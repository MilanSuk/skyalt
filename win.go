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
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"runtime"
	"strings"
	"time"
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
)

const SKYALT_INI_PATH = "ini.json"

func InitSDLGlobal() error {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		return fmt.Errorf("sdl.Init() failed: %w", err)
	}

	err = mix.Init(mix.INIT_FLAC | mix.INIT_MOD | mix.INIT_MP3 | mix.INIT_OGG)
	if err != nil {
		return fmt.Errorf("mix.Init() failed: %w", err)
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
	sdl.Quit()
}

type Win struct {
	io *WinIO

	window *sdl.Window
	render sdl.GLContext

	buff *WinPaintBuff

	winVisible  bool
	num_redraws int //bool ...

	lastClickUp OsV2
	numClicks   int

	fullscreen_now          bool
	fullscreen              bool
	recover_fullscreen_size OsV2

	cursors  []WinCursor
	cursorId int

	//blinking cursor
	cursorEdit          bool
	cursorTimeStart     float64
	cursorTimeEnd       float64
	cursorTimeLastBlink float64
	cursorCdA           byte

	images []*WinImage

	gph *WinGph

	particles *WinParticles

	stat       WinStats
	start_time int64

	services *WinServices

	quit bool

	anim_start_time float64
}

// disk can be nil
func NewWin() (*Win, error) {
	win := &Win{}

	win.buff = NewWinPaintBuff(win)

	if !OsFileExists(SKYALT_INI_PATH) {
		win.anim_start_time = OsTime()
	}

	var err error
	win.io, err = NewWinIO()
	if err != nil {
		return nil, fmt.Errorf("NewIO() failed: %w", err)
	}
	err = win.io.Open(SKYALT_INI_PATH)
	if err != nil {
		return nil, fmt.Errorf("Open() failed: %w", err)
	}

	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "2")

	// create SDL
	win.window, err = sdl.CreateWindow("SkyAlt", int32(win.io.Ini.WinX), int32(win.io.Ini.WinY), int32(win.io.Ini.WinW), int32(win.io.Ini.WinH), sdl.WINDOW_RESIZABLE|sdl.WINDOW_OPENGL)
	if err != nil {
		return nil, fmt.Errorf("CreateWindow() failed: %w", err)
	}

	win.render, err = win.window.GLCreateContext()
	if err != nil {
		return nil, fmt.Errorf("GLCreateContext() failed: %w", err)
	}
	err = gl.Init()
	if err != nil {
		return nil, fmt.Errorf("gl.Init() failed: %w", err)
	}

	win.gph = NewWinGph()

	sdl.EventState(sdl.DROPFILE, sdl.ENABLE)
	sdl.StartTextInput()

	// cursors
	win.cursors = append(win.cursors, WinCursor{"default", sdl.SYSTEM_CURSOR_ARROW, sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_ARROW)})
	win.cursors = append(win.cursors, WinCursor{"hand", sdl.SYSTEM_CURSOR_HAND, sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_HAND)})
	win.cursors = append(win.cursors, WinCursor{"ibeam", sdl.SYSTEM_CURSOR_IBEAM, sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_IBEAM)})
	win.cursors = append(win.cursors, WinCursor{"cross", sdl.SYSTEM_CURSOR_CROSSHAIR, sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_CROSSHAIR)})

	win.cursors = append(win.cursors, WinCursor{"res_col", sdl.SYSTEM_CURSOR_SIZEWE, sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_SIZEWE)})
	win.cursors = append(win.cursors, WinCursor{"res_row", sdl.SYSTEM_CURSOR_SIZENS, sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_SIZENS)})
	win.cursors = append(win.cursors, WinCursor{"res_nwse", sdl.SYSTEM_CURSOR_SIZENESW, sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_SIZENESW)}) // bug(already fixed) in SDL: https://github.com/libsdl-org/SDL/issues/2123
	win.cursors = append(win.cursors, WinCursor{"res_nesw", sdl.SYSTEM_CURSOR_SIZENWSE, sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_SIZENWSE)})
	win.cursors = append(win.cursors, WinCursor{"move", sdl.SYSTEM_CURSOR_SIZEALL, sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_SIZEALL)})

	win.cursors = append(win.cursors, WinCursor{"wait", sdl.SYSTEM_CURSOR_WAITARROW, sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_WAITARROW)})
	win.cursors = append(win.cursors, WinCursor{"no", sdl.SYSTEM_CURSOR_NO, sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_NO)})

	win.services = NewServices(&win.io.Ini)

	return win, nil
}

func (win *Win) Destroy() error {
	var err error

	win.io.Save(SKYALT_INI_PATH)

	if win.particles != nil {
		win.particles.Destroy()
	}

	err = win.io.Destroy()
	if err != nil {
		return fmt.Errorf("IO.Destroy() failed: %w", err)
	}

	for _, cur := range win.cursors {
		sdl.FreeCursor(cur.cursor)
	}

	win.gph.Destroy()

	sdl.GLDeleteContext(win.render)

	err = win.window.Destroy()
	if err != nil {
		return fmt.Errorf("Window.Destroy() failed: %w", err)
	}

	win.services.Destroy()

	return nil
}

