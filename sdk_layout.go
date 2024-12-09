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
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"path/filepath"
	"runtime"
	"time"
)

type Rect struct {
	X, Y, W, H float64
}

func (r Rect) Is() bool {
	return r.W > 0 && r.H > 0
}
func (r Rect) check() Rect {
	if r.W < 0 {
		r.W = 0
	}
	if r.H < 0 {
		r.H = 0
	}
	return r
}

func (r Rect) Cut(v float64) Rect {
	r.X += v
	r.Y += v
	r.W -= 2 * v
	r.H -= 2 * v
	return r.check()
}
func (r Rect) CutLeft(v float64) Rect {
	r.X += v
	r.W -= v
	return r.check()
}
func (r Rect) CutTop(v float64) Rect {
	r.Y += v
	r.H -= v
	return r.check()
}
func (r Rect) CutRight(v float64) Rect {
	r.W -= v
	return r.check()
}
func (r Rect) CutBottom(v float64) Rect {
	r.H -= v
	return r.check()
}
func (r Rect) GetPos(x, y float64) (float64, float64) {
	return r.X + r.W*x, r.Y + r.H*y
}
func (r Rect) Move(x, y float64) Rect {
	r.X -= x
	r.Y -= y
	return r
}

func Rect_centerFull(out Rect, in_w, in_h float64) Rect {
	var r Rect
	r.X = out.X
	r.Y = out.Y
	r.W = in_w
	r.H = in_h

	if out.W != in_w {
		r.X += (out.W - in_w) / 2
	}
	if out.Y != in_h {
		r.Y += (out.H - in_h) / 2
	}
	return r
}

func (r Rect) IsInside(x, y float64) bool {
	return (x > r.X && x < r.X+r.W && y > r.Y && y < r.Y+r.H)
}

func Color_Aprox(s color.RGBA, e color.RGBA, t float32) color.RGBA {
	var self color.RGBA
	self.R = byte(float32(s.R) + (float32(e.R)-float32(s.R))*t)
	self.G = byte(float32(s.G) + (float32(e.G)-float32(s.G))*t)
	self.B = byte(float32(s.B) + (float32(e.B)-float32(s.B))*t)
	self.A = byte(float32(s.A) + (float32(e.A)-float32(s.A))*t)
	return self
}

type LayoutPalette struct {
	P, S, T, E, B           color.RGBA
	OnP, OnS, OnT, OnE, OnB color.RGBA
}

func (pl *LayoutPalette) GetGrey(t float32) color.RGBA {
	return Color_Aprox(pl.S, pl.OnS, t)
}

type LayoutCmd struct {
	Hash   uint64
	Cmd    string
	Param1 string
	Param2 string
}

type LayoutDrawPrim struct {
	Type uint8

	Rect           Rect
	Sx, Sy, Ex, Ey float64

	Cd, Cd_over, Cd_down color.RGBA

	Border float64

	Text  string
	Text2 string

	Boolean bool

	Align_h uint8
	Align_v uint8

	Text_formating    bool
	Text_multiline    bool
	Text_linewrapping bool
	Text_selection    bool
	Text_editable     bool
}

type LayoutCR struct {
	Pos int     `json:",omitempty"`
	Min float64 `json:",omitempty"`
	Max float64 `json:",omitempty"`

	Resize_value float64 `json:",omitempty"`

	SetFromChild bool `json:",omitempty"`

	Caller_file string `json:",omitempty"`
	Caller_line int    `json:",omitempty"`
}

type LayoutScroll struct {
	Show   bool
	Narrow bool
}

type LayoutPickGrid struct {
	Cd                             color.RGBA
	Grid_x, Grid_y, Grid_w, Grid_h int
	Label                          string
}

