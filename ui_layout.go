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
	"strings"
)

type LayoutInput struct {
	Rect Rect

	IsStart  bool
	IsActive bool
	IsEnd    bool

	IsInside bool //rename IsOver ....
	IsUse    bool //IsActive? ....
	IsUp     bool //IsEnd? ....

	X, Y      float64
	Wheel     int
	NumClicks int
	AltClick  bool

	SetEdit   bool
	EditValue string
	EditEnter bool

	SetDropMove                    bool
	DragSrc_source, DragDst_source string
	DropSrc_pos, DropDst_pos       int

	Drop_path string

	Shortcut_key byte

	Pick LayoutPick
}
type LayoutCR struct {
	Pos int
	Min float64
	Max float64

	Resize_value float64

	SetFromChild_min float64
	SetFromChild_max float64
	SetFromChild_fix bool
}

func (cr *LayoutCR) IsFromChild() bool {
	return cr.SetFromChild_min > 0 || cr.SetFromChild_max > 0
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
	Radius               float64
}
type LayoutDrawLine struct {
	Cd, Cd_over, Cd_down color.RGBA
	Border               float64
	Sx, Sy, Ex, Ey       float64
}
type LayoutDrawFile struct {
	Cd, Cd_over, Cd_down color.RGBA

	ImagePath WinImagePath

	Align_h uint8
	Align_v uint8
}
type LayoutDrawText struct {
	Margin [4]float64 //Top, Bottom, Left, Right

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
	Password     bool
}

type LayoutDrawCursor struct {
	Name string
}
type LayoutDrawTooltip struct {
	Description string
	Force       bool
}

type LayoutDrawBrush struct {
	Cd     color.RGBA
	Points []OsV2
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
	Brush     *LayoutDrawBrush
}
type LayoutPaint struct {
	buffer []LayoutDrawPrim
}

func (layout *Layout) _recomputeHash() {
	h := sha256.New()

	//parent
	if layout.parent != nil {
		var pt [8]byte
		binary.LittleEndian.PutUint64(pt[:], layout.parent.UID)
		h.Write(pt[:])
	}

	//this
	h.Write(fmt.Appendf(nil, "%s(%d,%d,%d,%d)", layout.Name, layout.X, layout.Y, layout.W, layout.H))

	layout.UID = binary.LittleEndian.Uint64(h.Sum(nil))
}
func (layout *Layout) _build() {
	if layout.fnBuild != nil {
		layout.fnBuild(layout)
	}
	for _, it := range layout.childs {
		it._build()
	}
	for _, it := range layout.dialogs {
		it.Layout._build()
	}
}

func _newLayout(x, y, w, h int, name string, parent *Layout) *Layout {
	var ui *Ui
	if parent != nil {
		ui = parent.ui
	}

	layout := &Layout{X: x, Y: y, W: w, H: h, Name: name, Enable: true, EnableTouch: true, parent: parent, ui: ui}

	layout.scrollV.Init()
	layout.scrollH.Init()

	layout._recomputeHash()

	return layout
}

type LayoutTooltip struct {
	X, Y, W, H int
	Tooltip    string
}

// whole layout must be inside, not partially inside!
func (tip *LayoutTooltip) InInside(layout *Layout) bool {
	return tip.X <= layout.X &&
		tip.Y <= layout.Y &&
		tip.X+tip.W >= layout.X+layout.W &&
		tip.Y+tip.H >= layout.Y+layout.H
}

type Layout struct {
	ui     *Ui
	parent *Layout

	AppName  string
	ToolName string

	touch          bool
	touchDia       bool
	drawEnableFade bool

	dialog  *Layout //open
	childs  []*Layout
	dialogs []*LayoutDialog

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

	fnBuild       func(*Layout)
	fnDraw        func(Rect, *Layout) LayoutPaint
	fnInput       func(LayoutInput, *Layout)
	fnHasShortcut func(byte) bool
	fnSetEditbox  func(string, bool)
	fnGetEditbox  func() string

	dropMove func(src_i, dst_i int, aim_i int, src_source, dst_source string)
	dropFile func(path string)

	X, Y, W, H int
	Name       string

	Tooltip       string
	TooltipGroups []LayoutTooltip

	UID uint64

	App bool //touch crop

	Enable      bool
	EnableTouch bool

	Back_cd       color.RGBA
	Back_rounding bool
	Back_margin   float64
	Border_cd     color.RGBA

	Drag_group              string
	Drop_group              string
	Drag_source             string
	Drag_index              int
	Drop_h, Drop_v, Drop_in bool

	Editbox_name string

	UserCols []LayoutCR
	UserRows []LayoutCR

	fnAutoResize          func(layout *Layout)
	fnGetAutoResizeMargin func() [4]float64

	Cards_autoSpacing bool

	fnUpdate func() //here

	fnGetLLMTip func(layout *Layout) string
}

func NewUiLayoutDOM_root(ui *Ui) *Layout {
	root := _newLayout(0, 0, 0, 0, "Root", nil)
	root.UID = 1
	root.ui = ui
	root.App = true

	root.AppName = "Root"
	root.ToolName = "ShowRoot"
	return root
}

func (layout *Layout) GetSettings() *UiRootSettings {
	return &layout.ui.settings.Layouts
}

func (layout *Layout) GetPalette() *DevPalette {
	return layout.ui.router.services.sync.GetPalette()
}

func (layout *Layout) projectScroll() {
	layout.scrollV.wheel = layout.GetSettings().GetScrollV(layout.UID)
	layout.scrollH.wheel = layout.GetSettings().GetScrollH(layout.UID)

	for _, src := range layout.childs {
		src.projectScroll()
	}

	if layout.dialog != nil {
		layout.dialog.projectScroll()
	}
}

func (layout *Layout) setTouchEnable(parent_touch bool, parent_drawEnableFade bool) {
	layout.touch = parent_touch && layout.Enable && layout.EnableTouch
	layout.touchDia = true

	layout.drawEnableFade = !parent_drawEnableFade && !layout.Enable

	for _, it := range layout.childs {
		it.setTouchEnable(layout.touch, layout.drawEnableFade || parent_drawEnableFade)
	}
	if layout.dialog != nil {
		layout.dialog.setTouchEnable(true, false)
	}
}

func (layout *Layout) setTouchDialogDisable(ignoreDia *Layout) {
	layout.touchDia = false
	for _, it := range layout.childs {
		if it.App {
			it.setTouchDialogDisable(nil) //disable dialogs inside
		} else {
			it.setTouchDialogDisable(ignoreDia)
		}
	}

	if layout.dialog != nil && layout.dialog != ignoreDia {
		layout.dialog.setTouchDialogDisable(ignoreDia)
	}
}

func (layout *Layout) checkDialogs() {
	if layout.dialog != nil {
		if layout.ui.settings.FindDialog(layout.dialog.UID) != nil {
			layout.dialog.checkDialogs()
		} else {
			layout.dialog = nil
		}
	}

	for _, it := range layout.childs {
		it.checkDialogs()
	}
}

func (layout *Layout) SetTouchAll() {
	layout.checkDialogs()

	layout.setTouchEnable(true, false)

	//dialogs
	for i, dia := range layout.ui.settings.Dialogs {
		layDia := layout.FindUID(dia.UID)
		if layDia != nil {
			layApp := layDia.parent.GetApp()
			if layApp != nil {

				topApp_i := layout.ui.settings.GetHigherDialogApp(layApp, layout.ui)
				if i == topApp_i {
					layApp.setTouchDialogDisable(layDia)
				}
			}
		}
	}
}

