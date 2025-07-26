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
	"fmt"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type AppsRouterMsg struct {
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

func (msg *AppsRouterMsg) Stop() {
	msg.stop.Store(true)
}

func (msg *AppsRouterMsg) Done() {
	msg.out_done.Store(true)
}
func (msg *AppsRouterMsg) Progress(done float64, label string) bool {

	msg.progress_done = done
	msg.progress_label = label

	return msg.GetContinue()
}
func (msg *AppsRouterMsg) GetContinue() bool {
	return !msg.stop.Load()
}

type AppsRouter struct {
	server *AppsServer

	lock sync.Mutex

	msgs         map[uint64]*AppsRouterMsg
	msgs_counter atomic.Uint64 //to create unique msg_id

	apps map[string]*ToolsApp

	refresh_progress_time float64

	services *Services
}

func NewAppsRouter(start_port int, services *Services) (*AppsRouter, error) {
	router := &AppsRouter{}

	router.services = services
	router.services.fnCallBuildAsync = router.CallBuildAsync
	router.services.fnGetAppPortAndTools = router.GetAppPortAndTools

	router.server = NewAppsServer(start_port)
	router.msgs = make(map[uint64]*AppsRouterMsg)
	router.apps = make(map[string]*ToolsApp)

	//hot reload
	go func() {
		inited := false
		for {
			router._hotReload()
			if !inited {
				router.services.sync.Upload_LoadFiles()
				router.services.sync.Upload_deviceDefaultDPI()
				inited = true
			}

			time.Sleep(1000 * time.Millisecond)
		}
	}()

	//apps
	go router.RunNet()

	return router, nil
}

func (router *AppsRouter) Destroy() {
	for _, app := range router.apps {
		app.Destroy() //send exit
	}

	router.server.Destroy()

	router.Save()
}

func (router *AppsRouter) Tick() bool {
	refresh := false
	devApp := router.FindApp("Device")
	if devApp != nil {
		refresh = router.services.Tick(devApp.storage_changes)
		if refresh {
			router.CallUpdateDev()
		}
	}

	//ticks
	for _, app := range router.apps {
		if app.NeedRefresh() {
			refresh = true
		}
	}

	return refresh
}

func (router *AppsRouter) GetAppPortAndTools(appName string) (int, []*ToolsOpenAI_completion_tool, error) {

	var tools []*ToolsOpenAI_completion_tool
	app_port := -1

	if appName != "" {
		app := router.FindApp(appName)
		if app != nil {
			tools = app.GetAllSchemas()

			//start it
			err := app.CheckRun()
			if err != nil {
				return -1, nil, err
			}

			app_port = app.Process.port

		} else {
			return -1, nil, LogsErrorf("app '%s' not found", appName)
		}
	}

	return app_port, tools, nil
}

func (router *AppsRouter) Save() {
	router.lock.Lock()
	defer router.lock.Unlock()
}

func (router *AppsRouter) FindApp(appName string) *ToolsApp {
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
		LogsErrorf("app '%s' not found", appName)
	}

	return app
}

func (router *AppsRouter) GetRootApp() *ToolsApp {
	return router.FindApp("Root")
}

func (router *AppsRouter) GetSortedMsgs() []*AppsRouterMsg {
	router.lock.Lock()
	defer router.lock.Unlock()

	var sortedMsgs []*AppsRouterMsg
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
		jsParams, err = LogsJsonMarshal(params)
		if err != nil {
			return nil, err
		}
	}
	return jsParams, nil
}

func (router *AppsRouter) CallUpdateDev() {
	router.lock.Lock()
	defer router.lock.Unlock()

	for _, app := range router.apps {
		if app.Process.IsRunning() {
			_ToolsCaller_UpdateDev(app.Process.port)
		}
	}
}

