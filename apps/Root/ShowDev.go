package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// [ignore]
type ShowDev struct {
	AppName string
}

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

	type SdkToolsCodeError struct {
		File string
		Line int
		Col  int
		Msg  string
	}
	type SdkToolsMessages struct {
		Message   string
		Reasoning string
	}
	type SdkToolsPromptCode struct {
		Code   string
		Errors []SdkToolsCodeError
		Usage  LLMMsgUsage
	}

	type ToolsPromptTYPE int
	const (
		ToolsPrompt_STORAGE ToolsPromptTYPE = iota
		ToolsPrompt_FUNCTION
		ToolsPrompt_TOOL
	)

	type SdkToolsPrompt struct {
		Type   ToolsPromptTYPE
		Prompt string //LLM input

		//LLM output
		Messages []SdkToolsMessages

		//Code string
		CodeVersions []SdkToolsPromptCode

		//from code
		Name   string
		Schema json.RawMessage
		//Errors []SdkToolsCodeError

		//Usage LLMMsgUsage
	}
	type SdkToolsPromptGen struct {
		Name    string
		Message string
	}
	type SdkToolsPrompts struct {
		Changed bool

		Prompts  []*SdkToolsPrompt
		Err      string
		Err_line int

		StartPrompt string

		Generating_msg_id string
		Generating_items  []*SdkToolsPromptGen
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

	isGenerating := (sdk_app.Generating_msg_id != "")
	prompts_path := filepath.Join("..", app.Name, "skyalt")
	secrets_path := filepath.Join("..", app.Name, "secrets")

	filePrompts, _ := os.ReadFile(prompts_path)
	fileSecretsCipher, _ := os.ReadFile(secrets_path)

	var fileSecrets string
	if len(fileSecretsCipher) > 0 {
		plain, err := SdkDecryptAESGCM(fileSecretsCipher)
		if err != nil {
			return err
		}
		fileSecrets = string(plain)
	}

	{
		HeaderDiv := MainDiv.AddLayout(1, 0, 1, 1)
		HeaderDiv.SetColumn(1, 1, 100)
		HeaderDiv.SetColumn(2, 4, 4)

		//app settings
		SettingsDia := HeaderDiv.AddDialog("app_settings")
		st.buildSettings(SettingsDia, app, caller)
		SettingsBt := HeaderDiv.AddButton(0, 0, 1, 1, "")
		SettingsBt.Background = 0.5
		SettingsBt.IconPath = "resources/settings.png"
		SettingsBt.Icon_margin = 0.2
		SettingsBt.layout.Tooltip = "Show app Settings"
		SettingsBt.clicked = func() error {
			SettingsDia.OpenRelative(SettingsBt.layout, caller)
			return nil
		}

		HeaderDiv.AddText(1, 0, 1, 1, app.Name)

		{
			TabsDiv := HeaderDiv.AddLayout(2, 0, 1, 1)
			TabsDiv.SetColumn(0, 2, 2)
			TabsDiv.SetColumn(1, 2, 2)
			TabsDiv.Back_cd = UI_GetPalette().GetGrey(0.1)
			TabsDiv.Back_rounding = true

			PromptsBt := TabsDiv.AddButton(0, 0, 1, 1, "Prompts")
			PromptsBt.Background = 0.0
			PromptsBt.clicked = func() error {
				app.Dev.MainMode = "prompts"
				return nil
			}

			SecretsBt := TabsDiv.AddButton(1, 0, 1, 1, "Secrets")
			SecretsBt.Background = 0.0
			SecretsBt.clicked = func() error {
				app.Dev.MainMode = "secrets"
				return nil
			}

			switch app.Dev.MainMode {
			case "secrets":
				SecretsBt.Background = 1
			default: //"code"
				PromptsBt.Background = 1
			}
		}

	}

	if app.Dev.MainMode == "secrets" {
		ed := MainDiv.AddEditboxString(1, 1, 1, 1, &fileSecrets)
		ed.Linewrapping = true
		ed.Multiline = true
		ed.Align_v = 0
		ed.layout.Enable = !isGenerating
		ed.AutoSave = true //refresh "Save button"
		ed.changed = func() error {
			cipher, err := SdkEncryptAESGCM([]byte(fileSecrets))
			if err != nil {
				return err
			}
			if !bytes.Equal([]byte(fileSecrets), []byte(cipher)) {
				err = os.WriteFile(secrets_path, []byte(cipher), 0644)
				if err != nil {
					return err
				}
			}
			return nil
		}
	} else {
		prompts := string(filePrompts)

		if prompts == "" {
			prompts = "#Storage\n"
		}

		ed := MainDiv.AddEditboxString(1, 1, 1, 1, &prompts)
		ed.Linewrapping = true
		ed.Multiline = true
		ed.Align_v = 0
		ed.layout.Enable = !isGenerating
		ed.AutoSave = true //refresh "Save button"
		ed.changed = func() error {
			if !bytes.Equal(filePrompts, []byte(prompts)) {
				err = os.WriteFile(prompts_path, []byte(prompts), 0644)
				if err != nil {
					return err
				}
			}
			return nil
		}
	}

	{
		MainDiv.SetRowFromSub(2, 1, 10, true)
		FooterDiv := MainDiv.AddLayout(1, 2, 1, 1)
		FooterDiv.SetColumn(0, 1, 100)
		FooterDiv.SetColumn(1, 1, 5)
		FooterDiv.SetRowFromSub(0, 2, 5, true)

		//Note
		if app.Dev.MainMode == "secrets" {
			tx := FooterDiv.AddText(0, 0, 1, 1, "<alias> <value>\nExample: myemail@mail.com myActualEmail@gmail.com\nExample: password_1234 Ek7_sdf6m-o45-erc-er5_-df")
			tx.Align_v = 0
			tx.Cd = UI_GetPalette().GetGrey(0.5)
		} else {
			tx := FooterDiv.AddText(0, 0, 1, 1, "#storage //Describe structures for saving data.\n#function <name> //Describe background function.\n#tool <name> //Describe app's feature.\n#start //Enter prompt, which will be executed when new chat is created.")
			tx.Align_v = 0
			tx.Cd = UI_GetPalette().GetGrey(0.5)
		}

		{
			FooterRightDiv := FooterDiv.AddLayout(1, 0, 1, 1)
			FooterRightDiv.SetColumn(0, 1, 100)

			//generate
			{
				diff := sdk_app.Changed //(promptsFileTime != sdk_app.PromptsFileTime || secretsFileTime != sdk_app.SecretsFileTime)

				SaveDiv := FooterRightDiv.AddLayout(0, 0, 1, 1)

				if isGenerating {
					x := 0
					if !app.Dev.ShowSide {
						SaveDiv.SetColumn(x, 1, 100)
						CompBt := SaveDiv.AddButton(x, 0, 1, 1, "Show")
						CompBt.Background = 0.5
						CompBt.layout.Tooltip = "Show generation"
						CompBt.clicked = func() error {
							app.Dev.ShowSide = true
							return nil
						}
						x++
					}

					SaveDiv.SetColumn(x, 1, 100)
					StopBt := SaveDiv.AddButton(x, 0, 1, 1, "Stop")
					StopBt.Cd = UI_GetPalette().E
					StopBt.layout.Tooltip = "Stop generating"
					StopBt.clicked = func() error {
						callFuncMsgStop(sdk_app.Generating_msg_id)
						return nil
					}
				} else {
					SaveDiv.SetColumn(0, 1, 100)
					GenerateBt := SaveDiv.AddButton(0, 0, 1, 1, "Generate")
					GenerateBt.layout.Tooltip = "Save & Generate code"
					GenerateBt.clicked = func() error {

						app.Dev.PromptsHistory = append(app.Dev.PromptsHistory, string(filePrompts))

						callFuncGenerateApp(app.Name)
						return nil
					}
					GenerateBt.layout.Enable = diff
				}

				//history back/forward buttons ....
			}

			//Total price
			{
				sum := 0.0
				for _, prompt := range sdk_app.Prompts {
					for _, it := range prompt.CodeVersions {
						sum += it.Usage.Prompt_price + it.Usage.Input_cached_price + it.Usage.Completion_price + it.Usage.Reasoning_price
					}
				}
				tx := FooterRightDiv.AddText(0, 1, 1, 1, fmt.Sprintf("<i>Total: $%f", sum))
				tx.Align_v = 0
				tx.Align_h = 2
				tx.layout.Tooltip = "Total cost of generating code(including bugs fixes)"
			}
		}

		//Errors
		{
			n_errors := 0
			for _, prompt := range sdk_app.Prompts {
				if len(prompt.CodeVersions) > 0 {
					if len(prompt.CodeVersions[len(prompt.CodeVersions)-1].Errors) > 0 {
						n_errors++
					}
				}
			}
			if n_errors > 0 {
				tx := FooterDiv.AddText(0, 1, 2, 1, fmt.Sprintf("%d file(s) has compilation error(s)", n_errors))
				tx.Cd = UI_GetPalette().E
			}
		}

		/*FooterDiv.SetRow(1, 10, 10)
		FooterDiv.AddMediaPath(0, 1, 1, 1, "vid.mkv")

		FooterDiv.SetRow(2, 3, 3)
		FooterDiv.AddMediaPath(0, 2, 1, 1, "aud.mp3")*/
	}

	//Side panel
	if app.Dev.ShowSide {

		ui.SetColumn(0, 1, 100)
		ui.SetColumnResizable(1, 5, 25, 7)

		SideDiv := ui.AddLayout(1, 0, 1, 1)

		if isGenerating {
			SideDiv.SetColumn(0, 1, 100)

			{
				HeaderDiv := SideDiv.AddLayout(0, 0, 1, 1)
				HeaderDiv.SetColumn(1, 3, 100)
				//HeaderDiv.ScrollV.Hide = true
				HeaderDiv.ScrollH.Narrow = true

				CloseBt := HeaderDiv.AddButton(0, 0, 1, 1, ">>")
				CloseBt.layout.Tooltip = "Close side panel"
				CloseBt.Background = 0.25
				CloseBt.clicked = func() error {
					app.Dev.ShowSide = false //hide
					return nil
				}

				HeaderDiv.AddText(1, 0, 1, 1, fmt.Sprintf("Generating %d files", len(sdk_app.Generating_items)))
			}

			y := 1
			for _, it := range sdk_app.Generating_items {

				SideDiv.AddText(0, y, 1, 1, it.Name)
				y++

				SideDiv.SetRow(y, 2, 100)
				tx := SideDiv.AddText(0, y, 1, 1, it.Message)
				y++
				tx.Align_v = 0
				tx.layout.Back_cd = codeBackCd
				tx.layout.VScrollToTheBottom(true, caller)
			}

		} else {
			SideDiv.SetColumn(0, 1, 100)
			SideDiv.SetRow(1, 1, 100)

			{
				num_opened_versions := 0
				hasOpenedSchema := false
				var labels []string
				var values []string
				var icons []UIDropDownIcon
				for _, prompt := range sdk_app.Prompts {
					errStr := ""
					ncodes := len(prompt.CodeVersions)
					if ncodes > 0 {
						if len(prompt.CodeVersions[ncodes-1].Errors) > 0 { //last has error
							errStr = fmt.Sprintf(" [%d errors]", len(prompt.CodeVersions[ncodes-1].Errors))
						} else if ncodes > 1 {
							errStr = " [fixed]"
						}
					}

					labels = append(labels, prompt.Name+".go"+errStr)
					values = append(values, prompt.Name+".go")

					if prompt.Name+".go" == app.Dev.SideFile {
						hasOpenedSchema = (prompt.Type == ToolsPrompt_TOOL)
						num_opened_versions = len(prompt.CodeVersions)

						if app.Dev.SideFile_version < 0 || app.Dev.SideFile_version >= num_opened_versions {
							app.Dev.SideFile_version = num_opened_versions - 1
						}
					}

					var ic UIDropDownIcon
					switch prompt.Type {
					case ToolsPrompt_STORAGE:
						ic.Path = "resources/db.png"
						ic.Margin = 0.2
					case ToolsPrompt_FUNCTION:
						ic.Path = "resources/fx.png"
						ic.Margin = 0.2
					case ToolsPrompt_TOOL:
						ic.Path = "resources/tools.png"
						ic.Margin = 0.2
					}
					icons = append(icons, ic)
				}

				HeaderDiv := SideDiv.AddLayout(0, 0, 1, 1)

				//HeaderDiv.ScrollV.Hide = true
				HeaderDiv.ScrollH.Narrow = true

				hx := 0
				CloseBt := HeaderDiv.AddButton(hx, 0, 1, 1, ">>")
				hx++
				CloseBt.layout.Tooltip = "Close side panel"
				CloseBt.Background = 0.25
				CloseBt.clicked = func() error {
					app.Dev.ShowSide = false //hide
					return nil
				}

				HeaderDiv.SetColumn(hx, 3, 100)
				cb := HeaderDiv.AddDropDown(hx, 0, 1, 1, &app.Dev.SideFile, labels, values)
				hx++
				cb.Icons = icons
				cb.changed = func() error {
					app.Dev.SideFile_version = -1 //reset
					return nil
				}

				//code version
				if app.Dev.SideMode == "code" && num_opened_versions > 1 {
					var labels []string
					var values []string
					for i := range num_opened_versions {
						labels = append(labels, strconv.Itoa(1+i))
						values = append(labels, strconv.Itoa(1+i))
					}

					version := strconv.Itoa(1 + app.Dev.SideFile_version)
					HeaderDiv.SetColumnFromSub(hx, 1, 5, true)
					vr := HeaderDiv.AddDropDown(hx, 0, 1, 1, &version, labels, values)
					hx++
					vr.changed = func() error {
						v, _ := strconv.Atoi(version)
						app.Dev.SideFile_version = v - 1
						return nil
					}

				}

				hx++ //space

				{
					HeaderDiv.SetColumnFromSub(hx, 5, 100, true)
					TabsDiv := HeaderDiv.AddLayout(hx, 0, 1, 1)
					hx++
					TabsDiv.SetColumn(0, 2, 3)
					TabsDiv.SetColumn(1, 2, 3)
					TabsDiv.SetColumn(2, 2, 3)
					TabsDiv.Back_cd = UI_GetPalette().GetGrey(0.1)
					//TabsDiv.Border_cd = UI_GetPalette().P
					TabsDiv.Back_rounding = true

					if !hasOpenedSchema {
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
					SchemaBt.layout.Enable = hasOpenedSchema
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
			}

			if app.Dev.SideFile != "" {

				var side_prompt *SdkToolsPrompt
				var side_promptCode SdkToolsPromptCode

				app_toolName := strings.TrimSuffix(app.Dev.SideFile, filepath.Ext(app.Dev.SideFile))
				for _, prompt := range sdk_app.Prompts {
					if prompt.Name == app_toolName {
						side_prompt = prompt

						if app.Dev.SideFile_version >= 0 {

							callFuncPrint(fmt.Sprintf("-------------%d", app.Dev.SideFile_version))

							side_promptCode = prompt.CodeVersions[app.Dev.SideFile_version]
						}
						break
					}
				}

				if side_prompt != nil {
					switch app.Dev.SideMode {
					case "schema":
						tx := SideDiv.AddText(0, 1, 1, 1, string(side_prompt.Schema))
						//tx.Linewrapping = false
						tx.Align_v = 0
						tx.layout.Back_cd = codeBackCd

						//scroll
						//tx.layout.VScrollToTheBottom(true, caller)

					case "msg":
						var msgStr string
						for i := range side_prompt.Messages {
							m := side_prompt.Messages[i]
							msgStr += m.Reasoning
							if m.Reasoning == "" {
								msgStr += m.Message
							}
							if i+1 < len(side_prompt.Messages) {
								msgStr += "\n--- --- --- --- --- --- --- ---\n"
							}
						}
						tx := SideDiv.AddText(0, 1, 1, 1, msgStr)
						tx.Align_v = 0
						tx.layout.Back_cd = codeBackCd

					default:
						code := side_promptCode.Code
						if len(side_promptCode.Errors) > 0 {
							errCd := UI_GetPalette().E
							lines := strings.Split(code, "\n")
							for _, er := range side_promptCode.Errors {
								if er.Line >= 1 && er.Line <= len(lines) {
									lines[er.Line-1] = fmt.Sprintf("<rgba%d,%d,%d,255>%s</rgba>", errCd.R, errCd.G, errCd.B, lines[er.Line-1])
								}
							}

							code = strings.Join(lines, "\n")
						}

						tx := SideDiv.AddText(0, 1, 1, 1, code)
						tx.Linewrapping = false
						tx.Align_v = 0
						tx.layout.Back_cd = codeBackCd
						tx.EnableCodeFormating = true
					}

					//Price
					{
						tx := SideDiv.AddText(0, 2, 1, 1, fmt.Sprintf("<i>%s, $%f", side_promptCode.Usage.Model, side_promptCode.Usage.Prompt_price+side_promptCode.Usage.Input_cached_price+side_promptCode.Usage.Completion_price+side_promptCode.Usage.Reasoning_price))
						tx.Align_h = 2

						in := side_promptCode.Usage.Prompt_price
						inCached := side_promptCode.Usage.Input_cached_price
						out := side_promptCode.Usage.Completion_price + side_promptCode.Usage.Reasoning_price
						tx.layout.Tooltip = fmt.Sprintf("<b>%s</b>\n%s\nTime to first token: %s sec\nTime: %s sec\n%s tokens/sec\nTotal: $%s\n- Input: $%s(%d toks)\n- Cached: $%s(%d toks)\n- Output: $%s(%d+%d toks)",
							side_promptCode.Usage.Provider+":"+side_promptCode.Usage.Model,
							SdkGetDateTime(int64(side_promptCode.Usage.CreatedTimeSec)),
							strconv.FormatFloat(side_promptCode.Usage.TimeToFirstToken, 'f', 3, 64),
							strconv.FormatFloat(side_promptCode.Usage.DTime, 'f', 3, 64),
							strconv.FormatFloat(side_promptCode.Usage.GetSpeed(), 'f', 3, 64),
							strconv.FormatFloat(in+inCached+out, 'f', -1, 64),
							strconv.FormatFloat(in, 'f', -1, 64),
							side_promptCode.Usage.Prompt_tokens,
							strconv.FormatFloat(inCached, 'f', -1, 64),
							side_promptCode.Usage.Input_cached_tokens,
							strconv.FormatFloat(out, 'f', -1, 64),
							side_promptCode.Usage.Reasoning_tokens,
							side_promptCode.Usage.Completion_tokens)

					}

					//Errors
					{
						if len(side_promptCode.Errors) > 0 {
							SideDiv.AddText(0, 3, 1, 1, "Code errors:")
							SideDiv.SetRowFromSub(4, 1, 5, true)
							ErrsDiv := SideDiv.AddLayout(0, 4, 1, 1)
							ErrsDiv.ScrollH.Narrow = true
							ErrsDiv.SetColumnFromSub(0, 1, 100, true)
							for i, er := range side_promptCode.Errors {
								tx := ErrsDiv.AddText(0, i, 1, 1, fmt.Sprintf("%d:%d: %s", er.Line, er.Col, er.Msg))
								tx.Linewrapping = false
								tx.Cd = UI_GetPalette().E
							}
						}
					}
				}
			}
		}
	} else {
		ShowSideBt := ui.AddButton(3, 0, 1, 1, "<<")
		ShowSideBt.layout.Tooltip = "Show side panel"
		ShowSideBt.Background = 0.25
		ShowSideBt.clicked = func() error {
			app.Dev.ShowSide = true
			return nil
		}
	}

	return nil
}

func (st *ShowDev) buildSettings(dia *UIDialog, app *RootApp, caller *ToolCaller) {

	ui := &dia.UI

	ui.SetColumn(0, 2, 2)
	ui.SetColumn(1, 3, 7)
	y := 0

	//label
	ui.AddTextLabel(0, y, 2, 1, "Settings").Align_h = 1
	y++

	//edit app name
	ui.AddText(0, y, 1, 1, "Rename")
	name := app.Name
	RenameEd := ui.AddEditboxString(1, y, 1, 1, &name)
	RenameEd.changed = func() error {
		newName, err := callFuncRenameApp(app.Name, name)
		if err == nil {
			app.Name = newName
		}
		return err
	}
	y++

	y++ //space

	//change icon
	ui.AddText(0, y, 1, 1, "Icon")
	dstPath := filepath.Join("apps", app.Name, "icon")
	srcPath := dstPath
	IconBt := ui.AddFilePickerButton(1, y, 1, 1, &srcPath, false, false)
	IconBt.changed = func() error {
		return st._copyFile(filepath.Join("..", app.Name, "icon"), srcPath)
	}
	y++
	IconBt.Preview = true

	y++ //space

	//delete app
	DeleteBt := ui.AddButton(0, y, 2, 1, "Delete app")
	y++
	//DeleteBt.Cd = UI_GetPalette().E
	DeleteBt.ConfirmQuestion = fmt.Sprintf("Are you sure you want to permanently delete '%s' app", app.Name)
	DeleteBt.clicked = func() error {
		os.RemoveAll(filepath.Join("..", app.Name))
		dia.Close(caller)
		return nil
	}
}

func (st *ShowDev) _copyFile(dst, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

func _getFileTime(path string) int64 {
	inf, err := os.Stat(path)
	if err == nil && inf != nil {
		return inf.ModTime().UnixNano()
	}
	return 0
}