func (layout *Layout) Destroy() {
	if layout.dialog != nil {
		layout.dialog.Destroy()
	}
	for _, it := range layout.childs {
		it.Destroy()
	}

	layout.dialog = nil
	layout.dialogs = nil
	layout.cols.Clear()
	layout.rows.Clear()
	layout.childs = nil

}

func (layout *Layout) Cell() int {
	return layout.ui.Cell()
}

func (layout *Layout) FindGrid(x, y, w, h int) *Layout {
	for _, it := range layout.childs {
		if it.X == x && it.Y == y && it.W == w && it.H == h {
			return it
		}
	}
	return nil
}

func (layout *Layout) FindChildHash(uid uint64) *Layout {

	for _, it := range layout.childs {
		if it.UID == uid {
			return it
		}
	}
	return nil
}

func (layout *Layout) FindUID(UID uint64) *Layout {
	if layout.UID == UID {
		return layout
	}

	for _, dia := range layout.dialogs {
		d := dia.Layout.FindUID(UID)
		if d != nil {
			return d
		}
	}

	for _, it := range layout.childs {
		d := it.FindUID(UID)
		if d != nil {
			return d
		}
	}

	return nil
}

func (layout *Layout) FindEditbox(name string) *Layout {
	if layout.Editbox_name == name {
		return layout
	}

	if layout.dialog != nil {
		d := layout.dialog.FindEditbox(name)
		if d != nil {
			return d
		}
	}

	for _, it := range layout.childs {
		d := it.FindEditbox(name)
		if d != nil {
			return d
		}
	}

	return nil
}

func (layout *Layout) FindChildMaxArea() *Layout {
	var max_layout *Layout
	var max_area int
	for _, it := range layout.childs {
		area := it.canvas.Area()
		if area > max_area {
			max_layout = it
			max_area = area
		}
	}
	return max_layout
}

func (layout *Layout) FindShortcut(key byte) *Layout {
	if layout.CanTouch() && layout.fnHasShortcut != nil && layout.fnHasShortcut(key) {
		return layout
	}

	if layout.dialog != nil {
		d := layout.dialog.FindShortcut(key)
		if d != nil {
			return d
		}
	}

	for _, it := range layout.childs {
		d := it.FindShortcut(key)
		if d != nil {
			return d
		}
	}

	return nil
}

func (layout *Layout) Col(pos int, val float64) {
	layout.cols.findOrAdd(pos).min = val
}
func (layout *Layout) Row(pos int, val float64) {
	layout.rows.findOrAdd(pos).min = val
}

func (layout *Layout) ColMax(pos int, val float64) {
	layout.cols.findOrAdd(pos).max = val
}
func (layout *Layout) RowMax(pos int, val float64) {
	layout.rows.findOrAdd(pos).max = val
}

func (layout *Layout) ColResize(pos int, val float64) {
	if val > 0 {
		layout.cols.findOrAdd(pos).resize = &UiLayoutArrayRes{value: val}
	}
}

func (layout *Layout) RowResize(pos int, val float64) {
	if val > 0 {
		layout.rows.findOrAdd(pos).resize = &UiLayoutArrayRes{value: val}
	}
}

func (layout *Layout) rebuildLevel() bool {

	winRect := layout.ui.winRect //window

	layout.canvas = winRect
	layout.view = winRect
	layout.crop = winRect

	resized := false
	if !layout.IsBase() { //dom.IsDialog()
		//get size
		coord := layout.GetLevelSize()

		//coord
		diaS := layout.ui.settings.FindDialog(layout.UID)
		if diaS != nil {
			new_coord := diaS.GetDialogCoord(coord, layout.ui)
			resized = !coord.Size.Cmp(new_coord.Size)
			coord = new_coord
		}

		//set & rebuild with new size
		layout.canvas = coord
		layout.view = coord
		layout.crop = coord
	}

	return resized
}

func (layout *Layout) IsShown() bool {
	return layout.parent == nil || (layout.W != 0 && layout.H != 0)
}

func (layout *Layout) TouchDialogs(editUID, touchUID uint64) {
	var act *Layout
	var actE *Layout
	if editUID != 0 {
		actE = layout.FindUID(editUID)
	}
	if touchUID != 0 {
		act = layout.FindUID(touchUID)
	}

	if layout.ui.GetWin().io.Touch.Wheel != 0 {
		act = layout.FindUID(layout.ui.touch.canvasOver)
	}

	if actE != nil {
		actE.touchComp()
	}
	if act != nil && act != actE {
		act.touchComp()
	}
}

func (layout *Layout) GetApp() *Layout {
	for layout != nil {
		if layout.App {
			return layout
		}
		layout = layout.parent
	}
	return layout
}

func (layout *Layout) IsDialog() bool {
	return layout.ui.settings.FindDialog(layout.UID) != nil
}
func (layout *Layout) IsBase() bool {
	return layout.parent == nil
}

func (layout *Layout) IsLevel() bool {
	return layout.IsBase() || layout.IsDialog()
}

func (layout *Layout) updateFromChildCols(enableFix bool) (float64, float64) {
	//reset
	for i, c := range layout.UserCols {
		if c.IsFromChild() {
			layout.UserCols[i].Min = 0
			layout.UserCols[i].Max = 0
		}
	}

	//childs
	for _, it := range layout.childs {
		if it.IsShown() {
			min, max := it.updateFromChildCols(enableFix)

			if it.W == 1 {
				for i, c := range layout.UserCols {
					if c.IsFromChild() && it.X == c.Pos {

						layout.UserCols[i].Min = OsClampFloat(OsMaxFloat(c.Min, min), c.SetFromChild_min, c.SetFromChild_max)
						if !enableFix {
							layout.UserCols[i].Max = OsClampFloat(OsMaxFloat(c.Max, max), c.SetFromChild_min, c.SetFromChild_max)
						} else {

							if c.SetFromChild_fix {
								layout.UserCols[i].Max = layout.UserCols[i].Min
							} else {
								layout.UserCols[i].Max = c.SetFromChild_max
							}
						}
					}
				}
			}
		}
	}

	//repair
	for i, c := range layout.UserCols {
		if c.IsFromChild() && c.Min == 0 && c.Max == 0 { //un-changed
			layout.UserCols[i].Min = c.SetFromChild_min
			layout.UserCols[i].Max = c.SetFromChild_max
		}
	}

	if layout.dialog != nil {
		layout.dialog.updateFromChildCols(enableFix)
	}

	//update min/max size of components
	if layout.fnAutoResize != nil {
		layout.UserCols = nil
		layout.UserRows = nil

		if !enableFix {
			layout.canvas = OsV4{} //only when update cols, rows must know width!
		}

		layout.fnAutoResize(layout)
	}

	var min_px, max_px int
	{
		max_grid := 0
		for _, tx := range layout.childs {
			if tx.IsShown() {
				max_grid = OsMax(max_grid, tx.X+tx.W)
			}
		}
		min_px = layout.ui.CellWidth(float64(max_grid))
		max_px = min_px

		for _, c := range layout.UserCols {
			min_px += layout.ui.CellWidth(c.Min)
			max_px += layout.ui.CellWidth(c.Max)
			if c.Pos < max_grid {
				min_px -= layout.Cell()
				max_px -= layout.Cell()
			}
		}
	}

	if layout.scrollV.Is() {
		min_px += layout.scrollV._GetWidth(layout.ui)
		max_px += layout.scrollV._GetWidth(layout.ui)
	}
	if min_px > max_px {
		max_px = min_px
	}

	return float64(min_px) / float64(layout.Cell()), float64(max_px) / float64(layout.Cell())
}

