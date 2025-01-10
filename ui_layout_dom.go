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
	"math"
)

type LayoutPick struct {
	Line       int
	X, Y, W, H int
	Label      string
	Cd         color.RGBA //paintbrush color
	Points     []OsV2
}

func (a *LayoutPick) Cmp(b *LayoutPick) bool {
	return a.Line == b.Line &&
		a.X == b.X &&
		a.Y == b.Y &&
		a.W == b.W &&
		a.H == b.H
}

type Rect struct {
	X, Y, W, H float64
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

type LayoutCmd struct {
	Hash   uint64
	Cmd    string
	Param1 string
	Param2 string
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

	Pick    LayoutPick
	PickApp string
}
type LayoutCR struct {
	Pos int     `json:",omitempty"`
	Min float64 `json:",omitempty"`
	Max float64 `json:",omitempty"`

	Resize_value float64 `json:",omitempty"`

	SetFromChild_min float64 `json:",omitempty"`
	SetFromChild_max float64 `json:",omitempty"`

	Caller_file string `json:",omitempty"`
	Caller_line int    `json:",omitempty"`
}
type LayoutDialog struct {
	Layout *Layout
}
type LayoutScroll struct {
	Hide   bool
	Narrow bool
}

type LayoutDrawRect struct {
	Cd, Cd_over, Cd_down color.RGBA
	Border               float64
}
type LayoutDrawLine struct {
	Cd, Cd_over, Cd_down color.RGBA
	Border               float64
	Sx, Sy, Ex, Ey       float64
}
type LayoutDrawFile struct {
	Cd, Cd_over, Cd_down color.RGBA
	Url                  string
	Align_h              uint8
	Align_v              uint8
}
type LayoutDrawText struct {
	Margin float64

	Cd, Cd_over, Cd_down color.RGBA

	Text  string
	Ghost string

	Align_h uint8
	Align_v uint8

	Formating    bool
	Multiline    bool
	Linewrapping bool
	Selection    bool
	Editable     bool
	Refresh      bool
}
type LayoutDrawCursor struct {
	Name string
}
type LayoutDrawTooltip struct {
	Description string
	Force       bool
}
type LayoutDrawPrim struct {
	Rect Rect

	Rectangle *LayoutDrawRect
	Circle    *LayoutDrawRect
	Line      *LayoutDrawLine
	File      *LayoutDrawFile
	Text      *LayoutDrawText
	Cursor    *LayoutDrawCursor
	Tooltip   *LayoutDrawTooltip
}
type LayoutPaint struct {
	buffer []LayoutDrawPrim
}

type Layout struct {
	X, Y, W, H int
	Name       string

	dialogs []*LayoutDialog
	Childs  []*Layout
	Hash    uint64

	App bool //touch crop

	Enable      bool
	EnableTouch bool
	LLMTip      string

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

	UserCols       []LayoutCR
	UserRows       []LayoutCR
	UserCRFromText *LayoutDrawText

	ScrollV LayoutScroll
	ScrollH LayoutScroll

	Caller_file string `json:",omitempty"`
	Caller_line int    `json:",omitempty"`

	List_auto_spacing bool

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

func _newLayout(x, y, w, h int, name string, parent *Layout) *Layout {
	layout := &Layout{X: x, Y: y, W: w, H: h, Name: name, Enable: true, EnableTouch: true}
	layout.Hash = layout._computeHash(parent)
	return layout
}

func _newLayoutRoot() *Layout {
	root := _newLayout(0, 0, 0, 0, "Root", nil)
	root.App = true
	return root
}

type Layout3 struct {
	ui     *Ui
	parent *Layout3

	props Layout

	touch          bool
	touchDia       bool
	drawEnableFade bool

	dialog *Layout3 //open
	childs []*Layout3

	canvas OsV4
	view   OsV4
	crop   OsV4

	cols UiLayoutArray
	rows UiLayoutArray

	scrollV        UiLayoutScroll
	scrollH        UiLayoutScroll
	scrollOnScreen bool //show scroll all the time

	touchResizeHighlightCol bool
	touchResizeHighlightRow bool
	touchResizeIndex        int
	touchResizeIsActive     bool

	buffer []LayoutDrawPrim
}

func NewUiLayoutDOM(props Layout, parent *Layout3, ui *Ui) *Layout3 {
	dom := &Layout3{ui: ui, parent: parent, props: props}

	dom.scrollV.Init()
	dom.scrollH.Init()

	return dom
}

func NewUiLayoutDOM_root(ui *Ui) *Layout3 {
	return NewUiLayoutDOM(*_newLayoutRoot(), nil, ui)
}

func (dom *Layout3) GetSettings() *UiRootSettings {
	return &dom.ui.settings.Layouts
}

func (dom *Layout3) extractRects(rects map[uint64]Rect) {
	rects[dom.props.Hash] = dom._getRect()

	for _, it := range dom.childs {
		it.extractRects(rects)
	}
}

func (dom *Layout3) project(src *Layout) {
	dom.props = *src

	//scroll
	dom.scrollV.Show = !src.ScrollV.Hide
	dom.scrollH.Show = !src.ScrollH.Hide
	dom.scrollV.Narrow = src.ScrollV.Narrow
	dom.scrollH.Narrow = src.ScrollH.Narrow

	dom.scrollV.wheel = dom.GetSettings().GetScrollV(dom.props.Hash)
	dom.scrollH.wheel = dom.GetSettings().GetScrollH(dom.props.Hash)

	dom.buffer = nil

	//remove un-used
	for i := len(dom.childs) - 1; i >= 0; i-- {
		hash := dom.childs[i].props.Hash
		found := false
		for _, it := range dom.props.Childs {
			if it.Hash == hash {
				found = true
				break
			}
		}
		if !found {
			dom.childs[i].Destroy()
			dom.childs = append(dom.childs[:i], dom.childs[i+1:]...) //remove
		}
	}

	//add new
	for _, src := range dom.props.Childs {

		it := dom.FindChildHash(src.Hash)
		if it == nil {
			it = NewUiLayoutDOM(*src, dom, dom.ui)
			dom.childs = append(dom.childs, it)
		}

		it.project(src)
	}
}

func (dom *Layout3) hasDialog() bool {
	if dom.dialog != nil {
		return true
	}

	for _, it := range dom.childs {
		if it.props.App {
			continue //skip
		}
		if it.hasDialog() {
			return true
		}
	}

	return false
}

func (dom *Layout3) setTouchEnable(parent_touch bool, parent_drawEnableFade bool) {
	dom.touch = parent_touch && dom.props.Enable && dom.props.EnableTouch
	dom.touchDia = true

	dom.drawEnableFade = !parent_drawEnableFade || !dom.props.Enable

	for _, it := range dom.childs {
		it.setTouchEnable(dom.touch, dom.drawEnableFade || !parent_drawEnableFade)
	}
	if dom.dialog != nil {
		dom.dialog.setTouchEnable(true, false)
	}
}

func (dom *Layout3) setTouchDialogDisable(ignoreDia *Layout3) {
	dom.touchDia = false
	for _, it := range dom.childs {
		if it.props.App {
			it.setTouchDialogDisable(nil) //disable dialogs inside
		} else {
			it.setTouchDialogDisable(ignoreDia)
		}
	}

	if dom.dialog != nil && dom.dialog != ignoreDia {
		dom.dialog.setTouchDialogDisable(ignoreDia)
	}
}

func (dom *Layout3) checkDialogs() {
	if dom.dialog != nil {
		if dom.ui.settings.FindDialog(dom.dialog.props.Hash) != nil {
			dom.dialog.checkDialogs()
		} else {
			dom.dialog = nil
		}
	}

	for _, it := range dom.childs {
		it.checkDialogs()
	}
}

func (dom *Layout3) SetTouchAll() {
	dom.checkDialogs()

	dom.setTouchEnable(true, false)

	//dialogs
	for i, dia := range dom.ui.settings.Dialogs {
		layDia := dom.FindHash(dia.Hash)
		if layDia != nil {
			layApp := layDia.parent.GetApp()
			if layApp != nil {

				topApp_i := dom.ui.settings.GetHigherDialogApp(layApp, dom.ui)
				if i == topApp_i {
					layApp.setTouchDialogDisable(layDia)
				}
			}
		}
	}
}

func (dom *Layout3) Destroy() {
	if dom.dialog != nil {
		dom.dialog.Destroy()
	}
	for _, it := range dom.childs {
		it.Destroy()
	}
}

func (dom *Layout3) GetUis() *UiClients {
	return dom.ui.parent
}
func (dom *Layout3) GetIni() *WinIni {
	return &dom.ui.GetWin().io.Ini
}
func (dom *Layout3) GetEnv() *UiEnv {
	return dom.ui.parent.GetEnv()
}

func (dom *Layout3) Cell() int {
	return dom.ui.Cell()
}

func (dom *Layout3) GetServices() *WinServices {
	return dom.ui.GetWin().services
}

func (dom *Layout3) FindChildHash(hash uint64) *Layout3 {

	for _, it := range dom.childs {
		if it.props.Hash == hash {
			return it
		}
	}
	return nil
}

func (dom *Layout3) FindHash(hash uint64) *Layout3 {
	if dom.props.Hash == hash {
		return dom
	}

	if dom.dialog != nil {
		d := dom.dialog.FindHash(hash)
		if d != nil {
			return d
		}
	}

	for _, it := range dom.childs {
		d := it.FindHash(hash)
		if d != nil {
			return d
		}
	}

	return nil
}

func (dom *Layout3) FindFirstName(name string) *Layout3 {
	if dom.props.Name == name {
		return dom
	}

	if dom.dialog != nil {
		d := dom.dialog.FindFirstName(name)
		if d != nil {
			return d
		}
	}

	for _, it := range dom.childs {
		d := it.FindFirstName(name)
		if d != nil {
			return d
		}
	}

	return nil
}

func (dom *Layout3) FindChildMaxArea() *Layout3 {
	var max_layout *Layout3
	var max_area int
	for _, it := range dom.childs {
		area := it.canvas.Area()
		if area > max_area {
			max_layout = it
			max_area = area
		}
	}
	return max_layout
}

func (dom *Layout3) FindShortcut(key byte) *Layout3 {
	if dom.CanTouch() && dom.props.Shortcut_key == key {
		return dom
	}

	if dom.dialog != nil {
		d := dom.dialog.FindShortcut(key)
		if d != nil {
			return d
		}
	}

	for _, it := range dom.childs {
		d := it.FindShortcut(key)
		if d != nil {
			return d
		}
	}

	return nil
}

// It can 'out' multiple(addButton in for loop)!
func (dom *Layout3) FindAppLine(file string, line int, out *[]*Layout3) {
	if dom.props.Caller_file == file && dom.props.Caller_line == line {
		*out = append(*out, dom)
	}

	if dom.dialog != nil {
		dom.dialog.FindAppLine(file, line, out)
	}

	for _, it := range dom.childs {
		it.FindAppLine(file, line, out)
	}
}

func (dom *Layout3) Col(pos int, val float64) {
	dom.cols.findOrAdd(pos).min = float32(val)
}
func (dom *Layout3) Row(pos int, val float64) {
	dom.rows.findOrAdd(pos).min = float32(val)
}

func (dom *Layout3) ColMax(pos int, val float64) {
	dom.cols.findOrAdd(pos).max = float32(val)
}
func (dom *Layout3) RowMax(pos int, val float64) {
	dom.rows.findOrAdd(pos).max = float32(val)
}

func (dom *Layout3) ColResize(pos int, val float64) {
	if val > 0 {
		dom.cols.findOrAdd(pos).resize = &UiLayoutArrayRes{value: float32(val)}
	}
}

func (dom *Layout3) RowResize(pos int, val float64) {
	if val > 0 {
		dom.rows.findOrAdd(pos).resize = &UiLayoutArrayRes{value: float32(val)}
	}
}

func (dom *Layout3) rebuildLevel() {

	winRect := dom.ui.winRect

	dom.canvas = winRect
	dom.view = winRect
	dom.crop = winRect

	if !dom.IsBase() { //dom.IsDialog()
		//get size
		coord := dom.GetLevelSize()

		//coord
		diaS := dom.ui.settings.FindDialog(dom.props.Hash)
		if diaS != nil {
			coord = diaS.GetDialogCoord(coord, dom.ui)
		}

		//set & rebuild with new size
		dom.canvas = coord
		dom.view = coord
		dom.crop = coord
	}
}

func (dom *Layout3) IsShown() bool {
	return dom.parent == nil || (dom.props.W != 0 && dom.props.H != 0)
}

func (dom *Layout3) TouchDialogs(editHash, touchHash uint64) {
	var act *Layout3
	var actE *Layout3
	if editHash != 0 {
		actE = dom.FindHash(editHash)
	}
	if touchHash != 0 {
		act = dom.FindHash(touchHash)
	}

	if dom.ui.GetWin().io.Touch.Wheel != 0 {
		act = dom.FindHash(dom.ui.parent.touch.canvasOver)
	}

	if actE != nil {
		actE.touchComp()
	}
	if act != nil && act != actE {
		act.touchComp()
	}
}

func (dom *Layout3) GetApp() *Layout3 {
	for dom != nil {
		if dom.props.App {
			return dom
		}
		dom = dom.parent
	}
	return dom
}

func (dom *Layout3) IsDialog() bool {
	return dom.ui.settings.FindDialog(dom.props.Hash) != nil
}
func (dom *Layout3) IsBase() bool {
	return dom.parent == nil
}

func (dom *Layout3) IsLevel() bool {
	return dom.IsBase() || dom.IsDialog()
}

func (dom *Layout3) RebuildSoft() {
	if dom.IsDialog() { //is dialog
		dom.rebuildLevel()
	}

	dom.updateCoord(0, 0, 1, 1)

	for _, it := range dom.childs {
		if it.IsShown() {
			it.RebuildSoft()
		}
	}
}

func (dom *Layout3) relayout(setFromSubs bool) {

	if dom.IsLevel() {
		dom.rebuildLevel()
	}

	dom.updateCoord(0, 0, 1, 1)

	if dom.resizeFromPaintText() {
		dom.updateCoord(0, 0, 1, 1)
	}

	if dom.IsDialog() {
		dom.rebuildLevel() //for dialogs, it needs to know dialog size
		dom.updateCoord(0, 0, 1, 1)
	}

	//order List
	dom.rebuildList()

	for _, it := range dom.childs {
		if it.IsShown() {
			it.relayout(setFromSubs)
		}
	}

	if setFromSubs {
		changed := false

		for i, c := range dom.props.UserCols {
			if c.SetFromChild_min > 0 || c.SetFromChild_max > 0 {
				v := 1.0
				for _, it := range dom.childs {
					if it.props.X == c.Pos && it.props.W == 1 {
						v = OsMaxFloat(v, it._getWidth())
					}
				}
				v = OsClampFloat(v, c.SetFromChild_min, c.SetFromChild_max)
				dom.props.UserCols[i].Min = v
				dom.props.UserCols[i].Max = v

				changed = true
			}
		}

		for i, r := range dom.props.UserRows {
			if r.SetFromChild_min > 0 || r.SetFromChild_max > 0 {
				v := 1.0
				for _, it := range dom.childs {
					if it.props.Y == r.Pos && it.props.H == 1 {
						v = OsMaxFloat(v, it._getHeight())
					}
				}
				v = OsClampFloat(v, r.SetFromChild_min, r.SetFromChild_max)
				dom.props.UserRows[i].Min = v
				dom.props.UserRows[i].Max = v

				changed = true
			}
		}

		if changed {
			dom.relayout(false)
		}
	}
}

func (dom *Layout3) rebuildList() {

	if dom.props.Name != "_list" {
		return
	}

	max_width := dom._getWidth()

	//get max item size
	it_width := 0.0
	it_height := 0.0
	for _, it := range dom.childs {
		sz := it.GetLevelSize().Size
		it_width = OsMaxFloat(it_width, float64(sz.X)/float64(dom.Cell()))
		it_height = OsMaxFloat(it_height, float64(sz.Y)/float64(dom.Cell()))
	}

	//num cols/rows
	nx := int(max_width / it_width)
	if nx == 0 {
		nx = 1
	}
	ny := len(dom.childs) / nx
	if len(dom.childs)%nx > 0 {
		ny++
	}

	total_extra_space_w := max_width - float64(nx)*it_width
	space_between := total_extra_space_w / float64(nx-1)
	if !dom.props.List_auto_spacing {
		space_between = 0
	}

	//set cols/rows
	for x := 0; x < nx; x++ {
		dom.props.UserCols = append(dom.props.UserCols, LayoutCR{Pos: x*2 + 0, Min: it_width, Max: it_width})
		//dom.SetColumn(x*2+0, it_width, it_width)
		if x+1 < nx {
			dom.props.UserCols = append(dom.props.UserCols, LayoutCR{Pos: x*2 + 1, Min: 0, Max: space_between})
			//dom.SetColumn(x*2+1, 0, space_between)
		}
	}
	for y := 0; y < ny; y++ {
		dom.props.UserRows = append(dom.props.UserRows, LayoutCR{Pos: y*2 + 0, Min: it_height, Max: it_height})
		//dom.SetRow(y*2+0, it_height, it_height)
		if y+1 < ny {
			dom.props.UserRows = append(dom.props.UserRows, LayoutCR{Pos: y*2 + 1, Min: 0, Max: space_between})
			//dom.SetRow(y*2+1, 0, space_between)
		}
	}

	//set item grid poses
	i := 0
	for y := 0; y < ny; y++ {
		for x := 0; x < nx; x++ {
			if i < len(dom.props.Childs) {
				dom.childs[i].props.X = x * 2
				dom.childs[i].props.Y = y * 2
				dom.childs[i].props.W = 1
				dom.childs[i].props.H = 1
				i++
			}
		}
	}

	//update!
	dom.updateCoord(0, 0, 1, 1)
}

func (dom *Layout3) GetCd(cd, cd_over, cd_down color.RGBA) color.RGBA {
	if dom.CanTouch() {
		active := dom.IsMouseButtonPressed()
		inside := dom.IsMouseInside() && (active || !dom.IsMouseButtonUse())
		if active {
			if inside {
				cd = cd_down
			} else {
				cd = Color2_Aprox(cd_down, cd_over, 0.3)
			}
		} else {
			if inside {
				cd = cd_over
			}
		}
	}
	return cd
}

func (dom *Layout3) getCanvasPx(rect Rect) OsV4 {

	cell := float64(dom.Cell())

	var ret OsV4
	ret.Start.X = dom.canvas.Start.X + int(math.Round(rect.X*cell))
	ret.Start.Y = dom.canvas.Start.Y + int(math.Round(rect.Y*cell))
	ret.Size.X = int(math.Round(rect.W * cell))
	ret.Size.Y = int(math.Round(rect.H * cell))

	return ret
}

func (dom *Layout3) IsCropZero() bool {
	return dom.crop.IsZero()
}

func (dom *Layout3) renderBuffer(buffer []LayoutDrawPrim) {
	if dom.IsCropZero() {
		return
	}

	buff := dom.ui.GetWin().buff

	for _, it := range buffer {
		coord := dom.getCanvasPx(it.Rect)

		if it.Rectangle != nil {
			st := it.Rectangle
			frontCd := dom.GetCd(st.Cd, st.Cd_over, st.Cd_down)
			buff.AddRect(coord, frontCd, dom.ui.CellWidth(st.Border))
		}
		if it.Circle != nil {
			st := it.Circle
			frontCd := dom.GetCd(st.Cd, st.Cd_over, st.Cd_down)
			buff.AddCircle(coord, frontCd, dom.ui.CellWidth(st.Border))
		}
		if it.Line != nil {
			st := it.Line
			frontCd := dom.GetCd(st.Cd, st.Cd_over, st.Cd_down)

			var start, end OsV2
			start.X = coord.Start.X + int(float64(coord.Size.X)*st.Sx)
			start.Y = coord.Start.Y + int(float64(coord.Size.Y)*st.Sy)
			end.X = coord.Start.X + int(float64(coord.Size.X)*st.Ex)
			end.Y = coord.Start.Y + int(float64(coord.Size.Y)*st.Ey)

			buff.AddLine(start, end, frontCd, dom.ui.CellWidth(st.Border))
		}

		if it.File != nil {
			st := it.File
			frontCd := dom.GetCd(st.Cd, st.Cd_over, st.Cd_down)

			var tx, ty, sx, sy float64
			path := InitWinMedia_url(st.Url)
			buff.AddImage(path, coord, frontCd, OsV2{int(st.Align_h), int(st.Align_v)}, &tx, &ty, &sx, &sy, dom.GetPalette().E, dom.Cell())
		}

		if it.Text != nil {
			st := it.Text
			frontCd := dom.GetCd(st.Cd, st.Cd_over, st.Cd_down)

			coord = dom.getCanvasPx(it.Rect.Cut(st.Margin)) //recompute with margin

			prop := InitWinFontPropsDef(dom.Cell())

			prop.formating = st.Formating

			align := OsV2{int(st.Align_h), int(st.Align_v)}
			dom.ui._Text_draw(dom, coord, st.Text, st.Ghost, prop, frontCd, align, st.Selection, st.Editable, st.Multiline, st.Linewrapping)
		}

		if it.Cursor != nil {
			st := it.Cursor
			cq := coord.GetIntersect(dom.crop)
			if dom.CanTouch() && cq.Inside(dom.ui.GetWin().io.Touch.Pos) {
				dom.ui.GetWin().PaintCursor(st.Name)
			}
		}

		if it.Tooltip != nil {
			st := it.Tooltip
			force := st.Force
			if force && !dom.IsTouchActive() {
				force = false
			}
			if dom.CanTouch() && (force || !dom.GetUis().touch.IsActive()) {
				coord := coord.GetIntersect(buff.crop)

				if force {
					dom.ui.tooltip.SetForce(coord, true, st.Description, dom.ui.GetPalette().OnB)
				} else {
					dom.ui.tooltip.Set(coord, false, st.Description, dom.ui.GetPalette().OnB)
				}
			}
		}
	}
}

func (dom *Layout3) _getWidth() float64 {
	return float64(dom.canvas.Size.X) / float64(dom.Cell())
}
func (dom *Layout3) _getHeight() float64 {
	return float64(dom.canvas.Size.Y) / float64(dom.Cell())
}

func (dom *Layout3) _getRect() Rect {
	return Rect{X: 0, Y: 0, W: dom._getWidth(), H: dom._getHeight()}
}

func (dom *Layout3) GetPalette() *WinCdPalette {
	return dom.ui.GetPalette()
}

func (dom *Layout3) GetMouseX() float64 {
	if !dom.CanTouch() {
		return -1
	}

	posCell := dom.ui.GetWin().io.Touch.Pos.Sub(dom.canvas.Start).toV2f().DivV(float32(dom.Cell()))
	return float64(posCell.X)
}
func (dom *Layout3) GetMouseY() float64 {
	if !dom.CanTouch() {
		return -1
	}

	posCell := dom.ui.GetWin().io.Touch.Pos.Sub(dom.canvas.Start).toV2f().DivV(float32(dom.Cell()))
	return float64(posCell.Y)
}

func (dom *Layout3) GetMouseWheel() int {
	if !dom.CanTouch() {
		return 0
	}
	return dom.ui.GetWin().io.Touch.Wheel
}

func (dom *Layout3) IsCtrlPressed() bool {
	if !dom.CanTouch() {
		return false
	}
	return dom.ui.GetWin().io.Keys.Ctrl
}

func (dom *Layout3) IsMouseButtonDownStart() bool {
	if !dom.CanTouch() {
		return false
	}
	return dom.ui.GetWin().io.Touch.Start && dom.IsTouchActive() && !dom.IsCtrlPressed()
}
func (dom *Layout3) GetMouseButtonDown() bool {
	if !dom.CanTouch() {
		return false
	}
	return dom.IsTouchActive() && !dom.IsCtrlPressed()
}
func (dom *Layout3) GetMouseButtonUp() bool {
	if !dom.CanTouch() {
		return false
	}
	return dom.IsTouchEnd() && !dom.IsCtrlPressed()
}

func (dom *Layout3) IsMouseInside() bool {
	tx := dom.GetMouseX()
	ty := dom.GetMouseY()
	return (tx >= 0 && tx < dom._getWidth() && ty >= 0 && ty < dom._getHeight())
}
func (dom *Layout3) IsMouseButtonPressed() bool {
	return dom.IsMouseButtonDownStart() || dom.GetMouseButtonDown() || dom.GetMouseButtonUp()
}
func (dom *Layout3) IsMouseButtonUse() bool {
	if !dom.CanTouch() {
		return false
	}
	return dom.ui.parent.touch.IsCanvasActive() && !dom.IsCtrlPressed()
}

func (layout *Layout3) UpdateTouch() {

	if layout.CanTouch() {
		layout.Touch()
		layout.updateResizer()
	}

	//dialogs
	if layout.dialog != nil {
		layout.dialog.UpdateTouch()
	}

	//subs
	for _, it := range layout.childs {
		if it.IsShown() {
			it.UpdateTouch()
		}
	}
}

func (layout *Layout3) findTouch() *Layout3 {

	var found *Layout3
	if layout.IsMouseInside() {
		found = layout
	}

	//subs
	for _, it := range layout.childs {
		if it.IsShown() {
			l := it.findTouch()
			if l != nil {
				found = l
			}
		}
	}

	return found
}

func (dom *Layout3) resizeFromPaintText() (changed bool) {
	tx := dom.props.UserCRFromText
	if tx != nil {
		value := tx.Text
		if tx.Editable {
			if dom.ui.parent.edit.Is(dom) {
				value = dom.ui.parent.edit.temp
			}
		}

		coord := dom.canvas.Crop(dom.ui.CellWidth(tx.Margin * 2))

		prop := InitWinFontPropsDef(dom.Cell())

		var size OsV2f
		var mx, my int
		if tx.Multiline {
			max_line_px := dom.ui._UiText_getMaxLinePx(coord, tx.Multiline, tx.Linewrapping)
			mx, my = dom.ui.GetTextSizeMax(value, max_line_px, prop)
		} else {
			mx = dom.ui.GetTextSize(-1, value, prop).X
			my = 1
		}
		sizePx := OsV2{mx, my * prop.lineH}
		size = sizePx.toV2f().DivV(float32((dom.Cell()))) //conver into cell
		if !tx.Multiline {
			size.Y -= 0.5 //make space for narrow h-scroll
		}

		//add margin back
		size.X += float32(4 * tx.Margin)
		size.Y += float32(4 * tx.Margin)

		//column
		{
			if len(dom.props.UserCols) == 0 {
				dom.props.UserCols = make([]LayoutCR, 1)
			}
			dom.props.UserCols[0].Min = float64(size.X)
			dom.props.UserCols[0].Max = float64(size.X)

			changed = true
		}

		//row
		if tx.Multiline {
			if len(dom.props.UserRows) == 0 {
				dom.props.UserRows = make([]LayoutCR, 1)
			}
			dom.props.UserRows[0].Min = float64(size.Y)
			dom.props.UserRows[0].Max = float64(size.Y)

			changed = true
		}
	}

	return changed
}

func (layout *Layout3) findBufferText() (Rect, *LayoutDrawText) {
	for _, tx := range layout.buffer {
		if tx.Text != nil {
			return tx.Rect, tx.Text
		}
	}
	return Rect{}, nil
}

func (layout *Layout3) textComp() {

	if layout.CanTouch() {

		rect, tx := layout.findBufferText()
		if tx != nil {
			prop := InitWinFontPropsDef(layout.Cell())
			prop.formating = tx.Formating

			coord := layout.getCanvasPx(rect)
			align := OsV2{int(tx.Align_h), int(tx.Align_v)}

			layout.ui._Text_update(layout, coord, tx.Text, prop, align, tx.Selection, tx.Editable, true, tx.Multiline, tx.Linewrapping, tx.Refresh)
		}
	}

	if layout.dialog != nil {
		layout.dialog.textComp()
	}

	for _, tx := range layout.childs {
		if tx.IsShown() {
			tx.textComp()
		}
	}
}

func (layout *Layout3) touchComp() {
	if layout.CanTouch() {
		var in LayoutInput

		in.Rect = layout._getRect()
		in.X = layout.GetMouseX()
		in.Y = layout.GetMouseY()

		in.IsStart = layout.IsMouseButtonDownStart()
		in.IsActive = layout.IsMouseButtonPressed()
		in.IsEnd = layout.IsTouchEnd()

		in.IsUse = layout.IsMouseButtonUse()
		in.IsUp = layout.GetMouseButtonUp()

		in.IsInside = layout.IsMouseInside()
		in.Wheel = layout.GetMouseWheel()
		in.NumClicks = layout.ui.GetWin().io.Touch.NumClicks
		in.AltClick = layout.ui.GetWin().io.Touch.Rm

		if in.IsStart || in.IsActive || in.Wheel != 0 || in.IsUse || in.IsUp {
			layout.ui.parent.CallInput(&layout.props, &in) //err ...
		}
	}

	//drop file
	if layout.CanTouch() {
		drop_path := layout.GetUis().win.io.Touch.Drop_path
		if layout.IsMouseInside() && drop_path != "" {
			in := LayoutInput{Drop_path: drop_path}
			layout.ui.parent.CallInput(&layout.props, &in) //err ...

		}
	}

	//drag & drop layouts
	if layout.CanTouch() {
		if layout.GetUis().drag.IsOverDrop(layout) {
			if layout.ui.GetWin().io.Touch.End {
				srcDom := layout.ui.dom.FindHash(layout.GetUis().drag.srcHash)
				dstDom := layout.ui.dom.FindHash(layout.GetUis().drag.dstHash)
				if dstDom != nil {
					src_i := srcDom.props.Drag_index
					dst_i := dstDom.props.Drag_index
					dst_i = OsMoveElementIndex(src_i, dst_i, layout.GetUis().drag.pos)

					in := LayoutInput{SetDropMove: true, DropSrc: src_i, DropDst: dst_i}
					layout.ui.parent.CallInput(&dstDom.props, &in) //err ...
				}
			}
		}
	}
}

func (dom *Layout3) Draw() {
	buff := dom.ui.GetWin().buff
	buff.AddCrop(dom.CropWithScroll())
	buff.AddRect(buff.crop, dom.GetPalette().B, 0)

	//base
	dom.drawBuffers()

	//dialogs
	for _, dia := range dom.ui.settings.Dialogs {
		layDia := dom.FindHash(dia.Hash)
		if layDia != nil {
			layApp := layDia.GetApp()
			if layApp != nil {
				//alpha grey background
				backCanvas := layApp.crop
				buff.StartLevel(layDia.CropWithScroll(), dom.ui.GetPalette().B, backCanvas)
			}

			layDia.drawBuffers() //add renderToTexture optimalization ...
		}
	}

	//selection
	dom.ui.selection.Draw(buff, dom.ui)

}

func (dom *Layout3) drawBuffers() {
	buff := dom.ui.GetWin().buff

	buff.AddCrop(dom.CropWithScroll())
	dom._renderScroll()

	buff.AddCrop(dom.crop)

	if dom.props.Back_cd.A > 0 {
		r := dom.crop.Crop(dom.ui.CellWidth(dom.props.Back_margin))
		buff.AddRect(r, dom.props.Back_cd, 0) //background
	}

	if dom.props.Border_cd.A > 0 {
		r := dom.crop.Crop(dom.ui.CellWidth(dom.props.Back_margin))
		buff.AddRect(r, dom.props.Border_cd, dom.ui.CellWidth(0.03)) //background
	}

	dom.renderBuffer(dom.buffer)

	dom.drawResizer()

	//draw alpha rect = disable
	if dom.drawEnableFade && !dom.touch && (dom.parent == nil || dom.parent.touch) {
		buff.AddCrop(dom.crop)
		buff.AddRect(dom.canvas, color.RGBA{255, 255, 255, 150}, 0)
	}

	//subs
	for _, tx := range dom.childs {
		if tx.IsShown() {
			tx.drawBuffers()
		}
	}

	dom.drawDragAndDrop()
}

func (dom *Layout3) drawGrid(cd color.RGBA, w float64, depth int) {
	buff := dom.ui.GetWin().buff

	canvas := dom.canvas.Size

	mx := 1.0 //float64(dom.cols.GetSumOutput(-1)) / float64(canvas.X)
	my := 1.0 //float64(dom.rows.GetSumOutput(-1)) / float64(canvas.Y)

	var start, end OsV2
	width := dom.ui.CellWidth(w)

	cr := dom.crop.Crop(depth * width)
	buff.AddCrop(cr)

	//main border
	//rc := dom._getRect()
	//rc = rc.Cut(float64(depth) * 0.06)
	buff.AddRect(cr, cd, width)

	//columns
	start = dom.canvas.Start
	end = dom.canvas.End()
	end.Y = dom.canvas.Start.Y + int(float64(dom.canvas.Size.Y)*my)
	sum := int32(0)
	for _, c := range dom.cols.outputs {
		sum += c
		p := float64(sum) / float64(canvas.X)

		start.X = dom.canvas.Start.X + int(float64(dom.canvas.Size.X)*p)
		end.X = start.X
		buff.AddLine(start, end, cd, width)
	}
	//rest
	for start.X < dom.canvas.End().X {
		start.X += dom.Cell()
		end.X = start.X
		buff.AddLine(start, end, cd, width)
	}

	//rows
	sum = 0
	start = dom.canvas.Start
	end = dom.canvas.End()
	end.X = dom.canvas.Start.X + int(float64(dom.canvas.Size.X)*mx)
	for _, r := range dom.rows.outputs {
		sum += r
		p := float64(sum) / float64(canvas.Y)

		start.Y = dom.canvas.Start.Y + int(float64(dom.canvas.Size.Y)*p)
		end.Y = start.Y
		buff.AddLine(start, end, cd, width)
	}
	//rest
	for start.Y < dom.canvas.End().Y {
		start.Y += dom.Cell()
		end.Y = start.Y
		buff.AddLine(start, end, cd, width)
	}
}

func (dom *Layout3) drawDragAndDrop() {
	if !dom.CanTouch() {
		return
	}

	buff := dom.ui.GetWin().buff

	drag := &dom.GetUis().drag

	//activate
	if dom.props.Drag_group != "" && dom.IsTouchActiveSubs() {
		drag.Set(dom)
	}
	isDragged := drag.IsDraged(dom)
	isDrop := drag.IsOverDrop(dom)
	if isDragged || isDrop {
		buff.AddCrop(dom.crop)

		borderWidth := dom.ui.CellWidth(0.1)
		cd := dom.GetPalette().OnB
		cd.A = 100

		//draw drag
		if isDragged {
			buff.AddRect(dom.crop.Crop(borderWidth), cd, 0)
		}

		//draw drop
		if isDrop && dom.IsOver() {

			pos := SA_Drop_INSIDE

			r := dom.ui.GetWin().io.Touch.Pos.Sub(dom.crop.Middle())

			if dom.props.Drop_v && dom.props.Drop_h {
				arx := float32(OsAbs(r.X)) / float32(dom.crop.Size.X)
				ary := float32(OsAbs(r.Y)) / float32(dom.crop.Size.Y)
				if arx > ary {
					if r.X < 0 {
						pos = SA_Drop_H_LEFT
					} else {
						pos = SA_Drop_H_RIGHT
					}
				} else {
					if r.Y < 0 {
						pos = SA_Drop_V_LEFT
					} else {
						pos = SA_Drop_V_RIGHT
					}
				}
			} else if dom.props.Drop_v {
				if r.Y < 0 {
					pos = SA_Drop_V_LEFT
				} else {
					pos = SA_Drop_V_RIGHT
				}
			} else if dom.props.Drop_h {
				if r.X < 0 {
					pos = SA_Drop_H_LEFT
				} else {
					pos = SA_Drop_H_RIGHT
				}
			}

			//update
			drag.pos = pos
			drag.dstHash = dom.props.Hash

			//paint
			wx := float64(borderWidth) / float64(dom.canvas.Size.X)
			wy := float64(borderWidth) / float64(dom.canvas.Size.Y)
			switch pos {
			case SA_Drop_INSIDE:
				buff.AddRect(dom.crop, cd, borderWidth) //full rect
			case SA_Drop_V_LEFT:
				buff.AddRect(dom.crop.Cut(0, 0, 1, wy), cd, 0)
			case SA_Drop_V_RIGHT:
				buff.AddRect(dom.crop.Cut(0, 1-wy, 1, wy), cd, 0)
			case SA_Drop_H_LEFT:
				buff.AddRect(dom.crop.Cut(0, 0, wx, 1), cd, 0)
			case SA_Drop_H_RIGHT:
				buff.AddRect(dom.crop.Cut(1-wx, 0, wx, 1), cd, 0)
			}
		}
	}
}

func (dom *Layout3) Touch() {
	startTouch := dom.CanTouch() && dom.ui.GetWin().io.Touch.Start && !dom.IsCtrlPressed()
	over := dom.CanTouch() && dom.IsTouchPosInside() && !dom.GetUis().touch.IsResizeActive()

	if over && dom.CanTouch() {
		if startTouch {
			if !dom.GetUis().touch.IsScrollOrResizeActive() { //if lower resize or scroll is activated than don't rewrite it with higher canvas
				dom.GetUis().touch.Set(dom.props.Hash, 0, 0, 0)
			}
		}

		dom.GetUis().touch.canvasOver = dom.props.Hash
	}

	dom.touchScroll()
}

func (dom *Layout3) _isTouchCtrl() bool {
	return dom.CanTouch() && dom.IsCtrlPressed() && dom.IsTouchPosInside() //dom.ui.levels.IsLevelTop(dom)
}

func (dom *Layout3) IsTouchCtrlComp_start() bool {
	return dom._isTouchCtrl() && !dom.ui.GetWin().io.Touch.Rm && dom.ui.GetWin().io.Touch.Start
}
func (dom *Layout3) IsTouchCtrlComp_end() bool {
	return dom._isTouchCtrl() && !dom.ui.GetWin().io.Touch.Rm && dom.ui.GetWin().io.Touch.End
}

func (dom *Layout3) IsTouchCtrlGrid_start() bool {
	return dom._isTouchCtrl() && dom.ui.GetWin().io.Touch.Rm && dom.ui.GetWin().io.Touch.Start
}
func (dom *Layout3) IsTouchCtrlGrid_end() bool {
	return dom._isTouchCtrl() && dom.ui.GetWin().io.Touch.Rm && dom.ui.GetWin().io.Touch.End
}

func (dom *Layout3) CropWithScroll() OsV4 {
	crop := dom.crop

	if dom.scrollV.Is() {
		crop.Size.X += dom.scrollV._GetWidth(dom.ui)
	}

	if dom.scrollH.Is() {
		crop.Size.Y += dom.scrollH._GetWidth(dom.ui)
	}

	return crop
}

func (dom *Layout3) _renderScroll() {
	showBackground := dom.scrollOnScreen

	if dom.scrollV.Is() {
		scrollQuad := dom.scrollV.GetScrollBackCoordV(dom.view, dom.ui)
		dom.scrollV.DrawV(scrollQuad, showBackground, dom.ui)
	}

	if dom.scrollH.Is() {
		scrollQuad := dom.scrollH.GetScrollBackCoordH(dom.view, dom.ui)
		dom.scrollH.DrawH(scrollQuad, showBackground, dom.ui)
	}
}

func (dom *Layout3) isTouchScroll() (bool, bool) {
	insideScrollV := false
	insideScrollH := false
	if dom.scrollV.Is() {
		scrollQuad := dom.scrollV.GetScrollBackCoordV(dom.view, dom.ui)
		insideScrollV = scrollQuad.Inside(dom.ui.GetWin().io.Touch.Pos)
	}

	if dom.scrollH.Is() {
		scrollQuad := dom.scrollH.GetScrollBackCoordH(dom.view, dom.ui)
		insideScrollH = scrollQuad.Inside(dom.ui.GetWin().io.Touch.Pos)
	}
	return insideScrollV, insideScrollH
}

func (dom *Layout3) IsTouchPosInside() bool {
	return dom.crop.Inside(dom.ui.GetWin().io.Touch.Pos)
}

func (dom *Layout3) IsOver() bool {
	return dom.CanTouch() && dom.IsTouchPosInside() && !dom.GetUis().touch.IsResizeActive()
}

func (dom *Layout3) IsOverScroll() bool {
	insideScrollV, insideScrollH := dom.isTouchScroll()
	return dom.CanTouch() && (insideScrollV || insideScrollH)
}

func (dom *Layout3) IsTouchActive() bool {
	return dom.GetUis().touch.IsFnMove(dom.props.Hash, 0, 0, 0)
}

func (dom *Layout3) IsTouchActiveSubs() bool {
	if dom.IsTouchActive() {
		return true
	}
	for _, tx := range dom.childs {
		if tx.IsTouchActiveSubs() {
			return true
		}
	}
	return false
}

func (dom *Layout3) IsTouchInside() bool {
	inside := dom.IsOver()

	if !dom.IsTouchActive() && dom.CanTouch() && dom.GetUis().touch.IsActive() { // when click and move, other Buttons, etc. are disabled
		inside = false
	}

	return inside
}

func (dom *Layout3) IsTouchEnd() bool {
	return dom.CanTouch() && dom.ui.GetWin().io.Touch.End && dom.IsTouchActive() //doesn't have to be inside!
}

func (dom *Layout3) IsClicked(enable bool) (int, int, bool, bool, bool) {
	var click, rclick int
	var inside, active, end bool
	if enable {
		inside = dom.IsTouchInside()
		active = dom.IsTouchActive()
		end = dom.IsTouchEnd()

		touch := &dom.ui.GetWin().io.Touch

		if inside && end {
			click = 1
			rclick = OsTrn(touch.Rm, 1, 0)
		}

		if click > 0 {
			click = touch.NumClicks
		}
		if rclick > 0 {
			rclick = touch.NumClicks
		}
	}

	return click, rclick, inside, active, end
}

func (dom *Layout3) CanTouch() bool {
	return dom.touch && dom.touchDia
}

func (dom *Layout3) touchScroll() {
	hasScrollV := dom.scrollV.Is() //IsPure()
	hasScrollH := dom.scrollH.Is() //IsPure()

	enableInput := dom.CanTouch()

	//redraw := false
	if hasScrollV {
		if enableInput {
			if dom.scrollV.TouchV(dom) {
				dom.RebuildSoft()

				dom.GetSettings().SetScrollV(dom.props.Hash, dom.scrollV.wheel)
			}
		}
	}

	if hasScrollH {
		if enableInput {
			if dom.scrollH.TouchH(hasScrollV, dom) {
				dom.RebuildSoft()

				dom.GetSettings().SetScrollH(dom.props.Hash, dom.scrollH.wheel)
			}
		}
	}
}

func (dom *Layout3) convert(grid OsV4) OsV4 {
	cell := dom.Cell()

	c := dom.cols.Convert(cell, grid.Start.X, grid.Start.X+grid.Size.X)
	r := dom.rows.Convert(cell, grid.Start.Y, grid.Start.Y+grid.Size.Y)

	return OsV4{OsV2{c.X, r.X}, OsV2{c.Y, r.Y}}
}

func (dom *Layout3) ConvertMax(cell int, in OsV4) OsV4 {
	c := dom.cols.ConvertMax(cell, in.Start.X, in.Start.X+in.Size.X)
	r := dom.rows.ConvertMax(cell, in.Start.Y, in.Start.Y+in.Size.Y)

	return OsV4{OsV2{c.X, r.X}, OsV2{c.Y, r.Y}}
}

func (dom *Layout3) updateColsRows() {

	//reset
	dom.cols.Clear()
	dom.rows.Clear()

	//project
	for _, it := range dom.props.UserCols {
		if it.Resize_value > 0 {
			it.Resize_value = dom.GetSettings().GetCol(dom.props.Hash, it.Pos, it.Resize_value)
		}

		dom.Col(it.Pos, it.Min)
		dom.ColMax(it.Pos, it.Max)
		dom.ColResize(it.Pos, it.Resize_value)
	}
	for _, it := range dom.props.UserRows {
		if it.Resize_value > 0 {
			it.Resize_value = dom.GetSettings().GetRow(dom.props.Hash, it.Pos, it.Resize_value)
		}
		dom.Row(it.Pos, it.Min)
		dom.RowMax(it.Pos, it.Max)
		dom.RowResize(it.Pos, it.Resize_value)
	}
}

func (dom *Layout3) updateCoord(rx, ry, rw, rh float64) {

	dom.updateColsRows()

	isLevel := dom.IsLevel()
	if !isLevel {
		//	data := parent.convert(OsV4{OsV2{}, gridMax}).Size
		dom.view = dom.parent.convert(InitOsV4(dom.props.X, dom.props.Y, dom.props.W, dom.props.H))
		dom.view.Start = dom.parent.view.Start.Add(dom.view.Start)

		dom.view.Start.X += int(float64(dom.view.Size.X) * rx)
		dom.view.Start.Y += int(float64(dom.view.Size.Y) * ry)
		dom.view.Size.X = int(float64(dom.view.Size.X) * rw)
		dom.view.Size.Y = int(float64(dom.view.Size.Y) * rh)

		// move start by scroll
		dom.view.Start.X -= dom.parent.scrollH.GetWheel() //parent scroll
		dom.view.Start.Y -= dom.parent.scrollV.GetWheel()
	}

	// crop
	if !isLevel {
		dom.crop = dom.view.GetIntersect(dom.parent.crop)
	}

	//slow ...
	{
		makeSmallerX := dom.scrollV.Show
		makeSmallerY := dom.scrollH.Show
		gridMax := dom.GetGridMax(OsV2{})
		screen := dom.view.Size
		for dom.updateGridAndScroll(&screen, gridMax, &makeSmallerX, &makeSmallerY) {
		}
		dom.view.Size = screen
	}

	// crop
	if dom.scrollV.Is() {
		dom.crop.Size.X = OsMax(0, dom.crop.Size.X-dom.scrollV._GetWidth(dom.ui))
	}
	if dom.scrollH.Is() {
		dom.crop.Size.Y = OsMax(0, dom.crop.Size.Y-dom.scrollH._GetWidth(dom.ui))
	}

	dom.canvas.Start = dom.view.Start
	dom.canvas.Start.X -= dom.scrollH.GetWheel() //this scroll
	dom.canvas.Start.Y -= dom.scrollV.GetWheel()

	dom.canvas.Size.X = dom.cols.OutputAll()
	dom.canvas.Size.Y = dom.rows.OutputAll()
	dom.canvas = dom.canvas.Extend(dom.view)
}

func (dom *Layout3) GetGridMax(minSize OsV2) OsV2 {
	mx := minSize
	for _, tx := range dom.childs {
		if tx.IsShown() {
			mx = mx.Max(OsV2{X: tx.props.X + tx.props.W, Y: tx.props.Y + tx.props.H})
		}
	}

	mx = mx.Max(OsV2{dom.cols.NumInputs(), dom.rows.NumInputs()})
	return mx
}

func (dom *Layout3) updateArray(window OsV2, endGrid OsV2) {
	if endGrid.X > dom.cols.NumInputs() {
		dom.cols.Resize(int(endGrid.X))
	}
	if endGrid.Y > dom.rows.NumInputs() {
		dom.rows.Resize(int(endGrid.Y))
	}

	dom.cols.SetFills(dom.childs, true)
	dom.rows.SetFills(dom.childs, false)

	cell := dom.Cell()
	dom.cols.Update(cell, window.X)
	dom.rows.Update(cell, window.Y)
}

func (dom *Layout3) updateGridAndScroll(screen *OsV2, gridMax OsV2, makeSmallerX *bool, makeSmallerY *bool) bool {

	// update cols/rows
	dom.updateArray(*screen, gridMax)

	// get max
	data := dom.convert(OsV4{OsV2{}, gridMax}).Size

	// make canvas smaller
	hasScrollV := OsTrnBool(*makeSmallerX, data.Y > screen.Y, false)
	hasScrollH := OsTrnBool(*makeSmallerY, data.X > screen.X, false)
	if hasScrollV {
		screen.X -= dom.scrollV._GetWidth(dom.ui)
		*makeSmallerX = false
	}
	if hasScrollH {
		screen.Y -= dom.scrollH._GetWidth(dom.ui)
		*makeSmallerY = false
	}

	// save to scroll
	dom.scrollV.data_height = data.Y
	dom.scrollV.screen_height = screen.Y

	dom.scrollH.data_height = data.X
	dom.scrollH.screen_height = screen.X

	return hasScrollV || hasScrollH
}

func (dom *Layout3) GetLevelSize() OsV4 {

	dom.updateColsRows() //project .userCols -> .cols

	q := OsV4{OsV2{}, dom.GetGridMax(OsV2{1, 1})}

	q.Size = dom.ConvertMax(dom.Cell(), q).Size

	q.Start = dom.ui.winRect.Start
	q = q.GetIntersect(dom.ui.winRect)
	return q
}

func (layout *Layout3) ScrollIntoTop_vertical() {
	layout.scrollV.SetWheel(0)
	layout.GetSettings().SetScrollV(layout.props.Hash, 0)

}
func (layout *Layout3) ScrollIntoTop_horizontal() {
	layout.scrollH.SetWheel(0)
	layout.GetSettings().SetScrollH(layout.props.Hash, 0)
}
func (layout *Layout3) ScrollIntoBottom_vertical() {
	layout.scrollV.SetWheel(100000000) //100M
	layout.GetSettings().SetScrollV(layout.props.Hash, 100000000)
}
func (layout *Layout3) ScrollIntoBottom_horizontal() {
	layout.scrollH.SetWheel(100000000)
	layout.GetSettings().SetScrollH(layout.props.Hash, 100000000)
}
