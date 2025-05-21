package main

import (
	"fmt"
	"os"
	"path/filepath"
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

	//load chat
	var sourceChat *Chat
	chatID := int64(0)
	if source_root.Selected_tab_i >= 0 {
		if source_root.Selected_tab_i >= len(source_root.Tabs) {
			source_root.Selected_tab_i = len(source_root.Tabs) - 1
		}
		if source_root.Selected_tab_i >= 0 {
			chatID = source_root.Tabs[source_root.Selected_tab_i].ChatID
			sourceChat, err = NewChat(fmt.Sprintf("Chats/Chat-%d.json", chatID), caller)
			if err != nil {
				return err
			}

			if sourceChat != nil {
				//reload
				source_root.Tabs[source_root.Selected_tab_i].Label = sourceChat.Label
				source_root.Tabs[source_root.Selected_tab_i].Use_sources = sourceChat.GetListOfSources()
			}
		}
	}

	//save brush
	if sourceChat != nil {
		if st.AddBrush != nil {
			sourceChat.Input.MergePick(*st.AddBrush)
			ui.ActivateEditbox("chat_user_prompt", caller)
		}
	}

	ui.SetColumnResizable(0, 5, 15, 6)
	ui.SetColumn(1, 1, 100)
	ui.SetRow(0, 1, 100)

	//Side
	{
		SideDiv := ui.AddLayout(0, 0, 1, 1)
		SideDiv.SetColumn(0, 1, 100)

		HeaderDiv := SideDiv.AddLayout(0, 0, 1, 1)

		SideDiv.SetRow(1, 1, 100)
		TabsDiv := SideDiv.AddLayout(0, 1, 1, 1)

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
				bt.clicked = func() error {

					//empty tab is only in array, create file when there is actually some content ....

					chatID = time.Now().UnixMicro()
					sourceChat, err = NewChat(fmt.Sprintf("Chats/Chat-%d.json", chatID), caller)
					if err != nil {
						return nil
					}

					source_root.Tabs = slices.Insert(source_root.Tabs, 0, RootTab{Label: "Empty chat", ChatID: chatID})
					source_root.Selected_tab_i = 0
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
				if source_root.Show == "settings" {
					logoBt.Background = 1
				}
				logoBt.clicked = func() error {
					if source_root.Show == "settings" {
						source_root.Show = ""
					} else {
						source_root.Show = "settings"
					}
					return nil
				}
			}

		}

		//List of tabs
		{
			TabsDiv.SetColumn(0, 1, 100)
			yy := 0
			for i, tab := range source_root.Tabs {

				btChat := TabsDiv.AddButton(0, yy, 1, 1, tab.Label)
				yy++

				icon_path := "tools/icons/unknown.png" //chat icon
				if len(tab.Use_sources) > 0 {
					last_struct := tab.Use_sources[len(tab.Use_sources)-1]

					//avoid "DeviceSettings"
					if last_struct == "DeviceSettings" && len(tab.Use_sources) > 1 {
						last_struct = tab.Use_sources[len(tab.Use_sources)-2]
					}

					if _, err := os.Stat(filepath.Join("..", "tools/icons", last_struct+".png")); err == nil {
						icon_path = filepath.Join("tools/icons", last_struct+".png")
					}
					btChat.Tooltip = fmt.Sprintf("Files: %s", tab.Use_sources)
				}
				btChat.IconPath = icon_path
				btChat.Icon_margin = 0.25

				btChat.Align = 0
				btChat.Background = 0.2
				if i == source_root.Selected_tab_i {
					if source_root.Show != "" {
						btChat.Border = true
					} else {
						btChat.Background = 1 //selected
					}
				}
				btChat.clicked = func() error {
					source_root.Selected_tab_i = i
					source_root.Show = "" //reset
					ui.ActivateEditbox("chat_user_prompt", caller)
					return nil
				}

				btChat.Drag_group = "chat"
				btChat.Drop_group = "chat"
				btChat.Drag_index = i
				btChat.Drop_v = true
				btChat.dropMove = func(src_i, dst_i int, src_source, dst_source string) error {
					Layout_MoveElement(&source_root.Tabs, &source_root.Tabs, src_i, dst_i)
					source_root.Selected_tab_i = dst_i
					ui.ActivateEditbox("chat_user_prompt", caller)
					return nil
				}

				btClose := TabsDiv.AddButton(1, i, 1, 1, "X")
				btClose.Background = 0.2
				btClose.clicked = func() error {
					source_root.Tabs = slices.Delete(source_root.Tabs, i, i+1)

					if i < source_root.Selected_tab_i {
						source_root.Selected_tab_i--
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
				st.buildThreads(SideDiv.AddLayout(0, 2, 1, 1), msgs)
			}
		}
	}

	//Chat(or settings)
	{
		if source_root.Show == "settings" {
			err := st.buildSettings(ui.AddLayout(1, 0, 1, 1), caller, source_root)
			if err != nil {
				return err
			}
		} else if chatID != 0 {
			ChatDiv, err := ui.AddTool(1, 0, 1, 1, (&ShowChat{ChatID: chatID}).run, caller)
			if err != nil {
				return err
			}

			ChatDiv.Back_cd = caller.GetPalette().GetGrey(0.03)

			for _, br := range sourceChat.Input.Picks {
				ui.Paint_Brush(br.Cd.Cd, br.Points)
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
		ui.AddTool(1, y, 1, 1, (&ShowDeviceSettings{}).run, caller)
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
				bt.Cd = caller.GetPalette().E
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
				bt.Cd = caller.GetPalette().E
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
		tx.Cd = caller.GetPalette().E
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