func (layout *Layout) updateFromChildRows() (float64, float64) {
	//reset
	for i, r := range layout.UserRows {
		if r.IsFromChild() {
			layout.UserRows[i].Min = 0
			layout.UserRows[i].Max = 0
		}
	}

	//childs
	for _, it := range layout.childs {
		if it.IsShown() {
			min, _ := it.updateFromChildRows()

			if it.H == 1 {
				for i, r := range layout.UserRows {
					if r.IsFromChild() && it.Y == r.Pos {
						layout.UserRows[i].Min = OsClampFloat(OsMaxFloat(r.Min, min), r.SetFromChild_min, r.SetFromChild_max)
						if r.SetFromChild_fix {
							layout.UserRows[i].Max = layout.UserRows[i].Min
						} else {
							layout.UserRows[i].Max = r.SetFromChild_max
							//layout.UserRows[i].Max = OsClampFloat(OsMaxFloat(r.Max, max), r.SetFromChild_min, r.SetFromChild_max)
						}
					}
				}
			}
		}
	}

	//repair
	for i, r := range layout.UserRows {
		if r.IsFromChild() && r.Min == 0 && r.Max == 0 { //un-changed
			layout.UserRows[i].Min = r.SetFromChild_min
			layout.UserRows[i].Max = r.SetFromChild_max
		}
	}

	if layout.dialog != nil {
		layout.dialog.updateFromChildRows()
	}

	//update min/max size of components
	if layout.fnAutoResize != nil {
		layout.UserCols = nil
		layout.UserRows = nil

		layout.fnAutoResize(layout)
	}

	var min_px, max_px int
	{
		max_grid := 0
		for _, tx := range layout.childs {
			if tx.IsShown() {
				max_grid = OsMax(max_grid, tx.Y+tx.H)
			}
		}
		min_px = layout.ui.CellWidth(float64(max_grid))
		max_px = min_px

		for _, c := range layout.UserRows {
			min_px += layout.ui.CellWidth(c.Min)
			max_px += layout.ui.CellWidth(c.Max)
			if c.Pos < max_grid {
				min_px -= layout.Cell()
				max_px -= layout.Cell()
			}
		}
	}

	if layout.scrollH.Is() {
		min_px += layout.scrollH._GetWidth(layout.ui)
		max_px += layout.scrollH._GetWidth(layout.ui)
	}
	if min_px > max_px {
		max_px = min_px
	}

	return float64(min_px) / float64(layout.Cell()), float64(max_px) / float64(layout.Cell())
}

func (layout *Layout) resizeFromPaintText(value string, multiline bool, linewrapping bool, margin [4]float64) {

	prop := InitWinFontPropsDef(layout.Cell())

	var min, max OsV2
	if multiline {

		if layout.canvas.Size.X > 0 {
			canvas_width := layout.canvas.Inner(layout.ui.CellWidth(margin[0]), layout.ui.CellWidth(margin[1]), layout.ui.CellWidth(margin[2]), layout.ui.CellWidth(margin[3])).Size.X

			max_line_px := layout.ui._UiText_getMaxLinePx(canvas_width, multiline, linewrapping)
			min.X, min.Y = layout.ui.win.GetTextSizeMax(value, max_line_px, prop)

			//when vertical scroll will be show, the max_line_px must be smaller
			/*if (min.Y * prop.lineH) > layout.scrollV.screen {
				max_line_px = canvas_width - layout.scrollV._GetWidth(layout.ui) //minus scroller width

				max_line_px = layout.ui._UiText_getMaxLinePx(max_line_px, multiline, linewrapping)
				min.X, min.Y = layout.ui.win.GetTextSizeMax(value, max_line_px, prop)
			}*/

			max = min
		} else {

			lines := strings.Split(value, "\n")
			for _, ln := range lines {
				x := layout.ui.win.GetTextSize(-1, ln, prop).X
				min.X = OsMax(min.X, x)
			}
			min.Y = len(lines)

			max = min

			//if len(lines) > 1 && linewrapping {
			if linewrapping {
				min.X = OsMin(layout.ui.CellWidth(1.5), min.X) //should be sortest single word ....
			}
		}
	} else {
		min.X = layout.ui.win.GetTextSize(-1, value, prop).X
		min.Y = 1

		max = min
	}
	min.Y = OsMax(1, min.Y)

	min_sizePx := OsV2{min.X, min.Y * prop.lineH}
	min_sizePx.X += layout.ui.CellWidth(margin[2]) + layout.ui.CellWidth(margin[3])
	min_sizePx.Y += layout.ui.CellWidth(margin[0]) + layout.ui.CellWidth(margin[1])
	min_size_x := float64(min_sizePx.X) / float64(layout.Cell())
	min_size_y := float64(min_sizePx.Y) / float64(layout.Cell())

	max_sizePx := OsV2{max.X, max.Y * prop.lineH}
	max_sizePx.X += layout.ui.CellWidth(margin[2]) + layout.ui.CellWidth(margin[3])
	max_sizePx.Y += layout.ui.CellWidth(margin[0]) + layout.ui.CellWidth(margin[1])
	max_size_x := float64(max_sizePx.X) / float64(layout.Cell())
	max_size_y := float64(max_sizePx.Y) / float64(layout.Cell())

	//column
	{
		if len(layout.UserCols) == 0 {
			layout.UserCols = make([]LayoutCR, 1)
		}
		layout.UserCols[0].Min = min_size_x
		layout.UserCols[0].Max = max_size_x
	}

	//row
	if multiline {
		if len(layout.UserRows) == 0 {
			layout.UserRows = make([]LayoutCR, 1)
		}
		layout.UserRows[0].Min = min_size_y
		layout.UserRows[0].Max = max_size_y
	}
}

func (layout *Layout) _relayout() {

	if layout.IsLevel() {
		layout.rebuildLevel()
	}

	layout.updateCoord()

	if layout.IsDialog() {
		if layout.rebuildLevel() { //for dialogs, it needs to know dialog size
			layout.updateCoord()
		}
	}

	//order List
	layout.rebuildList()

	for _, it := range layout.childs {
		if it.IsShown() {
			it._relayout() //not Inner(), because parent could changed, which may influence the childs setRowFromSub()
		}
	}
}

