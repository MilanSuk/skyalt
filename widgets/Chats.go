package main

import (
	"fmt"
	"sort"
	"strconv"
	"time"
)

type Chats struct {
	Folder     string
	SearchChat string

	Selected_chat_id int64
}

func (layout *Layout) AddChats(x, y, w, h int, props *Chats) *Chats {
	layout._createDiv(x, y, w, h, "Chats", props.Build, nil, nil)
	return props
}

func (st *Chats) Build(layout *Layout) {
	if st.Folder == "" {
		st.Folder = "chats"
	}

	layout.SetColumnResizable(0, 2, 10, 5)
	layout.SetColumn(1, 1, 100)
	layout.SetRow(0, 1, 100)

	found_selChat := false

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

			files, _ := GetListOfFiles(st.Folder)

			//sort
			sort.Slice(files, func(i, j int) bool {
				ival, _ := strconv.Atoi(files[i].Name)
				jval, _ := strconv.Atoi(files[j].Name)
				return ival < jval
			})

			for _, file := range files {
				//for i := range st.Chats {
				fid, _ := strconv.Atoi(file.Name)

				if int64(fid) == st.Selected_chat_id {
					found_selChat = true
				}

				chat := OpenFilePath_Chat(fmt.Sprintf("%s/Chat-%s", st.Folder, file.Name))
				chat.file_name = file.Name //update!

				if !Search_Find(chat.Label, searchWords) {
					continue //skip
				}

				ListItem := ListChats.AddLayout(0, y, 1, 1)
				ListItem.SetColumn(0, 1, 100)
				/*ListItem.Drag_group = "chat"
				ListItem.Drop_group = "chat"
				ListItem.Drag_index = i
				ListItem.Drop_v = true
				ListItem.dropMove = func(src int, dst int) {
					if st.Selected_chat == st.Chats[src] {
						st.Selected_chat = st.Chats[dst]
					}
					Layout_MoveElement(&st.Chats, &st.Chats, src, dst)
				}*/

				bt := ListItem.AddButtonMenu(0, 0, 1, 1, chat.Label, "", 0)
				if st.Selected_chat_id == int64(fid) {
					bt.Background = 1
				}
				bt.clicked = func() {
					st.Selected_chat_id = int64(fid)
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
	if found_selChat {
		path := fmt.Sprintf("%s/Chat-%d", st.Folder, st.Selected_chat_id)
		layout.AddApp(1, 0, 1, 1, path)
		layout.RenameLayout(1, 0, 1, 1, path) //change name(from Chat), so it remembers scroll, resize info

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

func (st *Chats) addNewChat() {
	id := time.Now().UnixMilli()

	item := OpenFilePath_Chat(fmt.Sprintf("%s/Chat-%d", st.Folder, id))
	item.Reset()

	st.Selected_chat_id = id
}
