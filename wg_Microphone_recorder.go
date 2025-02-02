package main

import (
	"encoding/binary"
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

	cancel            bool
	start             func()
	Out_buffer        audio.IntBuffer
	Out_startUnixTime float64
	done              func(out audio.IntBuffer)
}

func (layout *Layout) AddMicrophone_recorder(x, y, w, h int, props *Microphone_recorder) *Microphone_recorder {
	layout._createDiv(x, y, w, h, "Microphone_recorder", props.Build, nil, nil)
	return props
}

/*func OpenMemory_Microphone_recorder(uid string) *Microphone_recorder {
	st := &Microphone_recorder{UID: uid}
	return OpenMemory(uid, st)
}*/

func (st *Microphone_recorder) Build(layout *Layout) {

	layout.SetColumn(0, 1, 100)
	layout.SetRow(0, 1, 100)

	layout.Enable = OpenFile_Microphone().Enable

	var bt *Button
	var btL *Layout
	if st.Label == "" {
		bt, btL = layout.AddButtonIcon2(0, 0, 1, 1, "resources/mic.png", 0.15, st.Tooltip)
	} else {
		bt, btL = layout.AddButtonMenu2(0, 0, 1, 1, st.Label, "resources/mic.png", 0.15)
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
func (st *Microphone_recorder) FindJob() *Job {
	return FindJob(st.UID)
}

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

	st.Out_buffer, st.Out_startUnixTime, err = g_microphone_malgo.Stop(st.UID)
	if err != nil {
		job.AddError(err)
		return
	}
	if !st.cancel && st.done != nil {
		st.done(st.Out_buffer)
	}
}

type MicMalgoRecord struct {
	data          []byte //int=hash
	startUnixTime float64
}

type MicMalgo struct {
	lock   sync.Mutex
	device *malgo.Device
	mics   map[string]*MicMalgoRecord //int=hash
}

var g_microphone_malgo MicMalgo

func (mlg *MicMalgo) _checkDevice() error {

	if mlg.mics == nil {
		mlg.mics = make(map[string]*MicMalgoRecord)
	}

	if g_microphone_malgo.device != nil {
		return nil
	}

	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return err
	}

	mic := OpenFile_Microphone()

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Duplex)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = uint32(mic.Channels)
	deviceConfig.Playback.Format = malgo.FormatS16
	deviceConfig.Playback.Channels = uint32(mic.Channels)
	deviceConfig.SampleRate = uint32(mic.SampleRate)
	deviceConfig.Alsa.NoMMap = 1

	onRecvFrames := func(pSample2, pSample []byte, framecount uint32) {
		for key := range mlg.mics {
			//there is 200ms delay between this and .Start()
			if mlg.mics[key].startUnixTime == 0 {
				mlg.mics[key].startUnixTime = float64(time.Now().UnixMilli()) / 1000
			}
			//add data
			mlg.mics[key].data = append(mlg.mics[key].data, pSample...) //add
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
		return false
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
		mlg.mics[uid] = &MicMalgoRecord{}
	}

	if !mlg.device.IsStarted() {
		//fmt.Println("-----+++++++------- start", time.Now().UnixMilli())
		err := mlg.device.Start()
		if err != nil {
			return err
		}
	}

	//fmt.Printf("Mic recording started: %d\n", hash)
	return nil
}

func (mlg *MicMalgo) Stop(uid string) (audio.IntBuffer, float64, error) {
	mlg.lock.Lock()
	defer mlg.lock.Unlock()

	err := mlg._checkDevice()
	if err != nil {
		return audio.IntBuffer{}, 0, err
	}

	//remove
	rec := mlg.mics[uid]
	delete(mlg.mics, uid)

	SampleRate := int(mlg.device.SampleRate())
	NumChannels := int(mlg.device.CaptureChannels())

	//stop device
	if len(mlg.mics) == 0 && mlg.device.IsStarted() {
		mlg.device.Uninit()
		mlg.device = nil
	}

	intData := make([]int, len(rec.data)/2)
	for i := 0; i+1 < len(rec.data); i += 2 {
		value := int(binary.LittleEndian.Uint16(rec.data[i : i+2]))
		intData[i/2] = value
	}

	//fmt.Printf("Mic recording stoped: %d\n", hash)
	return audio.IntBuffer{Data: intData, Format: &audio.Format{SampleRate: SampleRate, NumChannels: NumChannels}}, rec.startUnixTime, nil
}
