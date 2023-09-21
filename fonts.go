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
	"strings"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const SKYALT_FONT_0 = "resources/arial.ttf"
const SKYALT_FONT_1 = "resources/consola.ttf"

const SKYALT_FONT_TAB_WIDTH = 4

type FontLetter struct {
	texture *sdl.Texture
	x       int
	y       int
	len     int
}

func InitFontLetter(ch rune, font *ttf.Font, render *sdl.Renderer) (*FontLetter, error) {
	var self FontLetter

	tab := (ch == '\t')
	if tab {
		ch = ' '
	}

	// texture
	if render != nil {
		surface, err := font.RenderGlyphBlended(ch, sdl.Color{R: 255, G: 255, B: 255, A: 255})
		if err != nil {
			return nil, fmt.Errorf("RenderGlyphBlended() failed: %w", err)
		}
		defer surface.Free()

		self.texture, err = render.CreateTextureFromSurface(surface)
		if err != nil {
			return nil, fmt.Errorf("CreateTextureFromSurface() failed: %w", err)
		}
	}

	// coords
	mt, err := font.GlyphMetrics(ch)
	if err != nil {
		return nil, fmt.Errorf("GlyphMetrics() failed: %w", err)
	}
	self.x = mt.MinX
	self.y = mt.MinY // -FontLetter_size(self).y
	self.len = mt.Advance

	if tab {
		self.len *= SKYALT_FONT_TAB_WIDTH
	}

	return &self, nil
}

func (font *FontLetter) Free() error {

	if font.texture != nil {
		err := font.texture.Destroy()
		if err != nil {
			//err := fmt.Errorf("TextureDestroy() failed: %w", err)
			fmt.Printf("Error: TextureDestroy() failed: %s\n", err)
		}
	}

	return nil
}

func (font *FontLetter) Size() (OsV2, error) {

	if font.texture != nil {
		_, _, x, y, err := font.texture.Query()

		if err != nil {
			return OsV2{}, fmt.Errorf("TextureQuery() failed: %w", err)
		}
		return OsV2{int(x), int(y)}, nil
	}
	return OsV2{}, nil
}

type FontHeight struct {
	font    *ttf.Font
	letters map[int]FontLetter
}

func NewFontHeight(path string, h int) (*FontHeight, error) {
	var self FontHeight

	var err error
	self.font, err = ttf.OpenFont(path, int(h))
	if err != nil {
		return nil, fmt.Errorf("OpenFont() failed: %w", err)
	}

	self.letters = make(map[int]FontLetter)

	return &self, nil
}

func (font *FontHeight) Destroy() error {

	font.font.Close()

	for _, it := range font.letters {
		it.Free()
	}
	return nil
}

type Font struct {
	path    string
	heights map[int]*FontHeight
}

func NewFont(path string) *Font {
	var self Font
	self.path = path
	self.heights = make(map[int]*FontHeight)
	return &self
}

func (font *Font) Destroy() error {

	for _, it := range font.heights {
		it.Destroy()
	}
	return nil
}

func (font *Font) Get(ch rune, h int, render *sdl.Renderer) (FontLetter, error) {

	hh, ok := font.heights[h]
	if !ok {
		var err error
		hh, err = NewFontHeight(font.path, h)
		if err != nil {
			return FontLetter{}, fmt.Errorf("NewFontHeight() failed: %w", err)
		}
		font.heights[h] = hh
	}

	l, ok := hh.letters[int(ch)]
	if !ok || l.len == 0 || (render != nil && l.texture == nil) {
		ll, err := InitFontLetter(ch, hh.font, render)
		if err != nil {
			return FontLetter{}, fmt.Errorf("InitFontLetter() failed: %w", err)
		}
		hh.letters[int(ch)] = *ll
	}

	return l, nil
}

func (font *Font) GetDownY(text string, h int, render *sdl.Renderer) (int, error) {
	down_y := 0
	for _, ch := range text {
		l, err := font.Get(ch, h, render)
		if err != nil {
			return 0, fmt.Errorf("Start.Get() failed: %w", err)
		}
		if -l.y > down_y {
			down_y = -l.y
		}
	}
	return down_y, nil
}

