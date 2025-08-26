package main

/*
#cgo LDFLAGS: -lvlc
#include <vlc/vlc.h>
#include <stdlib.h>

typedef struct {
    void* pixels;
	int frame;
} video_ctx;

static void* video_lock_cb(void* data, void** p_pixels) {
    video_ctx* ctx = (video_ctx*)data;
    *p_pixels = ctx->pixels;
    return NULL;
}

static void video_unlock_cb(void* data, void* id, void* const* p_pixels) {
	video_ctx* ctx = (video_ctx*)data;
	ctx->frame++;
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
	"fmt"
	"log"
	"os"
	"sync"
	"time"
	"unsafe"
)

type VLC struct {
	instance *C.libvlc_instance_t

	media map[uint64]*VLCItem

	lock sync.Mutex
}

func NewVLC() *VLC {
	vlc := &VLC{}

	vlc.instance = C.libvlc_new(0, nil)
	if vlc.instance == nil {
		log.Fatal("libvlc_new() failed")
	}

	vlc.media = make(map[uint64]*VLCItem)

	return vlc
}

func (vlc *VLC) Destroy() {
	vlc.lock.Lock()
	defer vlc.lock.Unlock()

	for _, it := range vlc.media {
		it.Destroy()
	}

	C.libvlc_release(vlc.instance)
}

func (vlc *VLC) UpdateFileTimes() {

	vlc.lock.Lock()
	var ims []*VLCItem
	for _, it := range vlc.media {
		ims = append(ims, it)
	}
	vlc.lock.Unlock()

	//slow
	for _, sp := range ims {
		inf, err := os.Stat(sp.path)
		if err == nil && inf != nil {
			sp.check_file_time = inf.ModTime().UnixNano()
		}
	}
}

func (vlc *VLC) Maintenance(min_time int64, fnVlcChanged func(path string, playerID uint64)) {
	vlc.lock.Lock()
	defer vlc.lock.Unlock()

	for playerID, sp := range vlc.media {

		diff := (sp.check_file_time != sp.open_file_time)
		if diff {
			fnVlcChanged(sp.path, playerID) //file changed
		} else if sp.IsPlaying() {

			if C.int(sp.last_frame) != sp.videoCtx.frame {
				fnVlcChanged(sp.path, playerID) //new frame
				sp.last_frame = int(sp.videoCtx.frame)
			} else if sp.pixels_size == 0 {
				//also need to update time and seek for audio

				pos_ms := sp.GetSeek()
				d_ms := (pos_ms - sp.last_time_ms)
				if d_ms < 0 {
					d_ms *= -1
				}
				if d_ms > 800 { //every 800ms
					fnVlcChanged(sp.path, playerID) //new second
					sp.last_time_ms = pos_ms
				}
			}
		}

		if !sp.IsPlaying() {
			if (sp.last_use_time > 0 && sp.last_use_time < min_time) || diff { //diff(!), which mean it's delete here and .Add() later
				//fmt.Println("Maintenance() removing " + it.path)
				sp.Destroy()
				delete(vlc.media, playerID)
			}
		}
	}
}

func (vlc *VLC) Find(playerID uint64) *VLCItem {
	vlc.lock.Lock()
	defer vlc.lock.Unlock()

	item, found := vlc.media[playerID]
	if found {
		return item
	}
	return nil
}

func (vlc *VLC) Add(path string, playerID uint64) (*VLCItem, error) {

	//find
	item := vlc.Find(playerID)
	if item != nil {
		if item.path == path {
			return item, nil
		}

		item.Destroy()
		delete(vlc.media, playerID)
	}

	//create new media
	item, err := NewVLCItem(path, vlc)
	if err != nil {
		item = &VLCItem{path: path} //add it anyway, because file can be rewritten later. Error is return below.
	}

	vlc.lock.Lock()
	defer vlc.lock.Unlock()

	//add
	vlc.media[playerID] = item

	return item, err
}

type VLCItem struct {
	path   string
	player *C.libvlc_media_player_t
	media  *C.libvlc_media_t

	width       int
	height      int
	pixels_size int
	videoCtx    C.video_ctx

	last_frame   int
	last_time_ms int64

	last_use_time int64

	open_file_time  int64
	check_file_time int64
}

func NewVLCItem(path string, vlc *VLC) (*VLCItem, error) {
	sp := &VLCItem{path: path}

	//file_time
	inf, err := os.Stat(path)
	if err == nil && inf != nil {
		sp.open_file_time = inf.ModTime().UnixNano()
		sp.check_file_time = sp.open_file_time
	}

	// Create media player
	sp.player = C.libvlc_media_player_new(vlc.instance)
	if sp.player == nil {
		return nil, fmt.Errorf("VLC player creation failed")
	}

	// Load media
	mediaPath := C.CString(path)
	defer C.free(unsafe.Pointer(mediaPath))
	sp.media = C.libvlc_media_new_path(vlc.instance, mediaPath)
	if sp.media == nil {
		return nil, fmt.Errorf("media '%s' creation failed", path)
	}

	//get video size and duration
	{
		if C.libvlc_media_parse_with_options(sp.media, C.libvlc_media_parse_local, -1) != 0 {
			return nil, fmt.Errorf("failed to parse media")
		}

		// wait for parsing to complete (timeout after 5 seconds)
		for i := range 100 {
			status := C.libvlc_media_get_parsed_status(sp.media)
			if status == C.libvlc_media_parsed_status_done {
				break
			}
			if status == C.libvlc_media_parsed_status_failed {
				return nil, fmt.Errorf("media parsing failed")
			}
			// Sleep for 100ms
			time.Sleep(100 * time.Millisecond)
			if i == 99 {
				return nil, fmt.Errorf("media parsing timeout")
			}
		}

		// get tracks
		var tracks **C.libvlc_media_track_t
		trackCount := C.libvlc_media_tracks_get(sp.media, &tracks)
		if trackCount == 0 {
			return nil, fmt.Errorf("no tracks found in media")
		}
		defer C.libvlc_media_tracks_release(tracks, trackCount)

		// look for video track
		for i := 0; i < int(trackCount); i++ {
			cTrack := unsafe.Pointer(uintptr(unsafe.Pointer(tracks)) + uintptr(i)*unsafe.Sizeof(*tracks))
			track := *(**C.libvlc_media_track_t)(cTrack)

			if track.i_type == C.libvlc_track_video {
				video := *(**C.libvlc_video_track_t)(unsafe.Pointer(&track.anon0[0]))
				if video == nil {
					break
				}

				sp.width = int(video.i_width)
				sp.height = int(video.i_height)
			}
		}
	}

	if sp.width > 0 && sp.height > 0 {
		sp.pixels_size = sp.width * sp.height * 4 // RGBA
		sp.videoCtx.pixels = C.calloc(1, C.size_t(sp.pixels_size))

		C.setup_video_callbacks(sp.player, unsafe.Pointer(&sp.videoCtx))
	}

	// Set video format
	formatStr := C.CString("RGBA")
	defer C.free(unsafe.Pointer(formatStr))
	C.libvlc_video_set_format(sp.player, formatStr, C.uint(sp.width), C.uint(sp.height), C.uint(sp.width*4))

	C.libvlc_media_player_set_media(sp.player, sp.media)

	return sp, nil
}

func (sp *VLCItem) Destroy() {
	if sp.player == nil || sp.media == nil {
		return
	}

	C.libvlc_media_player_stop(sp.player)
	C.free(sp.videoCtx.pixels)
	sp.videoCtx.pixels = nil

	C.libvlc_media_player_release(sp.player)
	sp.player = nil

	C.libvlc_media_release(sp.media)
	sp.media = nil
}

func (sp *VLCItem) Pause() {
	if sp.player == nil {
		return
	}

	//if C.libvlc_media_player_is_playing(sp.player) > 0 {
	C.libvlc_media_player_pause(sp.player)
	//}
}

func (sp *VLCItem) Play() {
	if sp.player == nil {
		return
	}

	if C.libvlc_media_player_get_state(sp.player) == C.libvlc_Ended {
		C.libvlc_media_player_stop(sp.player)
	}

	C.libvlc_media_player_play(sp.player)
}

func (sp *VLCItem) IsPlaying() bool {
	if sp.player == nil {
		return false
	}

	return C.libvlc_media_player_is_playing(sp.player) > 0
}

func (sp *VLCItem) SetVolume(t float64) { //<0, 1>
	if sp.player == nil {
		return
	}

	C.libvlc_audio_set_volume(sp.player, C.int(t*100)) //(0 = mute, 100 = 0dB)
}

func (sp *VLCItem) GetVolume() float64 { //<0, 1>
	if sp.player == nil {
		return 0
	}

	vol := C.libvlc_audio_get_volume(sp.player) //(0 = mute, 100 = 0dB)
	return float64(vol) / 100
}

func (sp *VLCItem) GetDuration() int64 {
	if sp.media == nil {
		return 0
	}

	return int64(C.libvlc_media_get_duration(sp.media)) //return milliseconds
}

func (sp *VLCItem) GetSeek() int64 {
	if sp.player == nil {
		return 0
	}

	pos_ms := int64(C.libvlc_media_player_get_time(sp.player))
	if pos_ms < 0 {
		pos_ms = 0
	}
	return pos_ms
}

func (sp *VLCItem) SetSeek(pos_ms int64) {
	if sp.player == nil {
		return
	}

	curr_seek := sp.GetSeek()
	if pos_ms == curr_seek {
		return
	}

	end_ms := sp.GetDuration()
	if pos_ms >= end_ms {
		pos_ms = end_ms - 1
	}

	if !sp.IsPlaying() {
		sp.Play()
	}

	C.libvlc_media_player_set_time(sp.player, C.libvlc_time_t(pos_ms))

	if !sp.IsPlaying() {
		if sp.pixels_size > 0 {
			//video
			for C.int(sp.last_frame) == sp.videoCtx.frame {
				time.Sleep(1 * time.Millisecond)
			}
		} else {
			//audio
			for !sp.IsPlaying() {
				time.Sleep(10 * time.Millisecond)
			}
		}
		sp.Pause()
	}
}
