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
	AddBrush  *LayoutPick
	RunPrompt string
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

		//add brush
		if st.AddBrush != nil {
			source_chat.Input.MergePick(*st.AddBrush)
			ui.ActivateEditbox("chat_user_prompt", caller)
			st.AddBrush = nil
		}

		//run prompt(from PromptMenu)
		if st.RunPrompt != "" {
			_saveInstances() //save previous chat(and root selection)
			source_chat.Input.Text = st.RunPrompt
			st.RunPrompt = ""

			err := source_chat._sendIt(app.Name, caller, source_root, false)
			if err != nil {
				return err
			}
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

			logoBt := AppsDiv.AddButton(0, y, 1, 1, "")
			y++
			logoBt.Background = 0.2
			logoBt.IconPath = "resources/logo_small.png"
			logoBt.Icon_margin = 0.1
			logoBt.layout.Tooltip = "About"
			logoBt.clicked = func() error {
				AboutDia.OpenRelative(logoBt.layout, caller)
				return nil
			}
		}

		//divider
		{
			AppsDiv.SetRow(y, 0.1, 0.1)
			AppsDiv.AddDivider(0, y, 1, 1, true)
			y++
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
					dd := 1.0

					Apps2Div.SetRowFromSub(yy, 1, 100, true)

					BtDiv := Apps2Div.AddLayout(0, yy, 1, 1)
					BtDiv.SetColumn(0, 1, 100)
					BtDiv.SetRow(0, d, d)
					BtDiv.SetRow(1, dd, dd)
					BtDiv.Back_cd = Color_Aprox(UI_GetPalette().P, UI_GetPalette().B, 0.6) //UI_GetPalette().P
					BtDiv.Back_rounding = true

					bt = BtDiv.AddButton(0, 0, 1, 1, "")

					//Dev button
					btDev := BtDiv.AddButton(0, 1, 1, 1, "Build")
					btDev.layout.Tooltip = "Edit app"
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

				bt.Background = 0
				/*if i == source_root.Selected_app_i && !source_root.ShowSettings {
					bt.Background = 1
				}*/
				bt.layout.Tooltip = app.Name
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
				bt.dropMove = func(src_i, dst_i int, aim_i int, src_source, dst_source string) error {
					Layout_MoveElement(&source_root.Apps, &source_root.Apps, src_i, dst_i)
					source_root.Selected_app_i = dst_i
					ui.ActivateEditbox("chat_user_prompt", caller)
					return nil
				}

				yy++
			}
		}

		//divider
		{
			AppsDiv.SetRow(y, 0.1, 0.1)
			AppsDiv.AddDivider(0, y, 1, 1, true)
			y++
		}

		//Progress
		{
			msgs := callFuncGetMsgs()
			if len(msgs) > 0 {
				ProgressDia := AppsDiv.AddDialog("progress")
				ProgressDia.UI.SetColumn(0, 5, 15)
				ProgressDia.UI.SetRowFromSub(0, 1, 10, true)
				st.buildMessages(ProgressDia.UI.AddLayout(0, 0, 1, 1), msgs)

				ProgressBt := AppsDiv.AddButton(0, y, 1, 1, "")
				y++
				ProgressBt.Background = 0.25
				ProgressBt.IconPath = "resources/warning.png"
				ProgressBt.Icon_margin = 0.2
				ProgressBt.layout.Tooltip = fmt.Sprintf("%d new errors", len(msgs))
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
			newAppBt.layout.Tooltip = "Create new app"
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
			setBt.layout.Tooltip = "Show Settings"
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
			UsageBt.layout.Tooltip = "Spendings"
			UsageBt.clicked = func() error {
				UsageDia.OpenRelative(UsageBt.layout, caller)
				return nil
			}
		}

		//Log/Errors
		{
			logs := callFuncGetLogs()
			LogsDia := AppsDiv.AddDialog("logs")
			st.buildLog(&LogsDia.UI, logs, caller)

			hasNewErrs := (len(logs) > 0 && logs[len(logs)-1].Time > source_root.Last_log_time)
			label := "LOG"
			if hasNewErrs {
				n := 0
				for _, it := range logs {
					if it.Time > source_root.Last_log_time {
						n++
					}
				}
				label = fmt.Sprintf("(%d)", n)
			}

			LogBt := AppsDiv.AddButton(0, y, 1, 1, label)
			y++
			LogBt.Background = 0.25
			if hasNewErrs {
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

		//Microphone status
		micInfo := callFuncGetMicInfo()
		if micInfo.Active {
			MicBt := AppsDiv.AddButton(0, y, 1, 1, "")
			MicBt.layout.Tooltip = "Stop all microphone recordings"
			MicBt.IconPath = "resources/mic.png"
			MicBt.Icon_margin = 0.15
			MicBt.Border = true

			//animate icon color
			MicBt.layout.update = func() error {
				micInfo := callFuncGetMicInfo()

				MicBt.Cd = UI_GetPalette().E
				MicBt.Cd.A = byte(255 * micInfo.Decibels_normalized)
				if MicBt.Cd.A == 0 {
					MicBt.Cd.A = 1 //because 0=off
				}
				//MicBt.Cd = Color_Aprox(UI_GetPalette().B, UI_GetPalette().E, micInfo.Decibels_normalized)

				return nil
			}
			y++
			MicBt.clicked = func() error {
				callFuncStopMic()
				return nil
			}
		}

		//Media status
		mediaInfo := callFuncGetMediaInfo()
		if len(mediaInfo) > 0 {

			MediaDia := AppsDiv.AddDialog("media")
			{
				MediaDia.UI.SetColumnFromSub(0, 1, 20, true)
				MediaDia.UI.SetColumn(1, 1, 2)
				MediaDia.UI.SetColumn(2, 2, 5)
				yy := 0
				for playerID, it := range mediaInfo {
					//path
					tx := MediaDia.UI.AddText(0, yy, 1, 1, it.Path)
					tx.layout.Tooltip = fmt.Sprintf("%d", playerID)

					//time
					MediaDia.UI.AddText(1, yy, 1, 1, fmt.Sprintf("%d%%", int(float64(it.Seek)/float64(it.Duration)*100)))

					//pause ....

					//stop ....

					//volume ....

					vol := it.Volume
					volume := MediaDia.UI.AddSlider(2, yy, 1, 1, &vol, 0, 1, 0.1)
					volume.changed = func() error {
						//send back new value ....
						return nil
					}

					yy++
				}

				//stop all ....
				//pause all ....
			}

			MediaBt := AppsDiv.AddButton(0, y, 1, 1, "")
			MediaBt.layout.Tooltip = "Show media"
			MediaBt.IconPath = "resources/speaker.png"
			MediaBt.Icon_margin = 0.2
			MediaBt.Background = 0.25
			y++
			MediaBt.clicked = func() error {
				MediaDia.OpenRelative(MediaBt.layout, caller)
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
			ChatDiv.SetColumnResizable(0, 8, 20, 8)
			ChatDiv.SetColumn(1, 1, 100)
			ChatDiv.SetRow(0, 1, 100)

			//Chat(or settings)
			//note: must be called before, because it will update chat label
			{
				ChatDiv, err := ChatDiv.AddTool(1, 0, 1, 1, (&ShowApp{AppName: app.Name, ChatFileName: chat_fileName}).run, caller)
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

			//Side
			{
				SideDiv := ChatDiv.AddLayout(0, 0, 1, 1)
				SideDiv.SetColumn(0, 1, 100)
				SideDiv.SetRow(1, 1, 100)

				//Header
				{
					SideDiv.SetRow(0, 1, 1.5)
					HeaderDiv := SideDiv.AddLayout(0, 0, 1, 1)
					HeaderDiv.SetRow(0, 1, 100)
					HeaderDiv.ScrollH.Narrow = true
					HeaderDiv.ScrollV.Hide = true
					//New Tab
					{
						HeaderDiv.SetColumn(0, 2, 100)
						bt := HeaderDiv.AddButton(0, 0, 1, 1, "New Tab")
						bt.Background = 0.5
						bt.layout.Tooltip = "Create new tab"
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

							pos := app.NumPins() //skip pins
							app.Chats = slices.Insert(app.Chats, pos, RootChat{Label: "Empty", FileName: fileName})
							app.Selected_chat_i = pos

							ui.ActivateEditbox("chat_user_prompt", caller)

							SideDiv.VScrollToTheTop(caller)

							source_root.ShowSettings = false

							//if exist, prepare StartPrompt to run
							{
								type SdkToolsPrompts struct {
									StartPrompt string
									//..
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

								if sdk_app.StartPrompt != "" {
									_saveInstances() //save previous chat(and root selection)

									source_chat.Input.Text = sdk_app.StartPrompt

									//ch := ShowChat{AppName: app.Name, ChatFileName: fileName}
									return source_chat._sendIt(app.Name, caller, source_root, false)
								}
							}

							return nil
						}
					}

					HeaderDiv.SetColumn(1, 0.5, 0.5) //space

					//navigation
					if source_chat != nil {
						numUseMessages := source_chat.GetNumUserMessages()

						//HeaderDiv.SetColumn(1, 0.5, 0.5) //space

						HeaderDiv.SetColumnFromSub(2, 1, 100, true)
						NavDiv := HeaderDiv.AddLayout(2, 0, 1, 1)
						NavDiv.SetRow(0, 1, 100)

						nx := 0

						/*NavDiv.SetColumn(nx, 1.5, 1.5)
						homeBt := NavDiv.AddButton(nx, 0, 1, 1, "")
						nx++
						homeBt.layout.Enable = (source_chat.User_msg_i > 0)
						homeBt.layout.Tooltip = "First dashboard"
						homeBt.Background = 0.5
						homeBt.IconPath = "resources/home.png"
						homeBt.Icon_margin = 0.5
						homeBt.clicked = func() error {
							source_chat.User_msg_i = 0
							return nil
						}
						*/

						NavDiv.SetColumn(nx, 1.5, 1.5)
						backBt := NavDiv.AddButton(nx, 0, 1, 1, "<")
						nx++
						backBt.layout.Tooltip = "Previous dashboard"
						backBt.Background = 0.5
						backBt.layout.Enable = (source_chat.User_msg_i > 0)
						backBt.clicked = func() error {
							source_chat.User_msg_i--
							return nil
						}

						NavDiv.SetColumnFromSub(nx, 1.5, 100, true)
						inf := NavDiv.AddText(nx, 0, 1, 1, fmt.Sprintf("%d/%d", source_chat.User_msg_i+1, numUseMessages)) //...
						nx++
						inf.Align_h = 1
						//inf.layout.Tooltip = dashes.UI_func + "()"

						NavDiv.SetColumn(nx, 1.5, 1.5)
						forwardBt := NavDiv.AddButton(nx, 0, 1, 1, ">")
						nx++
						forwardBt.layout.Tooltip = "Next dashboard"
						forwardBt.Background = 0.5
						forwardBt.layout.Enable = (source_chat.User_msg_i+1 < numUseMessages)
						forwardBt.clicked = func() error {
							source_chat.User_msg_i++
							return nil
						}
					}
				}

				//List of tabs
				ListsDiv := SideDiv.AddLayout(0, 1, 1, 1)
				ListsDiv.SetColumn(0, 1, 100)

				var PinnedDiv *UI
				var TabsDiv *UI

				num_pins := app.NumPins()
				if num_pins > 0 {
					ListsDiv.SetRow(0, 0.7, 0.7)
					ListsDiv.AddDivider(0, 0, 1, 1, true).Label = "<small>Pins"
					ListsDiv.SetRowFromSub(1, 1, 100, true)
					PinnedDiv = ListsDiv.AddLayout(0, 1, 1, 1)
					PinnedDiv.SetColumn(1, 1, 100)

					ListsDiv.SetRow(2, 0.7, 0.7)
					ListsDiv.AddDivider(0, 2, 1, 1, true).Label = "<small>Tabs"
					ListsDiv.SetRowFromSub(3, 1, 100, true)
					TabsDiv = ListsDiv.AddLayout(0, 3, 1, 1)
					TabsDiv.SetColumn(1, 1, 100)

				} else {
					ListsDiv.SetRow(0, 1, 100)
					TabsDiv = ListsDiv.AddLayout(0, 0, 1, 1)
					TabsDiv.SetColumn(1, 1, 100)
				}

				yPinned := 0
				yTabs := 0
				for i, tab := range app.Chats {

					isSelected := (i == app.Selected_chat_i)

					var btChat *UIButton
					var btPin *UIButton
					if tab.Pinned {
						btPin = PinnedDiv.AddButton(0, yPinned, 1, 1, "")
						btPin.IconPath = "resources/unpin.png"
						btPin.layout.Tooltip = "Un-pin tab"

						btChat = PinnedDiv.AddButton(1, yPinned, 1, 1, tab.Label)
					} else {
						btPin = TabsDiv.AddButton(0, yTabs, 1, 1, "")
						btPin.IconPath = "resources/pin.png"
						btPin.layout.Tooltip = "Pin tab"

						btChat = TabsDiv.AddButton(1, yTabs, 1, 1, tab.Label)
					}

					btPin.Icon_margin = 0.25
					btPin.Background = 0.2
					btPin.clicked = func() error {
						if tab.Pinned {
							//move down
							app.Chats = slices.Insert(app.Chats, num_pins-1, tab)
							app.Chats = slices.Delete(app.Chats, i, i+1)
						}
						app.Chats[i].Pinned = !tab.Pinned
						return nil
					}

					btChat.layout.Tooltip = tab.Label
					btChat.Align = 0
					btChat.Background = 0.2
					if isSelected {
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
					btChat.dropMove = func(src_i, dst_i int, aim_i int, src_source, dst_source string) error {

						app.Chats[src_i].Pinned = (aim_i < num_pins)

						Layout_MoveElement(&app.Chats, &app.Chats, src_i, dst_i)

						if app.Selected_chat_i != dst_i {
							ui.ActivateEditbox("chat_user_prompt", caller)
						}
						app.Selected_chat_i = dst_i

						return nil
					}

					if isSelected { //show only when open
						//close
						var btClose *UIButton
						if tab.Pinned {
							btClose = PinnedDiv.AddButton(2, yPinned, 1, 1, "X")
						} else {
							btClose = TabsDiv.AddButton(2, yTabs, 1, 1, "X")
						}

						btClose.Background = 0.2
						btClose.clicked = func() error {
							return app.RemoveChat(app.Chats[i])
						}
					}

					if tab.Pinned {
						yPinned++
					} else {
						yTabs++
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
		ui.SetRowFromSub(y, 0, 100, true)
		ui.AddToolApp(1, y, 1, 1, "Device", "ShowDeviceSettings", nil, caller)
		y++
	}

	ui.AddDivider(1, y, 1, 1, true)
	y++

	// LLMs
	{
		ui.SetRowFromSub(y, 0, 100, true)
		ui.AddToolApp(1, y, 1, 1, "Device", "ShowLLMsSettings", nil, caller)
		y++
	}

	ui.AddDivider(1, y, 1, 1, true)
	y++

	return nil
}

func (st *ShowRoot) buildAbout(ui *UI) {
	ui.SetColumnFromSub(0, 5, 30, true)

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

	ui.SetColumnFromSub(0, 1, 30, true)
	ui.SetRowFromSub(0, 1, 15, true)

	total_price := 0.0
	{
		ItemsDiv := ui.AddLayout(0, 0, 1, 1)
		ItemsDiv.SetColumnFromSub(0, 1, 10, true)
		ItemsDiv.SetColumnFromSub(1, 1, 10, true)
		ItemsDiv.SetColumnFromSub(2, 1, 10, true)
		ItemsDiv.SetColumnFromSub(3, 1, 10, true)

		y := 0
		for i := len(usages) - 1; i >= 0; i-- {
			usg := &usages[i]

			ItemsDiv.AddText(0, y, 1, 1, usg.Model)
			ItemsDiv.AddText(1, y, 1, 1, SdkGetDateTime(int64(usg.CreatedTimeSec)))
			ItemsDiv.AddText(2, y, 1, 1, fmt.Sprintf("%.0fsec", usg.DTime))

			price := (usg.Prompt_price + usg.Input_cached_price + usg.Completion_price + usg.Reasoning_price)
			ItemsDiv.AddText(3, y, 1, 1, fmt.Sprintf("$%f", price))
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

	{
		HeaderDiv := ui.AddLayout(0, 0, 1, 1)
		HeaderDiv.SetColumn(0, 5, 100)
		HeaderDiv.SetColumn(1, 3, 3)
		HeaderDiv.SetColumn(2, 3, 5)

		HeaderDiv.AddTextLabel(0, 0, 1, 1, "Logs")

		ClearBt := HeaderDiv.AddButton(1, 0, 1, 1, "Clear")
		ClearBt.layout.Enable = (len(logs) > 0)
		ClearBt.Background = 0.5
		ClearBt.clicked = func() error {
			clearLogs()
			return nil
		}

		CopyBt := HeaderDiv.AddButton(2, 0, 1, 1, "Copy to clipboard")
		CopyBt.layout.Enable = (len(logs) > 0)
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

	ui.SetRowFromSub(1, 1, 15, true)
	ListDiv := ui.AddLayout(0, 1, 1, 1)
	ListDiv.SetColumn(0, 1, 4)
	ListDiv.SetColumn(1, 1, 26)

	MAX_N := 20
	y := 0
	for i := len(logs) - 1; i >= 0 && i >= len(logs)-MAX_N; i-- {
		it := logs[i]

		ListDiv.SetRowFromSub(y, 1, 5, true)

		ListDiv.AddText(0, y, 1, 1, SdkGetDateTime(int64(it.Time)))

		tx := ListDiv.AddText(1, y, 1, 1, it.Msg)
		tx.layout.Tooltip = it.Stack
		tx.Cd = UI_GetPalette().E

		y++
	}
	if len(logs) > MAX_N {
		ListDiv.AddText(0, y, 2, 1, "...").Align_h = 1
	}

}

func (st *ShowRoot) buildMessages(ui *UI, msgs []SdkMsg) {
	y := 0
	ui.SetColumn(0, 3, 100)
	ui.SetColumn(1, 2, 3)

	//Progress
	for _, msg := range msgs {
		label := msg.GetLabel()
		ui.SetRowFromSub(y, 1, 5, true)
		tx := ui.AddText(0, y, 1, 1, label)
		tx.layout.Tooltip = fmt.Sprintf("%s() - %s", msg.ToolName, label)
		bt := ui.AddButton(1, y, 1, 1, "Cancel")
		bt.Background = 0.5
		bt.clicked = func() error {
			callFuncMsgStop(msg.Id)
			return nil
		}
		y++
	}
}
