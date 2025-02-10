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
	"encoding/json"
	"fmt"
	"os"
)

const (
	NMSG_EXIT = 0
	NMSG_SAVE = 1

	NMSG_GET_ENV = 10
	NMSG_SET_ENV = 11

	NMSG_REFRESH = 20
	NMSG_REDRAW  = 30

	NMSG_INPUT = 40
)

type UiLayoutDraw struct {
	Hash   uint64
	Rect   Rect
	Buffer []LayoutDrawPrim
}

type Ui struct { //put into UiClients(and rename it) ........
	parent *UiClients

	winRect OsV4

	dom *Layout3

	settings UiSettings

	tooltip UiTooltip

	selection UiSelection

	refresh_next_time float64
	relayout          bool
	redrawBuffer      bool

	redrawHashes []UiLayoutDraw

	maintenance_tick int64

	last_root_layout *Layout
}

func NewUi(parent *UiClients) *Ui {
	ui := &Ui{parent: parent}

	ui.dom = NewUiLayoutDOM_root(ui)
	ui.settings.Layouts.Init()

	ui.tooltip.dom = ui.dom

	ui.Open()

	return ui
}

func (ui *Ui) Destroy() {
	ui.dom.Destroy()
}

func (ui *Ui) GetSettingsPath() string {
	return "layouts/Root.json"
}

func (ui *Ui) Open() bool {
	//settings
	{
		jsRead, err := os.ReadFile(ui.GetSettingsPath())
		if err != nil {
			fmt.Printf("Open() failed: %v\n", err)
			return false
		}

		if len(jsRead) > 0 {
			err := json.Unmarshal(jsRead, &ui.settings)
			if err != nil {
				fmt.Printf("Open() failed: %v\n", err)
				return false
			}
		}
	}
	return true
}

func (ui *Ui) Save() bool {
	//settings
	{
		js, err := json.MarshalIndent(ui.settings, "", "\t")
		if err != nil {
			fmt.Printf("Save() failed: %v\n", err)
			return false
		}
		err = os.WriteFile(ui.GetSettingsPath(), js, 0644)
		if err != nil {
			fmt.Printf("Save() failed: %v\n", err)
			return false
		}
	}
	return true
}

func (ui *Ui) GetWin() *Win {
	return ui.parent.win
}

func (ui *Ui) GetPalette() *WinCdPalette {
	return ui.parent.GetPalette()
}

func (ui *Ui) Cell() int {
	return ui.parent.Cell()
}
func (ui *Ui) CellWidth(width float64) int {
	t := int(width * float64(ui.Cell())) // cell is ~34
	if width > 0 && t <= 0 {
		t = 1 //at least 1px
	}
	return t
}

func (ui *Ui) GetTextSize(cur int, ln string, prop WinFontProps) OsV2 {
	return ui.GetWin().gph.GetTextSize(prop, cur, ln)
}
func (ui *Ui) GetTextSizeMax(text string, max_line_px int, prop WinFontProps) (int, int) {
	tx := ui.GetWin().gph.GetTextMax(text, max_line_px, prop)
	if tx == nil {
		return 0, 1
	}

	return tx.max_size_x, len(tx.lines)
}
func (ui *Ui) GetTextLines(text string, max_line_px int, prop WinFontProps) []WinGphLine {
	tx := ui.GetWin().gph.GetTextMax(text, max_line_px, prop)
	if tx == nil {
		return []WinGphLine{{s: 0, e: len(text)}}
	}

	return tx.lines
}
func (ui *Ui) GetTextNumLines(text string, max_line_px int, prop WinFontProps) int {
	tx := ui.GetWin().gph.GetTextMax(text, max_line_px, prop)
	if tx == nil {
		return 1
	}

	return len(tx.lines)
}

func (ui *Ui) GetTextPos(touchPx int, ln string, prop WinFontProps, coord OsV4, align OsV2) int {
	start := ui.GetTextStart(ln, prop, coord, align.X, align.Y, 1)

	return ui.GetWin().gph.GetTextPos(prop, (touchPx - start.X), ln)
}

func (ui *Ui) GetTextStartLine(ln string, prop WinFontProps, coord OsV4, align OsV2, numLines int) OsV2 {
	lnSize := ui.GetTextSize(-1, ln, prop)
	size := OsV2{lnSize.X, numLines * prop.lineH}
	return coord.Align(size, align)
}

func (ui *Ui) GetTextStart(ln string, prop WinFontProps, coord OsV4, align_h, align_v int, numLines int) OsV2 {
	//lineH
	lnSize := ui.GetTextSize(-1, ln, prop)
	size := OsV2{lnSize.X, numLines * prop.lineH}
	start := coord.Align(size, OsV2{align_h, align_v})

	//letters
	coord.Start = start
	coord.Size.X = size.X
	coord.Size.Y = prop.lineH
	start = coord.Align(lnSize, OsV2{align_h, 1}) //letters must be always in the middle of line

	return start
}

