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

type WinIniMap struct {
	Enable     bool
	Tiles_url  string
	Cache_path string

	Copyright     string
	Copyright_url string
}

type WinIniCloudChat struct {
	Enable  bool
	Api_key string

	ChatCompletion_url string
	STT_url            string
	TTS_url            string
}

type WinIniLocalChat struct {
	Folder string
	Addr   string
	Port   int
}

func (chat *WinIniCloudChat) Check() error {
	if !chat.Enable {
		return fmt.Errorf("service is disabled(goto Settings)")
	}

	if chat.Api_key == "" {
		return fmt.Errorf("API key is empty")
	}

	return nil
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

func (io *WinIO) setup(path string) error {
	return io.Save(path)
}

func (io *WinIO) Open(path string) error {

	if !OsFileExists(path) {
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

func (io *WinIO) Save(path string) error {

	file, err := json.MarshalIndent(&io.Ini, "", "\t")
	if err != nil {
		return fmt.Errorf("MarshalIndent() failed: %w", err)
	}

	err = os.WriteFile(path, file, 0644)
	if err != nil {
		return fmt.Errorf("WriteFile() failed: %w", err)
	}
	return nil
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

type WinCdPalette struct {
	P, S, T, E, B           color.RGBA
	OnP, OnS, OnT, OnE, OnB color.RGBA
}

const (
	CdPalette_White = uint8(0)

	CdPalette_P = uint8(1)
	CdPalette_S = uint8(2)
	CdPalette_T = uint8(3)
	CdPalette_E = uint8(4)
	CdPalette_B = uint8(5)
)

// light
func InitWinCdPalette_light() WinCdPalette {
	var pl WinCdPalette
	//Primary
	pl.P = color.RGBA{37, 100, 120, 255}
	pl.OnP = color.RGBA{255, 255, 255, 255}
	//Secondary
	pl.S = color.RGBA{85, 95, 100, 255}
	pl.OnS = color.RGBA{255, 255, 255, 255}
	//Tertiary
	pl.T = color.RGBA{90, 95, 115, 255}
	pl.OnT = color.RGBA{255, 255, 255, 255}
	//Err
	pl.E = color.RGBA{180, 40, 30, 255}
	pl.OnE = color.RGBA{255, 255, 255, 255}
	//Surface(background)
	pl.B = color.RGBA{250, 250, 250, 255}
	pl.OnB = color.RGBA{25, 27, 30, 255}
	return pl
}

// dark
func InitWinCdPalette_dark() WinCdPalette {
	var pl WinCdPalette
	pl.P = color.RGBA{150, 205, 225, 255}
	pl.OnP = color.RGBA{0, 50, 65, 255}

	pl.S = color.RGBA{190, 200, 205, 255}
	pl.OnS = color.RGBA{40, 50, 55, 255}

	pl.T = color.RGBA{195, 200, 220, 255}
	pl.OnT = color.RGBA{75, 35, 50, 255}

	pl.E = color.RGBA{240, 185, 180, 255}
	pl.OnE = color.RGBA{45, 45, 65, 255}

	pl.B = color.RGBA{25, 30, 30, 255}
	pl.OnB = color.RGBA{230, 230, 230, 255}
	return pl
}

func Color2_Aprox(s color.RGBA, e color.RGBA, t float32) color.RGBA {
	var self color.RGBA
	self.R = byte(float32(s.R) + (float32(e.R)-float32(s.R))*t)
	self.G = byte(float32(s.G) + (float32(e.G)-float32(s.G))*t)
	self.B = byte(float32(s.B) + (float32(e.B)-float32(s.B))*t)
	self.A = byte(float32(s.A) + (float32(e.A)-float32(s.A))*t)
	return self
}

func (pl *WinCdPalette) GetGrey(t float32) color.RGBA {
	return Color2_Aprox(pl.S, pl.OnS, t)
}

func (pl *WinCdPalette) GetCdOver(cd color.RGBA, inside bool, active bool) color.RGBA {
	if active {
		if inside {
			cd = Color2_Aprox(cd, pl.OnS, 0.4)
		} else {
			cd = Color2_Aprox(cd, pl.OnS, 0.3)
		}
	} else {
		if inside {
			cd = Color2_Aprox(cd, pl.S, 0.2)
		}
	}

	return cd
}

func (pl *WinCdPalette) GetCd2(cd color.RGBA, fade, enable, inside, active bool) color.RGBA {
	if fade || !enable {
		cd.A = 100
	}
	if enable {
		cd = pl.GetCdOver(cd, inside, active)
	}
	return cd
}

func (pl *WinCdPalette) GetCdI(i uint8) (color.RGBA, color.RGBA) {
	switch i {
	case CdPalette_White:
		return color.RGBA{255, 255, 255, 255}, color.RGBA{0, 0, 0, 255}
	case CdPalette_P:
		return pl.P, pl.OnP
	case CdPalette_S:
		return pl.S, pl.OnS
	case CdPalette_T:
		return pl.T, pl.OnT
	case CdPalette_E:
		return pl.E, pl.OnE
	case CdPalette_B:
		return pl.B, pl.OnB
	}

	return pl.P, pl.OnP
}

func (pl *WinCdPalette) GetCd(i uint8, fade, enable, inside, active bool) (color.RGBA, color.RGBA) {

	cd, onCd := pl.GetCdI(i)

	cd = pl.GetCd2(cd, fade, enable, inside, active)
	onCd = pl.GetCd2(onCd, fade, enable, inside, active)

	return cd, onCd
}
