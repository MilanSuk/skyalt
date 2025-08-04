package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

func isFileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func getFileExtensionFromUrl(path string) (string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", err
	}
	pos := strings.LastIndex(u.Path, ".")
	if pos == -1 {
		return "", errors.New("'.' not found")
	}
	return u.Path[pos+1 : len(u.Path)], nil
}
func IsUrl(path string) bool {
	path = strings.TrimSpace(path)
	path = strings.ToLower(path)
	return strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")
}

func getType(path string) (bool, int) {
	var ext string
	isUrl := IsUrl(path)
	if isUrl {
		ext, _ = getFileExtensionFromUrl(path)
	} else {
		ext = strings.ToLower(filepath.Ext(path))
	}

	img := (ext == ".png" || ext == ".webp" || ext == ".jpeg" || ext == ".jpg" || ext == ".gif" || ext == ".tiff" || ext == ".bmp")
	vid := (ext == ".mp4" || ext == ".mkv" || ext == ".webm" || ext == ".mov" || ext == ".avi" || ext == ".flv")
	aud := (ext == ".wav" || ext == ".mp3" || ext == ".opus" || ext == ".aac" || ext == ".ogg" || ext == ".flac" || ext == ".pcm")

	if !img && !vid && !aud {
		img = true //default = try to open as image
	}

	if vid {
		return isUrl, 1
	}
	if aud {
		return isUrl, 2
	}

	return isUrl, 0 //default = try to open as image
}

