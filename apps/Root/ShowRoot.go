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

	//run prompt(from PromptMenu)
	if st.RunPrompt != "" && source_chat != nil {
		_saveInstances() //save previous chat(and root selection)
		source_chat.Input.Text = st.RunPrompt
		st.RunPrompt = ""

		err := source_chat._sendIt(app.Name, caller, source_root, false)
		if err != nil {
			return err
		}
	}

	//save brush
	activate_prompt := false
	if source_chat != nil {
		//add brush
		if st.AddBrush != nil {
			st.AddBrush.Dash_i = source_chat.Selected_user_msg
			source_chat.Input.MergePick(*st.AddBrush)
			activate_prompt = true
			st.AddBrush = nil
		}
	}

	d := 1.5
	ui.SetColumn(0, d, d)
	ui.SetColumn(1, 1, Layout_MAX_SIZE)
	ui.SetRow(0, 1, Layout_MAX_SIZE)

	if source_root.EnableTextHighlighting {
		SearchDiv := ui.AddLayout(0, 1, 2, 1)
		SearchDiv.SetColumnFromSub(0, 1, 10, true)
		SearchDiv.SetColumn(1, 1, 100)

		SearchDiv.AddText(0, 0, 1, 1, "Search")
		SearchEditbox := SearchDiv.AddEditboxString(1, 0, 1, 1, &source_root.TextHighlighting)
		SearchEditbox.ActivateOnCreate = true

		callFuncSetTextHighlight(source_root.TextHighlighting)
	} else {
		callFuncSetTextHighlight("")
	}

	switch source_root.Show {
	case "chats":
		ui.AddToolApp(1, 0, 1, 1, "Chats", "Chats", "ShowChats", nil, caller)

	case "settings":
		err := st.buildSettings(ui.AddLayoutWithName(1, 0, 1, 1, "Settings"), caller)
		if err != nil {
			return err
		}
	default:
		if app != nil {

			if app.Dev.Enable {
				_, err := ui.AddTool(1, 0, 1, 1, fmt.Sprintf("dev_%s", app.Name), (&ShowDev{AppName: app.Name}).run, caller)
				if err != nil {
					return err
				}

			} else {
				AppDiv := ui.AddLayoutWithName(1, 0, 1, 1, "App")
				AppDiv.EnableBrush = true
				AppDiv.SetColumnResizable(0, 8, 20, 8)
				AppDiv.SetColumn(1, 1, Layout_MAX_SIZE)
				AppDiv.SetRow(0, 1, Layout_MAX_SIZE)

				//Chat(or settings)
				//note: must be called before, because it will update chat label

				if app.Selected_chat_i >= 0 {
					err = st.buildApp(AppDiv.AddLayoutWithName(1, 0, 1, 1, fmt.Sprintf("app_%s", app.Name)), activate_prompt, source_root, app, chat_fileName, source_chat, caller)
					//ChatDiv, err := AppDiv.AddTool(1, 0, 1, 1, fmt.Sprintf("chat_%s", app.Name), (&ShowApp{AppName: app.Name, ChatFileName: chat_fileName}).run, caller)
					if err != nil {
						return err
					}

					if source_chat != nil {
						for _, br := range source_chat.Input.Picks {
							if br.Dash_i == source_chat.Selected_user_msg {
								ui.Paint_Brush(br.Cd.Cd, br.Points)
							}
						}
					}
				}

				//Side
				{
					SideDiv := AppDiv.AddLayout(0, 0, 1, 1)
					SideDiv.SetColumn(0, 1, Layout_MAX_SIZE)
					SideDiv.SetRow(1, 1, Layout_MAX_SIZE)

					st.buildAppSideDiv(SideDiv, app, source_root, source_chat, caller)
				}
			}
		}
	}

	//Apps
	{
		AppsDiv := ui.AddLayout(0, 0, 1, 1)
		AppsDiv.SetColumn(0, 1, Layout_MAX_SIZE)
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

		//Normal chats
		{
			//AppsDiv.SetRow(y, d, d)
			chatsBt := AppsDiv.AddButton(0, y, 1, 1, "")
			y++
			chatsBt.IconPath = "resources/chat.png"
			chatsBt.Icon_margin = 0.2
			chatsBt.layout.Tooltip = "Chats"
			chatsBt.Background = 0.25
			if source_root.Show == "chats" {
				chatsBt.Background = 1
			}
			chatsBt.clicked = func() error {
				if source_root.Show != "chats" {
					source_root.Show = "chats"
				} else {
					source_root.Show = ""
				}
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
			AppsDiv.SetRow(y, 1, Layout_MAX_SIZE)
			Apps2Div := AppsDiv.AddLayout(0, y, 1, 1)
			y++
			Apps2Div.SetColumn(0, 1, Layout_MAX_SIZE)
			Apps2Div.ScrollV.Narrow = true
			Apps2Div.SetColumn(0, 1, Layout_MAX_SIZE)
			yy := 0
			for i, app := range source_root.Apps {

				if app.Name == "Device" || app.Name == "Chats" {
					continue
				}

				var bt *UIButton

				if i == source_root.Selected_app_i && source_root.Show == "" {
					dd := 1.0

					Apps2Div.SetRowFromSub(yy, 1, Layout_MAX_SIZE, true)

					BtDiv := Apps2Div.AddLayout(0, yy, 1, 1)
					BtDiv.SetColumn(0, 1, Layout_MAX_SIZE)
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
						source_root.Show = ""
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
				bt.IconPath = filepath.Join("apps", app.Name, "icon")
				bt.Icon_margin = 0.4

				bt.clicked = func() error {
					if source_root.Selected_app_i == i && source_root.Show == "" {
						app.Dev.Enable = false
					}
					source_root.Selected_app_i = i
					source_root.Show = ""
					return nil
				}

				bt.Drag_group = "app"
				bt.Drop_group = "app"
				bt.Drag_index = i
				bt.Drop_v = true
				bt.dropMove = func(src_i, dst_i int, aim_i int, src_source, dst_source string) error {
					Layout_MoveElement(&source_root.Apps, &source_root.Apps, src_i, dst_i)
					source_root.Selected_app_i = dst_i
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
				//ProgressBt.Background = 0.25
				ProgressBt.IconPath = "resources/running.png"
				ProgressBt.Icon_margin = 0.2
				ProgressBt.layout.Tooltip = fmt.Sprintf("Running %d jobs", len(msgs))
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
					source_root.Show = ""
					source_root.Selected_app_i = len(source_root.Apps) - 1
					source_root.Apps[len(source_root.Apps)-1].Dev.Enable = true
				}
				return nil
			}
		}

		//settings error
		var settingsErr string
		{
			type CheckLLMs struct {
				Out_AppProvider_error   string
				Out_CodeProvider_error  string
				Out_ImageProvider_error string
				Out_STTProvider_error   string
			}
			dataJs, _, err := CallToolApp("Device", "CheckLLMs", nil, caller)
			var check CheckLLMs
			if err == nil {
				LogsJsonUnmarshal(dataJs, &check)
			}

			if check.Out_AppProvider_error != "" {
				if settingsErr != "" {
					settingsErr += "\n"
				}
				settingsErr += "App provider error: " + check.Out_AppProvider_error
			}
			if check.Out_CodeProvider_error != "" {
				if settingsErr != "" {
					settingsErr += "\n"
				}
				settingsErr += "Code provider error: " + check.Out_CodeProvider_error
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
			if source_root.Show == "settings" {
				setBt.Background = 1
			}
			if settingsErr != "" {
				//setBt.Background = 1
				setBt.Cd = UI_GetPalette().E
				setBt.layout.Tooltip = settingsErr
			}

			setBt.clicked = func() error {
				if source_root.Show != "settings" {
					source_root.Show = "settings"
				} else {
					source_root.Show = ""
				}
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
			if len(logs) > 0 {
				LogsDia := AppsDiv.AddDialog("logs")
				st.buildLog(&LogsDia.UI, logs, caller)

				hasNewErrs := (len(logs) > 0 && logs[len(logs)-1].Time > source_root.Last_log_time)
				tootip := "Show log"
				if hasNewErrs {
					n := 0
					for _, it := range logs {
						if it.Time > source_root.Last_log_time {
							n++
						}
					}
					tootip = fmt.Sprintf("Log has %d new errors", n)
				}

				LogBt := AppsDiv.AddButton(0, y, 1, 1, "")
				y++
				LogBt.Background = 0.25
				LogBt.IconPath = "resources/warning.png"
				LogBt.Icon_margin = 0.2
				LogBt.layout.Tooltip = tootip
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

				MediaDia.UI.SetColumnFromSub(0, 1, 2, true)
				MediaDia.UI.SetColumnFromSub(1, 1, 10, true)
				MediaDia.UI.SetColumn(3, 2, 5)
				MediaDia.UI.SetColumnFromSub(4, 1, 20, true)
				yy := 0
				for playerID, it := range mediaInfo {

					//progress
					Progress := MediaDia.UI.AddText(0, yy, 1, 1, "xxx")
					Progress.layout.update = func() error {
						mediaInfo := callFuncGetMediaInfo()
						it := mediaInfo[playerID]
						Progress.Label = fmt.Sprintf("%d%%", int(float64(it.Seek)/float64(it.Duration)*100))
						return nil
					}

					//time
					MediaDia.UI.AddText(1, yy, 1, 1, fmt.Sprintf("%s / %s", SdkGetDTime(float64(it.Seek/1000)), SdkGetDTime(float64(it.Duration/1000))))

					//pause/play
					statusLabel := "▶"
					if it.IsPlaying {
						statusLabel = "⏸︎"
					}
					PauseBt := MediaDia.UI.AddButton(2, yy, 1, 1, statusLabel)
					PauseBt.Background = 0.5
					PauseBt.clicked = func() error {
						//send back new value ....
						return nil
					}

					//volume
					vol := it.Volume
					volume := MediaDia.UI.AddSlider(3, yy, 1, 1, &vol, 0, 1, 0.1)
					volume.changed = func() error {
						//send back new value ....
						return nil
					}

					//path
					tx := MediaDia.UI.AddText(4, yy, 1, 1, it.Path)
					tx.layout.Tooltip = fmt.Sprintf("%d", playerID)

					//remove
					StopBt := MediaDia.UI.AddButton(5, yy, 1, 1, "X")
					StopBt.Background = 0.5
					StopBt.clicked = func() error {
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

		//Text highlighting
		{
			SearchBt := AppsDiv.AddButton(0, y, 1, 1, "")
			SearchBt.layout.Tooltip = "Search"
			SearchBt.IconPath = "resources/search.png"
			SearchBt.Icon_margin = 0.2
			SearchBt.Background = 0.25
			if source_root.EnableTextHighlighting {
				SearchBt.Background = 1.0
			}
			SearchBt.Shortcut = 'f'
			y++
			SearchBt.clicked = func() error {
				source_root.EnableTextHighlighting = !source_root.EnableTextHighlighting
				return nil
			}
		}
	}

	return nil
}

func (st *ShowRoot) buildApp(ui *UI, activate_prompt bool, source_root *Root, app *RootApp, chat_fileName string, source_chat *Chat, caller *ToolCaller) error {

	ui.Back_cd = UI_GetPalette().GetGrey(0.03)

	ui.SetColumn(0, 1, Layout_MAX_SIZE)
	ui.SetRow(0, 1, Layout_MAX_SIZE)
	ui.SetRowFromSub(1, 1, g_ShowApp_prompt_height, true)

	isRunning := (callFuncFindMsgName(source_chat.GetChatID()) != nil) //(st.isRunning != nil && st.isRunning())

	dashes := source_chat.GetResponse(source_chat.Selected_user_msg)

	var dashUIs []*ChatMsg
	for _, msg := range dashes {
		if msg.HasUI() {
			dashUIs = append(dashUIs, msg)
		}
	}

	app.Chats[app.Selected_chat_i].Label = "" //reset

	dashW := 1
	if !app.ShowSide {
		dashW = 2
	}

	if len(dashUIs) > 0 {
		if len(dashUIs) == 1 {
			//1x Dash
			appUi, _ := ui.AddToolApp(0, 0, dashW, 1, fmt.Sprintf("chat_%s", source_chat.GetChatID()), app.Name, dashUIs[0].UI_func, []byte(dashUIs[0].UI_paramsJs), caller)
			appUi.changed = func(newParamsJs []byte) error {
				dashUIs[0].UI_paramsJs = string(newParamsJs) //save back changes
				return nil
			}
			app.Chats[app.Selected_chat_i].Label = appUi.findH1()

			appUi.App = true
		} else {
			//Multiple Dashes
			DashDiv := ui.AddLayoutWithName(0, 0, dashW, 1, fmt.Sprintf("chat_%s", source_chat.GetChatID()))
			DashDiv.SetColumn(0, 1, Layout_MAX_SIZE)
			DashDiv.App = true

			for i, dash := range dashUIs {
				DashDiv.SetRowFromSub(i, 1, Layout_MAX_SIZE, true)

				appUi, _ := DashDiv.AddToolApp(0, i, 1, 1, fmt.Sprintf("dash_%s_%d", source_chat.GetChatID(), i), app.Name, dash.UI_func, []byte(dash.UI_paramsJs), caller)
				appUi.changed = func(newParamsJs []byte) error {
					dash.UI_paramsJs = string(newParamsJs) //save back changes
					return nil
				}
				app.Chats[app.Selected_chat_i].Label = appUi.findH1()
			}
		}
	} else {
		//No dash = only message
		for i := len(dashes) - 1; i >= 0; i-- {
			dash := dashes[i]
			if dash.Content.Calls != nil && dash.Content.Calls.Content != "" {
				tx := ui.AddText(0, 0, dashW, 1, dash.Content.Calls.Content)
				tx.Align_h = 1
				break //done
			}
		}
	}

	if app.Chats[app.Selected_chat_i].Label == "" {
		app.Chats[app.Selected_chat_i].Label = source_chat.FindUserMessage(source_chat.Selected_user_msg)
	}

	//Prompt
	{
		DivInput := ui.AddLayoutWithName(0, 1, 1, 1, fmt.Sprintf("prompt_%s", source_chat.GetChatID()))
		d := 0.25
		dd := 0.25
		DivInput.SetColumn(0, d, d) //space
		DivInput.SetColumn(1, 1, Layout_MAX_SIZE)
		DivInput.SetColumn(2, d, d) //space
		DivInput.SetRow(0, d, d)
		DivInput.SetRowFromSub(1, 1, g_ShowApp_prompt_height-0.5, true)
		DivInput.SetRow(2, d, d)

		Div := DivInput.AddLayout(1, 1, 1, 1)

		Div.SetColumn(0, dd, dd) //space
		Div.SetColumn(1, 1, Layout_MAX_SIZE)
		Div.SetColumn(2, dd, dd) //space
		Div.SetRow(0, dd, dd)
		Div.SetRowFromSub(1, 1, g_ShowApp_prompt_height-1, true)
		Div.SetRow(2, dd, dd)

		Div.Back_cd = UI_GetPalette().B //GetGrey(0.05)
		Div.Back_rounding = true
		Div.Border_cd = UI_GetPalette().GetGrey(0.2)

		st.buildPrompt(Div.AddLayoutWithName(1, 1, 1, 1, "prompt"), activate_prompt, source_root, app, source_chat, caller)

		/*pr := ShowPrompt{AppName: st.AppName, ChatFileName: st.ChatFileName}
		_, err = Div.AddTool(1, 1, 1, 1, "prompt", pr.run, caller)
		if err != nil {
			return fmt.Errorf("buildInput() failed: %v", err)
		}*/
	}

	//Side panel
	if app.ShowSide {
		ui.SetColumnResizable(1, 5, 25, 7)
		SideDiv := ui.AddLayout(1, 0, 1, 2)
		SideDiv.SetColumn(0, 1, Layout_MAX_SIZE)
		SideDiv.SetRow(0, 1, Layout_MAX_SIZE)

		//Chat
		ChatDiv, err := SideDiv.AddTool(0, 0, 1, 1, "side", (&ShowChat{AppName: app.Name, ChatFileName: chat_fileName}).run, caller)
		if err != nil {
			return fmt.Errorf("ShowChat.run() failed: %v", err)
		}
		if isRunning {
			if source_chat.scroll_down {
				ChatDiv.VScrollToTheBottom(false, caller)
				source_chat.scroll_down = false //reset
			}
			ChatDiv.VScrollToTheBottom(true, caller)
		}

		//close panel
		CloseBt := SideDiv.AddButton(0, 1, 1, 1, ">>")
		CloseBt.layout.Tooltip = "Close chat panel"
		CloseBt.Background = 0.5
		CloseBt.clicked = func() error {
			app.ShowSide = false //hide
			return nil
		}

	} else {
		ShowSideBt := ui.AddButton(1, 1, 1, 1, "<<")
		ShowSideBt.layout.Tooltip = "Show chat panel"
		ShowSideBt.Background = 0.5
		ShowSideBt.clicked = func() error {
			app.ShowSide = true
			return nil
		}
	}

	return nil
}

func (ui *UI) findH1() string {

	for _, it := range ui.Items {
		if it.Text != nil && strings.HasPrefix(strings.ToLower(it.Text.Label), "<h1>") {

			str := it.Text.Label
			str = strings.ReplaceAll(str, "<h1>", "")
			str = strings.ReplaceAll(str, "</h1>", "")
			str = strings.ReplaceAll(str, "<H1>", "")
			str = strings.ReplaceAll(str, "</H1>", "")
			return str
		}
	}

	return ""
}

func (st *ShowRoot) buildPrompt(ui *UI, activate_prompt bool, source_root *Root, app *RootApp, source_chat *Chat, caller *ToolCaller) {

	isRunning := (callFuncFindMsgName(source_chat.GetChatID()) != nil) //(st.isRunning != nil && st.isRunning())

	input := &source_chat.Input

	preview_height := 0.0
	if len(input.Files) > 0 {
		preview_height = 2
	}

	ui.SetRowFromSub(0, 1, g_ShowApp_prompt_height-1-preview_height, true)

	/*if isRunning {
		MsgsDiv.VScrollToTheBottom(true, caller)
	}*/

	sendIt := func() error {
		source_chat.Messages.Messages = append(source_chat.Messages.Messages, source_chat.TempMessages.Messages...)
		return source_chat._sendIt(app.Name, caller, source_root, false)
	}

	x := 0
	y := 0
	{
		ui.SetColumnFromSub(x, 3, 5, true)
		DivStart := ui.AddLayout(x, y, 1, 1)
		DivStart.SetRow(0, 0, Layout_MAX_SIZE)
		DivStart.Enable = !isRunning
		x++

		xx := 0

		//Drop file
		{
			var filePath string
			fls := DivStart.AddFilePickerButton(xx, 1, 1, 1, &filePath, false, false)
			fls.changed = func() error {
				input.Files = append(input.Files, filePath)
				return nil
			}
			xx++
		}

		//Auto-send after recording
		{
			as := DivStart.AddCheckbox(xx, 1, 1, 1, "", &source_root.Autosend)
			as.layout.Tooltip = "Auto-send after recording"
			xx++
		}

		//Mic
		{
			mic := DivStart.AddMicrophone(xx, 1, 1, 1)
			mic.Transcribe = true
			mic.Transcribe_response_format = "verbose_json"
			mic.Shortcut = '\t'
			mic.Output_onlyTranscript = true

			mic.started = func() error {
				input.Text_mic = input.Text
				return nil
			}
			mic.finished = func(audio []byte, transcript string) error {

				//save
				err := input.SetVoice([]byte(transcript))
				if err != nil {
					return fmt.Errorf("SetVoice() failed: %v", err)
				}

				// auto-send
				if source_root.Autosend > 0 {
					sendIt()
				}

				return nil
			}
			xx++
		}

		//Reset brushes
		if len(input.Picks) > 0 {
			DivStart.SetColumn(xx, 2, 2)
			ClearBt := DivStart.AddButton(xx, 1, 1, 1, "Clear")
			ClearBt.Background = 0.5
			ClearBt.layout.Tooltip = "Remove Brushes"
			ClearBt.clicked = func() error {
				//remove
				//for i := range st.Picks {
				//	st.Text = strings.ReplaceAll(st.Text, st.Picks[i].Cd.GetLabel(), "")
				//}
				input.Text = ""
				input.Picks = nil
				return nil
			}
			xx++
		}
	}

	//Editbox
	{
		ui.SetColumn(x, 1, Layout_MAX_SIZE)
		prompt_editbox := ui.AddEditboxString(x, y, 1, 1, &input.Text)
		prompt_editbox.Ghost = "What can I do for you?"
		prompt_editbox.Multiline = input.Multilined
		prompt_editbox.enter = sendIt
		prompt_editbox.layout.Enable = !isRunning
		prompt_editbox.ActivateOnCreate = true
		if activate_prompt {
			prompt_editbox.Activate(caller)
		}

		x++
	}

	//switch multi-lined
	{
		DivML := ui.AddLayout(x, y, 1, 1)
		DivML.SetColumn(0, 1, Layout_MAX_SIZE)
		DivML.SetRow(0, 0, Layout_MAX_SIZE)
		DivML.Enable = !isRunning

		mt := DivML.AddButton(0, 1, 1, 1, "")
		mt.IconPath = "resources/multiline.png"
		mt.Icon_margin = 0.1
		mt.layout.Tooltip = "Enable/disable multi-line prompt"
		if !input.Multilined {
			mt.Background = 0
		}
		mt.clicked = func() error {
			input.Multilined = !input.Multilined
			return nil
		}
		x++

	}

	//Send button
	{
		ui.SetColumn(x, 2.5, 2.5)
		DivSend := ui.AddLayout(x, y, 1, 1)
		DivSend.SetColumn(0, 1, Layout_MAX_SIZE)
		DivSend.SetRow(0, 0, Layout_MAX_SIZE)
		if !isRunning {
			SendBt := DivSend.AddButton(0, 1, 1, 1, "Send")
			SendBt.IconPath = "resources/up.png"
			SendBt.Icon_margin = 0.2
			SendBt.Align = 0
			//SendBt.Tooltip = //name of "text" model ....
			SendBt.clicked = sendIt
		} else {
			StopBt := DivSend.AddButton(0, 1, 1, 1, "Stop")
			StopBt.Cd = UI_GetPalette().E
			StopBt.clicked = func() error {
				callFuncMsgStop(source_chat.GetChatID()) //stop
				return nil
			}
		}
		x++
	}
	y++

	//show file previews
	if len(input.Files) > 0 {
		ui.SetRow(y, preview_height, preview_height)
		ImgsCards := ui.AddLayoutCards(0, y, x, 1, true)
		y++

		for fi, file := range input.Files {
			ImgDia := ui.AddDialog("image_" + file)
			ImgDia.UI.SetColumn(0, 5, 12)
			ImgDia.UI.SetColumn(1, 3, 3)
			ImgDia.UI.SetRow(1, 5, 15)
			ImgDia.UI.AddMediaPath(0, 1, 2, 1, file)
			ImgDia.UI.AddText(0, 0, 1, 1, file)
			RemoveBt := ImgDia.UI.AddButton(1, 0, 1, 1, "Remove")
			RemoveBt.clicked = func() error {
				input.Files = slices.Delete(input.Files, fi, fi+1)
				ImgDia.Close(caller)
				return nil
			}

			imgLay := ImgsCards.AddItem()
			imgLay.SetColumn(0, 2, 2)
			imgLay.SetRow(0, 2, 2)
			imgBt := imgLay.AddButton(0, 0, 1, 1, "")
			imgBt.IconPath = file
			imgBt.Icon_margin = 0
			imgBt.layout.Tooltip = file

			imgBt.Background = 0
			imgBt.Cd = UI_GetPalette().B
			imgBt.Border = true
			imgBt.clicked = func() error {
				ImgDia.OpenRelative(imgBt.layout, caller)
				return nil
			}
		}

		//remove all files
		{
			delLay := ImgsCards.AddItem()
			delLay.SetColumn(0, 2, 2)
			delLay.SetRow(0, 2, 2)
			delBt := delLay.AddButton(0, 0, 1, 1, "Delete All")
			delBt.Background = 0.5
			delBt.clicked = func() error {
				input.Files = nil
				return nil
			}
		}
	}

	//LLMTips/Brushes
	if len(input.Picks) > 0 {
		ui.SetRowFromSub(y, 1, 8, true)
		TipsDiv := ui.AddLayout(0, y, x, 1)
		y++
		TipsDiv.SetColumn(0, 2, 2)
		TipsDiv.SetColumn(1, 1, Layout_MAX_SIZE)

		yy := 0
		for i, br := range input.Picks {
			found_i := source_chat.Input.FindPick(br.LLMTip)
			if found_i >= 0 && found_i < i { //unique
				continue //skip
			}

			TipsDiv.SetRowFromSub(yy, 1, 100, true)
			TipsDiv.AddText(0, yy, 1, 1, input.Picks[i].Cd.GetLabel())
			TipsDiv.AddText(1, yy, 1, 1, strings.TrimSpace(br.LLMTip)).setMultilined()
			yy++

			if i+1 < len(input.Picks) {
				TipsDiv.SetRow(yy, 0.1, 0.1)
				TipsDiv.AddDivider(0, yy, 2, 1, true)
				yy++
			}

		}
	}
}

func (st *ShowRoot) buildAppSideDiv(SideDiv *UI, app *RootApp, source_root *Root, source_chat *Chat, caller *ToolCaller) {

	//Header
	{
		SideDiv.SetRow(0, 1, 1.5)
		HeaderDiv := SideDiv.AddLayout(0, 0, 1, 1)
		HeaderDiv.SetRow(0, 1, Layout_MAX_SIZE)
		HeaderDiv.ScrollH.Narrow = true
		HeaderDiv.ScrollV.Hide = true
		//New Tab
		{
			HeaderDiv.SetColumn(0, 2, Layout_MAX_SIZE)
			bt := HeaderDiv.AddButton(0, 0, 1, 1, "New Tab")
			bt.Background = 0.5
			bt.layout.Tooltip = "Create new tab"
			bt.Shortcut = 't'
			bt.layout.Enable = (app != nil)
			bt.clicked = func() error {
				if app == nil {
					return fmt.Errorf("No app selected")
				}

				chat_fileName := fmt.Sprintf("Chat-%d.json", time.Now().UnixMicro())
				source_chat, err := NewChat(filepath.Join("..", app.Name, "Chats", chat_fileName))
				if err != nil {
					return nil
				}

				pos := app.NumPins() //skip pins
				app.Chats = slices.Insert(app.Chats, pos, RootChat{Label: "Empty", FileName: chat_fileName})
				app.Selected_chat_i = pos

				SideDiv.VScrollToTheTop(caller)

				source_root.Show = ""

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

			HeaderDiv.SetColumnFromSub(2, 1, Layout_MAX_SIZE, true)
			NavDiv := HeaderDiv.AddLayout(2, 0, 1, 1)
			NavDiv.SetRow(0, 1, Layout_MAX_SIZE)

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
			backBt.Shortcut = '←'
			backBt.Background = 0.5
			backBt.layout.Enable = (source_chat.Selected_user_msg > 0)
			backBt.clicked = func() error {
				source_chat.Selected_user_msg--
				return nil
			}

			NavDiv.SetColumnFromSub(nx, 1.5, Layout_MAX_SIZE, true)
			inf := NavDiv.AddText(nx, 0, 1, 1, fmt.Sprintf("%d/%d", source_chat.Selected_user_msg+1, numUseMessages)) //...
			nx++
			inf.Align_h = 1
			//inf.layout.Tooltip = dashes.UI_func + "()"

			NavDiv.SetColumn(nx, 1.5, 1.5)
			forwardBt := NavDiv.AddButton(nx, 0, 1, 1, ">")
			nx++
			forwardBt.layout.Tooltip = "Next dashboard"
			forwardBt.Shortcut = '→'
			forwardBt.Background = 0.5
			forwardBt.layout.Enable = (source_chat.Selected_user_msg+1 < numUseMessages)
			forwardBt.clicked = func() error {
				source_chat.Selected_user_msg++
				return nil
			}
		}
	}

	//List of tabs
	ListsDiv := SideDiv.AddLayout(0, 1, 1, 1)
	ListsDiv.SetColumn(0, 1, Layout_MAX_SIZE)

	var PinnedDiv *UI
	var TabsDiv *UI

	num_pins := app.NumPins()
	if num_pins > 0 {
		ListsDiv.SetRow(0, 0.7, 0.7)
		ListsDiv.AddDivider(0, 0, 1, 1, true).Label = "<small>Pins"
		ListsDiv.SetRowFromSub(1, 1, Layout_MAX_SIZE, true)
		PinnedDiv = ListsDiv.AddLayout(0, 1, 1, 1)
		PinnedDiv.SetColumn(1, 1, Layout_MAX_SIZE)

		ListsDiv.SetRow(2, 0.7, 0.7)
		ListsDiv.AddDivider(0, 2, 1, 1, true).Label = "<small>Tabs"
		ListsDiv.SetRowFromSub(3, 1, Layout_MAX_SIZE, true)
		TabsDiv = ListsDiv.AddLayout(0, 3, 1, 1)
		TabsDiv.SetColumn(1, 1, Layout_MAX_SIZE)

	} else {
		ListsDiv.SetRow(0, 1, Layout_MAX_SIZE)
		TabsDiv = ListsDiv.AddLayout(0, 0, 1, 1)
		TabsDiv.SetColumn(1, 1, Layout_MAX_SIZE)
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
			source_root.Show = ""

			return nil
		}

		btChat.Drag_group = "chat"
		btChat.Drop_group = "chat"
		btChat.Drag_index = i
		btChat.Drop_v = true
		btChat.dropMove = func(src_i, dst_i int, aim_i int, src_source, dst_source string) error {

			app.Chats[src_i].Pinned = (aim_i < num_pins)

			Layout_MoveElement(&app.Chats, &app.Chats, src_i, dst_i)

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
	ui.SetColumn(0, 1, Layout_MAX_SIZE)
	ui.SetColumn(1, 10, 16)
	ui.SetColumn(2, 1, Layout_MAX_SIZE)

	y := 0
	ui.AddTextLabel(1, y, 1, 1, "Settings").Align_h = 1
	y++
	y++

	//device settings
	{
		ui.SetRowFromSub(y, 0, Layout_MAX_SIZE, true)
		ui.AddToolApp(1, y, 1, 1, "device_settings", "Device", "ShowDeviceSettings", nil, caller)
		y++
	}

	ui.AddDivider(1, y, 1, 1, true)
	y++

	// LLMs
	{
		ui.SetRowFromSub(y, 0, Layout_MAX_SIZE, true)
		ui.AddToolApp(1, y, 1, 1, "llm_settings", "Device", "ShowLLMsSettings", nil, caller)
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

	//label
	ui.AddTextLabel(0, 0, 1, 1, "Spendings")

	//spendings
	total_price := 0.0
	{
		ui.SetRow(1, 1, 15) //ui.SetRowFromSub(1, 1, 15, true)
		ListDiv := ui.AddLayout(0, 1, 1, 1)
		ListDiv.SetColumnFromSub(0, 1, 10, true)
		ListDiv.SetColumnFromSub(1, 1, 10, true)
		ListDiv.SetColumnFromSub(2, 1, 10, true)
		ListDiv.SetColumnFromSub(3, 1, 10, true)

		//usages = usages[:2]
		y := 0
		for i := len(usages) - 1; i >= 0; i-- {
			usg := &usages[i]

			ListDiv.AddText(0, y, 1, 1, usg.Model)
			ListDiv.AddText(1, y, 1, 1, SdkGetDateTime(int64(usg.CreatedTimeSec)))
			ListDiv.AddText(2, y, 1, 1, SdkGetDTime(usg.DTime))

			price := (usg.Prompt_price + usg.Input_cached_price + usg.Completion_price + usg.Reasoning_price)
			ListDiv.AddText(3, y, 1, 1, fmt.Sprintf("$%f", price))
			total_price += price

			y++
		}
	}

	//space
	ui.SetRow(2, 0.1, 0.1)
	ui.AddDivider(0, 2, 1, 1, true)

	//Sum
	ui.AddText(0, 3, 1, 1, fmt.Sprintf("Total(%d): $%f", len(usages), total_price)).Align_h = 2

	//Note
	noteTx := ui.AddText(0, 4, 1, 1, "<i>numbers may not be accurate.")
	noteTx.Align_h = 1
	noteTx.Cd = UI_GetPalette().GetGrey(0.5)
}

func (st *ShowRoot) buildLog(ui *UI, logs []SdkLog, caller *ToolCaller) {
	ui.SetColumnFromSub(0, 1, 30, true)

	{
		HeaderDiv := ui.AddLayout(0, 0, 1, 1)
		HeaderDiv.SetColumn(0, 5, Layout_MAX_SIZE)
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
	ListDiv.SetColumnFromSub(0, 1, 10, true)
	ListDiv.SetColumnFromSub(1, 1, Layout_MAX_SIZE, true)

	MAX_N := 20
	y := 0
	for i := len(logs) - 1; i >= 0 && i >= len(logs)-MAX_N; i-- {
		it := logs[i]

		ListDiv.SetRowFromSub(y, 1, 5, true)

		ListDiv.AddText(0, y, 1, 1, SdkGetDateTime(int64(it.Time)))

		tx := ListDiv.AddText(1, y, 1, 1, strings.TrimSpace(it.Msg))
		tx.layout.Tooltip = it.Stack
		tx.Cd = UI_GetPalette().E
		tx.setMultilined()

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
