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
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unsafe"

	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

func InitSDLGlobal() error {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		return fmt.Errorf("sdl.Init() failed: %w", err)
	}

	err = ttf.Init()
	if err != nil {
		return fmt.Errorf("ttf.Init() failed: %w", err)
	}

	n, err := sdl.GetNumVideoDisplays()
	if err != nil {
		return fmt.Errorf("GetNumVideoDisplays() failed: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("0 video displays")
	}

	return nil
}
func DestroySDLGlobal() {
	ttf.Quit()
	sdl.Quit()
}

type Info struct {
	sum_frames int
	sum_time   int

	last_show_time int

	max_dt int

	out_worst_fps float64
	out_avg_fps   float64
}

func (ui *Info) Update(dt int) {

	time := OsTicks()

	ui.sum_time += OsMax(0, dt)

	if dt > ui.max_dt {
		ui.max_dt = dt
	}

	if ui.last_show_time+1000 < time { // every 1sec

		ui.out_worst_fps = OsTrnFloat(ui.max_dt > 0, 1/(float64(ui.max_dt)/1000.0), 0)
		ui.out_avg_fps = OsTrnFloat(ui.sum_time > 0, float64(ui.sum_frames)/(float64(ui.sum_time)/1000.0), 0)

		ui.sum_frames = 0
		ui.sum_time = 0
		ui.last_show_time = time
		ui.max_dt = 0
	}

	ui.sum_frames++
}

type Ui struct {
	io *IO

	window *sdl.Window
	render *sdl.Renderer

	startParticles bool
	particles      *Particles

	numClicks uint8

	redraw_num              int
	last_input_tick         int
	last_redraw_tick        int
	skip_draw_on_screen     bool
	num_skip_draw_on_screen int

	//poly   Poly
	images []*Image

	cursors []Cursor

	lastClickUp OsV2

	cursorId   int
	fullscreen bool

	recover_fullscreen_size OsV2

	cursorEdit          bool
	cursorTimeStart     float64
	cursorTimeEnd       float64
	cursorTimeLastBlink float64
	cursorCdA           byte
}

func IsCtrlActive() bool {
	state := sdl.GetKeyboardState()
	return state[sdl.SCANCODE_LCTRL] != 0 || state[sdl.SCANCODE_RCTRL] != 0
}

func IsShiftActive() bool {
	state := sdl.GetKeyboardState()
	return state[sdl.SCANCODE_LSHIFT] != 0 || state[sdl.SCANCODE_RSHIFT] != 0
}

func IsAltActive() bool {
	state := sdl.GetKeyboardState()
	return state[sdl.SCANCODE_LALT] != 0 || state[sdl.SCANCODE_RALT] != 0
}

func (ui *Ui) ResendInput() {
	if ui == nil {
		return
	}
	ui.redraw_num = 0
	ui.last_input_tick = OsTicks()
}

func (ui *Ui) GetMousePosition() OsV2 {

	x, y, _ := sdl.GetGlobalMouseState()

	w, h := ui.window.GetPosition()

	return OsV2_32(x, y).Sub(OsV2_32(w, h))
}

func (ui *Ui) GetScreenCoord() (OsV4, error) {

	w, h, err := ui.render.GetOutputSize()
	if err != nil {
		return OsV4{}, fmt.Errorf("GetOutputSize() failed: %w", err)
	}
	return OsV4{Start: OsV2{}, Size: OsV2_32(w, h)}, nil
}

