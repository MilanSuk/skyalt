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
	"strings"
	"time"
)

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
