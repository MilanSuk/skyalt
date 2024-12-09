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

import (
	"sync"

	"github.com/veandco/go-sdl2/mix"
)

var g_win_speaker_num = 0
var g_win_speaker_lock sync.Mutex
var g_win_speaker_volume = 1.0

type WinSpeaker struct {
	music *mix.Music
}

func NewWinSpeaker(path string) (*WinSpeaker, error) {
	speak := &WinSpeaker{}

	//global
	{
		g_win_speaker_lock.Lock()
		if g_win_speaker_num == 0 {
			err := mix.OpenAudio(48000, mix.DEFAULT_FORMAT, 2, 4096)
			if err != nil {
				g_win_speaker_lock.Unlock()
				return nil, err
			}
			WinSpeaker_setVolume(g_win_speaker_volume)
		}
		g_win_speaker_num++
		g_win_speaker_lock.Unlock()
	}

	var err error
	speak.music, err = mix.LoadMUS(path)
	if err != nil {
		speak.Destroy()
		return nil, err
	}

	err = speak.music.Play(1)
	if err != nil {
		speak.Destroy()
		return nil, err
	}

	return speak, nil
}

func (speak *WinSpeaker) Destroy() {
	if speak.music != nil {
		speak.music.Free()
	}

	//global
	{
		g_win_speaker_lock.Lock()
		g_win_speaker_num--
		if g_win_speaker_num == 0 {
			mix.CloseAudio()
		}
		g_win_speaker_lock.Unlock()
	}
}

// do not call, set win.io.Ini.Volume
func WinSpeaker_setVolume(t float64) { //<0, 1>
	g_win_speaker_volume = t
	mix.VolumeMusic(int(float64(mix.MAX_VOLUME) * t))
}
