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
	source_root, err := NewRoot("")
	if err != nil {
		return err
	}

	//refresh apps
	app, err := source_root.refreshApps()
	if err != nil {
		return err
	}

	//load chat
	var source_chat *Chat
	var chat_fileName string
	if app != nil {
		source_chat, chat_fileName, err = app.refreshChats(caller)
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

	d := 1.5
	ui.SetColumn(0, d, d)
	ui.SetColumn(1, 1, 100)
	ui.SetRow(0, 1, 100)

	//Apps
	{
		AppsDiv := ui.AddLayout(0, 0, 1, 1)
		AppsDiv.SetColumn(0, 1, 100)
		AppsDiv.SetRow(1, 1, 100)
		AppsDiv.Back_cd = UI_GetPalette().GetGrey(0.1)

		//Logo
		{
			AboutDia := AppsDiv.AddDialog("about")
			st.buildAbout(&AboutDia.UI)

			//logo := HeaderDiv.AddImagePath(x, 0, 1, 1, "resources/logo_small.png")
			logoBt := AppsDiv.AddButton(0, 0, 1, 1, "")
			logoBt.Icon_align = 1
			logoBt.Background = 0.2
			logoBt.IconPath = "resources/logo_small.png"
			logoBt.Icon_margin = 0.1
			logoBt.Tooltip = "About"
			logoBt.clicked = func() error {
				AboutDia.OpenRelative(logoBt.layout, caller)
				return nil
			}
		}

		//Settings
		{
			setBt := AppsDiv.AddButton(0, 2, 1, 1, "")
			setBt.IconPath = "resources/settings.png"
			setBt.Icon_margin = 0.2
			setBt.Icon_align = 1
			setBt.Tooltip = "Show Settings"
			setBt.Background = 0.25
			if source_root.ShowSettings {
				setBt.Background = 1
			}
			setBt.clicked = func() error {
				source_root.ShowSettings = !source_root.ShowSettings
				return nil
			}
		}

		//Error
		{
			//....
		}

		//Progress
		{
			msgs := GetMsgs()
			if len(msgs) > 0 {
				AboutDia := AppsDiv.AddDialog("about")
				AboutDia.UI.SetColumn(0, 5, 15)
				AboutDia.UI.SetRowFromSub(0, 1, 10)
				st.buildThreads(AboutDia.UI.AddLayout(0, 0, 1, 1), msgs)

				ProgressBt := AppsDiv.AddButton(0, 3, 1, 1, fmt.Sprintf("[%d]", len(msgs)))
				ProgressBt.Background = 0
				ProgressBt.clicked = func() error {
					AboutDia.OpenRelative(ProgressBt.layout, caller)
					return nil
				}
			}
		}

		//Apps
		{
			Apps2Div := AppsDiv.AddLayout(0, 1, 1, 1)
			Apps2Div.SetColumn(0, 1, 100)
			Apps2Div.ScrollV.Narrow = true
			Apps2Div.SetColumn(0, 1, 100)
			y := 0
			for i, app := range source_root.Apps {
				var bt *UIButton

				if i == source_root.Selected_app_i && !source_root.ShowSettings {
					dd := 0.8
					Apps2Div.SetRow(y, d+dd, d+dd)

					BtDiv := Apps2Div.AddLayout(0, y, 1, 1)
					BtDiv.SetColumn(0, 1, 100)
					BtDiv.SetRow(0, d, d)
					BtDiv.SetRow(1, dd, dd)
					BtDiv.Back_cd = UI_GetPalette().P

					bt = BtDiv.AddButton(0, 0, 1, 1, "")

					//Dev button
					btDev := BtDiv.AddButton(0, 1, 1, 1, "Build")
					btDev.Tooltip = "Edit app"
					btDev.Shortcut = 'b'
					if app.Dev.Enable {
						btDev.Background = 0.5
						//btDev.Cd = UI_GetPalette().S
						btDev.Label = "<b>" + btDev.Label
					}
					btDev.clicked = func() error {
						app.Dev.Enable = !app.Dev.Enable
						source_root.Selected_app_i = i
						source_root.ShowSettings = false
						return nil
					}

				} else {
					Apps2Div.SetRow(y, d, d)

					bt = Apps2Div.AddButton(0, y, 1, 1, "")
				}

				bt.Icon_align = 1
				bt.Background = 0.2
				if i == source_root.Selected_app_i && !source_root.ShowSettings {
					bt.Background = 1
				}
				bt.Tooltip = app.Name
				bt.IconPath = fmt.Sprintf("apps/%s/icon.png", app.Name)
				bt.Icon_margin = 0.4

				bt.clicked = func() error {
					if source_root.Selected_app_i == i {
						app.Dev.Enable = false
					}
					source_root.Selected_app_i = i
					source_root.ShowSettings = false
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

				y++
			}
		}
	}

	if source_root.ShowSettings {
		err := st.buildSettings(ui.AddLayout(1, 0, 1, 1), caller, source_root)
		if err != nil {
			return err
		}
	} else {

		if app == nil {
			return nil
		}

		if app.Dev.Enable {
			_, err := ui.AddTool(1, 0, 1, 1, (&ShowDev{AppName: app.Name}).run, caller)
			if err != nil {
				return err
			}

		} else {
			ChatDiv := ui.AddLayout(1, 0, 1, 1)
			ChatDiv.SetColumnResizable(0, 6, 15, 6)
			ChatDiv.SetColumn(1, 1, 100)
			ChatDiv.SetRow(0, 1, 100)

			//Side
			{
				SideDiv := ChatDiv.AddLayout(0, 0, 1, 1)
				SideDiv.SetColumn(0, 1, 100)
				SideDiv.SetRow(1, 1, 100)

				//New Chat button
				{
					bt := SideDiv.AddButton(0, 0, 1, 1, "New Chat")
					bt.Background = 0.5
					bt.Tooltip = "Create new chat"
					bt.Shortcut = 't'
					bt.layout.Enable = (app != nil)
					bt.clicked = func() error {
						if app == nil {
							return fmt.Errorf("No app selected")
						}

						fileName := fmt.Sprintf("Chat-%d.json", time.Now().UnixMicro())
						source_chat, err = NewChat(fmt.Sprintf("../%s/Chats/%s", app.Name, fileName))
						if err != nil {
							return nil
						}

						app.Chats = slices.Insert(app.Chats, 0, RootChat{Label: "Empty chat", FileName: fileName})
						app.Selected_chat_i = 0
						ui.ActivateEditbox("chat_user_prompt", caller)

						SideDiv.VScrollToTheTop(caller)

						source_root.ShowSettings = false

						return nil
					}
				}

				//List of tabs
				{
					TabsDiv := SideDiv.AddLayout(0, 1, 1, 1)
					TabsDiv.SetColumn(0, 1, 100)

					y := 0
					for i, tab := range app.Chats {

						btChat := TabsDiv.AddButton(0, y, 1, 1, tab.Label)
						btChat.Align = 0
						btChat.Background = 0.2
						if i == app.Selected_chat_i {
							btChat.Background = 1 //selected
						}
						btChat.clicked = func() error {
							app.Selected_chat_i = i
							source_root.ShowSettings = false
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

						btClose := TabsDiv.AddButton(1, y, 1, 1, "X")
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

						y++
					}
				}
			}

			//Chat(or settings)
			{
				ChatDiv, err := ChatDiv.AddTool(1, 0, 1, 1, (&ShowChat{AppName: app.Name, ChatFileName: chat_fileName}).run, caller)
				if err != nil {
					return err
				}

				ChatDiv.Back_cd = UI_GetPalette().GetGrey(0.03)

				if source_chat != nil {
					for _, br := range source_chat.Input.Picks {
						ui.Paint_Brush(br.Cd.Cd, br.Points)
					}
				}
			}
		}
	}

	return nil
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

			source_llm, err := NewLLMxAI("")
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

			source_wsp, err := NewLLMWhispercpp_wsp("")
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
		tx := ui.AddText(1, y, 1, 1, "Extended system prompt - memory")
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
	y := 0
	ui.SetColumn(0, 3, 100)
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
