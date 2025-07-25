package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

func getType(path string) int {
	ext := strings.ToLower(filepath.Ext(path))

	img := (ext == ".png" || ext == ".webp" || ext == ".jpeg" || ext == ".jpg" || ext == ".gif" || ext == ".tiff" || ext == ".bmp")
	vid := (ext == ".mp4" || ext == ".mkv" || ext == ".webm" || ext == ".mov" || ext == ".avi" || ext == ".flv")
	aud := (ext == ".wav" || ext == ".mp3" || ext == ".opus" || ext == ".aac" || ext == ".ogg" || ext == ".flac" || ext == ".pcm")

	if !img && !vid && !aud {
		img = true //default = try to open as image
	}

	if vid {
		return 1
	}
	if aud {
		return 2
	}

	return 0 //default = try to open as image
}

func main() {
	log.SetFlags(log.Llongfile) //log.LstdFlags | log.Lshortfile

	vlc := NewVLC()
	defer vlc.Destroy()

	imgs := NewImages()
	defer imgs.Destroy()

	/*
		{
			//Tests

			av, _ := vlc.Add("../vid.mkv", 0)
			//vlc.Add("../aud.mp3", 1)
			//imgs.Add("../resources/logo_small.png", nil)

			//av.Play()
			av.SetSeek(10000)

			for {
				imgs.UpdateFileTimes()
				vlc.UpdateFileTimes()

				min_time := time.Now().Add(-60 * time.Second).UnixNano()

				imgs.Maintenance(min_time)
				vlc.Maintenance(min_time)

				//fmt.Println(av.GetSeek(), av.GetDuration(), av.IsPlaying())

				p, a := vlc.Check("../vid.mkv", 0)
				//p2, a2 := vlc.Check("../aud.mp3", 1)
				if p == false {
					//av.Play() //repeat
				}

				//b := imgs.Check("../resources/logo_small.png")
				fmt.Println(p, a)
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

	cl := NewNetClient("localhost", server_port)
	defer cl.Destroy()

	var lock sync.Mutex

	go func() {
		for {
			//update file_times outside lock
			imgs.UpdateFileTimes()
			vlc.UpdateFileTimes()

			lock.Lock()
			{
				min_time := time.Now().Add(-60 * time.Second).UnixNano()

				imgs.Maintenance(min_time)
				vlc.Maintenance(min_time)
			}
			lock.Unlock()

			time.Sleep(10 * time.Second)
		}
	}()

	for {
		command := cl.ReadArray()
		lock.Lock()

		switch string(command) {
		case "exit":
			return

		case "check":
			path := cl.ReadArray()
			playerID := cl.ReadInt() //for audio/video only

			var playing bool
			diff := false
			tp := getType(string(path))
			if tp > 0 { //audio/video
				playing, diff = vlc.Check(string(path), playerID)
			} else {
				diff = imgs.Check(string(path))
			}

			cl.WriteBool(playing)
			cl.WriteBool(diff)

		case "type":
			path := cl.ReadArray() //must be set!

			tp := getType(string(path))
			cl.WriteInt(uint64(tp))

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
			cl.WriteArray(infoJs)

		case "frame":
			path := cl.ReadArray()   //must be set!
			blob := cl.ReadArray()   //for images only
			playerID := cl.ReadInt() //for audio/video only

			var errBytes []byte
			var width uint64
			var height uint64
			var data []byte
			var seek_ms uint64
			var play_duration uint64

			tp := getType(string(path))
			if tp > 0 { //audio/video
				item, err := vlc.Add(string(path), playerID)
				if err == nil {
					item.GetSeek()

					width = uint64(item.width)
					height = uint64(item.height)
					seek_ms = uint64(item.GetSeek())
					play_duration = uint64(item.GetDuration())
					data = unsafe.Slice((*byte)(item.videoCtx.pixels), item.pixels_size)
				} else {
					errBytes = []byte(err.Error())
				}
			} else { //image
				item, err := imgs.Add(string(path), blob)
				if err == nil {
					width = uint64(item.width)
					height = uint64(item.height)
					data = item.data
				} else {
					errBytes = []byte(err.Error())
				}
			}

			cl.WriteArray(errBytes)
			cl.WriteInt(width)
			cl.WriteInt(height)
			cl.WriteArray(data)
			cl.WriteInt(seek_ms)
			cl.WriteInt(play_duration)
			cl.WriteInt(uint64(tp))

		case "play": //or pause
			path := cl.ReadArray()
			playerID := cl.ReadInt()
			playIt := cl.ReadInt()

			tp := getType(string(path))
			if tp > 0 { //audio/video
				item, err := vlc.Add(string(path), playerID)
				if err == nil {
					if playIt > 0 {
						item.Play()
					} else {
						item.Pause()
					}
					cl.WriteArray(nil)
				} else {
					cl.WriteArray([]byte(err.Error()))
				}
			} else {
				cl.WriteArray(fmt.Appendf(nil, "'%s' audio/video format is not supported", string(path)))
			}

		case "seek":
			path := cl.ReadArray()
			playerID := cl.ReadInt()
			pos_ms := cl.ReadInt()

			tp := getType(string(path))
			if tp > 0 { //audio/video
				item, err := vlc.Add(string(path), playerID)
				if err == nil {
					item.SetSeek(int64(pos_ms))
					cl.WriteArray(nil)
				} else {
					cl.WriteArray([]byte(err.Error()))
				}
			} else {
				cl.WriteArray(fmt.Appendf(nil, "'%s' audio/video format is not supported", string(path)))
			}

		case "volume":
			path := cl.ReadArray()
			playerID := cl.ReadInt()
			volume := cl.ReadInt() //0-100

			tp := getType(string(path))
			if tp > 0 { //audio/video
				item, err := vlc.Add(string(path), playerID)
				if err == nil {
					item.SetVolume(float64(volume) / 100)
					cl.WriteArray(nil)
				} else {
					cl.WriteArray([]byte(err.Error()))
				}
			} else {
				cl.WriteArray(fmt.Appendf(nil, "'%s' audio/video format is not supported", string(path)))
			}

		}
		lock.Unlock()
	}
}
