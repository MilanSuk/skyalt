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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ToolsCmdItem struct {
	Schema   *ToolsOpenAI_completion_tool
	FileTime int64
}

type ToolsCmd struct {
	router     *ToolsRouter
	folderCode string

	lock sync.Mutex

	port int

	cmd        *exec.Cmd
	cmd_exited bool
	cmd_error  string

	Compile_error string
	CodeHash      int64
	Tools         map[string]*ToolsCmdItem
}

func NewToolsCmd(folderCode string, router *ToolsRouter) *ToolsCmd {
	tools := &ToolsCmd{folderCode: folderCode, router: router}
	tools.Tools = make(map[string]*ToolsCmdItem)

	fl, err := os.ReadFile(tools.GetToolsJsonPath())
	if err == nil {
		json.Unmarshal(fl, tools)
	}

	//hot reload
	go func() {
		for {
			err = tools._hotReload()
			router.log.Error(err)

			time.Sleep(1000 * time.Millisecond)
		}
	}()

	return tools
}

func (tools *ToolsCmd) Destroy() error {
	if tools.IsRunning() {
		cl, err := NewToolsClient("localhost", tools.port)
		if err != nil {
			return err
		}
		err = cl.WriteArray([]byte("exit"))
		if err != nil {
			return err
		}
	}

	return nil
}

func (tools *ToolsCmd) _save() error {
	tools.lock.Lock()
	defer tools.lock.Unlock()

	_, err := Tools_WriteJSONFile(tools.GetToolsJsonPath(), tools)
	if err != nil {
		return err
	}

	return nil

}

func (tools *ToolsCmd) GetToolsJsonPath() string {
	return filepath.Join(tools.folderCode, "tools.json")
}

func (totoolsl *ToolsCmd) GetGenGoFileName() string {
	return "a_gen.go"
}

func (tools *ToolsCmd) getFileName(toolName string) string {
	return "z" + toolName + ".go"
}
func (tools *ToolsCmd) getFilePath(toolName string) string {
	return filepath.Join(tools.folderCode, tools.getFileName(toolName))
}

func (tools *ToolsCmd) WaitUntilExited() string {
	n := 0
	for n < 100 && !tools.cmd_exited {
		time.Sleep(10 * time.Millisecond)
		n++
	}
	return tools.cmd_error
}

func (tools *ToolsCmd) IsRunning() bool {
	return tools.cmd != nil && !tools.cmd_exited
}

func (tools *ToolsCmd) GetSchemas() []*ToolsOpenAI_completion_tool {
	tools.lock.Lock()
	defer tools.lock.Unlock()

	var oai_list []*ToolsOpenAI_completion_tool

	for _, it := range tools.Tools {
		oai_list = append(oai_list, it.Schema)
	}
	return oai_list
}

func (tools *ToolsCmd) _hotReload() error {

	saveIt := false

	files, err := os.ReadDir(tools.folderCode)
	if err != nil {
		return err
	}

	//add new tools
	codeHash := int64(0)
	for _, info := range files {

		if info.IsDir() || filepath.Ext(info.Name()) != ".go" || info.Name() == tools.GetGenGoFileName() {
			continue
		}

		var fileTime int64
		inf, err := os.Stat(filepath.Join(tools.folderCode, info.Name()))
		if err != nil {
			return err
		}
		if inf != nil {
			fileTime = inf.ModTime().UnixNano()
		}
		codeHash += fileTime

		if !strings.HasPrefix(info.Name(), "z") {
			continue
		}

		toolName, _ := strings.CutSuffix(info.Name()[1:], ".go") //remove 'z' and '.go'
		item, found := tools.Tools[toolName]
		if !found {
			//add
			schema, err := tools.GetSchema(toolName)
			if err != nil {
				return err
			}

			if schema != nil { //not ignored
				tools.lock.Lock()
				tools.Tools[toolName] = &ToolsCmdItem{Schema: schema, FileTime: fileTime}
				tools.lock.Unlock()
				saveIt = true
			}

		} else {
			if item.FileTime != fileTime {
				//update
				schema, err := tools.GetSchema(toolName)
				if err != nil {
					return err
				}

				if schema != nil { //not ignored
					tools.lock.Lock()
					item.Schema = schema
					item.FileTime = fileTime
					tools.lock.Unlock()
					saveIt = true
				}
			}
		}

	}
	//remove deleted tools
	for toolName := range tools.Tools {
		found := false
		for _, file := range files {
			if !file.IsDir() && file.Name() == tools.getFileName(toolName) {
				found = true
				break
			}
		}
		if !found {
			tools.lock.Lock()
			delete(tools.Tools, toolName)
			tools.lock.Unlock()
			saveIt = true
		}
	}

	//remove old bins
	if tools.IsRunning() && tools.Compile_error == "" {
		exclude := tools.getBinName()
		for _, info := range files {
			if info.IsDir() || filepath.Ext(info.Name()) != ".bin" || info.Name() == exclude {
				continue
			}
			os.Remove(filepath.Join(tools.folderCode, info.Name()))
		}
	}

	if tools.CodeHash != codeHash || (tools.Compile_error == "" && !Tools_FileExists(filepath.Join(tools.folderCode, tools.getBinName()))) {
		tools.Compile_error = ""
		err = tools.compile(codeHash)
		if err != nil {
			tools.Compile_error = err.Error()
		} else {
			tools.Destroy() //stop it - CheckRun() will start lastest bin

			err = tools.router.tools.CheckRun()
			tools.router.log.Error(err)
		}

		saveIt = true
	}

	if saveIt {
		//save 'tools.json'
		tools._save()
	}

	return nil
}

