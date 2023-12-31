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
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// const SKYALT_FONT_PATH = "resources/Inter/Inter-V.ttf"
const SKYALT_FONT_PATH = "resources/arial.ttf"
const SKYALT_FONT_HEIGHT = 0.36
const SKYALT_FONT_TAB_WIDTH = 4

type FontFamilyName struct {
	Weight int
	Name   string
}

// add "Italic" for italics
var g_FontFamilyNames = []FontFamilyName{
	{100, "Thin"},
	{200, "ExtraLight"},
	{300, "Light"},
	{400, "Regular"},
	{500, "Medium"},
	{600, "SemiBold"},
	{700, "Bold"},
	{800, "ExtraBold"},
	{900, "Black"},
}

var g_Font_DEFAULT_Weight = 400

func GetFontFamilyNamesIndex(weight int) int {
	weight = OsClamp(weight, 100, 900)
	i := (weight / 100) - 1
	if weight%100 >= 50 {
		i++
	}
	return i
}

type FontLetter struct {
	texture *sdl.Texture
	x       int
	y       int
	h       int
	len     int
}

func NewFontLetter(ch rune, font *ttf.Font, style ttf.Style, render *sdl.Renderer) (*FontLetter, int, error) {
	var self FontLetter
	var bytes int

	tab := (ch == '\t')
	if tab {
		ch = ' '
	}

	font.SetStyle(style)
	font.SetHinting(ttf.HINTING_LIGHT)

	// texture
	if render != nil {
		surface, err := font.RenderGlyphBlended(ch, sdl.Color{R: 255, G: 255, B: 255, A: 255})
		if err != nil {
			return nil, 0, fmt.Errorf("RenderGlyphBlended() failed: %w", err)
		}
		defer surface.Free()

		bytes = int(surface.W * surface.H * 4)

		self.texture, err = render.CreateTextureFromSurface(surface)
		if err != nil {
			return nil, 0, fmt.Errorf("CreateTextureFromSurface() failed: %w", err)
		}
	}

	// coords
	mt, err := font.GlyphMetrics(ch)
	if err != nil {
		return nil, 0, fmt.Errorf("GlyphMetrics() failed: %w", err)
	}
	self.x = mt.MinX
	self.y = mt.MinY // -FontLetter_size(self).y
	self.h = mt.MaxY
	self.len = mt.Advance

	if tab {
		self.len *= SKYALT_FONT_TAB_WIDTH
	}

	return &self, bytes, nil
}

