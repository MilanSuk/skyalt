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
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"slices"
	"sync"
	"time"

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
	origSize OsV2

	path        WinImagePath
	loaded_blob []byte

	err error

	texture *WinTexture

	lastDrawTick int64

	file_timestamp int64
}

func NewWinImage(path WinImagePath, win *Win) *WinImage {
	img := &WinImage{path: path, file_timestamp: -1}

	img.path = path

	fnDone := func(bytes []byte, err error) {
		img.loaded_blob = bytes
		img.err = err
		win.SetRedrawNewImage()
	}

	if path.fnGetBlob != nil {
		img.err = path.fnGetBlob(fnDone)
	} else {
		img.err = fmt.Errorf("fnGetBlob is nil")
	}

	if img.path.fnGetTimestamp != nil {
		img.file_timestamp, img.err = img.path.fnGetTimestamp()
	}

	return img
}

func (img *WinImage) FreeTexture() error {
	if img.texture != nil {
		img.texture.Destroy()
	}

	img.texture = nil
	return nil
}

func (img *WinImage) GetBytes() int {
	if img.texture != nil {
		sz := img.texture.size
		return sz.X * sz.Y * 4

	}
	return 0
}

func (img *WinImage) Destroy() {
	if img.texture != nil {
		img.FreeTexture()
	}
}

func (img *WinImage) Maintenance() (bool, error) {
	// free un-used
	if img.texture != nil && !OsIsTicksIn(img.lastDrawTick, 60000) {
		img.FreeTexture()
		return false, nil
	}

	return true, nil
}

func (img *WinImage) Tick() error {

	return nil
}

func (img *WinImage) Draw(coord OsV4, depth int, cd color.RGBA) string {

	if len(img.loaded_blob) > 0 {
		img.texture, _, img.err = InitWinTextureFromBlob(img.loaded_blob)

		if img.texture != nil {
			img.origSize = img.texture.size
		}

		img.loaded_blob = nil //reset
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

	//hot reload
	go func() {
		for {
			img._hotReload()
			time.Sleep(1 * time.Second)
		}
	}()

	return img
}

func (imgs *WinImages) Destroy() {
	imgs.lock.Lock()
	defer imgs.lock.Unlock()

	for _, it := range imgs.images {
		it.Destroy()
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

func (imgs *WinImages) GetImagesBytes() int {
	imgs.lock.Lock()
	defer imgs.lock.Unlock()

	n := 0
	for _, img := range imgs.images {
		n += img.GetBytes()
	}
	return n
}

func (imgs *WinImages) AddImage(path WinImagePath) *WinImage {
	imgs.lock.Lock()
	defer imgs.lock.Unlock()

	//find
	for _, img := range imgs.images {
		if img.path.Cmp(&path) {
			return img
		}
	}

	//add
	img := NewWinImage(path, imgs.win)
	imgs.images = append(imgs.images, img)
	return img
}

func (imgs *WinImages) Maintenance() {
	imgs.lock.Lock()
	defer imgs.lock.Unlock()

	for i := len(imgs.images) - 1; i >= 0; i-- {
		ok, _ := imgs.images[i].Maintenance()
		if !ok {
			imgs.images = slices.Delete(imgs.images, i, i+1)
		}
	}
}

func (imgs *WinImages) _hotReload() {

	if imgs.lock.TryLock() {
		//imgs.lock.Lock()
		defer imgs.lock.Unlock()

		for i := len(imgs.images) - 1; i >= 0; i-- {
			img := imgs.images[i]

			if img.path.fnGetTimestamp != nil {
				timestamp, err := img.path.fnGetTimestamp()
				if err == nil {
					if img.file_timestamp != timestamp {
						imgs.images = slices.Delete(imgs.images, i, i+1)
						imgs.win.SetRedrawNewImage()
					}
				}
			}
		}
	}
}
