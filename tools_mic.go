package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gen2brain/malgo"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

type ToolsMicMalgoRecord struct {
	data          []byte //int=hash
	startUnixTime float64
	Stop          atomic.Bool

	fnFinished func(buff *audio.IntBuffer)
}

type ToolsMicMalgo struct {
	router *ToolsRouter

	lock   sync.Mutex
	device *malgo.Device
	mics   map[uint64]*ToolsMicMalgoRecord //int=hash

	decibels float64
}

func NewToolsMicMalgo(router *ToolsRouter) *ToolsMicMalgo {
	st := &ToolsMicMalgo{router: router}

	st.mics = make(map[uint64]*ToolsMicMalgoRecord)

	return st
}

func (mlg *ToolsMicMalgo) Destroy() {
	mlg.FinishAll(true)
}

func _ToolsMicMalgo_GetDB(samples []byte) float64 {
	if len(samples) < 2 || len(samples)%2 != 0 {
		return -100.0
	}

	// Calculate Root Mean Square
	var sumSquares float64
	sampleCount := len(samples) / 2
	for i := 0; i < len(samples); i += 2 {
		// Convert two bytes to int16 (little-endian)
		sample := int16(samples[i]) | int16(samples[i+1])<<8
		normalized := float64(sample) / 32768.0
		sumSquares += normalized * normalized
	}

	rms := math.Sqrt(sumSquares / float64(sampleCount))

	if rms < 1e-10 { // Avoid log(0)
		return -100.0
	}

	// Convert RMS to decibels
	return 20.0 * math.Log10(rms)
}

func (mlg *ToolsMicMalgo) _checkDevice() error {

	if mlg.device != nil {
		return nil
	}

	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return err
	}

	source_mic := &mlg.router.sync.Mic

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

			mlg.decibels = _ToolsMicMalgo_GetDB(pSample)
		}
	}

	mlg.device, err = malgo.InitDevice(ctx.Context, deviceConfig, malgo.DeviceCallbacks{Data: onRecvFrames})
	if err != nil {
		return err
	}

	return nil
}

func (mlg *ToolsMicMalgo) Find(uid uint64) *ToolsMicMalgoRecord {
	mlg.lock.Lock()
	defer mlg.lock.Unlock()

	return mlg.mics[uid]
}

func (mlg *ToolsMicMalgo) Start(uid uint64) (*ToolsMicMalgoRecord, error) {
	mlg.lock.Lock()
	defer mlg.lock.Unlock()

	err := mlg._checkDevice()
	if err != nil {
		return nil, err
	}

	//add
	_, found := mlg.mics[uid]
	if found {
		return nil, fmt.Errorf("mic UID '%d' already recording", uid)
	}

	mic := &ToolsMicMalgoRecord{}
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

func (mlg *ToolsMicMalgo) _finished(uid uint64, cancel bool) (audio.IntBuffer, error) {

	//remove
	mic, found := mlg.mics[uid]
	if !found {
		return audio.IntBuffer{}, fmt.Errorf("uid '%d' not found", uid)
	}
	delete(mlg.mics, uid)

	Sample_rate := int(mlg.device.SampleRate())
	NumChannels := int(mlg.device.CaptureChannels())

	//stop loop
	mic.Stop.Store(true)

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

	ret_buff := audio.IntBuffer{Data: intData, Format: &audio.Format{SampleRate: Sample_rate, NumChannels: NumChannels}}

	if mic.fnFinished != nil {
		mic.fnFinished(&ret_buff)
	}

	//fmt.Printf("Mic recording stoped: %d\n", hash)
	return ret_buff, nil
}

func (mlg *ToolsMicMalgo) Finished(uid uint64, cancel bool) (audio.IntBuffer, error) {
	mlg.lock.Lock()
	defer mlg.lock.Unlock()

	return mlg._finished(uid, cancel)
}

func (mlg *ToolsMicMalgo) FinishAll(cancel bool) {
	mlg.lock.Lock()
	defer mlg.lock.Unlock()

	for uid := range mlg.mics {
		mlg._finished(uid, cancel) //err ....
	}
}

// format: "wav", "mp3"
// sampleRate: 16000
func FFMpeg_convertIntoFile(input *audio.IntBuffer, format string, sampleRate int) ([]byte, error) {

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

	if format != "wav" || input.Format.SampleRate != sampleRate {
		compress_path := "temp/mic2." + format
		err := FFMpeg_convert(path, compress_path, sampleRate)
		if err != nil {
			return nil, err
		}
		path = compress_path
	}

	buff, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return buff, nil
}

func FFMpeg_convert(src, dst string, samples int) error {
	os.Remove(dst) //ffmpeg complains that 'file already exists'

	cmd := exec.Command("ffmpeg", "-i", src, "-ar", fmt.Sprintf("%d", samples), dst)
	cmd.Dir = ""
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