func (layout *Layout) rebuildList() {
	if !layout.IsTypeCards() {
		return
	}

	max_width := layout._getWidth()

	//get max item size
	it_width := 0.0
	it_height := 0.0
	for _, it := range layout.childs {
		sz := it.GetLevelSize().Size
		it_width = OsMaxFloat(it_width, float64(sz.X)/float64(layout.Cell()))
		it_height = OsMaxFloat(it_height, float64(sz.Y)/float64(layout.Cell()))
	}

	//num cols/rows
	nx := int(max_width / it_width)
	if nx == 0 {
		nx = 1
	}
	ny := len(layout.childs) / nx
	if len(layout.childs)%nx > 0 {
		ny++
	}

	total_extra_space_w := max_width - float64(nx)*it_width
	var space_between_x float64
	if layout.Cards_autoSpacing {
		space_between_x = total_extra_space_w / float64(nx+1)
	}
	var space_between_y float64
	if space_between_x > 0 {
		space_between_y = OsMinFloat(space_between_x, 1) //max 1
	}

	//set cols/rows
	layout.UserCols = nil
	layout.UserRows = nil
	for x := range nx {
		if space_between_x > 0 {
			layout.UserCols = append(layout.UserCols, LayoutCR{Pos: x*2 + 0, Min: 0, Max: space_between_x}) //space

			layout.UserCols = append(layout.UserCols, LayoutCR{Pos: x*2 + 1, Min: it_width, Max: it_width}) //item

			if x+1 == nx { //last
				layout.UserCols = append(layout.UserCols, LayoutCR{Pos: x*2 + 2, Min: 0, Max: space_between_x}) //space
			}
		} else {
			layout.UserCols = append(layout.UserCols, LayoutCR{Pos: x, Min: it_width, Max: it_width})
		}

	}
	for y := 0; y < ny; y++ {
		if space_between_y > 0 {

			layout.UserRows = append(layout.UserRows, LayoutCR{Pos: y*2 + 0, Min: it_height, Max: it_height}) //item

			if y+1 < ny { //not last
				layout.UserRows = append(layout.UserRows, LayoutCR{Pos: y*2 + 1, Min: space_between_y, Max: space_between_y}) //space
			}

		} else {
			layout.UserRows = append(layout.UserRows, LayoutCR{Pos: y, Min: it_height, Max: it_height})
		}
	}

	//set item grid poses
	i := 0
	for y := range ny {
		for x := range nx {
			if i < len(layout.childs) {
				if space_between_x > 0 {
					layout.childs[i].X = x*2 + 1
				} else {
					layout.childs[i].X = x
				}
				if space_between_y > 0 {
					layout.childs[i].Y = y * 2
				} else {
					layout.childs[i].Y = y
				}

				layout.childs[i].W = 1
				layout.childs[i].H = 1
				i++
			}
		}
	}

	//update!
	layout.updateCoord()
}

func (layout *Layout) GetCd(cd, cd_over, cd_down color.RGBA) color.RGBA {
	if layout.CanTouch() {
		active := layout.IsMouseButtonPressed()
		inside := layout.IsMouseInside() && (active || !layout.IsMouseButtonUse())
		if active {
			if inside {
				cd = cd_down
			} else {
				cd = Color_Aprox(cd_down, cd_over, 0.3)
			}
		} else {
			if inside {
				cd = cd_over
			}
		}
	}
	return cd
}

func (layout *Layout) getCanvasPx(rect Rect) OsV4 {
	cell := float64(layout.Cell())

	var ret OsV4
	ret.Start.X = layout.canvas.Start.X + int(math.Round(rect.X*cell))
	ret.Start.Y = layout.canvas.Start.Y + int(math.Round(rect.Y*cell))
	ret.Size.X = int(math.Round(rect.W * cell))
	ret.Size.Y = int(math.Round(rect.H * cell))

	return ret
}

func (layout *Layout) IsCropZero() bool {
	return layout.crop.IsZero()
}

func (layout *Layout) renderBuffer(buffer []LayoutDrawPrim) (hasBrush bool) {
	if layout.IsCropZero() {
		return false
	}

	buff := layout.ui.GetWin().buff

	for _, it := range buffer {
		coord := layout.getCanvasPx(it.Rect)

		if it.Rectangle != nil {
			st := it.Rectangle

			frontCd := layout.GetCd(st.Cd, st.Cd_over, st.Cd_down)
			if st.Radius > 0 {
				buff.AddRectRound(coord, layout.ui.CellWidth(st.Radius), frontCd, layout.ui.CellWidth(st.Border))
			} else {
				buff.AddRect(coord, frontCd, layout.ui.CellWidth(st.Border))
			}
		}
		if it.Circle != nil {
			st := it.Circle
			frontCd := layout.GetCd(st.Cd, st.Cd_over, st.Cd_down)
			buff.AddCircle(coord, frontCd, layout.ui.CellWidth(st.Border))
		}
		if it.Line != nil {
			st := it.Line
			frontCd := layout.GetCd(st.Cd, st.Cd_over, st.Cd_down)

			var start, end OsV2
			start.X = coord.Start.X + int(float64(coord.Size.X)*st.Sx)
			start.Y = coord.Start.Y + int(float64(coord.Size.Y)*st.Sy)
			end.X = coord.Start.X + int(float64(coord.Size.X)*st.Ex)
			end.Y = coord.Start.Y + int(float64(coord.Size.Y)*st.Ey)

			buff.AddLine(start, end, frontCd, layout.ui.CellWidth(st.Border))
		}

		if it.File != nil {
			st := it.File
			frontCd := layout.GetCd(st.Cd, st.Cd_over, st.Cd_down)

			var tx, ty, sx, sy float64
			buff.AddImage(st.ImagePath, coord, frontCd, OsV2{int(st.Align_h), int(st.Align_v)}, &tx, &ty, &sx, &sy, layout.GetPalette().GetGrey(0.5), layout.Cell())
		}

		if it.Text != nil {
			tx := it.Text
			frontCd := layout.GetCd(tx.Cd, tx.Cd_over, tx.Cd_down)

			prop := InitWinFontPropsDef(layout.Cell())

			prop.formating = tx.Formating

			var coordText OsV4
			if layout.fnAutoResize != nil {
				coordText = layout.canvas.Inner(layout.ui.CellWidth(tx.Margin[0]), layout.ui.CellWidth(tx.Margin[1]), layout.ui.CellWidth(tx.Margin[2]), layout.ui.CellWidth(tx.Margin[3]))
			} else {
				coordText = layout.getCanvasPx(it.Rect).Inner(layout.ui.CellWidth(tx.Margin[0]), layout.ui.CellWidth(tx.Margin[1]), layout.ui.CellWidth(tx.Margin[2]), layout.ui.CellWidth(tx.Margin[3]))
			}

			align := OsV2{int(tx.Align_h), int(tx.Align_v)}
			layout.ui._Text_draw(layout, coordText, tx.Text, tx.Ghost, prop, frontCd, align, tx.Selection, tx.Editable, tx.Multiline, tx.Linewrapping, tx.Password)

			//draw border
			if tx.Editable {
				width := 0.03
				if layout.ui.edit.Is(layout) {
					width *= 2
				}
				rounding := layout.ui.CellWidth(layout.getRounding())

				cd := layout.GetPalette().P
				if layout.ui.router.services.mic.Find(layout.UID) != nil {
					cd = layout.GetPalette().E
				}
				buff.AddRectRound(layout.canvas, rounding, cd, layout.ui.CellWidth(width))
			}
		}

		if it.Cursor != nil {
			st := it.Cursor
			cq := coord.GetIntersect(layout.crop)
			if layout.CanTouch() && cq.Inside(layout.ui.GetWin().io.Touch.Pos) {
				layout.ui.GetWin().PaintCursor(st.Name)
			}
		}

		if it.Tooltip != nil {
			st := it.Tooltip
			force := st.Force
			if force && !layout.IsTouchActive() {
				force = false
			}
			if layout.CanTouch() && (force || !layout.ui.touch.IsActive()) {
				coord := coord.GetIntersect(buff.crop)

				if force {
					layout.ui.tooltip.SetForce(coord, true, st.Description, layout.GetPalette().OnB)
				} else {
					layout.ui.tooltip.Set(coord, false, st.Description, layout.GetPalette().OnB, layout.ui)
				}
			}
		}

		if it.Brush != nil {
			hasBrush = true
		}
	}
	return
}