func main() {
	log.SetFlags(log.Llongfile) //log.LstdFlags | log.Lshortfile

	vlc := NewVLC()
	defer vlc.Destroy()

	imgs := NewImages()
	defer imgs.Destroy()

	cache := NewCache()
	defer cache.Destroy()

	/*{
		//Tests

		fmt.Println(cache.Get("https://tile.openstreetmap.org/5/5/4.png"))

		av, _ := vlc.Add("../vid.mkv", 0)
		//vlc.Add("../aud.mp3", 1)
		//imgs.Add("../resources/logo_small.png", nil)

		av.Play()
		//av.SetSeek(100000)

		time.Sleep(3 * time.Second)
		av.Pause()

		for {
			imgs.UpdateFileTimes()
			vlc.UpdateFileTimes()

			min_time := time.Now().Add(-60 * time.Second).UnixNano()

			fnImgChanged := func(path string) {

			}

			fnVlcChanged := func(path string, playerID uint64) {

			}

			imgs.Maintenance(min_time, fnImgChanged)
			vlc.Maintenance(min_time, fnVlcChanged)

			//fmt.Println(av.GetSeek(), av.GetDuration(), av.IsPlaying())

			//b := imgs.Check("../resources/logo_small.png")
			//fmt.Println(p, a)
			//fmt.Println(p2, a2)
			//fmt.Println(b)

			time.Sleep(1 * time.Second)
		}
	}*/

	if len(os.Args) < 2 {
		log.Fatal("missing 'server port' argument(s): ", os.Args)
	}

	server_port, err := strconv.Atoi(os.Args[1])
	if err != nil || server_port <= 0 {
		log.Fatal(err)
	}

	cl_tasks := NewNetClient("localhost", server_port)
	defer cl_tasks.Destroy()

	var client_updates_lock sync.Mutex
	client_updates := NewNetClient("localhost", server_port)
	defer func() {
		client_updates_lock.Lock()
		defer client_updates_lock.Unlock()
		client_updates.WriteArray([]byte("exit")) //exit
		client_updates.Destroy()
	}()

	//check if files changed
	go func() {
		for {
			imgs.UpdateFileTimes()
			vlc.UpdateFileTimes()

			time.Sleep(1 * time.Second) //1s
		}
	}()

	//send updates(file changed or new frame)
	go func() {
		for {
			min_time := time.Now().Add(-60 * time.Second).UnixNano()

			fnImgChanged := func(path string) {
				client_updates_lock.Lock()
				defer client_updates_lock.Unlock()
				client_updates.WriteArray([]byte("image"))
				client_updates.WriteArray([]byte(path))

			}
			fnVlcChanged := func(path string, playerID uint64) {
				client_updates_lock.Lock()
				defer client_updates_lock.Unlock()
				client_updates.WriteArray([]byte("video"))
				client_updates.WriteArray([]byte(path))
				client_updates.WriteInt(playerID)
			}

			imgs.Maintenance(min_time, fnImgChanged)
			vlc.Maintenance(min_time, fnVlcChanged)

			time.Sleep(10 * time.Millisecond) //10ms
		}
	}()

	for {
		command := cl_tasks.ReadArray()

		switch string(command) {
		case "exit":
			return

		case "type":
			path := cl_tasks.ReadArray() //must be set!

			_, tp := getType(string(path))
			cl_tasks.WriteInt(uint64(tp))

		case "info":
			type MediaItem struct {
				Path      string
				Volume    float64
				Seek      int64
				Duration  int64
				IsPlaying bool
			}
			//send as array and sort it by path? ....
			info := make(map[uint64]MediaItem)
			for playerID, it := range vlc.media {
				if it.GetSeek() > 0 {
					info[playerID] = MediaItem{Path: it.path, Volume: it.GetVolume(), Seek: it.GetSeek(), Duration: it.GetDuration(), IsPlaying: it.IsPlaying()}
				}
			}
			infoJs, err := json.Marshal(info)
			if err != nil {
				log.Fatal(err)
			}
			cl_tasks.WriteArray(infoJs)

		case "frame":
			path := cl_tasks.ReadArray()   //must be set!
			blob := cl_tasks.ReadArray()   //for images only
			playerID := cl_tasks.ReadInt() //for audio/video only

			var errBytes []byte
			var width uint64
			var height uint64
			var data []byte
			var seek_ms uint64
			var play_duration uint64

			isUrl, tp := getType(string(path))
			if tp > 0 { //audio/video
				item, err := vlc.Add(string(path), playerID)
				if err == nil {
					width = uint64(item.width)
					height = uint64(item.height)
					seek_ms = uint64(item.GetSeek())
					play_duration = uint64(item.GetDuration())
					data = unsafe.Slice((*byte)(item.videoCtx.pixels), item.pixels_size)
				} else {
					errBytes = []byte(err.Error())
				}
			} else { //image
				pathStr := string(path)
				if isUrl {
					pathStr, err = cache.Get(string(path)) //download or get /temp path
					if err != nil {
						errBytes = []byte(err.Error())
					}
				}

				item, err := imgs.Add(pathStr, blob)
				if err == nil {
					width = uint64(item.width)
					height = uint64(item.height)
					data = item.data
				} else {
					errBytes = []byte(err.Error())
				}
			}

			cl_tasks.WriteArray(errBytes)
			cl_tasks.WriteInt(width)
			cl_tasks.WriteInt(height)
			cl_tasks.WriteArray(data)
			cl_tasks.WriteInt(seek_ms)
			cl_tasks.WriteInt(play_duration)
			cl_tasks.WriteInt(uint64(tp))

		case "play": //or pause
			path := cl_tasks.ReadArray()
			playerID := cl_tasks.ReadInt()
			playIt := cl_tasks.ReadInt()

			_, tp := getType(string(path))
			if tp > 0 { //audio/video
				item, err := vlc.Add(string(path), playerID)
				if err == nil {
					if playIt > 0 {
						item.Play()
					} else {
						item.Pause()
					}
					cl_tasks.WriteArray(nil)
				} else {
					cl_tasks.WriteArray([]byte(err.Error()))
				}
			} else {
				cl_tasks.WriteArray(fmt.Appendf(nil, "'%s' audio/video format is not supported", string(path)))
			}

		case "seek":
			path := cl_tasks.ReadArray()
			playerID := cl_tasks.ReadInt()
			pos_ms := cl_tasks.ReadInt()

			_, tp := getType(string(path))
			if tp > 0 { //audio/video
				item, err := vlc.Add(string(path), playerID)
				if err == nil {
					item.SetSeek(int64(pos_ms))
					cl_tasks.WriteArray(nil)
				} else {
					cl_tasks.WriteArray([]byte(err.Error()))
				}
			} else {
				cl_tasks.WriteArray(fmt.Appendf(nil, "'%s' audio/video format is not supported", string(path)))
			}

		case "volume":
			path := cl_tasks.ReadArray()
			playerID := cl_tasks.ReadInt()
			volume := cl_tasks.ReadInt() //0-100

			_, tp := getType(string(path))
			if tp > 0 { //audio/video
				item, err := vlc.Add(string(path), playerID)
				if err == nil {
					item.SetVolume(float64(volume) / 100)
					cl_tasks.WriteArray(nil)
				} else {
					cl_tasks.WriteArray([]byte(err.Error()))
				}
			} else {
				cl_tasks.WriteArray(fmt.Appendf(nil, "'%s' audio/video format is not supported", string(path)))
			}

		}
	}
}
