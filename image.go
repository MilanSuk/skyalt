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

func FileParseUrl(url string, app *App) (string, string, bool, error) {

	file, found := strings.CutPrefix(url, "dbs:")
	if found {
		d := strings.Index(file, ":")
		if d >= 0 {
			if len(file[:d]) == 0 {
				//empty db
				return app.db.path, file[d+1:], false, nil //optional(table/column/row)
			}
			return app.db.root.folderDatabases + "/" + file[:d], file[d+1:], false, nil //optional(table/column/row)
		}

		if len(file) == 0 {
			//empty db
			return app.db.path, "", false, nil
		}
		return app.db.root.folderDatabases + "/" + file, "", false, nil
	}

	file, found = strings.CutPrefix(url, "app:")
	if found {
		d := strings.Index(file, ":")
		if d >= 0 {
			return app.getPath() + "/" + file[:d], file[d+1:], true, nil //optional(table/column/row)
		}
		return app.getPath() + "/" + file, "", true, nil
	}

	return "", "", false, fmt.Errorf("must start with 'dbs:' or 'app:'")
}

type MediaPath struct {
	root *Root

	path string

	//optional(blob)
	table  string
	column string
	row    int
}

func MediaParseUrl(url string, app *App) (MediaPath, error) {
	var ip MediaPath
	ip.root = app.db.root

	filePath, opt, _, err := FileParseUrl(url, app)
	if err != nil {
		return MediaPath{}, fmt.Errorf("DbParseUrl() failed: %w", err)
	}
	ip.path = filePath

	if len(opt) > 0 {
		//table
		d := strings.Index(opt, "/")
		if d <= 0 {
			return ip, errors.New("table '/' is missing")
		}
		ip.table = opt[:d]
		opt = opt[d+1:]

		//column
		d = strings.Index(opt, "/")
		if d <= 0 {
			return ip, errors.New("column '/' is missing")
		}
		ip.column = opt[:d]
		opt = opt[d+1:]

		//row
		var err error
		ip.row, err = strconv.Atoi(opt)
		if err != nil {
			return ip, err
		}
	}

	return ip, nil
}

func (ip *MediaPath) IsDb() bool {
	return len(ip.table) > 0
}
func (ip *MediaPath) IsFile() bool {
	return !ip.IsDb()
}

func (ip *MediaPath) GetString() string {
	return fmt.Sprintf("%s - %s/%s/%d", ip.path, ip.table, ip.column, ip.row)
}

func (a *MediaPath) Cmp(b *MediaPath) bool {
	return a.path == b.path && a.table == b.table && a.column == b.column && a.row == b.row
}

func (ip *MediaPath) GetBlob() ([]byte, *Db, error) {
	var data []byte
	var err error

	var db *Db
	if ip.IsDb() {
		//db blob
		db, err = ip.root.AddDb(ip.path)
		if err != nil {
			return nil, nil, fmt.Errorf("AddDb() failed: %w", err)
		}

		row := db.db.QueryRow("SELECT "+ip.column+" FROM "+ip.table+" WHERE _rowid_ = ?;", ip.row)
		if row == nil {
			return nil, nil, fmt.Errorf("QueryRow() failed")
		}

		err = row.Scan(&data)
		if err != nil {
			return nil, nil, fmt.Errorf("Scan() failed: %w", err)
		}
	} else {
		//file
		data, err = os.ReadFile(ip.path)
		if err != nil {
			return nil, nil, fmt.Errorf("ReadFile(%s) failed: %w", ip.path, err)
		}
	}

	return data, db, nil
}

type Image struct {
	origSize   OsV2
	maxUseSize OsV2

	path              MediaPath
	db                *Db
	dbWrite_loadTicks int64
	blobHash          OsHash

	texture *sdl.Texture

	lastDrawTick int64
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

func CreateTextureFromImage(img image.Image, render *sdl.Renderer) (*sdl.Texture, OsV2, error) {

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

	/*if inverserRGB {
		for i := 0; i < len(pixels); i++ {
			if i%4 != 3 { //skip alpha channel
				pixels[i] = 255 - pixels[i]
			}
		}
	}*/

	//copy(pixels, surf.Pixels()) //, surf.Pitch*surf.H)
	texture.Unlock()

	return texture, OsV2{W, H}, nil
}

func Image_LoadTexture(blob []byte, render *sdl.Renderer) (*sdl.Texture, error) {

	img, _, err := image.Decode(bytes.NewReader(blob))
	if err != nil {
		return nil, fmt.Errorf("Decode() failed: %w", err)
	}

	texture, _, err := CreateTextureFromImage(img, render)
	if err != nil {
		return nil, fmt.Errorf("CreateTextureFromImage() failed: %w", err)
	}

	return texture, nil
}

func NewImage(path MediaPath, render *sdl.Renderer) (*Image, error) {

	var img Image

	img.path = path

	var err error
	var blob []byte
	blob, img.db, err = path.GetBlob()
	if err != nil {
		return nil, fmt.Errorf("GetBlob() failed: %w", err)
	}
	if len(blob) == 0 {
		return nil, nil //empty = no error
	}

	if img.db != nil {
		img.dbWrite_loadTicks = img.db.lastWriteTicks
	}

	img.blobHash, err = InitOsHash(blob)
	if err != nil {
		return nil, fmt.Errorf("InitOsHash() failed: %w", err)
	}

	img.texture, err = Image_LoadTexture(blob, render)
	if err != nil {
		return nil, fmt.Errorf("Image_LoadTexture() failed: %w", err)
	}

	img.origSize, err = img.GetSize()
	if err != nil {
		return nil, fmt.Errorf("GetSize() failed: %w", err)
	}

	return &img, nil
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

	//is db changed
	if img.db != nil {
		if img.dbWrite_loadTicks < img.db.lastWriteTicks {
			blob, _, err := img.path.GetBlob()
			if err == nil {
				blobHash, err := InitOsHash(blob)
				if err == nil {
					img.dbWrite_loadTicks = img.db.lastWriteTicks

					if !img.blobHash.Cmp(&blobHash) {
						return false, nil //remove & later reload it
					}
				}
			}
		}
	}

	//is used
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
