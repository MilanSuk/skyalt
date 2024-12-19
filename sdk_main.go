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
)

const (
	NMSG_EXIT = 0
	NMSG_SAVE = 1

	NMSG_GET_ENV = 10
	NMSG_SET_ENV = 11

	NMSG_REFRESH = 20

	NMSG_INPUT = 30
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

func _build(layout *Layout) {
	if layout.fnBuild != nil {
		//fmt.Println("fnBuild", layout.Name)
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
	if g_OpenAI != nil {
		_write_file("OpenAI-OpenAI", g_OpenAI)
		g_OpenAI = nil
	}
	if g_Xai != nil {
		_write_file("Xai-Xai", g_Xai)
		g_Xai = nil
	}
	if g_Whispercpp != nil {
		_write_file("Whispercpp-Whispercpp", g_Whispercpp)
		g_Whispercpp = nil
	}
	if g_Assistant != nil {
		_write_file("Assistant-Assistant", g_Assistant)
		g_Assistant = nil
	}

	//...
}

func main_sdk(port int) {
	client, client_err := NewNetClient("localhost", port)
	if client_err != nil {
		log.Fatal(client_err)
	}

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
			_save()

		case NMSG_GET_ENV:
			client.WriteArray(OsMarshal(*NewFile_Env()))

		case NMSG_SET_ENV:
			data, err := client.ReadArray()
			if err != nil {
				log.Fatal(err)
			}
			OsUnmarshal(data, NewFile_Env())

		case NMSG_REFRESH:
			//build
			layout = _newLayoutRoot()
			layout.fnBuild = NewFile_Root().Build
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
			client.WriteInt(uint64(len(g__jobs.jobs)))

		case NMSG_INPUT:
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

			if layout == nil {
				log.Fatal(fmt.Errorf("layout == nil"))
			}

			inLayout := layout._findHash(hash)
			if inLayout == nil {
				log.Fatal(fmt.Errorf("layout %d not found", hash))
			}

			if in.Shortcut_key != 0 {
				inLayout := layout._findShortcut(in.Shortcut_key)
				if inLayout != nil {
					inLayout.fnInput(in, inLayout)
				}

			} else if in.Drop_path != "" {
				if inLayout.fnSetEditbox != nil {
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

			} else if in.Pick.File != "" {
				NewFile_Assistant().findPickOrAdd(in.Pick)

			} else if inLayout.fnInput != nil {
				inLayout.fnInput(in, inLayout)
			}

			client.WriteArray(OsMarshal(_getCmds()))
		}

		g__jobs.maintenance()
	}
}
