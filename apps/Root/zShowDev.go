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

	//check
	if len(app.Dev.Tools) == 0 {
		app.Dev.Tools = append(app.Dev.Tools, &RootTool{})
	}

	//add, delete Tools from list of files. Keep tools without name ....

	ui.SetColumn(0, 1, 100)
	ui.SetRow(0, 1, 100)

	if app.Dev.ShowTool != "" {
		ui.SetColumn(0, 1, 100)
		ui.SetColumnResizable(1, 5, 25, 7)

		CodeDiv := ui.AddLayout(1, 0, 1, 1)
		CodeDiv.SetColumn(0, 1, 100)
		CodeDiv.SetRowFromSub(1, 1, 100)

		{
			HeaderDiv := CodeDiv.AddLayout(0, 0, 1, 1)
			HeaderDiv.SetColumn(0, 3, 100)
			HeaderDiv.SetColumn(1, 2, 3)
			HeaderDiv.SetColumn(2, 2, 3)

			HeaderDiv.AddText(0, 0, 1, 1, "<b>"+app.Name+"/"+app.Dev.ShowTool+".go").Align_h = 0

			CodeBt := HeaderDiv.AddButton(1, 0, 1, 1, "Code")
			CodeBt.Background = 0.5
			CodeBt.clicked = func() error {
				app.Dev.ShowToolMode = "code"
				return nil
			}
			MessageBt := HeaderDiv.AddButton(2, 0, 1, 1, "Message")
			MessageBt.Background = 0.5
			MessageBt.clicked = func() error {
				app.Dev.ShowToolMode = "message"
				return nil
			}

			if app.Dev.ShowToolMode == "message" {
				MessageBt.Background = 1
			} else {
				CodeBt.Background = 1
			}
		}

		var tool *RootTool
		if app.Dev.ShowTool == "Structures" {
			tool = &app.Dev.Structures
		} else {
			for _, t := range app.Dev.Tools {
				if t.Name == app.Dev.ShowTool {
					tool = t
					break
				}
			}
		}

		if tool != nil {
			if app.Dev.ShowToolMode == "message" {

				tx := CodeDiv.AddText(0, 1, 1, 1, tool.Message)
				tx.Linewrapping = false
				tx.Align_v = 0
				tx.layout.Back_cd = UI_GetPalette().GetGrey(0.05)

				//scroll doesn't work? .............

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

	y := 1

	//app icon, change name, delete app
	{
		//drop app icon ...........
	}

	//Create/Edit tools
	{
		ListDiv := ui.AddLayout(0, 0, 1, 1)
		ListDiv.SetColumn(0, 1, 100)
		ListDiv.SetColumn(1, 10, 20)
		ListDiv.SetColumn(2, 1, 100)

		ListDiv.SetRowFromSub(y, 2, 100)
		app.Dev.Structures.Name = "Structures"
		st._showPrompt(app, &app.Dev.Structures, -1, ListDiv.AddLayout(1, y, 1, 1), caller)
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

		for i, tool := range app.Dev.Tools {
			ListDiv.SetRowFromSub(y, 2, 100)
			st._showPrompt(app, tool, i, ListDiv.AddLayout(1, y, 1, 1), caller)
			y++
			y++
		}
	}

	return nil
}

func (st *ShowDev) _showPrompt(app *RootApp, tool *RootTool, tool_i int, ui *UI, caller *ToolCaller) {
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
	BtsDiv.SetColumn(1, 1, 100)
	btm := 3.0
	BtsDiv.SetColumn(2, btm, btm)
	BtsDiv.SetColumn(3, btm, btm)
	BtsDiv.SetColumn(4, btm, btm)

	exist := (tool.Name != "") //&& file_exist ....

	if tool.Name != "Structures" {
		MoveBt := BtsDiv.AddButton(0, 0, 1, 1, "<small>↕")
		MoveBt.Background = 0.5
		MoveBt.Cd = UI_GetPalette().GetGrey(0.15)
		MoveBt.Drag_group = "tool"
		MoveBt.Drop_group = "tool"
		MoveBt.Drag_index = tool_i
		MoveBt.Drop_v = true
		MoveBt.dropMove = func(src_i, dst_i int, src_source, dst_source string) error {
			Layout_MoveElement(&app.Dev.Tools, &app.Dev.Tools, src_i, dst_i)
			return nil
		}
	}

	CodeBt := BtsDiv.AddButton(2, 0, 1, 1, "Show code")
	CodeBt.layout.Enable = exist
	CodeBt.Background = 0.5
	if app.Dev.ShowTool != "" && app.Dev.ShowTool == tool.Name {
		CodeBt.Label = "Hide code"
		CodeBt.Background = 1
		//CodeBt.Border = true
	}
	CodeBt.clicked = func() error {
		if app.Dev.ShowTool == tool.Name {
			app.Dev.ShowTool = ""
		} else {
			app.Dev.ShowTool = tool.Name
			app.Dev.ShowToolMode = "code"
		}
		return nil
	}

	IgnoreBt := BtsDiv.AddButton(3, 0, 1, 1, "Ignore")
	IgnoreBt.layout.Enable = exist
	IgnoreBt.Background = 0.5
	IgnoreBt.clicked = func() error {
		//rename file to .goo ....
		return nil
	}

	GenBt := BtsDiv.AddButton(4, 0, 1, 1, "Generate")
	//GenBt.layout.Enable = .... if prompt or file changed
	GenBt.clicked = func() error {

		app.Dev.ShowTool = tool.Name //open code panel
		app.Dev.ShowToolMode = "message"

		code := ""
		var err error
		if tool.Name == "Structures" {
			code, err = st.GenerateStructure(tool, caller)

		} else {
			code, err = st.GenerateFunction(tool, caller)

			tool.Name = "" //get name from code ......
		}

		app.Dev.ShowToolMode = "code" //switch to final code

		if err != nil {
			return err
		}
		err = os.WriteFile(tool.GetFilePath(app), []byte(code), 0644)
		if err != nil {
			return err
		}
		//save hash? ....

		return nil
	}

	//if compiler find bug in file, show it here ....
}

func (st *ShowDev) GenerateStructure(tool *RootTool, caller *ToolCaller) (string, error) {
	comp := st._prepareLLM(tool.Prompt)

	exampleFile := `
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
	comp.SystemMessage += "Here is the example file with code:\n```go" + exampleFile + "```\n"

	comp.SystemMessage += "Based on user message, rewrite above code. Your job is to design structures. Write functions only if user ask for them. You may write multiple structures, but output everything in one code block.\n"

	comp.SystemMessage += "Structures can't have pointers, because they will be saved as JSON, so instead of pointer(s) use ID which is saved in map[interger or string ID].\n"

	//maybe add old file structures, because it's needed that struct and attributes names are same ...............

	return st._runLLM(tool, &comp, caller)
}

func (st *ShowDev) GenerateFunction(tool *RootTool, caller *ToolCaller) (string, error) {
	comp := st._prepareLLM(tool.Prompt)

	//exampleApis := ``

	//exampleStructs := ``

	exampleFile := `
package main

type ExampleTool struct {
	//<tool's arguments>
}

func (st *ExampleTool) run(caller *ToolCaller, ui *UI) error {

	//<...>

	return nil
}
`

	//Arguments 'Out_' ....

	//add list of structures and their loading functions ....
	/*source_root, err := NewRoot("")
	if err != nil {
		return err
	}*/

	comp.SystemMessage = "You are a programmer. You write code in Go language.\n"
	comp.SystemMessage += "Here is the example file with code:\n```go" + exampleFile + "```\n"

	//Pick up Tool name. Here are the names of tools which already exists, don't use them.

	//..........

	return st._runLLM(tool, &comp, caller)
}

func (st *ShowDev) _prepareLLM(user_prompt string) LLMxAICompleteChat {
	var comp LLMxAICompleteChat
	comp.Model = "grok-3-mini" //grok-3-mini-fast
	comp.Temperature = 0.2
	comp.Max_tokens = 65536
	comp.Top_p = 0.7 //1.0
	comp.Frequency_penalty = 0
	comp.Presence_penalty = 0
	comp.Reasoning_effort = ""
	comp.Max_iteration = 1

	comp.UserMessage = user_prompt

	return comp
}

func (st *ShowDev) _runLLM(tool *RootTool, comp *LLMxAICompleteChat, caller *ToolCaller) (string, error) {
	tool.Message = ""
	comp.delta = func(msg *ChatMsg) {
		if msg.Content.Calls != nil {
			tool.Message = msg.Content.Calls.Content
		}
	}

	code := ""
	_, err := CallTool(comp.run, caller)
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

	return code, nil
}
