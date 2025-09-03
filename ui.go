/*
Copyright 2025 Milan Suk

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
	"os"
	"slices"
	"time"
)

type UiStats struct {
	Build         float64
	Changed       float64
	Relayout_soft float64 //sec
	Relayout_hard float64
}

func (s *UiStats) Get() (float64, float64, float64, float64) {
	return s.Build, s.Changed, s.Relayout_soft, s.Relayout_hard
}

type Ui struct {
	win    *Win
	router *AppsRouter

	winRect OsV4

	settings *UiSettings

	mainLayout *Layout

	refresh bool

	relayout_hard bool
	relayout_soft bool
	redrawBuffer  bool

	maintenance_tick int64

	redrawLayouts []uint64 //hashes

	datePage int64

	edit_history *UiTextHistoryArray
	tooltip      *UiTooltip
	touch        *UiInput
	edit         *UiEdit
	drag         *UiDrag
	selection    *UiSelection

	temp_ui   *UI
	temp_cmds []ToolCmd

	last_layout_updates_ticks int64

	runPrompt string

	stats UiStats
}

func NewUi(win *Win, router *AppsRouter) (*Ui, error) {
	ui := &Ui{win: win, router: router}

	ui.datePage = time.Now().Unix()

	ui.settings = &UiSettings{}

	ui.edit_history = &UiTextHistoryArray{}
	ui.tooltip = &UiTooltip{}
	ui.touch = &UiInput{}
	ui.edit = &UiEdit{}
	ui.drag = &UiDrag{}
	ui.selection = &UiSelection{}

	ui.mainLayout = NewUiLayoutDOM_root(ui)
	ui.settings.Layouts.Init()

	ui.open()

	ui.SetRefresh()

	win.getUiStats = ui.stats.Get

	return ui, nil
}
func (ui *Ui) Destroy() {
	ui.save()

	ui.mainLayout.Destroy()
}

func (ui *Ui) GetSettingsPath() string {
	return "layouts.json"
}

func (ui *Ui) open() bool {
	//settings
	{
		jsRead, err := os.ReadFile(ui.GetSettingsPath())
		if err != nil {
			return false
		}

		if len(jsRead) > 0 {
			err := LogsJsonUnmarshal(jsRead, ui.settings)
			if err != nil {
				return false
			}
		}
	}

	return true
}

func (ui *Ui) save() error {

	//save editbox
	if ui.edit.IsActive() {
		ui.edit.send(false, ui)
	}

	//settings
	_, err := Tools_WriteJSONFile(ui.GetSettingsPath(), ui.settings)
	if err != nil {
		return LogsErrorf("Save(%s) failed: %v\n", ui.GetSettingsPath(), err)
	}

	return nil
}

func (ui *Ui) GetWin() *Win {
	return ui.win
}

func (ui *Ui) UpdateIO(winRect OsV4) {

	keys := ui.win.io.Keys
	if keys.ZoomAdd {
		ui.router.services.sync.Upload_deviceDPI(OsClamp(ui.router.services.sync.Device.Dpi+3, 30, 5000))
		keys.ZoomAdd = false
	}
	if keys.ZoomSub {
		ui.router.services.sync.Upload_deviceDPI(OsClamp(ui.router.services.sync.Device.Dpi-3, 30, 5000))
		keys.ZoomSub = false
	}
	if keys.ZoomDef {
		ui.router.services.sync.Upload_deviceDPI(GetDeviceDPI())
		keys.ZoomDef = false
	}
	if keys.F2 {
		ui.router.services.sync.Upload_deviceStats(!ui.router.services.sync.Device.Stats)
		keys.F2 = false
	}
	if keys.F11 {
		ui.router.services.sync.Upload_deviceFullscreen(!ui.router.services.sync.Device.Fullscreen)
		keys.F11 = false
	}

	ui.win.fullscreen = ui.router.services.sync.Device.Fullscreen

	if !ui.winRect.Cmp(winRect) {
		ui.SetRefresh()
		ui.winRect = winRect //update
	}
}

func (ui *Ui) GetPalette() *DevPalette {
	return ui.router.services.sync.GetPalette()
}

func (ui *Ui) SetRefresh() {
	ui.refresh = true
}

func (ui *Ui) SetRelayoutHard() {
	ui.relayout_hard = true
}
func (ui *Ui) SetRelayoutSoft() {
	ui.relayout_soft = true
}

func (ui *Ui) SetRedrawBuffer() {
	ui.redrawBuffer = true
}

func (ui *Ui) NeedRedraw() bool {
	redraw := false

	if ui.tooltip.NeedRedraw() {
		ui.SetRedrawBuffer()
	}

	if ui.redrawBuffer {
		redraw = true
	}

	return redraw
}

func (ui *Ui) ResetIO() {
	ui.touch.Reset()
	ui.drag.Reset()
}

func (ui *Ui) Cell() int {
	return int(float64(ui.router.services.sync.Device.Dpi) / 2.5)
}
func (ui *Ui) CellWidth(width float64) int {
	t := OsRoundHalf(width * float64(ui.Cell())) // cell is ~34
	if width > 0 && t <= 0 {
		t = 1 //at least 1px
	}
	return t
}
func (ui *Ui) GetScrollThickness() int {
	return int(float64(ui.Cell()) * float64(ui.router.services.sync.Device.ScrollThick))
}

func (layout *Layout) _draw() {
	layout.buffer = nil

	canvas := Rect{X: 0, Y: 0, W: layout._getWidth(), H: layout._getHeight()}
	if layout.fnDraw != nil && canvas.Is() {
		paint := layout.fnDraw(canvas, layout)
		layout.buffer = paint.buffer
	}

	for _, it := range layout.childs {
		it._draw()
	}
}

func (ui *Ui) _relayout(layout *Layout) {

	layout.projectScroll()

	layout._relayoutCards()
	layout.updateFromChildCols(false)
	layout._relayout()
	layout.updateFromChildRows()
	layout._relayout()

	layout._relayoutCards()
	layout.updateFromChildCols(false)
	layout._relayout()
	layout.updateFromChildRows()
	layout._relayout()

	layout._relayoutCards()
	layout.updateFromChildCols(true)
	layout._relayout()
	layout.updateFromChildRows()
	layout._relayout()

	layout._draw()
}

func (ui *Ui) drawInner() {
	buff := ui.GetWin().buff

	buff.AddCrop(ui.mainLayout.CropWithScroll())
	buff.AddRect(buff.crop, ui.GetPalette().B, 0)

	//base
	ui.mainLayout._drawBuffers()

	//dialogs
	for _, dia := range ui.settings.Dialogs {
		layDia := ui.mainLayout.FindUID(dia.UID)
		if layDia != nil {
			layApp := layDia.GetApp()
			if layApp != nil {
				//alpha grey background
				backCanvas := layApp.CropWithScroll()
				buff.StartLevel(layDia.CropWithScroll(), ui.GetPalette().B, backCanvas, ui.CellWidth(layApp.getRounding()))
			}

			layDia._drawBuffers() //add renderToTexture optimalization ....
		}
	}

	//selection
	ui.selection.Draw(buff, ui)

	keys := ui.win.io.Keys
	if keys.Ctrl && keys.Shift {
		n := 0

		ui.GetTopLayout().postDraw(0, &n) //only top
	}
}
func (ui *Ui) Draw() {
	if ui.tooltip.touch(ui) {
		ui.SetRedrawBuffer()
	}

	win := ui.GetWin()
	win.buff.StartLevel(ui.mainLayout.canvas, ui.GetPalette().B, OsV4{}, 0)

	ui.drawInner()
	if win.io.Keys.Ctrl && !win.io.Keys.Shift {
		ui.GetWin().PaintCursor("cross") //brush
	}

	ui.tooltip.draw(ui)
	ui.GetWin().buff.FinalDraw()

	if !OsIsTicksIn(ui.maintenance_tick, 2000) {
		ui.GetWin().gph.Maintenance() //slow ....
		ui.maintenance_tick = OsTicks()
	}
}

func (ui *Ui) Tick() {
	if ui.win.io.Touch.Start {
		ui.ResetIO()
	}

	ui.redrawBuffer = false

	//shortcut
	ui.edit.shortcut_triggered = false
	keys := &ui.win.io.Keys
	if keys.HasChanged && (keys.Ctrl || keys.Plus || keys.Minus || keys.PageBackward || keys.PageForward) {
		var sh rune
		if keys.CtrlChar != "" {
			sh = rune(keys.CtrlChar[0])
		}
		if keys.Tab {
			sh = '\t'
		}
		if keys.ArrowL {
			sh = 37
		}
		if keys.ArrowU {
			sh = 38
		}
		if keys.ArrowR {
			sh = 39
		}
		if keys.ArrowD {
			sh = 40
		}
		if keys.Plus {
			sh = '+'
		}
		if keys.Minus {
			sh = '-'
		}
		if keys.PageBackward {
			sh = '←'
		}
		if keys.PageForward {
			sh = '→'
		}

		if sh != 0 {
			lay := ui.mainLayout.FindShortcut(sh)
			if lay != nil {
				if lay.fnInput != nil {
					lay.fnInput(LayoutInput{Shortcut_key: sh}, lay)
				}
				ui.edit.shortcut_triggered = true
			}
		}
	}

	addBrush := ui.selection.UpdateBrush(ui)
	if addBrush != nil {
		ui.SetRefresh()
	}

	root_runPrompt := ui.runPrompt
	if root_runPrompt != "" {
		ui.SetRefresh()
		ui.runPrompt = ""
	}

	if ui.router.Tick() {
		ui.SetRefresh()
	}

	ui.settings.Highlight_text = ui.router.text_highlight

	if (OsTicks() - ui.last_layout_updates_ticks) > 100 { //every 100ms
		ui.GetTopLayout().CallLayoutUpdates() //"Root", "ShowRoot", ui.mainLayout.UID)

		ui.last_layout_updates_ticks = OsTicks()
	}

	if ui.refresh {
		ui.refresh = false
		ui.relayout_hard = false //is in _refresh()
		ui.relayout_soft = false

		//save activated editbox
		if ui.edit.IsActive() {
			ui.edit.send(false, ui)
		}

		fnDone := func(dataJs []byte, uiGob []byte, cmdsGob []byte, err error, start_time float64) {
			if err != nil {
				return
			}

			ui.stats.Build = OsTime() - start_time

			var uii UI
			err = LogsGobUnmarshal(uiGob, &uii)
			if err != nil {
				return
			}
			ui.temp_ui = &uii

			var cmds []ToolCmd
			err = LogsGobUnmarshal(cmdsGob, &cmds)
			if err != nil {
				return
			}
			ui.temp_cmds = append(ui.temp_cmds, cmds...)
		}

		type ShowRoot struct {
			AddBrush  *LayoutPick
			RunPrompt string
		}
		ui.router.CallBuildAsync(1, "Root", "ShowRoot", ShowRoot{AddBrush: addBrush, RunPrompt: root_runPrompt}, ui._addLayout_FnProgress, fnDone)
	}

	if !ui.touch.IsActive() {
		if ui.temp_ui != nil {
			new_dom := NewUiLayoutDOM_root(ui)

			ui.temp_ui.addLayout(new_dom, new_dom.UID)

			new_dom._build()
			ui.mainLayout = new_dom
			ui.SetRelayoutHard()

			ui.temp_ui = nil
		}
	}

	//run commands
	{
		for i := 0; i < len(ui.temp_cmds); i++ {
			if ui.temp_cmds[i].Exe(ui) {
				//remove
				ui.temp_cmds = slices.Delete(ui.temp_cmds, i, i+1)
				i--

				ui.SetRelayoutHard()
			}
		}
		ui.temp_cmds = nil
	}

	if ui.settings.IsChanged() {
		ui.SetRelayoutHard()
	}

	if ui.relayout_hard {
		ui.relayout_hard = false
		ui.relayout_soft = false

		st := OsTime()

		//base
		ui._relayout(ui.mainLayout)

		//dialogs
		for _, dia := range ui.settings.Dialogs {
			diaLay := ui.mainLayout.FindUID(dia.UID)
			if diaLay != nil {
				if diaLay.parent != nil {
					diaLay.parent.dialog = diaLay //update layout.dialog
				}
				ui._relayout(diaLay)
			}
		}

		//maintenance
		for i := len(ui.settings.Dialogs) - 1; i >= 0; i-- {
			diaLay := ui.mainLayout.FindUID(ui.settings.Dialogs[i].UID)
			if diaLay == nil {
				ui.settings.Dialogs = slices.Delete(ui.settings.Dialogs, i, i+1)
			}
		}

		ui.stats.Relayout_hard = OsTime() - st

		ui.mainLayout.SetTouchAll()

		ui.SetRedrawBuffer()
	}

	if ui.relayout_soft {
		ui.relayout_soft = false

		st := OsTime()

		ui.mainLayout.updateCoordSoft()

		//dialogs
		for _, dia := range ui.settings.Dialogs {
			diaLay := ui.mainLayout.FindUID(dia.UID)
			if diaLay != nil {
				diaLay.updateCoordSoft()
			}
		}

		ui.stats.Relayout_soft = OsTime() - st
	}

	ui.mainLayout.UpdateTouch()
	ui.mainLayout.TouchDialogs(ui.edit.uid, ui.touch.canvas)

	if ui.win.io.Touch.Start || (ui.edit.IsActive() && (ui.touch.IsCanvasActive() || ui.win.io.Keys.HasChanged)) {
		ui.mainLayout.textComp()
	}

	// close all levels
	if ui.win.io.Keys.Shift && ui.win.io.Keys.Esc {
		ui.ResetIO()
		ui.settings.CloseAllDialogs(ui)
		ui.win.io.Keys.Esc = false
	}

	// touch
	ui.settings.CloseTouchDialogs(ui)

	if len(ui.redrawLayouts) > 0 {

		redrawLayouts := ui.redrawLayouts
		ui.redrawLayouts = nil

		for _, huidsh := range redrawLayouts {
			it := ui.mainLayout.FindUID(huidsh)
			if it != nil {
				paint := it.fnDraw(Rect{X: 0, Y: 0, W: it._getWidth(), H: it._getHeight()}, it)
				it.buffer = paint.buffer
			}
		}

		ui.SetRedrawBuffer()
	}

	if ui.win.io.Touch.End {
		ui.ResetIO()
	}

	ui.edit.Tick()

	if ui.router.Flush() { //only changed done msgs
		ui.SetRefresh()
	}

}

func (ui *Ui) _addLayout_FnProgress(cmdsGob [][]byte, err error, start_time float64) {
	if err != nil {
		return
	}
	for _, gob := range cmdsGob {
		var cmds []ToolCmd
		err := LogsGobUnmarshal(gob, &cmds)
		if err == nil {
			ui.temp_cmds = append(ui.temp_cmds, cmds...)
		}
	}
}
func (ui *Ui) _addLayout_FnIODone(dataJs []byte, uiGob []byte, cmdsGob []byte, err error, start_time float64) {
	if err != nil {
		return
	}

	var cmds []ToolCmd
	err = LogsGobUnmarshal(cmdsGob, &cmds)
	if err == nil {
		ui.temp_cmds = append(ui.temp_cmds, cmds...)
	}

	ui.stats.Changed = OsTime() - start_time
}
