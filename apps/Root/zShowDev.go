package main

import (
	"os"
	"regexp"
	"slices"
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

	//check
	if len(app.Dev.Tools) == 0 {
		app.Dev.Tools = append(app.Dev.Tools, &RootTool{})
	}

	ui.SetColumn(0, 1, 100)
	ui.SetRow(0, 1, 100)
	if app.Dev.ShowCode != "" {
		ui.SetColumn(0, 1, 100)
		ui.SetColumnResizable(1, 5, 25, 7)

		//code
		{
			CodeDiv := ui.AddLayout(1, 0, 1, 1)
			CodeDiv.SetColumn(0, 1, 100)
			CodeDiv.SetRowFromSub(1, 1, 100)
			CodeDiv.AddText(0, 0, 1, 1, app.Name+"/"+app.Dev.ShowCode+".go").Align_h = 1

			var tool *RootTool
			if app.Dev.ShowCode == "Structures" {
				tool = &app.Dev.Structures
			} else {
				for _, t := range app.Dev.Tools {
					if t.Name == app.Dev.ShowCode {
						tool = t
						break
					}
				}
			}

			if tool != nil {
				if tool.Message != "" {

					tx := CodeDiv.AddText(0, 1, 1, 1, tool.Message)
					tx.Linewrapping = false
					tx.Align_v = 0
					tx.layout.Back_cd = UI_GetPalette().GetGrey(0.05)

					//scroll nefunguje .............

					//scroll
					tx.layout.VScrollToTheBottom(true, caller)

				} else {
					fl, err := os.ReadFile(tool.GetFilePath(app))
					if err == nil {
						tx := CodeDiv.AddText(0, 1, 1, 1, string(fl))
						tx.Linewrapping = false
						tx.Align_v = 0
						tx.layout.Back_cd = UI_GetPalette().GetGrey(0.05)
					}
				}
			}
		}
	}

	{
		ListDiv := ui.AddLayout(0, 0, 1, 1)
		ListDiv.SetColumn(0, 1, 100)
		ListDiv.SetColumn(1, 10, 20)
		ListDiv.SetColumn(2, 1, 100)

		y := 1
		ListDiv.SetRowFromSub(y, 2, 100)
		app.Dev.Structures.Name = "Structures"
		st._showPrompt(app, &app.Dev.Structures, ListDiv.AddLayout(1, y, 1, 1), caller)
		y++
		y++

		//ui.AddDivider(1, y, 1, 1, true)
		//y++

		NewToolBt := ListDiv.AddButton(1, y, 1, 1, "Create new tool")
		NewToolBt.Background = 0.5
		y++
		y++

		NewToolBt.clicked = func() error {
			app.Dev.Tools = slices.Insert(app.Dev.Tools, 0, &RootTool{}) //1st
			return nil
		}

		for _, tool := range app.Dev.Tools {
			ListDiv.SetRowFromSub(y, 2, 100)
			st._showPrompt(app, tool, ListDiv.AddLayout(1, y, 1, 1), caller)
			y++
			y++
		}
	}

	//drop app icon ....

	return nil
}

