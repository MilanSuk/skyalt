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
}

func NewWinImage(path WinImagePath, win *Win) *WinImage {
	img := &WinImage{path: path}

	img.path = path

	fnDone := func(bytes []byte, err error) {
		img.loaded_blob = bytes
		img.err = err
		win.SetRedrawNewImage()
	}

	img.err = path.GetBlob(fnDone)

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

func (img *WinImage) Destroy() error {
	return img.FreeTexture()
}

func (img *WinImage) Maintenance(win *Win) (bool, error) {
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
