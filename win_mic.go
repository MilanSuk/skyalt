/*
Copyright 2024 Milan Suk

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this db except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

/*
typedef unsigned char Uint8;
void WinMic_OnAudio(void *userdata, Uint8 *stream, int length);
*/
import "C"
import (
	"os"
	"os/exec"
	"sync"
	"unsafe"

	"github.com/go-audio/audio"
	"github.com/veandco/go-sdl2/sdl"
)

type WinMic struct {
	spec   sdl.AudioSpec
	device sdl.AudioDeviceID

	lock sync.Mutex
	data audio.IntBuffer
}

// int32, but values are in range of int16
//var g_audio_data = audio.IntBuffer{Data: nil, SourceBitDepth: 16, Format: &audio.Format{NumChannels: 1, SampleRate: 44100}}
//var g_audio_mu sync.Mutex

var g_win_mics []*WinMic
var g_win_mics_lock sync.Mutex

const g_win_mics_SourceBitDepth = 16

//export WinMic_OnAudio
func WinMic_OnAudio(userdata unsafe.Pointer, _stream *C.Uint8, _length C.int) {
	g_win_mics_lock.Lock()
	defer g_win_mics_lock.Unlock()

	length := int(_length) / (g_win_mics_SourceBitDepth / 8)
	header := unsafe.Slice(_stream, length)
	src_buf := *(*[]int16)(unsafe.Pointer(&header))

	//prepare
	dst := make([]int, length) //int32, but values are in range of int16
	for i, v := range src_buf {
		dst[i] = int(v) //v=int16t, which is saved into int32(audio.IntBuffer is always []int32)
	}

	//add to all mics
	for _, mic := range g_win_mics {
		mic.data.Data = append(mic.data.Data, dst...)
	}
}

func NewWinMic() (*WinMic, error) {
	mic := &WinMic{}

	mic.data = audio.IntBuffer{Data: nil, SourceBitDepth: g_win_mics_SourceBitDepth, Format: &audio.Format{NumChannels: 1, SampleRate: 44100}}

	var spec sdl.AudioSpec
	spec.Freq = int32(mic.data.Format.SampleRate)
	spec.Format = sdl.AUDIO_S16 //audio_data.SourceBitDepth!!!
	spec.Channels = uint8(mic.data.Format.NumChannels)
	spec.Samples = 4046
	spec.Callback = sdl.AudioCallback(C.WinMic_OnAudio)
	//spec.UserData = unsafe.Pointer(mic)	//creates panic. needs to be C.malloc()

	var err error
	//defaultRecordingDeviceName := sdl.GetAudioDeviceName(0, true)
	mic.device, err = sdl.OpenAudioDevice("", true, &spec, nil, 0)
	if err != nil {
		return nil, err
	}

	//add to global
	g_win_mics_lock.Lock()
	g_win_mics = append(g_win_mics, mic)
	g_win_mics_lock.Unlock()

	mic.SetEnable(true)

	return mic, nil
}
func (mic *WinMic) Destroy() {
	sdl.CloseAudioDevice(mic.device)

	//remove from global
	g_win_mics_lock.Lock()
	for i := range g_win_mics {
		if g_win_mics[i] == mic {
			g_win_mics = append(g_win_mics[:i], g_win_mics[i+1:]...) //remove
			break
		}
	}
	g_win_mics_lock.Unlock()
}

func (mic *WinMic) SetEnable(record_now bool) {
	sdl.PauseAudioDevice(mic.device, !record_now)

	//BUG: when mic is enabled, it has ~2s warmup, when audio has alot of noise .........

	if record_now {
		sdl.ClearQueuedAudio(mic.device) //is this useful?
	} else {
		mic.Get() //clean buffer
	}
}

func (mic *WinMic) IsPlaying() bool {
	return sdl.GetAudioDeviceStatus(mic.device) == sdl.AUDIO_PLAYING
}

func (mic *WinMic) Get() audio.IntBuffer {
	mic.lock.Lock()
	defer mic.lock.Unlock()

	ret := mic.data
	mic.data.Data = nil

	return ret
}

func FFMpeg_convert(src, dst string) error {

	OsFileRemove(dst) //ffmpeg complains that 'file already exists'

	cmd := exec.Command("ffmpeg", "-i", src, dst)
	cmd.Dir = ""
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