type LayoutInput struct {
	Rect Rect

	IsStart  bool
	IsActive bool
	IsEnd    bool

	IsInside bool //rename IsOver ...
	IsUse    bool //IsActive? .....
	IsUp     bool //IsEnd? .....

	X, Y      float64
	Wheel     int
	NumClicks int
	AltClick  bool

	SetEdit   bool
	EditValue string
}
type Layout struct {
	X, Y, W, H int
	Name       string

	Childs []*Layout
	Hash   uint64

	Enable bool
	LLMTip string

	Shortcut_key byte

	Back_cd     color.RGBA
	Back_margin float64
	Border_cd   color.RGBA

	Drag_group              string
	Drop_group              string
	Drag_index              int
	Drop_h, Drop_v, Drop_in bool

	UserCols []LayoutCR
	UserRows []LayoutCR

	ScrollV LayoutScroll
	ScrollH LayoutScroll

	Caller_file string `json:",omitempty"`
	Caller_line int    `json:",omitempty"`

	fnBuild      func()
	fnDraw       func(Rect)
	fnInput      func(LayoutInput)
	fnSetEditbox func(string)

	Canvas Rect
	buffer []LayoutDrawPrim

	Prompt_label   string
	Prompt_comp_cd color.RGBA //highlight with rect
	Prompt_grids   []LayoutPickGrid

	Progress float64
	updated  bool
	redraw   bool
	done     bool
}

func (layout *Layout) _getName() string {
	return fmt.Sprintf("%s(%d,%d,%d,%d)", layout.Name, layout.X, layout.Y, layout.W, layout.H)
}
func (layout *Layout) _computeHash(parent *Layout) uint64 {
	if parent == nil {
		return 0
	}

	h := sha256.New()

	//parent
	var pt [8]byte
	binary.LittleEndian.PutUint64(pt[:], parent.Hash)
	h.Write(pt[:])

	//this
	h.Write([]byte(layout._getName()))

	return binary.LittleEndian.Uint64(h.Sum(nil))
}

func (layout *Layout) FindDialog(name string) *Layout {
	name = "_dialog_" + name

	for _, it := range layout.Childs {
		if it.Name == name {
			return it
		}
	}
	return nil
}

func (layout *Layout) _findChild(x, y, w, h int, name string) *Layout {
	for _, it := range layout.Childs {
		if it.X == x && it.Y == y && it.W == w && it.H == h && it.Name == name {
			return it
		}
	}
	return nil
}
func (layout *Layout) _findHash(hash uint64) *Layout {
	if layout.Hash == hash {
		return layout
	}

	for _, it := range layout.Childs {
		d := it._findHash(hash)
		if d != nil {
			return d
		}
	}
	return nil
}

func _newLayout(x, y, w, h int, name string, parent *Layout) *Layout {
	layout := &Layout{X: x, Y: y, W: w, H: h, Name: name, Enable: true} //, Canvas: Rect{0, 0, -1, -1}}
	layout.Hash = layout._computeHash(parent)
	return layout
}

func _newLayoutRoot() *Layout {
	return _newLayout(0, 0, 0, 0, "Root", nil)
}

func (layout *Layout) _createDiv(x, y, w, h int, name string, fnBuild func(), fnDraw func(rect Rect), fnInput func(touch LayoutInput)) *Layout {

	lay := layout._findChild(x, y, w, h, name)
	if lay == nil {
		lay = _newLayout(x, y, w, h, name, layout)
		layout.Childs = append(layout.Childs, lay)
	}

	lay.fnBuild = fnBuild
	lay.fnDraw = fnDraw
	lay.fnInput = fnInput

	var ok bool
	_, lay.Caller_file, lay.Caller_line, ok = runtime.Caller(2)
	if !ok {
		log.Fatal("runtime.Caller failed")
	}
	lay.Caller_file = filepath.Base(lay.Caller_file)

	return lay
}

func (layout *Layout) AddLayoutWithName(x, y, w, h int, name string) *Layout {
	name = "_layout_" + name
	return layout._createDiv(x, y, w, h, name, nil, nil, nil)
}
func (layout *Layout) AddLayout(x, y, w, h int) *Layout {
	return layout._createDiv(x, y, w, h, "_layout", nil, nil, nil)
}

func (layout *Layout) AddDialog(name string) *Layout {
	name = "_dialog_" + name

	lay := layout._findChild(0, 0, 0, 0, name)
	if lay == nil {
		lay = _newLayout(0, 0, 0, 0, name, layout)
		layout.Childs = append(layout.Childs, lay)
	}

	return lay
}

