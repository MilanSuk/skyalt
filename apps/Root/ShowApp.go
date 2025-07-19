package main

import (
	"fmt"
	"path/filepath"
)

// [ignore]
type ShowApp struct {
	AppName      string
	ChatFileName string
}

const g_ShowApp_prompt_height = 7

func (st *ShowApp) run(caller *ToolCaller, ui *UI) error {

	source_chat, err := NewChat(filepath.Join("..", st.AppName, "Chats", st.ChatFileName))
	if err != nil {
		return err
	}
	/*source_root, err := NewRoot("")
	if err != nil {
		return err
	}*/

	var dash *ChatMsg
	if source_chat.Dash_call_id != "" {
		dash = source_chat.FindUI(source_chat.Dash_call_id)
	}

	ui.SetColumn(0, 1, 100)
	ui.SetRow(0, 1, 100)
	ui.SetRowFromSub(1, 1, g_ShowApp_prompt_height, true)

	isRunning := (callFuncFindMsgName(source_chat.GetChatID()) != nil) //(st.isRunning != nil && st.isRunning())

	var MsgsDiv *UI
	if dash != nil {

		DashDiv := ui.AddLayout(0, 0, 1, 1)
		DashDiv.SetColumn(0, 1, 100)
		DashDiv.SetColumnResizable(1, 1, 50, 10)
		DashDiv.SetRow(0, 1, 100)

		//Dash
		{
			appUi, err := DashDiv.AddToolApp(0, 0, 1, 1, st.AppName, dash.UI_func, []byte(dash.UI_paramsJs), caller)
			if err != nil {
				return fmt.Errorf("AddToolApp() failed: %v", err)
			}
			appUi.changed = func(newParamsJs []byte) error {
				dash.UI_paramsJs = string(newParamsJs) //save back changes
				return nil
			}

		}

		ChatDiv := DashDiv.AddLayout(1, 0, 1, 1)
		ChatDiv.SetColumn(0, 1, 100)
		ChatDiv.SetRow(1, 1, 100)

		DashHeaderDiv := ChatDiv.AddLayout(0, 0, 1, 1)
		DashHeaderDiv.ScrollH.Narrow = true
		//DashHeaderDiv.ScrollV.Hide = true
		{
			preUI := source_chat.FindPreviousUI(source_chat.Dash_call_id)
			nxtUI := source_chat.FindNextUI(source_chat.Dash_call_id)

			btClose := DashHeaderDiv.AddButton(0, 0, 1, 1, "<<")
			btClose.layout.Tooltip = "Hide dashboard"
			btClose.Background = 0.5
			btClose.clicked = func() error {
				source_chat.Dash_call_id = "" //reset
				return nil
			}
			btBack := DashHeaderDiv.AddButton(2, 0, 1, 1, "<")
			btBack.layout.Tooltip = "Previous dashboard"
			btBack.Background = 0.5
			btBack.layout.Enable = (preUI != nil)
			btBack.clicked = func() error {
				if preUI != nil {
					source_chat.Dash_call_id = preUI.Content.Result.Tool_call_id
				}
				return nil
			}
			btForward := DashHeaderDiv.AddButton(3, 0, 1, 1, ">")
			btForward.layout.Tooltip = "Next dashboard"
			btForward.Background = 0.5
			btForward.layout.Enable = (nxtUI != nil)
			btForward.clicked = func() error {
				if nxtUI != nil {
					source_chat.Dash_call_id = nxtUI.Content.Result.Tool_call_id
				}
				return nil
			}

			//FuncName
			DashHeaderDiv.SetColumn(5, 3, 100)
			DashHeaderDiv.AddText(5, 0, 1, 1, dash.UI_func+"()").Multiline = false
		}

		//Chat
		MsgsDiv = ChatDiv.AddLayout(0, 1, 1, 1)
		ch := ShowChat{AppName: st.AppName, ChatFileName: st.ChatFileName}
		err = ch.run(caller, MsgsDiv)
		//err = st.buildShowMessages(MsgsDiv, caller, source_chat, source_root, isRunning)
		if err != nil {
			return fmt.Errorf("buildShowMessages1() failed: %v", err)
		}
		if isRunning {
			MsgsDiv.VScrollToTheBottom(true, caller)
		}

	} else {
		//Chat
		MsgsDiv = ui.AddLayout(0, 0, 1, 1)
		ch := ShowChat{AppName: st.AppName, ChatFileName: st.ChatFileName}
		err = ch.run(caller, MsgsDiv)
		//err = st.buildShowMessages(MsgsDiv, caller, source_chat, source_root, isRunning)
		if err != nil {
			return fmt.Errorf("buildShowMessages2() failed: %v", err)
		}
		if isRunning {
			MsgsDiv.VScrollToTheBottom(true, caller)
		}
	}

	//User prompt
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
		err = pr.run(caller, Div.AddLayout(1, 1, 1, 1))
		//err = st.buildInput(Div.AddLayout(1, 1, 1, 1), caller, source_chat, source_root, MsgsDiv, isRunning)
		if err != nil {
			return fmt.Errorf("buildInput() failed: %v", err)
		}
	}

	return nil
}
