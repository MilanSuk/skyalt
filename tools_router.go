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
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
)

type ToolsRouterMsg struct {
	user_uid string
	ui_uid   uint64
	funcName string

	start_time float64

	fnProgress func(cmds []ToolCmd, err error, start_time float64)                       //call when 'out_flushed_cmds'
	fnDone     func(bytes []byte, ui *UI, cmds []ToolCmd, err error, start_time float64) //called when msg is done

	stop     atomic.Bool
	out_done atomic.Bool

	progress_done  float64
	progress_label string

	out_bytes        []byte
	out_ui           *UI
	out_cmds         []ToolCmd
	out_flushed_cmds []ToolCmd
	out_error        error

	drawit bool
}

func (msg *ToolsRouterMsg) Stop() {
	msg.stop.Store(true)
}

func (msg *ToolsRouterMsg) Done() {
	msg.out_done.Store(true)
}

type ToolsRouter struct {
	server *ToolsServer

	lock sync.Mutex

	log ToolsLog

	msgs         map[uint64]*ToolsRouterMsg
	msgs_counter atomic.Uint64 //to create unique msg_id

	tools *ToolsCmd
	files *ToolsFiles

	refresh_progress_time float64
}

func NewToolsRouter(FolderTools string, FolderFiles string, start_port int) *ToolsRouter {
	router := &ToolsRouter{}
	router.server = NewToolsServer(start_port)
	router.msgs = make(map[uint64]*ToolsRouterMsg)
	router.log.Name = "ToolsRouter"

	router.files = NewToolsFiles(FolderFiles)
	router.tools = NewToolsCmd(FolderTools, router)

	type SdkMsg struct {
		Id             string
		FuncName       string
		Progress_label string
		Progress_done  float64
		Start_time     float64
	}

	//communicate with tools
	go func() {
		for {
			cl, err := router.server.Accept()
			if router.log.Error(err) != nil {
				continue
			}
			if cl == nil {
				break //close tool
			}

			//serve tool
			go func() {
				defer cl.Destroy()

				mode, err := cl.ReadArray()
				if router.log.Error(err) == nil {
					switch string(mode) {

					case "print":
						str, err := cl.ReadArray()
						if router.log.Error(err) == nil {
							fmt.Println("Router's print:", string(str))
						}

					case "read_file":
						path, err := cl.ReadArray()
						if router.log.Error(err) == nil {
							file, file_data := router.files.GetFile(string(path))
							if file != nil {
								err = cl.WriteInt(1)
								if router.log.Error(err) == nil {
									err = cl.WriteArray(file_data)
									router.log.Error(err)
								}
							} else {
								//file not exist
								err = cl.WriteInt(0)
								router.log.Error(err)
							}
						}
					case "write_file":
						file, err := cl.ReadArray()
						if router.log.Error(err) == nil {
							file_data, err := cl.ReadArray()
							if router.log.Error(err) == nil {
								router.files.SetFile(string(file), file_data)
							}
						}

					case "register":
						port, err := cl.ReadInt()
						if router.log.Error(err) == nil {
							router.lock.Lock()
							{
								router.tools.port = int(port)
							}
							router.lock.Unlock()
						}
					case "progress":
						msg_id, err := cl.ReadInt()
						if router.log.Error(err) == nil {
							done, err := cl.ReadInt()
							if router.log.Error(err) == nil {
								label, err := cl.ReadArray()
								if router.log.Error(err) == nil {

									stop := uint64(0) //not found = sub-ui has same msg_id(which ended)
									router.lock.Lock()
									{
										msg, found := router.msgs[msg_id]
										if found && msg != nil {
											msg.progress_done = float64(done) / 10000
											msg.progress_label = string(label)

											//stop = 0
											if msg.stop.Load() {
												stop = 1
											}
										}
									}
									router.lock.Unlock()

									err = cl.WriteInt(stop)
									router.log.Error(err)
								}
							}
						}

					case "add_cmds":
						msg_id, err := cl.ReadInt()
						if router.log.Error(err) == nil {
							cmdsJs, err := cl.ReadArray()
							if router.log.Error(err) == nil {

								router.lock.Lock()
								{
									msg, found := router.msgs[msg_id]
									if found && msg != nil {
										var cmds []ToolCmd
										err = json.Unmarshal(cmdsJs, &cmds)
										if router.log.Error(err) == nil {
											msg.out_flushed_cmds = cmds
										}
									}
								}
								router.lock.Unlock()
							}
						}

					case "get_msgs":
						var msgs []SdkMsg
						rmsgs := router.GetSortedMsgs()
						for _, m := range rmsgs {
							if m.drawit {
								msgs = append(msgs, SdkMsg{Id: m.user_uid, FuncName: m.funcName, Progress_label: m.progress_label, Progress_done: m.progress_done, Start_time: m.start_time})
							}
						}

						msgsJs, err := json.Marshal(msgs)
						if router.log.Error(err) == nil {
							err = cl.WriteArray(msgsJs)
							router.log.Error(err)
						}

					case "set_msg_name":
						msg_id, err := cl.ReadInt()
						if router.log.Error(err) == nil {
							user_id, err := cl.ReadArray()
							if router.log.Error(err) == nil {
								router.lock.Lock()
								{
									msg, found := router.msgs[msg_id]
									if found && msg != nil {
										msg.user_uid = string(user_id)
									}
								}
								router.lock.Unlock()
							}
						}

					case "find_msg":
						msg_id, err := cl.ReadArray()
						if router.log.Error(err) == nil {
							msg := router.FindMessage(string(msg_id))

							if msg != nil {
								err := cl.WriteInt(1) //exist
								router.log.Error(err)

								msg := SdkMsg{Id: string(msg_id), FuncName: msg.funcName, Progress_label: msg.progress_label, Progress_done: msg.progress_done, Start_time: msg.start_time}
								msgJs, err := json.Marshal(msg)
								if router.log.Error(err) == nil {
									err = cl.WriteArray(msgJs)
									router.log.Error(err)
								}

							} else {
								err := cl.WriteInt(0) //non-exist
								router.log.Error(err)
							}
						}

					case "stop_msg":
						user_uid, err := cl.ReadArray()
						if router.log.Error(err) == nil {
							msg := router.FindMessage(string(user_uid))
							if msg != nil {
								msg.Stop()
							}
						}

					case "get_tools_shemas":
						schemas := router.tools.GetAllSchemas()
						oaiJs, err := json.Marshal(schemas)
						router.log.Error(err)

						err = cl.WriteArray(oaiJs)
						router.log.Error(err)

					case "get_tools_shemas_by_source":
						sourceName, err := cl.ReadArray()
						if router.log.Error(err) == nil {
							schemas := router.tools.GetSchemasForSource(string(sourceName))
							js, err := json.Marshal(schemas)
							router.log.Error(err)

							err = cl.WriteArray(js)
							router.log.Error(err)
						}

					case "get_sources":
						js, err := json.Marshal(router.tools.Sources)
						router.log.Error(err)

						err = cl.WriteArray(js)
						router.log.Error(err)
					}
				}
			}()
		}
	}()

	return router
}