func (ui *Ui) NeedMaintenance() bool {
	return !OsIsTicksIn(ui.maintenance_tick, 1000)
}

func (ui *Ui) Maintenance() {
	ui.GetWin().gph.Maintenance() //slow ...

	ui.maintenance_tick = OsTicks()
}

func (ui *Ui) SetRefreshTime(next_time float64) {
	ui.refresh_next_time = next_time
}

func (ui *Ui) SetRefresh() {
	ui.SetRefreshTime(1) //as soon as possible
}
func (ui *Ui) SetRelayout() {
	ui.relayout = true
}

func (ui *Ui) SetRedrawLayout(hash uint64) {
	//find
	for _, it := range ui.redrawHashes {
		if it.Hash == hash {
			return //already in
		}
	}

	ui.redrawHashes = append(ui.redrawHashes, UiLayoutDraw{Hash: hash})
}

func (ui *Ui) SetRedrawBuffer() {
	ui.redrawBuffer = true
}

func (ui *Ui) _refreshLayout(newDia *Layout3) {
	lay := ui.last_root_layout._findHash(newDia.props.Hash)
	if lay == nil {
		return
	}

	//parent hash
	{
		lay_parent := ui.last_root_layout._findParent(lay)
		if lay_parent != nil {
			layParent := ui.dom.FindHash(lay_parent.Hash)
			if layParent != nil {
				newDia.parent = layParent
				layParent.dialog = newDia
			}
		}
	}

	//client.WriteArray(OsMarshal(lay))
	st := OsTime()
	//project & clear buffers
	newDia.project(lay)

	//relayout
	newDia.relayout(true)

	fmt.Printf("project() + Relayout(): %.4fsec\n", (OsTime() - st))

	//draw
	rects := make(map[uint64]Rect)
	newDia.extractRects(rects)

	out_buffs := make(map[uint64][]LayoutDrawPrim)

	st = OsTime()
	_draw(lay, rects, out_buffs)
	fmt.Println("_draw()", (OsTime() - st), "sec")

	//recv paint buffers
	{
		if len(out_buffs) > 0 {
			//fmt.Println("buff back", len(buffs))

			//set buffs
			for hash, buff := range out_buffs {
				it := newDia.FindHash(hash)
				if it != nil {
					it.buffer = buff
					//fmt.Println("-Buffer", it.props.Name, it.canvas.Size.Y)
					ui.SetRedrawBuffer()
				} else {
					fmt.Println("never should happen: buffs")
				}
			}
		}
	}
}

func (ui *Ui) _refresh() {

	st := OsTime()
	if ui.refresh_next_time == 0 || ui.refresh_next_time > st {
		return
	}

	ui.refresh_next_time = 0

	ui.last_root_layout = _newLayoutRoot()
	ui.last_root_layout.fnBuild = OpenFile_Root().Build
	ui.last_root_layout.fnInput = OpenFile_Root().Input
	_build(ui.last_root_layout)

	ui._refreshLayout(ui.dom)

	for _, dia := range ui.settings.Dialogs {

		newDia := NewUiLayoutDOM(Layout{Hash: dia.Hash}, nil, ui)

		ui._refreshLayout(newDia)
	}

	//recv cmds
	{
		cmds := _getCmds()
		ui._executeCmds(cmds)
	}

	ui.parent.CallGetEnv()

	ui.dom.SetTouchAll()

	fmt.Printf("Refreshed: %.4fsec\n", OsTime()-st)
}

func (ui *Ui) _redrawHashes() {
	if len(ui.redrawHashes) == 0 {
		return
	}

	st := OsTime()

	//update rects
	for i, it := range ui.redrawHashes {
		lay := ui.dom.FindHash(it.Hash)
		if lay != nil {
			ui.redrawHashes[i].Rect = lay._getRect()
		}
	}

	//redraw
	for i, it := range ui.redrawHashes {
		lay := ui.last_root_layout._findHash(it.Hash)
		if lay != nil && lay.fnDraw != nil && it.Rect.Is() {
			ui.redrawHashes[i].Buffer = lay.fnDraw(it.Rect, lay).buffer
		}
	}

	//project
	for _, it := range ui.redrawHashes {
		lay := ui.dom.FindHash(it.Hash)
		if lay != nil {
			lay.buffer = it.Buffer
		}
	}

	//reset
	ui.redrawHashes = nil
	ui.SetRedrawBuffer()

	//recv cmds
	{
		cmds := _getCmds()
		ui._executeCmds(cmds)
	}

	fmt.Printf("RedrawHashes: %.4fsec\n", OsTime()-st)
}

