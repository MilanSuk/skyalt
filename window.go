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

type Window struct {
	server *NetServer
	client *NerServerClient

	layout Layout
}

func NewWindow() (*Window, error) {
	win := &Window{}
	win.server = NewNetServer(10000)

	return win, nil
}

func (win *Window) Destroy() {
	if win.client != nil {
		win.client.WriteInt(NMSG_EXIT)
	}

	win.server.Destroy()

	//time.Sleep(5 * time.Second)
}

func (win *Window) RunSDK() {

	go main_sdk(win.server.port)

	var err error
	win.client, err = win.server.Accept()
	if err != nil {
		log.Fatal(err)
	}

	if win.client == nil {
		log.Fatal(fmt.Errorf("client == nil"))
	}
}

var hasStarted = false
var hasFinished = false

func (layout *Layout) FakeRelayout(items map[uint64]Rect) {

	layout.Canvas = Rect{W: 1, H: 1}

	items[layout.Hash] = layout.Canvas

	for _, it := range layout.Childs {
		it.FakeRelayout(items)
	}
}

func (win *Window) Tick() {

	//build
	{
		if !hasStarted {
			win.client.WriteInt(NMSG_REFRESH_START)
			//win.client.WriteInt(1) //update_id

			hasStarted = true
		} else if !hasFinished {

			win.client.WriteInt(NMSG_REFRESH_UPDATE)

			//project Layout3 to Layout  ...
			items := make(map[uint64]Rect)
			win.layout.FakeRelayout(items)

			err := win.client.WriteArray(OsMarshal(items))
			if err != nil {
				log.Fatal(err)
			}

			{
				buffs := make(map[uint64][]LayoutDrawPrim)
				data, err := win.client.ReadArray()
				if err != nil {
					log.Fatal(err)
				}
				OsUnmarshal(data, &buffs)

				if len(buffs) > 0 {
					fmt.Println("buff back", len(buffs))
					//project buffs ...
				}
			}

			{
				var layout Layout
				data, err := win.client.ReadArray()
				if err != nil {
					log.Fatal(err)
				}
				OsUnmarshal(data, &layout)
				win.layout = layout

				//project to Layout3 ...
				//Relayout() ...
			}

			{
				done, err := win.client.ReadInt()
				if err != nil {
					log.Fatal(err)
				}
				if done == 1 {
					hasFinished = true
					fmt.Println("finished")
				}
			}
		} else {
			time.Sleep(100 * time.Millisecond)
		}
	}

}
