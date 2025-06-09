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
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"image/color"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func _loadInstance[T any](file string, structName string, format string, defInst *T, save bool) (*T, error) {
	if file == "" {
		file = fmt.Sprintf("%s-%s.%s", structName, structName, format)
	}

	//find
	g_files_lock.Lock()
	inst, found := g_files[file]
	g_files_lock.Unlock()
	if found {
		inst.save = save
		return inst.st.(*T), nil
	}

	//get file data
	data, err := os.ReadFile(file)
	if err != nil {
		//is file exist
		if _, err := os.Stat(file); err == nil {
			Tool_Error(err)
		}
	}

	// Unpack
	if len(data) > 0 {
		if format == "json" {
			err := json.Unmarshal(data, defInst)
			if err != nil {
				return nil, err
			}
		} else if format == "xml" {
			err := xml.Unmarshal(data, defInst)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("%s format not supported", format)
		}
	}

	g_files_lock.Lock()
	g_files[file] = &_Instance{data: data, st: defInst, save: save}
	g_files_lock.Unlock()
	return defInst, nil
}

func _saveInstances() {

	g_files_lock.Lock()
	defer g_files_lock.Unlock()

	changed := false
	for path, it := range g_files {
		if !it.save {
			continue
		}

		var err error
		var js []byte
		switch strings.ToLower(filepath.Ext(path)) {
		case ".json":
			js, err = json.Marshal(it.st)
		case ".xml":
			js, err = xml.Marshal(it.st)
		}

		if err == nil && !bytes.Equal(it.data, js) {

			//create folder
			os.MkdirAll(filepath.Dir(path), os.ModePerm)

			//save file
			os.WriteFile(path, js, 0644)

			it.data = js

			changed = true
		}
	}

	if changed {
		cl, err := NewToolClient("localhost", g_main.router_port)
		if Tool_Error(err) == nil {
			defer cl.Destroy()

			err = cl.WriteArray([]byte("storage_changed"))
			if Tool_Error(err) == nil {
				err = cl.WriteArray([]byte(g_main.appName))
				Tool_Error(err)
			}
		}
	}
}

type SdkPalette struct {
	P, S, E, B         color.RGBA
	OnP, OnS, OnE, OnB color.RGBA
}

func Color_Aprox(s color.RGBA, e color.RGBA, t float64) color.RGBA {
	var self color.RGBA
	self.R = byte(float64(s.R) + (float64(e.R)-float64(s.R))*t)
	self.G = byte(float64(s.G) + (float64(e.G)-float64(s.G))*t)
	self.B = byte(float64(s.B) + (float64(e.B)-float64(s.B))*t)
	self.A = byte(float64(s.A) + (float64(e.A)-float64(s.A))*t)
	return self
}

func (pl *SdkPalette) GetGrey(t float64) color.RGBA {
	return Color_Aprox(pl.B, pl.OnB, t)
}

func UI_GetPalette() *SdkPalette {
	return &g_dev.Palette
}
func UI_GetDateFormat() string {
	return g_dev.DateFormat
}

type ToolProgram struct {
	appName     string
	router_port int
	server      *ToolServer
}

type _Instance struct {
	data []byte
	st   any
	save bool
}

type ToolDeviceSettings struct {
	Palette    SdkPalette
	DateFormat string
}

type ToolCaller struct {
	msg_id uint64
	ui_uid uint64

	last_send_progress_ms int64 //ms

	cmds []ToolCmd
}

func NewToolCaller() *ToolCaller {
	caller := &ToolCaller{}
	return caller
}

type ToolUI struct {
	parameters interface{}
	ui         *UI

	Caller *ToolCaller

	lock sync.Mutex
}

var g_uis_lock sync.Mutex
var g_uis map[uint64]*ToolUI //[ui_uid]

var g_files map[string]*_Instance
var g_files_lock sync.Mutex

var g_main ToolProgram

var g_dev ToolDeviceSettings

func main() {
	log.SetFlags(log.Llongfile) //log.LstdFlags | log.Lshortfile

	if len(os.Args) < 3 {
		log.Fatal("missing 'app name' and 'port' argument(s): ", os.Args)
	}

	g_main.appName = os.Args[1]

	var err error
	g_main.router_port, err = strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	g_main.server = NewToolServer(g_main.router_port + 100)

	g_uis = make(map[uint64]*ToolUI)
	g_files = make(map[string]*_Instance)

	defer g_main.server.Destroy()

	//report tool into router server
	{
		cl, err := NewToolClient("localhost", g_main.router_port)
		if err != nil {
			log.Fatal(err)
		}
		err = cl.WriteArray([]byte("register"))
		if err != nil {
			log.Fatal(err)
		}

		err = cl.WriteArray([]byte(g_main.appName))
		if err != nil {
			log.Fatal(err)
		}

		err = cl.WriteInt(uint64(g_main.server.port))
		if err != nil {
			log.Fatal(err)
		}
		cl.Destroy()
	}

	//init functions
	_callGlobalInits()
	defer _callGlobalDestroys()

	_updateDev()

	// main loop
	for {
		cl, err := g_main.server.Accept()
		if Tool_Error(err) != nil {
			continue
		}
		if cl == nil {
			break //close tool
		}

		mode, err := cl.ReadArray()
		if Tool_Error(err) == nil {

			switch string(mode) {

			case "exit":
				cl.Destroy()
				return

			case "update_dev":
				_updateDev()
				cl.Destroy()

			case "change":
				ok := false

				msg_id, err := cl.ReadInt()
				if Tool_Error(err) == nil {
					ui_uid, err := cl.ReadInt()
					if Tool_Error(err) == nil {
						changeJs, err := cl.ReadArray()
						if Tool_Error(err) == nil {
							var change SdkChange
							err = json.Unmarshal(changeJs, &change)
							if Tool_Error(err) == nil {
								g_uis_lock.Lock()
								ui, found := g_uis[ui_uid]
								g_uis_lock.Unlock()
								if found {
									ok = true
									go func() {
										ui.lock.Lock()
										defer ui.lock.Unlock()

										ui.Caller.msg_id = msg_id
										ui.Caller.cmds = nil

										out_error := ui.ui.runChange(change)

										if out_error == nil {
											if !ui.Caller._sendProgress(1, "") {
												out_error = errors.New("_interrupted_")
											}
										}

										//add callstack to error
										var output_errBytes []byte
										if out_error != nil {
											output_errBytes = []byte(out_error.Error() + fmt.Sprintf("\n%s(%.20s)", "_change_", string(changeJs)))
										}

										//send back
										{
											err = cl.WriteArray(output_errBytes) //error
											Tool_Error(err)

											dataJs, _ := json.Marshal(ui.parameters)
											err = cl.WriteArray(dataJs) //data
											Tool_Error(err)

											cmdsJs, _ := json.Marshal(ui.Caller.cmds)
											err = cl.WriteArray(cmdsJs) //commands
											Tool_Error(err)
										}
										cl.Destroy()

										_saveInstances()
									}()
								} else {
									fmt.Printf("UI UID %d not found\n", ui_uid)
								}
							}
						}
					}
				}
				if !ok {
					cl.Destroy()
				}

			case "call":
				ok := false
				caller := NewToolCaller()

				caller.msg_id, err = cl.ReadInt()
				if Tool_Error(err) == nil {
					caller.ui_uid, err = cl.ReadInt()
					if Tool_Error(err) == nil {
						funcName, err := cl.ReadArray()
						if Tool_Error(err) == nil {

							paramsJs, err := cl.ReadArray()
							if Tool_Error(err) == nil {
								ok = true
								go func() {
									ui := _newUIItem(0, 0, 1, 1)
									ui.UID = caller.ui_uid

									if len(paramsJs) == 0 {
										paramsJs = []byte("{}")
									}

									fnRun, out_params, err := FindToolRunFunc(string(funcName), paramsJs)
									out_error := err
									if Tool_Error(out_error) == nil {
										if fnRun != nil {
											out_error = fnRun(caller, ui)
										}
									}

									if out_error == nil {
										if caller.ui_uid != 0 {
											g_uis_lock.Lock()
											g_uis[caller.ui_uid] = &ToolUI{ui: ui,
												parameters: out_params,
												Caller:     caller}
											g_uis_lock.Unlock()
										}

										if !caller._sendProgress(1, "") {
											out_error = errors.New("_interrupted_")
										}
									}

									//out -> bytes
									var dataJs []byte
									var uiJs []byte
									var cmdsJs []byte
									if out_error == nil {
										dataJs, out_error = json.Marshal(out_params)
										Tool_Error(out_error)

										uiJs, out_error = json.Marshal(ui)
										Tool_Error(out_error)

										cmdsJs, out_error = json.Marshal(caller.cmds)
										Tool_Error(out_error)

										caller.cmds = nil
									}

									//add callstack to error
									var output_errBytes []byte
									if out_error != nil {
										output_errBytes = []byte(out_error.Error() + fmt.Sprintf("\n%s(%.20s)", funcName, string(paramsJs)))
									}

									//send result back
									err = cl.WriteArray(output_errBytes) //error
									if Tool_Error(err) == nil {
										err = cl.WriteArray(dataJs) //data
										if Tool_Error(err) == nil {
											err = cl.WriteArray(uiJs) //ui
											if Tool_Error(err) == nil {
												err = cl.WriteArray(cmdsJs) //cmds
												Tool_Error(err)
											}
										}
									}

									cl.Destroy()

									_saveInstances()
								}()
							}
						}
					}
				}
				if !ok {
					cl.Destroy()
				}
			}
		}
	}
}

