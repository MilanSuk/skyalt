package main

import (
	"time"
)

type Chat struct {
	UID  string
	Name string

	Instructions string
	Msgs         []ChatMsg

	TempMsg string

	Properties Llamacpp_completion_props
}

func (layout *Layout) AddChat(x, y, w, h int, props *Chat) *Chat {
	layout._createDiv(x, y, w, h, "Chat:"+props.UID, props.Build, nil, nil)
	return props
}

func (st *Chat) Build(layout *Layout) {
	chat := NewGlobal_Llamacpp_completion(st.UID)
	job := chat.FindJob()

	layout.SetColumn(0, 1, 100)
	layout.SetColumnResizable(1, 3, 15, 5)
	layout.SetRow(0, 0, 100)

	ChatDiv := layout.AddLayout(0, 0, 1, 1)
	{
		ChatDiv.SetColumn(0, 1, 100)
		ChatDiv.SetRow(0, 0, 100)
		ChatDiv.SetRowFromSub(1, 1, 100)

		MsgsDiv := ChatDiv.AddLayout(0, 0, 1, 1)
		{
			MsgsDiv.SetColumn(0, 1, 100)

			y := 0
			for i := range st.Msgs {
				//previous message
				MsgsDiv.SetRowFromSub(y, 1, 100)
				MsgsDiv.AddChatMsg(0, y, 1, 1, &st.Msgs[i])
				y++

				//space
				MsgsDiv.SetRow(y, 0.5, 0.5)
				y++
			}

			if job != nil {
				//user msg
				MsgsDiv.SetRowFromSub(y, 1, 100)
				MsgsDiv.AddChatMsg(0, y, 1, 1, &ChatMsg{Text: st.TempMsg, CreatedTimeSec: job.start_time_sec, CreatedBy: ""})
				y++

				MsgsDiv.SetRow(y, 0.5, 0.5) //space
				y++

				//generated msg
				MsgsDiv.SetRowFromSub(y, 1, 100)
				MsgsDiv.AddLlamacpp_completion(0, y, 1, 1, chat)
				y++

				MsgsDiv.SetRow(y, 0.5, 0.5) //space
				y++
			}
		}

		InputDiv := ChatDiv.AddLayout(0, 1, 1, 1)
		{
			InputDiv.SetColumn(0, 1, 100)
			InputDiv.SetColumn(1, 3, 3)
			InputDiv.SetRowFromSub(0, 1, 5)

			NewMsg := InputDiv.AddEditboxMultiline(0, 0, 2, 1, &st.TempMsg)
			NewMsg.enter = func() {
				if st.TempMsg == "" {
					return
				}

				chat := NewGlobal_Llamacpp_completion(st.UID)
				job := chat.FindJob()
				if job != nil {
					return
				}

				stTime := float64(time.Now().UnixMilli()) / 1000
				done := func(out string) {
					if out != "" {
						st.Msgs = append(st.Msgs, ChatMsg{Text: st.TempMsg, CreatedTimeSec: stTime, CreatedBy: ""})
						st.Msgs = append(st.Msgs, ChatMsg{Text: out, CreatedTimeSec: float64(time.Now().UnixMilli()) / 1000, CreatedBy: chat.Properties.Model})

						MsgsDiv.VScrollToTheBottom()
						st.TempMsg = ""
					}
				}

				chat.Properties = st.Properties
				chat.Properties.Messages = st.GetMessages(true)
				chat.done = done
				job = chat.Start()

				MsgsDiv.VScrollToTheBottom()
			}

			{
				ImgsList := InputDiv.AddLayoutList(0, 1, 1, 1, true)
				//already added small preview ...
				//+dialog preview + remove ...

				addImgLay := ImgsList.AddListSubItem()
				AddImgBt := addImgLay.AddButton(0, 0, 1, 1, "+")
				AddImgBt.clicked = func() {
					//relative dialog FilePicker ...
				}
			}

			SendBt, SendLay := InputDiv.AddButton2(1, 1, 1, 1, "Send")
			SendLay.Enable = (job == nil)
			SendBt.clicked = NewMsg.enter
		}
	}

	PropertiesDiv := layout.AddLayout(1, 0, 1, 1)
	{
		PropertiesDiv.SetColumn(0, 1, 100)
		PropertiesDiv.SetColumn(1, 1, 3.5)
		PropertiesDiv.SetRowResizable(1, 1, 10, 3) //Instructions editbox
		PropertiesDiv.SetRowFromSub(2, 1, 100)

		PropertiesDiv.AddText(0, 0, 1, 1, "Instructions")
		bt := PropertiesDiv.AddButton(1, 0, 1, 1, "Reset")
		bt.Background = 0.5
		bt.clicked = func() {
			st.Instructions = g__chat_instructions_default
		}
		PropertiesDiv.AddEditboxMultiline(0, 1, 2, 1, &st.Instructions)

		PropertiesDiv.AddLlamacpp_completion_props(0, 2, 2, 1, &st.Properties)
	}
}

func (chat *Chat) GetMessages(addTempMsg bool) []OpenAI_completion_msg {
	var Messages []OpenAI_completion_msg
	Messages = append(Messages, OpenAI_completion_msg{Role: "system", Content: chat.Instructions})

	for _, msg := range chat.Msgs {
		role := "user"
		if msg.CreatedBy != "" {
			role = "assistant"
		}
		Messages = append(Messages, OpenAI_completion_msg{Role: role, Content: msg.Text})
	}
	if addTempMsg {
		Messages = append(Messages, OpenAI_completion_msg{Role: "user", Content: chat.TempMsg})
	}

	return Messages
}