func (layout *Layout) AddDialogBorder(name string, title string, width float64) *Layout {

	dia := layout.AddDialog(name)
	dia.SetColumn(1, 1, width)
	//dia.SetRow(1, 1, height)
	dia.SetRowFromSub(1)
	dia.SetColumn(2, 1, 1)
	dia.SetRow(2, 1, 1)

	tx := dia.AddText(0, 0, 3, 1, title)
	tx.Align_h = 1

	return dia.AddLayout(1, 1, 1, 1)
}

func _extractFileName(path string) string {
	return filepath.Base(path)
}

func (layout *Layout) SetColumn(grid_x int, min_size, max_size float64) {
	_, caller_file, caller_line, ok := runtime.Caller(1)
	if !ok {
		fmt.Println("runtime.Caller failed")
	}

	newItem := LayoutCR{Pos: grid_x, Min: min_size, Max: max_size, Caller_file: _extractFileName(caller_file), Caller_line: caller_line}
	for i := range layout.UserCols {
		if layout.UserCols[i].Pos == grid_x {
			layout.UserCols[i] = newItem
			return
		}
	}

	layout.UserCols = append(layout.UserCols, newItem)
}

func (layout *Layout) SetRow(grid_y int, min_size, max_size float64) {
	_, caller_file, caller_line, ok := runtime.Caller(1)
	if !ok {
		fmt.Println("runtime.Caller failed")
	}

	newItem := LayoutCR{Pos: grid_y, Min: min_size, Max: max_size, Caller_file: _extractFileName(caller_file), Caller_line: caller_line}

	for i := range layout.UserRows {
		if layout.UserRows[i].Pos == grid_y {
			layout.UserRows[i] = newItem
			return
		}
	}

	layout.UserRows = append(layout.UserRows, newItem)
}

func (layout *Layout) SetColumnFromSub(grid_x int) {
	_, caller_file, caller_line, ok := runtime.Caller(1)
	if !ok {
		fmt.Println("runtime.Caller failed")
	}

	newItem := LayoutCR{Pos: grid_x, SetFromChild: true, Min: 1, Max: 1, Caller_file: _extractFileName(caller_file), Caller_line: caller_line}

	for i := range layout.UserCols {
		if layout.UserCols[i].Pos == grid_x {
			layout.UserCols[i] = newItem
			return
		}
	}

	layout.UserCols = append(layout.UserCols, newItem)
}

func (layout *Layout) SetRowFromSub(grid_y int) {
	_, caller_file, caller_line, ok := runtime.Caller(1)
	if !ok {
		fmt.Println("runtime.Caller failed")
	}

	newItem := LayoutCR{Pos: grid_y, SetFromChild: true, Min: 1, Max: 1, Caller_file: _extractFileName(caller_file), Caller_line: caller_line}

	for i := range layout.UserRows {
		if layout.UserRows[i].Pos == grid_y {
			layout.UserRows[i] = newItem
			return
		}
	}

	layout.UserRows = append(layout.UserRows, newItem)
}

func (layout *Layout) SetColumnResizable(grid_x int, min_size, max_size, default_size float64) {
	_, caller_file, caller_line, ok := runtime.Caller(1)
	if !ok {
		fmt.Println("runtime.Caller failed")
	}

	layout.UserCols = append(layout.UserCols, LayoutCR{Pos: grid_x, Min: min_size, Max: max_size, Resize_value: default_size, Caller_file: _extractFileName(caller_file), Caller_line: caller_line})

}
func (layout *Layout) SetRowResizable(grid_y int, min_size, max_size, default_size float64) {
	_, caller_file, caller_line, ok := runtime.Caller(1)
	if !ok {
		fmt.Println("runtime.Caller failed")
	}

	layout.UserRows = append(layout.UserRows, LayoutCR{Pos: grid_y, Min: min_size, Max: max_size, Resize_value: default_size, Caller_file: _extractFileName(caller_file), Caller_line: caller_line})
}

