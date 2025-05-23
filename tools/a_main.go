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
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

func _loadInstance[T any](file string, structName string, format string, defInst *T, save bool, caller *ToolCaller) (*T, error) {
	if file == "" {
		file = fmt.Sprintf("%s-%s.%s", structName, structName, format)
	}

	caller._addSourceStruct(structName)

	//find
	g_files_lock.Lock()
	inst, found := g_files[file]
	g_files_lock.Unlock()
	if found {
		inst.save = save
		return inst.st.(*T), nil
	}

	//get file from router
	var data []byte
	{
		cl, err := NewToolClient("localhost", g_main.router_port)
		if Tool_Error(err) == nil {
			defer cl.Destroy()

			err = cl.WriteArray([]byte("read_file"))
			if Tool_Error(err) == nil {

				err = cl.WriteArray([]byte(file))
				if Tool_Error(err) == nil {

					exist, err := cl.ReadInt()
					if Tool_Error(err) == nil {

						if exist > 0 {
							data, err = cl.ReadArray()
							Tool_Error(err)
						}
					}
				}
			}
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

func (caller *ToolCaller) _saveInstances() {

	g_files_lock.Lock()
	defer g_files_lock.Unlock()

	for path, it := range g_files {
		if !it.save {
			continue
		}

		var err error
		var data []byte
		switch strings.ToLower(filepath.Ext(path)) {
		case ".json":
			data, err = json.Marshal(it.st)
		case ".xml":
			data, err = xml.Marshal(it.st)
		}

		if err == nil && !bytes.Equal(it.data, data) {
			//send file to router
			cl, err := NewToolClient("localhost", g_main.router_port)
			if Tool_Error(err) == nil {
				defer cl.Destroy()

				err = cl.WriteArray([]byte("write_file"))
				if Tool_Error(err) == nil {

					err = cl.WriteArray([]byte(path))
					if Tool_Error(err) == nil {

						err = cl.WriteArray(data)
						Tool_Error(err)
					}
				}
			}

			it.data = data
		}
	}
}

type DevPalette struct {
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

func (pl *DevPalette) GetGrey(t float64) color.RGBA {
	return Color_Aprox(pl.B, pl.OnB, t)
}

func (caller *ToolCaller) GetPalette() *DevPalette {
	dev, err := NewDeviceSettings("", caller)
	if err == nil {
		return dev.GetPalette()
	}
	return &DevPalette{}
}
func (caller *ToolCaller) GetDateFormat() string {
	dev, err := NewDeviceSettings("", caller)
	if err == nil {
		return dev.DateFormat
	}
	return "eu"
}

type ToolProgram struct {
	router_port int
	server      *ToolServer
}

type _Instance struct {
	data []byte
	st   any
	save bool
}

type ToolCaller struct {
	msg_id uint64
	ui_uid uint64

	last_send_progress_ms int64 //ms

	cmds []ToolCmd

	source_structs []string
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

func main() {
	log.SetFlags(log.Llongfile) //log.LstdFlags | log.Lshortfile

	if len(os.Args) < 2 {
		log.Fatal("missing 'port' argument(s): ", os.Args)
	}

	router_port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	g_uis = make(map[uint64]*ToolUI)
	g_files = make(map[string]*_Instance)
	g_main.router_port = router_port
	g_main.server = NewToolServer(router_port + 100)
	defer g_main.server.Destroy()

	//report tool into router server
	{
		cl, err := NewToolClient("localhost", router_port)
		if err != nil {
			log.Fatal(err)
		}
		err = cl.WriteArray([]byte("register"))
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

											cmdsJs, _ := json.Marshal(ui.Caller.cmds)
											err = cl.WriteArray(cmdsJs) //commands
											Tool_Error(err)
										}
										cl.Destroy()

										ui.Caller._saveInstances()
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
							if len(paramsJs) == 0 {
								paramsJs = []byte("{}")
							}
							if Tool_Error(err) == nil {
								ok = true
								go func() {
									ui := _newUIItem(0, 0, 1, 1)
									ui.UID = caller.ui_uid

									var out_error error

									fnRun, out_params := FindToolRunFunc(string(funcName), paramsJs)
									if fnRun != nil {
										out_error = fnRun(caller, ui)
									}

									if caller.ui_uid != 0 {
										g_uis_lock.Lock()
										g_uis[caller.ui_uid] = &ToolUI{ui: ui,
											parameters: out_params,
											Caller:     caller}
										g_uis_lock.Unlock()
									}

									if out_error == nil {
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
										Tool_Error(err)

										uiJs, out_error = json.Marshal(ui)
										Tool_Error(err)

										cmdsJs, out_error = json.Marshal(caller.cmds)
										Tool_Error(err)

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

									caller._saveInstances()
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
		fmt.Printf("\033[31merror: %v\033[0m\n", err)
	}
	return err
}

func (caller *ToolCaller) _addSourceStruct(name string) {
	caller.source_structs = append(caller.source_structs, name)
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

type SdkTool struct {
	Name  string
	Attrs []string

	Running bool

	Compile_error string
	Cmd_error     string
}

type SdkMsg struct {
	Id             string
	FuncName       string
	Progress_label string
	Progress_done  float64
	Start_time     float64
}

func (msg *SdkMsg) GetLabel() string {
	label := msg.Progress_label
	if label == "" {
		label = msg.FuncName + "()"
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

func callFuncGetToolsShemas() []byte {
	cl, err := NewToolClient("localhost", g_main.router_port)
	if Tool_Error(err) == nil {
		defer cl.Destroy()

		err = cl.WriteArray([]byte("get_tools_shemas"))
		if Tool_Error(err) == nil {
			oaiJs, err := cl.ReadArray()
			if Tool_Error(err) == nil {
				return oaiJs
			}
		}
	}
	return nil
}

func callFuncGetToolsShemasBySource(source string) []byte {
	cl, err := NewToolClient("localhost", g_main.router_port)
	if Tool_Error(err) == nil {
		defer cl.Destroy()

		err = cl.WriteArray([]byte("get_tools_shemas_by_source"))
		if Tool_Error(err) == nil {
			err = cl.WriteArray([]byte(source))
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

type _ToolSource struct {
	FileTime    int64
	Description string
	Tools       []string
}

func callFuncGetSources(source string) map[string]*_ToolSource {
	cl, err := NewToolClient("localhost", g_main.router_port)
	if Tool_Error(err) == nil {
		defer cl.Destroy()

		err = cl.WriteArray([]byte("get_sources"))
		if Tool_Error(err) == nil {
			js, err := cl.ReadArray()
			if Tool_Error(err) == nil {
				var sources map[string]*_ToolSource
				err = json.Unmarshal(js, &sources)
				if Tool_Error(err) == nil {
					return sources
				}
			}
		}
	}
	return nil
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

type UI struct {
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

func (ui *UI) _addTool(x, y, w, h int, funcName string, fnRun func(caller *ToolCaller, ui *UI) error, caller *ToolCaller) (*UI, error) {
	ret_ui := _newUIItem(x, y, w, h)
	ui._addUISub(ret_ui, "")

	var out_error error
	if fnRun != nil {
		out_error = fnRun(caller, ret_ui)
	} else {
		out_error = fmt.Errorf("Tool function '%s' not found", funcName)
	}

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
		tx.Cd = caller.GetPalette().E
	}

	return ret_ui, out_error
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
func CallToolByName(funcName string, jsParams []byte, caller *ToolCaller) (*UI, interface{}, error) {
	fnRun, st := FindToolRunFunc(funcName, jsParams)
	ui, err := _callTool(funcName, fnRun, caller)
	return ui, st, err
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

	switch caller.GetDateFormat() {
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
