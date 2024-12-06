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
	"sync"
	"time"
)

const (
	NMSG_EXIT = 0
	NMSG_SAVE = 1

	NMSG_REFRESH_START  = 2
	NMSG_REFRESH_UPDATE = 3

	NMSG_CALL = 5
)

type ProgressItem struct {
	start  time.Time
	layout *Layout
}

type RefreshList struct {
	root *Root

	items map[uint64]ProgressItem //[hash]
	lock  sync.Mutex
}

func NewProgressList() *RefreshList {
	list := &RefreshList{}
	list.items = make(map[uint64]ProgressItem)

	list.root = NewFile_Root()
	list.root.layout = _newLayout(0, 0, 0, 0, "Root", nil)
	list.root.layout.fnBuild = list.root.Build

	return list
}

func (list *RefreshList) Add(layout *Layout) bool {
	list.lock.Lock()
	defer list.lock.Unlock()

	_, found := list.items[layout.Hash] //exist

	if !found {
		list.items[layout.Hash] = ProgressItem{layout: layout, start: time.Now()}
	}
	return found
}
func (list *RefreshList) Remove(layout *Layout) {
	list.lock.Lock()
	defer list.lock.Unlock()

	delete(list.items, layout.Hash)
}

func (list *RefreshList) Wait(timeout_ms int64) {
	st := time.Now()

	running := true
	for running {

		list.lock.Lock()
		num_wait := 0
		for _, it := range list.items {
			if it.layout.Progress > 0 {
				dt := float64(time.Since(it.start).Milliseconds())
				rest := int64((dt / it.layout.Progress) - dt)
				if rest < timeout_ms {
					num_wait++
				}

			} else {
				num_wait++
			}
		}
		list.lock.Unlock()

		if num_wait == 0 {
			running = false
		}

		if time.Since(st).Milliseconds() >= timeout_ms {
			running = false
		}
	}

	//fmt.Println("waited:", time.Since(st).Seconds())
}

func (list *RefreshList) _buildInit(layout *Layout) {
	if layout == nil {
		layout = list.root.layout
	}

	running := list.Add(layout) //add
	if !running {               //not needed!
		go func() {
			layout.Progress = 0.0
			layout.Childs = nil //not needed!
			layout.Canvas = Rect{}

			if layout.fnBuild != nil {
				layout.fnBuild()
			}
			layout.Progress = 1.0
			layout.updated = true

			list.Remove(layout) //remove

			for _, it := range layout.Childs {
				list._buildInit(it)
			}
		}()
	} else {
		fmt.Println("This should never happen: _buildInit()")
	}
}

func (list *RefreshList) _draw(layout *Layout, all_canvases map[uint64]Rect) {
	if layout == nil {
		layout = list.root.layout
	}

	canvas, found := all_canvases[layout.Hash]
	if found && layout.updated { //must be update!
		if layout.Canvas != canvas {
			layout.Canvas = canvas
			layout.done = false
			layout.redraw = true
		}
	}

	if layout.redraw {
		running := list.Add(layout) //add
		if !running {
			layout.redraw = false
			go func() {
				layout.Progress = 0.0
				layout.buffer = nil

				if layout.fnDraw != nil {
					layout.fnDraw()
				}
				layout.Progress = 1.0
				layout.done = true

				list.Remove(layout) //remove
			}()
		}
	}

	for _, it := range layout.Childs {
		list._draw(it, all_canvases)
	}
}

func (list *RefreshList) _extractBuffers(layout *Layout, out_buffs map[uint64][]LayoutDrawPrim) bool {
	if layout == nil {
		layout = list.root.layout
	}

	done := layout.done

	if layout.Canvas.Is() && len(layout.buffer) > 0 && layout.done {
		out_buffs[layout.Hash] = layout.buffer
		layout.buffer = nil
	}

	for _, it := range layout.Childs {
		if !list._extractBuffers(it, out_buffs) {
			done = false
		}
	}

	return done
}

func _save() {
	if g_Root != nil {
		_write_file("Root-Root", g_Root)
		g_Root = nil
	}
	//...
}

func main_sdk(port int) {
	client, client_err := NewNetClient("localhost", port)
	if client_err != nil {
		log.Fatal(client_err)
	}

	var progresses []*RefreshList

	for {
		msg, msg_err := client.ReadInt()
		if msg_err != nil {
			log.Fatal(msg_err)
		}
		switch msg {
		case NMSG_EXIT:
			return
		case NMSG_SAVE:
			_save()

		case NMSG_REFRESH_START:
			progress := NewProgressList()
			progress._buildInit(nil)
			progresses = append(progresses, progress)

		case NMSG_REFRESH_UPDATE:
			if len(progresses) == 0 {
				log.Fatal(fmt.Errorf("NMSG_REFRESH_START was not call or NMSG_REFRESH_UPDATE ended"))
			}

			progress := progresses[len(progresses)-1]

			//recv .Canvas
			items := make(map[uint64]Rect)
			{
				data, err := client.ReadArray()
				if err != nil {
					log.Fatal(err)
				}
				OsUnmarshal(data, &items)
			}

			progress._draw(nil, items)

			progress.Wait(100) //max 100ms

			//send buffers back
			refresh_finished := false
			{
				buffs := make(map[uint64][]LayoutDrawPrim)
				refresh_finished = progress._extractBuffers(nil, buffs)
				client.WriteArray(OsMarshal(buffs))
			}

			client.WriteArray(OsMarshal(progress.root.layout))

			if refresh_finished {
				if len(progress.items) > 0 {
					fmt.Println("This should never happen: main_sdk()")
				}

				client.WriteInt(1)
			} else {
				client.WriteInt(0)
			}

		case NMSG_CALL:
			hash, err := client.ReadInt()
			if err != nil {
				log.Fatal(err)
			}
			fnId, err := client.ReadInt()
			if err != nil {
				log.Fatal(err)
			}
			prms, err := client.ReadArray()
			if err != nil {
				log.Fatal(err)
			}

			layout := NewFile_Root().layout
			if layout == nil {
				log.Fatal(fmt.Errorf("root.layout == nil"))
			}
			layout = layout._findHash(hash)
			if layout == nil {
				log.Fatal(fmt.Errorf("layout_hash %d not found", hash))
			}

			//...
			fmt.Println(hash, fnId, prms)

		}

		//maintenance
		for i := len(progresses) - 2; i >= 0; i-- { //-2=not last
			if len(progresses[i].items) == 0 {
				progresses = append(progresses[:i], progresses[i+1:]...) //remove
			}
		}
	}
}