func (layout *Layout) VScrollToTheTop() {
	_addCmd(LayoutCmd{Hash: layout.Hash, Cmd: "VScrollToTheTop"})
}
func (layout *Layout) VScrollToTheBottom() {
	_addCmd(LayoutCmd{Hash: layout.Hash, Cmd: "VScrollToTheBottom"})
}
func (layout *Layout) HScrollToTheTop() {
	_addCmd(LayoutCmd{Hash: layout.Hash, Cmd: "HScrollToTheTop"})
}
func (layout *Layout) HScrollToTheBottom() {
	_addCmd(LayoutCmd{Hash: layout.Hash, Cmd: "HScrollToTheBottom"})
}

func (layout *Layout) OpenDialogCentered() {
	_addCmd(LayoutCmd{Hash: layout.Hash, Cmd: "OpenDialogCentered"})
}
func (layout *Layout) OpenDialogRelative(parent *Layout) {
	if parent != nil {
		js, err := json.Marshal(parent.Hash)
		if err == nil {
			_addCmd(LayoutCmd{Hash: layout.Hash, Cmd: "OpenDialogRelative", Param1: string(js)})
		}
	}
}
func (layout *Layout) OpenDialogOnTouch() {
	_addCmd(LayoutCmd{Hash: layout.Hash, Cmd: "OpenDialogOnTouch"})
}
func (layout *Layout) CloseDialog() {
	_addCmd(LayoutCmd{Hash: layout.Hash, Cmd: "CloseDialog"})
}

func (layout *Layout) SetClipboardText(text string) {
	_addCmd(LayoutCmd{Hash: 0, Cmd: "SetClipboardText", Param1: text})
}

func (layout *Layout) Refresh() {
	_addCmd(LayoutCmd{Hash: 0, Cmd: "Refresh"})
}

func (layout *Layout) CopyText() {
	_addCmd(LayoutCmd{Hash: layout.Hash, Cmd: "CopyText"})
}
func (layout *Layout) SelectAllText() {
	_addCmd(LayoutCmd{Hash: layout.Hash, Cmd: "SelectAllText"})
}
func (layout *Layout) CutText() {
	_addCmd(LayoutCmd{Hash: layout.Hash, Cmd: "CutText"})
}
func (layout *Layout) PasteText() {
	_addCmd(LayoutCmd{Hash: layout.Hash, Cmd: "PasteText"})
}

var g_theme_light = LayoutPalette{
	P:   color.RGBA{37, 100, 120, 255},
	OnP: color.RGBA{255, 255, 255, 255},

	S:   color.RGBA{85, 95, 100, 255},
	OnS: color.RGBA{255, 255, 255, 255},

	T:   color.RGBA{90, 95, 115, 255},
	OnT: color.RGBA{255, 255, 255, 255},

	E:   color.RGBA{180, 40, 30, 255},
	OnE: color.RGBA{255, 255, 255, 255},

	B:   color.RGBA{250, 250, 250, 255},
	OnB: color.RGBA{25, 27, 30, 255},
}

var g_theme_dark = LayoutPalette{
	P:   color.RGBA{150, 205, 225, 255},
	OnP: color.RGBA{0, 50, 65, 255},

	S:   color.RGBA{190, 200, 205, 255},
	OnS: color.RGBA{40, 50, 55, 255},

	T:   color.RGBA{195, 200, 220, 255},
	OnT: color.RGBA{75, 35, 50, 255},

	E:   color.RGBA{240, 185, 180, 255},
	OnE: color.RGBA{45, 45, 65, 255},

	B:   color.RGBA{25, 30, 30, 255},
	OnB: color.RGBA{230, 230, 230, 255},
}

func (layout *Layout) Cell() int { //number of pixels in one cell
	return int(float32(NewFile_Env().Dpi) / 2.5)
}

func (layout *Layout) GetPalette() *LayoutPalette {

	env := NewFile_Env()

	theme := env.Theme

	hour := time.Now().Hour()

	if env.UseDarkTheme {
		if (env.UseDarkThemeStart < env.UseDarkThemeEnd && hour >= env.UseDarkThemeStart && hour < env.UseDarkThemeEnd) ||
			(env.UseDarkThemeStart > env.UseDarkThemeEnd && (hour >= env.UseDarkThemeStart || hour < env.UseDarkThemeEnd)) {
			theme = "dark"
		}
	}

	switch theme {
	case "light":
		return &g_theme_light

	case "dark":
		return &g_theme_dark
	}

	return &env.CustomPalette
}

