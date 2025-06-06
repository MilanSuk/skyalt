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
	"path/filepath"
	"strings"
	"sync"
)

type ToolsSdkChange struct {
	UID         uint64
	ValueBytes  []byte
	ValueString string
	ValueFloat  float64
	ValueInt    int64
	ValueBool   bool
}

type ToolsAppItem struct {
	FileTime int64
	Schema   *ToolsOpenAI_completion_tool
}
type ToolsApp struct {
	router *ToolsRouter
	lock   sync.Mutex

	Process *ToolsAppProcess

	Prompts *ToolsPrompts

	Tools map[string]*ToolsAppItem

	storage_changes int64
}

func NewToolsApp(folderPath string, router *ToolsRouter) *ToolsApp {

	app := &ToolsApp{router: router}
	app.Tools = make(map[string]*ToolsAppItem)

	app.Process = NewToolsAppRun(folderPath)

	promptsFilePath := filepath.Join(folderPath, "skyalt")
	if Tools_FileExists(promptsFilePath) {
		app.Prompts = NewToolsPrompts()
	}

	fl, err := os.ReadFile(app.GetToolsJsonPath())
	if err == nil {
		json.Unmarshal(fl, app)
	}

	return app
}

func (app *ToolsApp) Destroy() error {
	err := app.Process.Destroy()
	if err != nil {
		return err
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
	return filepath.Join(app.Process.Compile.folderPath, "tools.json")
}

func (app *ToolsApp) getToolFileName(toolName string) string {
	return "z" + toolName + ".go"
}

func (app *ToolsApp) getToolFilePath(toolName string) string {
	return filepath.Join(app.Process.Compile.folderPath, app.getToolFileName(toolName))
}

func (app *ToolsApp) GetAllSchemas() []*ToolsOpenAI_completion_tool {
	app.lock.Lock()
	defer app.lock.Unlock()

	var schemas []*ToolsOpenAI_completion_tool

	if app.Prompts != nil {
		for _, prompt := range app.Prompts.Prompts {
			schemas = append(schemas, prompt.Schema)
		}
	} else {
		for _, tool := range app.Tools {
			schemas = append(schemas, tool.Schema)
		}
	}

	return schemas
}

func (app *ToolsApp) GetPrompt(toolName string) *ToolsPrompt {
	app.lock.Lock()
	defer app.lock.Unlock()

	if app.Prompts != nil {
		prompt := app.Prompts.FindPromptName(toolName)
		if prompt != nil {
			return prompt
		}
	}

	return nil
}

func (app *ToolsApp) CheckRun() error {
	app.lock.Lock()
	defer app.lock.Unlock()

	if app.Process.Compile.Error != "" {
		return fmt.Errorf("'%s' app has compilation error: %s", app.Process.Compile.folderPath, app.Process.cmd_error)
	}

	if app.Process.Compile.CodeHash == 0 {
		return fmt.Errorf("'%s' app is waiting for compilation", app.Process.Compile.folderPath)
	}

	return app.Process.CheckRun(app.router)
}

func (app *ToolsApp) Tick() error {

	saveIt := false

	if app.Prompts != nil {

		promptsFilePath := filepath.Join(app.Process.Compile.folderPath, "skyalt")
		codeHash := Tools_GetFileTime(promptsFilePath)

		if app.Process.Compile.NeedCompile(codeHash) {

			err := app.Prompts.Reload(app.Process.Compile.folderPath)
			if err != nil {
				return err
			}

			err = app.Prompts.Generate(app.Process.Compile.folderPath, app.router)
			if err != nil {
				return err
			}
			codeHash = Tools_GetFileTime(promptsFilePath) //refresh, because Generate() re-wrote tools names

			err = app.Prompts.WriteFiles(app.Process.Compile.folderPath)
			if err != nil {
				return err
			}

			codeErrors, err := app.Process.Compile.Compile(codeHash, app.router, app.Destroy)
			app.Prompts.SetCodeErrors(codeErrors)
			if err != nil {
				return err
			}

			saveIt = true
		}

		app.Prompts.PromptsFileTime = codeHash

	} else {

		files, err := os.ReadDir(app.Process.Compile.folderPath)
		if err != nil {
			return err
		}

		//add new tools
		codeHash := int64(0)
		//main.go
		codeHash += Tools_GetFileTime("apps/main.go")

		for _, info := range files {

			if info.IsDir() || filepath.Ext(info.Name()) != ".go" || info.Name() == "main.go" {
				continue
			}

			fileTime := Tools_GetFileTime(filepath.Join(app.Process.Compile.folderPath, info.Name()))
			codeHash += fileTime

			if !strings.HasPrefix(info.Name(), "z") {
				continue
			}

			toolName, _ := strings.CutSuffix(info.Name()[1:], ".go") //remove 'z' and '.go'
			item, found := app.Tools[toolName]
			if !found {
				//add
				schema, err := BuildToolsOpenAI_completion_tool(toolName, app.getToolFilePath(toolName), nil)
				if err != nil {
					return err
				}

				if schema != nil { //not ignored
					app.lock.Lock()
					app.Tools[toolName] = &ToolsAppItem{Schema: schema, FileTime: fileTime}
					app.lock.Unlock()

					saveIt = true
				}

			} else {
				if item.FileTime != fileTime {
					//update
					schema, err := BuildToolsOpenAI_completion_tool(toolName, app.getToolFilePath(toolName), nil)
					if err != nil {
						return err
					}

					if schema != nil { //not ignored
						app.lock.Lock()
						item.Schema = schema
						item.FileTime = fileTime
						app.lock.Unlock()

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

		if app.Process.Compile.NeedCompile(codeHash) {
			_, err := app.Process.Compile.Compile(codeHash, app.router, app.Destroy)
			if err != nil {
				return err
			}

			saveIt = true
		}

	}

	if saveIt {
		//save 'tools.json'
		app._save()
	}

	return nil
}
