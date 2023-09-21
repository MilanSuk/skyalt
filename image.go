/*
Copyright 2023 Milan Suk

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
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"strconv"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"
)

type ResourcePath struct {
	root *Root

	//is db
	db     string
	table  string
	column string
	row    int

	//is resource
	app   string
	asset string
	file  string
}

func InitResourcePath(root *Root, path string, app string) (ResourcePath, error) {
	var ip ResourcePath
	ip.root = root

	var found bool
	path, found = strings.CutPrefix(path, "db:")
	if found {
		//db
		d := strings.Index(path, "/")
		if d <= 0 {
			return ip, errors.New("1st '/' invalid")
		}
		ip.db = path[:d]
		path = path[d+1:]

		//table
		d = strings.Index(path, "/")
		if d <= 0 {
			return ip, errors.New("2nd '/' invalid")
		}
		ip.table = path[:d]
		path = path[d+1:]

		//column
		d = strings.Index(path, "/")
		if d <= 0 {
			return ip, errors.New("3rd '/' invalid")
		}
		ip.column = path[:d]
		path = path[d+1:]

		//row
		var err error
		ip.row, err = strconv.Atoi(path)
		if err != nil {
			return ip, err
		}
	}

	path, found = strings.CutPrefix(path, "asset:")
	if found {
		ip.app = app
		//asset
		d := strings.Index(path, "/")
		if d < 0 {
			return ip, errors.New("1st '/' invalid")
		}
		ip.asset = path[:d]
		path = path[d+1:]

		//name
		ip.file = path
	}

	return ip, nil
}

func (ip *ResourcePath) Is() bool {
	return true
}

func (ip *ResourcePath) GetString() string {
	if len(ip.db) > 0 {
		return fmt.Sprintf("db:%s/%s/%d", ip.table, ip.column, ip.row)

	} else if len(ip.app) > 0 {
		return fmt.Sprintf("asset:%s/%s/%s", ip.app, ip.asset, ip.file)
	}
	return ""
}

func (a *ResourcePath) Cmp(b *ResourcePath) bool {
	return a.db == b.db && a.table == b.table && a.column == b.column && a.row == b.row &&
		a.app == b.app && a.asset == b.asset && a.file == b.file
}

func (ip *ResourcePath) GetBlob() ([]byte, error) {
	var data []byte

	if len(ip.db) > 0 {
		db, err := ip.root.AddDb(ip.db)
		if err != nil {
			return nil, fmt.Errorf("AddDb() failed: %w", err)
		}

		res := db.db.QueryRow("SELECT "+ip.column+" FROM "+ip.table+" WHERE _rowid_ = ?;", ip.row)
		if res == nil {
			return nil, fmt.Errorf("QueryRow() failed")
		}

		err = res.Scan(&data)
		if err != nil {
			return nil, fmt.Errorf("Scan() failed: %w", err)
		}
	} else if len(ip.app) > 0 {

		app := ip.root.FindApp(ip.app, "", -1)
		if app == nil {
			return nil, fmt.Errorf("App(%s) not found", ip.app)
		}

		asset := app.FindAsset(ip.asset)
		if asset == nil {
			return nil, fmt.Errorf("Asset(%s.%s) not found", ip.app, ip.asset)
		}

		path := asset.getResourcesPath() + "/" + ip.file

		var err error
		data, err = os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("ReadFile(%s) failed: %w", path, err)
		}
	}

	return data, nil
}

type Image struct {
	origSize   OsV2
	maxUseSize OsV2

	path ResourcePath

	inverserRGB bool

	texture *sdl.Texture

	lastDrawTick int
}

func (img *Image) GetSize() (OsV2, error) {

	if img.texture != nil {
		_, _, x, y, err := img.texture.Query()
		return OsV2{int(x), int(y)}, err
	}
	return OsV2{}, nil
}

func InitImageGlobal() {
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	image.RegisterFormat("webp", "webp", webp.Decode, webp.DecodeConfig)
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("jpg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("gif", "gif", gif.Decode, gif.DecodeConfig)
	image.RegisterFormat("tiff", "tiff", tiff.Decode, tiff.DecodeConfig)
	image.RegisterFormat("bmp", "bmp", bmp.Decode, bmp.DecodeConfig)
}

func CreateTextureFromImage(img image.Image, inverserRGB bool, render *sdl.Renderer) (*sdl.Texture, OsV2, error) {

	W := img.Bounds().Max.X
	H := img.Bounds().Max.Y

	texture, err := render.CreateTexture(sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_STREAMING, int32(W), int32(H))
	if err != nil {
		return nil, OsV2{}, fmt.Errorf("CreateTexture() failed: %w", err)
	}
	texture.SetBlendMode(sdl.BLENDMODE_BLEND)

	pixels, _, err := texture.Lock(nil)
	if err != nil {
		return nil, OsV2{}, fmt.Errorf("texture Lock() failed: %w", err)
	}

	stride := W * 4
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			r, g, b, a := img.At(int(x), int(y)).RGBA()

			pixels[y*stride+x*4+0] = byte(b >> 8) //blue is 1st!
			pixels[y*stride+x*4+1] = byte(g >> 8)
			pixels[y*stride+x*4+2] = byte(r >> 8) //red is last!

			pixels[y*stride+x*4+3] = byte(a >> 8)
		}
	}

	if inverserRGB {
		for i := 0; i < len(pixels); i++ {
			if i%4 != 3 { //skip alpha channel
				pixels[i] = 255 - pixels[i]
			}
		}
	}

	//copy(pixels, surf.Pixels()) //, surf.Pitch*surf.H)
	texture.Unlock()

	return texture, OsV2{W, H}, nil
}

func Image_LoadTexture(blob []byte, inverserRGB bool, render *sdl.Renderer) (*sdl.Texture, error) {

	img, _, err := image.Decode(bytes.NewReader(blob))
	if err != nil {
		return nil, fmt.Errorf("Decode() failed: %w", err)
	}

	texture, _, err := CreateTextureFromImage(img, inverserRGB, render)
	if err != nil {
		return nil, fmt.Errorf("CreateTextureFromImage() failed: %w", err)
	}

	return texture, nil
}

func NewImage(path ResourcePath, inverserRGB bool, render *sdl.Renderer) (*Image, error) {

	var self Image

	self.path = path
	self.inverserRGB = inverserRGB

	var err error
	blob, err := path.GetBlob()
	if err != nil {
		return nil, fmt.Errorf("GetBlob() failed: %w", err)
	}
	if len(blob) == 0 {
		return nil, nil //empty = no error
	}

	self.texture, err = Image_LoadTexture(blob, inverserRGB, render)
	if err != nil {
		return nil, err
	}

	self.origSize, err = self.GetSize()
	if err != nil {
		return nil, fmt.Errorf("GetSize() failed: %w", err)
	}

	return &self, nil
}

func (img *Image) FreeTexture() error {
	if img.texture != nil {
		if err := img.texture.Destroy(); err != nil {
			return fmt.Errorf("Destroy() failed: %w", err)
		}
	}

	img.texture = nil
	return nil
}

func (img *Image) GetBytes() int64 {
	if img.texture != nil {
		sz, err := img.GetSize()
		if err == nil {
			return int64(sz.X * sz.Y * 4)
		}
	}
	return 0
}

func (img *Image) Destroy() error {
	return img.FreeTexture()
}

func (img *Image) Maintenance(render *sdl.Renderer) (bool, error) {

	if !img.maxUseSize.Is() && !OsIsTicksIn(img.lastDrawTick, 10000) {
		// free un-used
		if img.texture != nil && !OsIsTicksIn(img.lastDrawTick, 10000) {
			img.FreeTexture()
		}
		return false, nil
	}

	img.maxUseSize = OsV2{0, 0} // reset

	return true, nil
}

func (img *Image) Draw(coord OsV4, cd OsCd, render *sdl.Renderer) error {

	img.maxUseSize = coord.Size.Max(img.maxUseSize)

	if img.texture != nil {
		err := img.texture.SetColorMod(cd.R, cd.G, cd.B)
		if err != nil {
			return fmt.Errorf("Image.Draw() SetColorMod() failed: %w", err)
		}

		err = img.texture.SetAlphaMod(cd.A)
		if err != nil {
			return fmt.Errorf("Image.Draw() SetAlphaMod() failed: %w", err)
		}

		err = render.Copy(img.texture, nil, coord.GetSDLRect())
		if err != nil {
			return fmt.Errorf("Image.Draw() RenderCopy() failed: %w", err)
		}
	}

	img.lastDrawTick = OsTicks()
	return nil
}
