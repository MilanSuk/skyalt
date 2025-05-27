package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gen2brain/malgo"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

// Start microphone recording [ignore]
type RecordMicrophone struct {
	UID string //Unique ID of recording session

	Stop         bool
	CancelResult bool

	Format string //Output bytes format [Options: wav, mp3]

	Out_bytes []byte
}

func (st *RecordMicrophone) run(caller *ToolCaller, ui *UI) error {

	if st.Stop {
		mic := g_malgo.Find(st.UID)
		if mic == nil {
			return fmt.Errorf("Mic UID '%s' not found", st.UID)
		}

		//stop "play" function call
		mic.Stop.Store(true)

		if st.CancelResult {
			return nil
		}

		//return data
		if st.Format != "wav" && st.Format != "mp3" {
			return fmt.Errorf("unknown format")
		}
		out, err := g_malgo.Finished(st.UID, false)
		if err != nil {
			return err
		}
		st.Out_bytes, err = FFMpeg_convertIntoFile(&out, st.Format == "mp3")
		if err != nil {
			return err
		}

	} else {

		source_mic, err := NewMicrophoneSettings("", caller)
		if err != nil {
			return err
		}

		mic, err := g_malgo.Start(st.UID, source_mic)
		if err != nil {
			return err
		}

		for !mic.Stop.Load() {
			if source_mic.Enable && !caller.Progress(0, "Listening") {
				g_malgo.Finished(st.UID, true)
				return nil //cancelled
			}

			time.Sleep(10 * time.Millisecond)
		}
	}

	return nil
}

type MicMalgoRecord struct {
	data          []byte //int=hash
	startUnixTime float64
	Stop          atomic.Bool
}

type MicMalgo struct {
	lock   sync.Mutex
	device *malgo.Device
	mics   map[string]*MicMalgoRecord //int=hash
}

var g_malgo = MicMalgo{mics: make(map[string]*MicMalgoRecord)}

func (mlg *MicMalgo) _checkDevice(source_mic *MicrophoneSettings) error {

	if g_malgo.device != nil {
		return nil
	}

	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return err
	}

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Duplex)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = uint32(source_mic.Channels)
	deviceConfig.Playback.Format = malgo.FormatS16
	deviceConfig.Playback.Channels = uint32(source_mic.Channels)
	deviceConfig.SampleRate = uint32(source_mic.Sample_rate)
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

func (mlg *MicMalgo) Find(uid string) *MicMalgoRecord {
	mlg.lock.Lock()
	defer mlg.lock.Unlock()

	return mlg.mics[uid]
}

func (mlg *MicMalgo) Start(uid string, source_mic *MicrophoneSettings) (*MicMalgoRecord, error) {
	mlg.lock.Lock()
	defer mlg.lock.Unlock()

	err := mlg._checkDevice(source_mic)
	if err != nil {
		return nil, err
	}

	//add
	mic, found := mlg.mics[uid]
	if found {
		return nil, fmt.Errorf("Mic UID '%s' already recording", uid)
	}

	mic = &MicMalgoRecord{}
	mlg.mics[uid] = mic

	if !mlg.device.IsStarted() {
		err := mlg.device.Start()
		if err != nil {
			return nil, err
		}
	}

	//fmt.Printf("Mic recording started: %d\n", hash)
	return mic, nil
}

func (mlg *MicMalgo) Finished(uid string, cancel bool) (audio.IntBuffer, error) {
	mlg.lock.Lock()
	defer mlg.lock.Unlock()

	//remove
	mic, found := mlg.mics[uid]
	if !found {
		return audio.IntBuffer{}, fmt.Errorf("uid '%s' not found", uid)
	}
	delete(mlg.mics, uid)

	Sample_rate := int(mlg.device.SampleRate())
	NumChannels := int(mlg.device.CaptureChannels())

	//stop device
	if len(mlg.mics) == 0 && mlg.device.IsStarted() {
		mlg.device.Uninit()
		mlg.device = nil
	}

	if cancel {
		return audio.IntBuffer{}, nil
	}

	intData := make([]int, len(mic.data)/2)
	for i := 0; i+1 < len(mic.data); i += 2 {
		value := int(binary.LittleEndian.Uint16(mic.data[i : i+2]))
		intData[i/2] = value
	}

	//fmt.Printf("Mic recording stoped: %d\n", hash)
	return audio.IntBuffer{Data: intData, Format: &audio.Format{SampleRate: Sample_rate, NumChannels: NumChannels}}, nil
}

func FFMpeg_convertIntoFile(input *audio.IntBuffer, mp3 bool) ([]byte, error) {

	os.MkdirAll("temp", os.ModePerm)

	path := "temp/mic.wav"

	//file := &OsWriterSeeker{}	//....
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	enc := wav.NewEncoder(file, input.Format.SampleRate, 16, input.Format.NumChannels, 1)
	err = enc.Write(input)
	if err != nil {
		enc.Close()
		file.Close()
		return nil, err
	}
	enc.Close()
	file.Close()

	if mp3 {
		compress_path := "temp/mic2.mp3"
		err := FFMpeg_convert(path, compress_path)
		if err != nil {
			return nil, err
		}
		path = compress_path
	} else {
		resample_path := "temp/mic2.wav"
		err := FFMpeg_convert(path, resample_path)
		if err != nil {
			return nil, err
		}
		path = resample_path
	}

	buff, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return buff, nil
}

func FFMpeg_convert(src, dst string) error {
	os.Remove(dst) //ffmpeg complains that 'file already exists'

	cmd := exec.Command("ffmpeg", "-i", src, "-ar", "16000", dst)
	cmd.Dir = ""
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