func (layout *Layout) renderBufferBrush(buffer []LayoutDrawPrim) {
	if layout.IsCropZero() {
		return
	}

	buff := layout.ui.GetWin().buff

	for _, it := range buffer {
		if it.Brush != nil {
			backupDepth := buff.depth

			buff.depth = 900
			buff.AddBrush(layout.canvas.Start, it.Brush.Points, it.Brush.Cd, UiSelection_Thick(layout.ui), true)

			buff.depth = backupDepth
		}
	}

}

func (layout *Layout) numBrushes() int {
	n := 0

	for _, it := range layout.buffer {
		if it.Brush != nil {
			n++
		}
	}

	//subs
	for _, it := range layout.childs {
		if it.IsShown() {
			n += it.numBrushes()
		}
	}

	return n
}

func (layout *Layout) RedrawThis() {
	if len(layout.ui.redrawLayouts) > 0 && layout.ui.redrawLayouts[len(layout.ui.redrawLayouts)-1] == layout.UID {
		return //already added
	}
	layout.ui.redrawLayouts = append(layout.ui.redrawLayouts, layout.UID)
}

func (layout *Layout) _getWidth() float64 {
	return float64(layout.canvas.Size.X) / float64(layout.Cell())
}
func (layout *Layout) _getHeight() float64 {
	return float64(layout.canvas.Size.Y) / float64(layout.Cell())
}

func (layout *Layout) GetMouseX() float64 {
	if !layout.CanTouch() {
		return -1
	}

	return float64(layout.ui.GetWin().io.Touch.Pos.X-layout.canvas.Start.X) / float64(layout.Cell())
}
func (layout *Layout) GetMouseY() float64 {
	if !layout.CanTouch() {
		return -1
	}

	return float64(layout.ui.GetWin().io.Touch.Pos.Y-layout.canvas.Start.Y) / float64(layout.Cell())
}

func (layout *Layout) GetMouseWheel() int {
	if !layout.CanTouch() {
		return 0
	}
	return layout.ui.GetWin().io.Touch.Wheel
}

func (layout *Layout) IsCtrlPressed() bool {
	if !layout.CanTouch() {
		return false
	}
	return layout.ui.GetWin().io.Keys.Ctrl
}

func (layout *Layout) IsMouseButtonDownStart() bool {
	if !layout.CanTouch() {
		return false
	}
	return layout.ui.GetWin().io.Touch.Start && layout.IsTouchActive() && !layout.IsCtrlPressed()
}
func (layout *Layout) GetMouseButtonDown() bool {
	if !layout.CanTouch() {
		return false
	}
	return layout.IsTouchActive() && !layout.IsCtrlPressed()
}
func (layout *Layout) GetMouseButtonUp() bool {
	if !layout.CanTouch() {
		return false
	}
	return layout.IsTouchEnd() && !layout.IsCtrlPressed()
}

func (layout *Layout) IsMouseInside() bool {
	if !layout.CanTouch() {
		return false
	}
	return layout.crop.Inside(layout.ui.GetWin().io.Touch.Pos)
}
func (layout *Layout) IsMouseButtonPressed() bool {
	return layout.IsMouseButtonDownStart() || layout.GetMouseButtonDown() || layout.GetMouseButtonUp()
}
func (layout *Layout) IsMouseButtonUse() bool {
	if !layout.CanTouch() {
		return false
	}
	return layout.ui.touch.IsCanvasActive() && !layout.IsCtrlPressed()
}