func (router *ToolsRouter) Destroy() {

	router.tools.Destroy() //send exit

	router.server.Destroy()

	router.files.Destroy()

	router.Save()
}

func (router *ToolsRouter) Save() {
	router.lock.Lock()
	defer router.lock.Unlock()

	//files
	err := router.files.Save()
	router.log.Error(err)

}
func (router *ToolsRouter) GetSortedMsgs() []*ToolsRouterMsg {
	router.lock.Lock()
	defer router.lock.Unlock()

	var sortedMsgs []*ToolsRouterMsg
	for _, msg := range router.msgs {
		if msg != nil {
			sortedMsgs = append(sortedMsgs, msg)
		}
	}

	//sort by time
	sort.Slice(sortedMsgs, func(i, j int) bool {
		return sortedMsgs[i].start_time < sortedMsgs[j].start_time
	})
	return sortedMsgs
}

func _ToolsRouter_getJSON(params interface{}) ([]byte, error) {
	jsParams := []byte("{}")
	if params != nil {
		var err error
		jsParams, err = json.Marshal(params)
		if err != nil {
			return nil, err
		}
	}
	return jsParams, nil
}

func (router *ToolsRouter) CallChangeAsync(ui_uid uint64, funcName string, change SdkChange, fnProgress func(cmds []ToolCmd, err error, start_time float64), fnDone func(bytes []byte, uii *UI, cmds []ToolCmd, err error, start_time float64)) {
	msg := &ToolsRouterMsg{user_uid: "", funcName: "_change_" + funcName, fnProgress: fnProgress, fnDone: fnDone, start_time: OsTime()}
	msg_id := router.msgs_counter.Add(1)

	msg.drawit = true

	router.lock.Lock()
	router.msgs[msg_id] = msg
	router.lock.Unlock()

	//call tool change()
	go func() {
		msg.out_cmds, msg.out_error = _ToolsCaller_CallChange(router.tools.port, msg_id, ui_uid, change, router.log.Error)
		router.log.Error(msg.out_error)
		msg.Done()
	}()
}

