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

type ToolsAppItem struct {
	FileTime int64
	Schema   *ToolsOpenAI_completion_tool
}

type ToolsAppSource struct {
	FileTime    int64
	Description string
	Tools       []string
}

type ToolsApp struct {
	router *ToolsRouter
	folder string
	name   string

	lock sync.Mutex

	port int

	cmd        *exec.Cmd
	cmd_exited bool
	cmd_error  string

	Compile_error string
	CodeHash      int64
	Tools         map[string]*ToolsAppItem
	Sources       map[string]*ToolsAppSource

	storage_changes int64
}

func NewToolsApp(name string, folder string, router *ToolsRouter) *ToolsApp {
	app := &ToolsApp{name: name, folder: folder, router: router}
	app.Tools = make(map[string]*ToolsAppItem)
	app.Sources = make(map[string]*ToolsAppSource)
	app.Sources["_default_"] = &ToolsAppSource{}

	fl, err := os.ReadFile(app.GetToolsJsonPath())
	if err == nil {
		json.Unmarshal(fl, app)
	}

	return app
}

func (app *ToolsApp) Destroy() error {
	if app.IsRunning() {
		cl, err := NewToolsClient("localhost", app.port)
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

func (app *ToolsApp) _save() error {
	app.lock.Lock()
	defer app.lock.Unlock()

	_, err := Tools_WriteJSONFile(app.GetToolsJsonPath(), app)
	if err != nil {
		return err
	}

	return nil

}

func (app *ToolsApp) GetToolsJsonPath() string {
	return filepath.Join(app.folder, "tools.json")
}

func (app *ToolsApp) getToolFileName(toolName string) string {
	return "z" + toolName + ".go"
}
func (app *ToolsApp) getSourceFileName(sourceName string) string {
	return sourceName + ".go"
}

func (app *ToolsApp) getToolFilePath(toolName string) string {
	return filepath.Join(app.folder, app.getToolFileName(toolName))
}
func (app *ToolsApp) getSourceFilePath(sourceName string) string {
	return filepath.Join(app.folder, app.getSourceFileName(sourceName))
}

func (app *ToolsApp) WaitUntilExited() string {
	n := 0
	for n < 100 && !app.cmd_exited {
		time.Sleep(10 * time.Millisecond)
		n++
	}
	return app.cmd_error
}

func (app *ToolsApp) IsRunning() bool {
	return app.cmd != nil && !app.cmd_exited
}

func (app *ToolsApp) GetAllSchemas() []*ToolsOpenAI_completion_tool {
	app.lock.Lock()
	defer app.lock.Unlock()

	var schemas []*ToolsOpenAI_completion_tool

	for _, tool := range app.Tools {
		schemas = append(schemas, tool.Schema)
	}

	return schemas
}

func (app *ToolsApp) GetSchemas(toolNames []string) []*ToolsOpenAI_completion_tool {
	app.lock.Lock()
	defer app.lock.Unlock()

	var schemas []*ToolsOpenAI_completion_tool

	for _, toolName := range toolNames {
		tool, found := app.Tools[toolName]
		if found {
			schemas = append(schemas, tool.Schema)
		}
	}

	return schemas
}
func (app *ToolsApp) GetSchemasForSource(sourceName string) []*ToolsOpenAI_completion_tool {
	app.lock.Lock()
	defer app.lock.Unlock()

	var schemas []*ToolsOpenAI_completion_tool

	source, found := app.Sources[sourceName]
	if found {
		for _, toolName := range source.Tools {
			tool, found := app.Tools[toolName]
			if found {
				schemas = append(schemas, tool.Schema)
			}
		}
	}

	return schemas
}

func (app *ToolsApp) Tick() error {

	saveIt := false

	files, err := os.ReadDir(app.folder)
	if err != nil {
		return err
	}

	//create list of sources
	{
		//add new sources
		for _, info := range files {
			if info.IsDir() || filepath.Ext(info.Name()) != ".go" || info.Name() == "" || info.Name()[0] != strings.ToUpper(info.Name())[0] {
				continue
			}

			var fileTime int64
			inf, err := os.Stat(filepath.Join(app.folder, info.Name()))
			if err != nil {
				return err
			}
			if inf != nil {
				fileTime = inf.ModTime().UnixNano()
			}

			sourceName := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))

			item, found := app.Sources[sourceName]
			if !found {
				description, err := app.GetSourceDescription(sourceName)
				if err != nil {
					return err
				}

				app.lock.Lock()
				app.Sources[sourceName] = &ToolsAppSource{Description: description, FileTime: fileTime}
				app.lock.Unlock()

				saveIt = true
			} else {
				if item.FileTime != fileTime {
					//update
					description, err := app.GetSourceDescription(sourceName)
					if err != nil {
						return err
					}

					app.lock.Lock()
					item.Description = description
					item.FileTime = fileTime
					app.lock.Unlock()

					saveIt = true
				}
			}
		}

		//remove deleted sources
		for sourceName := range app.Sources {
			if sourceName == "_default_" {
				continue //skip
			}

			found := false
			for _, file := range files {
				if !file.IsDir() && file.Name() == app.getSourceFileName(sourceName) {
					found = true
					break
				}
			}
			if !found {
				app.lock.Lock()
				delete(app.Sources, sourceName)
				app.lock.Unlock()
				saveIt = true
			}
		}
	}

	//add new tools
	codeHash := int64(0)
	//main.go
	{
		var fileTime int64
		inf, err := os.Stat(filepath.Join(app.router.folderApps, "main.go"))
		if err != nil {
			return err
		}
		if inf != nil {
			fileTime = inf.ModTime().UnixNano()
		}
		codeHash += fileTime

	}
	for _, info := range files {

		if info.IsDir() || filepath.Ext(info.Name()) != ".go" || info.Name() == "main.go" {
			continue
		}

		var fileTime int64
		inf, err := os.Stat(filepath.Join(app.folder, info.Name()))
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
		item, found := app.Tools[toolName]
		if !found {
			//add
			schema, err := app.GetToolSchema(toolName)
			if err != nil {
				return err
			}

			if schema != nil { //not ignored
				app.lock.Lock()
				app.Tools[toolName] = &ToolsAppItem{Schema: schema, FileTime: fileTime}
				app.lock.Unlock()

				err := app.UpdateSources(toolName)
				if err != nil {
					return err
				}

				saveIt = true
			}

		} else {
			if item.FileTime != fileTime {
				//update
				schema, err := app.GetToolSchema(toolName)
				if err != nil {
					return err
				}

				if schema != nil { //not ignored
					app.lock.Lock()
					item.Schema = schema
					item.FileTime = fileTime
					app.lock.Unlock()

					err := app.UpdateSources(toolName)
					if err != nil {
						return err
					}

					saveIt = true
				}
			}
		}
	}

	//remove deleted tools
	for toolName := range app.Tools {
		found := false
		for _, file := range files {
			if !file.IsDir() && file.Name() == app.getToolFileName(toolName) {
				found = true
				break
			}
		}
		if !found {
			app.lock.Lock()
			delete(app.Tools, toolName)
			app.lock.Unlock()
			saveIt = true
		}
	}

	if app.CodeHash != codeHash || (app.Compile_error == "" && !Tools_FileExists(filepath.Join(app.folder, app.getBinName()))) {
		app.Compile_error = ""
		err = app.compile(codeHash)
		if err == nil {
			app.Destroy() //stop it

			//remove old bins
			if app.Compile_error == "" {
				exclude := app.getBinName()
				for _, info := range files {
					if info.IsDir() || filepath.Ext(info.Name()) != ".bin" || info.Name() == exclude {
						continue
					}
					os.Remove(filepath.Join(app.folder, info.Name()))
				}
			}

			//err = app.CheckRun()
			//app.router.log.Error(err)
		} else {
			app.Compile_error = err.Error()
		}

		saveIt = true
	}

	if saveIt {
		//save 'tools.json'
		app._save()
	}

	return nil
}