func Tool_Error(err error) error {
	if err != nil {
		stack := string(debug.Stack())

		str := fmt.Sprintf("\033[31merror: %v\nstack:%s\033[0m\n", err, stack)
		fmt.Println(str)
		callFuncPrint(str)
	}
	return err
}

// returns true=OK, false=interrupted
func (caller *ToolCaller) Progress(done float64, label string) bool {
	st := time.Now().UnixMilli()
	if (st - caller.last_send_progress_ms) < 500 {
		return true //ok
	}
	caller.last_send_progress_ms = st

	return caller._sendProgress(done, label)
}

func (caller *ToolCaller) _sendProgress(done float64, label string) bool {
	cl, err := NewToolClient("localhost", g_main.router_port)
	if Tool_Error(err) == nil {
		defer cl.Destroy()

		err = cl.WriteArray([]byte("progress"))
		if Tool_Error(err) == nil {
			err = cl.WriteInt(caller.msg_id)
			if Tool_Error(err) == nil {
				err = cl.WriteInt(uint64(done * 10000))
				if Tool_Error(err) == nil {
					err = cl.WriteArray([]byte(label))
					if Tool_Error(err) == nil {
						stop, err := cl.ReadInt()
						if Tool_Error(err) == nil {
							return stop == 0
						}
					}
				}
			}
		}
	}

	return false //stop
}

func (caller *ToolCaller) SendFlushCmd() {
	cmdsJs, err := json.Marshal(caller.cmds)
	if Tool_Error(err) != nil {
		return
	}

	cl, err := NewToolClient("localhost", g_main.router_port)
	if Tool_Error(err) == nil {
		defer cl.Destroy()

		err = cl.WriteArray([]byte("add_cmds"))
		if Tool_Error(err) == nil {
			err = cl.WriteInt(caller.msg_id)
			if Tool_Error(err) == nil {
				err = cl.WriteArray(cmdsJs)
				if Tool_Error(err) == nil {
					caller.cmds = nil //reset
				}
			}
		}
	}
}

func _updateDev() {
	js, err := os.ReadFile("../Device/DeviceSettings-DeviceSettings.json")
	if Tool_Error(err) != nil {
		return
	}

	type UiSyncDeviceSettings struct {
		DateFormat string

		Theme         string //light, dark, custom
		LightPalette  SdkPalette
		DarkPalette   SdkPalette
		CustomPalette SdkPalette
	}
	var st UiSyncDeviceSettings

	err = json.Unmarshal(js, &st)
	if Tool_Error(err) != nil {
		return
	}

	g_dev.DateFormat = st.DateFormat

	switch st.Theme {
	case "light":
		g_dev.Palette = st.LightPalette
	case "dark":
		g_dev.Palette = st.DarkPalette
	default:
		g_dev.Palette = st.CustomPalette
	}
}

type SdkTool struct {
	Name  string
	Attrs []string

	Running bool

	Compile_error string
	Cmd_error     string
}

type SdkMsg struct {
	Id             string
	AppName        string
	FuncName       string
	Progress_label string
	Progress_done  float64
	Start_time     float64
}

func (msg *SdkMsg) GetLabel() string {
	label := msg.Progress_label
	if label == "" {
		label = fmt.Sprintf("%s:%s()", msg.AppName, msg.FuncName)
	}

	if msg.Progress_done > 0 {
		//Percentage
		label += fmt.Sprintf(" - %.2f%%", msg.Progress_done*100)
	} else {
		//Time
		dt := time.Since(time.Unix(int64(msg.Start_time), 0))
		label += fmt.Sprintf("- %d:%02d:%02d", int(dt.Hours()), int(dt.Minutes())%60, int(dt.Seconds())%60)
	}
	return label
}

func GetMsgs() []SdkMsg {
	var msgs []SdkMsg

	cl, err := NewToolClient("localhost", g_main.router_port)
	if Tool_Error(err) == nil {
		defer cl.Destroy()

		err = cl.WriteArray([]byte("get_msgs"))
		if Tool_Error(err) == nil {
			msgsJs, err := cl.ReadArray()
			if Tool_Error(err) == nil {
				err = json.Unmarshal(msgsJs, &msgs)
				Tool_Error(err)
			}
		}
	}

	return msgs
}

func callFuncMsgStop(msg_id string) {
	cl, err := NewToolClient("localhost", g_main.router_port)
	if Tool_Error(err) == nil {
		defer cl.Destroy()

		err = cl.WriteArray([]byte("stop_msg"))
		if Tool_Error(err) == nil {
			err = cl.WriteArray([]byte(msg_id))
			Tool_Error(err)
		}
	}
}
func callFuncFindMsgName(name string) *SdkMsg {
	cl, err := NewToolClient("localhost", g_main.router_port)
	if Tool_Error(err) == nil {
		defer cl.Destroy()

		err = cl.WriteArray([]byte("find_msg"))
		if Tool_Error(err) == nil {
			err = cl.WriteArray([]byte(name))
			if Tool_Error(err) == nil {

				exist, err := cl.ReadInt()
				if exist > 0 && Tool_Error(err) == nil {
					msgJs, err := cl.ReadArray()
					if Tool_Error(err) == nil {
						var msg SdkMsg
						err = json.Unmarshal(msgJs, &msg)
						if Tool_Error(err) == nil {
							return &msg
						}
					}
				}
			}
		}
	}
	return nil
}

func (caller *ToolCaller) SetMsgName(name string) {
	cl, err := NewToolClient("localhost", g_main.router_port)
	if Tool_Error(err) == nil {
		defer cl.Destroy()

		err = cl.WriteArray([]byte("set_msg_name"))
		if Tool_Error(err) == nil {

			err = cl.WriteInt(caller.msg_id)
			if Tool_Error(err) == nil {
				err = cl.WriteArray([]byte(name))
				Tool_Error(err)
			}
		}
	}
}

func callFuncGetToolsShemas(appName string) []byte {
	cl, err := NewToolClient("localhost", g_main.router_port)
	if Tool_Error(err) == nil {
		defer cl.Destroy()

		err = cl.WriteArray([]byte("get_tools_shemas"))
		if Tool_Error(err) == nil {

			err = cl.WriteArray([]byte(appName))
			if Tool_Error(err) == nil {
				js, err := cl.ReadArray()
				if Tool_Error(err) == nil {
					return js
				}
			}
		}
	}
	return nil
}

