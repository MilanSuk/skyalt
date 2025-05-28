package main

import (
	"fmt"
	"os"
	"slices"
	"time"
)

// [ignore]
type ShowRoot struct {
	AddBrush *LayoutPick
}

func (st *ShowRoot) run(caller *ToolCaller, ui *UI) error {
	source_root, err := NewRoot("", caller)
	if err != nil {
		return err
	}

	//refresh apps
	app, err := st.refreshApps(source_root)
	if err != nil {
		return err
	}

	//load chat
	var source_chat *Chat
	var chat_fileName string
	if app != nil {
		source_chat, chat_fileName, err = st.refreshChats(app, caller)
		if err != nil {
			return err
		}
	}

	//save brush
	if source_chat != nil {
		if st.AddBrush != nil {
			source_chat.Input.MergePick(*st.AddBrush)
			ui.ActivateEditbox("chat_user_prompt", caller)
		}
	}

	ui.SetColumnResizable(0, 5, 15, 6)
	ui.SetColumn(1, 1, 100)
	ui.SetRow(0, 1, 100)

	//Side
	{
		SideDiv := ui.AddLayout(0, 0, 1, 1)
		SideDiv.SetColumn(0, 2, 2)
		SideDiv.SetColumn(1, 1, 100)

		HeaderDiv := SideDiv.AddLayout(0, 0, 2, 1)

		SideDiv.SetRow(1, 1, 100)
		AppsDiv := SideDiv.AddLayout(0, 1, 1, 1)
		TabsDiv := SideDiv.AddLayout(1, 1, 1, 1)

		//Header
		{
			x := 0

			//Logo
			{
				AboutDia := HeaderDiv.AddDialog("about")
				st.buildAbout(&AboutDia.UI)

				HeaderDiv.SetColumn(x, 2, 2)
				//logo := HeaderDiv.AddImagePath(x, 0, 1, 1, "resources/logo_small.png")
				logoBt := HeaderDiv.AddButton(x, 0, 1, 1, "")
				logoBt.Icon_align = 1
				logoBt.Background = 0.2
				logoBt.IconPath = "resources/logo_small.png"
				logoBt.Icon_margin = 0.2
				logoBt.Tooltip = "About"
				logoBt.clicked = func() error {
					AboutDia.OpenRelative(logoBt.layout, caller)
					return nil
				}
				x++
			}

			//New Tab button
			{
				HeaderDiv.SetColumn(x, 0, 100)

				bt := HeaderDiv.AddButton(x, 0, 1, 1, "New Tab")
				x++
				bt.Background = 0.5
				bt.Shortcut = 't'
				bt.layout.Enable = (app != nil)
				bt.clicked = func() error {
					if app == nil {
						return fmt.Errorf("No app selected")
					}

					fileName := fmt.Sprintf("Chat-%d.json", time.Now().UnixMicro())
					source_chat, err = NewChat(fmt.Sprintf("../%s/Chats/%s", app.Name, fileName), caller)
					if err != nil {
						return nil
					}

					app.Chats = slices.Insert(app.Chats, 0, RootChat{Label: "Empty chat", FileName: fileName})
					app.Selected_chat_i = 0
					ui.ActivateEditbox("chat_user_prompt", caller)

					TabsDiv.VScrollToTheTop(caller)

					return nil
				}
			}

			//Settings
			{
				logoBt := HeaderDiv.AddButton(x, 0, 1, 1, "")
				x++
				logoBt.IconPath = "resources/settings.png"
				logoBt.Icon_margin = 0.2
				logoBt.Tooltip = "Show Settings"
				logoBt.Background = 0.25
				if source_root.Mode == "settings" {
					logoBt.Background = 1
				}
				logoBt.clicked = func() error {
					if source_root.Mode == "settings" {
						source_root.Mode = ""
					} else {
						source_root.Mode = "settings"
					}
					return nil
				}
			}
		}

		//Apps
		{
			AppsDiv.ScrollV.Narrow = true
			AppsDiv.SetColumn(0, 1, 100)
			AppsDiv.Back_cd = UI_GetPalette().GetGrey(0.1)
			for i, app := range source_root.Apps {

				AppsDiv.SetRow(i, 2, 2)

				bt := AppsDiv.AddButton(0, i, 1, 1, "")
				bt.Icon_align = 1
				bt.Background = 0.2
				if i == source_root.Selected_app_i {
					bt.Background = 1
				}
				bt.Tooltip = app.Name
				bt.IconPath = fmt.Sprintf("apps/%s/icon.png", app.Name)
				bt.Icon_margin = 0.6

				bt.clicked = func() error {
					source_root.Selected_app_i = i
					return nil
				}

				bt.Drag_group = "app"
				bt.Drop_group = "app"
				bt.Drag_index = i
				bt.Drop_v = true
				bt.dropMove = func(src_i, dst_i int, src_source, dst_source string) error {
					Layout_MoveElement(&source_root.Apps, &source_root.Apps, src_i, dst_i)
					source_root.Selected_app_i = dst_i
					ui.ActivateEditbox("chat_user_prompt", caller)
					return nil
				}
			}
		}

		//List of tabs
		if app != nil {
			TabsDiv.SetColumn(0, 1, 100)
			yy := 0
			for i, tab := range app.Chats {

				btChat := TabsDiv.AddButton(0, yy, 1, 1, tab.Label)
				yy++

				btChat.Align = 0
				btChat.Background = 0.2
				if i == app.Selected_chat_i {
					if source_root.Mode != "" {
						btChat.Border = true
					} else {
						btChat.Background = 1 //selected
					}
				}
				btChat.clicked = func() error {
					app.Selected_chat_i = i
					source_root.Mode = "" //reset
					ui.ActivateEditbox("chat_user_prompt", caller)
					return nil
				}

				btChat.Drag_group = "chat"
				btChat.Drop_group = "chat"
				btChat.Drag_index = i
				btChat.Drop_v = true
				btChat.dropMove = func(src_i, dst_i int, src_source, dst_source string) error {
					Layout_MoveElement(&app.Chats, &app.Chats, src_i, dst_i)
					app.Selected_chat_i = dst_i
					ui.ActivateEditbox("chat_user_prompt", caller)
					return nil
				}

				btClose := TabsDiv.AddButton(1, i, 1, 1, "X")
				btClose.Background = 0.2
				btClose.clicked = func() error {

					//create "trash" folder
					os.MkdirAll(fmt.Sprintf("../%s/Chats/trash", app.Name), os.ModePerm)

					//copy file
					err = OsCopyFile(fmt.Sprintf("../%s/Chats/trash/%s", app.Name, app.Chats[i].FileName),
						fmt.Sprintf("../%s/Chats/%s", app.Name, app.Chats[i].FileName))

					if err != nil {
						return err
					}

					//remove file
					os.Remove(fmt.Sprintf("../%s/Chats/%s", app.Name, app.Chats[i].FileName))

					app.Chats = slices.Delete(app.Chats, i, i+1)
					if i < app.Selected_chat_i {
						app.Selected_chat_i--
					}
					return nil
				}
			}
		}

		//Threads
		{
			msgs := GetMsgs()
			if len(msgs) > 0 {
				SideDiv.SetRowFromSub(2, 0, 5)
				ThreadsDiv := SideDiv.AddLayout(0, 2, 2, 1)
				st.buildThreads(ThreadsDiv, msgs)
			}
		}
	}

	//Chat(or settings)
	{
		if source_root.Mode == "settings" {
			err := st.buildSettings(ui.AddLayout(1, 0, 1, 1), caller, source_root)
			if err != nil {
				return err
			}
		} else if source_chat != nil {

			ChatDiv, err := ui.AddTool(1, 0, 1, 1, (&ShowChat{AppName: app.Name, ChatFileName: chat_fileName}).run, caller)
			if err != nil {
				return err
			}

			ChatDiv.Back_cd = UI_GetPalette().GetGrey(0.03)

			for _, br := range source_chat.Input.Picks {
				ui.Paint_Brush(br.Cd.Cd, br.Points)
			}
		}
	}

	return nil
}