func (ui *Ui) SaveScreenshot() error {

	w, h, err := ui.render.GetOutputSize()
	if err != nil {
		return fmt.Errorf("GetOutputSize() failed: %w", err)
	}

	surface, err := sdl.CreateRGBSurface(0, w, h, 32, 0, 0, 0, 0)
	if err != nil {
		return fmt.Errorf("CreateRGBSurface() failed: %w", err)
	}
	defer surface.Free()

	//copies pixels
	err = ui.render.ReadPixels(nil, surface.Format.Format, unsafe.Pointer(&surface.Pixels()[0]), int(surface.Pitch))
	if err != nil {
		return fmt.Errorf("ReadPixels() failed: %w", err)
	}
	img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{int(surface.W), int(surface.H)}})
	for y := int32(0); y < surface.H; y++ {
		for x := int32(0); x < surface.W; x++ {
			b := surface.Pixels()[y*surface.W*4+x*4+0] //blue 1st
			g := surface.Pixels()[y*surface.W*4+x*4+1]
			r := surface.Pixels()[y*surface.W*4+x*4+2] //red last
			img.SetRGBA(int(x), int(y), color.RGBA{r, g, b, 255})
		}
	}

	// creates file
	file, err := os.Create("screenshot_" + time.Now().Format("2006-1-2_15-4-5") + ".png")
	if err != nil {
		return fmt.Errorf("Create() failed: %w", err)
	}
	defer file.Close()

	//saves PNG
	err = png.Encode(file, img)
	if err != nil {
		return fmt.Errorf("Encode() failed: %w", err)
	}

	return nil
}

func (ui *Ui) NumTextures() int {
	n := 0
	for _, img := range ui.images {
		if img.texture != nil {
			n++
		}
	}
	return n
}

func (ui *Ui) GetImagesBytes() int64 {
	bytes := int64(0)
	for _, img := range ui.images {
		bytes += img.GetBytes()
	}
	return bytes
}

func (ui *Ui) StartupAnim() (bool, error) {
	if ui.particles != nil {
		ui.particles.StartAnim(5)
	}

	running := true
	for ui.particles != nil && ui.particles.num_draw > 0 && running && !ui.io.touch.start {

		// clear
		err := ui.render.SetDrawColor(220, 220, 220, 255)
		if err != nil {
			return false, fmt.Errorf("SetDrawColor() failed: %w", err)
		}

		err = ui.render.Clear()
		if err != nil {
			return false, fmt.Errorf("RenderClear() failed: %w", err)
		}

		// particles
		ui.particles.Update()
		_, err = ui.particles.Draw(OsCd{255, 255, 255, 255}, ui.render)
		if err != nil {
			return false, fmt.Errorf("Particles.Draw() failed: %w", err)
		}

		// finish frame
		ui.render.Present()

		running, _ = ui.Event()
	}

	if ui.particles != nil {
		ui.particles.Clear()
	}

	return running, nil
}

func NewUi(iniPath string) (*Ui, error) {
	var ui Ui
	var err error

	// start anim
	ui.startParticles = !OsFileExists(iniPath)

	ui.io, err = NewIO()
	if err != nil {
		return nil, fmt.Errorf("NewIO() failed: %w", err)
	}
	err = ui.io.Open(iniPath)
	if err != nil {
		return nil, fmt.Errorf("Open() failed: %w", err)
	}

	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "2")

	// create SDL
	ui.window, err = sdl.CreateWindow("SkyAlt", int32(ui.io.ini.WinX), int32(ui.io.ini.WinY), int32(ui.io.ini.WinW), int32(ui.io.ini.WinH), sdl.WINDOW_RESIZABLE)
	if err != nil {
		return nil, fmt.Errorf("CreateWindow() failed: %w", err)
	}
	ui.render, err = sdl.CreateRenderer(ui.window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		return nil, fmt.Errorf("CreateRenderer() failed: %w", err)
	}
	sdl.EventState(sdl.DROPFILE, sdl.ENABLE)
	sdl.StartTextInput()

	// particles
	ui.particles, err = NewParticles(ui.render)
	if err != nil {
		return nil, fmt.Errorf("NewParticles() failed: %w", err)
	}

	// cursors
	ui.cursors = append(ui.cursors, Cursor{"default", sdl.SYSTEM_CURSOR_ARROW, sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_ARROW)})
	ui.cursors = append(ui.cursors, Cursor{"hand", sdl.SYSTEM_CURSOR_HAND, sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_HAND)})
	ui.cursors = append(ui.cursors, Cursor{"ibeam", sdl.SYSTEM_CURSOR_IBEAM, sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_IBEAM)})
	ui.cursors = append(ui.cursors, Cursor{"cross", sdl.SYSTEM_CURSOR_CROSSHAIR, sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_CROSSHAIR)})

	ui.cursors = append(ui.cursors, Cursor{"res_col", sdl.SYSTEM_CURSOR_SIZEWE, sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_SIZEWE)})
	ui.cursors = append(ui.cursors, Cursor{"res_row", sdl.SYSTEM_CURSOR_SIZENS, sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_SIZENS)})
	ui.cursors = append(ui.cursors, Cursor{"res_nwse", sdl.SYSTEM_CURSOR_SIZENESW, sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_SIZENESW)}) // bug(already fixed) in SDL: https://github.com/libsdl-org/SDL/issues/2123
	ui.cursors = append(ui.cursors, Cursor{"res_nesw", sdl.SYSTEM_CURSOR_SIZENWSE, sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_SIZENWSE)})
	ui.cursors = append(ui.cursors, Cursor{"move", sdl.SYSTEM_CURSOR_SIZEALL, sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_SIZEALL)})

	ui.cursors = append(ui.cursors, Cursor{"wait", sdl.SYSTEM_CURSOR_WAITARROW, sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_WAITARROW)})
	ui.cursors = append(ui.cursors, Cursor{"no", sdl.SYSTEM_CURSOR_NO, sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_NO)})

	return &ui, nil
}

