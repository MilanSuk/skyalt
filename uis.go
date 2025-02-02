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
	"fmt"
	"log"
	"time"
)

type UiError struct {
	Time_unix   int64
	Layout_hash uint64
	Error       string
}

type UiErrors struct {
	Items []UiError
}

type UiClients struct {
	win *Win

	edit_history UiTextHistoryArray
	touch        UiLayoutInput
	edit         UiLayoutEdit
	drag         UiLayoutDrag
	backup_cell  int

	palette_light WinCdPalette
	palette_dark  WinCdPalette

	ui *Ui
}

func NewUiClients(win *Win, port int) (*UiClients, error) {
	rs := &UiClients{win: win}

	rs.palette_light = g_theme_light
	rs.palette_dark = g_theme_dark

	rs.ui = NewUi(rs)
	rs.ui.SetRefresh()

	return rs, nil
}
func (rs *UiClients) Destroy() {
	rs.ExitWidgetsProcess()

	rs.ui.Destroy()

	WgFiles_Save()
}

func (rs *UiClients) ExitWidgetsProcess() {
	rs.Save()
}

func (rs *UiClients) Save() {
	rs.ui.Save() //layouts
}

func (rs *UiClients) NeedRedraw() bool {
	redraw := false

	if rs.ui.tooltip.NeedRedraw() {
		rs.ui.SetRedrawBuffer()
	}

	if rs.ui.redrawBuffer {
		redraw = true
	}

	return redraw
}

func (rs *UiClients) CallInput(props *Layout, in *LayoutInput) error {

	if rs.ui.last_root_layout == nil {
		return fmt.Errorf("last_root_layout == nil")
	}

	inLayout := rs.ui.last_root_layout._findHash(props.Hash)
	if inLayout == nil {
		log.Fatal(fmt.Errorf("layout %d not found", props.Hash))
	}

	if in.Shortcut_key != 0 {
		if inLayout.fnInput != nil {
			inLayout.fnInput(*in, inLayout)
		}

	} else if in.Drop_path != "" {
		if inLayout.dropFile != nil {
			inLayout.dropFile(in.Drop_path)
		}

	} else if in.SetDropMove {
		if inLayout.dropMove != nil {
			inLayout.dropMove(in.DropSrc_pos, in.DropDst_pos, in.DragSrc_source, in.DragDst_source)
		}

	} else if in.SetEdit {
		if inLayout.fnSetEditbox != nil {
			inLayout.fnSetEditbox(in.EditValue, in.EditEnter)
		}

	} else if in.Pick.Line > 0 {
		//in.Pick.time_sec = float64(time.Now().UnixMilli()) / 1000
		//OpenFile_AssistantChat().findPickOrAdd(in.Pick)
		//OpenFile_AssistantChat().AppName = in.PickApp

	} else if inLayout.fnInput != nil {
		inLayout.fnInput(*in, inLayout)
	}

	//recv cmds
	{
		cmds := _getCmds()
		rs.ui._executeCmds(cmds)
	}

	rs.ui.parent.CallGetEnv()
	rs.ui.SetRefresh()

	return nil
}

func (rs *UiClients) CallGetEnv() error {
	return nil
}
func (rs *UiClients) CallSetEnv() error {
	return nil
}

func (rs *UiClients) Tick() {
	if rs.win.io.Touch.Start {
		rs.ResetIO()
	}

	rs.ui.Tick() //!!!

	if rs.win.io.Touch.End {
		rs.ResetIO()
	}

	rs.maintenance()
}

func (rs *UiClients) Draw() {
	rs.ui.Draw()
}

func (rs *UiClients) GetEnv() *DeviceSettings {
	return OpenFile_DeviceSettings()
}

func (rs *UiClients) Cell() int {
	return int(float32(rs.GetEnv().Dpi) / 2.5)
}

func (rs *UiClients) GetPalette() *WinCdPalette {

	env := rs.GetEnv()
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
		return &rs.palette_light
	case "dark":
		return &rs.palette_dark
	}

	return &env.CustomPalette
}

func (rs *UiClients) UpdateIO() {

	env := rs.GetEnv()
	keys := rs.win.io.Keys
	if keys.ZoomAdd {
		env.Dpi = OsClamp(env.Dpi+3, 30, 5000)
		rs.CallSetEnv()
	}
	if keys.ZoomSub {
		env.Dpi = OsClamp(env.Dpi-3, 30, 5000)
		rs.CallSetEnv()
	}
	if keys.ZoomDef {
		env.Dpi = GetDeviceDPI()
		rs.CallSetEnv()
	}

	if keys.F2 {
		env.Stats = !env.Stats // switch
		rs.CallSetEnv()
	}
	if keys.F11 {
		env.Fullscreen = !env.Fullscreen // switch
		rs.CallSetEnv()
	}

	rs.win.fullscreen = env.Fullscreen
}

func (rs *UiClients) maintenance() {
	rs.edit.Maintenance(rs)
}

func (rs *UiClients) ResetIO() {
	rs.touch.Reset()
	rs.drag.Reset()
}