func (st *ShowRoot) refreshChats(app *RootApp, caller *ToolCaller) (*Chat, string, error) {

	chats_folder := fmt.Sprintf("../%s/Chats", app.Name)
	if _, err := os.Stat(chats_folder); os.IsNotExist(err) {
		//no chat folder
		app.Chats = nil
		return nil, "", nil //ok
	}

	fls, err := os.ReadDir(chats_folder)
	if err != nil {
		return nil, "", nil //maybe no chat
	}
	//add new chats
	for _, fl := range fls {
		if fl.IsDir() {
			continue
		}

		found := false
		for _, chat := range app.Chats {
			if chat.FileName == fl.Name() {
				found = true
				break
			}
		}
		if !found {
			app.Chats = append(app.Chats, RootChat{FileName: fl.Name()})
		}
	}
	//remove deleted chats
	for i := len(app.Chats) - 1; i >= 0; i-- {
		found := false
		for _, fl := range fls {
			if fl.IsDir() {
				continue
			}

			if fl.Name() == app.Chats[i].FileName {
				found = true
				break
			}
		}
		if !found {
			app.Chats = slices.Delete(app.Chats, i, i+1)
		}
	}

	//check selecte in range
	if app.Selected_chat_i >= 0 {
		if app.Selected_chat_i >= len(app.Chats) {
			app.Selected_chat_i = len(app.Chats) - 1
		}
	}

	//update and return
	if app.Selected_chat_i >= 0 {
		fileName := app.Chats[app.Selected_chat_i].FileName
		sourceChat, err := NewChat(fmt.Sprintf("../%s/Chats/%s", app.Name, fileName), caller)
		if err != nil {
			return nil, "", err
		}

		if sourceChat != nil {
			//reload
			app.Chats[app.Selected_chat_i].Label = sourceChat.Label
		}

		return sourceChat, fileName, nil
	}

	return nil, "", nil
}

