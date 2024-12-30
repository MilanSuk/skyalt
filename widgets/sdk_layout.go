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
	"fmt"
	"image/color"
	"log"
	"path/filepath"
	"runtime"
	"strings"
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
	Hide   bool
	Narrow bool
}

type LayoutPick struct {
	Line       int
	X, Y, W, H int

	Cd       color.RGBA //paintbrush color
	time_sec float64
}

func (a *LayoutPick) Cmp(b *LayoutPick) bool {
	return a.Line == b.Line &&
		a.X == b.X &&
		a.Y == b.Y &&
		a.W == b.W &&
		a.H == b.H
}

type LayoutInput struct {
	Rect Rect

	IsStart  bool
	IsActive bool
	IsEnd    bool

	IsInside bool //rename IsOver ...
	IsUse    bool //IsActive? ...
	IsUp     bool //IsEnd? ...

	X, Y      float64
	Wheel     int
	NumClicks int
	AltClick  bool

	SetEdit   bool
	EditValue string
	EditEnter bool

	SetDropMove      bool
	DropSrc, DropDst int

	Drop_path string

	Shortcut_key byte

	Pick LayoutPick
	//PickCode  string
	//PickImage []byte
}

type LayoutDialog struct {
	Opened bool
	Layout *Layout
}

func (dia *LayoutDialog) OpenCentered() {
	_addCmd(LayoutCmd{Hash: dia.Layout.Hash, Cmd: "OpenDialogCentered"})
}
func (dia *LayoutDialog) OpenRelative(parent *Layout) {
	if parent != nil {
		_addCmd(LayoutCmd{Hash: dia.Layout.Hash, Cmd: "OpenDialogRelative", Param1: string(OsMarshal(parent.Hash))})
	} else {
		dia.OpenCentered()
	}
}
func (dia *LayoutDialog) OpenOnTouch() {
	_addCmd(LayoutCmd{Hash: dia.Layout.Hash, Cmd: "OpenDialogOnTouch"})
}
func (dia *LayoutDialog) Close() {
	_addCmd(LayoutCmd{Hash: dia.Layout.Hash, Cmd: "CloseDialog"})
}

type Layout struct {
	X, Y, W, H int
	Name       string

	dialogs []*LayoutDialog
	Childs  []*Layout
	Hash    uint64

	App bool //touch crop

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

	dropMove func(src, dst int)

	dropFile func(path string)

	UserCols []LayoutCR
	UserRows []LayoutCR

	ScrollV LayoutScroll
	ScrollH LayoutScroll

	Caller_file string `json:",omitempty"`
	Caller_line int    `json:",omitempty"`

	fnBuild      func(*Layout)
	fnDraw       func(Rect, *Layout) LayoutPaint
	fnInput      func(LayoutInput, *Layout)
	fnSetEditbox func(string, bool)
}

func (layout *Layout) _getName() string {
	return fmt.Sprintf("%s(%d,%d,%d,%d)", layout.Name, layout.X, layout.Y, layout.W, layout.H)
}
func (layout *Layout) _computeHash(parent *Layout) uint64 {
	h := sha256.New()

	//parent
	if parent != nil {
		var pt [8]byte
		binary.LittleEndian.PutUint64(pt[:], parent.Hash)
		h.Write(pt[:])
	}

	//this
	h.Write([]byte(layout._getName()))

	return binary.LittleEndian.Uint64(h.Sum(nil))
}