func (router *AppsRouter) CallUpdateAsync(ui_uid uint64, sub_uid uint64, appName string, funcName string, fnProgress func(cmdsJs [][]byte, err error, start_time float64), fnDone func(dataJs []byte, uiJs []byte, cmdsJs []byte, err error, start_time float64)) {
	app := router.FindApp(appName)
	if app == nil {
		return
	}

	msg_id := router.msgs_counter.Add(1)
	msg := &AppsRouterMsg{msg_id: msg_id, user_uid: "", appName: appName, funcName: "_update_" + funcName, fnProgress: fnProgress, fnDone: fnDone, start_time: Tools_Time()}
	//msg.drawit = true

	router.lock.Lock()
	router.msgs[msg_id] = msg
	router.lock.Unlock()

	//call
	go func() {
		defer msg.Done()

		out_subUiJs, out_cmdsJs, out_error := _ToolsCaller_CallUpdate(app.Process.port, msg_id, ui_uid, sub_uid)
		msg.out_error = out_error

		if out_error == nil {
			msg.out_ui = out_subUiJs
			msg.out_cmds = out_cmdsJs
		}

	}()
}

func (router *AppsRouter) CallChangeAsync(ui_uid uint64, appName string, funcName string, change ToolsSdkChange, fnProgress func(cmdsJs [][]byte, err error, start_time float64), fnDone func(dataJs []byte, uiJs []byte, cmdsJs []byte, err error, start_time float64)) {
	app := router.FindApp(appName)
	if app == nil {
		return
	}

	msg_id := router.msgs_counter.Add(1)
	msg := &AppsRouterMsg{msg_id: msg_id, user_uid: "", appName: appName, funcName: "_change_" + funcName, fnProgress: fnProgress, fnDone: fnDone, start_time: Tools_Time()}

	msg.drawit = true

	router.lock.Lock()
	router.msgs[msg_id] = msg
	router.lock.Unlock()

	//call
	go func() {
		defer msg.Done()

		out_dataJs, out_cmdsJs, out_error := _ToolsCaller_CallChange(app.Process.port, msg_id, ui_uid, change)
		msg.out_error = out_error

		if out_error == nil {
			msg.out_data = out_dataJs
			msg.out_cmds = out_cmdsJs
		}

	}()
}

func (router *AppsRouter) CallBuildAsync(ui_uid uint64, appName string, funcName string, params interface{}, fnProgress func(cmdsJs [][]byte, err error, start_time float64), fnDone func(dataJs []byte, uiJs []byte, cmdsJs []byte, err error, start_time float64)) *AppsRouterMsg {
	app := router.FindApp(appName)
	if app == nil {
		return nil
	}

	msg_id := router.msgs_counter.Add(1)
	msg := &AppsRouterMsg{msg_id: msg_id, user_uid: "", ui_uid: ui_uid, appName: appName, funcName: funcName, fnProgress: fnProgress, fnDone: fnDone, start_time: Tools_Time()}

	router.lock.Lock()
	router.msgs[msg_id] = msg
	router.lock.Unlock()

	jsParams, err := _ToolsRouter_getJSON(params)
	if err != nil {
		return nil
	}

	//call
	go func() {
		defer msg.Done()

		//start it
		err = app.CheckRun()
		if err != nil {
			msg.out_error = err
			return
		}

		//call it - no parent!
		msg.out_data, msg.out_ui, msg.out_cmds, msg.out_error = _ToolsCaller_CallBuild(app.Process.port, msg_id, ui_uid, funcName, jsParams)
	}()

	return msg
}

func (router *AppsRouter) AddRecompileMsg(appName string) *AppsRouterMsg {
	msg_id := router.msgs_counter.Add(1)
	msg := &AppsRouterMsg{msg_id: msg_id, user_uid: "_compile_", ui_uid: 0, appName: appName, funcName: "_compile_", fnDone: nil, start_time: Tools_Time()}

	router.lock.Lock()
	router.msgs[msg_id] = msg
	router.lock.Unlock()

	return msg
}

func (router *AppsRouter) FindRecompileMsg(appName string) *AppsRouterMsg {
	router.lock.Lock()
	defer router.lock.Unlock()

	for _, msg := range router.msgs {
		if msg != nil && msg.user_uid == "_compile_" && msg.appName == appName {
			return msg
		}
	}
	return nil
}