func (font *Font) Start(text string, h int, coord OsV4, align OsV2, render *sdl.Renderer) (OsV2, error) {

	word_space := 0
	len := 0
	down_y := 0

	for _, ch := range text {
		l, err := font.Get(ch, h, render)
		if err != nil {
			return OsV2{}, fmt.Errorf("Start.Get() failed: %w", err)
		}
		len += (l.len + word_space)

		if -l.y > down_y {
			down_y = -l.y
		}
	}

	pos := coord.Start
	if align.X == 0 {
		// left
		// pos.x += H / 2
	} else if align.X == 1 {
		// center
		if len > coord.Size.X {
			pos.X = coord.Start.X // + H / 2
		} else {
			pos.X = coord.Middle().X - len/2
		}
	} else {
		// right
		pos.X = coord.End().X - len
	}

	// y
	if h >= coord.Size.Y {
		pos.Y += (coord.Size.Y - h) / 2
	} else {
		if align.Y == 0 {
			pos.Y = coord.Start.Y // + H / 2
		} else if align.Y == 1 {
			pos.Y += (coord.Size.Y - h) / 2
		} else if align.Y == 2 {
			pos.Y += (coord.Size.Y) - h
		}
	}

	return pos, nil
}

func (font *Font) Print(text string, h int, coord OsV4, align OsV2, color OsCd, cds []OsCd, blendingOn bool, render *sdl.Renderer) error {

	pos, err := font.Start(text, h, coord, align, render)
	if err != nil {
		return fmt.Errorf("Print.Start() failed: %w", err)
	}
	posStart := pos.X

	i := 0
	for _, ch := range text {
		if ch == '\n' {
			pos.X = posStart
			pos.Y += int(float32(h) * 1.7)
			i++
			continue
		}

		l, err := font.Get(ch, h, render)
		if err != nil {
			return fmt.Errorf("Print.Get() failed: %w", err)
		}

		sz, err := l.Size()
		if err != nil {
			return fmt.Errorf("Size() failed: %w", err)
		}

		var cd OsCd
		if len(cds) > 0 {
			cd = cds[i]
		} else {
			cd = color
		}

		if !blendingOn {
			l.texture.SetBlendMode(sdl.BLENDMODE_NONE)
		}

		err = l.texture.SetColorMod(cd.R, cd.G, cd.B)
		if err != nil {
			return fmt.Errorf("SetColorMod() failed: %w", err)
		}

		err = l.texture.SetAlphaMod(cd.A)
		if err != nil {
			return fmt.Errorf("SetAlphaMod() failed: %w", err)
		}

		err = render.Copy(l.texture, nil, &sdl.Rect{X: int32(pos.X), Y: int32(pos.Y), W: int32(sz.X), H: int32(sz.Y)})
		if err != nil {
			return fmt.Errorf("Copy() failed: %w", err)
		}

		pos.X += l.len
		i++
	}

	return nil
}

func (font *Font) GetPxPos(text string, h int, ch_pos int) (int, error) {

	px := 0

	i := 0
	for _, ch := range text {
		if i >= ch_pos {
			break
		}
		l, err := font.Get(ch, h, nil)
		if err != nil {
			return 0, fmt.Errorf("GetPxPos.Get() failed: %w", err)
		}
		px += l.len
		i++
	}

	return px, nil
}

func (font *Font) GetChPos(text string, h int, px int) (int, error) {

	px_act := 0

	i := 0
	for _, ch := range text {
		l, err := font.Get(ch, h, nil)
		if err != nil {
			return 0, fmt.Errorf("GetChPos.Get() failed: %w", err)
		}
		if px < (px_act + l.len/2) {
			return i, nil
		}

		px_act += l.len
		i++
	}

	return len(text), nil
}

func (font *Font) GetTextSize(text string, textH int, lineH int) (OsV2, error) {

	nlines := 0
	maxLineWidth := 0
	for _, line := range strings.Split(strings.TrimSuffix(text, "\n"), "\n") {
		maxLineWidth = OsMax(maxLineWidth, len(line))
		nlines++
	}

	x, err := font.GetPxPos(text, textH, maxLineWidth) // + textH
	if err != nil {
		return OsV2{}, fmt.Errorf("GetTextSize.GetPxPos() failed: %w", err)
	}
	y := nlines * lineH

	return OsV2{x, y}, nil
}

func (font *Font) GetTextPos(touchPos OsV2, text string, coord OsV4, h int, align OsV2) (int, error) {

	start, err := font.Start(text, h, coord, align, nil)
	if err != nil {
		return 0, fmt.Errorf("GetTextPos.Start() failed: %w", err)
	}
	return font.GetChPos(text, h, touchPos.X-start.X)
}

type Fonts struct {
	fonts []*Font
}

func NewFonts() *Fonts {
	var fonts Fonts
	return &fonts
}

func (fonts *Fonts) Destroy() error {
	for _, it := range fonts.fonts {
		it.Destroy()
	}
	return nil
}

func (fonts *Fonts) Get(path string) *Font {
	//find
	for _, f := range fonts.fonts {
		if f.path == path {
			return f
		}
	}

	//add
	f := NewFont(path)
	if f != nil {
		fonts.fonts = append(fonts.fonts, f)
	}
	return f
}