func (app *ToolsApp) getBinName() string {
	return strconv.FormatInt(app.CodeHash, 10) + ".bin"
}

func (app *ToolsApp) CheckRun() error {
	if !app.IsRunning() {

		app.lock.Lock()
		defer app.lock.Unlock()

		if app.Compile_error != "" {
			return fmt.Errorf("'%s' app has compilation error: %s", app.name, app.cmd_error)
		}

		if app.CodeHash == 0 {
			return fmt.Errorf("'%s' app is waiting for compilation", app.name)
		}

		if app.cmd_exited {
			app.WaitUntilExited()
		}

		app.cmd_exited = false
		app.cmd_error = ""
		app.port = 0
		app.cmd = nil

		//start
		cmd := exec.Command("./"+app.getBinName(), app.name, strconv.Itoa(app.router.server.port))
		cmd.Dir = app.folder
		OutStr := new(strings.Builder)
		ErrStr := new(strings.Builder)
		cmd.Stdout = OutStr
		cmd.Stderr = ErrStr
		err := cmd.Start()
		if err != nil {
			return fmt.Errorf("'%s' start failed: %w", app.name, err)
		}
		app.cmd = cmd //running

		fmt.Printf("App '%s' has started\n", app.name)

		//run tool
		go func() {
			app.cmd.Wait()

			if OutStr.Len() > 0 {
				fmt.Printf("'%s' app output: %s\n", app.name, OutStr.String())
			}
			if ErrStr.Len() > 0 {
				fmt.Printf("\033[31m'%s' app error:%s\033[0m\n", app.name, ErrStr.String())
			}

			wd, _ := os.Getwd()
			app.cmd_error = strings.ReplaceAll(ErrStr.String(), wd, "")
			app.cmd_exited = true
			app.cmd = nil
		}()

		//wait one second for recv a port
		{
			n := 0
			for n < 100 && app.port == 0 {
				time.Sleep(10 * time.Millisecond)
				n++
			}
			if app.port == 0 {
				fmt.Printf("'%s' app process hasn't connected in time\n", app.name)
			}
		}

	}

	return nil //ok
}

