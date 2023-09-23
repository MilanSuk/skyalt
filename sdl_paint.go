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
	"fmt"

	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
)

type PaintBuffTexture struct {
	texture *sdl.Texture
	size    OsV2
}

func NewPaintBuffTexture(size OsV2, render *sdl.Renderer) (*PaintBuffTexture, error) {
	var tex PaintBuffTexture

	var err error
	tex.texture, err = render.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, int32(size.X), int32(size.Y))
	if err != nil {
		return nil, fmt.Errorf("CreateTexture() failed: %w", err)
	}

	tex.size = size

	return &tex, nil
}

func (tex *PaintBuffTexture) Destroy() {
	if tex.texture != nil {
		tex.texture.Destroy() //err
	}
}

func (tex *PaintBuffTexture) SetRenderTarget(render *sdl.Renderer) (*sdl.Texture, error) {

	old_texture := render.GetRenderTarget()
	err := render.SetRenderTarget(tex.texture)
	if err != nil {
		return old_texture, fmt.Errorf("SetRenderTarget() failed: %w", err)
	}

	return old_texture, nil
}

func (tex *PaintBuffTexture) PrepareAlpha(render *sdl.Renderer) error {

	tex.texture.SetBlendMode(sdl.BLENDMODE_BLEND)
	err := render.SetDrawColor(0, 0, 0, 0) //full alpha
	if err != nil {
		return fmt.Errorf("SetDrawColor() failed: %w", err)
	}

	err = render.Clear()
	if err != nil {
		return fmt.Errorf("Clear() failed: %w", err)
	}
	return nil
}

func (tex *PaintBuffTexture) Copy(coord OsV4, render *sdl.Renderer) error {
	err := render.Copy(tex.texture, nil, coord.GetSDLRect())
	if err != nil {
		return fmt.Errorf("Copy() failed: %w", err)
	}
	return nil
}

type PaintBuffText struct {
	texture *PaintBuffTexture
	text    string
	font    *Font
	height  int
}

func NewPaintBuffText(text string, height int, font *Font, render *sdl.Renderer) (*PaintBuffText, error) {
	var bt PaintBuffText

	sz, err := font.GetTextSize(text, height, height)
	if err != nil {
		return nil, fmt.Errorf("GetTextSize() failed: %w", err)
	}
	{
		down_y, err := font.GetDownY(text, height, render)
		if err != nil {
			return nil, fmt.Errorf("GetTextSize() failed: %w", err)
		}
		sz.Y += down_y
	}

	bt.texture, err = NewPaintBuffTexture(sz, render)
	if err != nil {
		return nil, fmt.Errorf("NewPaintBuffTexture() failed: %w", err)
	}

	pre, err := bt.texture.SetRenderTarget(render)
	if err != nil {
		return nil, fmt.Errorf("SetRenderTarget() failed: %w", err)
	}

	err = bt.texture.PrepareAlpha(render)
	if err != nil {
		return nil, fmt.Errorf("PrepareAlpha() failed: %w", err)
	}

	err = font.Print(text, height, OsV4{OsV2{}, sz}, OsV2{0, 0}, OsCd_white(), nil, false, render)
	if err != nil {
		return nil, fmt.Errorf("Print() failed: %w", err)
	}

	render.SetRenderTarget(pre) //backup

	bt.text = text
	bt.font = font
	bt.height = height

	return &bt, nil
}

func (bt *PaintBuffText) Destroy() {
	if bt.texture != nil {
		bt.texture.Destroy() //err
	}
}

func (bt *PaintBuffText) Copy(coord OsV4, align OsV2, cd OsCd, render *sdl.Renderer) error {

	q := OsV4{coord.Start, bt.texture.size}

	//align X
	switch align.X {
	case 1:
		q.Start.X += (coord.Size.X - bt.texture.size.X) / 2

	case 2:
		q.Start.X = coord.End().X - bt.texture.size.X
	}

	//align Y
	switch align.Y {
	case 1:
		q.Start.Y += (coord.Size.Y - bt.texture.size.Y) / 2

	case 2:
		q.Start.Y = coord.End().Y - bt.texture.size.Y
	}

	bt.texture.texture.SetColorMod(cd.R, cd.G, cd.B)
	bt.texture.texture.SetAlphaMod(cd.A)
	err := bt.texture.Copy(q, render)
	if err != nil {
		return fmt.Errorf("Copy() failed: %w", err)
	}

	return nil

}

type PaintBuffTextCache struct {
	textures []*PaintBuffText
}

func NewPaintBuffTextCache() *PaintBuffTextCache {
	var btc PaintBuffTextCache
	return &btc
}

func (btc *PaintBuffTextCache) Destroy() {
	for _, it := range btc.textures {
		it.Destroy()
	}
}

func (btc *PaintBuffTextCache) Find(text string, height int, font *Font, cd OsCd) *PaintBuffText {
	for _, it := range btc.textures {
		if it.text == text && it.height == height && it.font == font {
			return it
		}
	}
	return nil
}

