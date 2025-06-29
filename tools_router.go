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
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type ToolsRouterMsg struct {
	msg_id   uint64
	user_uid string
	ui_uid   uint64
	appName  string
	funcName string

	start_time float64

	fnProgress func(cmdsJs [][]byte, err error, start_time float64)                           //call when 'out_flushed_cmds'
	fnDone     func(dataJs []byte, uiJs []byte, cmdsJs []byte, err error, start_time float64) //called when msg is done

	stop     atomic.Bool
	out_done atomic.Bool

	progress_done  float64
	progress_label string

	out_data []byte
	out_ui   []byte
	out_cmds []byte

	out_flushed_cmds [][]byte
	out_error        error

	drawit bool
}

func (msg *ToolsRouterMsg) Stop() {
	msg.stop.Store(true)
}

func (msg *ToolsRouterMsg) Done() {
	msg.out_done.Store(true)
}
func (msg *ToolsRouterMsg) Progress(done float64, label string) bool {

	msg.progress_done = done
	msg.progress_label = label

	return !msg.stop.Load()
}

type ToolsRouter struct {
	server *ToolsServer

	lock sync.Mutex

	log ToolsLog

	msgs         map[uint64]*ToolsRouterMsg
	msgs_counter atomic.Uint64 //to create unique msg_id

	apps map[string]*ToolsApp

	refresh_progress_time float64

	llms *LLMs
	sync *ToolsSync

	mic *ToolsMicMalgo
}

func NewToolsRouter(start_port int) (*ToolsRouter, error) {
	var err error

	router := &ToolsRouter{}
	router.mic = NewToolsMicMalgo(router)
	router.server = NewToolsServer(start_port)
	router.msgs = make(map[uint64]*ToolsRouterMsg)
	router.apps = make(map[string]*ToolsApp)
	router.log.Name = "ToolsRouter"

	router.llms, err = NewLLMs(router)
	if err != nil {
		return nil, err
	}

	router.sync, err = NewToolsSync(router)
	if err != nil {
		return nil, err
	}

	//hot reload
	go func() {
		inited := false
		for {
			router._hotReload()
			if !inited {
				router.sync.Upload_LoadFiles()
				router.sync.Upload_deviceDefaultDPI()
				inited = true
			}

			time.Sleep(1000 * time.Millisecond)
		}
	}()

	//apps
	go router.RunNet()

	return router, nil
}

func (router *ToolsRouter) Destroy() {
	for _, app := range router.apps {
		app.Destroy() //send exit
	}

	router.server.Destroy()

	router.mic.Destroy()

	router.Save()
}

func (router *ToolsRouter) Save() {
	router.lock.Lock()
	defer router.lock.Unlock()
}

func (router *ToolsRouter) FindApp(appName string) *ToolsApp {
	router.lock.Lock()
	app := router.apps[appName]
	router.lock.Unlock()

	if app == nil {
		router._reloadAppList()

		//try again
		router.lock.Lock()
		app = router.apps[appName]
		router.lock.Unlock()
	}

	if app == nil {
		router.log.Error(fmt.Errorf("app '%s' not found", appName))
	}

	return app
}