func (ui *Ui) Destroy() error {
	var err error

	err = ui.io.Destroy()
	if err != nil {
		return fmt.Errorf("IO.Destroy() failed: %w", err)
	}

	for _, img := range ui.images {
		err = img.Destroy()
		if err != nil {
			return fmt.Errorf("Image.Destroy() failed: %w", err)
		}
	}

	for _, cur := range ui.cursors {
		sdl.FreeCursor(cur.cursor)
	}

	if ui.particles != nil {
		err = ui.particles.Destroy()
		if err != nil {
			return fmt.Errorf("Particles.Destroy() failed: %w", err)
		}
	}

	err = ui.render.Destroy()
	if err != nil {
		return fmt.Errorf("Render.Destroy() failed: %w", err)
	}
	err = ui.window.Destroy()
	if err != nil {
		return fmt.Errorf("Window.Destroy() failed: %w", err)
	}

	return nil
}

func (ui *Ui) Event() (bool, error) {

	io := ui.io

	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() { // some cases have RETURN(don't miss it in tick), some (can be missed in tick)!

		switch val := event.(type) {
		case sdl.QuitEvent:
			fmt.Println("Exiting ..")
			return false, nil

		case sdl.WindowEvent:
			if val.Event == sdl.WINDOWEVENT_MOVED {
			} else if val.Event == sdl.WINDOWEVENT_SIZE_CHANGED {
				ui.ResendInput()
			} else if val.Event == sdl.WINDOWEVENT_SHOWN {
				ui.ResendInput()
			} else if val.Event == sdl.WINDOWEVENT_HIDDEN {
				ui.ResendInput()
			}

		case sdl.MouseMotionEvent:
			ui.ResendInput()

		case sdl.MouseButtonEvent:

			ui.numClicks = val.Clicks
			if val.Clicks > 1 {
				if ui.lastClickUp.Distance(OsV2_32(val.X, val.Y)) > float32(ui.Cell())/5 { //7px error space
					ui.numClicks = 1
				}
			}

			io.touch.pos = OsV2_32(val.X, val.Y)
			io.touch.rm = (val.Button != sdl.ButtonLeft)
			//io.touch.numClicks = val.Clicks

			if val.Type == sdl.MOUSEBUTTONDOWN {
				io.touch.start = true
				sdl.CaptureMouse(true) // keep getting info even mouse is outside window

			} else if val.Type == sdl.MOUSEBUTTONUP {

				ui.lastClickUp = io.touch.pos
				io.touch.end = true
				sdl.CaptureMouse(false)
			}

			ui.ResendInput()
			return true, nil

		case sdl.MouseWheelEvent:

			if IsCtrlActive() { // zoom

				if val.Y > 0 {
					io.SetDPI(io.GetDPI() + 3)
				}
				if val.Y < 0 {
					io.SetDPI(io.GetDPI() - 3)
				}
			} else {
				io.touch.wheel = -int(val.Y) // divide by -WHEEL_DELTA
			}

			ui.ResendInput()
			return true, nil

		case sdl.KeyboardEvent:
			if val.Type == sdl.KEYDOWN {

				if IsCtrlActive() {
					if val.Keysym.Sym == sdl.K_PLUS || val.Keysym.Sym == sdl.K_KP_PLUS {
						io.SetDPI(io.GetDPI() + 3)
					}
					if val.Keysym.Sym == sdl.K_MINUS || val.Keysym.Sym == sdl.K_KP_MINUS {
						io.SetDPI(io.GetDPI() - 3)
					}
					if val.Keysym.Sym == sdl.K_0 || val.Keysym.Sym == sdl.K_KP_0 {
						dpi, err := _IO_getDPI()
						if err == nil {
							io.SetDPI(dpi)
						}
					}
				}

				keys := &io.keys

				keys.esc = val.Keysym.Sym == sdl.K_ESCAPE
				keys.enter = (val.Keysym.Sym == sdl.K_RETURN || val.Keysym.Sym == sdl.K_RETURN2 || val.Keysym.Sym == sdl.K_KP_ENTER)

				keys.arrowU = val.Keysym.Sym == sdl.K_UP
				keys.arrowD = val.Keysym.Sym == sdl.K_DOWN
				keys.arrowL = val.Keysym.Sym == sdl.K_LEFT
				keys.arrowR = val.Keysym.Sym == sdl.K_RIGHT
				keys.home = val.Keysym.Sym == sdl.K_HOME
				keys.end = val.Keysym.Sym == sdl.K_END
				keys.pageU = val.Keysym.Sym == sdl.K_PAGEUP
				keys.pageD = val.Keysym.Sym == sdl.K_PAGEDOWN

				keys.copy = val.Keysym.Sym == sdl.K_COPY || (IsCtrlActive() && val.Keysym.Sym == sdl.K_c)
				keys.cut = val.Keysym.Sym == sdl.K_CUT || (IsCtrlActive() && val.Keysym.Sym == sdl.K_x)
				keys.paste = val.Keysym.Sym == sdl.K_PASTE || (IsCtrlActive() && val.Keysym.Sym == sdl.K_v)
				keys.selectAll = val.Keysym.Sym == sdl.K_SELECT || (IsCtrlActive() && val.Keysym.Sym == sdl.K_a)
				keys.backward = val.Keysym.Sym == sdl.K_AC_FORWARD || (IsCtrlActive() && !IsShiftActive() && val.Keysym.Sym == sdl.K_z)
				keys.forward = val.Keysym.Sym == sdl.K_AC_BACK || (IsCtrlActive() && val.Keysym.Sym == sdl.K_y) || (IsCtrlActive() && IsShiftActive() && val.Keysym.Sym == sdl.K_z)

				keys.tab = val.Keysym.Sym == sdl.K_TAB

				keys.delete = val.Keysym.Sym == sdl.K_DELETE
				keys.backspace = val.Keysym.Sym == sdl.K_BACKSPACE

				keys.f1 = val.Keysym.Sym == sdl.K_F1
				keys.f2 = val.Keysym.Sym == sdl.K_F2
				keys.f3 = val.Keysym.Sym == sdl.K_F3
				keys.f4 = val.Keysym.Sym == sdl.K_F4
				keys.f5 = val.Keysym.Sym == sdl.K_F5
				keys.f6 = val.Keysym.Sym == sdl.K_F6
				keys.f7 = val.Keysym.Sym == sdl.K_F7
				keys.f8 = val.Keysym.Sym == sdl.K_F8
				keys.f9 = val.Keysym.Sym == sdl.K_F9
				keys.f10 = val.Keysym.Sym == sdl.K_F10
				keys.f11 = val.Keysym.Sym == sdl.K_F11
				keys.f12 = val.Keysym.Sym == sdl.K_F12

				let := val.Keysym.Sym
				if OsIsTextWord(rune(let)) || let == ' ' {
					if IsCtrlActive() {
						keys.ctrlChar = string(let) //string([]byte{byte(let)})
					}
					if IsAltActive() {
						keys.altChar = string(let)
					}
				}
			}

			ui.ResendInput()
			return true, nil

		case sdl.TextInputEvent:
			if !(IsCtrlActive() && len(val.Text) > 0 && val.Text[0] == ' ') { // ignore ctrl+space
				io.keys.text += string(val.Text[:])
			}

			ui.ResendInput()
			return true, nil

		case sdl.DropEvent:
			io.touch.drop_path = val.File
			io.touch.drop_name = filepath.Base(val.File)
			io.touch.drop_ext = filepath.Ext(val.File)

			ui.ResendInput()
			return true, nil

		}
	}

	return true, nil
}