func (router *ToolsRouter) CallAsync(ui_uid uint64, funcName string, params interface{}, fnProgress func(cmds []ToolCmd, err error, start_time float64), fnDone func(bytes []byte, ui *UI, cmds []ToolCmd, err error, start_time float64)) (*ToolsRouterMsg, error) {
	msg := &ToolsRouterMsg{user_uid: "", ui_uid: ui_uid, funcName: funcName, fnProgress: fnProgress, fnDone: fnDone, start_time: OsTime()}
	msg_id := router.msgs_counter.Add(1)

	router.lock.Lock()
	router.msgs[msg_id] = msg
	router.lock.Unlock()

	jsParams, err := _ToolsRouter_getJSON(params)
	if router.log.Error(err) != nil {
		return nil, err
	}

	//call tool
	go func() {
		//start it
		err = router.tools.CheckRun()
		if router.log.Error(err) != nil {
			msg.out_error = err
			return
		}

		//call it
		msg.out_bytes, msg.out_ui, msg.out_cmds, msg.out_error = _ToolsCaller_CallTool(router.tools.port, msg_id, ui_uid, funcName, jsParams, router.log.Error)
		router.log.Error(msg.out_error)
		msg.Done()
	}()

	return msg, nil
}

func (router *ToolsRouter) AddRecompileMsg() *ToolsRouterMsg {
	msg := &ToolsRouterMsg{user_uid: "_compile_", ui_uid: 0, funcName: "_compile_", fnDone: nil, start_time: OsTime()}
	msg_id := router.msgs_counter.Add(1)

	router.lock.Lock()
	router.msgs[msg_id] = msg
	router.lock.Unlock()

	return msg
}

func (router *ToolsRouter) FindRecompileMsg() *ToolsRouterMsg {
	return router.FindMessage("_compile_")
}

func (router *ToolsRouter) FindMessage(user_uid string) *ToolsRouterMsg {
	router.lock.Lock()
	defer router.lock.Unlock()

	for _, msg := range router.msgs {
		if msg != nil && msg.user_uid == user_uid {
			return msg
		}
	}
	return nil
}

func (router *ToolsRouter) NeedMsgRedraw() bool {
	router.lock.Lock()
	defer router.lock.Unlock()

	TM := OsTime()
	for _, msg := range router.msgs {
		if (TM-msg.start_time) > 0.9 && msg.ui_uid != 1 {
			msg.drawit = true
		}
		if msg.drawit {
			return true
		}
	}
	return false
}

func (router *ToolsRouter) Flush() bool {
	redraw := false

	var doneMsgs []*ToolsRouterMsg
	router.lock.Lock()
	{
		//redraw progress/threads with delay
		{
			TM := OsTime()
			msg_redraw := false
			for _, msg := range router.msgs {
				if (TM-msg.start_time) > 0.9 && msg.ui_uid != 1 {
					msg.drawit = true
					msg_redraw = true
				}
			}

			if msg_redraw && router.refresh_progress_time < TM {
				redraw = true
				router.refresh_progress_time = TM + 1
			}
		}

		//flush cmds
		for _, msg := range router.msgs {
			if msg != nil && len(msg.out_flushed_cmds) > 0 {
				msg.fnProgress(msg.out_flushed_cmds, nil, msg.start_time)
				msg.out_flushed_cmds = nil
			}
		}

		//done messages
		for msg_id, msg := range router.msgs {
			if msg != nil && msg.out_done.Load() {

				if msg.drawit {
					redraw = true
				}

				if msg.fnDone != nil {
					doneMsgs = append(doneMsgs, msg)
				}

				delete(router.msgs, msg_id)
			}
		}
	}
	router.lock.Unlock()

	for _, msg := range doneMsgs {
		msg.fnDone(msg.out_bytes, msg.out_ui, msg.out_cmds, msg.out_error, msg.start_time)
	}

	router.files.Tick()

	return redraw
}
