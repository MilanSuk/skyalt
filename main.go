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
	"image/color"
	"time"
)

func main() {

	InitImageGlobal()

	//SDL
	err := InitSDLGlobal()
	if err != nil {
		fmt.Printf("InitSDLGlobal() failed: %v\n", err)
		return
	}
	defer DestroySDLGlobal()

	//Window(GL)
	win, err := NewWin()
	if err != nil {
		fmt.Printf("NewWin() failed: %v\n", err)
		return
	}
	defer win.Destroy()

	clients, err := NewUiClients(win, 10000)
	if err != nil {
		fmt.Printf("NewUIs() failed: %v\n", err)
		return
	}
	defer clients.Destroy()

	//Main loop
	pl := clients.GetPalette()
	particles_cd := Color_Aprox(pl.P, pl.B, 0.5)
	num_sleeps := 0
	run := true

	Agent_initGlobal()
	defer Agent_destroyGlobal()

	for run {
		//window
		var win_redraw bool
		var err error
		run, win_redraw, err = win.UpdateIO()
		if err != nil {
			fmt.Printf("UpdateIO() failed: %v\n", err)
			return
		}

		if win.RenderProgressWelcome() {
			win.StartRender(color.RGBA{220, 220, 220, 255})
			win.RenderProgress(particles_cd, 0)
			win.EndRender(true, clients.GetEnv().Stats)

		} else {
			clients.UpdateIO()

			if !clients.ui.winRect.Cmp(win.GetScreenCoord()) {
				clients.ui.SetRefresh()
				clients.ui.winRect = win.GetScreenCoord() //update
			}

			//tooltip delay
			if !win_redraw && clients.NeedRedraw() {
				win_redraw = true
			}

			clients.Tick()

			if win_redraw {
				win.StartRender(color.RGBA{220, 220, 220, 255})
				clients.Draw()
				win.EndRender(true, clients.GetEnv().Stats)

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