func (st *ShowRoot) refreshApps(source_root *Root) (*RootApp, error) {
	fls, err := os.ReadDir("..")
	if err != nil {
		return nil, err
	}
	//add new apps
	for _, fl := range fls {
		if !fl.IsDir() || fl.Name() == "Root" {
			continue
		}

		found := false
		for _, app := range source_root.Apps {
			if app.Name == fl.Name() {
				found = true
				break
			}
		}
		if !found {
			source_root.Apps = append(source_root.Apps, &RootApp{Name: fl.Name()})
		}
	}
	//remove deleted app
	for i := len(source_root.Apps) - 1; i >= 0; i-- {
		found := false
		for _, fl := range fls {
			if !fl.IsDir() || fl.Name() == "Root" {
				continue
			}

			if fl.Name() == source_root.Apps[i].Name {
				found = true
				break
			}
		}
		if !found {
			source_root.Apps = slices.Delete(source_root.Apps, i, i+1)
		}
	}

	//check selecte in range
	if source_root.Selected_app_i >= 0 {
		if source_root.Selected_app_i >= len(source_root.Apps) {
			source_root.Selected_app_i = len(source_root.Apps) - 1
		}
	}
	//return
	if source_root.Selected_app_i >= 0 {
		return source_root.Apps[source_root.Selected_app_i], nil
	}

	return nil, nil
}

