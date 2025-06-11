package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// [ignore]
type ShowDev struct {
	AppName string
}

//remove 'z' from files ....
//all struct and utils must be in Structures.go ....

func (st *ShowDev) run(caller *ToolCaller, ui *UI) error {
	source_root, err := NewRoot("")
	if err != nil {
		return err
	}

	//refresh apps
	app, err := source_root.refreshApps()
	if err != nil {
		return err
	}

	codeBackCd := UI_GetPalette().GetGrey(0.05)

	ui.SetColumn(0, 1, 100)
	ui.SetRow(0, 1, 100)

	MainDiv := ui.AddLayout(0, 0, 1, 1)
	MainDiv.SetColumn(0, 1, 100)
	MainDiv.SetColumn(1, 10, 20)
	MainDiv.SetColumn(2, 1, 100)
	MainDiv.SetRow(1, 1, 100)
	MainDiv.SetRow(2, 2, 2)

	type SdkLLMMsgUsage struct {
		Prompt_tokens       int
		Input_cached_tokens int
		Completion_tokens   int
		Reasoning_tokens    int

		Prompt_price       float64
		Input_cached_price float64
		Completion_price   float64
		Reasoning_price    float64
	}
	type SdkToolsCodeError struct {
		File string
		Line int
		Col  int
		Msg  string
	}
	type SdkToolsPrompt struct {
		Prompt string //LLM input

		//LLM output
		Message   string
		Reasoning string

		Code string

		//from code
		Name   string
		Schema json.RawMessage
		Errors []SdkToolsCodeError

		Usage SdkLLMMsgUsage
	}
	type SdkToolsPrompts struct {
		PromptsFileTime int64

		Prompts  []*SdkToolsPrompt
		Err      string
		Err_line int

		Generating_name string
		Generating_msg  string
	}
	var sdk_app SdkToolsPrompts
	appJs, err := callFuncGetToolData(app.Name)
	if err != nil {
		return err
	}
	err = json.Unmarshal(appJs, &sdk_app)
	if err != nil {
		return err
	}

	isGenerating := (sdk_app.Generating_name != "")
	{
		prompts_path := filepath.Join("..", app.Name, "skyalt")
		filePromptsStr, _ := os.ReadFile(prompts_path)

		if app.Dev.PromptsFileTime != sdk_app.PromptsFileTime {
			app.Dev.Prompts = string(filePromptsStr) //rewrite from file
			app.Dev.PromptsFileTime = sdk_app.PromptsFileTime
		}

		diff := (string(filePromptsStr) != app.Dev.Prompts)

		HeaderDiv := MainDiv.AddLayout(1, 0, 1, 1)
		HeaderDiv.SetColumn(0, 1, 100)
		HeaderDiv.SetColumn(1, 3, 3)
		HeaderDiv.SetColumn(2, 3, 3)

		HeaderDiv.AddText(0, 0, 1, 1, app.Name)

		CancelBt := HeaderDiv.AddButton(1, 0, 1, 1, "Cancel")
		CancelBt.Background = 0.5
		CancelBt.ConfirmQuestion = "Are you sure?"
		CancelBt.Tooltip = "Trash changes"
		CancelBt.clicked = func() error {
			app.Dev.Prompts = string(filePromptsStr)
			return nil
		}
		SaveBt := HeaderDiv.AddButton(2, 0, 1, 1, "Save")
		SaveBt.Tooltip = "Save & Generate code"
		SaveBt.clicked = func() error {
			os.WriteFile(prompts_path, []byte(app.Dev.Prompts), 0644)
			return nil
		}

		if isGenerating {
			xx := 3
			if !app.Dev.ShowSide {
				HeaderDiv.SetColumn(xx, 3, 3)
				CompBt := HeaderDiv.AddButton(xx, 0, 1, 1, "Generating")
				CompBt.Tooltip = "Show generation"
				CompBt.clicked = func() error {
					app.Dev.ShowSide = true
					return nil
				}
				xx++
			}

			HeaderDiv.SetColumn(xx, 3, 3)
			StopBt := HeaderDiv.AddButton(xx, 0, 1, 1, "Stop")
			StopBt.Cd = UI_GetPalette().E
			StopBt.Tooltip = "Stop generating"
			StopBt.clicked = func() error {
				//....
				return nil
			}
			xx++
		}

		SaveBt.layout.Enable = diff && !isGenerating
		CancelBt.layout.Enable = diff && !isGenerating

		//back/forward buttons ....
	}

	{
		ed := MainDiv.AddEditboxString(1, 1, 1, 1, &app.Dev.Prompts)
		ed.Align_v = 0
		ed.Linewrapping = true
		ed.Multiline = true
		ed.layout.Enable = !isGenerating
	}

	//Note
	{
		MainDiv.SetRowFromSub(2, 1, 10)
		FooterDiv := MainDiv.AddLayout(1, 2, 1, 1)
		FooterDiv.SetColumn(0, 1, 100)
		FooterDiv.SetColumn(1, 1, 3)
		FooterDiv.SetRowFromSub(0, 1, 5)
		tx := FooterDiv.AddText(0, 0, 1, 1, "#storage //Describe structures for saving data.\n#<NameOfTool> //Describe app's feature.")
		tx.Align_v = 0
		tx.Cd = UI_GetPalette().GetGrey(0.5)

		//Total price
		{
			sum := 0.0
			for _, it := range sdk_app.Prompts {
				sum += it.Usage.Prompt_price + it.Usage.Input_cached_price + it.Usage.Completion_price + it.Usage.Reasoning_price
			}
			tx := FooterDiv.AddText(1, 0, 1, 1, fmt.Sprintf("<i>Total: $%.3f", sum))
			tx.Align_v = 0
			tx.Align_h = 2
		}

		//Total errors
		{
			n_errors := 0
			for _, it := range sdk_app.Prompts {
				if len(it.Errors) > 0 {
					n_errors++
				}
			}
			if n_errors > 0 {
				tx := FooterDiv.AddText(0, 1, 1, 1, fmt.Sprintf("%d file(s) has compilation error(s)", n_errors))
				tx.Cd = UI_GetPalette().E
			}
		}
	}

	//app icon, change name, delete app
	/*{
		HeaderDiv := MainDiv.AddLayout(1, y, 1, 1)
		HeaderDiv.SetColumn(0, 2, 2)
		HeaderDiv.SetColumn(1, 1, 1)

		//app name
		//....

		//delete app
		//....

		//change icon
		path := filepath.Join("apps", app.Name, "icon.png")
		IconBt := HeaderDiv.AddFilePickerButton(1, 0, 1, 1, &path, false, false)
		IconBt.Preview = true
		y++
	}*/

	//Side panel
	if app.Dev.ShowSide {

		ui.SetColumn(0, 1, 100)
		ui.SetColumnResizable(1, 5, 25, 7)

		SideDiv := ui.AddLayout(1, 0, 1, 1)

		if isGenerating {
			SideDiv.SetColumn(0, 1, 100)
			SideDiv.SetRow(1, 1, 100)

			{
				HeaderDiv := SideDiv.AddLayout(0, 0, 1, 1)
				HeaderDiv.SetColumn(0, 3, 100)
				HeaderDiv.ScrollV.Hide = true
				HeaderDiv.ScrollH.Narrow = true

				HeaderDiv.AddText(0, 0, 1, 1, "Generating code for <i>"+sdk_app.Generating_name)

				CloseBt := HeaderDiv.AddButton(1, 0, 1, 1, "X")
				CloseBt.Background = 0.25
				CloseBt.clicked = func() error {
					app.Dev.ShowSide = false //hide
					return nil
				}
			}

			//....
			tx := SideDiv.AddText(0, 1, 1, 1, string(sdk_app.Generating_msg))
			//tx.Linewrapping = false
			tx.Align_v = 0
			tx.layout.Back_cd = codeBackCd

			tx.layout.VScrollToTheBottom(true, caller)

		} else {

			SideDiv.SetColumn(0, 1, 100)
			SideDiv.SetRow(1, 1, 100)

			{
				var labels []string
				var values []string
				for _, it := range sdk_app.Prompts {
					errStr := ""
					if len(it.Errors) > 0 {
						errStr = fmt.Sprintf(" (%d errors)", len(it.Errors))
					}
					labels = append(labels, it.Name+".go"+errStr)
					values = append(values, it.Name+".go")

				}

				HeaderDiv := SideDiv.AddLayout(0, 0, 1, 1)
				HeaderDiv.SetColumn(0, 3, 100)
				HeaderDiv.SetColumn(1, 1, 1)
				HeaderDiv.SetColumnFromSub(2, 5, 100)
				HeaderDiv.ScrollV.Hide = true
				HeaderDiv.ScrollH.Narrow = true

				HeaderDiv.AddCombo(0, 0, 1, 1, &app.Dev.SideFile, labels, values)

				{
					TabsDiv := HeaderDiv.AddLayout(2, 0, 1, 1)
					TabsDiv.SetColumn(0, 2, 3)
					TabsDiv.SetColumn(1, 2, 3)
					TabsDiv.SetColumn(2, 2, 3)
					TabsDiv.Back_cd = UI_GetPalette().GetGrey(0.1)
					//TabsDiv.Border_cd = UI_GetPalette().P
					TabsDiv.Back_rounding = true

					isStruct := app.Dev.SideFile == "Storage.go"
					if isStruct {
						if app.Dev.SideMode == "schema" {
							app.Dev.SideMode = "code"
						}
					}

					CodeBt := TabsDiv.AddButton(0, 0, 1, 1, "Code")
					CodeBt.Background = 0.0
					CodeBt.clicked = func() error {
						app.Dev.SideMode = "code"
						return nil
					}
					SchemaBt := TabsDiv.AddButton(1, 0, 1, 1, "Schema")
					SchemaBt.Background = 0.0
					SchemaBt.layout.Enable = !isStruct
					SchemaBt.clicked = func() error {
						app.Dev.SideMode = "schema"
						return nil
					}

					MsgBt := TabsDiv.AddButton(2, 0, 1, 1, "Thinking")
					MsgBt.Background = 0.0
					MsgBt.clicked = func() error {
						app.Dev.SideMode = "msg"
						return nil
					}

					switch app.Dev.SideMode {
					case "schema":
						SchemaBt.Background = 1
					case "msg":
						MsgBt.Background = 1
					default: //"code"
						CodeBt.Background = 1
					}
				}

				CloseBt := HeaderDiv.AddButton(3, 0, 1, 1, "X")
				CloseBt.Background = 0.25
				CloseBt.clicked = func() error {
					app.Dev.ShowSide = false //hide
					return nil
				}
			}

			if app.Dev.SideFile != "" {

				var prompt *SdkToolsPrompt
				app_toolName := strings.TrimSuffix(app.Dev.SideFile, filepath.Ext(app.Dev.SideFile))
				for _, it := range sdk_app.Prompts {
					if it.Name == app_toolName {
						prompt = it
						break
					}
				}

				if prompt != nil {
					switch app.Dev.SideMode {
					case "schema":
						tx := SideDiv.AddText(0, 1, 1, 1, string(prompt.Schema))
						//tx.Linewrapping = false
						tx.Align_v = 0
						tx.layout.Back_cd = codeBackCd

						//scroll
						//tx.layout.VScrollToTheBottom(true, caller)

					case "msg":
						tx := SideDiv.AddText(0, 1, 1, 1, prompt.Reasoning)
						tx.Align_v = 0
						tx.layout.Back_cd = codeBackCd

					default:
						fl, err := os.ReadFile(filepath.Join("..", app.Name, app.Dev.SideFile))
						if err == nil {
							tx := SideDiv.AddText(0, 1, 1, 1, string(fl))
							tx.Linewrapping = false
							tx.Align_v = 0
							tx.layout.Back_cd = codeBackCd
							tx.EnableCodeFormating = true
						}
					}

					//Price
					{
						tx := SideDiv.AddText(0, 2, 1, 1, fmt.Sprintf("<i>$%.3f", prompt.Usage.Prompt_price+prompt.Usage.Input_cached_price+prompt.Usage.Completion_price+prompt.Usage.Reasoning_price))
						tx.Align_h = 2
					}

					//Errors
					{
						if len(prompt.Errors) > 0 {
							SideDiv.AddText(0, 3, 1, 1, "Code errors").Align_h = 1
							SideDiv.SetRowFromSub(4, 1, 5)
							ErrsDiv := SideDiv.AddLayout(0, 4, 1, 1)
							ErrsDiv.SetColumn(0, 1, 100)
							for i, er := range prompt.Errors {
								tx := ErrsDiv.AddText(0, i, 1, 1, fmt.Sprintf("%d:%d: %s", er.Line, er.Col, er.Msg))
								tx.Cd = UI_GetPalette().E
							}
						}
					}
				}
			}
		}
	} else {
		ShowSideBt := ui.AddButton(3, 0, 1, 1, "<<")
		ShowSideBt.Tooltip = "Show side panel"
		ShowSideBt.Background = 0.25
		ShowSideBt.clicked = func() error {
			app.Dev.ShowSide = true
			return nil
		}
	}

	return nil
}
