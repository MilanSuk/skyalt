/*
Copyright 2023 Milan Suk

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

	"github.com/veandco/go-sdl2/sdl"
)

type Edit struct {
	uid, next_uid        *LayoutDiv
	setFirstEditbox, tab bool

	temp, orig string
	start, end OsV2

	last_edit string //for every SA_Editbox call
}

type Drag struct {
	div   *LayoutDiv
	group string
	id    uint64 //maybe string? ...
}

type Keys struct {
	text string

	ctrlChar string
	altChar  string

	clipboard string

	shift  bool
	ctrl   bool
	alt    bool
	esc    bool
	enter  bool
	arrowU bool
	arrowD bool
	arrowL bool
	arrowR bool
	home   bool
	end    bool
	pageU  bool
	pageD  bool

	tab bool

	delete    bool
	backspace bool

	copy      bool
	cut       bool
	paste     bool
	selectAll bool

	backward bool
	forward  bool

	f1  bool
	f2  bool
	f3  bool
	f4  bool
	f5  bool
	f6  bool
	f7  bool
	f8  bool
	f9  bool
	f10 bool
	f11 bool
	f12 bool
}

type Touch struct {
	pos       OsV2
	wheel     int
	numClicks uint8
	start     bool
	end       bool
	rm        bool // right/middle button

	drop_name string
	drop_path string
	drop_ext  string
}

type Poly struct {
	x []int16
	y []int16
}

func (v *OsV4) GetSDLRect() *sdl.Rect {
	return &sdl.Rect{X: int32(v.Start.X), Y: int32(v.Start.Y), W: int32(v.Size.X), H: int32(v.Size.Y)}
}

func (poly *Poly) Clear() {
	poly.x = poly.x[0:0]
	poly.y = poly.y[0:0]
}

func (poly *Poly) Add(pos OsV2) {
	poly.x = append(poly.x, int16(pos.X))
	poly.y = append(poly.y, int16(pos.Y))
}

type Cursor struct {
	name   string
	tp     sdl.SystemCursor
	cursor *sdl.Cursor
}

type Ini struct {
	Dpi         int
	Dpi_default int

	Date  int
	Theme int

	TimeZone int

	Fullscreen bool
	Stats      bool
	Grid       bool

	Languages              []string
	WinX, WinY, WinW, WinH int

	Hosting_enable bool
	Hosting_addr   string
}

type IO struct {
	touch Touch
	keys  Keys

	edit Edit
	drag Drag

	ini Ini
}

func NewIO() (*IO, error) {
	var io IO

	err := io._IO_setDefault()
	if err != nil {
		return nil, fmt.Errorf("_IO_setDefault() failed: %w", err)
	}

	io.edit.setFirstEditbox = true

	return &io, nil
}

func (io *IO) Destroy() error {
	return nil
}

func (io *IO) ResetTouchAndKeys() {
	io.touch = Touch{}
	io.keys = Keys{}
}

func (io *IO) Open(path string) error {

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
		err = json.Unmarshal(file, &io.ini)
		if err != nil {
			return fmt.Errorf("Unmarshal() failed: %w", err)
		}
	}

	err = io._IO_setDefault()
	if err != nil {
		return fmt.Errorf("_IO_setDefault() failed: %w", err)
	}
	return nil
}

func (io *IO) Save(path string) error {

	file, err := json.MarshalIndent(&io.ini, "", "")
	if err != nil {
		return fmt.Errorf("MarshalIndent() failed: %w", err)
	}

	err = os.WriteFile(path, file, 0644)
	if err != nil {
		return fmt.Errorf("WriteFile() failed: %w", err)
	}
	return nil
}

func _IO_getDPI() (int, error) {
	dpi, _, _, err := sdl.GetDisplayDPI(0)
	if err != nil {
		return 0, fmt.Errorf("GetDisplayDPI() failed: %w", err)
	}
	return int(dpi), nil
}

func (io *IO) _IO_setDefault() error {

	isDefault := (io.ini.Dpi_default == 0)

	io.SetDeviceDPI()

	//dpi
	if io.ini.Dpi == 0 {
		dpi, err := _IO_getDPI()
		if err != nil {
			return fmt.Errorf("_IO_getDPI() failed: %w", err)
		}
		io.ini.Dpi = dpi
	}

	//timezone
	if isDefault {
		io.ini.TimeZone = OsTimeZone()
	}

	//date
	if isDefault {
		io.ini.Date = OsTrn((OsTimeZone() <= -3 && OsTimeZone() >= -10), 1, 0)
	}

	//languages
	if len(io.ini.Languages) == 0 {
		io.ini.Languages = append(io.ini.Languages, "en")
	}

	//window coord
	if io.ini.WinW == 0 || io.ini.WinH == 0 {
		io.ini.WinX = 50
		io.ini.WinY = 50
		io.ini.WinW = 1280
		io.ini.WinH = 720
	}

	if len(io.ini.Hosting_addr) == 0 {
		io.ini.Hosting_addr = "localhost:8080"
		io.ini.Hosting_enable = true
	}

	return nil
}

func (io *IO) GetDPI() int {
	return OsClamp(io.ini.Dpi, 30, 5000)
}
func (io *IO) SetDPI(dpi int) {
	io.ini.Dpi = OsClamp(dpi, 30, 5000)
}

func (io *IO) Cell() int {
	return int(float32(io.GetDPI()) / 2.5) // 2.9f
}

func (io *IO) GetThemeCd() OsCd {

	cd := OsCd{90, 180, 180, 255} // ocean
	switch io.ini.Theme {
	case 1:
		cd = OsCd{200, 100, 80, 255}
	case 2:
		cd = OsCd{130, 170, 210, 255}
	case 3:
		cd = OsCd{130, 180, 130, 255}
	case 4:
		cd = OsCd{160, 160, 160, 255}
	}
	return cd
}

func (io *IO) GetCoord() OsV4 {
	return OsV4{Start: OsV2{}, Size: OsV2{X: io.ini.WinW, Y: io.ini.WinH}}
}

func (io *IO) SetDeviceDPI() error {
	dpi, err := _IO_getDPI()
	if err != nil {
		return fmt.Errorf("_IO_getDPI() failed: %w", err)
	}
	io.ini.Dpi_default = dpi
	return nil
}
