/*
Copyright 2024 Milan Suk

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this db except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"path/filepath"
	"slices"
	"time"
)

type Root struct {
	Chats      []string
	TrashChats []string

	Search      string
	TrashSearch string

	Selected int
	Show     string

	AgentSelected string
}

func (st *Root) Build(layout *Layout) {

	st.refreshChats()

	layout.SetColumnResizable(0, 5, 10, 7)
	layout.SetColumn(1, 1, 100)
	layout.SetRow(0, 1, 100)

	{
		sideDiv := layout.AddLayout(0, 0, 1, 1)
		sideDiv.SetColumn(0, 1, 100)

		y := 0
		{
			headDiv := sideDiv.AddLayout(0, y, 1, 1)
			y++
			headDiv.SetColumn(0, 2, 2)
			headDiv.SetColumn(1, 1, 100)

			logo := headDiv.AddImageCd(0, 0, 1, 1, "resources/logo_small.png", color.RGBA{0, 0, 0, 255})
			logo.Align_h = 0
			logo.Margin = 0.2
			logo.Tooltip = "v0.1" //....

			createChatBt := headDiv.AddButton(1, 0, 1, 1, "New chat")
			createChatBt.Background = 0.5
			//createChatBt.Align = 0
			createChatBt.clicked = func() {
				st.AddNewChat()
			}
		}

		sideDiv.AddSearch(0, y, 1, 1, &st.Search, "")
		y++

		{
			sideDiv.SetRow(y, 1, 100)
			chatsDiv := sideDiv.AddLayout(0, y, 1, 1)
			y++
			chatsDiv.SetColumn(0, 1, 100)

			words := Search_Prepare(st.Search)
			yy := 00
			for i, it := range st.Chats {

				chat := OpenFile_Chat(it)

				if !Search_Find(chat.Description, words) {
					continue
				}

				createChatBt, chatLay := chatsDiv.AddButtonMenu2(0, yy, 1, 1, chat.Description, "", 0)
				createChatBt.Tooltip = it
				chatLay.Drag_group = "chat"
				chatLay.Drop_group = "chat"
				chatLay.Drag_index = i
				chatLay.Drag_source = "live"
				chatLay.Drop_v = true
				chatLay.dropMove = func(src_i, dst_i int, src_source, dst_source string) {
					st.MoveChat(src_i, dst_i, src_source, dst_source)
				}
				if i == st.Selected {
					if st.Show != "" {
						createChatBt.Border = true
						createChatBt.Background = 0.25 //highlight light
					} else {
						createChatBt.Background = 1 //highlight full
					}
				}
				createChatBt.clicked = func() {
					st.SelectChat(i)
				}

				contextDia := chatsDiv.AddDialog(fmt.Sprintf("chat_%d", i))
				{
					contextDia.Layout.SetColumn(0, 3, 5)
					y := 0
					dupBt := contextDia.Layout.AddButtonMenu(0, y, 1, 1, "Duplicate", "resources/duplicate.png", 0.2)
					dupBt.Background = 0.25
					dupBt.clicked = func() {
						cp := st.CopyChat(it)
						st.Chats = slices.Insert(st.Chats, i+1, cp)

						st.SelectChat(i + 1)
						contextDia.Close()
					}
					y++

					contextDia.Layout.AddDivider(0, y, 1, 1, true)
					contextDia.Layout.SetRow(y, 0.1, 0.1)
					y++

					delBt := contextDia.Layout.AddButtonMenu(0, y, 1, 1, "Delete", "resources/delete.png", 0.2)
					y++
					delBt.clicked = func() {

						st.TrashChats = append(st.TrashChats, st.Chats[i]) //add
						st.Chats = slices.Delete(st.Chats, i, i+1)         //remove

						if i < st.Selected {
							st.SelectChat(st.Selected - 1)
						}

						contextDia.Close()
					}
				}

				contextBt, contextLay := chatsDiv.AddButtonIcon2(1, yy, 1, 1, "resources/context.png", 0.2, "")
				contextBt.Background = 0.25
				contextBt.clicked = func() {
					contextDia.OpenRelative(contextLay)
				}

				yy++
			}
		}

		sideDiv.SetRow(y, 0.1, 0.1)
		sideDiv.AddDivider(0, y, 1, 1, true)
		y++

		{
			bottomDiv := sideDiv.AddLayout(0, y, 1, 1)

			st.addIconButton(0, "disk", "resources/folders.png", "Disk", bottomDiv)
			st.addIconButton(1, "tools", "resources/tools.png", "Tools", bottomDiv)
			st.addIconButton(2, "passwords", "resources/key.png", "Passwords", bottomDiv)
			st.addIconButton(3, "settings", "resources/settings.png", "Settings", bottomDiv)
			st.addIconButton(4, "trash", "resources/delete.png", "Recover deleted chats", bottomDiv)
		}
	}

	if st.Show == "disk" {
		//....

	} else if st.Show == "tools" {
		//....

	} else if st.Show == "passwords" {
		//....

	} else if st.Show == "settings" {
		//settings
		setDiv := layout.AddLayout(1, 0, 1, 1)
		setDiv.SetColumn(0, 1, 100)
		setDiv.SetColumn(1, 10, 16)
		setDiv.SetColumn(2, 1, 100)

		//device settings
		setDiv.SetRowFromSub(0, 0, 100)
		setDiv.AddSettings(1, 0, 1, 1, OpenFile_DeviceSettings())

		setDiv.AddDivider(1, 1, 1, 1, true)

		//agents
		{
			agentPathes, _ := OpenDir_agents_properties() //err ....
			var agents []string
			for _, ag := range agentPathes {
				agents = append(agents, filepath.Base(ag))
			}
			if st.AgentSelected == "" && len(agentPathes) > 0 {
				st.AgentSelected = agentPathes[0]
			}
			setDiv.AddTabs(1, 2, 1, 1, &st.AgentSelected, agents, agentPathes)
			setDiv.SetRowFromSub(3, 0, 100)
			setDiv.AddAgents_properties(1, 3, 1, 1, OpenFile_Agent_properties(st.AgentSelected))
		}

		setDiv.AddDivider(1, 4, 1, 1, true)

		//about
		setDiv.SetRowFromSub(5, 0, 100)
		AboutDiv := setDiv.AddLayout(1, 5, 1, 1)
		{
			AboutDiv.SetColumn(0, 1, 100)
			AboutDiv.SetRow(0, 2, 4)

			AboutDiv.AddImageCd(0, 0, 1, 1, "resources/logo.png", color.RGBA{0, 0, 0, 255})

			Version := AboutDiv.AddText(0, 1, 1, 1, "v0.1")
			Version.Align_h = 1

			Url := AboutDiv.AddButton(0, 2, 1, 1, "github.com/milansuk/skyalt/")
			Url.Background = 0
			Url.BrowserUrl = "https://github.com/milansuk/skyalt/"

			License := AboutDiv.AddText(0, 3, 1, 1, "This program is distributed under the terms of Apache License, Version 2.0.")
			License.Align_h = 1

			Copyright := AboutDiv.AddText(0, 4, 1, 1, "This program comes with absolutely no warranty.")
			Copyright.Align_h = 1

			Warranty := AboutDiv.AddText(0, 5, 1, 1, "Copyright © 2025 - SkyAlt team")
			Warranty.Align_h = 1
		}
	} else if st.Show == "trash" {
		//trash
		TrashDiv := layout.AddLayout(1, 0, 1, 1)
		TrashDiv.SetColumn(0, 1, 100)
		TrashDiv.SetColumn(1, 1, 15)
		TrashDiv.SetColumn(2, 1, 100)
		TrashDiv.SetRow(4, 1, 100)
		TrashDiv.SetRow(7, 1, 1)

		title := TrashDiv.AddText(1, 0, 1, 1, "Deleted chats")
		title.Align_h = 1

		TrashDiv.AddSearch(1, 2, 1, 1, &st.TrashSearch, "")
		words := Search_Prepare(st.TrashSearch)

		ListDiv := TrashDiv.AddLayout(1, 4, 1, 1)
		ListDiv.SetColumn(0, 3, 100)
		ListDiv.SetColumn(1, 3, 4)
		ListDiv.SetColumn(2, 3, 4)
		yy := 0
		for i, it := range st.TrashChats {

			chat := OpenFile_Chat(it)

			if !Search_Find(chat.Description, words) {
				continue
			}

			OpenBt, OpenBtLay := ListDiv.AddButtonMenu2(0, yy, 1, 1, chat.Description, "", 0)
			OpenBt.Background = 0.25
			OpenBt.clicked = func() {
				//open Chat as dialog ...
			}
			OpenBtLay.Drag_group = "chat"
			OpenBtLay.Drop_group = "chat"
			OpenBtLay.Drag_index = i
			OpenBtLay.Drag_source = "trash"
			OpenBtLay.Drop_v = true
			OpenBtLay.dropMove = func(src_i, dst_i int, src_source, dst_source string) {
				st.MoveChat(src_i, dst_i, src_source, dst_source)
			}

			RecoverBt := ListDiv.AddButton(1, yy, 1, 1, "Recover")
			RecoverBt.Background = 0.5
			RecoverBt.clicked = func() {
				st.Chats = append(st.Chats, st.TrashChats[i])        //add
				st.TrashChats = slices.Delete(st.TrashChats, i, i+1) //delete
			}

			DeleteBt := ListDiv.AddButtonConfirm(2, yy, 1, 1, "Delete", "Are you sure, you wanna delete the chat permanently?")
			DeleteBt.confirmed = func() {
				RemoveFile_Chat(it)
				st.TrashChats = slices.Delete(st.TrashChats, i, i+1)
			}

			yy++
		}

		EmptyBt := TrashDiv.AddButtonConfirm(1, 6, 1, 1, "Empty trash", "Are you sure, you wanna delete all chats permanently?")
		EmptyBt.Enable = len(st.TrashChats) > 0
		EmptyBt.confirmed = func() {
			for _, it := range st.TrashChats {
				RemoveFile_Chat(it)
			}
			st.TrashChats = nil
		}

	} else if len(st.Chats) == 0 {
		//create first chat
		EmptyDiv := layout.AddLayout(1, 0, 1, 1)
		EmptyDiv.SetColumn(0, 1, 100)
		EmptyDiv.SetColumn(1, 4, 4)
		EmptyDiv.SetColumn(2, 1, 100)
		EmptyDiv.SetRow(0, 1, 100)
		EmptyDiv.SetRow(2, 1, 100)

		AddNew := EmptyDiv.AddButton(1, 1, 1, 1, "Create first Chat")
		AddNew.clicked = func() {
			st.AddNewChat()
		}

	} else {
		//chat
		file := st.Chats[st.Selected]
		agent := OpenFile_Chat(file)

		ChatDiv := layout.AddLayout(1, 0, 1, 1)
		ChatDiv.SetRow(0, 1, 100)

		if agent != nil && agent.Selected_sub_call_id == "" {
			//1x centered chat
			ChatDiv.SetColumn(0, 1, 100)
			ChatDiv.AddChat(0, 0, 1, 1, &Chat{file_name: file, agent: agent, parent_agent: nil, center: true})
		} else {
			//multiple chats
			ChatDiv.SetColumn(0, 0, 0)
			x := 1 //as centered(x=1), so it remembers scroll position
			var parent_agent *Agent
			for agent != nil {
				ChatDiv.SetColumnResizable(x, 10, 100, 20)
				ChatDiv.AddChat(x, 0, 1, 1, &Chat{file_name: file, agent: agent, parent_agent: parent_agent, center: false})
				x++

				if agent.Selected_sub_call_id != "" {
					parent_agent = agent
					agent = agent.FindSubAgent(agent.Selected_sub_call_id)
				} else {
					agent = nil
				}
			}
			ChatDiv.SetColumn(x, 1, 1) //extra space so resizer can be grabbed
		}
	}
}

func (st *Root) addIconButton(x int, show string, icon string, label string, layout *Layout) {
	layout.SetColumn(x, 0.5, 100)

	trashBt := layout.AddButtonIcon(x, 0, 1, 1, icon, 0.15, label)
	trashBt.Background = 0.5
	if st.Show == show {
		trashBt.Background = 1
	}
	x++
	trashBt.clicked = func() {
		if st.Show == show {
			st.Show = ""
		} else {
			st.Show = show
		}
	}
}

func (st *Root) MoveChat(src_i, dst_i int, src_source, dst_source string) {
	src := &st.Chats
	dst := &st.Chats

	if src_source == "trash" {
		src = &st.TrashChats
	}
	if dst_source == "trash" {
		dst = &st.TrashChats
	}

	Layout_MoveElement(src, dst, src_i, dst_i)

	if dst_source != "trash" {
		st.Selected = dst_i
	}
}

func (st *Root) SelectChat(i int) {
	st.Selected = i
	st.Show = ""
}

func (st *Root) AddNewChat() {
	path := fmt.Sprintf("chats/%d", time.Now().UnixMilli())
	agent := OpenFile_Chat(path)
	*agent = *NewAgent("", "", "")
	agent.Description = "Chat"

	st.Chats = append(st.Chats, path)
	st.SelectChat(len(st.Chats) - 1)
}

func (st *Root) CopyChat(src_path string) string {

	//create
	dst_path := fmt.Sprintf("chats/%d", time.Now().UnixMilli())
	dst_agent := OpenFile_Chat(dst_path)

	//copy
	src_agent := OpenFile_Chat(src_path)
	js, _ := json.Marshal(src_agent)
	json.Unmarshal(js, dst_agent)

	return dst_path
}

func Root_findInList(str string, items []string) bool {
	for _, it := range items {
		if it == str {
			return true
		}
	}
	return false
}

func (st *Root) refreshChats() {
	files, _ := OpenDir_Chats() //err ...

	//add new files
	for _, path := range files {
		found := Root_findInList(path, st.Chats)
		if !found {
			found = Root_findInList(path, st.TrashChats)
		}

		if !found {
			st.Chats = append(st.Chats, path)
		}
	}

	//remove if file doesn't exist
	for i := len(st.Chats) - 1; i >= 0; i-- {
		if !Root_findInList(st.Chats[i], files) {
			st.Chats = slices.Delete(st.Chats, i, i+1)
		}
	}
	for i := len(st.TrashChats) - 1; i >= 0; i-- {
		if !Root_findInList(st.TrashChats[i], files) {
			st.TrashChats = slices.Delete(st.TrashChats, i, i+1)
		}
	}

	//check boundary
	if st.Selected >= len(st.Chats) {
		st.Selected = len(st.Chats) - 1
	}
}