func (tools *ToolsCmd) getBinName() string {
	return strconv.FormatInt(tools.CodeHash, 10) + ".bin"
}

func (tools *ToolsCmd) CheckRun() error {
	if !tools.IsRunning() {

		tools.lock.Lock()
		defer tools.lock.Unlock()

		if tools.Compile_error != "" {
			return errors.New(tools.cmd_error)
		}

		if tools.cmd_exited {
			tools.WaitUntilExited()
		}

		tools.cmd_exited = false
		tools.cmd_error = ""
		tools.port = 0
		tools.cmd = nil

		//start
		cmd := exec.Command(filepath.Join("..", tools.folderCode)+"/./"+tools.getBinName(), strconv.Itoa(tools.router.server.port))
		cmd.Dir = tools.router.files.folder
		OutStr := new(strings.Builder)
		ErrStr := new(strings.Builder)
		cmd.Stdout = OutStr
		cmd.Stderr = ErrStr
		fmt.Printf("Tools started\n")
		err := cmd.Start()
		if err != nil {
			return err
		}
		tools.cmd = cmd //running

		//run tool
		go func() {
			tools.cmd.Wait()

			if OutStr.Len() > 0 {
				fmt.Printf("Tools output: %s\n", OutStr.String())
			}
			if ErrStr.Len() > 0 {
				fmt.Printf("\033[31mTools error:%s\033[0m\n", ErrStr.String())
			}

			wd, _ := os.Getwd()
			tools.cmd_error = strings.ReplaceAll(ErrStr.String(), wd, "")
			tools.cmd_exited = true
			tools.cmd = nil
		}()
	}

	//wait one second for recv a port
	{
		n := 0
		for n < 100 && tools.port == 0 {
			time.Sleep(10 * time.Millisecond)
			n++
		}
		if tools.port == 0 {
			fmt.Printf("Tools process hasn't connected in time\n")
		}
	}

	return nil //ok
}

func (tools *ToolsCmd) compile(codeHash int64) error {
	tools.lock.Lock()
	defer tools.lock.Unlock()

	tools.CodeHash = codeHash

	msg := tools.router.AddRecompileMsg()
	defer msg.Done()

	msg.progress_label = "Generating tools code"
	{
		var strInits strings.Builder
		var strFrees strings.Builder
		var strCalls strings.Builder

		//start
		strInits.WriteString("func _callGlobalInits() {\n\tvar err error\n")
		strFrees.WriteString("func _callGlobalDestroys() {\n\tvar err error\n")
		strCalls.WriteString("func FindToolRunFunc(funcName string, jsParams []byte) (func(caller *ToolCaller, ui *UI) error, interface{}) {\n\tswitch funcName {\n")

		files, err := os.ReadDir(tools.folderCode)
		if err != nil {
			return err
		}
		for _, info := range files {
			if info.IsDir() || filepath.Ext(info.Name()) != ".go" || !strings.HasPrefix(info.Name(), "z") {
				continue
			}

			stName := info.Name()[1 : len(info.Name())-3]

			fl, err := os.ReadFile(filepath.Join(tools.folderCode, info.Name()))
			if err != nil {
				return err
			}
			flstr := string(fl)

			//add init
			if strings.Index(flstr, stName+"_global_init") > 0 {
				strInits.WriteString(fmt.Sprintf(`
		err = %s_global_init()
		if err != nil {
			log.Fatal(err)
		}
			`, stName))
			}

			//add destroy
			if strings.Index(flstr, stName+"_global_destroy") > 0 {
				strFrees.WriteString(fmt.Sprintf(`
		defer func() {
			err = %s_global_destroy()
			if err != nil {
				log.Fatal(err)
			}
		}()
			`, stName))
			}

			//add call
			strCalls.WriteString(fmt.Sprintf(`	case "%s":
				st := %s{}
				err := json.Unmarshal(jsParams, &st)
				if err != nil {
					return nil, nil
				}
				return st.run, &st
		`, stName, stName))

		}

		//finish
		strInits.WriteString("}\n")
		strFrees.WriteString("}\n")
		strCalls.WriteString("\n\t}\n\treturn nil, nil\n}\n")

		var strFinal strings.Builder
		strFinal.WriteString(`package main
import (
	"encoding/json"
	"log"
)
`)
		strFinal.WriteString(strInits.String())
		strFinal.WriteString(strFrees.String())
		strFinal.WriteString(strCalls.String())

		err = os.WriteFile(filepath.Join(tools.folderCode, tools.GetGenGoFileName()), []byte(strFinal.String()), 0644)
		if err != nil {
			return err
		}
	}
	msg.progress_done = 0.2

	//fix files
	msg.progress_label = "Fixing tools code"
	{
		fmt.Printf("Fixing ... ")
		st := float64(time.Now().UnixMilli()) / 1000
		cmd := exec.Command("goimports", "-l", "-w", ".")
		cmd.Dir = tools.folderCode
		var stderr bytes.Buffer
		cmd.Stderr = &stderr //os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("goimports failed: %s", stderr.String())
		}
		fmt.Printf("done in %.3fsec\n", (float64(time.Now().UnixMilli())/1000)-st)
	}

	msg.progress_done = 0.4

	//update packages
	msg.progress_label = "Updating tools packages"
	{
		fmt.Printf("Updating packages ... ")
		st := float64(time.Now().UnixMilli()) / 1000

		if !Tools_FileExists(filepath.Join(tools.folderCode, "go.mod")) {
			//create
			cmd := exec.Command("go", "mod", "init", "skyalt_tools")
			cmd.Dir = tools.folderCode
			var stderr bytes.Buffer
			cmd.Stderr = &stderr //os.Stderr
			cmd.Stdout = os.Stdout
			err := cmd.Run()
			if err != nil {
				return fmt.Errorf("go mod init failed: %s", stderr.String())
			}
		}

		//update
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = tools.folderCode
		var stderr bytes.Buffer
		cmd.Stderr = &stderr //os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("go mod tidy failed: %s", stderr.String())
		}

		fmt.Printf("done in %.3fsec\n", (float64(time.Now().UnixMilli())/1000)-st)
	}
	msg.progress_done = 0.6

	//compile
	msg.progress_label = "Compiling tools code"
	{
		fmt.Printf("Compiling ... ")
		st := float64(time.Now().UnixMilli()) / 1000
		cmd := exec.Command("go", "build", "-o", tools.getBinName())
		cmd.Dir = tools.folderCode
		var stderr bytes.Buffer
		cmd.Stderr = &stderr //os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("compiler failed: %s", stderr.String())
		}
		fmt.Printf("done in %.3fsec\n", (float64(time.Now().UnixMilli())/1000)-st)
	}
	msg.progress_done = 1.0

	return nil
}

