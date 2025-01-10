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

type UiEnv struct {
	DateFormat  string
	Volume      float64
	Dpi         int
	Dpi_default int

	Fullscreen bool
	Stats      bool

	Theme             string
	ThemePalette      WinCdPalette
	CustomPalette     WinCdPalette
	UseDarkTheme      bool
	UseDarkThemeStart int //hours from midnight
	UseDarkThemeEnd   int
}

func (env *UiEnv) Check() bool {

	backup := *env

	//date
	if env.DateFormat == "" {
		_, zn := time.Now().Zone()
		zn = zn / 3600

		if zn <= -3 && zn >= -10 {
			env.DateFormat = "us"
		} else {
			env.DateFormat = "eu"
		}
	}

	//dpi
	if env.Dpi <= 0 {
		env.Dpi = GetDeviceDPI()
	}
	if env.Dpi < 30 {
		env.Dpi = 30
	}
	if env.Dpi > 5000 {
		env.Dpi = 5000
	}

	if env.Dpi_default != GetDeviceDPI() {
		env.Dpi_default = GetDeviceDPI()
	}

	//theme
	if env.Theme == "" {
		env.Theme = "light"
	}
	if env.CustomPalette.P.A == 0 {
		env.CustomPalette = InitWinCdPalette_light()
	}

	return *env == backup
}

type UiClients struct {
	win *Win

	edit_history UiTextHistoryArray
	touch        UiLayoutInput
	edit         UiLayoutEdit
	drag         UiLayoutDrag
	backup_cell  int

	refreshTicks int64

	palette_light WinCdPalette
	palette_dark  WinCdPalette

	ui *Ui

	env             UiEnv
	env_initialized bool

	server *NetServer
	client *NerServerClient

	compile *Compile
}

func NewUiClients(win *Win, port int) (*UiClients, error) {
	rs := &UiClients{win: win}

	rs.palette_light = InitWinCdPalette_light()
	rs.palette_dark = InitWinCdPalette_dark()

	rs.ui = NewUi(rs)
	rs.ui.SetRefresh()

	rs.server = NewNetServer(port)

	rs.compile = NewCompile(rs)

	return rs, nil
}
func (rs *UiClients) Destroy() {
	rs.ExitWidgetsProcess()
	rs.server.Destroy()

	rs.compile.Destroy()

	rs.ui.Destroy()
}

func (rs *UiClients) ExitWidgetsProcess() {
	rs.Save()
	if rs.client != nil {
		rs.client.WriteInt(NMSG_EXIT)
	}
}

func (rs *UiClients) Save() {
	rs.ui.Save() //layouts

	if rs.client != nil {
		rs.client.WriteInt(NMSG_SAVE) //widgets
	}
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
	err := rs.client.WriteInt(NMSG_INPUT)
	if err != nil {
		return err
	}
	err = rs.client.WriteInt(props.Hash)
	if err != nil {
		return err
	}
	err = rs.client.WriteArray(OsMarshal(in))
	if err != nil {
		return err
	}

	//recv cmds
	{
		var cmds []LayoutCmd
		data, err := rs.client.ReadArray()
		if err != nil {
			log.Fatal(err)
		}
		OsUnmarshal(data, &cmds)
		rs.ui._executeCmds(cmds)

		//fmt.Println("in", in)
		//fmt.Println("cmds", cmds)
	}

	rs.ui.parent.CallGetEnv()
	rs.ui.SetRefresh()

	return nil
}

func (rs *UiClients) CallGetEnv() error {
	rs.client.WriteInt(NMSG_GET_ENV)

	data, err := rs.client.ReadArray()
	if err != nil {
		return err
	}
	OsUnmarshal(data, &rs.env)

	if !rs.env.Check() {
		err = rs.CallSetEnv()
		if err != nil {
			return err
		}
	}

	return nil
}
func (rs *UiClients) CallSetEnv() error {
	err := rs.client.WriteInt(NMSG_SET_ENV)
	if err != nil {
		return err
	}
	err = rs.client.WriteArray(OsMarshal(rs.env))
	if err != nil {
		return err
	}

	return nil
}

func (rs *UiClients) _check_env_init() error {
	if !rs.env_initialized {

		err := rs.CallGetEnv() //check inside
		if err != nil {
			return err
		}

		rs.env_initialized = true
	}
	return nil
}

func (rs *UiClients) Tick() {

	err := rs.compile.Tick()
	if err != nil {
		fmt.Println(err)
		return
	}

	if !rs.compile.running.Load() {
		//show it's not running ...
		return
	}

	err = rs._check_env_init()
	if err != nil {
		fmt.Println(err)
		return
	}

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

func (rs *UiClients) GetEnv() *UiEnv {
	return &rs.env
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

	//refresh
	if !OsIsTicksIn(rs.refreshTicks, 1000) {
		rs.refreshTicks = OsTicks()
	}

}

func (rs *UiClients) ResetIO() {
	rs.touch.Reset()
	rs.drag.Reset()
}
