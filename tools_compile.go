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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type ToolsCodeError struct {
	File string
	Line int
	Col  int
	Msg  string
}

type ToolsAppCompile struct {
	appName string

	Error    string
	CodeHash int64
}

func NewToolsAppCompile(appName string) *ToolsAppCompile {
	app := &ToolsAppCompile{appName: appName}
	return app
}

func (app *ToolsAppCompile) GetFolderPath() string {
	return filepath.Join("apps", app.appName)
}

func (app *ToolsAppCompile) GetBinName() string {
	return strconv.FormatInt(app.CodeHash, 10) + ".bin"
}

func (app *ToolsAppCompile) NeedCompile(codeHash int64) bool {
	return app.CodeHash != codeHash || (app.Error == "" && !Tools_FileExists(filepath.Join(app.GetFolderPath(), app.GetBinName())))
}

func (app *ToolsAppCompile) Compile(codeHash int64, router *ToolsRouter, stopProcess func() error) ([]ToolsCodeError, error) {

	codeErrors, err := app._compile(codeHash, router)
	if err == nil {
		err := stopProcess() //stop it
		if err != nil {
			return nil, err
		}

		//remove old bins
		if app.Error == "" {
			exclude := app.GetBinName()

			files, err := os.ReadDir(app.GetFolderPath())
			if err != nil {
				return nil, err
			}
			for _, info := range files {
				if info.IsDir() || filepath.Ext(info.Name()) != ".bin" || info.Name() == exclude {
					continue
				}
				os.Remove(filepath.Join(app.GetFolderPath(), info.Name()))
			}
		}
	} else {
		app.Error = err.Error()
	}

	return codeErrors, err
}

func (app *ToolsAppCompile) _compile(codeHash int64, router *ToolsRouter) ([]ToolsCodeError, error) {

	app.Error = ""

	app.CodeHash = codeHash

	msg := router.AddRecompileMsg(app.appName)
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

		files, err := os.ReadDir(app.GetFolderPath())
		if err != nil {
			return nil, err
		}
		for _, info := range files {
			if info.IsDir() || filepath.Ext(info.Name()) != ".go" || !strings.HasPrefix(info.Name(), "z") {
				continue
			}

			stName := info.Name()[1 : len(info.Name())-3]

			fl, err := os.ReadFile(filepath.Join(app.GetFolderPath(), info.Name()))
			if err != nil {
				return nil, err
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

		mainGo, err := os.ReadFile("apps/main.go")
		if err != nil {
			return nil, err
		}
		strFinal.WriteString(string(mainGo))
		strFinal.WriteString("\n")
		strFinal.WriteString(strInits.String())
		strFinal.WriteString(strFrees.String())
		strFinal.WriteString(strCalls.String())

		err = os.WriteFile(filepath.Join(app.GetFolderPath(), "main.go"), []byte(strFinal.String()), 0644)
		if err != nil {
			return nil, err
		}
	}
	msg.progress_done = 0.2

	//fix files
	msg.progress_label = "Fixing tools code"
	{
		fmt.Printf("Fixing '%s' ... ", app.GetFolderPath())
		st := float64(time.Now().UnixMilli()) / 1000
		cmd := exec.Command("goimports", "-l", "-w", ".")
		cmd.Dir = app.GetFolderPath()
		var stderr bytes.Buffer
		cmd.Stderr = &stderr //os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {

			var codeErrors []ToolsCodeError
			lines := strings.Split(stderr.String(), "\n")
			for _, line := range lines {
				itErr, _ := _ToolsAppCompile_parseErrorString(line)
				codeErrors = append(codeErrors, itErr)
			}

			return codeErrors, fmt.Errorf("goimports failed: %s", stderr.String())
		}
		fmt.Printf("done in %.3fsec\n", (float64(time.Now().UnixMilli())/1000)-st)
	}

	msg.progress_done = 0.4

	//update packages
	msg.progress_label = "Updating tools packages"
	{
		fmt.Printf("Updating packages '%s' ... ", app.GetFolderPath())
		st := float64(time.Now().UnixMilli()) / 1000

		if !Tools_FileExists(filepath.Join(app.GetFolderPath(), "go.mod")) {
			//create
			cmd := exec.Command("go", "mod", "init", "skyalt_tool")
			cmd.Dir = app.GetFolderPath()
			var stderr bytes.Buffer
			cmd.Stderr = &stderr //os.Stderr
			cmd.Stdout = os.Stdout
			err := cmd.Run()
			if err != nil {
				return nil, fmt.Errorf("go mod init failed: %s", stderr.String())
			}
		}

		//update
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = app.GetFolderPath()
		var stderr bytes.Buffer
		cmd.Stderr = &stderr //os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			return nil, fmt.Errorf("go mod tidy failed: %s", stderr.String())
		}

		fmt.Printf("done in %.3fsec\n", (float64(time.Now().UnixMilli())/1000)-st)
	}
	msg.progress_done = 0.6

	//compile
	msg.progress_label = "Compiling tools code"
	{
		fmt.Printf("Compiling '%s' ... ", app.GetFolderPath())
		st := float64(time.Now().UnixMilli()) / 1000
		cmd := exec.Command("go", "build", "-o", app.GetBinName())
		cmd.Dir = app.GetFolderPath()
		var stderr bytes.Buffer
		cmd.Stderr = &stderr //os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			var codeErrors []ToolsCodeError
			lines := strings.Split(stderr.String(), "\n")
			for _, line := range lines {
				itErr, err := _ToolsAppCompile_parseErrorString(line)
				if err == nil {
					codeErrors = append(codeErrors, itErr)
				}
			}

			return codeErrors, fmt.Errorf("compiler failed: %s", stderr.String())
		}
		fmt.Printf("done in %.3fsec\n", (float64(time.Now().UnixMilli())/1000)-st)
	}
	msg.progress_done = 1.0

	return nil, nil
}

func _ToolsAppCompile_parseErrorString(errStr string) (ToolsCodeError, error) {
	// Split the string by colons
	parts := strings.SplitN(errStr, ":", 3)
	if len(parts) < 3 {
		return ToolsCodeError{Line: -1, Msg: errStr}, fmt.Errorf("invalid error string format: %s", errStr)
	}

	// Extract file name
	file := parts[0]

	// Extract line number
	lineStr := parts[1]
	line, err := strconv.Atoi(lineStr)
	if err != nil {
		return ToolsCodeError{Line: -1, Msg: errStr}, fmt.Errorf("invalid line number: %s", lineStr)
	}

	// Split the remaining part to get column and message
	remaining := parts[2]
	colonIndex := strings.Index(remaining, ":")
	if colonIndex == -1 {
		return ToolsCodeError{Line: -1, Msg: errStr}, fmt.Errorf("invalid error string format, missing column or message: %s", errStr)
	}

	// Extract column number
	colStr := strings.TrimSpace(remaining[:colonIndex])
	col, err := strconv.Atoi(colStr)
	if err != nil {
		return ToolsCodeError{Line: -1, Msg: errStr}, fmt.Errorf("invalid column number: %s", colStr)
	}

	// Extract message
	msg := strings.TrimSpace(remaining[colonIndex+1:])

	return ToolsCodeError{
		File: file,
		Line: line,
		Col:  col,
		Msg:  msg,
	}, nil
}
