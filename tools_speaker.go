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

type ToolsSpeakers struct {
	router *ToolsRouter

	lock sync.Mutex

	speakers []*ToolsSpeaker
	paused   bool
}

func NewToolsSpeakers(router *ToolsRouter) *ToolsSpeakers {
	sps := &ToolsSpeakers{router: router}

	return sps
}

func (sps *ToolsSpeakers) Destroy() {
	sps.RemoveAll()
}

func (sps *ToolsSpeakers) activateDevice() error {
	if len(sps.speakers) == 0 {
		err := mix.OpenAudio(48000, mix.DEFAULT_FORMAT, 2, 4096)
		if err != nil {
			return err
		}
		mix.VolumeMusic(int(float64(mix.MAX_VOLUME) * sps.router.sync.Device.Volume))
	}
	return nil
}
func (sps *ToolsSpeakers) deactivateDevice() {
	if len(sps.speakers) == 0 {
		mix.CloseAudio()
	}
}

func (sps *ToolsSpeakers) Add(path string) error {
	sps.lock.Lock()
	defer sps.lock.Unlock()

	sps.activateDevice()

	sp, _ := NewToolsSpeaker(path) //err ....

	for _, it := range sps.speakers {
		it.Pause()
	}

	sps.speakers = append(sps.speakers, sp)
	//....

	return nil
}

func (sps *ToolsSpeakers) Remove(path string) {
	sps.lock.Lock()
	defer sps.lock.Unlock()

	for _, it := range sps.speakers {
		if it.path == path {
			it.Destroy()
		}
	}

	//....

	sps.deactivateDevice()
}

func (sps *ToolsSpeakers) RemoveAll() {
	for _, it := range sps.speakers {
		it.Destroy()
	}
	sps.speakers = nil
	sps.deactivateDevice()
}

//update volume ....

type ToolsSpeaker struct {
	path  string
	music *mix.Music
}

func NewToolsSpeaker(path string) (*ToolsSpeaker, error) {
	speak := &ToolsSpeaker{path: path}

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

func (speak *ToolsSpeaker) Destroy() {
	if speak.music != nil {
		speak.music.Free()
	}
}
func (speak *ToolsSpeaker) Pause() {
	if speak.music != nil {
		//speak.music.
	}
}