func (ui *Ui) Maintenance() {
	for i := len(ui.images) - 1; i >= 0; i-- {
		ok, _ := ui.images[i].Maintenance(ui.render)
		if !ok {
			ui.images = append(ui.images[:i], ui.images[i+1:]...)
		}
	}
}

func (ui *Ui) UpdateIO() (bool, error) {
	if ui == nil {
		return true, nil
	}

	// startup animation
	if ui.startParticles {
		ui.startParticles = false
		ok, err := ui.StartupAnim()
		if err != nil {
			return ok, fmt.Errorf("StartupAnim() failed: %w", err)
		}
		return ok, nil
	}

	ui.fullscreen = ui.io.ini.Fullscreen

	ok, err := ui.Event()
	if err != nil {
		return ok, fmt.Errorf("Event() failed: %w", err)
	}
	if !ok {
		return false, nil
	}

	// update Ui
	io := ui.io

	{
		start := OsV2_32(ui.window.GetPosition())
		size := OsV2_32(ui.window.GetSize())
		io.ini.WinX = start.X
		io.ini.WinY = start.Y
		io.ini.WinW = size.X
		io.ini.WinH = size.Y
	}

	io.SetDeviceDPI()

	if !io.touch.start && !io.touch.end && !io.touch.rm {
		io.touch.pos = ui.GetMousePosition()
	}
	io.touch.numClicks = ui.numClicks
	if io.touch.end {
		ui.numClicks = 0
	}

	// input.sleep = TRUE
	io.keys.shift = IsShiftActive()
	io.keys.alt = IsAltActive()
	io.keys.ctrl = IsCtrlActive()

	if io.keys.f2 {
		io.ini.Stats = !io.ini.Stats // switch
		ui.ResendInput()
	}

	if io.keys.f3 {
		io.ini.Grid = !io.ini.Grid // switch
		ui.ResendInput()
	}

	if io.keys.f8 {
		err := ui.SaveScreenshot()
		if err != nil {
			return true, fmt.Errorf("SaveScreenshot() failed: %w", err)
		}
	}

	if io.keys.f11 {
		io.ini.Fullscreen = !io.ini.Fullscreen // switch
		ui.ResendInput()
	}

	if io.keys.paste {
		text, err := sdl.GetClipboardText()
		if err != nil {
			fmt.Println("Error: UpdateIO.GetClipboardText() failed: %w", err)
		}
		io.keys.clipboard = strings.Trim(text, "\r")

	}

	ui.cursorId = 0

	return true, nil
}

