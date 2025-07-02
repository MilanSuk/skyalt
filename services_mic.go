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

type ServicesMicRecord struct {
	data          []byte //int=hash
	startUnixTime float64
	Stop          atomic.Bool

	fnFinished func(buff *audio.IntBuffer)
}

type ServicesMicInfo struct {
	Active              bool
	Decibels            float64
	Decibels_normalized float64
}

type ServicesMic struct {
	services *Services

	lock   sync.Mutex
	device *malgo.Device
	mics   map[uint64]*ServicesMicRecord //int=hash

	info ServicesMicInfo
}

func NewServicesMic(services *Services) *ServicesMic {
	mic := &ServicesMic{services: services}

	mic.mics = make(map[uint64]*ServicesMicRecord)

	return mic
}

func (mic *ServicesMic) Destroy() {
	mic.FinishAll(true)
}

func _ServicesMic_GetDB(samples []byte) float64 {
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

func (mic *ServicesMic) _checkDevice() error {

	if mic.device != nil {
		return nil
	}

	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return err
	}

	source_mic := &mic.services.sync.Mic

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Duplex)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = uint32(source_mic.Channels)
	deviceConfig.Playback.Format = malgo.FormatS16
	deviceConfig.Playback.Channels = uint32(source_mic.Channels)
	deviceConfig.SampleRate = uint32(source_mic.Sample_rate)
	deviceConfig.Alsa.NoMMap = 1

	onRecvFrames := func(pSample2, pSample []byte, framecount uint32) {
		for key := range mic.mics {
			//there is 200ms delay between this and .Start()
			if mic.mics[key].startUnixTime == 0 {
				mic.mics[key].startUnixTime = float64(time.Now().UnixMilli()) / 1000
			}
			//add data
			mic.mics[key].data = append(mic.mics[key].data, pSample...) //add

			mic.info.Decibels = _ServicesMic_GetDB(pSample)
			mic.updateDecibels_normalized()
		}
	}

	mic.device, err = malgo.InitDevice(ctx.Context, deviceConfig, malgo.DeviceCallbacks{Data: onRecvFrames})
	if err != nil {
		return err
	}

	return nil
}

func (mic *ServicesMic) updateDecibels_normalized() {
	d := OsClampFloat(mic.info.Decibels, -60, -20)
	d += 20

	diff := 40.0
	if diff > 0 {
		d += diff //<0-100> where 100 is max volume
		d /= diff //<0-1>
	}

	mic.info.Decibels_normalized = d
}

func (mic *ServicesMic) Find(uid uint64) *ServicesMicRecord {
	mic.lock.Lock()
	defer mic.lock.Unlock()

	return mic.mics[uid]
}

func (mic *ServicesMic) Start(uid uint64) (*ServicesMicRecord, error) {
	mic.lock.Lock()
	defer mic.lock.Unlock()

	err := mic._checkDevice()
	if err != nil {
		return nil, err
	}

	//add
	_, found := mic.mics[uid]
	if found {
		return nil, LogsErrorf("mic UID '%d' already recording", uid)
	}

	micr := &ServicesMicRecord{}
	mic.mics[uid] = micr
	mic.info.Active = true

	if !mic.device.IsStarted() {
		err := mic.device.Start()
		if err != nil {
			return nil, err
		}
	}

	//fmt.Printf("Mic recording started: %d\n", hash)
	return micr, nil
}

func (mic *ServicesMic) _finished(uid uint64, cancel bool) (audio.IntBuffer, error) {

	//remove
	micr, found := mic.mics[uid]
	if !found {
		return audio.IntBuffer{}, LogsErrorf("uid '%d' not found", uid)
	}
	delete(mic.mics, uid)

	Sample_rate := int(mic.device.SampleRate())
	NumChannels := int(mic.device.CaptureChannels())

	//stop loop
	micr.Stop.Store(true)

	//stop device
	if len(mic.mics) == 0 && mic.device.IsStarted() {
		mic.device.Uninit()
		mic.device = nil
	}

	if len(mic.mics) == 0 {
		//deactivate
		mic.info.Active = false
		mic.info.Decibels = -100 //silence
		mic.info.Decibels_normalized = 0
	}

	if cancel {
		return audio.IntBuffer{}, nil
	}

	intData := make([]int, len(micr.data)/2)
	for i := 0; i+1 < len(micr.data); i += 2 {
		value := int(binary.LittleEndian.Uint16(micr.data[i : i+2]))
		intData[i/2] = value
	}

	ret_buff := audio.IntBuffer{Data: intData, Format: &audio.Format{SampleRate: Sample_rate, NumChannels: NumChannels}}

	if micr.fnFinished != nil {
		micr.fnFinished(&ret_buff)
	}

	//fmt.Printf("Mic recording stoped: %d\n", hash)
	return ret_buff, nil
}

func (mic *ServicesMic) Finished(uid uint64, cancel bool) (audio.IntBuffer, error) {
	mic.lock.Lock()
	defer mic.lock.Unlock()

	return mic._finished(uid, cancel)
}

func (mic *ServicesMic) FinishAll(cancel bool) {
	mic.lock.Lock()
	defer mic.lock.Unlock()

	for uid := range mic.mics {
		mic._finished(uid, cancel) //err ....
	}
}

// format: "wav", "mp3"
// sampleRate: 16000
func FFMpeg_convertIntoFile(input *audio.IntBuffer, format string, sampleRate int) ([]byte, error) {

	os.MkdirAll("temp", os.ModePerm)

	path := "temp/mic.wav"

	//file := &OsWriterSeeker{}	//....
	file, err := os.Create(path)
	if LogsError(err) != nil {
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