func (app *ToolsApp) compile(codeHash int64) error {
	app.lock.Lock()
	defer app.lock.Unlock()

	app.CodeHash = codeHash

	msg := app.router.AddRecompileMsg(app.name)
	defer msg.Done()

	msg.progress_label = "Generating tools code"
	{
		var strInits strings.Builder
		var strFrees strings.Builder
		var strCalls strings.Builder

		//start
		strInits.WriteString("func _callGlobalInits() {\n\n")
		strFrees.WriteString("func _callGlobalDestroys() {\n\n")
		strCalls.WriteString("func FindToolRunFunc(funcName string, jsParams []byte) (func(caller *ToolCaller, ui *UI) error, interface{}, error) {\n\tswitch funcName {\n")

		files, err := os.ReadDir(app.folder)
		if err != nil {
			return err
		}
		for _, info := range files {
			if info.IsDir() || filepath.Ext(info.Name()) != ".go" || !strings.HasPrefix(info.Name(), "z") {
				continue
			}

			stName := info.Name()[1 : len(info.Name())-3]

			fl, err := os.ReadFile(filepath.Join(app.folder, info.Name()))
			if err != nil {
				return err
			}
			flstr := string(fl)

			//add init
			if strings.Index(flstr, stName+"_global_init") > 0 {
				strInits.WriteString(fmt.Sprintf(`
	{
		err := %s_global_init()
		if err != nil {
			log.Fatal(err)
		}
	}
`, stName))
			}

			//add destroy
			if strings.Index(flstr, stName+"_global_destroy") > 0 {
				strFrees.WriteString(fmt.Sprintf(`
	{
		err := %s_global_destroy()
		if err != nil {
			log.Fatal(err)
		}
	}
`, stName))
			}

			//add call
			strCalls.WriteString(fmt.Sprintf(`	case "%s":
				st := %s{}
				err := json.Unmarshal(jsParams, &st)
				if err != nil {
					return nil, nil, err
				}
				return st.run, &st, nil
		`, stName, stName))

		}

		//finish
		strInits.WriteString("}\n")
		strFrees.WriteString("}\n")
		strCalls.WriteString("\n\t}\n\treturn nil, nil, fmt.Errorf(\"Function '%s' not found\", funcName)\n}\n")

		var strFinal strings.Builder
		/*strFinal.WriteString(`package main
		import (
			"encoding/json"
			"log"
		)
		`)*/

		mainGo, err := os.ReadFile(filepath.Join(app.router.folderApps, "main.go"))
		if err != nil {
			return err
		}
		strFinal.WriteString(string(mainGo))
		strFinal.WriteString("\n")
		strFinal.WriteString(strInits.String())
		strFinal.WriteString(strFrees.String())
		strFinal.WriteString(strCalls.String())

		err = os.WriteFile(filepath.Join(app.folder, "main.go"), []byte(strFinal.String()), 0644)
		if err != nil {
			return err
		}
	}
	msg.progress_done = 0.2

	//fix files
	msg.progress_label = "Fixing tools code"
	{
		fmt.Printf("Fixing '%s' ... ", app.name)
		st := float64(time.Now().UnixMilli()) / 1000
		cmd := exec.Command("goimports", "-l", "-w", ".")
		cmd.Dir = app.folder
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
		fmt.Printf("Updating packages '%s' ... ", app.name)
		st := float64(time.Now().UnixMilli()) / 1000

		if !Tools_FileExists(filepath.Join(app.folder, "go.mod")) {
			//create
			cmd := exec.Command("go", "mod", "init", "skyalt_tool")
			cmd.Dir = app.folder
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
		cmd.Dir = app.folder
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
		fmt.Printf("Compiling '%s' ... ", app.name)
		st := float64(time.Now().UnixMilli()) / 1000
		cmd := exec.Command("go", "build", "-o", app.getBinName())
		cmd.Dir = app.folder
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

func _ToolsCaller_UpdateDev(port int, fnLog func(err error) error) error {
	cl, err := NewToolsClient("localhost", port)
	if fnLog(err) == nil {
		defer cl.Destroy()

		//send
		err = cl.WriteArray([]byte("update_dev"))
		fnLog(err)
	}

	return fmt.Errorf("connection failed")
}

// Function was copied from Server code
func _ToolsCaller_CallChange(port int, msg_id uint64, ui_uid uint64, change SdkChange, fnLog func(err error) error) ([]byte, []byte, error) {
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
								dataJs, err := cl.ReadArray()
								if fnLog(err) == nil {
									cmdsJs, err := cl.ReadArray()
									fnLog(err)

									if len(errStr) > 0 {
										return nil, nil, errors.New(string(errStr))
									}

									return dataJs, cmdsJs, nil
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

// Function was copied from Server code
func _ToolsCaller_CallTool(port int, msg_id uint64, ui_uid uint64, funcName string, params []byte, fnLog func(err error) error) ([]byte, []byte, []byte, error) {
	cl, err := NewToolsClient("localhost", port)
	if fnLog(err) == nil {
		defer cl.Destroy()

		//send
		err = cl.WriteArray([]byte("call"))
		if fnLog(err) == nil {
			err = cl.WriteInt(msg_id) //msg_id
			if fnLog(err) == nil {

				err = cl.WriteInt(ui_uid) //UI UID
				if fnLog(err) == nil {
					err = cl.WriteArray([]byte(funcName)) //function name
					if fnLog(err) == nil {
						err = cl.WriteArray(params) //params
						if fnLog(err) == nil {

							errStr, err := cl.ReadArray() //output error
							if fnLog(err) == nil {
								out_dataJs, err := cl.ReadArray() //output data
								if fnLog(err) == nil {
									out_uiJs, err := cl.ReadArray() //output UI
									if fnLog(err) == nil {
										out_cmdsJs, err := cl.ReadArray() //output cmds
										if fnLog(err) == nil {

											var out_err error
											if len(errStr) > 0 {
												out_err = errors.New(string(errStr))
											}

											return out_dataJs, out_uiJs, out_cmdsJs, out_err
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