func (ui *Ui) StartRender() error {
	if ui == nil {
		return nil
	}

	err := ui.render.SetClipRect(nil)
	if err != nil {
		return fmt.Errorf("SetClipRect() failed: %w", err)
	}

	err = ui.render.SetDrawColor(220, 220, 220, 255)
	if err != nil {
		return fmt.Errorf("SetDrawColor() failed: %w", err)
	}

	err = ui.render.Clear()
	if err != nil {
		return fmt.Errorf("RenderClear() failed: %w", err)
	}

	return nil
}

func (ui *Ui) EndRender() error {
	if ui == nil {
		return nil
	}

	if !ui.skip_draw_on_screen || ui.num_skip_draw_on_screen > 60 {
		ui.render.Present()
		ui.num_skip_draw_on_screen = 0
	} else {
		ui.num_skip_draw_on_screen++
	}
	ui.skip_draw_on_screen = false

	ui.last_redraw_tick = OsTicks()

	if ui.cursorId >= 0 {
		if ui.cursorId >= len(ui.cursors) {
			return errors.New("cursorID is out of range")
		}
		sdl.SetCursor(ui.cursors[ui.cursorId].cursor)
	}

	if ui.fullscreen != ui.io.ini.Fullscreen {
		fullFlag := uint32(0)
		if ui.io.ini.Fullscreen {
			ui.recover_fullscreen_size = OsV2_32(ui.window.GetSize())
			fullFlag = uint32(sdl.WINDOW_FULLSCREEN_DESKTOP)
		}
		err := ui.window.SetFullscreen(fullFlag)
		if err != nil {
			return fmt.Errorf("SetFullscreen() failed: %w", err)
		}
		if fullFlag == 0 {
			ui.window.SetSize(ui.recover_fullscreen_size.Get32()) //otherwise, wierd bug happens
		}
	}

	if len(ui.io.keys.clipboard) > 0 {
		sdl.SetClipboardText(ui.io.keys.clipboard)
	}

	ui.io.ResetTouchAndKeys()

	ui.Maintenance()
	return nil
}