func (st *ShowDev) _showPrompt(app *RootApp, tool *RootTool, ui *UI, caller *ToolCaller) {
	m := 0.25 //maring
	ui.SetColumn(0, m, m)
	ui.SetColumn(1, 1, 100)
	ui.SetColumn(2, m, m)
	ui.SetRow(0, m, m)
	ui.SetRowFromSub(2, 1, 10) //prompt
	ui.SetRow(4, m, m)

	ui.Back_cd = UI_GetPalette().GetGrey(0.08)
	ui.Back_rounding = true

	name := tool.Name
	if name == "" {
		name = "<Tool's name>"
	}
	ui.AddText(1, 1, 1, 1, name)
	ed := ui.AddEditboxString(1, 2, 1, 1, &tool.Prompt)
	ed.Multiline = true

	BtsDiv := ui.AddLayout(1, 3, 1, 1)
	BtsDiv.SetColumn(0, 1, 100)

	btm := 3.0
	BtsDiv.SetColumn(1, btm, btm)
	BtsDiv.SetColumn(2, btm, btm)
	BtsDiv.SetColumn(3, btm, btm)

	exist := (tool.Name != "") //&& file_exist ....

	CodeBt := BtsDiv.AddButton(1, 0, 1, 1, "Show code")
	CodeBt.layout.Enable = exist
	CodeBt.Background = 0.5
	if app.Dev.ShowCode != "" && app.Dev.ShowCode == tool.Name {
		CodeBt.Label = "Hide code"
		CodeBt.Background = 1
		//CodeBt.Border = true
	}
	CodeBt.clicked = func() error {
		if app.Dev.ShowCode == tool.Name {
			app.Dev.ShowCode = ""
		} else {
			app.Dev.ShowCode = tool.Name
		}
		tool.Message = "" //reset
		return nil
	}

	IgnoreBt := BtsDiv.AddButton(2, 0, 1, 1, "Ignore")
	IgnoreBt.layout.Enable = exist
	IgnoreBt.Background = 0.5
	IgnoreBt.clicked = func() error {
		//rename file to .goo ....
		return nil
	}

	GenBt := BtsDiv.AddButton(3, 0, 1, 1, "Generate")
	//GenBt.layout.Enable = .... if prompt or file changed
	GenBt.clicked = func() error {

		app.Dev.ShowCode = tool.Name //open code panel

		if tool.Name == "Structures" {
			code, err := st.GenerateStructure(tool, caller)
			if err != nil {
				return err
			}

			err = os.WriteFile(tool.GetFilePath(app), []byte(code), 0644)
			if err != nil {
				return err
			}
			//save hash? ....
			return nil
		} else {
			//GenerateTool()....
		}
		return nil
	}

	//move(reorder) ....

	//remove 'z' from files ....
	//all struct and utils must be in Structures.go ....

	//if compiler find bug in file, show it here ....
}

func (st *ShowDev) GenerateStructure(tool *RootTool, caller *ToolCaller) (string, error) {

	var comp LLMxAICompleteChat
	comp.Model = "grok-3-mini" //grok-3-mini-fast
	comp.Temperature = 0.2
	comp.Max_tokens = 65536
	comp.Top_p = 0.7 //1.0
	comp.Frequency_penalty = 0
	comp.Presence_penalty = 0
	comp.Reasoning_effort = ""

	comp.UserMessage = tool.Prompt
	comp.Max_iteration = 1
	{
		exampleStructFile := `
package main

type ExampleStruct struct {
	//<attributes>
}

func NewExampleStruct(filePath string) (*ExampleStruct, error) {
	st := &ExampleStruct{}

	//<set 'st' default values here>

	return _loadInstance(filePath, "ExampleStruct", "json", st, true)
}

//<structures functions here>
`
		comp.SystemMessage = "You are a programmer. You write code in Go language.\n"

		comp.SystemMessage += "Here is the example file with code:\n```go" + exampleStructFile + "```\n"

		comp.SystemMessage += "Based on user message, rewrite above code. Your job is to design structures. Write functions only if user ask for them. You may write multiple structures, but output everything in one code block.\n"

		comp.SystemMessage += "Structures can't have pointers, because they will be saved as JSON, so instead of pointer(s) use ID which is saved in map[interger or string ID].\n"

		//maybe add old file structures, because it's needed that struct and attributes names are same ...............
	}

	tool.Message = ""
	comp.delta = func(msg *ChatMsg) {
		if msg.Content.Calls != nil {
			tool.Message = msg.Content.Calls.Content
		}
	}

	code := ""
	_, err := CallTool(comp.run, caller)
	tool.Message = "" //reset
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile("(?s)```(?:go|golang)\n(.*?)\n```")
	matches := re.FindAllStringSubmatch(comp.Out_last_message, -1)

	var goCode strings.Builder
	for _, match := range matches {
		if len(match) > 1 {
			goCode.WriteString(match[1])
			goCode.WriteString("\n")
		}
	}

	code = strings.TrimSpace(goCode.String())

	return code, err
}
