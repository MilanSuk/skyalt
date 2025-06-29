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
	"image/color"
	"os"

	"github.com/veandco/go-sdl2/sdl"
)

type WinKeys struct {
	HasChanged bool

	Text string

	CtrlChar string
	AltChar  string

	//Clipboard string

	Shift  bool
	Ctrl   bool
	Alt    bool
	Esc    bool
	Enter  bool
	ArrowU bool
	ArrowD bool
	ArrowL bool
	ArrowR bool
	Home   bool
	End    bool
	PageU  bool
	PageD  bool

	Tab   bool
	Space bool

	Delete    bool
	Backspace bool

	Copy      bool
	Cut       bool
	Paste     bool
	SelectAll bool
	RecordMic bool

	Backward bool
	Forward  bool

	F1  bool
	F2  bool
	F3  bool
	F4  bool
	F5  bool
	F6  bool
	F7  bool
	F8  bool
	F9  bool
	F10 bool
	F11 bool
	F12 bool

	ZoomAdd bool
	ZoomSub bool
	ZoomDef bool
}

type WinTouch struct {
	Pos       OsV2
	Wheel     int
	NumClicks int
	Start     bool
	End       bool
	Rm        bool // right/middle button

	Drop_path string

	wheel_last_sec float64
}

type WinCursor struct {
	name   string
	tp     sdl.SystemCursor
	cursor *sdl.Cursor
}

type WinIni struct {
	WinX, WinY, WinW, WinH int
}

type WinIO struct {
	Touch WinTouch
	Keys  WinKeys
	Ini   WinIni
}

func NewWinIO() (*WinIO, error) {
	var io WinIO

	err := io._IO_setDefault()
	if err != nil {
		return nil, fmt.Errorf("_IO_setDefault() failed: %w", err)
	}

	return &io, nil
}

func (io *WinIO) Destroy() error {
	return nil
}

func (io *WinIO) ResetTouchAndKeys() {
	io.Touch = WinTouch{}
	io.Keys = WinKeys{}
}

func (io *WinIO) setup(path string) {
	io.Save(path)
}

func (io *WinIO) Open(path string) error {

	if !Tools_IsFileExists(path) {
		io.setup(path)
	}

	//create ini if not exist
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("OpenFile() failed: %w", err)
	}
	f.Close()

	//load ini
	file, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("ReadFile() failed: %w", err)
	}

	if len(file) > 0 {
		err = json.Unmarshal(file, &io.Ini)
		if err != nil {
			return fmt.Errorf("WinIO - Unmarshal() failed: %w", err)
		}
	}

	err = io._IO_setDefault()
	if err != nil {
		return fmt.Errorf("_IO_setDefault() failed: %w", err)
	}

	return nil
}

func (io *WinIO) Save(path string) {
	_, err := Tools_WriteJSONFile(path, io.Ini)
	if err != nil {
		fmt.Printf("Save(%s) failed: %v\n", path, err)
	}

}

func (io *WinIO) _IO_setDefault() error {
	//window coord
	if io.Ini.WinW == 0 || io.Ini.WinH == 0 {
		io.Ini.WinX = 50
		io.Ini.WinY = 50
		io.Ini.WinW = 1280
		io.Ini.WinH = 720
	}

	return nil
}

func GetDeviceDPI() int {
	dpi, _, _, err := sdl.GetDisplayDPI(0)
	if err != nil {
		fmt.Printf("GetDisplayDPI() failed: %v\n", err)
		return 100
	}
	return int(dpi)
}

type DevPalette struct {
	P, S, E, B         color.RGBA
	OnP, OnS, OnE, OnB color.RGBA
}

func (pl *DevPalette) GetGrey(t float64) color.RGBA {
	return Color_Aprox(pl.B, pl.OnB, t)
}

func Color_Aprox(s color.RGBA, e color.RGBA, t float64) color.RGBA {
	var self color.RGBA
	self.R = byte(float64(s.R) + (float64(e.R)-float64(s.R))*t)
	self.G = byte(float64(s.G) + (float64(e.G)-float64(s.G))*t)
	self.B = byte(float64(s.B) + (float64(e.B)-float64(s.B))*t)
	self.A = byte(float64(s.A) + (float64(e.A)-float64(s.A))*t)
	return self
}