func (font *FontLetter) Destroy() {
	if font.texture != nil {
		err := font.texture.Destroy()
		if err != nil {
			fmt.Printf("Error: TextureDestroy() failed: %s\n", err)
		}
	}
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

type Font struct {
	isFamily bool
	path     string

	files   map[int]*ttf.Font        //[height]
	letters map[[16]byte]*FontLetter //[height + weight(bold) + (regular/italic) + ch]

	bytes int
}

func NewFont(path string) *Font {
	var font Font

	font.isFamily = OsFolderExists(path)
	font.path = path

	font.files = make(map[int]*ttf.Font)
	font.letters = make(map[[16]byte]*FontLetter)

	return &font
}
func (font *Font) Destroy() {
	for _, it := range font.files {
		it.Close()
	}

	for _, it := range font.letters {
		it.Destroy()
	}
}

func (font *Font) GetStyle(weight int, italic bool) ttf.Style {
	style := ttf.STYLE_NORMAL
	if !font.isFamily {
		if weight > g_Font_DEFAULT_Weight {
			style = ttf.STYLE_BOLD
		}
		if italic {
			style = ttf.STYLE_ITALIC
		}
	}
	return style
}

func (font *Font) findFile(height int, weight int) (*ttf.Font, error) {

	file, found := font.files[height]
	if !found {
		//add font
		path := font.path
		if font.isFamily {
			family_i := GetFontFamilyNamesIndex(weight)
			path += "-" + g_FontFamilyNames[family_i].Name + ".otf" //bug: It's not "path/Inter", ale pouze "path/" ... + check finish "/" ......
		}

		var err error
		file, err = ttf.OpenFont(path, height)
		if err != nil {
			return nil, fmt.Errorf("OpenFont() failed: %w", err)
		}
		font.files[height] = file
	}
	return file, nil
}

func (font *Font) addLetter(letterId [16]byte, ch rune, height int, weight int, italic bool, render *sdl.Renderer) (*FontLetter, error) {

	//find file
	var err error
	file, err := font.findFile(height, weight)
	if err != nil {
		return nil, fmt.Errorf("findFile() failed: %w", err)
	}

	//add
	letter, bytes, err := NewFontLetter(ch, file, font.GetStyle(weight, italic), render)
	if err != nil {
		return nil, fmt.Errorf("NewFontLetter() failed: %w", err)
	}
	font.letters[letterId] = letter
	font.bytes += bytes
	return letter, nil
}

func (font *Font) Get(ch rune, height int, weight int, italic bool, render *sdl.Renderer) (*FontLetter, error) {

	var letterId [16]byte
	binary.LittleEndian.PutUint32(letterId[0:4], uint32(ch))
	binary.LittleEndian.PutUint32(letterId[4:8], uint32(height))
	binary.LittleEndian.PutUint32(letterId[8:12], uint32(weight))
	binary.LittleEndian.PutUint32(letterId[12:16], uint32(OsTrn(italic, 1, 0)))

	//find letter
	letter, found := font.letters[letterId]
	if found {
		if letter.len == 0 || (render != nil && letter.texture == nil) {

			//reload again(with texture)
			var err error
			letter, err = font.addLetter(letterId, ch, height, weight, italic, render)
			if err != nil {
				return nil, fmt.Errorf("addLetter() failed: %w", err)
			}
		}

		return letter, nil
	}

	//add letter
	var err error
	letter, err = font.addLetter(letterId, ch, height, weight, italic, render)
	if err != nil {
		return nil, fmt.Errorf("addLetter() failed: %w", err)
	}

	return letter, nil
}

func (font *Font) processLetter(text string, origW int, origH int, weight *int, italic *bool, height *int, skip *int) bool {

	if *skip > 0 {
		*skip -= 1
		return false
	}

	//new line = reset
	if strings.HasPrefix(text, "\n") {
		*weight = origW
		*italic = false
		*height = origH
	}

	//bold + italic
	if strings.HasPrefix(text, "***") || strings.HasPrefix(text, "___") {
		*weight = OsTrn(*weight != origW, origW, origW*3/2) //bold
		*italic = !*italic
		*skip = 2
		return false
	}

	//bold
	if strings.HasPrefix(text, "**") {
		*weight = OsTrn(*weight != origW, origW, origW*3/2)
		*skip = 1
		return false
	}

	//italic
	if strings.HasPrefix(text, "__") {
		*italic = !*italic
		*skip = 1
		return false
	}

	//smaller
	if strings.HasPrefix(text, "###") {
		*height = OsTrn(*height != origH, origH, int(float64(origH)*0.9))
		*skip = 2
		return false
	}

	//taller
	if strings.HasPrefix(text, "##") {
		*height = OsTrn(*height != origH, origH, int(float64(origH)*1.2))
		*skip = 1
		return false
	}

	return true
}

func (font *Font) GetPxPos(text string, w int, h int, ch_pos int, enableFormating bool) (int, error) {

	px := 0

	weight := w
	italic := false
	height := h
	skip := 0

	i := 0
	for p, ch := range text {
		if enableFormating && !font.processLetter(text[p:], w, h, &weight, &italic, &height, &skip) {
			i++
			continue
		}

		if i >= ch_pos {
			break
		}
		l, err := font.Get(ch, height, weight, italic, nil)
		if err != nil {
			return 0, fmt.Errorf("GetPxPos.Get() failed: %w", err)
		}
		px += l.len
		i++
	}

	return px, nil
}

func (font *Font) GetDownY(text string, w int, h int, enableFormating bool, render *sdl.Renderer) (int, error) {

	weight := w
	italic := false
	height := h
	skip := 0

	down_y := 0
	for p, ch := range text {
		if enableFormating && !font.processLetter(text[p:], w, h, &weight, &italic, &height, &skip) {
			continue
		}

		l, err := font.Get(ch, height, weight, italic, render)
		if err != nil {
			return 0, fmt.Errorf("Start.Get() failed: %w", err)
		}
		if -l.y > down_y {
			down_y = -l.y
		}
	}
	return down_y, nil
}

func (font *Font) Start(text string, w int, h int, coord OsV4, align OsV2, enableFormating bool, render *sdl.Renderer) (OsV2, error) {
	word_space := 0
	len := 0
	//down_y := 0
	max_tex_h := 0

	weight := w
	italic := false
	height := h
	skip := 0

	for p, ch := range text {
		if enableFormating && !font.processLetter(text[p:], w, h, &weight, &italic, &height, &skip) {
			continue
		}

		l, err := font.Get(ch, height, weight, italic, render)
		if err != nil {
			return OsV2{}, fmt.Errorf("Start.Get() failed: %w", err)
		}
		len += (l.len + word_space)

		//if -l.y > down_y {
		//	down_y = -l.y
		//}

		//sz, _ := l.Size()
		max_tex_h = OsMax(max_tex_h, h) //sz.y

	}

	h = max_tex_h //+ down_y

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

func (font *Font) GetChPos(text string, w int, h int, px int, enableFormating bool) (int, error) {

	px_act := 0

	weight := w
	italic := false
	height := h
	skip := 0

	i := 0
	for p, ch := range text {
		if enableFormating && !font.processLetter(text[p:], w, h, &weight, &italic, &height, &skip) {
			i++
			continue
		}

		l, err := font.Get(ch, height, weight, italic, nil)
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

func (font *Font) GetTouchPos(touchPos OsV2, text string, coord OsV4, w int, h int, align OsV2, enableFormating bool) (int, error) {

	start, err := font.Start(text, w, h, coord, align, enableFormating, nil)
	if err != nil {
		return 0, fmt.Errorf("Start() failed: %w", err)
	}
	return font.GetChPos(text, w, h, touchPos.X-start.X, enableFormating)
}

func (font *Font) GetTextSize(text string, w int, h int, lineH int, enableFormating bool) (OsV2, error) {

	nlines := 0
	maxLineWidth := 0
	for _, line := range strings.Split(strings.TrimSuffix(text, "\n"), "\n") {
		maxLineWidth = OsMax(maxLineWidth, len(line))
		nlines++
	}

	x, err := font.GetPxPos(text, w, h, maxLineWidth, enableFormating) // + textH
	if err != nil {
		return OsV2{}, fmt.Errorf("GetPxPos() failed: %w", err)
	}
	y := nlines * lineH

	return OsV2{x, y}, nil
}

func (font *Font) Print(text string, w int, h int, coord OsV4, align OsV2, color OsCd, cds []OsCd, blendingOn bool, enableFormating bool, render *sdl.Renderer) error {
	pos, err := font.Start(text, w, h, coord, align, enableFormating, render)
	if err != nil {
		return fmt.Errorf("Print.Start() failed: %w", err)
	}
	posStart := pos.X
	max_h := h

	weight := w
	italic := false
	height := h
	skip := 0

	i := 0
	for p, ch := range text {
		if enableFormating && !font.processLetter(text[p:], w, h, &weight, &italic, &height, &skip) {
			continue
		}
		max_h = OsMax(max_h, height)

		if ch == '\n' {
			pos.X = posStart
			pos.Y += int(float32(max_h) * 1.7)
			max_h = h
			i++
			continue
		}

		l, err := font.Get(ch, height, weight, italic, render)
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
			l.texture.SetBlendMode(sdl.BLENDMODE_NONE) //BUG: when textures(letters) are overlaping ... render whole string, not every letter? ...
		} else {
			l.texture.SetBlendMode(sdl.BLENDMODE_BLEND)
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

type Fonts struct {
	fonts map[string]*Font
}

func NewFonts() *Fonts {
	var fonts Fonts
	fonts.fonts = make(map[string]*Font)
	return &fonts
}
func (fonts *Fonts) Destroy() {
	for _, it := range fonts.fonts {
		it.Destroy()
	}
}

func (fonts *Fonts) Bytes() int {
	bytes := 0
	for _, it := range fonts.fonts {
		bytes += it.bytes
	}
	return bytes
}

func (fonts *Fonts) Maintenance() {
	for k, it := range fonts.fonts {
		if it.bytes > 2*1024*1024 { //over 2MB
			it.Destroy()
			delete(fonts.fonts, k)
		}
	}
}

func (fonts *Fonts) Get(path string) *Font {

	if len(path) == 0 {
		path = SKYALT_FONT_PATH
	}

	//find
	font, found := fonts.fonts[path]
	if found {
		return font
	}

	//add
	font = NewFont(path)
	fonts.fonts[path] = font
	return font
}