func (layout *Layout) GetDateFormat() string {
	return NewFile_Env().DateFormat
}

func (layout *Layout) WriteError(err error) error {
	//who calls this function and write it ........
	if err != nil {
		NewFile_Logs().AddError(err, 0)
	}
	return err
}

func (layout *Layout) Paint_rect(rect Rect, cd, cd_over, cd_down color.RGBA, borderWidth float64) {
	layout.buffer = append(layout.buffer, LayoutDrawPrim{Type: 1, Rect: rect, Cd: cd, Cd_over: cd_over, Cd_down: cd_down, Border: borderWidth})
}

func (layout *Layout) Paint_circle(rect Rect, cd, cd_over, cd_down color.RGBA, borderWidth float64) {
	layout.buffer = append(layout.buffer, LayoutDrawPrim{Type: 2, Rect: rect, Cd: cd, Cd_over: cd_over, Cd_down: cd_down, Border: borderWidth})
}

func (layout *Layout) Paint_circleRad(rect Rect, x, y float64, rad_cells float64, cd, cd_over, cd_down color.RGBA, borderWidth float64) Rect {
	if rad_cells <= 0 {
		return Rect{}
	}

	rect.X += rect.W * x
	rect.Y += rect.H * y
	rect.W = rad_cells
	rect.H = rad_cells

	//move
	rect.X -= rect.W / 2
	rect.Y -= rect.H / 2

	layout.Paint_circle(rect, cd, cd, cd, 0)
	return rect
}

func (layout *Layout) Paint_file(rect Rect, fromDb bool, path string, cd, cd_over, cd_down color.RGBA, align_h, align_v uint8) {
	preFix := "file:"
	if fromDb {
		preFix = "db:"
	}

	layout.buffer = append(layout.buffer, LayoutDrawPrim{Type: 3,
		Rect: rect,
		Cd:   cd, Cd_over: cd_over, Cd_down: cd_down,
		Text:    preFix + path,
		Align_h: align_h,
		Align_v: align_v,
	})
}

func (layout *Layout) Paint_line(rect Rect, sx, sy, ex, ey float64, cd color.RGBA, width float64) {
	layout.buffer = append(layout.buffer, LayoutDrawPrim{Type: 4,
		Rect: rect,
		Cd:   cd, Cd_over: cd, Cd_down: cd,
		Border: width,
		Sx:     sx,
		Sy:     sy,
		Ex:     ex,
		Ey:     ey,
	})
}

func (layout *Layout) Paint_text(rect Rect, text string, ghost string,
	frontCd, frontCd_over, frontCd_down color.RGBA,

	selection, editable bool,
	align_h uint8, align_v uint8,
	formating bool, multiline bool, linewrapping bool,
	margin float64) {

	rect = rect.Cut(margin)

	layout.buffer = append(layout.buffer, LayoutDrawPrim{Type: 5,
		Rect: rect,
		Cd:   frontCd, Cd_over: frontCd_over, Cd_down: frontCd_down,

		Text:              text,
		Text2:             ghost,
		Align_h:           align_h,
		Align_v:           align_v,
		Text_formating:    formating,
		Text_multiline:    multiline,
		Text_linewrapping: linewrapping,
		Text_selection:    selection,
		Text_editable:     editable,
	})
}

func (layout *Layout) Paint_cursorEx(rect Rect, name string) {
	layout.buffer = append(layout.buffer, LayoutDrawPrim{Type: 6,
		Rect: rect,
		Text: name,
	})
}
func (layout *Layout) Paint_cursor(name string, rect Rect) {
	layout.Paint_cursorEx(rect, name)
}

func (layout *Layout) Paint_tooltip(text string, rect Rect) {
	layout.Paint_tooltipEx(rect, text, false)
}
func (layout *Layout) Paint_tooltipEx(rect Rect, text string, force bool) {
	if text != "" {
		layout.buffer = append(layout.buffer, LayoutDrawPrim{Type: 7,
			Rect:    rect,
			Text:    text,
			Boolean: force,
		})
	}
}
