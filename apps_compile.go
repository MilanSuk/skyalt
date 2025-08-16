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

	Error       string
	SdkFileTime int64
	AppFileTime int64
}

func NewToolsAppCompile(appName string) *ToolsAppCompile {
	cmpl := &ToolsAppCompile{appName: appName}
	return cmpl
}

func (cmpl *ToolsAppCompile) GetFolderPath() string {
	return filepath.Join("apps", cmpl.appName)
}

func (cmpl *ToolsAppCompile) GetBinName() string {
	return strconv.FormatInt(cmpl.AppFileTime, 10) + ".bin"
}
func (cmpl *ToolsAppCompile) GetBinPath() string {
	return filepath.Join(cmpl.GetFolderPath(), cmpl.GetBinName())
}

func (cmpl *ToolsAppCompile) RemoveOldBins() error {
	if cmpl.Error == "" {
		exclude := cmpl.GetBinName()

		files, err := os.ReadDir(cmpl.GetFolderPath())
		if err != nil {
			return err
		}
		for _, info := range files {
			if info.IsDir() || filepath.Ext(info.Name()) != ".bin" || info.Name() == exclude {
				continue
			}
			os.Remove(filepath.Join(cmpl.GetFolderPath(), info.Name()))
		}
	}
	return nil
}

func (cmpl *ToolsAppCompile) BuildMainFile(prompts []*ToolsPrompt) error {
	var strInits strings.Builder
	var strFrees strings.Builder
	var strCalls strings.Builder

	//start
	strInits.WriteString("func _callGlobalInits() {\n\n")
	strFrees.WriteString("func _callGlobalDestroys() {\n\n")
	strCalls.WriteString("func FindToolRunFunc(toolName string, jsParams []byte) (func(caller *ToolCaller, ui *UI) error, interface{}, error) {\n\tswitch toolName {\n")

	for _, prompt := range prompts {
		if prompt.Type != ToolsPrompt_TOOL {
			continue
		}

		fl, err := os.ReadFile(filepath.Join(cmpl.GetFolderPath(), prompt.Name+".go"))
		if err != nil {
			cmpl.Error = err.Error()
			return err
		}
		flstr := string(fl)

		//add init
		if strings.Index(flstr, prompt.Name+"_global_init") > 0 {
			strInits.WriteString(fmt.Sprintf(`
	{
		err := %s_global_init()
		if err != nil {
			log.Fatal(err)
		}
	}
`, prompt.Name))
		}

		//add destroy
		if strings.Index(flstr, prompt.Name+"_global_destroy") > 0 {
			strFrees.WriteString(fmt.Sprintf(`
	{
		err := %s_global_destroy()
		if err != nil {
			log.Fatal(err)
		}
	}
`, prompt.Name))
		}

		//add call
		strCalls.WriteString(fmt.Sprintf(`	case "%s":
				st := %s{}
				err := json.Unmarshal(jsParams, &st)
				if err != nil {
					return nil, nil, err
				}
				return st.run, &st, nil
		`, prompt.Name, prompt.Name))

	}

	//finish
	strInits.WriteString("}\n")
	strFrees.WriteString("}\n")
	strCalls.WriteString("\n\t}\n\treturn nil, nil, fmt.Errorf(\"Function '%s' not found\", toolName)\n}\n")

	var strFinal strings.Builder

	mainGo, err := os.ReadFile("sdk/sdk.go")
	if err != nil {
		cmpl.Error = err.Error()
		return err
	}
	strFinal.WriteString(string(mainGo))
	strFinal.WriteString("\n")
	strFinal.WriteString(strInits.String())
	strFinal.WriteString(strFrees.String())
	strFinal.WriteString(strCalls.String())

	err = os.WriteFile(filepath.Join(cmpl.GetFolderPath(), "main.go"), []byte(strFinal.String()), 0644)
	if err != nil {
		cmpl.Error = err.Error()
		return err
	}
	return nil
}

func (cmpl *ToolsAppCompile) _compile(sdkFileTime, appFileTime int64, noBinary bool, msg *AppsRouterMsg) ([]ToolsCodeError, error) {

	cmpl.Error = ""

	cmpl.SdkFileTime = sdkFileTime
	cmpl.AppFileTime = appFileTime

	//fix files
	msg.progress_label = "Fixing tools code " + cmpl.GetFolderPath()
	{
		fmt.Printf("Fixing '%s' ...\n", cmpl.GetFolderPath())
		st := float64(time.Now().UnixMilli()) / 1000
		cmd := exec.Command("goimports", "-l", "-w", ".")
		cmd.Dir = cmpl.GetFolderPath()
		var stderr bytes.Buffer
		cmd.Stderr = &stderr //os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run() //can rewrite the file
		if err != nil {

			var codeErrors []ToolsCodeError
			lines := strings.Split(stderr.String(), "\n")
			for _, line := range lines {
				codeErrLn, err := _ToolsAppCompile_parseErrorString(line)
				if err == nil {
					codeErrors = append(codeErrors, codeErrLn)
				}
			}

			cmpl.Error = stderr.String()
			return codeErrors, nil
		}
		fmt.Printf("Fixing '%s' done in %.3fsec\n", cmpl.GetFolderPath(), (float64(time.Now().UnixMilli())/1000)-st)
	}

	msg.progress_done = 0.25

	//update packages
	msg.progress_label = "Updating tools packages " + cmpl.GetFolderPath()
	{
		fmt.Printf("Updating packages '%s' ...\n", cmpl.GetFolderPath())
		st := float64(time.Now().UnixMilli()) / 1000

		if !Tools_IsFileExists(filepath.Join(cmpl.GetFolderPath(), "go.mod")) {
			//create
			cmd := exec.Command("go", "mod", "init", "skyalt_tool")
			cmd.Dir = cmpl.GetFolderPath()
			var stderr bytes.Buffer
			cmd.Stderr = &stderr //os.Stderr
			cmd.Stdout = os.Stdout
			err := cmd.Run()
			if err != nil {
				err = LogsErrorf("go mod init failed: %s", stderr.String())
				cmpl.Error = stderr.String()
				return nil, err
			}
		}

		//update
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = cmpl.GetFolderPath()
		var stderr bytes.Buffer
		cmd.Stderr = &stderr //os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			err = LogsErrorf("go mod tidy failed: %s", stderr.String())
			cmpl.Error = err.Error()
			return nil, err
		}

		fmt.Printf("Updating '%s' done in %.3fsec\n", cmpl.GetFolderPath(), (float64(time.Now().UnixMilli())/1000)-st)
	}
	msg.progress_done = 0.5

	//compile
	msg.progress_label = "Compiling tools code " + cmpl.GetFolderPath()
	{
		fmt.Printf("Compiling '%s' ...\n", cmpl.GetFolderPath())
		st := float64(time.Now().UnixMilli()) / 1000

		outName := cmpl.GetBinName()
		if noBinary {
			outName = "/dev/null"
		}
		cmd := exec.Command("go", "build", "-gcflags=-e", "-o", outName) //(-gcflags="-e") = show all errors
		cmd.Dir = cmpl.GetFolderPath()
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

			cmpl.Error = stderr.String()
			return codeErrors, nil
		}
		fmt.Printf("Compiling '%s' done in %.3fsec\n", cmpl.GetFolderPath(), (float64(time.Now().UnixMilli())/1000)-st)
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