func (ui *Ui) NeedRedraw() bool {
	if ui == nil {
		return true
	}

	redraw := (ui.redraw_num < 10)

	if !redraw && !OsIsTicksIn(ui.last_redraw_tick, 5000) { // redraws after 5sec without activity
		ui.redraw_num = 0
	}

	if ui.cursorEdit {
		if redraw {
			ui.cursorEdit = false
		}

		tm := OsTime()

		if ui.redraw_num == 0 {
			ui.cursorTimeEnd = tm + 5 //startAfterSleep/continue blinking after mouse move
		}

		if (tm - ui.cursorTimeStart) < 0.5 {
			//star
			ui.cursorCdA = 255
		} else if tm > ui.cursorTimeEnd {
			//sleep
			if ui.cursorCdA == 0 { //last draw must be full
				ui.cursorCdA = 255
				ui.redraw_num = 8 //redraw 1x
			}

		} else if (tm - ui.cursorTimeLastBlink) > 0.5 {
			//blink swap
			if ui.cursorCdA > 0 {
				ui.cursorCdA = 0
			} else {
				ui.cursorCdA = 255
			}
			ui.redraw_num = 8 //redraw 1x
			ui.cursorTimeLastBlink = tm
		}
	}

	ui.redraw_num++
	return redraw
}
func (ui *Ui) SetNoSleep() {
	if ui == nil {
		return
	}

	if OsTicks() < ui.last_input_tick+5000 { // can be use 5sec after last mouse/keyboard action
		ui.redraw_num = 0 // redraw
	}
}
func (ui *Ui) SetLayoutChange() {
	ui.skip_draw_on_screen = true
	ui.redraw_num = 0 // redraw
}

func (ui *Ui) PaintCursor(name string) error {
	if ui == nil {
		return nil
	}

	for i, cur := range ui.cursors {
		if strings.EqualFold(cur.name, name) {
			ui.cursorId = i

			//ui.addFrameHash_int(i)
			return nil
		}
	}

	return errors.New("Cursor(" + name + ") not found: ")
}

func _Ui_boxSE(render *sdl.Renderer, start OsV2, end OsV2, cd OsCd) {

	if start.X != end.X && start.Y != end.Y {
		gfx.BoxRGBA(render, int32(start.X), int32(start.Y), int32(end.X-1), int32(end.Y-1), cd.R, cd.G, cd.B, cd.A)
	}
}

func _Ui_boxSE_border(render *sdl.Renderer, start OsV2, end OsV2, cd OsCd, thick int) {
	_Ui_boxSE(render, start, OsV2{end.X, start.Y + thick}, cd) // top
	_Ui_boxSE(render, OsV2{start.X, end.Y - thick}, end, cd)   // bottom
	_Ui_boxSE(render, start, OsV2{start.X + thick, end.Y}, cd) // left
	_Ui_boxSE(render, OsV2{end.X - thick, start.Y}, end, cd)   // right
}