func (layout *Layout) FindDialog(name string) *LayoutDialog {
	for _, it := range layout.dialogs {
		if it.Layout != nil && it.Layout.Name == name {
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

	for _, it := range layout.dialogs {
		d := it.Layout._findHash(hash)
		if d != nil {
			return d
		}
	}

	return nil
}

func (layout *Layout) _findParent(find *Layout) *Layout {

	for _, it := range layout.Childs {
		if it == find {
			return layout
		}
		d := it._findParent(find)
		if d != nil {
			return d
		}
	}

	for _, it := range layout.dialogs {
		if it.Layout == find {
			return layout
		}
		d := it.Layout._findParent(find)
		if d != nil {
			return d
		}
	}

	return nil
}

func _newLayout(x, y, w, h int, name string, parent *Layout) *Layout {
	layout := &Layout{X: x, Y: y, W: w, H: h, Name: name, Enable: true}
	layout.Hash = layout._computeHash(parent)
	return layout
}

func _newLayoutRoot() *Layout {
	root := _newLayout(0, 0, 0, 0, "Root", nil)
	root.App = true
	return root
}

func (layout *Layout) _createDiv(x, y, w, h int, name string, fnBuild func(layout *Layout), fnDraw func(rect Rect, layout *Layout) LayoutPaint, fnInput func(in LayoutInput, layout *Layout)) *Layout {

	var lay *Layout
	//find
	for _, it := range layout.Childs {
		if it.X == x && it.Y == y && it.W == w && it.H == h && it.Name == name {
			lay = it
			break
		}
	}

	//add
	if lay == nil {
		lay = _newLayout(x, y, w, h, name, layout)
		layout.Childs = append(layout.Childs, lay)
	}

	//set
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

func (layout *Layout) AddDialog(name string) *LayoutDialog {

	dia := layout.FindDialog(name)
	if dia == nil {
		dia = &LayoutDialog{Layout: _newLayout(0, 0, 0, 0, name, layout)}
		layout.dialogs = append(layout.dialogs, dia)
	} else {
		fmt.Println("Dialog already exist")
	}

	return dia
}

func (layout *Layout) AddDialogBorder(name string, title string, width float64) (*LayoutDialog, *Layout) {

	dia := layout.AddDialog(name)
	lay := dia.Layout
	lay.SetColumn(1, 1, width)
	//dia.SetRow(1, 1, height)
	lay.SetRowFromSub(1)
	lay.SetColumn(2, 1, 1)
	lay.SetRow(2, 1, 1)

	tx := lay.AddText(0, 0, 3, 1, title)
	tx.Align_h = 1

	return dia, lay.AddLayout(1, 1, 1, 1)
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

func (layout *Layout) Redraw() {
	_addCmd(LayoutCmd{Hash: layout.Hash, Cmd: "Redraw"})
}
func Layout_RefreshDelayed() {
	_addCmd(LayoutCmd{Hash: 0, Cmd: "RefreshDelayed"})
}
func Layout_Recompile() {
	_addCmd(LayoutCmd{Hash: 0, Cmd: "Compile"})
}
func Layout_ResetBrushes() {
	_addCmd(LayoutCmd{Hash: 0, Cmd: "ResetBrushes"})
}

func Layout_SetClipboardText(text string) {
	_addCmd(LayoutCmd{Hash: 0, Cmd: "SetClipboardText", Param1: text})
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

func Layout_Cell() int { //number of pixels in one cell
	return int(float32(NewFile_Settings().Dpi) / 2.5)
}

func Paint_GetPalette() *LayoutPalette {

	env := NewFile_Settings()

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

func Layout_GetDateFormat() string {
	return NewFile_Settings().DateFormat
}

func Layout_WriteError(err error) error {
	//who calls this function and write it ...
	if err != nil {
		NewFile_Logs().AddError(err, 0)
	}
	return err
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
type LayoutPaint struct {
	buffer []LayoutDrawPrim
}

func (paint *LayoutPaint) Rect(rect Rect, cd, cd_over, cd_down color.RGBA, borderWidth float64) {
	paint.buffer = append(paint.buffer, LayoutDrawPrim{Type: 1, Rect: rect, Cd: cd, Cd_over: cd_over, Cd_down: cd_down, Border: borderWidth})
}

func (paint *LayoutPaint) Circle(rect Rect, cd, cd_over, cd_down color.RGBA, borderWidth float64) {
	paint.buffer = append(paint.buffer, LayoutDrawPrim{Type: 2, Rect: rect, Cd: cd, Cd_over: cd_over, Cd_down: cd_down, Border: borderWidth})
}

func (paint *LayoutPaint) CircleRad(rect Rect, x, y float64, rad_cells float64, cd, cd_over, cd_down color.RGBA, borderWidth float64) Rect {
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

	paint.Circle(rect, cd, cd, cd, 0)
	return rect
}

func (paint *LayoutPaint) File(rect Rect, fromDb bool, path string, cd, cd_over, cd_down color.RGBA, align_h, align_v uint8) {
	preFix := "file:"
	if fromDb {
		preFix = "db:"
	}

	paint.buffer = append(paint.buffer, LayoutDrawPrim{Type: 3,
		Rect: rect,
		Cd:   cd, Cd_over: cd_over, Cd_down: cd_down,
		Text:    preFix + path,
		Align_h: align_h,
		Align_v: align_v,
	})
}

func (paint *LayoutPaint) Line(rect Rect, sx, sy, ex, ey float64, cd color.RGBA, width float64) {
	paint.buffer = append(paint.buffer, LayoutDrawPrim{Type: 4,
		Rect: rect,
		Cd:   cd, Cd_over: cd, Cd_down: cd,
		Border: width,
		Sx:     sx,
		Sy:     sy,
		Ex:     ex,
		Ey:     ey,
	})
}

func (paint *LayoutPaint) Text(rect Rect, text string, ghost string, frontCd, frontCd_over, frontCd_down color.RGBA, selection, editable bool, align_h uint8, align_v uint8, formating bool, multiline bool, linewrapping bool, margin float64) {
	rect = rect.Cut(margin)

	paint.buffer = append(paint.buffer, LayoutDrawPrim{Type: 5,
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

func (paint *LayoutPaint) CursorEx(rect Rect, name string) {
	paint.buffer = append(paint.buffer, LayoutDrawPrim{Type: 6,
		Rect: rect,
		Text: name,
	})
}
func (paint *LayoutPaint) Cursor(name string, rect Rect) {
	paint.CursorEx(rect, name)
}

func (paint *LayoutPaint) Tooltip(text string, rect Rect) {
	paint.TooltipEx(rect, text, false)
}
func (paint *LayoutPaint) TooltipEx(rect Rect, text string, force bool) {
	if text != "" {
		paint.buffer = append(paint.buffer, LayoutDrawPrim{Type: 7,
			Rect:    rect,
			Text:    text,
			Boolean: force,
		})
	}
}

func Layout_GetMonthText(month int) string {
	switch month {
	case 1:
		return "January"
	case 2:
		return "February"
	case 3:
		return "March"
	case 4:
		return "April"
	case 5:
		return "May"
	case 6:
		return "June"
	case 7:
		return "July"
	case 8:
		return "August"
	case 9:
		return "September"
	case 10:
		return "October"
	case 11:
		return "November"
	case 12:
		return "December"
	}
	return ""
}

func Layout_GetDayTextFull(day int) string {
	switch day {
	case 1:
		return "Monday"
	case 2:
		return "Tuesday"
	case 3:
		return "Wednesday"
	case 4:
		return "Thursday"
	case 5:
		return "Friday"
	case 6:
		return "Saturday"
	case 7:
		return "Sunday"
	}
	return ""
}

func Layout_GetDayTextShort(day int) string {
	switch day {
	case 1:
		return "Mon"
	case 2:
		return "Tue"
	case 3:
		return "Wed"
	case 4:
		return "Thu"
	case 5:
		return "Fri"
	case 6:
		return "Sat"
	case 7:
		return "Sun"
	}
	return ""
}

func Layout_ConvertTextTime(unix_sec int64) string {
	tm := time.Unix(unix_sec, 0)
	return fmt.Sprintf("%.02d:%.02d", tm.Hour(), tm.Minute())
}

func Layout_ConvertTextDate(unix_sec int64) string {
	tm := time.Unix(unix_sec, 0)
	//dd := Date_GetDateFromTime(unix_sec)

	switch NewFile_Settings().DateFormat {
	case "eu":
		return fmt.Sprintf("%d/%d/%d", tm.Day(), int(tm.Month()), tm.Year())

	case "us":
		return fmt.Sprintf("%d/%d/%d", int(tm.Month()), tm.Day(), tm.Year())

	case "iso":
		return fmt.Sprintf("%d-%02d-%02d", tm.Year(), int(tm.Month()), tm.Day())

	case "text":
		return fmt.Sprintf("%s %d, %d", Layout_GetMonthText(int(tm.Month())), tm.Day(), tm.Year())

	case "2base":
		return fmt.Sprintf("%d %d-%d", tm.Year(), int(tm.Month()), tm.Day())
	}

	return ""
}
func Layout_ConvertTextDateTime(unix_sec int64) string {
	return Layout_ConvertTextDate(unix_sec) + " " + Layout_ConvertTextTime(unix_sec)
}

func PrepareSearch(search string) []string {

	search = strings.ToLower(search)

	search = strings.ReplaceAll(search, "\n", " ")
	search = strings.ReplaceAll(search, "\t", " ")
	search = strings.ReplaceAll(search, ";", " ")
	search = strings.ReplaceAll(search, ",", " ")
	search = strings.ReplaceAll(search, ".", " ")
	search = strings.ReplaceAll(search, "?", " ")
	search = strings.ReplaceAll(search, "-", " ")

	//split
	words := strings.Split(search, " ")

	//remove empty items
	n := 0
	for i, s := range words {
		if s != "" {
			words[n] = words[i]
			n++
		}
	}
	words = words[:n]

	return words
}
func Search(str string, words []string) bool {

	if len(words) == 0 {
		return true
	}

	str = strings.ToLower(str)

	for _, w := range words {
		if !strings.Contains(str, w) {
			return false
		}
	}
	return true
}

func Layout_MoveElement[T any](array_src *[]T, array_dst *[]T, src int, dst int) {
	//move(by swap one-by-one)
	if array_src == array_dst {
		for i := src; i < dst; i++ {
			(*array_dst)[i], (*array_dst)[i+1] = (*array_dst)[i+1], (*array_dst)[i]
		}
		for i := src; i > dst; i-- {
			(*array_dst)[i], (*array_dst)[i-1] = (*array_dst)[i-1], (*array_dst)[i]
		}
	} else {
		backup := (*array_src)[src]

		//remove
		*array_src = append((*array_src)[:src], (*array_src)[src+1:]...)

		//insert
		if dst < len(*array_dst) {
			*array_dst = append((*array_dst)[:dst+1], (*array_dst)[dst:]...)
			(*array_dst)[dst] = backup
		} else {
			*array_dst = append(*array_dst, backup)
			dst = len(*array_dst) - 1
		}
	}
}
