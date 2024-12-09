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
	"log"
)

type Ui struct {
	parent *UiClients

	winRect OsV4

	dom *Layout3

	levels UiLevels

	tooltip UiTooltip

	SelectComp_active bool
	SelectGrid_active bool
	Selected_start    OsV2
	Selected_hash     uint64

	ShowGrid bool

	relayout     bool
	redrawBuffer bool

	hasBuildStarted  bool
	hasBuildFinished bool

	hasTouchesActive bool

	maintenance_tick int64
}

func NewUi(parent *UiClients) *Ui {
	ui := &Ui{parent: parent}

	ui.dom = NewUiLayoutDOM_root(ui)
	ui.levels = InitUiLevels(ui)

	ui.tooltip.dom = ui.dom

	ui.Open()

	return ui
}

func (ui *Ui) Destroy() {
	ui.levels.Destroy()
	ui.dom.Destroy()
}

func (ui *Ui) Open() bool {
	//settings
	err := ui.levels.Open()
	if err != nil {
		fmt.Printf("Open() failed: %v\n", err)
		return false
	}

	return true
}

func (ui *Ui) Save() bool {
	//settings
	err := ui.levels.Save()
	if err != nil {
		fmt.Printf("Save() failed: %v\n", err)
		return false
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

func (ui *Ui) IsLevelBase(dom *Layout3) bool {
	return ui.dom == dom
}

func (ui *Ui) SetRefresh() {
	ui.hasBuildStarted = false
	ui.hasBuildFinished = false
}
func (ui *Ui) SetRelayout() {
	ui.relayout = true
}
func (ui *Ui) SetRedraw() {
	ui.redrawBuffer = true
}

func (ui *Ui) _touch() {

	if ui.hasTouchesActive {

		err := ui.parent.client.WriteInt(NMSG_INPUT_UPDATE)
		if err != nil {
			log.Fatal(err)
		}

		{
			var cmds []LayoutCmd
			data, err := ui.parent.client.ReadArray()
			if err != nil {
				log.Fatal(err)
			}
			OsUnmarshal(data, &cmds)
			ui._executeCmds(cmds)
		}

		done, err := ui.parent.client.ReadInt()
		if err != nil {
			log.Fatal(err)
		}
		if done > 0 {
			ui.hasTouchesActive = false
			ui.SetRefresh()

			ui.parent.CallGetEnv()
		}
	}
}

func (ui *Ui) _refresh() {

	//build
	{
		if !ui.hasBuildStarted {
			ui.parent.client.WriteInt(NMSG_REFRESH_START)

			ui.hasBuildStarted = true
		} else if !ui.hasBuildFinished {

			ui.parent.client.WriteInt(NMSG_REFRESH_UPDATE)

			//send list of rects
			rects := make(map[uint64]Rect)
			ui.dom.extractRects(rects)
			err := ui.parent.client.WriteArray(OsMarshal(rects))
			if err != nil {
				log.Fatal(err)
			}

			//relayout
			{
				var layout Layout
				data, err := ui.parent.client.ReadArray()
				if err != nil {
					log.Fatal(err)
				}
				OsUnmarshal(data, &layout)

				//project
				ui.dom.project(&layout)

				//recompute layout
				ui.SetRelayout()
			}

			//recv paint buffers
			{
				buffs := make(map[uint64][]LayoutDrawPrim)
				data, err := ui.parent.client.ReadArray()
				if err != nil {
					log.Fatal(err)
				}
				OsUnmarshal(data, &buffs)

				if len(buffs) > 0 {
					fmt.Println("buff back", len(buffs))

					//set buffs
					for hash, buff := range buffs {
						it := ui.dom.FindHash(hash)
						if it != nil {
							it.buffer = buff
							//fmt.Println("-Buffer", it.props.Name, it.canvas.Size.Y)
							ui.SetRedraw()
						} else {
							fmt.Println("never should happen: buffs")
						}
					}
				}
			}

			//recv cmds
			{
				var cmds []LayoutCmd
				data, err := ui.parent.client.ReadArray()
				if err != nil {
					log.Fatal(err)
				}
				OsUnmarshal(data, &cmds)
				ui._executeCmds(cmds)
			}

			{
				done, err := ui.parent.client.ReadInt()
				if err != nil {
					log.Fatal(err)
				}
				if done > 0 {
					ui.hasBuildFinished = true
					ui.SetRedraw()
					fmt.Println("finished")

					ui.parent.CallGetEnv()
				}
			}
		}
	}

}

func (ui *Ui) _executeCmds(cmds []LayoutCmd) {

	edit := &ui.parent.edit

	for _, cmd := range cmds {

		var layout *Layout3
		if cmd.Hash != 0 {
			layout = ui.dom.FindHash(cmd.Hash)
			if layout == nil {
				fmt.Printf("warning: hash %d not found\n", cmd.Hash)
			}
		}

		switch cmd.Cmd {
		case "VScrollToTheTop":
			if layout != nil {
				layout.ScrollIntoTop_vertical()
			}
		case "VScrollToTheBottom":
			if layout != nil {
				layout.ScrollIntoBottom_vertical()
			}
		case "HScrollToTheTop":
			if layout != nil {
				layout.ScrollIntoTop_horizontal()
			}
		case "HScrollToTheBottom":
			if layout != nil {
				layout.ScrollIntoBottom_horizontal()
			}

		case "OpenDialogCentered":
			if layout != nil {
				layout.OpenDialogCentered()
			}

		case "OpenDialogRelative":
			if layout != nil {
				var parent_hash uint64
				err := json.Unmarshal([]byte(cmd.Param1), &parent_hash)
				if err != nil {
					continue
				}
				parent_layout := ui.dom.FindHash(parent_hash)
				if parent_layout == nil {
					fmt.Printf("warning: parent_hash %d not found\n", parent_hash)
					continue
				}

				layout.OpenDialogRelative(parent_layout)
			}

		case "OpenDialogOnTouch":
			if layout != nil {
				layout.OpenDialogOnTouch()
			}

		case "CloseDialog":
			if layout != nil {
				layout.CloseDialog()
			}

		case "SetClipboardText":
			ui.parent.win.SetClipboardText(cmd.Param1)

		case "Refresh":
			ui.SetRefresh()

		case "Compile":
			//ui.lastRecompileTicks = 0 //reset, so next Maintenance() will recompile ...

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

	ui.redrawBuffer = false

	if ui.parent.backup_cell != ui.Cell() { //zoom
		ui.SetRelayout()
		ui.SetRefresh() //without this _refresh() will send old(!) rects, but cell is new.
		ui.parent.backup_cell = ui.Cell()
	}

	if ui.relayout {
		ui.levels.Relayout()
		ui.relayout = false
	}

	ui._refresh()

	ui._touch()

	win := ui.GetWin()

	// close all levels
	if win.io.Keys.Shift && win.io.Keys.Esc {
		ui.parent.ResetIO()
		ui.levels.CloseAllDialogs()
		win.io.Keys.Esc = false
	}

	// touch
	ui.levels.CloseTouchDialogs()

	if ui.hasBuildFinished {
		ui.levels.TryCloseDialogs()
	}

	{
		topLevel := ui.levels.GetTopLevel()
		topLevel.domPtr.updateShortcut()
		topLevel.domPtr.updateTouch()

		touchHash := ui.parent.touch.canvas
		editHash := ui.parent.edit.hash

		var act *Layout3
		var actE *Layout3
		if editHash != 0 {
			actE = ui.dom.FindHash(editHash)
		}

		if touchHash != 0 {
			act = ui.dom.FindHash(touchHash)
		} else {
			act = topLevel.domPtr.findTouch()
		}

		if actE != nil {
			actE.touchComp()
		}
		if act != nil && act != actE {
			act.touchComp()
		}
	}

	ui.dom.textComp()

	if ui.GetWin().io.Touch.End {
		ui.Selected_hash = 0
		ui.SelectGrid_active = false
		ui.SelectComp_active = false
	}

	//ui.levels.TryCloseDialogs()

}

func (ui *Ui) Draw() {

	if ui.tooltip.touch() {
		ui.SetRedraw()
	}

	win := ui.GetWin()
	win.buff.StartLevel(ui.dom.canvas, ui.GetPalette().B, OsV4{})

	ui.levels.drawBuffer()
	ui.tooltip.draw()
	ui.GetWin().buff.FinalDraw()

	if ui.NeedMaintenance() {
		ui.Maintenance()
	}
}
