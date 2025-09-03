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
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type AppsRouterMsg struct {
	msg_id   uint64
	msg_name string

	actionName string

	start_time float64

	fnProgress func(cmdsGob [][]byte, err error, start_time float64)                            //call when 'out_flushed_cmds'
	fnDone     func(dataJs []byte, uiGob []byte, cmdsGob []byte, err error, start_time float64) //called when msg is done

	stop     atomic.Bool
	out_done atomic.Bool

	progress_done  float64
	progress_label string

	out_dataJs  []byte
	out_uiGob   []byte
	out_cmdsGob []byte

	out_flushed_cmdsGob [][]byte
	out_error           error

	drawit bool
}

func NewAppsRouterMsg(msg_id uint64, msg_name string, actionName string,
	fnProgress func(cmdsGob [][]byte, err error, start_time float64),
	fnDone func(dataJs []byte, uiGob []byte, cmdsGob []byte, err error, start_time float64)) *AppsRouterMsg {

	msg := &AppsRouterMsg{msg_id: msg_id, msg_name: msg_name, fnProgress: fnProgress, fnDone: fnDone}

	msg.start_time = Tools_Time()
	msg.actionName = actionName

	return msg
}

func (msg *AppsRouterMsg) SetName(appName, toolName, msgName string) {
	msg.msg_name = fmt.Sprintf("%s_%s_%s", appName, toolName, msgName)
}

