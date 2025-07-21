package main

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

// [ignore]
type ShowPrompt struct {
	AppName      string
	ChatFileName string
}

func (st *ShowPrompt) run(caller *ToolCaller, ui *UI) error {

	source_chat, err := NewChat(filepath.Join("..", st.AppName, "Chats", st.ChatFileName))
	if err != nil {
		return err
	}
	source_root, err := NewRoot("")
	if err != nil {
		return err
	}

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
		return source_chat._sendIt(st.AppName, caller, source_root, false)
	}

	x := 0
	y := 0
	{
		ui.SetColumnFromSub(x, 3, 5, true)
		DivStart := ui.AddLayout(x, y, 1, 1)
		DivStart.SetRow(0, 0, 100)
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
		ui.SetColumn(x, 1, 100)
		ed := ui.AddEditboxString(x, y, 1, 1, &input.Text)
		ed.Ghost = "What can I do for you?"
		ed.Multiline = input.Multilined
		ed.enter = sendIt
		ed.Name = "chat_user_prompt"
		ed.layout.Enable = !isRunning
		x++
	}

	//switch multi-lined
	{
		DivML := ui.AddLayout(x, y, 1, 1)
		DivML.SetColumn(0, 1, 100)
		DivML.SetRow(0, 0, 100)
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
		DivSend.SetColumn(0, 1, 100)
		DivSend.SetRow(0, 0, 100)
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
		ui.SetRowFromSub(y, 1, 5, true)
		TipsDiv := ui.AddLayout(0, y, x, 1)
		y++
		TipsDiv.SetColumn(0, 2, 2)
		TipsDiv.SetColumn(1, 1, 100)

		yy := 0
		for i, br := range input.Picks {
			found_i := source_chat.Input.FindPick(br.LLMTip)
			if found_i >= 0 && found_i < i { //unique
				continue //skip
			}

			TipsDiv.SetRowFromSub(yy, 1, 3, true)
			TipsDiv.AddText(0, yy, 1, 1, input.Picks[i].Cd.GetLabel())
			TipsDiv.AddText(1, yy, 1, 1, strings.TrimSpace(br.LLMTip))
			yy++
		}
	}

	return nil
}