func (layout *Layout) UpdateTouch() {

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

func (layout *Layout) findBufferText() (Rect, *LayoutDrawText) {
	for _, tx := range layout.buffer {
		if tx.Text != nil {
			return tx.Rect, tx.Text
		}
	}
	return Rect{}, nil
}

func (layout *Layout) textComp() {

	if layout.CanTouch() {

		rect, tx := layout.findBufferText()
		if tx != nil {
			prop := InitWinFontPropsDef(layout.Cell())
			prop.formating = tx.Formating

			var coordText OsV4
			if layout.fnAutoResize != nil {
				coordText = layout.canvas.Inner(layout.ui.CellWidth(tx.Margin[0]), layout.ui.CellWidth(tx.Margin[1]), layout.ui.CellWidth(tx.Margin[2]), layout.ui.CellWidth(tx.Margin[3]))
			} else {
				coordText = layout.getCanvasPx(rect).Inner(layout.ui.CellWidth(tx.Margin[0]), layout.ui.CellWidth(tx.Margin[1]), layout.ui.CellWidth(tx.Margin[2]), layout.ui.CellWidth(tx.Margin[3]))
			}
			align := OsV2{int(tx.Align_h), int(tx.Align_v)}

			layout.ui._Text_update(layout, coordText, tx.Margin, tx.Text, prop, align, tx.Selection, tx.Editable, true, tx.Multiline, tx.Linewrapping, tx.Refresh)
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

func (layout *Layout) findParentScroll() *Layout {
	if layout == nil || !layout.CanTouch() {
		return nil
	}
	if layout.scrollV.Is() || layout.scrollH.Is() {
		return layout
	}
	return layout.parent.findParentScroll()
}

func (layout *Layout) touchComp() {
	if layout.CanTouch() {
		var in LayoutInput

		in.Rect = Rect{X: 0, Y: 0, W: layout._getWidth(), H: layout._getHeight()}
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
			if layout.fnInput != nil {
				layout.fnInput(in, layout)
			}
		}
	}
}

func (layout *Layout) Draw() {
	buff := layout.ui.GetWin().buff
	buff.AddCrop(layout.CropWithScroll())
	buff.AddRect(buff.crop, layout.GetPalette().B, 0)

	//base
	layout._drawBuffers()

	//dialogs
	for _, dia := range layout.ui.settings.Dialogs {
		layDia := layout.FindUID(dia.UID)
		if layDia != nil {
			layApp := layDia.GetApp()
			if layApp != nil {
				//alpha grey background
				backCanvas := layApp.CropWithScroll()
				buff.StartLevel(layDia.CropWithScroll(), layout.GetPalette().B, backCanvas, layout.ui.CellWidth(layout.getRounding()))
			}

			layDia._drawBuffers() //add renderToTexture optimalization ....
		}
	}

	//selection
	layout.ui.selection.Draw(buff, layout.ui)

	keys := layout.ui.win.io.Keys
	if keys.Ctrl && keys.Shift {
		n := 0

		postLayout := layout //only top
		for _, dia := range layout.ui.settings.Dialogs {
			layDia := layout.FindUID(dia.UID)
			if layDia != nil {
				postLayout = layDia
			}
		}

		postLayout.postDraw(0, &n)
	}
}

func (layout *Layout) postDraw(depth int, num_cds *int) {
	if layout.IsTypeLayout() || layout.IsTypeCards() {
		cd := Layout3_Get_prompt_color(*num_cds)
		cd.Cd.A = 150

		layout.drawGrid(cd.Cd, 0.03, depth)

		(*num_cds)++
		depth++
	}

	//subs
	for _, it := range layout.childs {
		if it.IsShown() {
			it.postDraw(depth, num_cds)
		}
	}
}

func (layout *Layout) _getRoundingInner(rectIn OsV4) float64 {
	if layout.parent == nil {
		return layout.ui.router.services.sync.GetRounding()
	}

	rad := layout.parent._getRoundingInner(rectIn) //from parent

	if layout.Back_rounding && rad > 0 {
		rectOutter := layout.canvas.Crop(layout.ui.CellWidth(layout.Back_margin))

		s := rectIn.Start.Sub(rectOutter.Start)
		e := rectOutter.End().Sub(rectIn.End())
		pad := OsClamp(0, OsMin(s.X, s.Y), OsMin(e.X, e.Y))

		if pad > 0 {
			r := layout.ui.CellWidth(rad) - pad //inner_rad = outter_rad - pad
			rad = OsMaxFloat(3/float64(layout.Cell()), float64(r)/float64(layout.Cell()))
		}
	}

	return rad
}
func (layout *Layout) getRounding() float64 {
	return layout._getRoundingInner(layout.canvas.Crop(layout.ui.CellWidth(layout.Back_margin)))
}

func (layout *Layout) _drawBuffers() {
	buff := layout.ui.GetWin().buff

	buff.AddCrop(layout.CropWithScroll())
	layout._renderScroll()

	buff.AddCrop(layout.crop)

	rad := 0
	if layout.Back_rounding {
		rad = layout.ui.CellWidth(layout.getRounding())
	}

	if layout.Back_cd.A > 0 {
		rc := layout.canvas.Crop(layout.ui.CellWidth(layout.Back_margin))
		buff.AddRectRound(rc, rad, layout.Back_cd, 0) //background
	}

	if layout.Border_cd.A > 0 {
		rc := layout.canvas.Crop(layout.ui.CellWidth(layout.Back_margin))
		buff.AddRectRound(rc, rad, layout.Border_cd, layout.ui.CellWidth(0.03)) //background
	}

	hasBrush := layout.renderBuffer(layout.buffer)

	layout.drawResizer()

	//subs
	for _, tx := range layout.childs {
		if tx.IsShown() {
			tx._drawBuffers()
		}
	}

	layout.drawDragAndDrop()

	//draw alpha rect = disable
	if layout.drawEnableFade && !layout.touch && (layout.parent == nil || layout.parent.touch) {
		buff.AddCrop(layout.crop)
		buff.AddRectRound(layout.canvas, rad, color.RGBA{255, 255, 255, 150}, 0)
	}

	if hasBrush {
		buff.AddCrop(layout.crop)
		layout.renderBufferBrush(layout.buffer)

	}
}

func (layout *Layout) drawGrid(cd color.RGBA, w float64, depth int) {
	buff := layout.ui.GetWin().buff

	canvas := layout.canvas.Size

	mx := 1.0
	my := 1.0

	var start, end OsV2
	width := layout.ui.CellWidth(w)

	cr := layout.crop.Crop(depth * width)
	buff.AddCrop(cr)

	//main border
	buff.AddRect(cr, cd, width)

	//columns
	start = layout.canvas.Start
	end = layout.canvas.End()
	end.Y = layout.canvas.Start.Y + int(float64(layout.canvas.Size.Y)*my)
	sum := 0
	for _, c := range layout.cols.outputs {
		sum += c
		p := float64(sum) / float64(canvas.X)

		start.X = layout.canvas.Start.X + int(float64(layout.canvas.Size.X)*p)
		end.X = start.X
		buff.AddLine(start, end, cd, width)
	}
	//rest
	for start.X < layout.canvas.End().X {
		start.X += layout.Cell()
		end.X = start.X
		buff.AddLine(start, end, cd, width)
	}

	//rows
	sum = 0
	start = layout.canvas.Start
	end = layout.canvas.End()
	end.X = layout.canvas.Start.X + int(float64(layout.canvas.Size.X)*mx)
	for _, r := range layout.rows.outputs {
		sum += r
		p := float64(sum) / float64(canvas.Y)

		start.Y = layout.canvas.Start.Y + int(float64(layout.canvas.Size.Y)*p)
		end.Y = start.Y
		buff.AddLine(start, end, cd, width)
	}
	//rest
	for start.Y < layout.canvas.End().Y {
		start.Y += layout.Cell()
		end.Y = start.Y
		buff.AddLine(start, end, cd, width)
	}
}

func (layout *Layout) drawDragAndDrop() {
	if !layout.CanTouch() {
		return
	}

	buff := layout.ui.GetWin().buff

	drag := layout.ui.drag

	//activate
	if layout.Drag_group != "" && layout.IsTouchActiveSubs() {
		drag.Set(layout)
	}
	isDragged := drag.IsDraged(layout)
	isDrop := drag.IsOverDrop(layout)
	if isDragged || isDrop {
		buff.AddCrop(layout.crop)

		borderWidth := layout.ui.CellWidth(0.1)
		cd := layout.GetPalette().OnB
		cd.A = 100

		//draw drag
		if isDragged {
			buff.AddRect(layout.crop.Crop(borderWidth), cd, 0)
		}

		//draw drop
		if isDrop && layout.IsOver() {

			pos := SA_Drop_INSIDE

			r := layout.ui.GetWin().io.Touch.Pos.Sub(layout.crop.Middle())

			if layout.Drop_v && layout.Drop_h {
				arx := float64(OsAbs(r.X)) / float64(layout.crop.Size.X)
				ary := float64(OsAbs(r.Y)) / float64(layout.crop.Size.Y)
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
			} else if layout.Drop_v {
				if r.Y < 0 {
					pos = SA_Drop_V_LEFT
				} else {
					pos = SA_Drop_V_RIGHT
				}
			} else if layout.Drop_h {
				if r.X < 0 {
					pos = SA_Drop_H_LEFT
				} else {
					pos = SA_Drop_H_RIGHT
				}
			}

			//update
			drag.pos = pos
			drag.dstUID = layout.UID

			//paint
			wx := float64(borderWidth) / float64(layout.canvas.Size.X)
			wy := float64(borderWidth) / float64(layout.canvas.Size.Y)
			switch pos {
			case SA_Drop_INSIDE:
				buff.AddRect(layout.crop, cd, borderWidth) //full rect
			case SA_Drop_V_LEFT:
				buff.AddRect(layout.crop.Cut(0, 0, 1, wy), cd, 0)
			case SA_Drop_V_RIGHT:
				buff.AddRect(layout.crop.Cut(0, 1-wy, 1, wy), cd, 0)
			case SA_Drop_H_LEFT:
				buff.AddRect(layout.crop.Cut(0, 0, wx, 1), cd, 0)
			case SA_Drop_H_RIGHT:
				buff.AddRect(layout.crop.Cut(1-wx, 0, wx, 1), cd, 0)
			}
		}
	}
}

func (layout *Layout) Touch() {
	startTouch := layout.CanTouch() && layout.ui.GetWin().io.Touch.Start && !layout.IsCtrlPressed()
	over := layout.CanTouch() && layout.IsTouchPosInside() && !layout.ui.touch.IsResizeActive()

	if over && layout.CanTouch() {
		if startTouch {
			if !layout.ui.touch.IsScrollOrResizeActive() { //if lower resize or scroll is activated than don't rewrite it with higher canvas
				layout.ui.touch.Set(layout.UID, 0, 0, 0)
			}
		}

		layout.ui.touch.canvasOver = layout.UID
	}

	layout.touchScroll()

	// drop file
	if layout.CanTouch() {
		drop_path := layout.ui.win.io.Touch.Drop_path
		if layout.IsMouseInside() && drop_path != "" {
			if layout.dropFile != nil {
				layout.dropFile(drop_path)
			}
		}
	}

	// drag & drop layouts
	if layout.CanTouch() {
		drag := layout.ui.drag
		if drag.IsOverDrop(layout) && drag.IsDroped(layout) {
			if layout.ui.GetWin().io.Touch.End {
				srcDom := layout.ui.mainLayout.FindUID(drag.srcUID)
				dstDom := layout

				src_i := srcDom.Drag_index
				aim_i := dstDom.Drag_index
				dst_i := UiDrag_MoveElementIndex(src_i, aim_i, drag.pos, srcDom.Drag_source != dstDom.Drag_source)

				if srcDom.Drag_source != dstDom.Drag_source || src_i != dst_i || src_i != aim_i {
					if layout.dropMove != nil {
						layout.dropMove(src_i, dst_i, aim_i, srcDom.Drag_source, dstDom.Drag_source)
					}
				}
			}
		}
	}
}

func (layout *Layout) CropWithScroll() OsV4 {
	crop := layout.crop

	if layout.scrollV.Is() {
		crop.Size.X += layout.scrollV._GetWidth(layout.ui)
	}

	if layout.scrollH.Is() {
		crop.Size.Y += layout.scrollH._GetWidth(layout.ui)
	}

	return crop
}

func (layout *Layout) _renderScroll() {
	showBackground := layout.scrollOnScreen

	if layout.scrollV.Is() {
		scrollQuad := layout.scrollV.GetScrollBackCoordV(layout.view, layout.ui)
		layout.scrollV.DrawV(scrollQuad, showBackground, layout.ui)
	}

	if layout.scrollH.Is() {
		scrollQuad := layout.scrollH.GetScrollBackCoordH(layout.view, layout.ui)
		layout.scrollH.DrawH(scrollQuad, showBackground, layout.ui)
	}
}

func (layout *Layout) isTouchScroll() (bool, bool) {
	insideScrollV := false
	insideScrollH := false
	if layout.scrollV.Is() {
		scrollQuad := layout.scrollV.GetScrollBackCoordV(layout.view, layout.ui)
		insideScrollV = scrollQuad.Inside(layout.ui.GetWin().io.Touch.Pos)
	}

	if layout.scrollH.Is() {
		scrollQuad := layout.scrollH.GetScrollBackCoordH(layout.view, layout.ui)
		insideScrollH = scrollQuad.Inside(layout.ui.GetWin().io.Touch.Pos)
	}
	return insideScrollV, insideScrollH
}

func (layout *Layout) IsTouchPosInside() bool {
	return layout.crop.Inside(layout.ui.GetWin().io.Touch.Pos)
}

func (layout *Layout) IsTouchPosInsideOrScroll() bool {
	return layout.crop.Inside(layout.ui.GetWin().io.Touch.Pos) || layout.ui.touch.IsScrollOrResizeActive()
}

func (layout *Layout) IsOver() bool {
	return layout.CanTouch() && layout.IsTouchPosInside() && !layout.ui.touch.IsResizeActive()
}

func (layout *Layout) IsOverScroll() bool {
	insideScrollV, insideScrollH := layout.isTouchScroll()
	return layout.CanTouch() && (insideScrollV || insideScrollH)
}

func (layout *Layout) IsTouchActive() bool {
	return layout.ui.touch.IsFnMove(layout.UID, 0, 0, 0)
}

func (layout *Layout) IsTouchActiveSubs() bool {
	if layout.IsTouchActive() {
		return true
	}
	for _, tx := range layout.childs {
		if tx.IsTouchActiveSubs() {
			return true
		}
	}
	return false
}

func (layout *Layout) IsTouchInside() bool {
	inside := layout.IsOver()

	if !layout.IsTouchActive() && layout.CanTouch() && layout.ui.touch.IsActive() { // when click and move, other Buttons, etc. are disabled
		inside = false
	}

	return inside
}

func (layout *Layout) IsTouchEnd() bool {
	return layout.CanTouch() && layout.ui.GetWin().io.Touch.End && layout.IsTouchActive() //doesn't have to be inside!
}

func (layout *Layout) IsClicked(enable bool) (int, int, bool, bool, bool) {
	var click, rclick int
	var inside, active, end bool
	if enable {
		inside = layout.IsTouchInside()
		active = layout.IsTouchActive()
		end = layout.IsTouchEnd()

		touch := &layout.ui.GetWin().io.Touch

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

func (layout *Layout) CanTouch() bool {
	return layout.touch && layout.touchDia
}

func (layout *Layout) touchScroll() {
	hasScrollV := layout.scrollV.Is()
	hasScrollH := layout.scrollH.Is()

	enableInput := layout.CanTouch()

	//redraw := false
	if hasScrollV {
		if enableInput {
			if layout.scrollV.TouchV(layout) {
				wheel := layout.scrollV.wheel
				if layout.scrollV.IsDown() {
					wheel = UiRootSettings_GetMaxScroll()
				}
				layout.GetSettings().SetScrollV(layout.UID, wheel)
				layout.ui.SetRelayoutSoft()
			}
		}
	}

	if hasScrollH {
		if enableInput {
			if layout.scrollH.TouchH(hasScrollV, layout) {
				wheel := layout.scrollH.wheel
				if layout.scrollH.IsDown() {
					wheel = UiRootSettings_GetMaxScroll()
				}
				layout.GetSettings().SetScrollH(layout.UID, wheel)
				layout.ui.SetRelayoutSoft()
			}
		}
	}
}

func (layout *Layout) convert(grid OsV4) OsV4 {
	cell := layout.Cell()

	c := layout.cols.Convert(cell, grid.Start.X, grid.Start.X+grid.Size.X)
	r := layout.rows.Convert(cell, grid.Start.Y, grid.Start.Y+grid.Size.Y)

	return OsV4{OsV2{c.X, r.X}, OsV2{c.Y, r.Y}}
}

func (layout *Layout) ConvertMax(cell int, in OsV4) OsV4 {
	c := layout.cols.ConvertMax(cell, in.Start.X, in.Start.X+in.Size.X)
	r := layout.rows.ConvertMax(cell, in.Start.Y, in.Start.Y+in.Size.Y)

	return OsV4{OsV2{c.X, r.X}, OsV2{c.Y, r.Y}}
}

func (layout *Layout) updateColsRows() {

	//reset
	layout.cols.Clear()
	layout.rows.Clear()

	//project
	for _, it := range layout.UserCols {
		if it.Resize_value > 0 {
			it.Resize_value = layout.GetSettings().GetCol(layout.UID, it.Pos, it.Resize_value)
		}

		layout.Col(it.Pos, it.Min)
		layout.ColMax(it.Pos, it.Max)
		layout.ColResize(it.Pos, it.Resize_value)
	}
	for _, it := range layout.UserRows {
		if it.Resize_value > 0 {
			it.Resize_value = layout.GetSettings().GetRow(layout.UID, it.Pos, it.Resize_value)
		}
		layout.Row(it.Pos, it.Min)
		layout.RowMax(it.Pos, it.Max)
		layout.RowResize(it.Pos, it.Resize_value)
	}
}

func (layout *Layout) _updateCoordInner() {
	//maybe use these later for layout margin? ....
	rx := 0.0
	ry := 0.0
	rw := 1.0
	rh := 1.0

	isLevel := layout.IsLevel()
	if !isLevel {
		layout.view = layout.parent.convert(InitOsV4(layout.X, layout.Y, layout.W, layout.H))
		layout.view.Start = layout.parent.view.Start.Add(layout.view.Start)

		layout.view.Start.X += int(float64(layout.view.Size.X) * rx)
		layout.view.Start.Y += int(float64(layout.view.Size.Y) * ry)
		layout.view.Size.X = int(float64(layout.view.Size.X) * rw)
		layout.view.Size.Y = int(float64(layout.view.Size.Y) * rh)

		// move start by scroll
		layout.view.Start.X -= layout.parent.scrollH.GetWheel() //parent scroll
		layout.view.Start.Y -= layout.parent.scrollV.GetWheel()

	}

	// crop
	if !isLevel {
		layout.crop = layout.view.GetIntersect(layout.parent.crop)
	}

	//slow ....
	{
		makeSmallerX := layout.scrollV.Show
		makeSmallerY := layout.scrollH.Show
		gridMax := layout.GetGridMax(OsV2{})
		screen := layout.view.Size
		for layout.updateGridAndScroll(&screen, gridMax, &makeSmallerX, &makeSmallerY) {
		}
		layout.view.Size = screen
	}

	// crop
	if layout.scrollV.Is() {
		layout.crop.Size.X = OsMax(0, layout.crop.Size.X-layout.scrollV._GetWidth(layout.ui))
	}
	if layout.scrollH.Is() {
		layout.crop.Size.Y = OsMax(0, layout.crop.Size.Y-layout.scrollH._GetWidth(layout.ui))
	}

	layout.canvas.Start = layout.view.Start
	layout.canvas.Start.X -= layout.scrollH.GetWheel() //this scroll
	layout.canvas.Start.Y -= layout.scrollV.GetWheel()

	layout.canvas.Size.X = layout.cols.OutputAll()
	layout.canvas.Size.Y = layout.rows.OutputAll()
	layout.canvas = layout.canvas.Extend(layout.view)
}

func (layout *Layout) updateCoordSoft() {
	if !layout.IsLevel() {
		layout._updateCoordInner()
	}
	for _, it := range layout.childs {
		it.updateCoordSoft()
	}
}

func (layout *Layout) updateCoord() {
	layout.updateColsRows()
	layout._updateCoordInner()
}

func (layout *Layout) GetGridMax(minSize OsV2) OsV2 {
	mx := minSize
	for _, tx := range layout.childs {
		if tx.IsShown() {
			mx = mx.Max(OsV2{X: tx.X + tx.W, Y: tx.Y + tx.H})
		}
	}

	for _, col := range layout.UserCols {
		mx.X = OsMax(col.Pos+1, mx.X)
	}
	for _, row := range layout.UserRows {
		mx.Y = OsMax(row.Pos+1, mx.Y)
	}

	return mx
}

func (layout *Layout) updateArray(window OsV2, endGrid OsV2) {
	if endGrid.X > layout.cols.NumInputs() {
		layout.cols.Resize(int(endGrid.X))
	}
	if endGrid.Y > layout.rows.NumInputs() {
		layout.rows.Resize(int(endGrid.Y))
	}

	layout.cols.SetFills(layout.childs, true)
	layout.rows.SetFills(layout.childs, false)

	cell := layout.Cell()
	layout.cols.Update(cell, window.X)
	layout.rows.Update(cell, window.Y)
}

func (layout *Layout) updateGridAndScroll(screen *OsV2, gridMax OsV2, makeSmallerX *bool, makeSmallerY *bool) bool {
	// update cols/rows
	layout.updateArray(*screen, gridMax)

	// get max
	data := layout.convert(OsV4{OsV2{}, gridMax}).Size

	// make canvas smaller
	hasScrollV := OsTrnBool(*makeSmallerX, data.Y > screen.Y, false)
	hasScrollH := OsTrnBool(*makeSmallerY, data.X > screen.X, false)
	if hasScrollV {
		screen.X -= layout.scrollV._GetWidth(layout.ui)
		*makeSmallerX = false
	}
	if hasScrollH {
		screen.Y -= layout.scrollH._GetWidth(layout.ui)
		*makeSmallerY = false
	}

	// save to scroll
	layout.scrollV.data = data.Y
	layout.scrollV.screen = screen.Y

	layout.scrollH.data = data.X
	layout.scrollH.screen = screen.X

	return hasScrollV || hasScrollH
}

func (layout *Layout) GetLevelSize() OsV4 {

	layout.updateColsRows() //project .userCols -> .cols

	q := OsV4{OsV2{}, layout.GetGridMax(OsV2{1, 1})}

	q.Size = layout.ConvertMax(layout.Cell(), q).Size

	q.Start = layout.ui.winRect.Start
	q = q.GetIntersect(layout.ui.winRect)
	return q
}

func (layout *Layout) VScrollToTheTop() {
	layout.scrollV.SetWheel(0)
	layout.GetSettings().SetScrollV(layout.UID, 0)
	layout.ui.SetRelayoutSoft()

}
func (layout *Layout) HScrollToTheLeft() {
	layout.scrollH.SetWheel(0)
	layout.GetSettings().SetScrollH(layout.UID, 0)
	layout.ui.SetRelayoutSoft()
}
func (layout *Layout) VScrollToTheBottom() {
	layout.GetSettings().SetScrollV(layout.UID, UiRootSettings_GetMaxScroll())
	layout.ui.SetRelayoutSoft()
}
func (layout *Layout) VScrollToTheBottomIf() {
	//only when scroll is at the bottom
	if layout.GetSettings().GetScrollV(layout.UID) == UiRootSettings_GetMaxScroll() {
		layout.GetSettings().SetScrollV(layout.UID, UiRootSettings_GetMaxScroll())
		layout.ui.SetRelayoutSoft()
	}
}

func (layout *Layout) HScrollToTheRight() {
	layout.GetSettings().SetScrollH(layout.UID, UiRootSettings_GetMaxScroll())
	layout.ui.SetRelayoutSoft()
}

func Layout_buildLLMTip(tp string, label string, value_quotes bool, tip string) string {

	str := ""
	if tp != "" {
		str += "Type: " + tp
	}

	if label != "" {
		if value_quotes {
			label = "\"" + label + "\""
		}

		if str != "" {
			str += ", "
		}
		str += "Label: " + label
	}

	if tip != "" {
		if str != "" {
			str += ", "
		}
		str += "Tip: " + tip
	}

	return str
}
func Layout_addTip(str string, tooltip string) string {
	if str != "" {
		str += " part of "
	}
	str += "(" + tooltip + ")"
	return str
}

func (layout *Layout) GetLLMTip() string {
	var final string

	for layout != nil {
		var tip string
		if layout.fnGetLLMTip != nil {
			tip = layout.fnGetLLMTip(layout)
		}

		if tip != "" {
			final = Layout_addTip(final, tip)
		}

		if layout.parent != nil {
			for _, gr := range layout.parent.TooltipGroups {
				if gr.InInside(layout) {
					if gr.Tooltip != "" {
						final = Layout_addTip(final, Layout_buildLLMTip("", "", false, gr.Tooltip))
					}
				}
			}
		}

		layout = layout.parent
	}

	if final != "" && !strings.HasSuffix(final, ".") {
		final += "."
	}

	return final
}

func (layout *Layout) CallLayoutUpdates() {

	if layout.fnUpdate != nil {
		layout.fnUpdate()

		layout._build()
		layout._relayout()
		layout._draw()
		layout.ui.SetRedrawBuffer()
	}

	for _, it := range layout.childs {
		it.CallLayoutUpdates()
	}

}