func (router *ToolsRouter) GetRootApp() *ToolsApp {
	return router.FindApp("Root")
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

func _ToolsRouter_getJSON(params any) ([]byte, error) {
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

func (router *ToolsRouter) CallUpdateDev() {
	router.lock.Lock()
	defer router.lock.Unlock()

	for _, app := range router.apps {
		if app.Process.IsRunning() {
			_ToolsCaller_UpdateDev(app.Process.port, router.log.Error)
		}
	}
}

func (router *ToolsRouter) CallUpdateAsync(ui_uid uint64, sub_uid uint64, appName string, funcName string, fnProgress func(cmdsJs [][]byte, err error, start_time float64), fnDone func(dataJs []byte, uiJs []byte, cmdsJs []byte, err error, start_time float64)) {
	app := router.FindApp(appName)
	if app == nil {
		return
	}

	msg_id := router.msgs_counter.Add(1)
	msg := &ToolsRouterMsg{msg_id: msg_id, user_uid: "", appName: appName, funcName: "_update_" + funcName, fnProgress: fnProgress, fnDone: fnDone, start_time: Tools_Time()}
	//msg.drawit = true

	router.lock.Lock()
	router.msgs[msg_id] = msg
	router.lock.Unlock()

	//call
	go func() {
		defer msg.Done()

		out_subUiJs, out_cmdsJs, out_error := _ToolsCaller_CallUpdate(app.Process.port, msg_id, ui_uid, sub_uid, router.log.Error)
		msg.out_error = out_error

		if router.log.Error(out_error) == nil {
			msg.out_ui = out_subUiJs
			msg.out_cmds = out_cmdsJs
		}

	}()
}

func (router *ToolsRouter) CallChangeAsync(ui_uid uint64, appName string, funcName string, change ToolsSdkChange, fnProgress func(cmdsJs [][]byte, err error, start_time float64), fnDone func(dataJs []byte, uiJs []byte, cmdsJs []byte, err error, start_time float64)) {
	app := router.FindApp(appName)
	if app == nil {
		return
	}

	msg_id := router.msgs_counter.Add(1)
	msg := &ToolsRouterMsg{msg_id: msg_id, user_uid: "", appName: appName, funcName: "_change_" + funcName, fnProgress: fnProgress, fnDone: fnDone, start_time: Tools_Time()}

	msg.drawit = true

	router.lock.Lock()
	router.msgs[msg_id] = msg
	router.lock.Unlock()

	//call
	go func() {
		defer msg.Done()

		out_dataJs, out_cmdsJs, out_error := _ToolsCaller_CallChange(app.Process.port, msg_id, ui_uid, change, router.log.Error)
		msg.out_error = out_error

		if router.log.Error(out_error) == nil {
			msg.out_data = out_dataJs
			msg.out_cmds = out_cmdsJs
		}

	}()
}

func (router *ToolsRouter) CallBuildAsync(ui_uid uint64, appName string, funcName string, params interface{}, fnProgress func(cmdsJs [][]byte, err error, start_time float64), fnDone func(dataJs []byte, uiJs []byte, cmdsJs []byte, err error, start_time float64)) *ToolsRouterMsg {
	app := router.FindApp(appName)
	if app == nil {
		return nil
	}

	msg_id := router.msgs_counter.Add(1)
	msg := &ToolsRouterMsg{msg_id: msg_id, user_uid: "", ui_uid: ui_uid, appName: appName, funcName: funcName, fnProgress: fnProgress, fnDone: fnDone, start_time: Tools_Time()}

	router.lock.Lock()
	router.msgs[msg_id] = msg
	router.lock.Unlock()

	jsParams, err := _ToolsRouter_getJSON(params)
	if router.log.Error(err) != nil {
		return nil
	}

	//call
	go func() {
		defer msg.Done()

		//start it
		err = app.CheckRun()
		if router.log.Error(err) != nil {
			msg.out_error = err
			return
		}

		//call it - no parent!
		msg.out_data, msg.out_ui, msg.out_cmds, msg.out_error = _ToolsCaller_CallBuild(app.Process.port, msg_id, ui_uid, funcName, jsParams, router.log.Error)
		router.log.Error(msg.out_error)
	}()

	return msg
}

func (router *ToolsRouter) AddRecompileMsg(appName string) *ToolsRouterMsg {
	msg_id := router.msgs_counter.Add(1)
	msg := &ToolsRouterMsg{msg_id: msg_id, user_uid: "_compile_", ui_uid: 0, appName: appName, funcName: "_compile_", fnDone: nil, start_time: Tools_Time()}

	router.lock.Lock()
	router.msgs[msg_id] = msg
	router.lock.Unlock()

	return msg
}

func (router *ToolsRouter) FindRecompileMsg(appName string) *ToolsRouterMsg {
	router.lock.Lock()
	defer router.lock.Unlock()

	for _, msg := range router.msgs {
		if msg != nil && msg.user_uid == "_compile_" && msg.appName == appName {
			return msg
		}
	}
	return nil
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

	TM := Tools_Time()
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
			TM := Tools_Time()
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
		msg.fnDone(msg.out_data, msg.out_ui, msg.out_cmds, msg.out_error, msg.start_time)
	}

	return redraw
}

func (router *ToolsRouter) RunNet() {

	type SdkMsg struct {
		Id             string
		AppName        string
		FuncName       string
		Progress_label string
		Progress_done  float64
		Start_time     float64
	}

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
					appName, err := cl.ReadArray()
					if router.log.Error(err) == nil {
						str, err := cl.ReadArray()
						if router.log.Error(err) == nil {
							fmt.Printf("Router's print '%s' app: %s\n", string(appName), string(str))
						}
					}

				case "register":
					appName, err := cl.ReadArray()
					if router.log.Error(err) == nil {
						port, err := cl.ReadInt()
						if router.log.Error(err) == nil {
							app := router.FindApp(string(appName))
							if app != nil {
								app.Process.port = int(port)
							}
						}
					}

				case "generate_app":
					appName, err := cl.ReadArray()
					if router.log.Error(err) == nil {
						app := router.FindApp(string(appName))
						if app != nil {
							app.Tick(true) //err ....
						}
					}

				case "get_llm_usage":
					usage := router.llms.GetUsage()

					usageJs, err := json.Marshal(usage)
					router.log.Error(err)

					err = cl.WriteArray(usageJs)
					router.log.Error(err)

				case "rename_app":
					oldName, err := cl.ReadArray()
					if router.log.Error(err) == nil {
						newNameBytes, err := cl.ReadArray()
						if router.log.Error(err) == nil {
							app := router.FindApp(string(oldName))
							if app != nil {
								newName, renameErr := app.Rename(string(newNameBytes))
								if renameErr == nil {
									router._reloadAppList()
								}

								//send back
								err = cl.WriteArray([]byte(newName))
								router.log.Error(err)

								var errBytes []byte
								if renameErr != nil {
									errBytes = []byte(renameErr.Error())
								}
								err = cl.WriteArray(errBytes)
								router.log.Error(err)
							}
						}
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

				case "sub_call":
					msg_id, err := cl.ReadInt()
					if router.log.Error(err) == nil {
						ui_uid, err := cl.ReadInt()
						if router.log.Error(err) == nil {
							appName, err := cl.ReadArray()
							if router.log.Error(err) == nil {
								funcName, err := cl.ReadArray()
								if router.log.Error(err) == nil {
									jsParams, err := cl.ReadArray()
									if router.log.Error(err) == nil {
										var dataJs []byte
										var uiJs []byte
										var cmdsJs []byte
										var out_Error error
										app := router.FindApp(string(appName))
										if app != nil {
											//start it
											out_Error = app.CheckRun()
											if router.log.Error(out_Error) == nil {
												//call it
												dataJs, uiJs, cmdsJs, out_Error = _ToolsCaller_CallBuild(app.Process.port, msg_id, ui_uid, string(funcName), jsParams, router.log.Error)
												router.log.Error(out_Error)
											}
										} else {
											out_Error = fmt.Errorf("app '%s' not found", string(appName))
											router.log.Error(out_Error)
										}

										errStr := ""
										if out_Error != nil {
											errStr = out_Error.Error()
										}

										err = cl.WriteArray(dataJs)
										router.log.Error(err)
										err = cl.WriteArray(uiJs)
										router.log.Error(err)
										err = cl.WriteArray(cmdsJs)
										router.log.Error(err)
										err = cl.WriteArray([]byte(errStr))
										router.log.Error(err)
									}
								}
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
									msg.out_flushed_cmds = append(msg.out_flushed_cmds, cmdsJs)
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
							msgs = append(msgs, SdkMsg{Id: m.user_uid, AppName: m.appName, FuncName: m.funcName, Progress_label: m.progress_label, Progress_done: m.progress_done, Start_time: m.start_time})
						}
					}

					msgsJs, err := json.Marshal(msgs)
					if router.log.Error(err) == nil {
						err = cl.WriteArray(msgsJs)
						router.log.Error(err)
					}

				case "get_logs":
					start_i, err := cl.ReadInt()
					if router.log.Error(err) == nil {
						logs := router.log.Get(int(start_i))
						logsJs, err := json.Marshal(logs)
						router.log.Error(err)
						err = cl.WriteArray(logsJs)
						router.log.Error(err)
					}

				case "get_mic_info":
					micJs, err := json.Marshal(&router.mic.info)
					router.log.Error(err)
					err = cl.WriteArray(micJs)
					router.log.Error(err)

				case "stop_mic":
					router.mic.FinishAll(false)

				case "set_msg_name":
					user_uid, err := cl.ReadInt()
					if router.log.Error(err) == nil {
						user_id, err := cl.ReadArray()
						if router.log.Error(err) == nil {
							router.lock.Lock()
							{
								msg, found := router.msgs[user_uid]
								if found && msg != nil {
									msg.user_uid = string(user_id)
								}
							}
							router.lock.Unlock()
						}
					}

				case "find_msg_name":
					user_uid, err := cl.ReadArray()
					if router.log.Error(err) == nil {
						msg := router.FindMessage(string(user_uid))

						if msg != nil {
							err := cl.WriteInt(1) //exist
							router.log.Error(err)

							msg := SdkMsg{Id: string(user_uid), AppName: msg.appName, FuncName: msg.funcName, Progress_label: msg.progress_label, Progress_done: msg.progress_done, Start_time: msg.start_time}
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
					appName, err := cl.ReadArray()
					if router.log.Error(err) == nil {

						var schemas []*ToolsOpenAI_completion_tool
						app := router.FindApp(string(appName))
						if app != nil {
							schemas = app.GetAllSchemas()
						}

						oaiJs, err := json.Marshal(schemas)
						router.log.Error(err)

						err = cl.WriteArray(oaiJs)
						router.log.Error(err)
					}

				case "get_tool_data":
					appName, err := cl.ReadArray()
					if router.log.Error(err) == nil {

						var promptsJs []byte

						app := router.FindApp(string(appName))
						if app != nil {
							promptsJs, err = json.MarshalIndent(app.Prompts, "", " ")
							router.log.Error(err)
						}
						err = cl.WriteArray(promptsJs)
						router.log.Error(err)
					}

				case "storage_changed":
					appName, err := cl.ReadArray()
					if router.log.Error(err) == nil {
						app := router.FindApp(string(appName))
						if app != nil {
							app.storage_changes++
						}
					}

				case "llm_complete":
					msg_id, err := cl.ReadInt()
					if router.log.Error(err) == nil {
						compJs, err := cl.ReadArray()
						if router.log.Error(err) == nil {

							var msg *ToolsRouterMsg
							router.lock.Lock()
							{
								msg = router.msgs[msg_id]
							}
							router.lock.Unlock()

							//exe
							if msg != nil {
								var comp LLMComplete
								err := json.Unmarshal(compJs, &comp)
								if router.log.Error(err) == nil {

									comp.delta = func(msg *ChatMsg) {
										msgJs, err := json.Marshal(msg)
										if router.log.Error(err) == nil {
											err = cl.WriteArray(msgJs) //send delta
											router.log.Error(err)
										}
									}

									usecase := "chat"
									if comp.AppName != "" {
										usecase = "tools"
									}
									err = router.llms.Complete(&comp, msg, usecase)
									if router.log.Error(err) == nil {
										//save back
										compJs, err = json.Marshal(&comp)
										router.log.Error(err)
									}
								}
							}

							//send back
							err = cl.WriteArray(nil) //empty delta
							router.log.Error(err)

							err = cl.WriteArray(compJs)
							router.log.Error(err)
						}
					}

				case "llm_transcribe":
					msg_id, err := cl.ReadInt()
					if router.log.Error(err) == nil {
						compJs, err := cl.ReadArray()
						if router.log.Error(err) == nil {

							var msg *ToolsRouterMsg
							router.lock.Lock()
							{
								msg = router.msgs[msg_id]
							}
							router.lock.Unlock()

							//exe
							if msg != nil {
								var comp LLMTranscribe
								err := json.Unmarshal(compJs, &comp)
								if router.log.Error(err) == nil {

									err = router.llms.Transcribe(&comp)
									if router.log.Error(err) == nil {
										//save back
										compJs, err = json.Marshal(&comp)
										router.log.Error(err)
									}
								}
							}

							//send back
							err = cl.WriteArray(nil) //empty delta
							router.log.Error(err)

							err = cl.WriteArray(compJs)
							router.log.Error(err)
						}
					}
				}
			}
		}()
	}
}

func (router *ToolsRouter) _reloadAppList() {
	files, err := os.ReadDir("apps")
	if router.log.Error(err) != nil {
		return
	}

	router.lock.Lock()
	defer router.lock.Unlock()

	//add new apps
	for _, info := range files {
		if !info.IsDir() {
			continue
		}

		appName := info.Name()

		_, found := router.apps[appName]
		if !found {
			app, err := NewToolsApp(appName, router)
			if router.log.Error(err) == nil {
				router.apps[appName] = app
			}
		}
	}
	//remove deleted apps
	for appName := range router.apps {
		found := false
		for _, info := range files {
			if info.IsDir() && appName == info.Name() {
				found = true
				break
			}
		}
		if !found {
			router.apps[appName].Destroy() //exit

			delete(router.apps, appName)
		}
	}
}

func (router *ToolsRouter) _hotReload() {

	router._reloadAppList()

	//ticks
	for appName, app := range router.apps {
		err := app.Tick(false)
		if err != nil {
			router.log.Error(fmt.Errorf("%s: %w", appName, err))
		}
	}
}
