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
	"path/filepath"
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

type ToolsApp struct {
	router *AppsRouter
	lock   sync.Mutex

	Process *ToolsAppProcess

	Prompts ToolsPrompts

	storage_changes int64
}

func NewToolsApp(appName string, router *AppsRouter) (*ToolsApp, error) {
	app := &ToolsApp{router: router}

	app.Process = NewToolsAppProcess(appName)

	fl, err := os.ReadFile(app.GetToolsJsonPath())
	if err == nil {
		err = LogsJsonUnmarshal(fl, app)
		if err != nil {
			return nil, err
		}
	}

	return app, nil
}

func (app *ToolsApp) Destroy() error {
	return app.StopProcess(false)
}
func (app *ToolsApp) StopProcess(waitTillExit bool) error {
	err := app.Process.Destroy(waitTillExit)
	if err != nil {
		return err
	}

	err = app.Prompts.Destroy()
	if err != nil {
		return err
	}

	return nil
}

func (app *ToolsApp) Rename(newName string) (string, error) {

	oldName := app.Process.Compile.appName

	//check
	newName = _ToolsPrompt_getValidFileName(newName)
	if newName == "" {
		return oldName, LogsErrorf("Rename failed: empty newName")
	}

	app.lock.Lock()
	defer app.lock.Unlock()

	//stop
	app.Process.Destroy(true)

	//rename
	err := os.Rename(filepath.Join("apps", oldName), filepath.Join("apps", newName))
	if err != nil {
		return oldName, err
	}

	return newName, nil
}

func (app *ToolsApp) _save() error {
	_, err := Tools_WriteJSONFile(app.GetToolsJsonPath(), app)
	if err != nil {
		return err
	}

	return nil
}

func (app *ToolsApp) GetToolsJsonPath() string {
	return filepath.Join(app.Process.Compile.GetFolderPath(), "tools.json")
}

func (app *ToolsApp) GetAllSchemas() []*ToolsOpenAI_completion_tool {
	app.lock.Lock()
	defer app.lock.Unlock()

	var schemas []*ToolsOpenAI_completion_tool

	for _, prompt := range app.Prompts.Prompts {
		if prompt.Schema != nil {
			schemas = append(schemas, prompt.Schema)
		}
	}

	return schemas
}

func (app *ToolsApp) GetPrompt(toolName string) *ToolsPrompt {
	app.lock.Lock()
	defer app.lock.Unlock()

	prompt := app.Prompts.FindPromptName(toolName)
	if prompt != nil {
		return prompt
	}

	return nil
}

func (app *ToolsApp) CheckRun() error {
	app.lock.Lock()
	defer app.lock.Unlock()

	if app.Process.Compile.Error != "" {
		return fmt.Errorf("'%s' app has compilation error: %s", app.Process.Compile.GetFolderPath(), app.Process.Compile.Error) //don't log
	}

	if app.Process.Compile.AppFileTime == 0 {
		return fmt.Errorf("'%s' app is waiting for compilation", app.Process.Compile.GetFolderPath()) //don't log
	}

	return app.Process.CheckRun(app.router)
}

func (app *ToolsApp) getPromptFilePath() string {
	return filepath.Join(app.Process.Compile.GetFolderPath(), "skyalt")
}

func (app *ToolsApp) getSecretsFilePath() string {
	return filepath.Join(app.Process.Compile.GetFolderPath(), "secrets")
}

func (app *ToolsApp) getPromptFileTime() (int64, int64, bool, error) {
	promptsFileTime := Tools_GetFileTime(app.getPromptFilePath())
	secretsFileTime := Tools_GetFileTime(app.getSecretsFilePath())

	sdkFileTime := Tools_GetFileTime("sdk/sdk.go")

	if promptsFileTime > 0 {
		fileTappFilesTimeme := sdkFileTime + promptsFileTime + secretsFileTime

		return sdkFileTime, fileTappFilesTimeme, true, nil
	} else {

		folderPath := app.Process.Compile.GetFolderPath()
		files, err := os.ReadDir(folderPath)
		if err != nil {
			return sdkFileTime, -1, false, err
		}

		//add new tools
		appFilesTime := sdkFileTime
		for _, info := range files {
			if info.IsDir() || filepath.Ext(info.Name()) != ".go" || info.Name() == "main.go" {
				continue
			}
			appFilesTime += Tools_GetFileTime(filepath.Join(folderPath, info.Name()))
		}

		return sdkFileTime, appFilesTime, false, nil
	}
}