func (ui *Ui) _executeCmds(cmds []LayoutCmd) {

	edit := &ui.parent.edit

	for _, cmd := range cmds {

		switch cmd.Cmd {
		case "VScrollToTheTop":
			layout := ui.dom.FindHash(cmd.Hash)
			if layout != nil {
				layout.ScrollIntoTop_vertical()
			} else {
				fmt.Printf("warning: hash %d not found\n", cmd.Hash)
			}

		case "VScrollToTheBottom":
			layout := ui.dom.FindHash(cmd.Hash)
			if layout != nil {
				layout.ScrollIntoBottom_vertical()
			} else {
				fmt.Printf("warning: hash %d not found\n", cmd.Hash)
			}

		case "HScrollToTheLeft":
			layout := ui.dom.FindHash(cmd.Hash)
			if layout != nil {
				layout.ScrollIntoTop_horizontal()
			} else {
				fmt.Printf("warning: hash %d not found\n", cmd.Hash)
			}

		case "HScrollToTheRight":
			layout := ui.dom.FindHash(cmd.Hash)
			if layout != nil {
				layout.ScrollIntoBottom_horizontal()
			} else {
				fmt.Printf("warning: hash %d not found\n", cmd.Hash)
			}

		case "OpenDialogCentered":
			ui.settings.OpenDialog(cmd.Hash, 0, OsV2{})

		case "OpenDialogRelative":
			var parent_hash uint64
			OsUnmarshal([]byte(cmd.Param1), &parent_hash)
			ui.settings.OpenDialog(cmd.Hash, parent_hash, OsV2{})

		case "OpenDialogOnTouch":
			ui.settings.OpenDialog(cmd.Hash, 0, ui.GetWin().io.Touch.Pos)

		case "CloseDialog":
			dia := ui.settings.FindDialog(cmd.Hash)
			if dia != nil {
				ui.settings.CloseDialog(dia)
			}

		case "SetClipboardText":
			ui.parent.win.SetClipboardText(cmd.Param1)

		case "Redraw":
			ui.SetRedrawLayout(cmd.Hash)

		case "RefreshDelayed":
			ui.SetRefreshTime(OsTime() + 0.5)

		//case "Compile":
		//	ui.parent.compile.recompile = true

		case "CopyText":
			edit.KeyCopy = true

		case "SelectAllText":
			edit.KeySelectAll = true

		case "CutText":
			edit.KeyCut = true

		case "PasteText":
			edit.KeyPaste = true
		}
	}
}

func (ui *Ui) Tick() {
	win := ui.GetWin()

	ui.redrawBuffer = false

	if ui.parent.backup_cell != ui.Cell() { //zoom
		ui.SetRelayout()
		ui.SetRefresh() //without this _refresh() will send old(!) rects, but cell is new.
		ui.parent.backup_cell = ui.Cell()
	}

	//shortcut
	keys := &win.io.Keys
	if !ui.parent.edit.IsActive() && keys.Ctrl && keys.HasChanged {
		var sh byte
		if keys.CtrlChar != "" {
			sh = keys.CtrlChar[0]
		}
		if keys.Tab {
			sh = '\t'
		}

		if sh != 0 {
			lay := ui.dom.FindShortcut(sh)
			if lay != nil {
				in := LayoutInput{Shortcut_key: sh}
				ui.parent.CallInput(&lay.props, &in)
			}
		}
	}

	if ui.relayout {
		ui.dom.relayout(true)
		ui.relayout = false
	}

	ui._refresh()
	if ui.last_root_layout != nil {
		ui._redrawHashes()
	}

	ui.dom.UpdateTouch()
	ui.dom.TouchDialogs(ui.parent.edit.hash, ui.parent.touch.canvas)

	ui.selection.UpdateComp(ui)

	ui.dom.textComp()

	// close all levels
	if win.io.Keys.Shift && win.io.Keys.Esc {
		ui.parent.ResetIO()
		ui.settings.CloseAllDialogs()
		win.io.Keys.Esc = false
	}

	// touch
	if ui.settings.CloseTouchDialogs(ui) {
		ui.dom.SetTouchAll()
	}
}

func (ui *Ui) Draw() {
	if ui.tooltip.touch() {
		ui.SetRedrawBuffer()
	}

	win := ui.GetWin()
	win.buff.StartLevel(ui.dom.canvas, ui.GetPalette().B, OsV4{})

	ui.dom.Draw()
	ui.tooltip.draw()
	ui.GetWin().buff.FinalDraw()

	if ui.NeedMaintenance() {
		ui.Maintenance()
	}
}