func (win *Win) StopProgress() {
	if win.particles != nil {
		win.particles.Destroy()
		win.particles = nil
	}
}
func (win *Win) SetProgress(done float64, showProc bool, description string, cell int) {
	if win.particles == nil {
		var err error
		win.particles, err = NewWinParticles(win)
		if err != nil {
			fmt.Printf("NewParticles() failed: %v\n", err)
			return
		}
	}
	win.particles.done = done
	win.particles.showProc = showProc
	win.particles.description = description
	win.particles.cell = cell
}

func IsSpaceActive() bool {
	state := sdl.GetKeyboardState()
	return state[sdl.SCANCODE_SPACE] != 0
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

func (win *Win) GetMousePosition() (OsV2, bool, bool) {
	x, y, state := sdl.GetGlobalMouseState()

	wx, wy := win.window.GetPosition()
	ww, wh := win.window.GetSize()
	return OsV2_32(x, y).Sub(OsV2_32(wx, wy)), (state != 0 && state != sdl.ButtonLMask()), InitOsV4(int(wx), int(wy), int(ww), int(wh)).Inside(OsV2_32(x, y))
}

func (win *Win) GetOutputSize() (int, int) {
	w, h := win.window.GLGetDrawableSize()
	return int(w), int(h)
}

func (win *Win) GetScreenCoord() OsV4 {
	w, h := win.GetOutputSize()
	return OsV4{Start: OsV2{}, Size: OsV2{w, h}}
}

func (win *Win) GetScreenshot(coord OsV4) (*image.RGBA, error) {

	surface, err := sdl.CreateRGBSurface(0, int32(coord.Size.X), int32(coord.Size.Y), 32, 0, 0, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("CreateRGBSurface() failed: %w", err)
	}
	defer surface.Free()

	//copies pixels
	winH := win.GetScreenCoord().Size.Y
	gl.ReadPixels(int32(coord.Start.X), int32(winH-(coord.Start.Y+coord.Size.Y)), int32(coord.Size.X), int32(coord.Size.Y), gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&surface.Pixels()[0])) //int(surface.Pitch) ...

	img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{int(surface.W), int(surface.H)}})
	for y := int32(0); y < surface.H; y++ {
		for x := int32(0); x < surface.W; x++ {
			r := surface.Pixels()[y*surface.W*4+x*4+0]
			g := surface.Pixels()[y*surface.W*4+x*4+1]
			b := surface.Pixels()[y*surface.W*4+x*4+2]
			img.SetRGBA(int(x), int(surface.H-1-y), color.RGBA{r, g, b, 255})
		}
	}
	return img, nil
}

