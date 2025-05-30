package main

import "slices"

// [ignore]
type ShowDev struct {
	AppName string
}

func (st *ShowDev) run(caller *ToolCaller, ui *UI) error {
	source_root, err := NewRoot("", caller)
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
	if app.Dev.DevShowCode != "" {
		ui.SetColumn(0, 1, 100)
		ui.SetColumnResizable(1, 5, 15, 7)

		//code
		{
			CodeDiv := ui.AddLayout(1, 0, 1, 1)
			CodeDiv.SetColumn(0, 1, 100)
			CodeDiv.AddText(0, 0, 1, 1, app.Dev.DevShowCode).Align_h = 1
		}
	}

	{
		ListDiv := ui.AddLayout(0, 0, 1, 1)
		ListDiv.SetColumn(0, 1, 100)
		ListDiv.SetColumn(1, 10, 20)
		ListDiv.SetColumn(2, 1, 100)

		y := 1
		ListDiv.SetRowFromSub(y, 2, 100)
		app.Dev.Structures.Name = "Structure(s)"
		st._showPrompt(&app.Dev, &app.Dev.Structures, ListDiv.AddLayout(1, y, 1, 1))
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
			st._showPrompt(&app.Dev, tool, ListDiv.AddLayout(1, y, 1, 1))
			y++
			y++
		}
	}

	//drop app icon ....

	return nil
}

func (st *ShowDev) _showPrompt(dev *RootDev, tool *RootTool, ui *UI) {
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
		name = "<un-named>"
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

	CodeBt := BtsDiv.AddButton(1, 0, 1, 1, "Open code")
	CodeBt.layout.Enable = exist
	CodeBt.Background = 0.5
	if dev.DevShowCode != "" && dev.DevShowCode == tool.Name {
		CodeBt.Label = "Hide code"
		CodeBt.Background = 1
		//CodeBt.Border = true
	}
	CodeBt.clicked = func() error {
		if dev.DevShowCode == tool.Name {
			dev.DevShowCode = ""
		} else {
			dev.DevShowCode = tool.Name
		}
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
	//GenBt.layout.Enable = ....
	GenBt.clicked = func() error {
		//...
		//save hash of file ....
		return nil
	}

	//move(reorder) ....

	//remove 'z' from files ....
	//all struct and utils must be in Structures.go ....

	//if compiler find bug in file, show it here ....
}