func callFuncGetToolData(appName string) ([]byte, error) {
	cl, err := NewToolClient("localhost", g_main.router_port)
	if Tool_Error(err) == nil {
		defer cl.Destroy()

		err = cl.WriteArray([]byte("get_tool_data"))
		if Tool_Error(err) == nil {

			err = cl.WriteArray([]byte(appName))
			if Tool_Error(err) == nil {
				promptsJs, err := cl.ReadArray()
				if Tool_Error(err) == nil {
					return promptsJs, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("connection failed")
}

func (caller *ToolCaller) callFuncSubCall(ui_uid uint64, appName string, funcName string, jsParams []byte) ([]byte, []byte, error) {
	cl, err := NewToolClient("localhost", g_main.router_port)
	if Tool_Error(err) == nil {
		defer cl.Destroy()

		err = cl.WriteArray([]byte("sub_call"))
		if Tool_Error(err) == nil {

			err = cl.WriteInt(caller.msg_id)
			if Tool_Error(err) == nil {
				err = cl.WriteInt(ui_uid)
				if Tool_Error(err) == nil {
					err = cl.WriteArray([]byte(appName))
					if Tool_Error(err) == nil {
						err = cl.WriteArray([]byte(funcName))
						if Tool_Error(err) == nil {
							if len(jsParams) == 0 {
								jsParams = []byte("{}")
							}
							err = cl.WriteArray(jsParams)
							if Tool_Error(err) == nil {

								dataJs, err := cl.ReadArray()
								if Tool_Error(err) == nil {
									uiJs, err := cl.ReadArray()
									if Tool_Error(err) == nil {
										cmdsJs, err := cl.ReadArray()
										if Tool_Error(err) == nil {
											errBytes, err := cl.ReadArray()
											if Tool_Error(err) == nil {
												if len(errBytes) > 0 {
													err = errors.New(string(errBytes))
												}

												//add cmds
												var cmds []ToolCmd
												err = json.Unmarshal(cmdsJs, &cmds)
												if Tool_Error(err) == nil {
													caller.cmds = append(caller.cmds, cmds...)
												}

												return dataJs, uiJs, err
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return nil, nil, fmt.Errorf("connection failed")
}

func callFuncPrint(str string) {
	cl, err := NewToolClient("localhost", g_main.router_port)
	if Tool_Error(err) == nil {
		defer cl.Destroy()

		err = cl.WriteArray([]byte("print"))
		if Tool_Error(err) == nil {
			err = cl.WriteArray([]byte(str))
			Tool_Error(err)
		}
	}
}

func (caller *ToolCaller) _addCmd(cmd ToolCmd) {
	caller.cmds = append(caller.cmds, cmd)
}

func _newUIItem(x, y, w, h int) *UI {
	item := &UI{X: x, Y: y, W: w, H: h}

	item.Enable = true
	item.EnableTouch = true

	return item
}

func (ui *UI) Is() bool {
	return len(ui.Items) > 0
}

func (ui *UI) _computeUID(parent *UI, name string) {
	h := sha256.New()

	//parent
	if parent != nil {
		var pt [8]byte
		binary.LittleEndian.PutUint64(pt[:], parent.UID)
		h.Write(pt[:])
	}

	//this
	h.Write([]byte(fmt.Sprintf("%s: %d,%d,%d,%d", name, ui.X, ui.Y, ui.W, ui.H)))

	ui.UID = binary.LittleEndian.Uint64(h.Sum(nil))
}

func (parent *UI) _addUISub(ui *UI, name string) {
	parent.Items = append(parent.Items, ui)
	ui._computeUID(parent, name)
}

func _callTool(funcName string, fnRun func(caller *ToolCaller, ui *UI) error, caller *ToolCaller) (*UI, error) {
	if fnRun != nil {
		ui := &UI{}
		out_error := fnRun(caller, ui)

		if out_error == nil {
			if !caller._sendProgress(1, "") {
				out_error = errors.New("_interrupted_")
			}
		}

		return ui, out_error
	}
	return nil, fmt.Errorf("Tool function '%s' not found", funcName)
}

func CallTool(fnRun func(caller *ToolCaller, ui *UI) error, caller *ToolCaller) (*UI, error) {
	return _callTool("", fnRun, caller)
}

func CallToolApp(appName string, funcName string, jsParams []byte, caller *ToolCaller) ([]byte, *UI, error) {
	dataJs, uiJs, err := caller.callFuncSubCall(0, appName, funcName, jsParams)

	var ui UI
	if err == nil {
		err = json.Unmarshal(uiJs, &ui)
		Tool_Error(err)
	}

	return dataJs, &ui, err
}

func (ui *UI) _findUID(uid uint64) *UI {
	if ui.UID == uid {
		return ui
	}

	//dialogs
	for _, dia := range ui.Dialogs {
		f := dia.UI._findUID(uid)
		if f != nil {
			return f
		}
	}

	//subs
	for _, it := range ui.Items {
		f := it._findUID(uid)
		if f != nil {
			return f
		}
	}

	return nil
}

type SdkChange struct {
	UID         uint64
	ValueBytes  []byte
	ValueString string
	ValueFloat  float64
	ValueInt    int64
	ValueBool   bool
}

func (ui *UI) runChange(change SdkChange) error {
	it := ui._findUID(change.UID)
	if it == nil {
		return fmt.Errorf("UID %d not found", change.UID)
	}

	//sub-app
	if it.changed != nil {
		return it.changed(change.ValueBytes)
	}

	if it.Text != nil {
		if it.Text.dropFile != nil {
			var pathes []string
			err := json.Unmarshal(change.ValueBytes, &pathes)
			if err != nil {
				return err
			}
			return it.Text.dropFile(pathes)
		}
	}

	if it.Editbox != nil {

		diff := false
		if it.Editbox.Value != nil {
			diff = (*it.Editbox.Value != change.ValueString)
			*it.Editbox.Value = change.ValueString
		}
		if it.Editbox.ValueInt != nil {
			diff = (*it.Editbox.ValueInt != int(change.ValueInt))
			*it.Editbox.ValueInt = int(change.ValueInt)
		}
		if it.Editbox.ValueFloat != nil {
			diff = (*it.Editbox.ValueFloat != change.ValueFloat)
			*it.Editbox.ValueFloat = change.ValueFloat
		}

		if diff && it.Editbox.changed != nil {
			return it.Editbox.changed()
		}
		if change.ValueBool && it.Editbox.enter != nil {
			return it.Editbox.enter()
		}
	}
	if it.Slider != nil {
		if it.Slider.Value != nil {
			*it.Slider.Value = change.ValueFloat
		}
		if it.Slider.changed != nil {
			return it.Slider.changed()
		}
	}
	if it.Combo != nil {
		if it.Combo.Value != nil {
			*it.Combo.Value = change.ValueString
		}
		if it.Combo.changed != nil {
			return it.Combo.changed()
		}
	}
	if it.Switch != nil {
		if it.Switch.Value != nil {
			*it.Switch.Value = change.ValueBool
		}
		if it.Switch.changed != nil {
			return it.Switch.changed()
		}
	}
	if it.Checkbox != nil {
		if it.Checkbox.Value != nil {
			*it.Checkbox.Value = change.ValueFloat
		}
		if it.Checkbox.changed != nil {
			return it.Checkbox.changed()
		}
	}
	if it.FilePickerButton != nil {
		if it.FilePickerButton.Path != nil {
			*it.FilePickerButton.Path = change.ValueString
		}
		if it.FilePickerButton.changed != nil {
			return it.FilePickerButton.changed()
		}
	}
	if it.DatePickerButton != nil {
		if it.DatePickerButton.Date != nil {
			*it.DatePickerButton.Date = change.ValueInt
		}
		if it.DatePickerButton.changed != nil {
			return it.DatePickerButton.changed()
		}
	}
	if it.ColorPickerButton != nil {
		if it.ColorPickerButton.Cd != nil {
			fmt.Sscanf(change.ValueString, "%d %d %d %d", &it.ColorPickerButton.Cd.R, &it.ColorPickerButton.Cd.G, &it.ColorPickerButton.Cd.B, &it.ColorPickerButton.Cd.A)
		}

		if it.ColorPickerButton.changed != nil {
			return it.ColorPickerButton.changed()
		}
	}

	if it.OsmMap != nil {
		fmt.Sscanf(change.ValueString, "%f %f %f", it.OsmMap.Lon, it.OsmMap.Lat, it.OsmMap.Zoom)
	}

	if it.Button != nil {
		if change.ValueString != "" {
			if it.Button.dropMove != nil {
				var src_i int
				var dst_i int
				var src_source string
				var dst_source string
				fmt.Sscanf(change.ValueString, "%d %d %s %s", &src_i, &dst_i, &src_source, &dst_source)

				return it.Button.dropMove(src_i, dst_i, src_source, dst_source)
			}
		} else {
			if it.Button.clicked != nil {
				return it.Button.clicked()
			}
		}
	}

	return nil
}

func Layout_GetMonthText(month int) string {
	switch month {
	case 1:
		return "January"
	case 2:
		return "February"
	case 3:
		return "March"
	case 4:
		return "April"
	case 5:
		return "May"
	case 6:
		return "June"
	case 7:
		return "July"
	case 8:
		return "August"
	case 9:
		return "September"
	case 10:
		return "October"
	case 11:
		return "November"
	case 12:
		return "December"
	}
	return ""
}

func Layout_GetDayTextFull(day int) string {
	switch day {
	case 1:
		return "Monday"
	case 2:
		return "Tuesday"
	case 3:
		return "Wednesday"
	case 4:
		return "Thursday"
	case 5:
		return "Friday"
	case 6:
		return "Saturday"
	case 7:
		return "Sunday"
	}
	return ""
}

func Layout_GetDayTextShort(day int) string {
	switch day {
	case 1:
		return "Mon"
	case 2:
		return "Tue"
	case 3:
		return "Wed"
	case 4:
		return "Thu"
	case 5:
		return "Fri"
	case 6:
		return "Sat"
	case 7:
		return "Sun"
	}
	return ""
}

func Layout_ConvertTextTime(unix_sec int64) string {
	tm := time.Unix(unix_sec, 0)
	return fmt.Sprintf("%.02d:%.02d", tm.Hour(), tm.Minute())
}

func (caller *ToolCaller) ConvertTextDate(unix_sec int64) string {
	tm := time.Unix(unix_sec, 0)

	switch UI_GetDateFormat() {
	case "eu":
		return fmt.Sprintf("%d/%d/%d", tm.Day(), int(tm.Month()), tm.Year())

	case "us":
		return fmt.Sprintf("%d/%d/%d", int(tm.Month()), tm.Day(), tm.Year())

	case "iso":
		return fmt.Sprintf("%d-%02d-%02d", tm.Year(), int(tm.Month()), tm.Day())

	case "text":
		return fmt.Sprintf("%s %d, %d", Layout_GetMonthText(int(tm.Month())), tm.Day(), tm.Year())

	case "2base":
		return fmt.Sprintf("%d %d-%d", tm.Year(), int(tm.Month()), tm.Day())
	}

	return ""
}
func (caller *ToolCaller) ConvertTextDateTime(unix_sec int64) string {
	return caller.ConvertTextDate(unix_sec) + " " + Layout_ConvertTextTime(unix_sec)
}

func OsCopyFile(dst, src string) error {
	srcFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !srcFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	if err != nil {
		return err
	}

	err = os.Chmod(dst, srcFileStat.Mode())
	if err != nil {
		return err
	}

	return err
}

func Layout_MoveElement[T any](array_src *[]T, array_dst *[]T, src int, dst int) {
	//move(by swap one-by-one)
	if array_src == array_dst {
		for i := src; i < dst; i++ {
			(*array_dst)[i], (*array_dst)[i+1] = (*array_dst)[i+1], (*array_dst)[i]
		}
		for i := src; i > dst; i-- {
			(*array_dst)[i], (*array_dst)[i-1] = (*array_dst)[i-1], (*array_dst)[i]
		}
	} else {
		backup := (*array_src)[src]

		//remove
		*array_src = slices.Delete((*array_src), src, src+1)

		//insert
		if dst < len(*array_dst) {
			*array_dst = append((*array_dst)[:dst+1], (*array_dst)[dst:]...)
			(*array_dst)[dst] = backup
		} else {
			*array_dst = append(*array_dst, backup)
			//dst = len(*array_dst) - 1
		}
	}
}

//--- Network ---

type ToolServerInfo struct {
	bytes_written atomic.Int64
	bytes_read    atomic.Int64
}

func (info *ToolServerInfo) AddReadBytes(size int) {
	info.bytes_read.Add(int64(size))
}
func (info *ToolServerInfo) AddWrittenBytes(size int) {
	info.bytes_written.Add(int64(size))
}

type ToolServerClient struct {
	info *ToolServerInfo
	conn net.Conn
}

type ToolServer struct {
	port     int
	listener net.Listener
	exiting  bool

	info *ToolServerInfo
}

func NewToolServer(port int) *ToolServer {
	server := &ToolServer{}

	port_last := port + 1000
	for port < port_last {
		var err error
		server.listener, err = net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			break
		}
		port++
	}
	if port == port_last {
		log.Fatal(fmt.Errorf("can not Listen()"))
	}
	server.port = port
	server.info = &ToolServerInfo{}

	fmt.Printf("Server is running on port: %d\n", server.port)
	return server
}

func (server *ToolServer) Destroy() {
	server.exiting = true
	server.listener.Close()

	fmt.Printf("Server port: %d closed\n", server.port)
}

func (server *ToolServer) Accept() (*ToolServerClient, error) {
	conn, err := server.listener.Accept()
	if err != nil {
		if server.exiting {
			return nil, nil
		}
		return nil, err
	}
	return &ToolServerClient{info: server.info, conn: conn}, nil
}

func (client *ToolServerClient) Destroy() {
	client.conn.Close()
}

func (client *ToolServerClient) ReadInt() (uint64, error) {
	var sz [8]byte
	_, err := client.conn.Read(sz[:])
	if err != nil {
		return 0, err
	}
	client.info.AddReadBytes(8)

	return binary.LittleEndian.Uint64(sz[:]), nil
}

func (client *ToolServerClient) WriteInt(value uint64) error {
	var val [8]byte
	binary.LittleEndian.PutUint64(val[:], value)
	_, err := client.conn.Write(val[:])
	if err != nil {
		return err
	}

	client.info.AddWrittenBytes(8)
	return nil
}

func (client *ToolServerClient) ReadArray() ([]byte, error) {
	//recv size
	size, err := client.ReadInt()
	if err != nil {
		return nil, err
	}

	//recv data
	data := make([]byte, size)
	p := 0
	for p < int(size) {
		n, err := client.conn.Read(data[p:])
		if err != nil {
			return nil, err
		}
		p += n
	}

	client.info.AddReadBytes(int(size))

	return data, nil
}

func (client *ToolServerClient) WriteArray(data []byte) error {
	//send size
	err := client.WriteInt(uint64(len(data)))
	if err != nil {
		return err
	}

	//send data
	_, err = client.conn.Write(data)
	if err != nil {
		return err
	}
	client.info.AddWrittenBytes(len(data))

	return nil
}

type ToolClient struct {
	conn *net.TCPConn
}

func NewToolClient(addr string, port int) (*ToolClient, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	return &ToolClient{conn: conn}, nil
}
func (client *ToolClient) Destroy() {
	client.conn.Close()
}

func (client *ToolClient) ReadInt() (uint64, error) {
	var sz [8]byte
	_, err := client.conn.Read(sz[:])
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint64(sz[:]), nil
}

func (client *ToolClient) WriteInt(value uint64) error {
	var val [8]byte
	binary.LittleEndian.PutUint64(val[:], value)
	_, err := client.conn.Write(val[:])
	if err != nil {
		return err
	}
	return nil
}

func (client *ToolClient) ReadArray() ([]byte, error) {
	//recv size
	size, err := client.ReadInt()
	if err != nil {
		return nil, err
	}

	//recv data
	data := make([]byte, size)
	p := 0
	for p < int(size) {
		n, err := client.conn.Read(data[p:])
		if err != nil {
			return nil, err
		}
		p += n
	}

	return data, nil
}

func (client *ToolClient) WriteArray(data []byte) error {
	//send size
	err := client.WriteInt(uint64(len(data)))
	if err != nil {
		return err
	}

	//send data
	_, err = client.conn.Write(data)
	if err != nil {
		return err
	}

	return nil
}

//--- Ui ---

type UI struct {
	AppName  string `json:",omitempty"`
	FuncName string `json:",omitempty"`

	UID        uint64
	X, Y, W, H int
	LLMTip     string `json:",omitempty"`

	Cols  []UIGridSize
	Rows  []UIGridSize
	Items []*UI

	Dialogs []*UIDialog

	Enable        bool       `json:",omitempty"`
	EnableTouch   bool       `json:",omitempty"`
	Back_cd       color.RGBA `json:",omitempty"`
	Back_margin   float64    `json:",omitempty"`
	Back_rounding bool       `json:",omitempty"`
	Border_cd     color.RGBA `json:",omitempty"`
	ScrollV       UIScroll
	ScrollH       UIScroll

	//"omit empty" OR "Type string + Props interface{}" ....
	//Layout            *UI
	List              *UIList              `json:",omitempty"`
	Text              *UIText              `json:",omitempty"`
	Editbox           *UIEditbox           `json:",omitempty"`
	Button            *UIButton            `json:",omitempty"`
	Slider            *UISlider            `json:",omitempty"`
	FilePickerButton  *UIFilePickerButton  `json:",omitempty"`
	DatePickerButton  *UIDatePickerButton  `json:",omitempty"`
	ColorPickerButton *UIColorPickerButton `json:",omitempty"`
	Combo             *UICombo             `json:",omitempty"`
	Switch            *UISwitch            `json:",omitempty"`
	Checkbox          *UICheckbox          `json:",omitempty"`
	Divider           *UIDivider           `json:",omitempty"`
	OsmMap            *UIOsmMap            `json:",omitempty"`
	ChartLines        *UIChartLines        `json:",omitempty"`
	ChartColumns      *UIChartColumns      `json:",omitempty"`
	Image             *UIImage             `json:",omitempty"`
	YearCalendar      *UIYearCalendar      `json:",omitempty"`
	MonthCalendar     *UIMonthCalendar     `json:",omitempty"`
	DayCalendar       *UIDayCalendar       `json:",omitempty"`

	Paint []UIPaint `json:",omitempty"`

	changed func(newParams []byte) error
}

type UIText struct {
	layout  *UI
	Label   string
	Align_h int
	Align_v int
	Cd      color.RGBA
	Tooltip string

	Selection    bool
	Formating    bool
	Multiline    bool
	Linewrapping bool

	EnableDropFile bool
	dropFile       func(pathes []string) error
}
type UIEditbox struct {
	layout     *UI
	Name       string
	Error      string
	Value      *string
	ValueFloat *float64
	ValueInt   *int
	Precision  int
	Ghost      string
	Tooltip    string
	Password   bool

	Align_h int //0=left, 1=center, 2=right
	Align_v int //0=top, 1=center, 2=bottom

	Formating    bool
	Multiline    bool
	Linewrapping bool

	changed func() error
	enter   func() error
}
type UISlider struct {
	layout *UI
	Error  string
	Value  *float64
	Min    float64
	Max    float64
	Step   float64

	changed func() error
}
type UIButton struct {
	layout  *UI
	Label   string
	Tooltip string
	Align   int

	Shortcut byte

	Background  float64
	Border      bool
	IconBlob    []byte
	IconPath    string
	Icon_align  int
	Icon_margin float64
	BrowserUrl  string
	Cd          color.RGBA

	ConfirmQuestion string

	Drag_group              string
	Drop_group              string
	Drag_source             string
	Drag_index              int
	Drop_h, Drop_v, Drop_in bool

	clicked func() error

	dropMove func(src_i, dst_i int, src_source, dst_source string) error
}

type UIFilePickerButton struct {
	layout      *UI
	Error       string
	Path        *string
	Preview     bool
	OnlyFolders bool

	changed func() error
}
type UIDatePickerButton struct {
	layout   *UI
	Error    string
	Date     *int64
	Page     *int64
	ShowTime bool
	changed  func() error
}
type UIColorPickerButton struct {
	layout  *UI
	Error   string
	Cd      *color.RGBA
	changed func() error
}

type UICombo struct {
	layout      *UI
	Error       string
	Value       *string
	Labels      []string
	Values      []string
	DialogWidth float64

	changed func() error
}

type UISwitch struct {
	layout  *UI
	Error   string
	Label   string
	Tooltip string
	Value   *bool

	changed func() error
}

type UICheckbox struct {
	layout  *UI
	Error   string
	Label   string
	Tooltip string
	Value   *float64

	changed func() error
}

type UIDivider struct {
	layout     *UI
	Horizontal bool
}

type UIOsmMapLoc struct {
	Lon   float64
	Lat   float64
	Label string
}
type UIOsmMapLocators struct {
	Locators []UIOsmMapLoc
	clicked  func(i int, caller *ToolCaller)
}

type UIOsmMapSegmentTrk struct {
	Lon  float64
	Lat  float64
	Ele  float64
	Time string
	Cd   color.RGBA
}
type UIOsmMapSegment struct {
	Trkpts []UIOsmMapSegmentTrk
	Label  string
}
type UIOsmMapRoute struct {
	Segments []UIOsmMapSegment
}
type UIOsmMap struct {
	layout         *UI
	Lon, Lat, Zoom *float64
	Locators       []UIOsmMapLocators
	Routes         []UIOsmMapRoute
}

type UIChartPoint struct {
	X  float64
	Y  float64
	Cd color.RGBA
}

type UIChartLine struct {
	Points []UIChartPoint
	Label  string
	Cd     color.RGBA
}

type UIChartLines struct {
	layout *UI

	Lines []UIChartLine

	X_unit, Y_unit        string
	Bound_x0, Bound_y0    bool
	Point_rad, Line_thick float64
	Draw_XHelpLines       bool
	Draw_YHelpLines       bool
}

type UIChartColumnValue struct {
	Value float64
	Label string
	Cd    color.RGBA
}

type UIChartColumn struct {
	Values []UIChartColumnValue
}

type UIChartColumns struct {
	layout *UI

	X_unit, Y_unit string
	Bound_y0       bool
	Y_as_time      bool
	Columns        []UIChartColumn
	X_Labels       []string
	ColumnMargin   float64
}

type UIImage struct {
	layout *UI

	Blob    []byte
	Path    string
	Tooltip string

	Cd          color.RGBA
	Draw_border bool

	Margin  float64
	Align_h int
	Align_v int

	Translate_x, Translate_y float64
	Scale_x, Scale_y         float64
}

type UIList struct {
	layout *UI

	AutoSpacing bool
	//Items       []*UI
}

type UIGridSize struct {
	Pos int
	Min float64
	Max float64

	Default_resize float64

	SetFromChild_min float64 `json:",omitempty"`
	SetFromChild_max float64 `json:",omitempty"`
}
type UIScroll struct {
	Hide   bool
	Narrow bool
}
type UIPaint_Rect struct {
	Cd, Cd_over, Cd_down color.RGBA
	Width                float64
	X, Y, W, H           float64
}
type UIPaint_Circle struct {
	Cd, Cd_over, Cd_down color.RGBA
	Rad                  float64
	Width                float64
	X, Y                 float64
}
type UIPaint_Line struct {
	Cd, Cd_over, Cd_down color.RGBA
	Width                float64
	Sx, Sy, Ex, Ey       float64
}
type UIPaint_Text struct {
	Text                 string
	Cd, Cd_over, Cd_down color.RGBA
	Align_v              int
	Align_h              int
	Formating            bool
	Multiline            bool
	Linewrapping         bool
	Sx, Sy, Ex, Ey       float64
}
type UIPaintBrushPoint struct {
	X int
	Y int
}
type UIPaint_Brush struct {
	Cd     color.RGBA
	Points []UIPaintBrushPoint
}

type UIPaint struct {
	Rectangle *UIPaint_Rect   `json:",omitempty"`
	Circle    *UIPaint_Circle `json:",omitempty"`
	Line      *UIPaint_Line   `json:",omitempty"`
	Text      *UIPaint_Text   `json:",omitempty"`
	Brush     *UIPaint_Brush  `json:",omitempty"`
}

type UICalendarEvent struct {
	EventID int64
	GroupID int64

	Title string

	Start    int64 //unix time
	Duration int64 //seconds

	Color color.RGBA
}

type UIYearCalendar struct {
	layout *UI
	Year   int
}
type UIMonthCalendar struct {
	layout *UI
	Year   int
	Month  int //1=January, 2=February, etc.

	Events []UICalendarEvent
}
type UIDayCalendar struct {
	layout *UI
	Days   []int64
	Events []UICalendarEvent
}

type ToolCmd struct {
	Dialog_Open_UID     uint64 `json:",omitempty"`
	Dialog_Relative_UID uint64 `json:",omitempty"`
	Dialog_OnTouch      bool   `json:",omitempty"`

	Dialog_Close_UID uint64 `json:",omitempty"`

	Editbox_Activate string `json:",omitempty"`

	VScrollToTheTop      uint64 `json:",omitempty"`
	VScrollToTheBottom   uint64 `json:",omitempty"`
	VScrollToTheBottomIf uint64 `json:",omitempty"`
	HScrollToTheLeft     uint64 `json:",omitempty"`
	HScrollToTheRight    uint64 `json:",omitempty"`

	SetClipboardText string `json:",omitempty"`
}

type UIDialog struct {
	UID string
	UI  UI
}

func (dia *UIDialog) OpenCentered(caller *ToolCaller) {
	caller._addCmd(ToolCmd{Dialog_Open_UID: dia.UI.UID})
}
func (dia *UIDialog) OpenRelative(relative *UI, caller *ToolCaller) {
	caller._addCmd(ToolCmd{Dialog_Open_UID: dia.UI.UID, Dialog_Relative_UID: relative.UID})
}
func (dia *UIDialog) OpenOnTouch(caller *ToolCaller) {
	caller._addCmd(ToolCmd{Dialog_Open_UID: dia.UI.UID, Dialog_OnTouch: true})
}
func (dia *UIDialog) Close(caller *ToolCaller) {
	caller._addCmd(ToolCmd{Dialog_Close_UID: dia.UI.UID})
}

func (ui *UI) ActivateEditbox(editbox_name string, caller *ToolCaller) {
	caller._addCmd(ToolCmd{Editbox_Activate: editbox_name})
}

func (ui *UI) VScrollToTheTop(caller *ToolCaller) {
	caller._addCmd(ToolCmd{VScrollToTheTop: ui.UID})
}
func (ui *UI) VScrollToTheBottom(onlyWhenAtBottom bool, caller *ToolCaller) {
	if onlyWhenAtBottom {
		caller._addCmd(ToolCmd{VScrollToTheBottomIf: ui.UID})
	} else {
		caller._addCmd(ToolCmd{VScrollToTheBottom: ui.UID})
	}
}
func (ui *UI) HScrollToTheLeft(caller *ToolCaller) {
	caller._addCmd(ToolCmd{HScrollToTheLeft: ui.UID})
}
func (ui *UI) HScrollToTheRight(caller *ToolCaller) {
	caller._addCmd(ToolCmd{HScrollToTheRight: ui.UID})
}

func (caller *ToolCaller) SetClipboardText(text string) {
	caller._addCmd(ToolCmd{SetClipboardText: text})
}

func (ui *UI) SetColumn(pos int, min, max float64) {
	for i := range ui.Cols {
		if ui.Cols[i].Pos == pos {
			ui.Cols[i].Min = min
			ui.Cols[i].Max = max
			return
		}
	}
	ui.Cols = append(ui.Cols, UIGridSize{Pos: pos, Min: min, Max: max})
}
func (ui *UI) SetRow(pos int, min, max float64) {
	for i := range ui.Rows {
		if ui.Rows[i].Pos == pos {
			ui.Rows[i].Min = min
			ui.Rows[i].Max = max
			return
		}
	}
	ui.Rows = append(ui.Rows, UIGridSize{Pos: pos, Min: min, Max: max})
}

func (ui *UI) SetColumnResizable(pos int, min, max, default_size float64) {
	for i := range ui.Cols {
		if ui.Cols[i].Pos == pos {
			ui.Cols[i].Min = min
			ui.Cols[i].Max = max
			ui.Cols[i].Default_resize = default_size
			return
		}
	}
	ui.Cols = append(ui.Cols, UIGridSize{Pos: pos, Min: min, Max: max, Default_resize: default_size})
}
func (ui *UI) SetRowResizable(pos int, min, max, default_size float64) {
	for i := range ui.Rows {
		if ui.Rows[i].Pos == pos {
			ui.Rows[i].Min = min
			ui.Rows[i].Max = max
			ui.Rows[i].Default_resize = default_size
			return
		}
	}
	ui.Rows = append(ui.Rows, UIGridSize{Pos: pos, Min: min, Max: max, Default_resize: default_size})
}

func (ui *UI) SetColumnFromSub(grid_y int, min_size, max_size float64) {
	newItem := UIGridSize{Pos: grid_y, SetFromChild_min: min_size, SetFromChild_max: max_size}

	for i := range ui.Cols {
		if ui.Cols[i].Pos == grid_y {
			ui.Cols[i] = newItem
			return
		}
	}
	ui.Cols = append(ui.Cols, newItem)
}

func (ui *UI) SetRowFromSub(grid_y int, min_size, max_size float64) {
	newItem := UIGridSize{Pos: grid_y, SetFromChild_min: min_size, SetFromChild_max: max_size}

	for i := range ui.Rows {
		if ui.Rows[i].Pos == grid_y {
			ui.Rows[i] = newItem
			return
		}
	}
	ui.Rows = append(ui.Rows, newItem)
}

func (ui *UI) AddText(x, y, w, h int, label string) *UIText {
	item := &UIText{Label: label, Align_h: 0, Align_v: 1, Selection: true, Formating: true, Multiline: true, Linewrapping: true, layout: _newUIItem(x, y, w, h)}
	item.layout.Text = item
	ui._addUISub(item.layout, "")
	return item
}
func (ui *UI) AddTextLabel(x, y, w, h int, value string) *UIText {
	txt := ui.AddText(x, y, w, h, "<b>"+value+"</b>")
	txt.Align_h = 1
	return txt
}

func (ui *UI) AddEditboxString(x, y, w, h int, value *string) *UIEditbox {
	item := &UIEditbox{Value: value, Align_v: 1, Formating: true, layout: _newUIItem(x, y, w, h)}
	item.layout.Editbox = item
	ui._addUISub(item.layout, "")
	return item
}
func (ui *UI) AddEditboxInt(x, y, w, h int, value *int) *UIEditbox {
	item := &UIEditbox{ValueInt: value, Align_v: 1, Formating: true, layout: _newUIItem(x, y, w, h)}
	item.layout.Editbox = item
	ui._addUISub(item.layout, "")
	return item
}
func (ui *UI) AddEditboxFloat(x, y, w, h int, value *float64, precision int) *UIEditbox {
	item := &UIEditbox{ValueFloat: value, Align_v: 1, Precision: precision, Formating: true, layout: _newUIItem(x, y, w, h)}
	item.layout.Editbox = item
	ui._addUISub(item.layout, "")
	return item
}

func (ui *UI) AddSlider(x, y, w, h int, value *float64, min, max, step float64) *UISlider {
	item := &UISlider{Value: value, Min: min, Max: max, Step: step, layout: _newUIItem(x, y, w, h)}
	item.layout.Slider = item
	ui._addUISub(item.layout, "")
	return item
}

func (ui *UI) AddButton(x, y, w, h int, label string) *UIButton {
	item := &UIButton{Label: label, Background: 1, Align: 1, layout: _newUIItem(x, y, w, h)}
	item.layout.Button = item
	ui._addUISub(item.layout, "")
	return item
}

func (ui *UI) AddYearCalendar(x, y, w, h int, Year int) *UIYearCalendar {
	item := &UIYearCalendar{Year: Year, layout: _newUIItem(x, y, w, h)}
	item.layout.YearCalendar = item
	ui._addUISub(item.layout, "")
	return item
}

func (ui *UI) AddMonthCalendar(x, y, w, h int, Year int, Month int, Events []UICalendarEvent) *UIMonthCalendar {
	item := &UIMonthCalendar{Year: Year, Month: Month, Events: Events, layout: _newUIItem(x, y, w, h)}
	item.layout.MonthCalendar = item
	ui._addUISub(item.layout, "")
	return item
}
func (ui *UI) AddDayCalendar(x, y, w, h int, Days []int64, Events []UICalendarEvent) *UIDayCalendar {
	item := &UIDayCalendar{Days: Days, Events: Events, layout: _newUIItem(x, y, w, h)}
	item.layout.DayCalendar = item
	ui._addUISub(item.layout, "")
	return item
}

func (ui *UI) AddFilePickerButton(x, y, w, h int, path *string, preview bool, onlyFolders bool) *UIFilePickerButton {
	item := &UIFilePickerButton{Path: path, Preview: preview, OnlyFolders: onlyFolders, layout: _newUIItem(x, y, w, h)}
	item.layout.FilePickerButton = item
	ui._addUISub(item.layout, "")
	return item
}
func (ui *UI) AddDatePickerButton(x, y, w, h int, date *int64, page *int64, showTime bool) *UIDatePickerButton {
	item := &UIDatePickerButton{Date: date, Page: page, ShowTime: showTime, layout: _newUIItem(x, y, w, h)}
	item.layout.DatePickerButton = item
	ui._addUISub(item.layout, "")
	return item
}
func (ui *UI) AddColorPickerButton(x, y, w, h int, cd *color.RGBA) *UIColorPickerButton {
	item := &UIColorPickerButton{Cd: cd, layout: _newUIItem(x, y, w, h)}
	item.layout.ColorPickerButton = item
	ui._addUISub(item.layout, "")
	return item
}

func (ui *UI) AddCombo(x, y, w, h int, value *string, labels []string, values []string) *UICombo {
	item := &UICombo{Value: value, Labels: labels, Values: values, DialogWidth: 5, layout: _newUIItem(x, y, w, h)}
	item.layout.Combo = item
	ui._addUISub(item.layout, "")
	return item
}

func (ui *UI) AddSwitch(x, y, w, h int, label string, value *bool) *UISwitch {
	item := &UISwitch{Label: label, Value: value, layout: _newUIItem(x, y, w, h)}
	item.layout.Switch = item
	ui._addUISub(item.layout, "")
	return item
}
func (ui *UI) AddCheckbox(x, y, w, h int, label string, value *float64) *UICheckbox {
	item := &UICheckbox{Label: label, Value: value, layout: _newUIItem(x, y, w, h)}
	item.layout.Checkbox = item
	ui._addUISub(item.layout, "")
	return item
}
func (ui *UI) AddDivider(x, y, w, h int, horizontal bool) *UIDivider {
	item := &UIDivider{Horizontal: horizontal, layout: _newUIItem(x, y, w, h)}
	item.layout.Divider = item
	ui._addUISub(item.layout, "")
	return item
}

func (ui *UI) AddOsmMap(x, y, w, h int, lon, lat, zoom *float64) *UIOsmMap {
	item := &UIOsmMap{Lon: lon, Lat: lat, Zoom: zoom, layout: _newUIItem(x, y, w, h)}
	item.layout.OsmMap = item
	ui._addUISub(item.layout, "")
	return item
}
func (mp *UIOsmMap) AddLocators(loc UIOsmMapLocators) {
	mp.Locators = append(mp.Locators, loc)
}
func (mp *UIOsmMap) AddRoute(route UIOsmMapRoute) {
	mp.Routes = append(mp.Routes, route)
}

func (ui *UI) AddChartLines(x, y, w, h int, Lines []UIChartLine) *UIChartLines {
	item := &UIChartLines{Lines: Lines, Point_rad: 0.2, Line_thick: 0.03, layout: _newUIItem(x, y, w, h)}
	item.layout.ChartLines = item
	ui._addUISub(item.layout, "")
	return item
}
func (ui *UI) AddChartColumns(x, y, w, h int, columns []UIChartColumn, x_labels []string) *UIChartColumns {
	item := &UIChartColumns{Columns: columns, X_Labels: x_labels, ColumnMargin: 0.2, layout: _newUIItem(x, y, w, h)}
	item.layout.ChartColumns = item
	ui._addUISub(item.layout, "")
	return item
}

func (ui *UI) _addImage(x, y, w, h int, path string, blob []byte, cd color.RGBA) *UIImage {
	item := &UIImage{Path: path, Blob: blob, Align_h: 1, Align_v: 1, Margin: 0.1, Cd: cd, layout: _newUIItem(x, y, w, h)}
	item.layout.Image = item
	ui._addUISub(item.layout, "")
	return item
}
func (ui *UI) AddImagePath(x, y, w, h int, path string) *UIImage {
	return ui._addImage(x, y, w, h, path, nil, color.RGBA{255, 255, 255, 255})
}
func (ui *UI) AddImageBlob(x, y, w, h int, blob []byte) *UIImage {
	return ui._addImage(x, y, w, h, "", blob, color.RGBA{255, 255, 255, 255})
}

func (ui *UI) AddLayoutList(x, y, w, h int, autoSpacing bool) *UIList {
	item := &UIList{AutoSpacing: autoSpacing, layout: _newUIItem(x, y, w, h)}
	item.layout.List = item

	ui._addUISub(item.layout, "")
	return item
}
func (list *UIList) AddItem() *UI {
	item := _newUIItem(0, len(list.layout.Items), 1, 1)
	list.layout._addUISub(item, "")
	return item
}

func (ui *UI) AddLayout(x, y, w, h int) *UI {
	item := _newUIItem(x, y, w, h)
	ui._addUISub(item, "")
	return item
}
func (ui *UI) AddLayoutWithName(x, y, w, h int, name string) *UI {
	item := _newUIItem(x, y, w, h)
	ui._addUISub(item, name)
	return item
}

func (ui *UI) FindDialog(name string) *UIDialog {
	for _, dia := range ui.Dialogs {
		if dia.UID == name {
			return dia
		}
	}
	return nil
}
func (ui *UI) AddDialog(uid string) *UIDialog {
	dia := ui.FindDialog(uid)
	if dia == nil {
		dia = &UIDialog{UID: uid, UI: *_newUIItem(0, 0, 0, 0)}
		ui.Dialogs = append(ui.Dialogs, dia)
		dia.UI._computeUID(ui, uid)
	} else {
		fmt.Println("Dialog already exist")
	}
	return dia
}

func (ui *UI) AddDialogBorder(name string, title string) (*UIDialog, *UI) {
	dia := ui.AddDialog(name)
	lay := dia.UI
	lay.SetColumnFromSub(1, 1, 100)
	lay.SetRowFromSub(1, 1, 100)
	lay.SetColumn(2, 1, 1)
	lay.SetRow(2, 1, 1)

	tx := lay.AddText(0, 0, 3, 1, title)
	tx.Align_h = 1

	return dia, lay.AddLayout(1, 1, 1, 1)
}

func (ui *UI) AddTool(x, y, w, h int, fnRun func(caller *ToolCaller, ui *UI) error, caller *ToolCaller) (*UI, error) {
	ret_ui := _newUIItem(x, y, w, h)
	ui._addUISub(ret_ui, "")

	out_error := fnRun(caller, ret_ui)

	if out_error == nil {
		if !caller._sendProgress(1, "") {
			out_error = errors.New("_interrupted_")
		}
	}

	if out_error != nil {
		ret_ui.Cols = nil
		ret_ui.Rows = nil
		ret_ui.Items = nil

		ret_ui.SetColumn(0, 1, 100)
		ret_ui.SetRow(0, 1, 100)
		tx := ret_ui.AddText(0, 0, 1, 1, fmt.Sprintf("<i>Error: %s</i>", out_error.Error()))
		tx.Align_h = 1
		tx.Align_v = 1
		tx.Cd = UI_GetPalette().E
	}

	return ret_ui, out_error
}

func (ui *UI) AddToolApp(x, y, w, h int, appName string, funcName string, jsParams []byte, caller *ToolCaller) (*UI, error) {
	ret_ui := _newUIItem(x, y, w, h)
	ui._addUISub(ret_ui, "")

	//call router
	_, uiJs, err := caller.callFuncSubCall(ret_ui.UID, appName, funcName, jsParams)
	if err == nil {
		err := json.Unmarshal(uiJs, ret_ui)
		if Tool_Error(err) == nil {
			ret_ui.X = x
			ret_ui.Y = y
			ret_ui.W = w
			ret_ui.H = h
			ret_ui.AppName = appName
			ret_ui.FuncName = funcName

			return ret_ui, nil
		}
	}

	//error
	ret_ui.SetColumn(0, 1, 100)
	ret_ui.SetRow(0, 1, 100)
	tx := ret_ui.AddText(0, 0, 1, 1, fmt.Sprintf("<i>Error</i>"))
	tx.Align_h = 1
	tx.Align_v = 1
	tx.Cd = UI_GetPalette().E

	return ret_ui, nil
}

func (ui *UI) Paint_Rect(x, y, w, h float64, cd, cd_over, cd_down color.RGBA, width float64) {
	ui.Paint = append(ui.Paint, UIPaint{Rectangle: &UIPaint_Rect{
		Cd: cd, Cd_over: cd, Cd_down: cd,
		Width: width,
		X:     x,
		Y:     y,
		W:     w,
		H:     h,
	}})
}

func (ui *UI) Paint_CircleOnPos(x, y float64, rad float64, cd, cd_over, cd_down color.RGBA, width float64) {
	ui.Paint = append(ui.Paint, UIPaint{Circle: &UIPaint_Circle{
		Cd: cd, Cd_over: cd_over, Cd_down: cd_down,
		Rad:   rad,
		Width: width,
		X:     x,
		Y:     y,
	}})
}

func (ui *UI) Paint_Line(sx, sy, ex, ey float64, cd color.RGBA, width float64) {
	ui.Paint = append(ui.Paint, UIPaint{Line: &UIPaint_Line{
		Cd: cd, Cd_over: cd, Cd_down: cd,
		Width: width,
		Sx:    sx,
		Sy:    sy,
		Ex:    ex,
		Ey:    ey,
	}})
}

func (ui *UI) Paint_Text(sx, sy, ex, ey float64, Text string, cd color.RGBA, Align_v int, Align_h int, Formating bool, Multiline bool, Linewrapping bool) {
	ui.Paint = append(ui.Paint, UIPaint{Text: &UIPaint_Text{
		Cd: cd, Cd_over: cd, Cd_down: cd,
		Text:         Text,
		Align_v:      Align_v,
		Align_h:      Align_h,
		Formating:    Formating,
		Multiline:    Multiline,
		Linewrapping: Linewrapping,
		Sx:           sx,
		Sy:           sy,
		Ex:           ex,
		Ey:           ey,
	}})
}

func (ui *UI) Paint_Brush(cd color.RGBA, pts []UIPaintBrushPoint) {
	ui.Paint = append(ui.Paint, UIPaint{Brush: &UIPaint_Brush{
		Cd:     cd,
		Points: pts,
	}})
}

type LLMCompletion struct {
	UID string

	Temperature       float64
	Max_tokens        int
	Top_p             float64
	Frequency_penalty float64
	Presence_penalty  float64
	Reasoning_effort  string //"low", "medium", "high"

	AppName string //load tools from

	PreviousMessages []byte //[]*ChatMsg
	SystemMessage    string
	UserMessage      string
	UserFiles        []string

	Response_format string

	Max_iteration int

	Out_StatusCode   int
	Out_messages     []byte //[]*ChatMsg
	Out_last_message string

	delta func(msgJs []byte) //msg *ChatMsg
}

func (comp *LLMCompletion) Run(caller *ToolCaller) error {
	compJs, err := json.Marshal(comp)
	if err != nil {
		return err
	}

	cl, err := NewToolClient("localhost", g_main.router_port)
	if Tool_Error(err) == nil {
		defer cl.Destroy()

		err = cl.WriteArray([]byte("llm_complete"))
		if Tool_Error(err) == nil {

			err = cl.WriteInt(caller.msg_id)
			if Tool_Error(err) == nil {

				err = cl.WriteArray(compJs)
				if Tool_Error(err) == nil {

					//delta(s)
					for {
						msgJs, err := cl.ReadArray()
						if Tool_Error(err) == nil && len(msgJs) > 0 {
							if comp.delta != nil {
								comp.delta(msgJs)
							}
						} else {
							break
						}
					}

					//result
					compJs, err = cl.ReadArray()
					if Tool_Error(err) == nil {
						err = json.Unmarshal(compJs, comp)
						if Tool_Error(err) == nil {
							return nil
						}
					}
				}
			}
		}
	}

	return fmt.Errorf("connection failed")
}

type LLMTranscribe struct {
	UID string

	AudioBlob    []byte
	BlobFileName string //ext.... (blob.wav, blob.mp3)

	Temperature     float64 //0
	Response_format string

	Out_StatusCode int
	Out_Output     []byte
}

func (comp *LLMTranscribe) Run(caller *ToolCaller) error {
	compJs, err := json.Marshal(comp)
	if err != nil {
		return err
	}

	cl, err := NewToolClient("localhost", g_main.router_port)
	if Tool_Error(err) == nil {
		defer cl.Destroy()

		err = cl.WriteArray([]byte("llm_transcribe"))
		if Tool_Error(err) == nil {

			err = cl.WriteInt(caller.msg_id)
			if Tool_Error(err) == nil {

				err = cl.WriteArray(compJs)
				if Tool_Error(err) == nil {

					//delta(s)
					for {
						msgJs, err := cl.ReadArray()
						if Tool_Error(err) == nil && len(msgJs) > 0 {
							//if comp.delta != nil {
							//	comp.delta(msgJs)
							//}
						} else {
							break
						}
					}

					//result
					compJs, err = cl.ReadArray()
					if Tool_Error(err) == nil {
						err = json.Unmarshal(compJs, comp)
						if Tool_Error(err) == nil {
							return nil
						}
					}
				}
			}
		}
	}

	return fmt.Errorf("connection failed")
}

func (caller *ToolCaller) StartRecordingMicrophone(mic_uid string) error {
	cl, err := NewToolClient("localhost", g_main.router_port)
	if Tool_Error(err) == nil {
		defer cl.Destroy()

		err = cl.WriteArray([]byte("start_microphone"))
		if Tool_Error(err) == nil {

			err = cl.WriteInt(caller.msg_id)
			if Tool_Error(err) == nil {

				err = cl.WriteArray([]byte(mic_uid))
				if Tool_Error(err) == nil {

					//recv
					errBytes, err := cl.ReadArray()
					if Tool_Error(err) == nil {
						if len(errBytes) > 0 {
							err = errors.New(string(errBytes))
						}
						return Tool_Error(err)
					}
				}
			}
		}
	}
	return fmt.Errorf("StartRecordingMicrophone() connection failed")
}

func (caller *ToolCaller) StopRecordingMicrophone(mic_uid string, cancel bool, format string) ([]byte, error) {
	cl, err := NewToolClient("localhost", g_main.router_port)
	if Tool_Error(err) == nil {
		defer cl.Destroy()

		err = cl.WriteArray([]byte("stop_microphone"))
		if Tool_Error(err) == nil {

			err = cl.WriteArray([]byte(mic_uid))
			if Tool_Error(err) == nil {

				cancelInt := uint64(0)
				if cancel {
					cancelInt = 1
				}
				err = cl.WriteInt(cancelInt)
				if Tool_Error(err) == nil {

					err = cl.WriteArray([]byte(format))
					if Tool_Error(err) == nil {

						//recv
						errBytes, err := cl.ReadArray()
						if Tool_Error(err) == nil {
							outBytes, err := cl.ReadArray()
							if Tool_Error(err) == nil {

								if len(errBytes) > 0 {
									err = errors.New(string(errBytes))
									Tool_Error(err)
								}
								return outBytes, err
							}
						}
					}
				}
			}
		}
	}
	return nil, fmt.Errorf("StopRecordingMicrophone() connection failed")
}

//--- Auto generated ---
