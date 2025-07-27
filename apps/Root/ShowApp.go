package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// [ignore]
type ShowApp struct {
	AppName      string
	ChatFileName string
}

const g_ShowApp_prompt_height = 7

func (st *ShowApp) run(caller *ToolCaller, ui *UI) error {
	source_root, err := NewRoot("")
	if err != nil {
		return err
	}
	//refresh apps
	app, err := source_root.refreshApps()
	if err != nil {
		return err
	}

	if app == nil {
		return nil //err ....
	}

	if app.Selected_chat_i < 0 {
		return nil //err ...
	}

	source_chat, err := NewChat(filepath.Join("..", st.AppName, "Chats", st.ChatFileName))
	if err != nil {
		return err
	}

	ui.SetColumn(0, 1, 100)
	ui.SetRow(0, 1, 100)
	ui.SetRowFromSub(1, 1, g_ShowApp_prompt_height, true)

	isRunning := (callFuncFindMsgName(source_chat.GetChatID()) != nil) //(st.isRunning != nil && st.isRunning())

	dashes := source_chat.GetResponse(source_chat.User_msg_i)

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
			appUi, err := ui.AddToolApp(0, 0, dashW, 1, st.AppName, dashUIs[0].UI_func, []byte(dashUIs[0].UI_paramsJs), caller)
			if err != nil {
				return fmt.Errorf("AddToolApp() failed: %v", err)
			}
			appUi.changed = func(newParamsJs []byte) error {
				dashUIs[0].UI_paramsJs = string(newParamsJs) //save back changes
				return nil
			}
			app.Chats[app.Selected_chat_i].Label = appUi.findH1()

		} else {
			//Multiple Dashes

			DashDiv := ui.AddLayout(0, 0, dashW, 1)
			DashDiv.SetColumn(0, 1, 100)

			for i, dash := range dashUIs {
				DashDiv.SetRowFromSub(i, 1, 100, true)

				appUi, err := DashDiv.AddToolApp(0, i, 1, 1, st.AppName, dash.UI_func, []byte(dash.UI_paramsJs), caller)
				if err != nil {
					return fmt.Errorf("AddToolApp() failed: %v", err)
				}
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
		app.Chats[app.Selected_chat_i].Label = source_chat.FindUserMessage(source_chat.User_msg_i)
	}

	//Prompt
	{
		DivInput := ui.AddLayout(0, 1, 1, 1)
		d := 0.25
		dd := 0.25
		DivInput.SetColumn(0, d, d) //space
		DivInput.SetColumn(1, 1, 100)
		DivInput.SetColumn(2, d, d) //space
		DivInput.SetRow(0, d, d)
		DivInput.SetRowFromSub(1, 1, g_ShowApp_prompt_height-0.5, true)
		DivInput.SetRow(2, d, d)

		Div := DivInput.AddLayout(1, 1, 1, 1)

		Div.SetColumn(0, dd, dd) //space
		Div.SetColumn(1, 1, 100)
		Div.SetColumn(2, dd, dd) //space
		Div.SetRow(0, dd, dd)
		Div.SetRowFromSub(1, 1, g_ShowApp_prompt_height-1, true)
		Div.SetRow(2, dd, dd)

		Div.Back_cd = UI_GetPalette().B //GetGrey(0.05)
		Div.Back_rounding = true
		Div.Border_cd = UI_GetPalette().GetGrey(0.2)

		pr := ShowPrompt{AppName: st.AppName, ChatFileName: st.ChatFileName}
		_, err = Div.AddTool(1, 1, 1, 1, pr.run, caller)
		if err != nil {
			return fmt.Errorf("buildInput() failed: %v", err)
		}
	}

	//Side panel
	if app.ShowSide {
		ui.SetColumnResizable(1, 5, 25, 7)
		SideDiv := ui.AddLayout(1, 0, 1, 2)
		SideDiv.SetColumn(0, 1, 100)
		SideDiv.SetRow(0, 1, 100)

		//Chat
		ChatDiv, err := SideDiv.AddTool(0, 0, 1, 1, (&ShowChat{AppName: st.AppName, ChatFileName: st.ChatFileName}).run, caller)
		if err != nil {
			return fmt.Errorf("ShowChat.run() failed: %v", err)
		}
		if isRunning {
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
