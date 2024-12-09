package main

import (
	"encoding/binary"
	"sync"

	"github.com/gen2brain/malgo"
	"github.com/go-audio/audio"
)

func (st *Microphone_recorder) Build() {
	st.lock.Lock()
	defer st.lock.Unlock()

	st.layout.SetColumn(0, 1, 100)
	st.layout.SetRow(0, 1, 100)

	st.layout.Enable = NewFile_Microphone().Enable

	var bt *Button
	if st.Label == "" {
		bt = st.layout.AddButton(0, 0, 1, 1, NewButtonIcon("resources/mic.png", 0.15, st.Tooltip))
	} else {
		bt = st.layout.AddButton(0, 0, 1, 1, NewButtonMenu(st.Label, "resources/mic.png", 0.15))
	}

	isRecording := st.IsRecording()
	if isRecording {
		bt.Background = 1 //active
		bt.Cd = st.layout.GetPalette().E
	} else {
		bt.Background = st.Background //no recording
	}
	bt.layout.Shortcut_key = st.Shortcut_key

	bt.clicked = func() {
		if !isRecording {
			//start
			st.StartRecording()
		} else {
			//finish
			st.FinishRecording()
		}
	}
}

func (st *Microphone_recorder) IsRecording() bool {
	return g_microphone_malgo.Find(st.layout.Hash)
}
func (st *Microphone_recorder) StartRecording() {
	err := g_microphone_malgo.Start(st.layout.Hash, NewFile_Microphone())
	if err != nil {
		st.layout.WriteError(err)
		return
	}
	if st.start != nil {
		st.start()
	}
}
func (st *Microphone_recorder) FinishRecording() {
	buff, err := g_microphone_malgo.Stop(st.layout.Hash, NewFile_Microphone())
	if err != nil {
		st.layout.WriteError(err)
		return
	}

	if st.done != nil {
		st.done(buff) //...
	}
}
func (st *Microphone_recorder) CancelRecording() {
	_, err := g_microphone_malgo.Stop(st.layout.Hash, NewFile_Microphone())
	if err != nil {
		st.layout.WriteError(err)
		return
	}
}

type MicMalgo struct {
	lock   sync.Mutex
	device *malgo.Device
	mics   map[uint64][]byte //int=hash
}

var g_microphone_malgo MicMalgo

func (mlg *MicMalgo) _checkDevice(mic *Microphone) error {

	if mlg.mics == nil {
		mlg.mics = make(map[uint64][]byte)
	}

	if g_microphone_malgo.device != nil {
		return nil
	}

	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return err
	}

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

func (mlg *MicMalgo) Find(hash uint64) bool {
	mlg.lock.Lock()
	defer mlg.lock.Unlock()

	if mlg.mics == nil {
		mlg.mics = make(map[uint64][]byte)
	}

	_, found := mlg.mics[hash]
	return found
}

func (mlg *MicMalgo) Start(hash uint64, mic *Microphone) error {
	mlg.lock.Lock()
	defer mlg.lock.Unlock()

	err := mlg._checkDevice(mic)
	if err != nil {
		return err
	}

	//add
	_, found := mlg.mics[hash]
	if !found {
		mlg.mics[hash] = nil
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

func (mlg *MicMalgo) Stop(hash uint64, mic *Microphone) (audio.IntBuffer, error) {
	mlg.lock.Lock()
	defer mlg.lock.Unlock()

	err := mlg._checkDevice(mic)
	if err != nil {
		return audio.IntBuffer{}, err
	}

	//remove
	audioData := mlg.mics[hash]
	delete(mlg.mics, hash)

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