// Function was copied from Server code
func _ToolsCaller_CallChange(port int, msg_id uint64, ui_uid uint64, change SdkChange, fnLog func(err error) error) ([]ToolCmd, error) {
	cl, err := NewToolsClient("localhost", port)
	if fnLog(err) == nil {
		defer cl.Destroy()

		//send
		err = cl.WriteArray([]byte("change"))
		if fnLog(err) == nil {
			err = cl.WriteInt(msg_id)
			if fnLog(err) == nil {
				err = cl.WriteInt(ui_uid)
				if fnLog(err) == nil {
					changeJs, err := json.Marshal(change)
					if fnLog(err) == nil {
						err = cl.WriteArray(changeJs)
						if fnLog(err) == nil {

							errStr, err := cl.ReadArray()
							if fnLog(err) == nil {
								cmdsJs, err := cl.ReadArray()
								fnLog(err)

								var cmds []ToolCmd
								json.Unmarshal(cmdsJs, &cmds)

								if len(errStr) > 0 {
									return nil, errors.New(string(errStr))
								}

								return cmds, nil
							}
						}
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("connection failed")
}

// Function was copied from Server code
func _ToolsCaller_CallTool(port int, msg_id uint64, next_ui_uid uint64, funcName string, params []byte, fnLog func(err error) error) ([]byte, *UI, []ToolCmd, error) {
	cl, err := NewToolsClient("localhost", port)
	if fnLog(err) == nil {
		defer cl.Destroy()

		//send
		err = cl.WriteArray([]byte("call"))
		if fnLog(err) == nil {
			err = cl.WriteInt(msg_id) //msg_id
			if fnLog(err) == nil {
				err = cl.WriteInt(next_ui_uid) //UI UID
				if fnLog(err) == nil {
					err = cl.WriteArray([]byte(funcName)) //function name
					if fnLog(err) == nil {
						err = cl.WriteArray(params) //params
						if fnLog(err) == nil {

							errStr, err := cl.ReadArray() //output error
							if fnLog(err) == nil {
								out_data, err := cl.ReadArray() //output data
								if fnLog(err) == nil {
									out_ui, err := cl.ReadArray() //output UI
									if fnLog(err) == nil {
										out_cmds, err := cl.ReadArray() //output cmds
										if fnLog(err) == nil {
											var ui UI
											json.Unmarshal(out_ui, &ui)

											var cmds []ToolCmd
											json.Unmarshal(out_cmds, &cmds)

											var out_err error
											if len(errStr) > 0 {
												out_err = errors.New(string(errStr))
											}

											return out_data, &ui, cmds, out_err
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

	return nil, nil, nil, fmt.Errorf("connection failed")
}
