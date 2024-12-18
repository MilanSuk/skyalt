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
	"go/ast"
	"go/parser"
	"go/token"
	"image/color"
	"math"
	"os"
	"strings"
)

type LayoutPick struct {
	File                           string
	Line                           int
	Grid_x, Grid_y, Grid_w, Grid_h int
	Tip                            string

	Cd       color.RGBA
	Time_sec float64
}

func InitLayoutPick(file string, line int, grid OsV4, tip string) LayoutPick {
	return LayoutPick{File: file,
		Line:   line,
		Grid_x: grid.Start.X, Grid_y: grid.Start.Y, Grid_w: grid.Size.X, Grid_h: grid.Size.Y,
		Tip:      tip,
		Time_sec: OsTime(),
	}
}

type Layout3 struct {
	ui     *Ui
	parent *Layout3

	props Layout

	touch bool

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
	dom.scrollV.wheel = dom.GetSettings().GetScrollV(dom.props.Hash)
	dom.scrollH.wheel = dom.GetSettings().GetScrollH(dom.props.Hash)

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

func (dom *Layout3) setTouchEnable(touch bool) {
	dom.touch = touch && dom.props.Enable

	for _, it := range dom.childs {
		it.setTouchEnable(dom.touch)
	}
	if dom.dialog != nil {
		dom.dialog.setTouchEnable(true)
	}
}

func (dom *Layout3) setTouchDisable(subDialogs bool) {
	dom.touch = false
	for _, it := range dom.childs {

		if it.props.App {
			it.setTouchDisable(true) //disable dialogs inside
		} else {
			it.setTouchDisable(subDialogs)
		}
	}

	if dom.dialog != nil && subDialogs {
		dom.dialog.setTouchDisable(true)
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

	dom.setTouchEnable(true)

	//dialogs
	for _, dia := range dom.ui.settings.Dialogs {
		layDia := dom.FindHash(dia.Hash)
		if layDia != nil {
			layApp := layDia.parent.GetApp()
			if layApp != nil {
				//dia.appHash = layApp.props.Hash
				layApp.setTouchDisable(false)
			}
		}
	}
}

func (dom *Layout3) ResetPicks() {
	dom.props.Prompt_comp_cd = color.RGBA{}
	dom.props.Prompt_grids = nil

	if dom.dialog != nil {
		dom.dialog.ResetPicks()
	}
	for _, it := range dom.childs {
		it.ResetPicks()
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

func (dom *Layout3) FindAppLine(file string, line int) *Layout3 {
	if /*dom.caller_app &&*/ dom.props.Caller_file == file && dom.props.Caller_line == line {
		return dom
	}

	if dom.dialog != nil {
		d := dom.dialog.FindAppLine(file, line)
		if d != nil {
			return d
		}
	}

	for _, it := range dom.childs {
		d := it.FindAppLine(file, line)
		if d != nil {
			return d
		}
	}
	return nil
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

func (dom *Layout3) TouchDialogs(editHash, touchHash uint64, check bool) {

	if dom.touch && check {
		//dom.updateShortcut()
		dom.updateTouch()

		var act *Layout3
		var actE *Layout3
		if editHash != 0 {
			actE = dom.FindHash(editHash)
		}

		if touchHash != 0 {
			act = dom.FindHash(touchHash)
		} else {
			act = dom.findTouch()
		}

		if actE != nil {
			actE.touchComp()
		}
		if act != nil && act != actE {
			act.touchComp()
		}
	}

	if dom.dialog != nil {
		dom.dialog.TouchDialogs(editHash, touchHash, true)
		return
	}

	for _, it := range dom.childs {
		it.TouchDialogs(editHash, touchHash, false)
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

func (dom *Layout3) Relayout(setFromSubs bool) {

	if dom.IsLevel() {
		dom.rebuildLevel()
	}

	dom.updateCoord(0, 0, 1, 1)

	/*if dom.fnResize != nil {
		dom.fnResize(dom)
		dom.updateCoord(0, 0, 1, 1)
	}*/

	if dom.IsDialog() {
		dom.rebuildLevel() //for dialogs, it needs to know dialog size
		dom.updateCoord(0, 0, 1, 1)
	}

	//order List
	//dom.rebuildList()

	for _, it := range dom.childs {
		if it.IsShown() {
			it.Relayout(setFromSubs)
		}
	}

	if setFromSubs {
		changed := false

		for i, c := range dom.props.UserCols {
			if c.SetFromChild {
				v := 1.0
				for _, it := range dom.childs {
					if it.props.X == c.Pos && it.props.W == 1 {
						v = OsMaxFloat(v, it._getWidth())
					}
				}
				dom.props.UserCols[i].Min = v
				dom.props.UserCols[i].Max = v

				changed = true
			}
		}

		for i, r := range dom.props.UserRows {
			if r.SetFromChild {
				v := 1.0
				for _, it := range dom.childs {
					if it.props.Y == r.Pos && it.props.H == 1 {
						v = OsMaxFloat(v, it._getHeight())
					}
				}
				dom.props.UserRows[i].Min = v
				dom.props.UserRows[i].Max = v

				changed = true
			}
		}

		if changed {
			dom.Relayout(false)
		}

	}

	if setFromSubs {
		//if dom.props.Name == "_layout" && dom.canvas.Size.Y == 37 {
		//	fmt.Println("bug")
		//}

		//fmt.Println("-Relayout", dom.props.Name, dom.canvas.Size.Y)
		//dom.needRedraw = dom.needRedraw || (oldRect != dom._getRect())
	}
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
		frontCd := dom.GetCd(it.Cd, it.Cd_over, it.Cd_down)

		switch it.Type {
		case 1:
			buff.AddRect(coord, frontCd, dom.ui.CellWidth(it.Border))
		case 2:
			buff.AddCircle(coord, frontCd, dom.ui.CellWidth(it.Border))
		case 3:
			var tx, ty, sx, sy float64
			path := InitWinMedia_url(it.Text)
			buff.AddImage(path, coord, frontCd, OsV2{int(it.Align_h), int(it.Align_v)}, &tx, &ty, &sx, &sy, dom.GetPalette().E, dom.Cell())

		case 4:
			var start, end OsV2
			start.X = coord.Start.X + int(float64(coord.Size.X)*it.Sx)
			start.Y = coord.Start.Y + int(float64(coord.Size.Y)*it.Sy)
			end.X = coord.Start.X + int(float64(coord.Size.X)*it.Ex)
			end.Y = coord.Start.Y + int(float64(coord.Size.Y)*it.Ey)

			buff.AddLine(start, end, frontCd, dom.ui.CellWidth(it.Border))

		case 5:
			prop := InitWinFontPropsDef(dom.Cell())

			prop.formating = it.Text_formating
			coord := dom.getCanvasPx(it.Rect)

			align := OsV2{int(it.Align_h), int(it.Align_v)}
			dom.ui._Text_draw(dom, coord, it.Text, it.Text2, prop, frontCd, align, it.Text_selection, it.Text_editable, it.Text_multiline, it.Text_linewrapping)

		case 6:
			cq := coord.GetIntersect(dom.crop)
			if dom.CanTouch() && cq.Inside(dom.ui.GetWin().io.Touch.Pos) {
				dom.ui.GetWin().PaintCursor(it.Text)
			}

		case 7:
			force := it.Boolean
			if force && !dom.IsTouchActive() {
				force = false
			}

			if dom.CanTouch() && (force || !dom.GetUis().touch.IsAnyActive()) {
				coord := coord.GetIntersect(buff.crop)

				if force {
					dom.ui.tooltip.SetForce(coord, true, it.Text, dom.ui.GetPalette().OnB)
				} else {
					dom.ui.tooltip.Set(coord, false, it.Text, dom.ui.GetPalette().OnB)
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

func (dom *Layout3) GetDateFormat() string {
	return dom.GetEnv().DateFormat
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
	return dom.IsTouchAnyActive() && !dom.IsCtrlPressed()
}

func (layout *Layout3) updateTouch() {

	layout.Touch()

	layout.updateResizer()

	//subs
	for _, it := range layout.childs {
		if it.IsShown() {
			it.updateTouch()
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

func (layout *Layout3) findBufferText() int {
	for i, it := range layout.buffer {
		if it.Type == 5 {
			return i
		}
	}
	return -1
}

func (layout *Layout3) textComp() {

	if !layout.CanTouch() {
		return
	}

	ti := layout.findBufferText()
	if ti >= 0 {
		it := layout.buffer[ti]

		prop := InitWinFontPropsDef(layout.Cell())
		prop.formating = it.Text_formating

		coord := layout.getCanvasPx(it.Rect)
		align := OsV2{int(it.Align_h), int(it.Align_v)}

		layout.ui._Text_update(layout, coord, it.Text, prop, align, it.Text_selection, it.Text_editable, true, it.Text_multiline, true, it.Text_linewrapping)
	}

	for _, it := range layout.childs {
		if it.IsShown() {
			it.textComp()
		}
	}
}

func (layout *Layout3) touchComp() {
	//dom.drop_path = ""
	/*if layout.DropFile != nil && layout.CanTouch() {
		drop_path := layout.GetUis().win.io.Touch.Drop_path
		if layout.IsMouseInside() && drop_path != "" {
			layout.DropFile(drop_path)
			layout.ui.UpdateCode()
		}
	}*/

	if layout.CanTouch() {

		var in LayoutInput
		{
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
		}

		if in.IsStart || in.IsActive || (in.Wheel != 0) || in.NumClicks > 0 || in.IsUse || in.IsUp {
			layout.ui.parent.CallInput(&layout.props, &in) //err ...
		}
	}

	if layout.CanTouch() {
		drop_path := layout.GetUis().win.io.Touch.Drop_path
		if layout.IsMouseInside() && drop_path != "" {
			in := LayoutInput{Drop_path: drop_path}
			layout.ui.parent.CallInput(&layout.props, &in) //err ...

		}
	}

	//select component
	if layout.IsTouchCtrlComp_start() { // && layout.caller_app {
		layout.ui.SelectComp_active = true
		layout.ui.Selected_hash = layout.props.Hash //GetParentLevelDepth(dom.ui.SelectionDepth).hash
	}

	if layout.ui.SelectComp_active && layout.ui.Selected_hash == layout.props.Hash && layout.IsTouchCtrlComp_end() {

		layout.ui.parent.CallPick(InitLayoutPick(layout.props.Caller_file, layout.props.Caller_line, OsV4{}, layout.props.LLMTip))
		layout.ui.SetRefresh()

		layout.ui.Selected_hash = 0
		layout.ui.SelectComp_active = false
	}

	{
		if layout.IsTouchCtrlGrid_start() /*&& layout.IsDiv()*/ { //&& layout.caller_app {
			layout.ui.SelectGrid_active = true
			layout.ui.Selected_hash = layout.props.Hash //GetParentLevelDepth(dom.ui.SelectionDepth).hash
		}

		if layout.ui.SelectGrid_active && layout.ui.Selected_hash == layout.props.Hash {

			touch_rel := layout.ui.GetWin().io.Touch.Pos.Sub(layout.canvas.Start)
			c := layout.cols.GetCloseCell(touch_rel.X)
			r := layout.rows.GetCloseCell(touch_rel.Y)

			if layout.IsTouchCtrlGrid_start() {
				layout.props.Prompt_grids = append(layout.props.Prompt_grids, LayoutPickGrid{Cd: color.RGBA{0, 0, 0, 255}})
				layout.ui.Selected_start = OsV2{c, r}
			}

			coord := InitOsV4ab(layout.ui.Selected_start, OsV2{c, r})
			coord.Size.X++
			coord.Size.Y++

			if len(layout.props.Prompt_grids) > 0 {
				pg := &layout.props.Prompt_grids[len(layout.props.Prompt_grids)-1]
				//if !InitOsV4(pg.Grid_x, pg.Grid_y, pg.Grid_w, pg.Grid_h).Cmp(coord) {
				//layout.GetUis().RedrawComp()
				//layout.drawComp(0)
				//}
				//update
				pg.Grid_x = coord.Start.X
				pg.Grid_y = coord.Start.Y
				pg.Grid_w = coord.Size.X
				pg.Grid_h = coord.Size.Y
			}

			if layout.IsTouchCtrlGrid_end() {

				//right not, Caller position of Add<widget>(). For grid we need position of Build()
				path := "sdk/" + layout.props.Name + ".go"
				fileCode, err := os.ReadFile(path)
				if err == nil {

					build_pos, _, _, _, err := UiClients_findBuildFunction(path, string(fileCode), layout.props.Name)
					if err == nil && build_pos >= 0 {
						//rewrite Caller
						layout.props.Caller_file = layout.props.Name + ".go"
						layout.props.Caller_line = strings.Count(string(fileCode[:build_pos]), "\n") + 1

						layout.ui.parent.CallPick(InitLayoutPick(layout.props.Caller_file, layout.props.Caller_line, coord, layout.props.LLMTip))
						layout.ui.SetRefresh()

						layout.ui.Selected_hash = 0
						layout.ui.SelectGrid_active = false
					}
				}
			}
		}
	}
}

func UiClients_findBuildFunction(ghostPath string, code string, stName string) (int, int, int, int, error) {
	node, err := parser.ParseFile(token.NewFileSet(), ghostPath, code, parser.ParseComments)
	if err != nil {
		return -1, -1, -1, -1, err
	}

	build_pos := -1
	touch_pos := -1
	draw_pos := -1
	shortcut_pos := -1
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:

			tp := ""
			if x.Recv != nil && len(x.Recv.List) > 0 {
				tp = string(code[x.Recv.List[0].Type.Pos()-1 : x.Recv.List[0].Type.End()-1])
			}

			//function
			if tp == "*"+stName {
				if x.Name.Name == "Build" {
					build_pos = int(x.Pos())
				}
				if x.Name.Name == "Touch" {
					touch_pos = int(x.Pos())
				}
				if x.Name.Name == "Draw" {
					draw_pos = int(x.Pos())
				}
				if x.Name.Name == "Shortcut" {
					shortcut_pos = int(x.Pos())
				}
			}
		}
		return true
	})

	return build_pos, touch_pos, draw_pos, shortcut_pos, nil
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
				buff.StartLevel(layDia.crop, dom.ui.GetPalette().B, backCanvas)
			}

			layDia.drawBuffers() //add renderToTexture optimalization ...
		}
	}
}

func (dom *Layout3) drawBuffers() {
	buff := dom.ui.GetWin().buff

	buff.AddCrop(dom.CropWithScroll())
	dom._renderScroll()

	buff.AddCrop(dom.crop)

	if dom.props.Back_cd.A > 0 {
		r := dom.crop.AddSpace(dom.ui.CellWidth(dom.props.Back_margin))
		buff.AddRect(r, dom.props.Back_cd, 0) //background
	}

	if dom.props.Border_cd.A > 0 {
		r := dom.crop.AddSpace(dom.ui.CellWidth(dom.props.Back_margin))
		buff.AddRect(r, dom.props.Border_cd, dom.ui.CellWidth(0.03)) //background
	}

	dom.renderBuffer(dom.buffer)

	dom.drawResizer()

	//subs
	for _, it := range dom.childs {
		if it.IsShown() {
			it.drawBuffers()
		}
	}

	dom.draw_grid()
	dom.draw_drag_and_drop()
}

func (dom *Layout3) draw_grid() {
	buff := dom.ui.GetWin().buff

	//show grid
	if dom.ui.ShowGrid || dom.GetUis().win.io.Keys.Ctrl /*&& !dom.ui.IsRootApp() && dom.IsDiv()*/ {
		canvas := dom.canvas.Size
		cd := dom.GetPalette().OnB
		cd.A = 30

		mx := float64(dom.cols.GetSumOutput(-1)) / float64(canvas.X)
		my := float64(dom.rows.GetSumOutput(-1)) / float64(canvas.Y)

		var start, end OsV2

		buff.AddCrop(dom.crop)

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
			buff.AddLine(start, end, cd, dom.ui.CellWidth(0.03))
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
			buff.AddLine(start, end, cd, dom.ui.CellWidth(0.03))
		}
	}

	//if dom.ui.EditMode
	{
		buff.AddCrop(dom.crop)

		//prompt - component
		if dom.props.Prompt_comp_cd.A > 0 {
			//component
			rc := dom._getRect()
			rc = rc.Cut(0.1)
			cd := dom.props.Prompt_comp_cd

			buff.AddRect(dom.getCanvasPx(rc), cd, dom.ui.CellWidth(0.06))
			buff.AddText("<h2>"+dom.props.Prompt_label, InitWinFontPropsDef(dom.Cell()), cd, dom.getCanvasPx(rc.Cut(0.1)), OsV2{0, 0}, 0, 1)
		}

		//prompt - grids
		for _, it := range dom.props.Prompt_grids {
			//grid
			cell := float64(dom.Cell())

			start := OsV2{it.Grid_x, it.Grid_y}
			end := OsV2{it.Grid_x + it.Grid_w, it.Grid_y + it.Grid_h}
			sx := dom.cols.GetSumOutput(start.X)
			ex := dom.cols.GetSumOutput(end.X)
			sy := dom.rows.GetSumOutput(start.Y)
			ey := dom.rows.GetSumOutput(end.Y)

			rc := dom._getRect()
			rc.X += float64(sx) / cell
			rc.Y += float64(sy) / cell
			rc.W = float64(ex-sx) / cell
			rc.H = float64(ey-sy) / cell

			rc = rc.Cut(0.1)
			cd := it.Cd
			cd.A = 30

			buff.AddRect(dom.getCanvasPx(rc), cd, 0)
			buff.AddText("<h2>"+it.Label, InitWinFontPropsDef(dom.Cell()), it.Cd, dom.getCanvasPx(rc.Cut(0.1)), OsV2{0, 0}, 0, 1)
		}
	}
}

func (dom *Layout3) draw_drag_and_drop() {
	buff := dom.ui.GetWin().buff

	//drag & drop
	//activate
	if dom.props.Drag_group != "" && dom.IsTouchActiveSubs() {
		dom.GetUis().drag.Set(dom)
	}
	isDragged := dom.GetUis().drag.IsDraged(dom)
	isDrop := dom.GetUis().drag.IsOverDrop(dom)
	if isDragged || isDrop {
		buff.AddCrop(dom.crop)

		borderWidth := dom.ui.CellWidth(0.1)
		cd := dom.GetPalette().OnB
		cd.A = 100

		//draw drag
		if isDragged {
			buff.AddRect(dom.crop.AddSpace(borderWidth), cd, 0)
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

			/*q := dom.crop
			if dom.Drop_v {
				q = q.AddSpaceY(dom.crop.Size.Y / 3)
			}
			if dom.Drop_h {
				q = q.AddSpaceX(dom.crop.Size.X / 3)
			}*/

			/*if !dom.Drop_v && !dom.Drop_h {
				pos = 0
			} else if q.Inside(touchPos) {
				pos = 0
			}*/

			//...

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

			//move child item
			if dom.CanTouch() && dom.ui.GetWin().io.Touch.End {
				//srcDom
				/*srcDom := dom.GetUis().drag.dom
				src_i := srcDom.props.Drag_index
				dst_i := dom.props.Drag_index

				if dom.Moved != nil {
					dst_i = OsMoveElementIndex(src_i, dst_i, pos)
					dom.Moved(src_i, dst_i)
				}*/

				dom.ui.SetRefresh()
				//maybe in update()? ...
				//...
			}
		}
	}
}

func (dom *Layout3) Touch() {
	startTouch := dom.CanTouch() && dom.ui.GetWin().io.Touch.Start && !dom.IsCtrlPressed()
	over := dom.CanTouch() && dom.IsTouchPosInside() && !dom.GetUis().touch.IsResizeActive()

	if over && startTouch && dom.CanTouch() {
		if !dom.GetUis().touch.IsScrollOrResizeActive() { //if lower resize or scroll is activated than don't rewrite it with higher canvas
			dom.GetUis().touch.Set(dom.props.Hash, 0, 0, 0)
		}
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
	for _, it := range dom.childs {
		if it.IsTouchActiveSubs() {
			return true
		}
	}
	return false
}

func (dom *Layout3) IsTouchAnyActive() bool {
	return dom.GetUis().touch.IsAnyActive()
}

func (dom *Layout3) IsTouchInside() bool {
	inside := dom.IsOver()

	if !dom.IsTouchActive() && dom.CanTouch() && dom.GetUis().touch.IsAnyActive() { // when click and move, other Buttons, etc. are disabled
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
	return dom.touch
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

	//slow ......
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
	for _, it := range dom.childs {
		if it.IsShown() {
			mx = mx.Max(OsV2{X: it.props.X + it.props.W, Y: it.props.Y + it.props.H})
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
