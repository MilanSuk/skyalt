package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
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
		source_chat, chat_fileName, err = app.refreshChats()
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
		AppsDiv.Back_cd = UI_GetPalette().GetGrey(0.1)

		y := 0

		//Logo
		{
			AboutDia := AppsDiv.AddDialog("about")
			st.buildAbout(&AboutDia.UI)

			//logo := HeaderDiv.AddImagePath(x, 0, 1, 1, "resources/logo_small.png")
			logoBt := AppsDiv.AddButton(0, y, 1, 1, "")
			y++
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

		//Apps
		{
			AppsDiv.SetRow(y, 1, 100)
			Apps2Div := AppsDiv.AddLayout(0, y, 1, 1)
			y++
			Apps2Div.SetColumn(0, 1, 100)
			Apps2Div.ScrollV.Narrow = true
			Apps2Div.SetColumn(0, 1, 100)
			yy := 0
			for i, app := range source_root.Apps {
				var bt *UIButton

				if i == source_root.Selected_app_i && !source_root.ShowSettings {
					dd := 0.8
					Apps2Div.SetRowFromSub(yy, 1, 100)

					BtDiv := Apps2Div.AddLayout(0, yy, 1, 1)
					BtDiv.SetColumn(0, 1, 100)
					BtDiv.SetRow(0, d, d)
					BtDiv.SetRow(1, dd, dd)
					BtDiv.Back_cd = Color_Aprox(UI_GetPalette().P, UI_GetPalette().B, 0.6) //UI_GetPalette().P
					BtDiv.Back_rounding = true

					bt = BtDiv.AddButton(0, 0, 1, 1, "")

					//Dev button
					btDev := BtDiv.AddButton(0, 1, 1, 1, "Build")
					btDev.Tooltip = "Edit app"
					btDev.Shortcut = 'b'
					btDev.Background = 0
					if app.Dev.Enable {
						btDev.Background = 1 //0.5
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
					Apps2Div.SetRow(yy, d, d)

					bt = Apps2Div.AddButton(0, yy, 1, 1, "")
				}

				bt.Icon_align = 1
				bt.Background = 0
				/*if i == source_root.Selected_app_i && !source_root.ShowSettings {
					bt.Background = 1
				}*/
				bt.Tooltip = app.Name
				bt.IconPath = fmt.Sprintf("apps/%s/icon", app.Name)
				bt.Icon_margin = 0.4

				bt.clicked = func() error {
					if source_root.Selected_app_i == i && !source_root.ShowSettings {
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

				yy++
			}
		}

		//Progress
		{
			msgs := callFuncGetMsgs()
			if len(msgs) > 0 {
				ProgressDia := AppsDiv.AddDialog("progress")
				ProgressDia.UI.SetColumn(0, 5, 15)
				ProgressDia.UI.SetRowFromSub(0, 1, 10)
				st.buildThreads(ProgressDia.UI.AddLayout(0, 0, 1, 1), msgs)

				ProgressBt := AppsDiv.AddButton(0, y, 1, 1, fmt.Sprintf("[%d]", len(msgs)))
				y++
				ProgressBt.Background = 0
				ProgressBt.clicked = func() error {
					ProgressDia.OpenRelative(ProgressBt.layout, caller)
					return nil
				}
			}
		}

		//create an app
		{
			newAppBt := AppsDiv.AddButton(0, y, 1, 1, "<h1>+")
			y++
			newAppBt.Tooltip = "Create new app"
			newAppBt.Background = 0.25
			newAppBt.clicked = func() error {

				appName := st.findUniqueAppName("app", source_root)
				if appName != "" {
					os.MkdirAll(filepath.Join("..", appName), os.ModePerm)
					os.WriteFile(filepath.Join("..", appName, "skyalt"), []byte(""), 0644)

					//refresh apps & select and set new one
					source_root.refreshApps()
					source_root.ShowSettings = false
					source_root.Selected_app_i = len(source_root.Apps) - 1
					source_root.Apps[len(source_root.Apps)-1].Dev.Enable = true
				}
				return nil
			}
		}

		//Settings
		{
			setBt := AppsDiv.AddButton(0, y, 1, 1, "")
			y++
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

		//Spendings
		{
			usageJs := callFuncGetLLMUsage()
			UsageDia := AppsDiv.AddDialog("usage")
			st.buildUsage(&UsageDia.UI, usageJs)

			UsageBt := AppsDiv.AddButton(0, y, 1, 1, "$")
			y++
			UsageBt.Background = 0.25
			UsageBt.Tooltip = "Spendings"
			UsageBt.clicked = func() error {
				UsageDia.OpenRelative(UsageBt.layout, caller)
				return nil
			}
		}
		//Log/Errors
		{
			logs := callFuncGetLogs()
			LogsDia := AppsDiv.AddDialog("progress")
			st.buildLog(&LogsDia.UI, logs, caller)

			LogBt := AppsDiv.AddButton(0, y, 1, 1, "LOG")
			y++
			LogBt.Background = 0.25
			if len(logs) > 0 && logs[len(logs)-1].Time > source_root.Last_log_time {
				LogBt.Background = 1
				LogBt.Cd = UI_GetPalette().E
			}
			LogBt.clicked = func() error {
				LogsDia.OpenRelative(LogBt.layout, caller)

				if len(logs) > 0 {
					source_root.Last_log_time = logs[len(logs)-1].Time
				}
				return nil
			}
		}

	}

	if source_root.ShowSettings {
		err := st.buildSettings(ui.AddLayout(1, 0, 1, 1), caller)
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
						source_chat, err = NewChat(filepath.Join("..", app.Name, "Chats", fileName))
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
							os.MkdirAll(filepath.Join("..", app.Name, "Chats", "trash"), os.ModePerm)

							//copy file
							err = OsCopyFile(filepath.Join("..", app.Name, "Chats", "trash", app.Chats[i].FileName),
								filepath.Join("..", app.Name, "Chats", app.Chats[i].FileName))
							if err != nil {
								return err
							}

							//remove file
							os.Remove(filepath.Join("..", app.Name, "Chats", app.Chats[i].FileName))

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

func (st *ShowRoot) findUniqueAppName(prefix string, root *Root) string {
	id := 1
	for id < 1000 {
		name := fmt.Sprintf("%s_%d", prefix, id)
		if !root.IsAppExist(name) {
			return name
		}
		id++
	}

	return ""
}

func (st *ShowRoot) buildSettings(ui *UI, caller *ToolCaller) error {
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
		y++
	}

	ui.AddDivider(1, y, 1, 1, true)
	y++

	// LLMs
	{
		ui.SetRowFromSub(y, 0, 100)
		ui.AddToolApp(1, y, 1, 1, "Device", "ShowLLMsSettings", nil, caller)
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

func (st *ShowRoot) buildUsage(ui *UI, usageJs []byte) {

	var usages []LLMMsgUsage
	err := json.Unmarshal(usageJs, &usages)
	if err != nil {
		//err ....
		return
	}

	ui.SetColumn(0, 1, 10)
	ui.SetRowFromSub(0, 1, 15)

	total_price := 0.0
	{
		ListDiv := ui.AddLayout(0, 0, 1, 1)
		ListDiv.SetColumnFromSub(0, 1, 6)
		ListDiv.SetColumnFromSub(1, 1, 5)
		ListDiv.SetColumnFromSub(2, 1, 2)
		ListDiv.SetColumnFromSub(3, 1, 4)

		y := 0
		for i := len(usages) - 1; i >= 0; i-- {
			usg := &usages[i]

			ListDiv.AddText(0, y, 1, 1, usg.Model)
			ListDiv.AddText(1, y, 1, 1, SdkGetDateTime(int64(usg.CreatedTimeSec)))
			ListDiv.AddText(2, y, 1, 1, fmt.Sprintf("%.0fsec", usg.DTime))

			price := (usg.Prompt_price + usg.Input_cached_price + usg.Completion_price + usg.Reasoning_price)
			ListDiv.AddText(3, y, 1, 1, fmt.Sprintf("$%f", price))
			total_price += price

			y++
		}
	}

	//space
	ui.SetRow(1, 0.1, 0.1)
	ui.AddDivider(0, 1, 1, 1, true)

	//Sum
	ui.AddText(0, 2, 1, 1, fmt.Sprintf("Total(%d): $%f", len(usages), total_price)).Align_h = 2

	//Note
	noteTx := ui.AddText(0, 3, 1, 1, "<i>numbers may not be accurate.")
	noteTx.Align_h = 1
	noteTx.Cd = UI_GetPalette().GetGrey(0.5)
}

func (st *ShowRoot) buildLog(ui *UI, logs []SdkLog, caller *ToolCaller) {
	ui.SetColumn(0, 1, 20)

	slices.Reverse(logs) //newest top
	if len(logs) > 20 {
		logs = logs[:20] //cut
	}

	{
		HeaderDiv := ui.AddLayout(0, 0, 1, 1)
		HeaderDiv.SetColumn(0, 5, 100)
		HeaderDiv.SetColumn(1, 3, 5)

		HeaderDiv.AddTextLabel(0, 0, 1, 1, "Logs")

		CopyBt := HeaderDiv.AddButton(1, 0, 1, 1, "Copy to clipboard")
		CopyBt.Background = 0.5
		CopyBt.clicked = func() error {
			var str strings.Builder
			for _, it := range logs {
				str.WriteString("Msg:")
				str.WriteString(it.Msg)
				str.WriteString("\n")
				str.WriteString("Stack:")
				str.WriteString(it.Stack)
				str.WriteString("\n")
			}

			caller.SetClipboardText(str.String())
			return nil
		}
	}

	ui.SetRowFromSub(1, 1, 15)
	ListDiv := ui.AddLayout(0, 1, 1, 1)
	ListDiv.SetColumn(0, 1, 4)
	ListDiv.SetColumn(1, 1, 26)

	y := 0
	for _, it := range logs {
		ListDiv.SetRowFromSub(y, 1, 5)

		ListDiv.AddText(0, y, 1, 1, SdkGetDateTime(int64(it.Time)))

		tx := ListDiv.AddText(1, y, 1, 1, it.Msg)
		tx.Tooltip = it.Stack
		tx.Cd = UI_GetPalette().E

		y++
	}

}

func (st *ShowRoot) buildThreads(ui *UI, msgs []SdkMsg) {
	y := 0
	ui.SetColumn(0, 3, 100)
	ui.SetColumn(1, 2, 3)

	//Progress
	for _, msg := range msgs {
		label := msg.GetLabel()
		ui.SetRowFromSub(y, 1, 5)
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
}