func (win *Win) SaveScreenshot() error {
	img, err := win.GetScreenshot(win.GetScreenCoord())
	if err != nil {
		return err
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

func (win *Win) NumTextures() int {
	n := 0
	for _, it := range win.images {
		if it.texture != nil {
			n++
		}
	}
	return n
}

func (win *Win) GetImagesBytes() int {
	n := 0
	for _, it := range win.images {
		n += it.GetBytes()
	}
	return n
}

func (win *Win) FindImage(path WinMedia) *WinImage {
	for _, it := range win.images {
		if it.path.Cmp(&path) {
			return it
		}
	}
	return nil
}

func (win *Win) AddImage(path WinMedia) (*WinImage, error) {

	img := win.FindImage(path)

	if img == nil {
		var err error
		img, err = NewWinImage(path, win.services)
		if err != nil {
			return nil, fmt.Errorf("NewImage() failed: %w", err)
		}

		if img != nil {
			win.images = append(win.images, img)
		}
	}

	return img, nil
}

var g_dropPath string //dirty trick, because, when drop, the mouse position is invalid

func (win *Win) Event() (bool, bool, error) {
	io := win.io
	inputChanged := false

	//sdl.WaitEvent()
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() { // some cases have RETURN(don't miss it in tick), some (can be missed in tick)!

		switch val := event.(type) {
		case *sdl.QuitEvent:
			fmt.Println("Exiting ...")
			return false, false, nil

		case *sdl.WindowEvent:
			switch val.Event {
			case sdl.WINDOWEVENT_SIZE_CHANGED:
				inputChanged = true
			case sdl.WINDOWEVENT_MOVED:
				inputChanged = true
			case sdl.WINDOWEVENT_SHOWN:
				win.winVisible = true
				inputChanged = true
			case sdl.WINDOWEVENT_HIDDEN:
				win.winVisible = false
				inputChanged = true
			}

		case *sdl.MouseMotionEvent:
			inputChanged = true

		case *sdl.MouseButtonEvent:
			win.numClicks = int(val.Clicks)
			if val.Clicks > 1 {
				if win.lastClickUp.Distance(OsV2_32(val.X, val.Y)) > float32(GetDeviceDPI())/13 { //7px error space
					win.numClicks = 1
				}
			}

			io.Touch.Pos = OsV2_32(val.X, val.Y)
			io.Touch.Rm = (val.Button != sdl.BUTTON_LEFT)

			if val.Type == sdl.MOUSEBUTTONDOWN {
				io.Touch.Start = true
				sdl.CaptureMouse(true) // keep getting info even mouse is outside window

			} else if val.Type == sdl.MOUSEBUTTONUP {

				win.lastClickUp = io.Touch.Pos
				io.Touch.End = true
				sdl.CaptureMouse(false)
			}
			return true, true, nil

		case *sdl.MouseWheelEvent:
			if IsCtrlActive() { // zoom

				if val.Y > 0 {
					io.Keys.ZoomAdd = true
				}
				if val.Y < 0 {
					io.Keys.ZoomSub = true
				}
			} else {
				io.Touch.Wheel = -int(val.Y) // divide by -WHEEL_DELTA
				io.Touch.wheel_last_sec = OsTime()
			}
			return true, true, nil

		case *sdl.KeyboardEvent:
			if val.Type == sdl.KEYDOWN {

				if IsCtrlActive() {
					if val.Keysym.Sym == sdl.K_PLUS || val.Keysym.Sym == sdl.K_KP_PLUS {
						io.Keys.ZoomAdd = true
					}
					if val.Keysym.Sym == sdl.K_MINUS || val.Keysym.Sym == sdl.K_KP_MINUS {
						io.Keys.ZoomSub = true
					}
					if val.Keysym.Sym == sdl.K_0 || val.Keysym.Sym == sdl.K_KP_0 {
						io.Keys.ZoomDef = true
					}
				}

				keys := &io.Keys

				keys.Esc = val.Keysym.Sym == sdl.K_ESCAPE
				keys.Enter = (val.Keysym.Sym == sdl.K_RETURN || val.Keysym.Sym == sdl.K_RETURN2 || val.Keysym.Sym == sdl.K_KP_ENTER)

				keys.ArrowU = val.Keysym.Sym == sdl.K_UP
				keys.ArrowD = val.Keysym.Sym == sdl.K_DOWN
				keys.ArrowL = val.Keysym.Sym == sdl.K_LEFT
				keys.ArrowR = val.Keysym.Sym == sdl.K_RIGHT
				keys.Home = val.Keysym.Sym == sdl.K_HOME
				keys.End = val.Keysym.Sym == sdl.K_END
				keys.PageU = val.Keysym.Sym == sdl.K_PAGEUP
				keys.PageD = val.Keysym.Sym == sdl.K_PAGEDOWN

				keys.Copy = val.Keysym.Sym == sdl.K_COPY || (IsCtrlActive() && val.Keysym.Sym == sdl.K_c)
				keys.Cut = val.Keysym.Sym == sdl.K_CUT || (IsCtrlActive() && val.Keysym.Sym == sdl.K_x)
				keys.Paste = val.Keysym.Sym == sdl.K_PASTE || (IsCtrlActive() && val.Keysym.Sym == sdl.K_v)
				keys.SelectAll = val.Keysym.Sym == sdl.K_SELECT || (IsCtrlActive() && val.Keysym.Sym == sdl.K_a)
				keys.Backward = val.Keysym.Sym == sdl.K_AC_FORWARD || (IsCtrlActive() && !IsShiftActive() && val.Keysym.Sym == sdl.K_z)
				keys.Forward = val.Keysym.Sym == sdl.K_AC_BACK || (IsCtrlActive() && val.Keysym.Sym == sdl.K_y) || (IsCtrlActive() && IsShiftActive() && val.Keysym.Sym == sdl.K_z)

				keys.Tab = val.Keysym.Sym == sdl.K_TAB
				keys.Space = val.Keysym.Sym == sdl.K_SPACE

				keys.Delete = val.Keysym.Sym == sdl.K_DELETE
				keys.Backspace = val.Keysym.Sym == sdl.K_BACKSPACE

				keys.F1 = val.Keysym.Sym == sdl.K_F1
				keys.F2 = val.Keysym.Sym == sdl.K_F2
				keys.F3 = val.Keysym.Sym == sdl.K_F3
				keys.F4 = val.Keysym.Sym == sdl.K_F4
				keys.F5 = val.Keysym.Sym == sdl.K_F5
				keys.F6 = val.Keysym.Sym == sdl.K_F6
				keys.F7 = val.Keysym.Sym == sdl.K_F7
				keys.F8 = val.Keysym.Sym == sdl.K_F8
				keys.F9 = val.Keysym.Sym == sdl.K_F9
				keys.F10 = val.Keysym.Sym == sdl.K_F10
				keys.F11 = val.Keysym.Sym == sdl.K_F11
				keys.F12 = val.Keysym.Sym == sdl.K_F12

				let := val.Keysym.Sym
				if OsIsTextWord(rune(let)) || let == ' ' {
					if IsCtrlActive() {
						keys.CtrlChar = string(let) //string([]byte{byte(let)})
					}
					if IsAltActive() {
						keys.AltChar = string(let)
					}
				}

				keys.HasChanged = true
			}
			return true, true, nil

		case *sdl.TextInputEvent:
			if !(IsCtrlActive() && len(val.Text) > 0 && val.Text[0] == ' ') { // ignore ctrl+space
				io.Keys.Text += val.GetText()
				io.Keys.HasChanged = true
			}
			return true, true, nil

		case *sdl.DropEvent:
			if val.Type == sdl.DROPFILE {
				g_dropPath = val.File
			}
			return true, true, nil
		}
	}

	return true, inputChanged, nil
}

func (win *Win) Maintenance() {
	for i := len(win.images) - 1; i >= 0; i-- {
		ok, _ := win.images[i].Maintenance(win)
		if !ok {
			win.images = append(win.images[:i], win.images[i+1:]...)
		}
	}

	win.gph.Maintenance()
}

func (win *Win) needRedraw(inputChanged bool) bool {
	if win == nil {
		return true
	}

	if win.cursorEdit {
		if inputChanged {
			win.cursorEdit = false
		}

		tm := OsTime()

		if inputChanged {
			win.cursorTimeEnd = tm + 5 //startAfterSleep/continue blinking after mouse move
		}

		if (tm - win.cursorTimeStart) < 0.5 {
			//star
			win.cursorCdA = 255
		} else if tm > win.cursorTimeEnd {
			//sleep
			if win.cursorCdA == 0 { //last draw must be full
				win.cursorCdA = 255
				inputChanged = true //redraw
			}
		} else if (tm - win.cursorTimeLastBlink) > 0.5 {
			//blink swap
			if win.cursorCdA > 0 {
				win.cursorCdA = 0
			} else {
				win.cursorCdA = 255
			}
			inputChanged = true //redraw
			win.cursorTimeLastBlink = tm
		}
	}

	return inputChanged
}

func (win *Win) SetRedraw() {
	win.num_redraws = 0
}

func (win *Win) UpdateIO() (bool, bool, error) {
	if win == nil {
		return true, false, nil
	}

	run, redraw, err := win.Event()
	if err != nil {
		return run, true, fmt.Errorf("Event() failed: %w", err)
	}
	if !run {
		return false, redraw, nil
	}

	if win.quit {
		return false, redraw, nil
	}

	if win.needRedraw(redraw) {
		win.SetRedraw()
		redraw = true
	}

	//one more time
	if win.num_redraws == 0 {
		redraw = true
	}

	//draw one more time after click start/end
	if !win.io.Touch.Start && !win.io.Touch.End && !win.io.Touch.Rm {
		win.num_redraws++
	}

	// update Win
	io := win.io

	{
		start := OsV2_32(win.window.GetPosition())
		size := OsV2_32(win.window.GetSize())
		io.Ini.WinX = start.X
		io.Ini.WinY = start.Y
		io.Ini.WinW = size.X
		io.Ini.WinH = size.Y
	}

	//io.SetDeviceDPI()

	if !io.Touch.Start && !io.Touch.End && !io.Touch.Rm {
		var inside bool
		io.Touch.Pos, io.Touch.Rm, inside = win.GetMousePosition()

		//drop file
		if inside && g_dropPath != "" {
			win.io.Touch.Drop_path = g_dropPath
			g_dropPath = ""
			win.SetRedraw()
		}
	}
	io.Touch.NumClicks = win.numClicks
	if io.Touch.End {
		win.numClicks = 0
	}

	io.Keys.Shift = IsShiftActive()
	io.Keys.Alt = IsAltActive()
	io.Keys.Ctrl = IsCtrlActive()
	io.Keys.Space = IsSpaceActive()

	if io.Keys.F8 {
		err := win.SaveScreenshot()
		if err != nil {
			return true, true, fmt.Errorf("SaveScreenshot() failed: %w", err)
		}
	}

	win.cursorId = 0

	win.services.Maintenance()

	return true, redraw, nil
}

func (win *Win) StartRender(clearCd color.RGBA) error {
	if win == nil {
		return nil
	}

	//GL setup
	{
		screen := win.GetScreenCoord()

		gl.Enable(gl.SCISSOR_TEST)

		gl.Enable(gl.DEPTH_TEST)
		gl.ClearColor(float32(clearCd.R)/255, float32(clearCd.G)/255, float32(clearCd.B)/255, float32(clearCd.A)/255)
		gl.ClearDepth(1)
		gl.DepthFunc(gl.LEQUAL)
		gl.Viewport(0, 0, int32(screen.Size.X), int32(screen.Size.Y))

		gl.MatrixMode(gl.PROJECTION)
		gl.LoadIdentity()
		gl.Ortho(0, float64(screen.Size.X), float64(screen.Size.Y), 0, -1000, 1000)
		gl.MatrixMode(gl.MODELVIEW)
		gl.LoadIdentity()

		gl.Enable(gl.POINT_SMOOTH)
		gl.Hint(gl.POINT_SMOOTH_HINT, gl.NICEST)
		gl.Enable(gl.LINE_SMOOTH)
		gl.Hint(gl.LINE_SMOOTH_HINT, gl.NICEST)
		//gl.Enable(gl.POLYGON_SMOOTH)
		//gl.Hint(gl.POLYGON_SMOOTH_HINT, gl.NICEST)

		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		gl.Enable(gl.ALPHA_TEST)
		gl.ShadeModel(gl.SMOOTH)

		gl.Enable(gl.TEXTURE_2D)
	}

	w, h := win.GetOutputSize()
	gl.Scissor(0, 0, int32(w), int32(h))

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	win.start_time = OsTicks()
	return nil
}

func (win *Win) RenderProgressWelcome() bool {
	if win.anim_start_time > 0 {
		done := (OsTime() - win.anim_start_time) / 7
		win.SetProgress(done, false, "", 0)
		if win.io.Touch.Start {
			//reset
			win.StopProgress()
			win.anim_start_time = 0
		}

		return true
	}
	return false
}

func (win *Win) RenderProgress(cd_orig color.RGBA, depth int) {

	if win.particles != nil {
		if !win.particles.Tick(cd_orig, depth, win) {
			//reset(0 visible particles)
			win.StopProgress()
			win.anim_start_time = 0
		}
		win.SetRedraw()
	}
}
func (win *Win) EndRender(present bool, show_stats bool) error {
	if win == nil {
		return nil
	}

	win.stat.Update(int(OsTicks() - win.start_time))
	if show_stats {
		win.renderStats()
	}
	//fmt.Println("Num Texts", len(win.gph.texts))

	if present {
		win.window.GLSwap()

		if win.cursorId >= 0 {
			if win.cursorId >= len(win.cursors) {
				return errors.New("cursorID is out of range")
			}
			sdl.SetCursor(win.cursors[win.cursorId].cursor)
		}
	}

	if win.fullscreen != win.fullscreen_now {
		fullFlag := uint32(0)
		if win.fullscreen {
			win.recover_fullscreen_size = OsV2_32(win.window.GetSize())
			fullFlag = uint32(sdl.WINDOW_FULLSCREEN_DESKTOP)
		}
		err := win.window.SetFullscreen(fullFlag)
		if err != nil {
			return fmt.Errorf("SetFullscreen() failed: %w", err)
		}
		if fullFlag == 0 {
			win.window.SetSize(win.recover_fullscreen_size.Get32()) //otherwise, wierd bug happens
		}

		win.fullscreen_now = win.fullscreen
	}

	return nil
}

func (win *Win) Finish() {
	win.io.ResetTouchAndKeys()

	win.Maintenance()
}

func (win *Win) SetClipboardText(text string) {
	sdl.SetClipboardText(text)
}
func (win *Win) GetClipboardText() string {
	text, err := sdl.GetClipboardText()
	if err != nil {
		fmt.Println("Error: UpdateIO.GetClipboardText() failed: %w", err)
	}
	return strings.Trim(text, "\r")
}

func (win *Win) renderStats() error {

	cell := int(float64(GetDeviceDPI()) / 2.5)
	props := InitWinFontPropsDef(cell)

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	text := fmt.Sprintf("FPS(worst: %.1f, best: %.1f, avg: %.1f), Memory(%d imgs: %.2fMB, process: %.2fMB), Threads(%d), Text(live: %d, created: %d, removed: %d)",
		win.stat.out_worst_fps, win.stat.out_best_fps, win.stat.out_avg_fps,
		win.NumTextures(), float64(win.GetImagesBytes())/1024/1024, float64(mem.Sys)/1024/1024, runtime.NumGoroutine(),
		len(win.gph.texts), win.gph.texts_num_created, win.gph.texts_num_remove)

	sz := win.GetTextSize(-1, text, props)

	cq := OsV4{win.GetScreenCoord().Middle().Sub(sz.MulV(0.5)), sz}

	win.SetClipRect(cq)
	depth := 990 //...
	win.DrawRect(cq.Start, cq.End(), depth, color.RGBA{255, 255, 255, 255})

	win.DrawText(text, props, color.RGBA{255, 50, 50, 255}, cq, depth, OsV2{0, 1}, 0, 1)

	return nil
}

func (win *Win) SetClipRect(coord OsV4) {
	if win == nil {
		return
	}

	_, winH := win.GetOutputSize()
	gl.Scissor(int32(coord.Start.X), int32(winH-(coord.Start.Y+coord.Size.Y)), int32(coord.Size.X), int32(coord.Size.Y))
}

func (win *Win) PaintCursor(name string) error {
	if win == nil {
		return nil
	}

	for i, cur := range win.cursors {
		if strings.EqualFold(cur.name, name) {
			win.cursorId = i
			return nil
		}
	}

	return errors.New("Cursor(" + name + ") not found: ")
}

func (win *Win) DrawPointStart() {
	gl.Begin(gl.POINTS)
}
func (win *Win) DrawPointCdI(pos OsV2, depth int, cd color.RGBA) {
	gl.Color4ub(cd.R, cd.G, cd.B, cd.A)
	gl.Vertex3i(int32(pos.X), int32(pos.Y), int32(depth))
}
func (win *Win) DrawPointCdF(pos OsV2f, depth int, cd color.RGBA) {
	gl.Color4ub(cd.R, cd.G, cd.B, cd.A)
	gl.Vertex3f(float32(pos.X), float32(pos.Y), float32(depth))
}

func (win *Win) DrawPointEnd() {
	gl.End()
}

func (win *Win) DrawRect(start OsV2, end OsV2, depth int, cd color.RGBA) {
	if start.X != end.X && start.Y != end.Y {
		gl.Color4ub(cd.R, cd.G, cd.B, cd.A)

		gl.Begin(gl.QUADS)
		{
			gl.Vertex3f(float32(start.X), float32(start.Y), float32(depth))
			gl.Vertex3f(float32(end.X), float32(start.Y), float32(depth))
			gl.Vertex3f(float32(end.X), float32(end.Y), float32(depth))
			gl.Vertex3f(float32(start.X), float32(end.Y), float32(depth))
		}
		gl.End()

		//gl.Enable(gl.POLYGON_SMOOTH)
	}
}

func (win *Win) DrawRect_border(start OsV2, end OsV2, depth int, cd color.RGBA, thick int) {
	win.DrawRect(start, OsV2{end.X, start.Y + thick}, depth, cd) // top
	win.DrawRect(OsV2{start.X, end.Y - thick}, end, depth, cd)   // bottom
	win.DrawRect(start, OsV2{start.X + thick, end.Y}, depth, cd) // left
	win.DrawRect(OsV2{end.X - thick, start.Y}, end, depth, cd)   // right
}

func (win *Win) DrawRectRound(coord OsV4, rad int, depth int, cd color.RGBA, thick int, grad bool) {

	rad = OsMin(rad, coord.Size.X/2)
	rad = OsMin(rad, coord.Size.Y/2)

	gl.Color4ub(cd.R, cd.G, cd.B, cd.A)

	if thick > 0 {
		s := coord.Start
		e := coord.End()

		gl.LineWidth(float32(thick))
		gl.Begin(gl.LINES)

		//top
		gl.Vertex3f(float32(s.X+rad), float32(s.Y), float32(depth))
		gl.Vertex3f(float32(e.X-rad), float32(s.Y), float32(depth))
		//bottom
		gl.Vertex3f(float32(s.X+rad), float32(e.Y), float32(depth))
		gl.Vertex3f(float32(e.X-rad), float32(e.Y), float32(depth))
		//left
		gl.Vertex3f(float32(s.X), float32(s.Y+rad), float32(depth))
		gl.Vertex3f(float32(s.X), float32(e.Y-rad), float32(depth))
		//right
		gl.Vertex3f(float32(e.X), float32(s.Y+rad), float32(depth))
		gl.Vertex3f(float32(e.X), float32(e.Y-rad), float32(depth))

		gl.End()
	} else {
		gl.Begin(gl.QUADS)

		if !grad {
			//top
			_WinTexture_drawQuadNoUVs(InitOsV4(coord.Start.X+rad, coord.Start.Y, coord.Size.X-2*rad, rad), depth)
			//middle
			_WinTexture_drawQuadNoUVs(InitOsV4(coord.Start.X, coord.Start.Y+rad, coord.Size.X, coord.Size.Y-2*rad), depth)
			//bottom
			_WinTexture_drawQuadNoUVs(InitOsV4(coord.Start.X+rad, coord.Start.Y+coord.Size.Y-rad, coord.Size.X-2*rad, rad), depth)
		} else {
			//top
			_WinTexture_drawQuadNoUVs_cd(InitOsV4(coord.Start.X+rad, coord.Start.Y, coord.Size.X-2*rad, rad), depth, cd, [4]byte{0, 0, cd.A, cd.A})
			//bottom
			_WinTexture_drawQuadNoUVs_cd(InitOsV4(coord.Start.X+rad, coord.Start.Y+coord.Size.Y-rad, coord.Size.X-2*rad, rad), depth, cd, [4]byte{cd.A, cd.A, 0, 0})
			//left
			_WinTexture_drawQuadNoUVs_cd(InitOsV4(coord.Start.X, coord.Start.Y+rad, rad, coord.Size.Y-2*rad), depth, cd, [4]byte{0, cd.A, cd.A, 0})
			//right
			_WinTexture_drawQuadNoUVs_cd(InitOsV4(coord.Start.X+coord.Size.X-rad, coord.Start.Y+rad, rad, coord.Size.Y-2*rad), depth, cd, [4]byte{cd.A, 0, 0, cd.A})
			//middle
			gl.Color4ub(cd.R, cd.G, cd.B, cd.A)
			_WinTexture_drawQuadNoUVs(InitOsV4(coord.Start.X+rad, coord.Start.Y+rad, coord.Size.X-2*rad, coord.Size.Y-2*rad), depth)
		}

		gl.End()
	}

	if rad > 0 {
		rad += thick //inside GetCircle(), radius is smaller by thick

		coord = coord.Crop(-thick) //make it larger, because 'rad += thick'
		s := coord.Start
		e := coord.End()

		var circle *WinGphItemCircle
		if !grad {
			circle = win.gph.GetCircle(OsV2{rad * 2, rad * 2}, float64(thick), OsV2f{})
		} else {
			circle = win.gph.GetCircleGrad(OsV2{rad * 2, rad * 2}, OsV2f{}, 0.5)
		}

		circle.item.DrawUV(InitOsV4(s.X, s.Y, rad, rad), depth, cd, OsV2f{0, 0}, OsV2f{0.5, 0.5})
		circle.item.DrawUV(InitOsV4(e.X-rad, s.Y, rad, rad), depth, cd, OsV2f{0.5, 0}, OsV2f{1, 0.5})
		circle.item.DrawUV(InitOsV4(s.X, e.Y-rad, rad, rad), depth, cd, OsV2f{0, 0.5}, OsV2f{0.5, 1})
		circle.item.DrawUV(InitOsV4(e.X-rad, e.Y-rad, rad, rad), depth, cd, OsV2f{0.5, 0.5}, OsV2f{1, 1})
	}
}

func (win *Win) DrawCicle(mid OsV2, rad OsV2, depth int, cd color.RGBA, thick int) {

	circle := win.gph.GetCircle(rad.MulV(2), float64(thick), OsV2f{})
	if circle != nil {
		mid.X++
		mid.Y++
		circle.item.DrawCut(InitOsV4Mid(mid, circle.size), depth, cd)
	}
}

func (win *Win) DrawLine(start OsV2, end OsV2, depth int, thick int, cd color.RGBA) {

	gl.Color4ub(cd.R, cd.G, cd.B, cd.A)

	v := end.Sub(start)
	if !v.IsZero() {

		if start.Y == end.Y {
			win.DrawRect(start, OsV2{end.X, start.Y + thick}, depth, cd) // horizontal
		} else if start.X == end.X {
			win.DrawRect(start, OsV2{start.X + thick, end.Y}, depth, cd) // vertical
		} else {
			gl.LineWidth(float32(thick))
			gl.Begin(gl.LINES)
			gl.Vertex3f(float32(start.X), float32(start.Y), float32(depth))
			gl.Vertex3f(float32(end.X), float32(end.Y), float32(depth))
			gl.End()
		}
	}
}

func (win *Win) DrawLines(points []OsV2, depth int, thick int, cd color.RGBA, renderEndings bool) {
	if len(points) == 0 {
		return
	}

	//cd
	gl.Color4ub(cd.R, cd.G, cd.B, cd.A)

	//lines
	/*gl.LineWidth(float32(thick))
	gl.Begin(gl.LINE_STRIP)
	for _, pt := range points {
		gl.Vertex3f(float32(pt.X), float32(pt.Y), float32(depth))
	}
	gl.End()*/

	//gl.Disable(gl.POLYGON_SMOOTH)
	gl.Begin(gl.QUADS)
	var last_pt OsV2f
	for i, pt := range points {
		if i == 0 {
			last_pt = pt.toV2f()
			continue
		}
		s := last_pt
		e := pt.toV2f()
		v := e.Sub(s)
		v.X, v.Y = v.Y, -v.X
		v = v.MulV(float32(thick/2) / v.Len())

		a := s.Sub(v)
		b := s.Add(v)
		c := e.Add(v)
		d := e.Sub(v)

		gl.Vertex3f(a.X, a.Y, float32(depth))
		gl.Vertex3f(b.X, b.Y, float32(depth))
		gl.Vertex3f(c.X, c.Y, float32(depth))
		gl.Vertex3f(d.X, d.Y, float32(depth))

		last_pt = pt.toV2f()
	}
	gl.End()
	//gl.Enable(gl.POLYGON_SMOOTH)

	if renderEndings {
		thick = int(float64(thick) * 0.9)
		circle := win.gph.GetCircle(OsV2{thick, thick}, 0, OsV2f{})
		if circle == nil {
			return
		}
		for _, pt := range points {
			circle.item.DrawCut(InitOsV4Mid(pt, circle.size), depth, cd)
		}
	}
}

func _Win_getBezierPoint(t float64, a, b, c, d OsV2f) OsV2f {
	af := a.MulV(float32(math.Pow(t, 3)))
	bf := b.MulV(float32(3 * math.Pow(t, 2) * (1 - t)))
	cf := c.MulV(float32(3 * t * math.Pow((1-t), 2)))
	df := d.MulV(float32(math.Pow((1 - t), 3)))

	return af.Add(bf).Add(cf).Add(df)
}

func (win *Win) DrawBezier(a OsV2, b OsV2, c OsV2, d OsV2, depth int, thick int, cd color.RGBA, dash_px float32, move float32) {

	gl.Color4ub(cd.R, cd.G, cd.B, cd.A)

	aa := a.toV2f()
	bb := b.toV2f()
	cc := c.toV2f()
	dd := d.toV2f()

	gl.LineWidth(float32(thick))
	if dash_px > 0 {
		gl.Begin(gl.LINES)
	} else {
		gl.Begin(gl.LINE_STRIP)
	}

	{
		//compute length
		len := float32(0)
		last_a := a.toV2f()
		N := 10
		div := 1 / float64(N)
		for t := float64(0); t <= 1.001; t += div {
			len += last_a.Sub(_Win_getBezierPoint(t, aa, bb, cc, dd)).Len()
		}

		N = OsTrn(dash_px > 0, int(len/dash_px), int(len/5)) // 5 = 5px jump
		div = 1 / float64(N)

		pre_p := _Win_getBezierPoint(0-div, aa, bb, cc, dd)
		for t := float64(0); t <= 1.001; t += div {
			p := _Win_getBezierPoint(t, aa, bb, cc, dd)

			if move != 0 {
				old_p := p
				v := p.Sub(pre_p)
				v.X, v.Y = -v.Y, v.X
				v = v.MulV(move / v.Len())
				p = p.Add(v)
				pre_p = old_p
			}

			gl.Vertex3f(p.X, p.Y, float32(depth))

		}
	}
	gl.End()
}

func (win *Win) GetBezier(a OsV2, b OsV2, c OsV2, d OsV2, t float64) (OsV2f, OsV2f) {
	aa := a.toV2f()
	bb := b.toV2f()
	cc := c.toV2f()
	dd := d.toV2f()

	s := _Win_getBezierPoint(t, aa, bb, cc, dd)
	e := _Win_getBezierPoint(t+0.01, aa, bb, cc, dd)

	return s, e.Sub(s)
}

func (win *Win) GetPoly(points []OsV2f, width float64) *WinGphItemPoly {
	return win.gph.GetPoly(points, width)
}

func (win *Win) DrawPolyQuad(pts [4]OsV2f, uvs [4]OsV2f, poly *WinGphItemPoly, depth int, cd color.RGBA) {
	poly.item.DrawPointsUV(pts, uvs, depth, cd)
}

func (win *Win) DrawPolyRect(rect OsV4, poly *WinGphItemPoly, depth int, cd color.RGBA) {
	poly.item.DrawCut(rect, depth, cd)
}
func (win *Win) DrawPolyStart(start OsV2, poly *WinGphItemPoly, depth int, cd color.RGBA) {
	win.DrawPolyRect(OsV4{Start: start, Size: poly.size}, poly, depth, cd)
}

func (win *Win) DrawText(ln string, prop WinFontProps, frontCd color.RGBA, coord OsV4, depth int, align OsV2, yLine, numLines int) { // single line
	item := win.gph.GetText(prop, ln)
	if item != nil {
		start := win.GetTextStart(ln, prop, coord, align.X, align.Y, numLines)
		start.Y += yLine * prop.lineH

		item.item.DrawCutCds(OsV4{Start: start, Size: item.size}, depth, frontCd, item.cds) //InitOsCdWhite())
	}
}

func (win *Win) GetTextSize(cur int, ln string, prop WinFontProps) OsV2 {
	return win.gph.GetTextSize(prop, cur, ln)
}
func (win *Win) GetTextSizeMax(text string, max_line_px int, prop WinFontProps) (int, int) {
	tx := win.gph.GetTextMax(text, max_line_px, prop)
	if tx == nil {
		return 0, 1
	}

	return tx.max_size_x, len(tx.lines)
}
func (win *Win) GetTextLines(text string, max_line_px int, prop WinFontProps) []WinGphLine {
	tx := win.gph.GetTextMax(text, max_line_px, prop)
	if tx == nil {
		return []WinGphLine{{s: 0, e: len(text)}}
	}

	return tx.lines
}
func (win *Win) GetTextNumLines(text string, max_line_px int, prop WinFontProps) int {
	tx := win.gph.GetTextMax(text, max_line_px, prop)
	if tx == nil {
		return 1
	}

	return len(tx.lines)
}

func (win *Win) GetTextPos(touchPx int, ln string, prop WinFontProps, coord OsV4, align OsV2) int {
	start := win.GetTextStart(ln, prop, coord, align.X, align.Y, 1)

	return win.gph.GetTextPos(prop, (touchPx - start.X), ln)
}

func (win *Win) GetTextStartLine(ln string, prop WinFontProps, coord OsV4, align OsV2, numLines int) OsV2 {
	lnSize := win.GetTextSize(-1, ln, prop)
	size := OsV2{lnSize.X, numLines * prop.lineH}
	return coord.Align(size, align)
}

func (win *Win) GetTextStart(ln string, prop WinFontProps, coord OsV4, align_h, align_v int, numLines int) OsV2 {

	//lineH
	lnSize := win.GetTextSize(-1, ln, prop)
	size := OsV2{lnSize.X, numLines * prop.lineH}
	start := coord.Align(size, OsV2{align_h, align_v})

	//letters
	coord.Start = start
	coord.Size.X = size.X
	coord.Size.Y = prop.lineH
	start = coord.Align(lnSize, OsV2{align_h, 1}) //letters must be always in the middle of line

	return start
}

func (win *Win) SetTextCursorMove() {
	win.cursorTimeStart = OsTime()
	win.cursorTimeEnd = win.cursorTimeStart + 5
	win.cursorCdA = 255
}
