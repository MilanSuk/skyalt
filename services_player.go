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
	"slices"
	"sync"
	"time"
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

func (sps *ServicesPlayer) Add(path string) (*ServicesPlayerItem, error) {
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

	return sp, nil
}

func (sps *ServicesPlayer) Remove(path string) {
	sps.lock.Lock()
	defer sps.lock.Unlock()

	//stop and remove 'path'
	for i := len(sps.media) - 1; i >= 0; i-- {
		it := sps.media[i]
		if it.path == path {
			it.Destroy()
			sps.media = slices.Delete(sps.media, i, i+1)
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

	//get video size and duration
	{
		// Parse media to get track information
		if C.libvlc_media_parse_with_options(sp.media, C.libvlc_media_parse_local, -1) != 0 {
			return nil, LogsErrorf("failed to parse media")
		}

		// Wait for parsing to complete (timeout after 5 seconds)
		for i := range 50 {
			status := C.libvlc_media_get_parsed_status(sp.media)
			if status == C.libvlc_media_parsed_status_done {
				break
			}
			if status == C.libvlc_media_parsed_status_failed {
				return nil, LogsErrorf("media parsing failed")
			}
			// Sleep for 100ms
			time.Sleep(100 * time.Millisecond)
			if i == 49 {
				return nil, LogsErrorf("media parsing timeout")
			}
		}

		// Get tracks
		var tracks **C.libvlc_media_track_t
		trackCount := C.libvlc_media_tracks_get(sp.media, &tracks)
		if trackCount == 0 {
			return nil, LogsErrorf("no tracks found in media")
		}
		defer C.libvlc_media_tracks_release(tracks, trackCount)

		// Look for video track
		for i := 0; i < int(trackCount); i++ {

			cTrack := unsafe.Pointer(uintptr(unsafe.Pointer(tracks)) + uintptr(i)*unsafe.Sizeof(*tracks))
			track := *(**C.libvlc_media_track_t)(cTrack)

			if track.i_type == C.libvlc_track_video {
				video := *(**C.libvlc_video_track_t)(unsafe.Pointer(&track.anon0[0]))
				if video == nil {
					break
				}

				sp.size.X = int(video.i_width)
				sp.size.Y = int(video.i_height)
			}
		}

		duration := C.libvlc_media_get_duration(sp.media)
		sp.duration_ms = int(duration)
	}

	if sp.size.Is() {
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

	//sp.Play()

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
	C.libvlc_media_player_stop(sp.player)
	C.free(sp.pixels)
	C.libvlc_media_player_release(sp.player)
	C.libvlc_media_release(sp.media)
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