func (btc *PaintBuffTextCache) Draw(text string, height int, font *Font, cd OsCd, coord OsV4, align OsV2, render *sdl.Renderer) error {
	var err error

	if len(text) == 0 || height <= 0 || font == nil {
		return nil
	}

	it := btc.Find(text, height, font, cd)
	if it == nil {
		//add
		it, err = NewPaintBuffText(text, height, font, render)
		if err != nil {
			return fmt.Errorf("NewPaintBuffText() failed: %w", err)
		}

		btc.textures = append(btc.textures, it)
	}

	err = it.Copy(coord, align, cd, render)
	if err != nil {
		return fmt.Errorf("Copy() failed: %w", err)
	}

	return nil
}

type PaintBuff struct {
	ui       *Ui
	lastCrop OsV4

	texture        *PaintBuffTexture
	backup_texture *sdl.Texture

	text_cache *PaintBuffTextCache
}

func NewPaintBuff(ui *Ui) *PaintBuff {
	var b PaintBuff
	b.ui = ui
	b.text_cache = NewPaintBuffTextCache()
	return &b
}

func (b *PaintBuff) Destroy() {
	if b.texture != nil {
		b.texture.Destroy()
	}
	b.text_cache.Destroy()
}

func (b *PaintBuff) StartLevel(coord OsV4) error {

	var err error

	tex_size, _ := b.ui.GetScreenCoord()

	//cmp
	if b.texture == nil {
		b.texture, err = NewPaintBuffTexture(tex_size.Size, b.ui.render)
		if err != nil {
			return fmt.Errorf("NewPaintBuffTexture() failed: %w", err)
		}

	} else if !b.texture.size.Cmp(tex_size.Size) {
		b.texture.Destroy()
		b.texture, err = NewPaintBuffTexture(tex_size.Size, b.ui.render)
		if err != nil {
			return fmt.Errorf("NewPaintBuffTexture() failed: %w", err)
		}
	}

	//set
	b.backup_texture, err = b.texture.SetRenderTarget(b.ui.render)
	if err != nil {
		return fmt.Errorf("SetRenderTarget() failed: %w", err)
	}

	err = b.texture.PrepareAlpha(b.ui.render)
	if err != nil {
		return fmt.Errorf("PrepareAlpha() failed: %w", err)
	}

	b.Reset(coord) //background

	return nil
}

func (b *PaintBuff) EndLevel() error {
	err := b.ui.render.SetRenderTarget(b.backup_texture)
	if err != nil {
		return fmt.Errorf("SetRenderTarget() failed: %w", err)
	}

	return nil
}

func (b *PaintBuff) Draw() error {
	err := b.ui.render.SetRenderTarget(nil)
	if err != nil {
		return fmt.Errorf("CreateTexture() failed: %w", err)
	}

	sz, err := b.ui.GetScreenCoord()
	if err != nil {
		return fmt.Errorf("GetScreenCoord() failed: %w", err)
	}
	err = b.ui.render.SetClipRect(sz.GetSDLRect())
	if err != nil {
		return fmt.Errorf("SetClipRect() failed: %w", err)
	}

	if b.texture != nil {
		//fade
		_Ui_boxSE(b.ui.render, sz.Start, sz.End(), OsCd{0, 0, 0, 80})

		//copy texture
		b.texture.Copy(sz, b.ui.render)
	}

	return nil
}

func (b *PaintBuff) Reset(crop OsV4) {
	b.lastCrop = crop

	//background
	b.AddCrop(crop)
	b.AddRect(crop, OsCd_white(), 0)
}

func (b *PaintBuff) AddCrop(coord OsV4) OsV4 {
	err := b.ui.render.SetClipRect(coord.GetSDLRect())
	if err != nil {
		return OsV4{}
	}

	old := b.lastCrop
	b.lastCrop = coord
	return old
}

func (b *PaintBuff) AddRect(coord OsV4, cd OsCd, thick int) {
	start := coord.Start
	end := coord.End()
	if thick == 0 {
		_Ui_boxSE(b.ui.render, start, end, cd)
	} else {
		_Ui_boxSE_border(b.ui.render, start, end, cd, thick)
	}

}

func (b *PaintBuff) AddLine(start OsV2, end OsV2, cd OsCd, thick int) {
	v := end.Sub(start)
	if !v.IsZero() {
		_Ui_line(b.ui.render, start, end, thick, cd)
	}
}

func (b *PaintBuff) AddCircle(coord OsV4, cd OsCd, thick int) {
	p := coord.Middle()
	if thick == 0 {
		gfx.FilledEllipseRGBA(b.ui.render, int32(p.X), int32(p.Y), int32(coord.Size.X/2), int32(coord.Size.Y/2), cd.R, cd.G, cd.B, cd.A)
		gfx.AAEllipseRGBA(b.ui.render, int32(p.X), int32(p.Y), int32(coord.Size.X/2), int32(coord.Size.Y/2), cd.R, cd.G, cd.B, cd.A)
	} else {
		gfx.AAEllipseRGBA(b.ui.render, int32(p.X), int32(p.Y), int32(coord.Size.X/2), int32(coord.Size.Y/2), cd.R, cd.G, cd.B, cd.A)
	}
}