func (msg *AppsRouterMsg) HasName(appName, toolName, msgName string) bool {
	return msg.msg_name == fmt.Sprintf("%s_%s_%s", appName, toolName, msgName)
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

	msgs           map[uint64]*AppsRouterMsg
	msgs_counter   atomic.Uint64 //to create unique msg_id
	text_highlight string

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

func (router *AppsRouter) GetNumMsgs() int {
	return len(router.msgs)
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

func (router *AppsRouter) CallUpdateAsync(sub_uid uint64, fnProgress func(cmdsGob [][]byte, err error, start_time float64), fnDone func(dataJs []byte, uiGob []byte, cmdsGob []byte, err error, start_time float64)) {
	ui_uid := uint64(1)

	app := router.FindApp("Root")
	if app == nil {
		return
	}

	msg_id := router.msgs_counter.Add(1)
	msg := NewAppsRouterMsg(msg_id, "", "update", fnProgress, fnDone)

	router.lock.Lock()
	router.msgs[msg_id] = msg
	router.lock.Unlock()

	//call
	go func() {
		defer msg.Done()

		out_subUiGob, out_cmdsGob, out_error := _ToolsCaller_CallUpdate(app.Process.port, msg_id, ui_uid, sub_uid)
		msg.out_error = out_error

		if out_error == nil {
			msg.out_uiGob = out_subUiGob
			msg.out_cmdsGob = out_cmdsGob
		}

	}()
}

func (router *AppsRouter) CallChangeAsync(change ToolsSdkChange, fnProgress func(cmdsGob [][]byte, err error, start_time float64), fnDone func(dataJs []byte, uiGob []byte, cmdsGob []byte, err error, start_time float64)) {
	ui_uid := uint64(1)

	app := router.FindApp("Root")
	if app == nil {
		return
	}

	msg_id := router.msgs_counter.Add(1)
	msg := NewAppsRouterMsg(msg_id, "", "change", fnProgress, fnDone)

	msg.drawit = true

	router.lock.Lock()
	router.msgs[msg_id] = msg
	router.lock.Unlock()

	//call
	go func() {
		defer msg.Done()

		out_dataJs, out_cmdsGob, out_error := _ToolsCaller_CallChange(app.Process.port, msg_id, ui_uid, change)
		msg.out_error = out_error

		if out_error == nil {
			msg.out_dataJs = out_dataJs
			msg.out_cmdsGob = out_cmdsGob
		}

	}()
}

func (router *AppsRouter) CallBuildAsync(ui_uid uint64, appName, toolName string, params interface{}, fnProgress func(cmdsGob [][]byte, err error, start_time float64), fnDone func(dataJs []byte, uiGob []byte, cmdsGob []byte, err error, start_time float64)) *AppsRouterMsg {
	app := router.FindApp(appName)
	if app == nil {
		return nil
	}

	msg_id := router.msgs_counter.Add(1)
	msg := NewAppsRouterMsg(msg_id, "", "build", fnProgress, fnDone)

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
		msg.out_dataJs, msg.out_uiGob, msg.out_cmdsGob, msg.out_error = _ToolsCaller_CallBuild(app.Process.port, msg_id, ui_uid, toolName, jsParams)
	}()

	return msg
}

func (router *AppsRouter) AddLocalRecompileMsg(appName string) *AppsRouterMsg {
	msg_id := router.msgs_counter.Add(1)
	msg := NewAppsRouterMsg(msg_id, "", "compile", nil, nil)
	msg.SetName(appName, "", "_compile_")

	router.lock.Lock()
	router.msgs[msg_id] = msg
	router.lock.Unlock()

	return msg
}

func (router *AppsRouter) FindLocalRecompileMsg(appName string) *AppsRouterMsg {
	return router.FindMessageName(appName, "", "_compile_")
}

func (router *AppsRouter) FindMessageName(appName, toolName, msgName string) *AppsRouterMsg {
	router.lock.Lock()
	defer router.lock.Unlock()

	for _, msg := range router.msgs {
		if msg != nil && msg.HasName(appName, toolName, msgName) {
			return msg
		}
	}
	return nil
}

func (router *AppsRouter) Flush() bool {
	refresh := false

	var doneMsgs []*AppsRouterMsg
	router.lock.Lock()
	{
		//refresh progress/threads with delay
		{
			TM := Tools_Time()
			msg_refresh := false
			for _, msg := range router.msgs {
				if (TM-msg.start_time) > 0.1 && msg.actionName != "build" { //give msg 0.1sec to finish, then start refreshing(below)
					msg.drawit = true
					msg_refresh = true
				}
			}

			if msg_refresh && router.refresh_progress_time < TM {
				refresh = true
				router.refresh_progress_time = TM + 1 //next is 1 second later
			}
		}

		//flush cmds
		for _, msg := range router.msgs {
			if msg != nil && len(msg.out_flushed_cmdsGob) > 0 {
				msg.fnProgress(msg.out_flushed_cmdsGob, nil, msg.start_time)
				msg.out_flushed_cmdsGob = nil
			}
		}

		//done messages
		for msg_id, msg := range router.msgs {
			if msg != nil && msg.out_done.Load() {

				if msg.drawit {
					refresh = true
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
		msg.fnDone(msg.out_dataJs, msg.out_uiGob, msg.out_cmdsGob, msg.out_error, msg.start_time)
	}

	return refresh
}

func (router *AppsRouter) RunNet() {

	type SdkMsg struct {
		Id         string
		ActionName string

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
					msg_id, err := cl.ReadInt()
					if err == nil {
						appName, err := cl.ReadArray()
						if err == nil {

							router.lock.Lock()
							msg, _ := router.msgs[msg_id]
							router.lock.Unlock()

							var retErr error
							app := router.FindApp(string(appName))
							if app != nil {
								retErr = app.Tick(msg) //err ....
							} else {
								retErr = fmt.Errorf("app '%s' not found", string(appName))
							}

							var errStr string
							if retErr != nil {
								errStr = retErr.Error()
							}
							cl.WriteArray([]byte(errStr))
						}
					}

				case "get_llm_usage":
					usage := router.services.llms.GetUsage()

					usageJs, _ := LogsJsonMarshal(usage)

					cl.WriteArray(usageJs)

				case "rename_app":
					oldNameBytes, err := cl.ReadArray()
					if err == nil {
						newNameBytes, err := cl.ReadArray()
						if err == nil {
							app := router.FindApp(string(oldNameBytes))
							if app != nil {
								if string(oldNameBytes) != string(newNameBytes) {
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
								} else {
									cl.WriteArray(newNameBytes)
									cl.WriteArray(nil) //err
								}

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

				case "run_app": //and run app
					appName, err := cl.ReadArray()
					if err == nil {
						var out_Error error
						app := router.FindApp(string(appName))
						var app_port uint64
						if app != nil {
							//start it
							out_Error = app.CheckRun()

							//get port
							if out_Error == nil {
								app_port = uint64(app.Process.port)
							}

						} else {
							out_Error = LogsErrorf("app '%s' not found", string(appName))
						}

						errStr := ""
						if out_Error != nil {
							errStr = out_Error.Error()
						}

						cl.WriteInt(app_port)
						cl.WriteArray([]byte(errStr))
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
									msg.out_flushed_cmdsGob = append(msg.out_flushed_cmdsGob, cmdsJs)
								}
							}
							router.lock.Unlock()
						}
					}

				case "get_msgs":
					var final_msgs []SdkMsg
					sorted_msgs := router.GetSortedMsgs()
					for _, msg := range sorted_msgs {
						if msg != nil && msg.drawit {

							final_msgs = append(final_msgs, SdkMsg{Id: msg.msg_name,
								ActionName:     msg.actionName,
								Progress_label: msg.progress_label, Progress_done: msg.progress_done,
								Start_time: msg.start_time})
						}
					}

					msgsJs, _ := LogsJsonMarshal(final_msgs)
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

				case "set_text_highlight":
					text, err := cl.ReadArray()
					if err == nil {
						router.text_highlight = string(text)
					}

				case "set_msg_name":
					appName, err := cl.ReadArray()
					if err == nil {
						toolName, err := cl.ReadArray()
						if err == nil {
							msg_name, err := cl.ReadArray()
							if err == nil {
								msg_id, err := cl.ReadInt()
								if err == nil {
									router.lock.Lock()
									{
										msg, found := router.msgs[msg_id]
										if found && msg != nil {
											msg.SetName(string(appName), string(toolName), string(msg_name))
										}
									}
									router.lock.Unlock()
								}
							}
						}
					}

				case "find_msg":
					appName, err := cl.ReadArray()
					if err == nil {
						toolName, err := cl.ReadArray()
						if err == nil {
							msg_name, err := cl.ReadArray()
							if err == nil {
								msg := router.FindMessageName(string(appName), string(toolName), string(msg_name))

								if msg != nil {
									cl.WriteInt(1) //exist

									msg := SdkMsg{Id: string(msg_name),
										ActionName:     msg.actionName,
										Progress_label: msg.progress_label, Progress_done: msg.progress_done,
										Start_time: msg.start_time}
									msgJs, _ := LogsJsonMarshal(msg)
									cl.WriteArray(msgJs)

								} else {
									cl.WriteInt(0) //non-exist
								}
							}
						}
					}

				case "stop_msg":
					appName, err := cl.ReadArray()
					if err == nil {
						toolName, err := cl.ReadArray()
						if err == nil {
							msg_name, err := cl.ReadArray()
							if err == nil {
								msg := router.FindMessageName(string(appName), string(toolName), string(msg_name))
								if msg != nil {
									msg.Stop()
								}
							}
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

				case "llm_find":
					appName, err := cl.ReadArray()
					if err == nil {
						toolName, err := cl.ReadArray()
						if err == nil {
							llm_name, err := cl.ReadArray()
							if err == nil {
								comp := router.services.llms.Find(string(appName), string(toolName), string(llm_name))
								cl.WriteInt(uint64(OsTrn(comp != nil, 1, 0)))
								if comp != nil {
									cl.WriteArray([]byte(comp.wip_answer))
								} else {
									cl.WriteArray(nil)
								}

							}
						}
					}

				case "llm_stop":
					appName, err := cl.ReadArray()
					if err == nil {
						toolName, err := cl.ReadArray()
						if err == nil {
							llm_name, err := cl.ReadArray()
							if err == nil {
								comp := router.services.llms.Find(string(appName), string(toolName), string(llm_name))
								if comp != nil {
									if comp.msg != nil {
										comp.msg.Stop()
									} else {
										LogsErrorf("it.msg == nil")
									}
								}
							}
						}
					}

				case "llm_complete":
					msg_id, err := cl.ReadInt()
					if err == nil {
						compJs, err := cl.ReadArray()
						if err == nil {
							msg := router.FindMsg(msg_id)
							if msg != nil {
								var comp LLMComplete
								err := LogsJsonUnmarshal(compJs, &comp)
								if err == nil {

									comp.delta = func(msg *ChatMsg) {

										if msg.Content.Calls != nil && msg.Content.Calls.Content != "" {
											comp.wip_answer = msg.Content.Calls.Content
											comp.wip_answer, _ = strings.CutSuffix(comp.wip_answer, ChatMsg_GetDivAfterReasoning()) //cut reasoning divider
										}
										msgJs, _ := LogsJsonMarshal(msg)

										cl.WriteArray(msgJs) //send delta raw
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
							msg := router.FindMsg(msg_id)
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

func (router *AppsRouter) FindMsg(msg_id uint64) *AppsRouterMsg {
	router.lock.Lock()
	defer router.lock.Unlock()
	return router.msgs[msg_id]
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

	var wg sync.WaitGroup

	//tick Root app first
	rootApp := router.FindApp("Root")
	if rootApp != nil {
		wg.Add(1)
		go func() {
			rootApp.Tick(nil)
			wg.Done()
		}()
	}

	//ticks
	for _, app := range router.apps {
		if app == rootApp {
			continue
		}

		wg.Add(1)
		go func() {
			app.Tick(nil)
			wg.Done()
		}()
	}

	wg.Wait()
}
