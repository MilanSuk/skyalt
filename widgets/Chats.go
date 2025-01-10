package main

import (
	"fmt"
	"math/rand"
)

type Chats struct {
	SearchChat string

	Chats         []Chat
	Selected_chat int
}

func (layout *Layout) AddChats(x, y, w, h int, props *Chats) *Chats {
	layout._createDiv(x, y, w, h, "Chats", props.Build, nil, nil)
	return props
}

var g_Chats *Chats

func OpenFile_Chats() *Chats {
	if g_Chats == nil {
		g_Chats = &Chats{}
		_read_file("Chats-Chats", g_Chats)
	}
	return g_Chats
}

func (st *Chats) Build(layout *Layout) {

	st.checkUIDs()
	//st.saveJobs()

	layout.SetColumnResizable(0, 2, 10, 5)
	layout.SetColumn(1, 1, 100)
	layout.SetRow(0, 1, 100)

	if st.Selected_chat >= len(st.Chats) {
		st.Selected_chat = len(st.Chats) - 1
	}

	//chat
	var selChat *Chat
	if st.Selected_chat >= 0 && st.Selected_chat < len(st.Chats) {
		selChat = &st.Chats[st.Selected_chat]
	}

	//list of chats
	ListDiv := layout.AddLayout(0, 0, 1, 1)
	{
		ListDiv.SetColumn(0, 1, 100)
		ListDiv.SetRow(1, 1, 100)

		//head
		{
			ListDiv.AddSearch(0, 0, 1, 1, &st.SearchChat, "")
		}

		ListChats := ListDiv.AddLayout(0, 1, 1, 1)
		{
			ListChats.SetColumn(0, 1, 100)

			y := 0
			searchWords := Search_Prepare(st.SearchChat)
			for i, ch := range st.Chats {
				chat_i := i
				if !Search_Find(ch.Name, searchWords) {
					continue //skip
				}

				ListItem := ListChats.AddLayout(0, y, 1, 1)
				ListItem.SetColumn(0, 1, 100)
				ListItem.Drag_group = "chat"
				ListItem.Drop_group = "chat"
				ListItem.Drag_index = i
				ListItem.Drop_v = true
				ListItem.dropMove = func(src int, dst int) {
					Layout_MoveElement(&st.Chats, &st.Chats, src, dst)
					if st.Selected_chat == src {
						st.Selected_chat = dst
					}
				}

				bt := ListItem.AddButtonMenu(0, 0, 1, 1, ch.Name, "", 0)
				if st.Selected_chat == i {
					bt.Background = 1
				}
				bt.clicked = func() {
					st.Selected_chat = chat_i
				}

				y++
			}

			AddNew := ListChats.AddButton(0, y, 1, 1, "+")
			AddNew.Background = 0.5
			AddNew.Tooltip = "Create new Chat"
			AddNew.clicked = func() {
				st.addNewChat()
			}
		}
	}

	//chat
	if selChat != nil {
		layout.AddChat(1, 0, 1, 1, selChat)
	} else {
		EmptyChat := layout.AddLayout(1, 0, 1, 1)
		EmptyChat.SetColumn(0, 1, 100)
		EmptyChat.SetColumn(1, 4, 4)
		EmptyChat.SetColumn(2, 1, 100)
		EmptyChat.SetRow(0, 1, 100)
		EmptyChat.SetRow(2, 1, 100)

		AddNew := EmptyChat.AddButton(1, 1, 1, 1, "Create first Chat")
		AddNew.clicked = func() {
			st.addNewChat()
		}
	}
}

const g__chat_instructions_default = "You are an AI assistant. Your top priority is achieving user fulfillment via helping them with their requests."

func (st *Chats) checkUIDs() {
	for i := range st.Chats {
		if st.Chats[i].UID == "" {
			st.Chats[i].UID = fmt.Sprintf("Chat:%d", rand.Int())
		}
	}
}

func (st *Chats) addNewChat() {
	item := Chat{
		UID:          "",
		Name:         "New Chat",
		Instructions: g__chat_instructions_default,
	}
	item.Properties.Reset()

	st.Chats = append(st.Chats, item)

	st.Selected_chat = len(st.Chats) - 1

	st.checkUIDs()
}
