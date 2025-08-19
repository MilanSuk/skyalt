package main

import (
	"fmt"
	"path/filepath"
	"slices"
	"time"
)

// [ignore]
type ShowChats struct {
}

func (st *ShowChats) run(caller *ToolCaller, ui *UI) error {
	source_root, err := NewChats("")
	if err != nil {
		return err
	}

	//load chat
	source_chat, chat_fileName, err := source_root.refreshChats()

	ui.SetColumnResizable(0, 4, 15, 7)
	ui.SetColumn(1, 1, Layout_MAX_SIZE)
	ui.SetRow(0, 1, Layout_MAX_SIZE)

	var prompt_editbox *UIEditbox
	{
		MainDiv := ui.AddLayout(1, 0, 1, 1)
		MainDiv.SetColumn(0, 1, Layout_MAX_SIZE)
		MainDiv.SetRow(0, 1, Layout_MAX_SIZE)
		MainDiv.SetRowFromSub(1, 1, g_ShowApp_prompt_height, true)

		//messages
		_, err = MainDiv.AddTool(0, 0, 1, 1, fmt.Sprintf("chat_%s", chat_fileName), (&ShowChat{ChatFileName: chat_fileName}).run, caller)
		if err != nil {
			return err
		}

		//prompt
		{
			DivInput := MainDiv.AddLayout(0, 1, 1, 1)
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

			prompt_editbox = st.buildPrompt(Div.AddLayoutWithName(1, 1, 1, 1, "prompt"), source_root, source_chat, caller)
			if err != nil {
				return fmt.Errorf("buildInput() failed: %v", err)
			}
		}
	}

	//side
	{
		SideDiv := ui.AddLayout(0, 0, 1, 1)
		SideDiv.SetColumn(0, 1, Layout_MAX_SIZE)
		SideDiv.SetRow(1, 1, Layout_MAX_SIZE)
		st.buildSideDiv(SideDiv, prompt_editbox, source_root, source_chat, caller)
	}

	return nil
}

func (st *ShowChats) buildSideDiv(SideDiv *UI, prompt_editbox *UIEditbox, root *Root, source_chat *Chat, caller *ToolCaller) {

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
			bt := HeaderDiv.AddButton(0, 0, 1, 1, "New Chat")
			bt.Background = 0.5
			bt.layout.Tooltip = "Create new chat"
			bt.Shortcut = 't'
			bt.clicked = func() error {
				//create new
				fileName := fmt.Sprintf("Chat-%d.json", time.Now().UnixMicro())
				_, err := NewChat(filepath.Join("Chats", fileName))
				if err != nil {
					return nil
				}

				//add
				pos := root.NumPins() //skip pins
				root.Chats = slices.Insert(root.Chats, pos, RootChat{Label: fmt.Sprintf("Chat %s", SdkGetDateTime(time.Now().Unix())), FileName: fileName})
				root.Selected_chat_i = pos

				prompt_editbox.Activate(caller)

				return nil
			}
		}
	}

	//List of tabs
	ListsDiv := SideDiv.AddLayout(0, 1, 1, 1)
	ListsDiv.SetColumn(0, 1, Layout_MAX_SIZE)

	var PinnedDiv *UI
	var TabsDiv *UI

	num_pins := root.NumPins()
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
	for i, tab := range root.Chats {

		isSelected := (i == root.Selected_chat_i)

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
				root.Chats = slices.Insert(root.Chats, num_pins-1, tab)
				root.Chats = slices.Delete(root.Chats, i, i+1)
			}
			root.Chats[i].Pinned = !tab.Pinned
			return nil
		}

		btChat.layout.Tooltip = tab.Label
		btChat.Align = 0
		btChat.Background = 0.2
		if isSelected {
			btChat.Background = 1 //selected
		}
		btChat.clicked = func() error {
			root.Selected_chat_i = i
			prompt_editbox.Activate(caller)
			return nil
		}

		btChat.Drag_group = "chat"
		btChat.Drop_group = "chat"
		btChat.Drag_index = i
		btChat.Drop_v = true
		btChat.dropMove = func(src_i, dst_i int, aim_i int, src_source, dst_source string) error {

			root.Chats[src_i].Pinned = (aim_i < num_pins)

			Layout_MoveElement(&root.Chats, &root.Chats, src_i, dst_i)

			if root.Selected_chat_i != dst_i {
				prompt_editbox.Activate(caller)
			}
			root.Selected_chat_i = dst_i

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
				return root.RemoveChat(root.Chats[i])
			}
		}

		if tab.Pinned {
			yPinned++
		} else {
			yTabs++
		}
	}
}

func (st *ShowChats) buildPrompt(ui *UI, source_root *Root, source_chat *Chat, caller *ToolCaller) *UIEditbox {

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
		return source_chat._sendIt(caller, source_root, false)
	}

	x := 0
	y := 0
	{
		ui.SetColumnFromSub(x, 2, 5, true)
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
	}

	//Editbox
	var prompt_editbox *UIEditbox
	{
		ui.SetColumn(x, 1, Layout_MAX_SIZE)
		prompt_editbox = ui.AddEditboxString(x, y, 1, 1, &input.Text)
		prompt_editbox.Ghost = "What can I do for you?"
		prompt_editbox.Multiline = input.Multilined
		prompt_editbox.enter = sendIt
		prompt_editbox.layout.Enable = !isRunning
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
	/*if len(input.Picks) > 0 {
		ui.SetRowFromSub(y, 1, 5, true)
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

			TipsDiv.SetRowFromSub(yy, 1, 3, true)
			TipsDiv.AddText(0, yy, 1, 1, input.Picks[i].Cd.GetLabel())
			TipsDiv.AddText(1, yy, 1, 1, strings.TrimSpace(br.LLMTip)).setMultilined()
			yy++
		}
	}*/

	return prompt_editbox
}
