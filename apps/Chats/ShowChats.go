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
	root, err := NewChats("")
	if err != nil {
		return err
	}

	//load chat
	source_chat, chat_fileName, err := root.refreshChats()

	ui.SetColumnResizable(0, 4, 15, 7)
	ui.SetColumn(1, 1, Layout_MAX_SIZE)
	ui.SetRow(0, 1, Layout_MAX_SIZE)

	//side
	{
		SideDiv := ui.AddLayout(0, 0, 1, 1)
		SideDiv.SetColumn(0, 1, Layout_MAX_SIZE)
		SideDiv.SetRow(1, 1, Layout_MAX_SIZE)
		st.buildSideDiv(SideDiv, root, source_chat, caller)
	}

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

			pr := ShowPrompt{ChatFileName: chat_fileName}
			_, err = Div.AddTool(1, 1, 1, 1, "prompt", pr.run, caller)
			if err != nil {
				return fmt.Errorf("buildInput() failed: %v", err)
			}
		}
	}

	return nil
}

func (st *ShowChats) buildSideDiv(SideDiv *UI, root *Root, source_chat *Chat, caller *ToolCaller) {

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
				root.Chats = slices.Insert(root.Chats, pos, RootChat{Label: "Empty", FileName: fileName})
				root.Selected_chat_i = pos

				SideDiv.ActivateEditbox("chat_user_prompt", caller)

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
			SideDiv.ActivateEditbox("chat_user_prompt", caller)
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
				SideDiv.ActivateEditbox("chat_user_prompt", caller)
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
