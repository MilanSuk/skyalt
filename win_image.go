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
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"slices"
	"sync"
	"sync/atomic"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"
)

func InitImageGlobal() {
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	image.RegisterFormat("webp", "webp", webp.Decode, webp.DecodeConfig)
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("jpg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("gif", "gif", gif.Decode, gif.DecodeConfig)
	image.RegisterFormat("tiff", "tiff", tiff.Decode, tiff.DecodeConfig)
	image.RegisterFormat("bmp", "bmp", bmp.Decode, bmp.DecodeConfig)
}

type WinImage struct {
	path WinImagePath

	loaded_lock sync.Mutex
	loaded_rgba []byte
	loaded_size OsV2

	num_loads atomic.Uint64

	err error

	texture *WinTexture

	lastDrawTick int64
}

func NewWinImage(path WinImagePath, win *Win, inited func()) *WinImage {
	img := &WinImage{path: path}

	img.path = path

	go img._loadFromMedia(win, inited)

	return img
}

func (img *WinImage) FreeTexture(win *Win) {
	if img.texture != nil {
		img.texture.Destroy()
	}
	img.texture = nil
}

func (img *WinImage) _loadFromMedia(win *Win, inited func()) {
	defer img.num_loads.Add(1)

	if inited != nil {
		defer inited()
	}

	w, h, rgba, play_pos, play_duration, tp, err := win.services.media.Frame(img.path.path, img.path.blob, img.path.playerID, img.err == nil)

	img.loaded_lock.Lock()
	defer img.loaded_lock.Unlock()

	img.err = err
	if err != nil {
		return
	}

	img.path.play_pos = play_pos
	img.path.play_duration = play_duration
	img.path.tp = tp
	img.loaded_size = OsV2{w, h}
	img.loaded_rgba = rgba

	win.SetRedrawNewImage()

	//fmt.Println(img.path.path, play_pos, play_duration)
	//fmt.Println("_loadFromMedia()", time.Now().UnixMilli())
}

func (img *WinImage) SetPlay(playIt bool, win *Win) {
	img.path.is_playing = playIt
	win.services.media.Play(img.path.path, img.path.playerID, playIt)
}

func (img *WinImage) SetSeek(pos_ms uint64, win *Win) {
	img.path.play_pos = pos_ms
	win.services.media.Seek(img.path.path, img.path.playerID, pos_ms)
}

func (img *WinImage) GetBytes() int {
	if img.texture != nil {
		sz := img.texture.size
		return sz.X * sz.Y * 4

	}
	return 0
}

func (img *WinImage) Destroy(win *Win) {
	img.FreeTexture(win)
}

func (img *WinImage) Maintenance(win *Win) (bool, error) {
	// free un-used
	if img.texture != nil && !OsIsTicksIn(img.lastDrawTick, 60000) {
		img.FreeTexture(win)

		return false, nil
	}

	return true, nil
}

func (img *WinImage) Tick() error {

	return nil
}

func (img *WinImage) Draw(coord OsV4, depth int, cd color.RGBA, win *Win) string {
	if len(img.loaded_rgba) > 0 {
		img.loaded_lock.Lock()
		defer img.loaded_lock.Unlock()

		if img.texture == nil || !img.texture.size.Cmp(img.loaded_size) {
			//new
			img.texture = InitWinTextureFromImageRGBAPix(img.loaded_rgba, img.loaded_size)
		} else {
			//update
			//fmt.Println("UpdateContent()", time.Now().UnixMilli())
			img.texture.UpdateContent(img.loaded_rgba)
		}

		img.loaded_rgba = nil //reset
	}

	if coord.Size.Is() {
		img.lastDrawTick = OsTicks()
	}

	if img.texture != nil {
		img.texture.DrawQuad(coord, depth, cd)
	} else {
		if img.err != nil {
			return img.err.Error()
		}

		return "loading ..."
	}

	return ""
}

type WinImages struct {
	win    *Win
	images []*WinImage
	lock   sync.Mutex
}

func NewWinImages(win *Win) *WinImages {
	img := &WinImages{win: win}
	return img
}

func (imgs *WinImages) Destroy(win *Win) {
	imgs.lock.Lock()
	defer imgs.lock.Unlock()

	for _, it := range imgs.images {
		it.Destroy(win)
	}
}

func (imgs *WinImages) NumTextures() int {
	imgs.lock.Lock()
	defer imgs.lock.Unlock()

	n := 0
	for _, img := range imgs.images {
		if img.texture != nil {
			n++
		}
	}
	return n
}

func (imgs *WinImages) GetBytes() int {
	imgs.lock.Lock()
	defer imgs.lock.Unlock()

	n := 0
	for _, img := range imgs.images {
		n += img.GetBytes()
	}
	return n
}

func (imgs *WinImages) Add(path WinImagePath, inited func()) *WinImage {
	imgs.lock.Lock()
	defer imgs.lock.Unlock()

	//find
	for _, img := range imgs.images {
		if img.path.Cmp(&path) {
			return img
		}
	}

	//add
	img := NewWinImage(path, imgs.win, inited)
	imgs.images = append(imgs.images, img)
	return img
}

func (imgs *WinImages) Maintenance(win *Win) {
	imgs.lock.Lock()
	defer imgs.lock.Unlock()

	for i := len(imgs.images) - 1; i >= 0; i-- {
		ok, _ := imgs.images[i].Maintenance(win)
		if !ok {
			imgs.images = slices.Delete(imgs.images, i, i+1)
		}
	}
}

func (imgs *WinImages) Tick(win *Win) {
	imgs.lock.Lock()
	defer imgs.lock.Unlock()

	redraw := false

	//video changed || audio is playing(get seek pos)
	changed := imgs.win.services.media.GetChanged()
	for _, it := range changed {
		for _, img := range imgs.images {
			if (it.Type == 0 && img.path.path == it.path) || (it.Type == 1 && img.path.playerID == it.playerID) {
				img._loadFromMedia(win, nil)
				//fmt.Println("------------", it, time.Now().UnixMilli())
				redraw = true
				break //found
			}
		}
	}

	if redraw {
		win.SetRedrawNewImage()
	}
}
