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
#cgo LDFLAGS: -lvlc
#include <vlc/vlc.h>
#include <stdlib.h>

typedef struct {
    void* pixels;
    int width;
    int height;
} video_ctx;

static void* video_lock_cb(void* data, void** p_pixels) {
    video_ctx* ctx = (video_ctx*)data;
    *p_pixels = ctx->pixels;
    return NULL;
}

static void video_unlock_cb(void* data, void* id, void* const* p_pixels) {
    // Nothing to do
}

static void video_display_cb(void* data, void* id) {
    // Nothing to do - OpenGL will handle display
}

static void setup_video_callbacks(libvlc_media_player_t* player, void* data) {
    libvlc_video_set_callbacks(player, video_lock_cb, video_unlock_cb, video_display_cb, data);
}
*/
import "C"

import (
	"sync"
	"unsafe"
)

type ServicesPlayer struct {
	services *Services

	vlcInstance *C.libvlc_instance_t

	lock sync.Mutex

	media   []*ServicesPlayerItem
	playing bool //false = pause all

	last_volume float64
}

func NewServicesPlayer(services *Services) (*ServicesPlayer, error) {
	sps := &ServicesPlayer{services: services}
	sps.playing = true
	sps.last_volume = 1
	return sps, nil
}

func (sps *ServicesPlayer) Destroy() {
	sps.RemoveAll()
}

func (sps *ServicesPlayer) _tryActivateDevice() error {
	if sps.vlcInstance == nil {
		sps.vlcInstance = C.libvlc_new(0, nil)
		if sps.vlcInstance == nil {
			return LogsErrorf("VLC instance creation failed")
		}
	}
	return nil
}
func (sps *ServicesPlayer) _tryDeactivateDevice() {
	if len(sps.media) == 0 && sps.vlcInstance != nil {
		C.libvlc_release(sps.vlcInstance)
		sps.vlcInstance = nil
	}
}

func (sps *ServicesPlayer) _setPlay(play bool) {
	if play {
		//play last
		if len(sps.media) > 0 {
			sps.media[len(sps.media)-1].Play()
		}

	} else {
		for _, it := range sps.media {
			it.Pause()
		}
	}
	sps.playing = play
}

func (sps *ServicesPlayer) SetPlay(play bool) {
	sps.lock.Lock()
	defer sps.lock.Unlock()

	sps._setPlay(play)
}

func (sps *ServicesPlayer) Add(path string) error {
	sps.lock.Lock()
	defer sps.lock.Unlock()

	sps._tryActivateDevice()

	//create new media
	sp, err := NewServicesPlayerItem(path, sps)
	if err != nil {
		sps._tryDeactivateDevice()
	}

	//pause older
	sps._setPlay(false)

	//add new
	sps.media = append(sps.media, sp)

	if sps.playing {
		sp.Play()
	}

	return nil
}

func (sps *ServicesPlayer) Remove(path string) {
	sps.lock.Lock()
	defer sps.lock.Unlock()

	//stop and remove 'path'
	for _, it := range sps.media {
		if it.path == path {
			it.Destroy()
		}
	}

	//start playing latest
	sps._setPlay(true)

	sps._tryDeactivateDevice()
}

func (sps *ServicesPlayer) RemoveAll() {
	sps.lock.Lock()
	defer sps.lock.Unlock()

	for _, it := range sps.media {
		it.Destroy()
	}
	sps.media = nil
	sps._tryDeactivateDevice()
}

func (sps *ServicesPlayer) Tick() {
	sps.lock.Lock()
	defer sps.lock.Unlock()

	volume := sps.services.sync.Device.Volume

	if volume != sps.last_volume {
		for _, it := range sps.media {
			it.SetVolume(volume)
		}
		sps.last_volume = volume
	}
}

type ServicesPlayerItem struct {
	path   string
	player *C.libvlc_media_player_t
	media  *C.libvlc_media_t

	size   OsV2
	pixels unsafe.Pointer

	duration_ms int
}

func NewServicesPlayerItem(path string, speakers *ServicesPlayer) (*ServicesPlayerItem, error) {
	sp := &ServicesPlayerItem{path: path}

	// Create media player
	sp.player = C.libvlc_media_player_new(speakers.vlcInstance)
	if sp.player == nil {
		return nil, LogsErrorf("VLC player creation failed")
	}

	// Load media
	mediaPath := C.CString(path)
	defer C.free(unsafe.Pointer(mediaPath))
	sp.media = C.libvlc_media_new_path(speakers.vlcInstance, mediaPath)
	if sp.media == nil {
		return nil, LogsErrorf("media '%s' creation failed", path)
	}

	var width, height C.uint
	result := C.libvlc_video_get_size(sp.player, 0, &width, &height)

	if result == 0 && width > 0 && height > 0 {
		sp.size.X = int(width)
		sp.size.Y = int(height)

		pixelSize := sp.size.X * sp.size.Y * 4 // RGBA
		sp.pixels = C.malloc(C.size_t(pixelSize))

		var videoCtx C.video_ctx
		videoCtx.pixels = sp.pixels
		videoCtx.width = C.int(sp.size.X)
		videoCtx.height = C.int(sp.size.Y)

		// Setup VLC video callbacks
		C.setup_video_callbacks(sp.player, unsafe.Pointer(&videoCtx))
	}

	// Set video format
	formatStr := C.CString("RGBA")
	defer C.free(unsafe.Pointer(formatStr))
	C.libvlc_video_set_format(sp.player, formatStr, C.uint(sp.size.X), C.uint(sp.size.Y), C.uint(sp.size.X*4))

	C.libvlc_media_player_set_media(sp.player, sp.media)

	duration := C.libvlc_media_player_get_length(sp.player)
	sp.duration_ms = int(duration)

	return sp, nil
}

func (sp *ServicesPlayerItem) GetSeek() (int, int) {
	pos_ms := C.libvlc_media_player_get_time(sp.player)
	return int(pos_ms), sp.duration_ms
}

func (sp *ServicesPlayerItem) SetSeek(pos_ms int) {
	if pos_ms < 0 {
		pos_ms = 0
	}
	if pos_ms > sp.duration_ms {
		pos_ms = sp.duration_ms
	}
	C.libvlc_media_player_set_time(sp.player, C.libvlc_time_t(pos_ms))
}

func (sp *ServicesPlayerItem) Destroy() {
	C.libvlc_media_release(sp.media)

	C.libvlc_media_player_stop(sp.player)
	C.libvlc_media_player_release(sp.player)
	C.free(sp.pixels)
}

func (sp *ServicesPlayerItem) Pause() {
	if C.libvlc_media_player_is_playing(sp.player) > 0 {
		C.libvlc_media_player_pause(sp.player)
	}
}

func (sp *ServicesPlayerItem) Play() {
	C.libvlc_media_player_play(sp.player)
}

func (sp *ServicesPlayerItem) SetVolume(t float64) { //<0, 1>
	C.libvlc_audio_set_volume(sp.player, C.int(t*100)) //(0 = mute, 100 = 0dB)
}