func (app *ToolsApp) NeedRefresh() bool {
	if app.Prompts.refresh {
		app.Prompts.refresh = false
		return true
	}
	return false
}

func (app *ToolsApp) Tick(generate bool) error {
	app.lock.Lock()
	defer app.lock.Unlock()

	defer app.Prompts.ResetGenMsgs("")

	sdkFileTime, appFilesTime, hasPrompts, err := app.getPromptFileTime()
	if err != nil {
		return err
	}

	binFileMissing := !Tools_IsFileExists(app.Process.Compile.GetBinPath()) && app.Process.Compile.Error == ""

	if !generate {
		old := app.Prompts.Changed
		app.Prompts.Changed = (app.Process.Compile.AppFileTime != appFilesTime || binFileMissing)
		if !app.Prompts.Changed {
			return nil //ok
		}
		if old != app.Prompts.Changed {
			app.Prompts.refresh = true
		}
	}

	restart := false

	if !hasPrompts {
		msg := app.router.AddRecompileMsg(app.Process.Compile.appName)
		defer msg.Done()

		err := app.Prompts._reloadFromCodeFiles(app.Process.Compile.GetFolderPath())
		if err != nil {
			return err
		}

		err = app.Process.Compile.BuildMainFile(app.Prompts.Prompts) //sdk.go -> main.go
		if err != nil {
			return err
		}

		codeErrors, err := app.Process.Compile._compile(sdkFileTime, appFilesTime, false, msg)
		if LogsError(err) != nil {
			return err
		}
		app.Prompts.SetCodeErrors(codeErrors, app)
		restart = true
	} else {

		if !generate {
			if len(app.Prompts.Prompts) > 0 { //must exist

				//only recompile(for example: sdk.go changed)
				if app.Process.Compile.SdkFileTime != sdkFileTime || binFileMissing {
					msg := app.router.AddRecompileMsg(app.Process.Compile.appName)
					defer msg.Done()

					err = app.Process.Compile.BuildMainFile(app.Prompts.Prompts) //sdk.go -> main.go
					if err != nil {
						return err
					}

					codeErrors, err := app.Process.Compile._compile(sdkFileTime, appFilesTime, false, msg)
					if LogsError(err) != nil {
						return err
					}
					app.Prompts.SetCodeErrors(codeErrors, app)
					restart = true
				}
			}
		} else {
			saved, err := app.Prompts._reloadFromPromptFile(app.Process.Compile.GetFolderPath())
			if err != nil {
				return err
			}
			if saved {
				sdkFileTime, appFilesTime, _, err = app.getPromptFileTime() //refresh after save
				if err != nil {
					return err
				}
			}

			secrets, err := NewToolsSecrets(app.getSecretsFilePath())
			if err != nil {
				return err
			}

			err = app.Prompts.RemoveOldCodeFiles(app.Process.Compile.GetFolderPath())
			if err != nil {
				return err
			}

			msg := app.router.AddRecompileMsg(app.Process.Compile.appName)
			defer msg.Done()

			MAX_Errors_tries := 10

			err = app.Process.Compile.BuildMainFile(nil) //sdk.go -> main.go
			if err != nil {
				return err
			}

			app.Prompts.ResetGenMsgs(msg.msg_name)

			//Storage
			storagePrompt := app.Prompts.FindStorage()
			hasStorage := (storagePrompt != nil && storagePrompt.Prompt != "")
			if hasStorage {

				for i := range MAX_Errors_tries {
					if i == 0 {
						msg.progress_label = "Generating Storage code"
					} else {
						msg.progress_label = "Fixing Storage code"
					}

					err = app.Prompts.generatePromptCode(storagePrompt, msg, app.router.services.llms)
					if err != nil {
						return err
					}

					err = app.Prompts.WriteFiles(app.Process.Compile.GetFolderPath(), secrets) //rewrite(remove old) files
					if err != nil {
						return err
					}

					codeErrors, err := app.Process.Compile._compile(sdkFileTime, appFilesTime, true, msg)
					if err != nil {
						return err
					}
					app.Prompts.SetCodeErrors(codeErrors, app)
					if len(codeErrors) == 0 {
						break
					}
					if i+1 == MAX_Errors_tries {
						return fmt.Errorf("failed to generage Storage.go")
					}
				}
			}

			//Functions
			if app.Prompts.HasFunction() {
				for i := range MAX_Errors_tries {
					if i == 0 {
						msg.progress_label = "Generating Functions code"
					} else {
						msg.progress_label = "Fixing Functions code"
					}

					//generate code
					var wg sync.WaitGroup
					var genErr error
					for _, prompt := range app.Prompts.Prompts {
						if prompt.Type != ToolsPrompt_FUNCTION || prompt.IsCodeWithoutErrors() {
							continue
						}

						wg.Add(1)
						go func() {
							defer wg.Done()
							err = app.Prompts.generatePromptCode(prompt, msg, app.router.services.llms)
							if err != nil {
								genErr = err
							}
						}()
					}
					wg.Wait()
					if genErr != nil {
						return genErr
					}

					err = app.Prompts.WriteFiles(app.Process.Compile.GetFolderPath(), secrets) //rewrite(remove old) files
					if err != nil {
						return err
					}

					codeErrors, err := app.Process.Compile._compile(sdkFileTime, appFilesTime, true, msg)
					if err != nil {
						return err
					}
					app.Prompts.SetCodeErrors(codeErrors, app)
					if len(codeErrors) == 0 {
						break
					}
					if i+1 == MAX_Errors_tries {
						return fmt.Errorf("failed to generage Storage.go")
					}
				}
			}

			//Tools
			for i := range MAX_Errors_tries {

				if i == 0 {
					msg.progress_label = "Generating Tools code"
				} else {
					msg.progress_label = "Fixing Tools code"
				}

				//generate code
				var wg sync.WaitGroup
				var genErr error
				for _, prompt := range app.Prompts.Prompts {
					if prompt.Type != ToolsPrompt_TOOL || prompt.IsCodeWithoutErrors() {
						continue
					}

					wg.Add(1)
					go func() {
						defer wg.Done()
						err = app.Prompts.generatePromptCode(prompt, msg, app.router.services.llms)
						if err != nil {
							genErr = err
						}
					}()
				}
				wg.Wait()
				if genErr != nil {
					return genErr
				}
				if !msg.GetContinue() {
					break
				}

				err = app.Prompts.WriteFiles(app.Process.Compile.GetFolderPath(), secrets)
				if err != nil {
					return err
				}

				err = app.Process.Compile.BuildMainFile(app.Prompts.Prompts) //sdk.go -> main.go
				if err != nil {
					return err
				}

				codeErrors, err := app.Process.Compile._compile(sdkFileTime, appFilesTime, false, msg)
				if err != nil {
					return err
				}
				app.Prompts.SetCodeErrors(codeErrors, app)
				if len(codeErrors) == 0 {
					break
				}
				if i+1 == MAX_Errors_tries {
					return fmt.Errorf("failed to generage Tools")
				}
			}

			restart = true
		}
	}

	if restart {
		err = app.StopProcess(true) //stop it
		if err != nil {
			return err
		}
		err = app.Process.Compile.RemoveOldBins()
		if err != nil {
			return err
		}

		err = app.Prompts.UpdateSchemas()
		if err != nil {
			return err
		}
	}

	//save 'tools.json'
	return app._save()
}
