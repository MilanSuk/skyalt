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
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"time"
)

type Ui struct {
	win    *Win
	router *ToolsRouter

	winRect OsV4

	settings *UiSettings

	mainLayout *Layout

	refresh bool

	relayout     bool
	redrawBuffer bool

	maintenance_tick int64

	redrawLayouts []uint64 //hashes

	sync *UiSync

	datePage int64

	edit_history *UiTextHistoryArray
	tooltip      *UiTooltip
	touch        *UiInput
	edit         *UiEdit
	drag         *UiDrag
	selection    *UiSelection

	temp_ui   *UI
	temp_cmds []ToolCmd
}

func NewUi(win *Win, router *ToolsRouter) (*Ui, error) {
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

	var err error
	ui.sync, err = NewUiSync(router)
	if err != nil {
		return nil, err
	}

	ui.SetRefresh()

	return ui, nil
}
func (ui *Ui) Destroy() {
	ui.sync.Destroy()

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
			//fmt.Printf("Open() failed: %v\n", err)
			return false
		}

		if len(jsRead) > 0 {
			err := json.Unmarshal(jsRead, ui.settings)
			if err != nil {
				fmt.Printf("Unmarshal() %s file failed: %v\n", ui.GetSettingsPath(), err)
				return false
			}
		}
	}

	return true
}

func (ui *Ui) save() bool {

	//save editbox
	if ui.edit.IsActive() {
		ui.edit.send(false, ui)
	}

	//settings
	{
		_, err := Tools_WriteJSONFile(ui.GetSettingsPath(), ui.settings)
		if err != nil {
			fmt.Printf("Save(%s) failed: %v\n", ui.GetSettingsPath(), err)
			return false
		}
	}

	return true
}

func (ui *Ui) GetWin() *Win {
	return ui.win
}

func (ui *Ui) UpdateIO(winRect OsV4) {

	keys := ui.win.io.Keys
	if keys.ZoomAdd {
		ui.sync.Upload_deviceDPI(OsClamp(ui.sync.device_settings.Dpi+3, 30, 5000), ui.router)
		keys.ZoomAdd = false
	}
	if keys.ZoomSub {
		ui.sync.Upload_deviceDPI(OsClamp(ui.sync.device_settings.Dpi-3, 30, 5000), ui.router)
		keys.ZoomSub = false
	}
	if keys.ZoomDef {
		ui.sync.Upload_deviceDPI(GetDeviceDPI(), ui.router)
		keys.ZoomDef = false
	}
	if keys.F2 {
		ui.sync.Upload_deviceStats(!ui.sync.device_settings.Stats, ui.router)
		keys.F2 = false
	}
	if keys.F11 {
		ui.sync.Upload_deviceFullscreen(!ui.sync.device_settings.Fullscreen, ui.router)
		keys.F11 = false
	}

	ui.win.fullscreen = ui.sync.device_settings.Fullscreen

	if !ui.winRect.Cmp(winRect) {
		ui.SetRefresh()
		ui.winRect = winRect //update
	}
}

func (ui *Ui) SetRefresh() {
	ui.refresh = true
}

func (ui *Ui) SetRelayout() {
	ui.relayout = true
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
	return int(float64(ui.sync.device_settings.Dpi) / 2.5)
}
func (ui *Ui) CellWidth(width float64) int {
	t := int(width * float64(ui.Cell())) // cell is ~34
	if width > 0 && t <= 0 {
		t = 1 //at least 1px
	}
	return t
}
func (ui *Ui) GetScrollThickness() int {
	return int(float64(ui.Cell()) * float64(ui.sync.device_settings.ScrollThick))
}

func _draw(layout *Layout) {
	layout.buffer = nil

	canvas := Rect{X: 0, Y: 0, W: layout._getWidth(), H: layout._getHeight()}
	if layout.fnDraw != nil && canvas.Is() {
		paint := layout.fnDraw(canvas, layout)
		layout.buffer = paint.buffer
	}

	for _, it := range layout.childs {
		_draw(it)
	}
}

func (ui *Ui) _relayout(layout *Layout) {

	layout.projectScroll()

	//relayout
	layout.clearAutoColsRows()
	layout._relayout()

	_draw(layout)
}

func (ui *Ui) Draw() {
	if ui.tooltip.touch(ui) {
		ui.SetRedrawBuffer()
	}

	win := ui.GetWin()
	win.buff.StartLevel(ui.mainLayout.canvas, ui.sync.GetPalette().B, OsV4{}, 0)

	ui.mainLayout.Draw()
	if win.io.Keys.Ctrl {
		ui.GetWin().PaintCursor("cross")
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
	if keys.Ctrl && keys.HasChanged {
		var sh byte
		if keys.CtrlChar != "" {
			sh = keys.CtrlChar[0]
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

	brush := ui.selection.UpdateComp(ui)
	if brush != nil {
		ui.SetRefresh()
	}

	if ui.sync.NeedRefresh(ui.router) {
		ui.SetRefresh()
	}

	if ui.refresh {
		ui.refresh = false
		ui.relayout = false //is in _refresh()

		//save activated editbox
		if ui.edit.IsActive() {
			ui.edit.send(false, ui)
		}

		fnProgress := func(cmds []ToolCmd, err error, start_time float64) {
			if err != nil {
				return
			}
			ui.temp_cmds = append(ui.temp_cmds, cmds...)
		}
		fnDone := func(bytes []byte, uii *UI, cmds []ToolCmd, err error, start_time float64) {

			fmt.Printf("_refresh(): %.4fsec\n", OsTime()-start_time)

			if err != nil {
				return
			}

			if uii != nil {
				ui.temp_ui = uii
			}

			ui.temp_cmds = append(ui.temp_cmds, cmds...)
		}

		type ShowRoot struct {
			AddBrush *LayoutPick
		}
		ui.router.CallAsync(1, "ShowRoot", ShowRoot{AddBrush: brush}, fnProgress, fnDone)
	}

	if !ui.touch.IsActive() {
		if ui.temp_ui != nil {
			ui.sync.ReloadSettings(ui.router)

			new_dom := NewUiLayoutDOM_root(ui)

			fnProgress := func(cmds []ToolCmd, err error, start_time float64) {
				if err != nil {
					return
				}
				ui.temp_cmds = append(ui.temp_cmds, cmds...)
			}
			fnDone := func(bytes []byte, uii *UI, cmds []ToolCmd, err error, start_time float64) {
				if err != nil {
					return
				}
				ui.temp_cmds = append(ui.temp_cmds, cmds...)
				fmt.Printf("_changed(): %.4fsec\n", OsTime()-start_time)
			}

			ui.temp_ui.addLayout(new_dom, "ShowRoot", 1, fnProgress, fnDone)

			new_dom._build()
			ui.mainLayout = new_dom
			ui.SetRelayout()

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

				ui.SetRelayout()
			}
		}
		ui.temp_cmds = nil
	}

	if ui.settings.IsChanged() {
		ui.SetRelayout()
	}

	if ui.relayout {
		ui.relayout = false

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
			} else {
				fmt.Println("dialog not found")
			}
		}

		fmt.Printf("_relayout(): %.4fsec\n", OsTime()-st)

		ui.mainLayout.SetTouchAll()

		ui.SetRedrawBuffer()
	}

	ui.mainLayout.UpdateTouch()
	ui.mainLayout.TouchDialogs(ui.edit.uid, ui.touch.canvas)

	ui.mainLayout.textComp()

	// close all levels
	if ui.win.io.Keys.Shift && ui.win.io.Keys.Esc {
		ui.ResetIO()
		ui.settings.CloseAllDialogs()
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
