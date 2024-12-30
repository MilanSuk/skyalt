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
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	NMSG_EXIT = 0
	NMSG_SAVE = 1

	NMSG_GET_ENV = 10
	NMSG_SET_ENV = 11

	NMSG_REFRESH = 20
	NMSG_REDRAW  = 30

	NMSG_INPUT = 40
)

var g_cmds []LayoutCmd
var g_cmds_lock sync.Mutex

func _addCmd(cmd LayoutCmd) {
	g_cmds_lock.Lock()
	defer g_cmds_lock.Unlock()

	g_cmds = append(g_cmds, cmd)
}
func _getCmds() []LayoutCmd {
	if len(g__jobs.jobs) > 0 {
		Layout_RefreshDelayed() //calls _addCmd() which has lock() inside
	}

	g_cmds_lock.Lock()
	defer g_cmds_lock.Unlock()

	cmds := g_cmds
	g_cmds = nil
	return cmds
}

func _build(layout *Layout) {
	if layout.fnBuild != nil {
		layout.fnBuild(layout)
	}
	for _, it := range layout.Childs {
		_build(it)
	}
	for _, it := range layout.dialogs {
		_build(it.Layout)
	}
}

func _draw(layout *Layout, rects map[uint64]Rect, out_buffs map[uint64][]LayoutDrawPrim) {
	canvas, found := rects[layout.Hash]
	if found {
		if layout.fnDraw != nil && canvas.Is() {
			//fmt.Println("fnDraw", layout.Name, canvas.H)
			paint := layout.fnDraw(canvas, layout)
			if len(paint.buffer) > 0 {
				out_buffs[layout.Hash] = paint.buffer
			}
		}
	}

	for _, it := range layout.Childs {
		_draw(it, rects, out_buffs)
	}
}

// func main_sdk(port int) {
func main() {
	if len(os.Args) < 2 {
		log.Fatal("missing 'port' argument: ", os.Args)
	}
	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	client, client_err := NewNetClient("localhost", port)
	if client_err != nil {
		log.Fatal(client_err)
	}

	defer g__processes.Destroy()

	var layout *Layout
	for {
		msg, msg_err := client.ReadInt()
		if msg_err != nil {
			log.Fatal(msg_err)
		}
		switch msg {
		case NMSG_EXIT:
			return
		case NMSG_SAVE:
			os.Mkdir("data", os.ModePerm)
			_skyalt_save()

		case NMSG_GET_ENV:
			client.WriteArray(OsMarshal(*OpenFile_Settings()))

		case NMSG_SET_ENV:
			data, err := client.ReadArray()
			if err != nil {
				log.Fatal(err)
			}
			OsUnmarshal(data, OpenFile_Settings())

		case NMSG_REFRESH:
			//build
			layout = _newLayoutRoot()
			root := &Root{}
			layout.fnBuild = root.Build
			_build(layout)

			//root & dialogs
			hash, err := client.ReadInt()
			if err != nil {
				log.Fatal(err)
			}

			for hash > 0 {
				lay := layout._findHash(hash)

				if lay != nil {
					client.WriteInt(1) //found
				} else {
					client.WriteInt(0) //not found
				}

				if lay != nil {
					//parent hash
					{
						lay_parent := layout._findParent(lay)
						if lay_parent != nil {
							client.WriteInt(lay_parent.Hash)
						} else {
							client.WriteInt(0) //no parent
						}
					}

					client.WriteArray(OsMarshal(lay))

					//draw
					rects := make(map[uint64]Rect)
					{
						data, err := client.ReadArray()
						if err != nil {
							log.Fatal(err)
						}
						OsUnmarshal(data, &rects)
					}
					out_buffs := make(map[uint64][]LayoutDrawPrim)
					_draw(lay, rects, out_buffs)
					client.WriteArray(OsMarshal(out_buffs))
				}

				//read next hash
				hash, err = client.ReadInt()
				if err != nil {
					log.Fatal(err)
				}
			}

			client.WriteArray(OsMarshal(_getCmds()))

		case NMSG_REDRAW:
			if layout == nil {
				log.Fatal(fmt.Errorf("layout == nil"))
			}

			//recv
			js, err := client.ReadArray()
			if err != nil {
				log.Fatal(err)
			}
			type Draw struct {
				Hash   uint64
				Rect   Rect
				Buffer []LayoutDrawPrim
			}
			var redrawHashes []Draw
			OsUnmarshal(js, &redrawHashes)

			//redraw
			for i, it := range redrawHashes {
				lay := layout._findHash(it.Hash)
				if lay != nil && lay.fnDraw != nil && it.Rect.Is() {
					redrawHashes[i].Buffer = lay.fnDraw(it.Rect, lay).buffer
				}
			}

			//send
			client.WriteArray(OsMarshal(redrawHashes))
			client.WriteArray(OsMarshal(_getCmds()))

		case NMSG_INPUT:
			if layout == nil {
				log.Fatal(fmt.Errorf("layout == nil"))
			}

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

			inLayout := layout._findHash(hash)
			if inLayout == nil {
				log.Fatal(fmt.Errorf("layout %d not found", hash))
			}

			if in.Shortcut_key != 0 {
				if inLayout.fnInput != nil {
					inLayout.fnInput(in, inLayout)
				}

			} else if in.Drop_path != "" {
				if inLayout.dropFile != nil {
					inLayout.dropFile(in.Drop_path)
				}

			} else if in.SetDropMove {
				if inLayout.dropMove != nil {
					inLayout.dropMove(in.DropSrc, in.DropDst)
				}

			} else if in.SetEdit {
				if inLayout.fnSetEditbox != nil {
					inLayout.fnSetEditbox(in.EditValue, in.EditEnter)
				}

			} else if in.Pick.Line > 0 {
				in.Pick.time_sec = float64(time.Now().UnixMilli()) / 1000
				OpenFile_AssistantChat().findPickOrAdd(in.Pick)
				OpenFile_AssistantChat().AppName = in.PickApp

			} else if inLayout.fnInput != nil {
				inLayout.fnInput(in, inLayout)
			}

			client.WriteArray(OsMarshal(_getCmds()))
		}

		g__jobs.maintenance()
		g__processes.maintenance()
	}
}
