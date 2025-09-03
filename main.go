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
)

func main() {
	log.SetFlags(log.Llongfile) //log.LstdFlags | log.Lshortfile

	/*{
		f, err := os.Create("profile.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}*/

	InitImageGlobal()

	//SDL
	err := InitSDLGlobal()
	if err != nil {
		log.Fatalf("InitSDLGlobal() failed: %v\n", err)
	}
	defer DestroySDLGlobal()

	media, err := NewMedia(10000)
	if err != nil {
		log.Fatalf("NewMedia() failed: %v\n", err)
	}
	defer media.Destroy()

	//Services
	services, err := NewServices(media)
	if err != nil {
		log.Fatalf("NewServices() failed: %v\n", err)
	}
	defer services.Destroy()

	//Window(GL)
	win, err := NewWin(services)
	if err != nil {
		log.Fatalf("NewWin() failed: %v\n", err)
	}
	defer win.Destroy()

	//Tools
	router, err := NewAppsRouter(9000, services)
	if err != nil {
		log.Fatalf("NewToolsRouter() failed: %v\n", err)
	}
	defer router.Destroy()

	//UI
	ui, err := NewUi(win, router)
	if err != nil {
		log.Fatalf("NewUIs() failed: %v\n", err)
	}
	defer ui.Destroy()

	//Loop
	run := true
	redraw := true
	for run {

		var err error
		run, redraw, err = win.UpdateIO(redraw || router.GetNumMsgs() > 0)
		if err != nil {
			log.Fatalf("UpdateIO() failed: %v\n", err)
		}

		rootApp := router.GetRootApp()
		if rootApp != nil && !rootApp.Process.IsRunning() {

			win.StartRender(color.RGBA{220, 220, 220, 255})

			if rootApp.Process.Compile.Error != "" {
				//error
				win.RenderError("Error: " + rootApp.Process.Compile.Error)
			} else {
				msg := router.FindLocalRecompileMsg("Root")
				if msg != nil && !msg.out_done.Load() {
					//compiling
					pl := ui.GetPalette()
					particles_cd := Color_Aprox(pl.P, pl.B, 0.5)
					win.RenderProgress(msg.progress_label, particles_cd, 0, ui.Cell())
				} else {
					//try run it
					rootApp.CheckRun()
				}
			}

			win.EndRender(true, ui.router.services.sync.GetStats())

		} else {
			win.StopProgress()

			ui.UpdateIO(win.GetScreenCoord())

			ui.Tick()

			//tooltip delay
			if !redraw && ui.NeedRedraw() {
				redraw = true
			}

			if redraw {
				win.StartRender(color.RGBA{220, 220, 220, 255})
				ui.Draw()
				win.EndRender(true, ui.router.services.sync.GetStats())
			}
		}
		win.Finish()
	}
}
