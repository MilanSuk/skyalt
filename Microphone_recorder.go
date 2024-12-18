package main

import (
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/gen2brain/malgo"
	"github.com/go-audio/audio"
)

type Microphone_recorder struct {
	UID string

	Label        string
	Tooltip      string
	Shortcut_key byte
	Background   float64

	cancel bool
	start  func()
	Out    audio.IntBuffer
	done   func()
}

func (layout *Layout) AddMicrophone_recorder(x, y, w, h int, props *Microphone_recorder) *Microphone_recorder {
	layout._createDiv(x, y, w, h, "Microphone_recorder", props.Build, nil, nil)
	return props
}

var g_global_Microphone_recorder = make(map[string]*Microphone_recorder)

func NewGlobal_Microphone_recorder(uid string) *Microphone_recorder {
	uid = fmt.Sprintf("Microphone_recorder:%s", uid)

	st, found := g_global_Microphone_recorder[uid]
	if !found {
		st = &Microphone_recorder{UID: uid}
		g_global_Microphone_recorder[uid] = st
	}
	return st
}
func (st *Microphone_recorder) Build(layout *Layout) {

	layout.SetColumn(0, 1, 100)
	layout.SetRow(0, 1, 100)

	layout.Enable = NewFile_Microphone().Enable

	var bt *Button
	var btL *Layout
	if st.Label == "" {
		bt, btL = layout.AddButton2(0, 0, 1, 1, NewButtonIcon("resources/mic.png", 0.15, st.Tooltip))
	} else {
		bt, btL = layout.AddButton2(0, 0, 1, 1, NewButtonMenu(st.Label, "resources/mic.png", 0.15))
	}

	job := FindJob(st.UID)
	if job != nil {
		bt.Background = 1 //active
		bt.Cd = Paint_GetPalette().E
	} else {
		bt.Background = st.Background //no recording
	}
	btL.Shortcut_key = st.Shortcut_key

	bt.clicked = func() {
		if job == nil {
			st.Start()
		} else {
			job.Stop()
		}
	}
}

func (st *Microphone_recorder) Start() *Job {
	st.cancel = false
	return StartJob(st.UID, "Recording from microphone", st.Run)
}
func (st *Microphone_recorder) Cancel() {
	job := FindJob(st.UID)
	if job != nil {
		st.cancel = true
		job.Stop()
	}
}
func (st *Microphone_recorder) IsRunning() bool {
	return FindJob(st.UID) != nil
}

var g__mic_lock sync.Mutex
var g__mic_device *malgo.Device

func (st *Microphone_recorder) Run(job *Job) {

	err := g_microphone_malgo.Start(st.UID)
	if err != nil {
		job.AddError(err)
		return
	}

	if st.start != nil {
		st.start()
	}

	for job.IsRunning() {
		time.Sleep(10 * time.Millisecond)
	}

	st.Out, err = g_microphone_malgo.Stop(st.UID)
	if err != nil {
		job.AddError(err)
		return
	}
	if !st.cancel && st.done != nil {
		st.done()
	}
}

type MicMalgo struct {
	lock   sync.Mutex
	device *malgo.Device
	mics   map[string][]byte //int=hash
}

var g_microphone_malgo MicMalgo

func (mlg *MicMalgo) _checkDevice() error {

	if mlg.mics == nil {
		mlg.mics = make(map[string][]byte)
	}

	if g_microphone_malgo.device != nil {
		return nil
	}

	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return err
	}

	mic := NewFile_Microphone()

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Duplex)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = uint32(mic.Channels)
	deviceConfig.Playback.Format = malgo.FormatS16
	deviceConfig.Playback.Channels = uint32(mic.Channels)
	deviceConfig.SampleRate = uint32(mic.SampleRate)
	deviceConfig.Alsa.NoMMap = 1

	onRecvFrames := func(pSample2, pSample []byte, framecount uint32) {
		for key := range mlg.mics {
			mlg.mics[key] = append(mlg.mics[key], pSample...) //add
		}
	}

	mlg.device, err = malgo.InitDevice(ctx.Context, deviceConfig, malgo.DeviceCallbacks{Data: onRecvFrames})
	if err != nil {
		return err
	}

	return nil
}

func (mlg *MicMalgo) Find(uid string) bool {
	mlg.lock.Lock()
	defer mlg.lock.Unlock()

	if mlg.mics == nil {
		mlg.mics = make(map[string][]byte)
	}

	_, found := mlg.mics[uid]
	return found
}

func (mlg *MicMalgo) Start(uid string) error {
	mlg.lock.Lock()
	defer mlg.lock.Unlock()

	err := mlg._checkDevice()
	if err != nil {
		return err
	}

	//add
	_, found := mlg.mics[uid]
	if !found {
		mlg.mics[uid] = nil
	}

	if !mlg.device.IsStarted() {
		err := mlg.device.Start()
		if err != nil {
			return err
		}
	}

	//fmt.Printf("Mic recording started: %d\n", hash)
	return nil
}

func (mlg *MicMalgo) Stop(uid string) (audio.IntBuffer, error) {
	mlg.lock.Lock()
	defer mlg.lock.Unlock()

	err := mlg._checkDevice()
	if err != nil {
		return audio.IntBuffer{}, err
	}

	//remove
	audioData := mlg.mics[uid]
	delete(mlg.mics, uid)

	SampleRate := int(mlg.device.SampleRate())
	NumChannels := int(mlg.device.CaptureChannels())

	//stop device
	if len(mlg.mics) == 0 && mlg.device.IsStarted() {
		mlg.device.Uninit()
		mlg.device = nil
	}

	intData := make([]int, len(audioData)/2)
	for i := 0; i+1 < len(audioData); i += 2 {
		value := int(binary.LittleEndian.Uint16(audioData[i : i+2]))
		intData[i/2] = value
	}

	//fmt.Printf("Mic recording stoped: %d\n", hash)
	return audio.IntBuffer{Data: intData, Format: &audio.Format{SampleRate: SampleRate, NumChannels: NumChannels}}, nil
}