func (router *AppsRouter) FindMessage(user_uid string) *AppsRouterMsg {
	router.lock.Lock()
	defer router.lock.Unlock()

	for _, msg := range router.msgs {
		if msg != nil && msg.user_uid == user_uid {
			return msg
		}
	}
	return nil
}

func (router *AppsRouter) NeedMsgRedraw() bool {
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

func (router *AppsRouter) Flush() bool {
	redraw := false

	var doneMsgs []*AppsRouterMsg
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

func (router *AppsRouter) RunNet() {

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
		if LogsError(err) != nil {
			continue
		}
		if cl == nil {
			break //close tool
		}

		//serve tool
		go func() {
			defer cl.Destroy()

			mode, err := cl.ReadArray()
			if err == nil {
				switch string(mode) {

				case "print":
					appName, err := cl.ReadArray()
					if err == nil {
						str, err := cl.ReadArray()
						if err == nil {
							fmt.Printf("Router's print '%s' app: %s\n", string(appName), string(str))
						}
					}

				case "register":
					appName, err := cl.ReadArray()
					if err == nil {
						port, err := cl.ReadInt()
						if err == nil {
							app := router.FindApp(string(appName))
							if app != nil {
								app.Process.port = int(port)
							}
						}
					}

				case "generate_app":
					appName, err := cl.ReadArray()
					if err == nil {
						app := router.FindApp(string(appName))
						if app != nil {
							app.Tick(true) //err ....
						}
					}

				case "get_llm_usage":
					usage := router.services.llms.GetUsage()

					usageJs, _ := LogsJsonMarshal(usage)

					cl.WriteArray(usageJs)

				case "rename_app":
					oldName, err := cl.ReadArray()
					if err == nil {
						newNameBytes, err := cl.ReadArray()
						if err == nil {
							app := router.FindApp(string(oldName))
							if app != nil {
								newName, renameErr := app.Rename(string(newNameBytes))
								if renameErr == nil {
									router._reloadAppList()
								}

								//send back
								cl.WriteArray([]byte(newName))

								var errBytes []byte
								if renameErr != nil {
									errBytes = []byte(renameErr.Error())
								}
								cl.WriteArray(errBytes)
							}
						}
					}

				case "progress":
					msg_id, err := cl.ReadInt()
					if err == nil {
						done, err := cl.ReadInt()
						if err == nil {
							label, err := cl.ReadArray()
							if err == nil {

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

								cl.WriteInt(stop)
							}
						}
					}

				case "sub_call":
					msg_id, err := cl.ReadInt()
					if err == nil {
						ui_uid, err := cl.ReadInt()
						if err == nil {
							appName, err := cl.ReadArray()
							if err == nil {
								funcName, err := cl.ReadArray()
								if err == nil {
									jsParams, err := cl.ReadArray()
									if err == nil {
										var dataJs []byte
										var uiJs []byte
										var cmdsJs []byte
										var out_Error error
										app := router.FindApp(string(appName))
										if app != nil {
											//start it
											out_Error = app.CheckRun()
											if out_Error == nil {
												//call it
												dataJs, uiJs, cmdsJs, out_Error = _ToolsCaller_CallBuild(app.Process.port, msg_id, ui_uid, string(funcName), jsParams)
											}
										} else {
											out_Error = LogsErrorf("app '%s' not found", string(appName))
										}

										errStr := ""
										if out_Error != nil {
											errStr = out_Error.Error()
										}

										cl.WriteArray(dataJs)
										cl.WriteArray(uiJs)
										cl.WriteArray(cmdsJs)
										cl.WriteArray([]byte(errStr))
									}
								}
							}
						}
					}

				case "add_cmds":
					msg_id, err := cl.ReadInt()
					if err == nil {
						cmdsJs, err := cl.ReadArray()
						if err == nil {

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

					msgsJs, _ := LogsJsonMarshal(msgs)
					cl.WriteArray(msgsJs)

				case "get_logs":
					start_i, err := cl.ReadInt()
					if err == nil {
						logs := g_logs.Get(int(start_i))
						logsJs, _ := LogsJsonMarshal(logs)

						cl.WriteArray(logsJs)
					}

				case "get_mic_info":
					micJs, _ := LogsJsonMarshal(&router.services.mic.info)

					cl.WriteArray(micJs)

				case "stop_mic":
					router.services.mic.FinishAll(false)

				case "get_media_info":
					infoJs, _ := router.services.media.GetInfo()
					cl.WriteArray(infoJs)

				case "set_msg_name":
					user_uid, err := cl.ReadInt()
					if err == nil {
						user_id, err := cl.ReadArray()
						if err == nil {
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
					if err == nil {
						msg := router.FindMessage(string(user_uid))

						if msg != nil {
							cl.WriteInt(1) //exist

							msg := SdkMsg{Id: string(user_uid), AppName: msg.appName, FuncName: msg.funcName, Progress_label: msg.progress_label, Progress_done: msg.progress_done, Start_time: msg.start_time}
							msgJs, _ := LogsJsonMarshal(msg)
							cl.WriteArray(msgJs)

						} else {
							cl.WriteInt(0) //non-exist
						}
					}

				case "stop_msg":
					user_uid, err := cl.ReadArray()
					if err == nil {
						msg := router.FindMessage(string(user_uid))
						if msg != nil {
							msg.Stop()
						}
					}

				case "get_tools_shemas":
					appName, err := cl.ReadArray()
					if err == nil {

						var schemas []*ToolsOpenAI_completion_tool
						app := router.FindApp(string(appName))
						if app != nil {
							schemas = app.GetAllSchemas()
						}

						oaiJs, _ := LogsJsonMarshal(schemas)

						cl.WriteArray(oaiJs)
					}

				case "get_tool_data":
					appName, err := cl.ReadArray()
					if err == nil {

						var promptsJs []byte

						app := router.FindApp(string(appName))
						if app != nil {
							promptsJs, _ = LogsJsonMarshalIndent(&app.Prompts)
						}
						cl.WriteArray(promptsJs)
					}

				case "storage_changed":
					appName, err := cl.ReadArray()
					if err == nil {
						app := router.FindApp(string(appName))
						if app != nil {
							app.storage_changes++
						}
					}

				case "llm_complete":
					msg_id, err := cl.ReadInt()
					if err == nil {
						compJs, err := cl.ReadArray()
						if err == nil {

							var msg *AppsRouterMsg
							router.lock.Lock()
							{
								msg = router.msgs[msg_id]
							}
							router.lock.Unlock()

							//exe
							if msg != nil {
								var comp LLMComplete
								err := LogsJsonUnmarshal(compJs, &comp)
								if err == nil {

									comp.delta = func(msg *ChatMsg) {
										msgJs, _ := LogsJsonMarshal(msg)
										cl.WriteArray(msgJs) //send delta
									}

									usecase := "chat"
									if comp.AppName != "" {
										usecase = "tools"
									}
									err = router.services.llms.Complete(&comp, msg, usecase)
									if err == nil {
										//save back
										compJs, _ = LogsJsonMarshal(&comp)
									}
								}
							}

							//send back
							cl.WriteArray(nil) //empty delta
							cl.WriteArray(compJs)
						}
					}

				case "llm_transcribe":
					msg_id, err := cl.ReadInt()
					if err == nil {
						compJs, err := cl.ReadArray()
						if err == nil {

							var msg *AppsRouterMsg
							router.lock.Lock()
							{
								msg = router.msgs[msg_id]
							}
							router.lock.Unlock()

							//exe
							if msg != nil {
								var comp LLMTranscribe
								err := LogsJsonUnmarshal(compJs, &comp)
								if err == nil {

									err = router.services.llms.Transcribe(&comp)
									if err == nil {
										//save back
										compJs, _ = LogsJsonMarshal(&comp)
									}
								}
							}

							//send back
							cl.WriteArray(nil) //empty delta
							cl.WriteArray(compJs)
						}
					}
				}
			}
		}()
	}
}

func (router *AppsRouter) _reloadAppList() {
	files, err := os.ReadDir("apps")
	if LogsError(err) != nil {
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
			if err == nil {
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

func (router *AppsRouter) _hotReload() {

	router._reloadAppList()

	//ticks
	for _, app := range router.apps {
		app.Tick(false)
	}
}