func (st *ShowRoot) buildSettings(ui *UI, caller *ToolCaller, root *Root) error {
	ui.SetColumn(0, 1, 100)
	ui.SetColumn(1, 10, 16)
	ui.SetColumn(2, 1, 100)

	y := 0
	ui.AddTextLabel(1, y, 1, 1, "Settings").Align_h = 1
	y++
	y++

	//device settings
	{
		ui.SetRowFromSub(y, 0, 100)
		ui.AddToolApp(1, y, 1, 1, "Device", "ShowDeviceSettings", nil, caller)
		//ui.AddTool(1, y, 1, 1, (&ShowDeviceSettings{}).run, caller)
		y++
	}

	ui.AddDivider(1, y, 1, 1, true)
	y++

	// LLMs
	{
		tx := ui.AddText(1, y, 1, 1, "LLMs")
		tx.Align_h = 1
		y++

		//xAI
		{
			setDia := ui.AddDialog("xai_settings")
			setDia.UI.SetColumn(0, 1, 20)
			setDia.UI.SetRowFromSub(0, 1, 100)
			setDia.UI.AddTool(0, 0, 1, 1, (&ShowLLMxAISettings{}).run, caller)

			bt := ui.AddButton(1, y, 1, 1, "xAI settings")
			//bt.Align = 0
			bt.Background = 0.5
			bt.clicked = func() error {
				setDia.OpenCentered(caller)
				return nil
			}

			source_llm, err := NewLLMxAI("", caller)
			if err != nil {
				return err
			}
			err = source_llm.Check(caller)
			if err != nil {
				bt.Cd = UI_GetPalette().E
				bt.Tooltip = "Error: " + err.Error()
			}
			y++
		}

		//Whisper.cpp
		{
			setDia := ui.AddDialog("whispercpp_settings")
			setDia.UI.SetColumn(0, 1, 20)
			setDia.UI.SetRowFromSub(0, 1, 100)
			setDia.UI.AddTool(0, 0, 1, 1, (&ShowLLMWhispercppSettings{}).run, caller)

			bt := ui.AddButton(1, y, 1, 1, "Whisper.cpp settings")
			//bt.Align = 0
			bt.Background = 0.5
			bt.clicked = func() error {
				setDia.OpenCentered(caller)
				return nil
			}

			source_wsp, err := NewLLMWhispercpp_wsp("", caller)
			if err != nil {
				return err
			}
			err = source_wsp.Check()
			if err != nil {
				bt.Cd = UI_GetPalette().E
				bt.Tooltip = "Error: " + err.Error()
			}
			y++
		}

		y++ //space
	}

	//Memory
	{
		tx := ui.AddText(1, y, 1, 1, "System prompt extended")
		tx.Align_h = 1
		tx.Tooltip = "Things you want to share with LLM agent.\nThe text is added to every system prompt."
		y++

		ui.SetRowFromSub(y, 2, 5)
		mem := ui.AddEditboxString(1, y, 1, 1, &root.Memory)
		mem.Multiline = true
		mem.Align_v = 0
		mem.Ghost = "Things you want to share with LLM agent."
		y++
	}

	ui.AddDivider(1, y, 1, 1, true)
	y++

	return nil
}

func (st *ShowRoot) buildAbout(ui *UI) {
	ui.SetColumnFromSub(0, 5, 30)

	y := 0

	Version := ui.AddText(0, y, 1, 1, "v0.1") //....
	Version.Align_h = 1
	y++

	Url := ui.AddButton(0, y, 1, 1, "github.com/milansuk/skyalt/")
	Url.Background = 0
	Url.BrowserUrl = "https://github.com/milansuk/skyalt/"
	y++

	Copyright := ui.AddText(0, y, 1, 1, "This program comes with absolutely no warranty.")
	Copyright.Align_h = 1
	y++

	License := ui.AddText(0, y, 1, 1, "This program is distributed under the terms of Apache License, Version 2.0.")
	License.Align_h = 1
	y++
}

func (st *ShowRoot) buildThreads(ui *UI, msgs []SdkMsg) {
	//router := layout.ui.router

	y := 0
	ui.SetColumn(0, 1, 100)
	ui.SetColumn(1, 2, 3)

	//Progress
	for _, msg := range msgs {
		label := msg.GetLabel()
		tx := ui.AddText(0, y, 1, 1, label)
		tx.Tooltip = fmt.Sprintf("%s() - %s", msg.FuncName, label)
		bt := ui.AddButton(1, y, 1, 1, "Cancel")
		bt.Background = 0.5
		bt.clicked = func() error {
			callFuncMsgStop(msg.Id)
			return nil
		}
		y++
	}

	//Errors ....
	/*errs := router.log.GetList(time.Now().Unix() - 10)
	for _, er := range errs {
		tx := ui.AddText(0, y, 2, 1, "Error: "+er.err.Error())
		tx.Cd = UI_GetPalette().E
		y++
	}

	if y > 0 {
		if y == len(errs) {
			ui.SetRefreshDelay(3) //errors
		} else {
			ui.SetRefreshDelay(1) //update
		}
	}*/
}
