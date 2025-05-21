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
	"image/color"
	"log"
	"time"
)

func main() {
	log.SetFlags(log.Llongfile) //log.LstdFlags | log.Lshortfile

	InitImageGlobal()

	//SDL
	err := InitSDLGlobal()
	if err != nil {
		log.Fatalf("InitSDLGlobal() failed: %v\n", err)
	}
	defer DestroySDLGlobal()

	//Window(GL)
	win, err := NewWin()
	if err != nil {
		log.Fatalf("NewWin() failed: %v\n", err)
	}
	defer win.Destroy()

	//Tools
	router := NewToolsRouter("tools", "files", 8000)
	defer router.Destroy()

	//UI
	ui, err := NewUi(win, router)
	if err != nil {
		log.Fatalf("NewUIs() failed: %v\n", err)
	}
	defer ui.Destroy()

	//Loop
	num_sleeps := 0
	run := true
	for run {
		var win_redraw bool
		var err error
		run, win_redraw, err = win.UpdateIO()
		if err != nil {
			log.Fatalf("UpdateIO() failed: %v\n", err)
		}

		if !router.tools.IsRunning() {

			win.StartRender(color.RGBA{220, 220, 220, 255})

			if router.tools.Compile_error != "" {
				//error
				win.RenderError("Error: " + router.tools.Compile_error)

			} else {
				msg := router.FindRecompileMsg()
				if msg != nil && !msg.out_done.Load() {
					//compiling
					pl := ui.sync.GetPalette()
					particles_cd := Color_Aprox(pl.P, pl.B, 0.5)
					win.RenderProgress(msg.progress_label, particles_cd, 0, ui.Cell())
				} else {
					//try run it
					router.tools.CheckRun()
				}

			}
			win.EndRender(true, ui.sync.GetStats())

		} else {
			win.StopProgress()

			ui.UpdateIO(win.GetScreenCoord())

			ui.Tick()

			//tooltip delay
			if !win_redraw && ui.NeedRedraw() {
				win_redraw = true
			}

			if win_redraw {
				win.StartRender(color.RGBA{220, 220, 220, 255})
				ui.Draw()
				win.EndRender(true, ui.sync.GetStats())

				num_sleeps = 0 //reset
			} else {
				if num_sleeps > 60 {
					time.Sleep((1000 / 5) * time.Millisecond) //deep sleep
				} else {
					time.Sleep((1000 / 60) * time.Millisecond) //light sleep
				}

				num_sleeps++
			}
		}
		win.Finish()
	}
}
