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
	"sync/atomic"
	"time"
)

const (
	NMSG_EXIT = 0
	NMSG_SAVE = 1

	NMSG_GET_ENV = 10
	NMSG_SET_ENV = 11

	NMSG_REFRESH_START  = 20
	NMSG_REFRESH_UPDATE = 21

	NMSG_INPUT_START  = 30
	NMSG_INPUT_UPDATE = 31
)

var g_cmds []LayoutCmd
var g_cmds_lock sync.Mutex

func _addCmd(cmd LayoutCmd) {
	g_cmds_lock.Lock()
	defer g_cmds_lock.Unlock()

	g_cmds = append(g_cmds, cmd)
}
func _getCmds() []LayoutCmd {
	g_cmds_lock.Lock()
	defer g_cmds_lock.Unlock()

	cmds := g_cmds
	g_cmds = nil
	return cmds
}

type RefreshItem struct {
	start  time.Time
	layout *Layout
}

type RefreshList struct {
	root *Root

	items map[uint64]RefreshItem //[hash]

	buffs map[uint64][]LayoutDrawPrim

	lock sync.Mutex
}

func NewProgressList(layout *Layout) *RefreshList {
	list := &RefreshList{}
	list.items = make(map[uint64]RefreshItem)

	list.buffs = make(map[uint64][]LayoutDrawPrim)

	list.root = NewFile_Root()
	if layout != nil {
		list.root.layout = layout
	} else {
		list.root.layout = _newLayoutRoot()
	}
	list.root.layout.fnBuild = list.root.Build

	return list
}

func (list *RefreshList) AddBuffer(hash uint64, buffer []LayoutDrawPrim) {
	list.lock.Lock()
	defer list.lock.Unlock()
	list.buffs[hash] = buffer
}

func (list *RefreshList) GetBuffs() map[uint64][]LayoutDrawPrim {
	list.lock.Lock()
	defer list.lock.Unlock()
	buffs := list.buffs
	list.buffs = make(map[uint64][]LayoutDrawPrim)
	return buffs
}

func (list *RefreshList) Add(layout *Layout) bool {
	list.lock.Lock()
	defer list.lock.Unlock()

	_, found := list.items[layout.Hash] //exist

	if !found {
		list.items[layout.Hash] = RefreshItem{layout: layout, start: time.Now()}
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
			//layout.Childs = nil //not needed!
			layout.Canvas = Rect{0, 0, -1, -1}

			if layout.fnBuild != nil {
				fmt.Println("fnBuild", layout.Name, layout.Canvas.H)
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
		log.Fatal("This should never happen: _buildInit()")
	}
}

func (list *RefreshList) _draw(layout *Layout, rects map[uint64]Rect) {
	if layout == nil {
		layout = list.root.layout
	}

	canvas, found := rects[layout.Hash]
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

				if layout.fnDraw != nil && layout.Canvas.Is() {
					fmt.Println("fnDraw", layout.Name, layout.Canvas.H)
					layout.fnDraw(layout.Canvas)
				}
				layout.Progress = 1.0
				list.AddBuffer(layout.Hash, layout.buffer)
				layout.buffer = nil
				layout.done = true

				list.Remove(layout) //remove
			}()
		}
	}

	for _, it := range layout.Childs {
		list._draw(it, rects)
	}
}

func (list *RefreshList) _isLayoutDone(layout *Layout) bool {
	if layout == nil {
		layout = list.root.layout
	}

	done := layout.done

	for _, it := range layout.Childs {
		if !list._isLayoutDone(it) {
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
	if g_Env != nil {
		_write_file("Env-Env", g_Env)
		g_Env = nil
	}
	if g_Logs != nil {
		_write_file("Logs-Logs", g_Logs)
		g_Logs = nil
	}
	if g_Counter != nil {
		_write_file("Counter-Counter", g_Counter)
		g_Counter = nil
	}
	if g_Microphone != nil {
		_write_file("Microphone-Microphone", g_Microphone)
		g_Microphone = nil
	}

	//...
}

func main_sdk(port int) {
	client, client_err := NewNetClient("localhost", port)
	if client_err != nil {
		log.Fatal(client_err)
	}

	var progresses []*RefreshList
	var last_finished_layout *Layout

	type Input struct {
		hash uint64
		prm  LayoutInput
		done atomic.Bool
	}
	var inputs []*Input

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

		case NMSG_GET_ENV:
			client.WriteArray(OsMarshal(*NewFile_Env()))

		case NMSG_SET_ENV:
			data, err := client.ReadArray()
			if err != nil {
				log.Fatal(err)
			}
			OsUnmarshal(data, NewFile_Env())

		case NMSG_REFRESH_START:
			progress := NewProgressList(nil)
			progress._buildInit(nil)
			progresses = append(progresses, progress)

		case NMSG_REFRESH_UPDATE:
			if len(progresses) == 0 {
				log.Fatal(fmt.Errorf("NMSG_REFRESH_START was not call or NMSG_REFRESH_UPDATE ended"))
			}

			progress := progresses[len(progresses)-1]

			//receive .Canvas
			rects := make(map[uint64]Rect)
			{
				data, err := client.ReadArray()
				if err != nil {
					log.Fatal(err)
				}
				OsUnmarshal(data, &rects)
			}

			progress._draw(nil, rects)

			progress.Wait(100) //max 100ms

			//send buffers back
			refresh_finished := progress._isLayoutDone(nil)

			client.WriteArray(OsMarshal(progress.root.layout))
			client.WriteArray(OsMarshal(progress.GetBuffs()))
			client.WriteArray(OsMarshal(_getCmds()))

			if refresh_finished {
				if len(progress.items) > 0 {
					log.Fatal("This should never happen: main_sdk()")
				}
				last_finished_layout = progress.root.layout

				client.WriteInt(1)
			} else {
				client.WriteInt(0)
			}

		case NMSG_INPUT_START:
			hash, err := client.ReadInt()
			if err != nil {
				log.Fatal(err)
			}
			param, err := client.ReadArray()
			if err != nil {
				log.Fatal(err)
			}
			var in LayoutInput
			OsUnmarshal(param, &in)

			if last_finished_layout == nil {
				break //log.Fatal(fmt.Errorf("Refresh hasn'v finished yet"))
			}

			layout := last_finished_layout._findHash(hash)
			if layout == nil {
				log.Fatal(fmt.Errorf("layout %d not found", hash))
			}

			if in.SetEdit {
				t := &Input{hash: hash, prm: in}
				if layout.fnSetEditbox != nil {
					go func() {
						layout.fnSetEditbox(in.EditValue)
						t.done.Store(true)
					}()
					inputs = append(inputs, t)
				}

			} else if layout.fnInput != nil {
				t := &Input{hash: hash, prm: in}
				go func() {
					layout.fnInput(in)
					t.done.Store(true)
				}()
				inputs = append(inputs, t)
			}

		case NMSG_INPUT_UPDATE:
			//maintenance
			for i := len(inputs) - 1; i >= 0; i-- {
				if inputs[i].done.Load() {
					inputs = append(inputs[:i], inputs[i+1:]...) //remove
				}
			}

			client.WriteArray(OsMarshal(_getCmds()))

			if len(inputs) == 0 {
				client.WriteInt(1)
			} else {
				client.WriteInt(0)
			}
		}

		//maintenance
		for i := len(progresses) - 2; i >= 0; i-- { //-2=not last
			if len(progresses[i].items) == 0 {
				progresses = append(progresses[:i], progresses[i+1:]...) //remove
			}
		}
	}
}