func _Ui_line(render *sdl.Renderer, start OsV2, end OsV2, thick int, cd OsCd) {

	v := end.Sub(start)
	if !v.IsZero() {

		hThick := thick / 2

		if thick == 1 {
			gfx.AALineRGBA(render, int32(start.X), int32(start.Y), int32(end.X), int32(end.Y), cd.R, cd.G, cd.B, cd.A)
		} else {
			l := v.Len()

			x := int(float32(v.Y)/l) * hThick
			y := int(float32(-v.X)/l) * hThick

			vx := []int16{int16(start.X + x), int16(end.X + x), int16(end.X - x), int16(start.X - x)}
			vy := []int16{int16(start.Y + y), int16(end.Y + y), int16(end.Y - y), int16(start.Y - y)}

			gfx.FilledPolygonRGBA(render, vx, vy, cd.R, cd.G, cd.B, cd.A)
		}

	}
}

func (ui *Ui) SetTextCursorMove() {
	ui.cursorTimeStart = OsTime()
	ui.cursorTimeEnd = ui.cursorTimeStart + 5
	ui.cursorCdA = 255
}

func (ui *Ui) Cell() int {
	return ui.io.Cell()
}

func (ui *Ui) RenderTile(text string, coord OsV4, cd OsCd, font *Font) error {
	if ui == nil {
		return nil
	}

	textH := ui.io.GetDPI() / 7

	num_lines := strings.Count(text, "\n") + 1
	cq := coord
	lineH := int(float32(textH) * 1.7)
	cq.Size, _ = font.GetTextSize(text, textH, lineH)
	cq = cq.AddSpaceX((lineH - textH) / -2)

	cq = OsV4_relativeSurround(coord, cq, OsV4{OsV2{}, OsV2{X: ui.io.ini.WinW, Y: ui.io.ini.WinH}})

	err := ui.render.SetClipRect(cq.GetSDLRect())
	if err != nil {
		fmt.Printf("SetClipRect() failed: %v\n", err)
	}
	_Ui_boxSE(ui.render, cq.Start, cq.End(), OsCd_white())
	_Ui_boxSE_border(ui.render, cq.Start, cq.End(), OsCd_black(), 1)
	cq.Size.Y /= num_lines
	err = font.Print(text, textH, cq, OsV2{1, 1}, cd, nil, true, ui.render)
	if err != nil {
		fmt.Printf("Print() failed: %v\n", err)
	}

	return err
}

func (ui *Ui) RenderInfoStats(ui_info *Info, vm_info *Info, font *Font) error {
	if ui == nil {
		return nil
	}

	textH := ui.io.GetDPI() / 6

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	text := fmt.Sprintf("worst FPS(ui: %.1f, vm: %.1f), avg FPS(ui: %.1f, vm: %.1f), Memory(imgs(%dx): %.2fMB, process: %.2fMB), Threads(%d)",
		ui_info.out_worst_fps, vm_info.out_worst_fps,
		ui_info.out_avg_fps, vm_info.out_avg_fps,
		ui.NumTextures(), float64(ui.GetImagesBytes())/1024/1024, float64(mem.Sys)/1024/1024,
		runtime.NumGoroutine())
	//netStats.num_connections_opened-netStats.num_connections_closed, netStats.num_sends, netStats.num_recvs)	//, Net(connections: %d, send: %dx, recv: %dx)

	sz, _ := font.GetTextSize(text, textH, int(float32(textH)*1.2))

	cq := OsV4{ui.io.GetCoord().Middle().Sub(sz.MulV(0.5)), sz}

	err := ui.render.SetClipRect(cq.GetSDLRect())
	if err != nil {
		fmt.Printf("SetClipRect() failed: %v\n", err)
	}
	_Ui_boxSE(ui.render, cq.Start, cq.End(), OsCd_white())
	err = font.Print(text, textH, cq, OsV2{0, 1}, OsCd{255, 50, 50, 255}, nil, true, ui.render)
	if err != nil {
		fmt.Printf("Print() failed: %v\n", err)
	}

	return nil
}