func PaintImage_load(path ResourcePath, inverserRGB bool, ui *Ui) (*Image, error) {
	var img *Image
	for _, it := range ui.images {
		if it.path.Cmp(&path) && it.inverserRGB == inverserRGB {
			img = it
			break
		}
	}

	if img == nil {
		var err error
		img, err = NewImage(path, inverserRGB, ui.render)
		if err != nil {
			return nil, fmt.Errorf("NewImage() failed: %w", err)
		}

		if img != nil {
			ui.images = append(ui.images, img)
		}
	}

	return img, nil
}

func (b *PaintBuff) AddImage(path ResourcePath, inverserRGB bool, coord OsV4, cd OsCd, alignV int, alignH int, fill bool) {
	img, err := PaintImage_load(path, inverserRGB, b.ui)
	if err != nil {
		b.AddText(path.GetString()+" has error", coord, path.root.fonts.Get(SKYALT_FONT_PATH), OsCd_error(), path.root.ui.io.GetDPI()/8, OsV2{1, 1}, nil)
		return
	}
	if img == nil {
		return //image is empty
	}

	var q OsV4
	if !fill {
		rect_size := OsV2_InRatio(coord.Size, img.origSize)
		q = OsV4_center(coord, rect_size)
	} else {
		q.Start = coord.Start
		q.Size = OsV2_OutRatio(coord.Size, img.origSize)
	}

	if alignH == 0 {
		q.Start.X = coord.Start.X
	} else if alignH == 1 {
		q.Start.X = OsV4_centerFull(coord, q.Size).Start.X
	} else if alignH == 2 {
		q.Start.X = coord.End().X - q.Size.X
	}

	if alignV == 0 {
		q.Start.Y = coord.Start.Y
	} else if alignV == 1 {
		q.Start.Y = OsV4_centerFull(coord, q.Size).Start.Y
	} else if alignV == 2 {
		q.Start.Y = coord.End().Y - q.Size.Y
	}

	imgRectBackup := b.AddCrop(b.lastCrop.GetIntersect(coord))

	if img != nil {
		err := img.Draw(q, cd, b.ui.render)
		if err != nil {
			fmt.Printf("Draw() failed: %v\n", err)
		}
	}

	b.AddCrop(imgRectBackup)
}

func (b *PaintBuff) AddText(text string, coord OsV4, font *Font, cd OsCd, h int, align OsV2, cds []OsCd) {

	//cached
	/*err := b.text_cache.Draw(text, h, font, cd, coord, align, b.ui.render)
	if err != nil {
		fmt.Printf("Draw() failed: %v\n", err)
	}*/

	//no caching
	err := font.Print(text, h, coord, align, cd, cds, true, b.ui.render)
	if err != nil {
		fmt.Printf("Print() failed: %v\n", err)
	}
}

func (b *PaintBuff) AddTextBack(rangee OsV2, text string, coord OsV4, font *Font, cd OsCd, h int, align OsV2, underline bool, addSpaceY bool) error {
	if rangee.X == rangee.Y {
		return nil
	}

	start, err := font.Start(text, h, coord, align, nil)
	if err != nil {
		return fmt.Errorf("Start() failed: %w", err)
	}

	var rng OsV2
	rng.X, err = font.GetPxPos(text, h, rangee.X)
	if err != nil {
		return fmt.Errorf("GetPxPos(1) failed: %w", err)
	}
	rng.Y, err = font.GetPxPos(text, h, rangee.Y)
	if err != nil {
		return fmt.Errorf("GetPxPos(2) failed: %w", err)
	}
	rng.Sort()

	if rng.X != rng.Y {
		if underline {
			Y := coord.Start.Y + coord.Size.Y
			b.AddRect(OsV4{Start: OsV2{start.X + rng.X, Y - 2}, Size: OsV2{rng.Y, 2}}, cd, 0)
		} else {
			c := InitOsQuad(start.X+rng.X, coord.Start.Y, rng.Y-rng.X, coord.Size.Y)
			if addSpaceY {
				c = c.AddSpaceY((coord.Size.Y - h) / 4) //smaller height
			}
			b.AddRect(c, cd, 0)
		}
	}
	return nil
}

func (b *PaintBuff) AddTextCursor(text string, coord OsV4, font *Font, cd OsCd, h int, align OsV2, cursorPos int, cell int) (OsV4, error) {
	b.ui.cursorEdit = true
	cd.A = b.ui.cursorCdA

	start, err := font.Start(text, h, coord, align, nil)
	if err != nil {
		return OsV4{}, fmt.Errorf("TextCursor().Start() failed: %w", err)
	}

	ex, err := font.GetPxPos(text, h, cursorPos)
	if err != nil {
		return OsV4{}, fmt.Errorf("TextCursor().GetPxPos() failed: %w", err)
	}

	cursorQuad := InitOsQuad(start.X+ex, coord.Start.Y, OsMax(1, cell/15), coord.Size.Y)
	cursorQuad = cursorQuad.AddSpaceY((coord.Size.Y - h) / 4) //smaller height

	b.AddRect(cursorQuad, cd, 0)

	return cursorQuad, nil
}
