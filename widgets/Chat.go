package main

import (
	"time"
)

type Chat struct {
	file_name string

	Label string

	Instructions string
	Msgs         []ChatMsg

	Input ChatInput

	Service              string
	Llamacpp_Properties  Llamacpp_completion_props
	OpenAI_Properties    OpenAI_completion_props
	Anthropic_Properties Anthropic_completion_props
	Xai_Properties       Xai_completion_props
	Groq_Properties      Groq_completion_props
}

func (layout *Layout) AddChat(x, y, w, h int, props *Chat) *Chat {
	layout._createDiv(x, y, w, h, "Chat", props.Build, nil, nil)
	return props
}

func (st *Chat) Build(layout *Layout) {
	var job *Job
	switch st.Service {
	case "llamacpp":
		job = OpenMemory_Llamacpp_completion(st.file_name).FindJob()
	case "openai":
		job = OpenMemory_OpenAI_completion(st.file_name).FindJob()
	case "anthropic":
		job = OpenMemory_Anthropic_completion(st.file_name).FindJob()
	case "xai":
		job = OpenMemory_Xai_completion(st.file_name).FindJob()
	case "groq":
		job = OpenMemory_Groq_completion(st.file_name).FindJob()
	}

	layout.SetColumn(0, 1, 100)
	layout.SetColumnResizable(1, 3, 15, 10)
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
				MsgsDiv.AddChatMsg(0, y, 1, 1, &ChatMsg{Text: st.Input.Text, CreatedTimeSec: job.start_time_sec, CreatedBy: ""})
				y++

				MsgsDiv.SetRow(y, 0.5, 0.5) //space
				y++

				//generated msg
				MsgsDiv.SetRowFromSub(y, 1, 100)

				switch st.Service {
				case "llamacpp":
					MsgsDiv.AddLlamacpp_completion(0, y, 1, 1, OpenMemory_Llamacpp_completion(st.file_name))
				case "openai":
					MsgsDiv.AddOpenAI_completion(0, y, 1, 1, OpenMemory_OpenAI_completion(st.file_name))
				case "anthropic":
					MsgsDiv.AddAnthropic_completion(0, y, 1, 1, OpenMemory_Anthropic_completion(st.file_name))
				case "xai":
					MsgsDiv.AddXai_completion(0, y, 1, 1, OpenMemory_Xai_completion(st.file_name))
				case "groq":
					MsgsDiv.AddGroq_completion(0, y, 1, 1, OpenMemory_Groq_completion(st.file_name))
				}
				y++

				MsgsDiv.SetRow(y, 0.5, 0.5) //space
				y++
			}
		}

		InputDiv := ChatDiv.AddChatInput(0, 1, 1, 1, &st.Input)
		ChatDiv.FindLayout(0, 1, 1, 1).Enable = (job == nil)
		InputDiv.sended = func() {
			stTime := float64(time.Now().UnixMilli()) / 1000
			var model string

			switch st.Service {
			case "llamacpp":
				chat := OpenMemory_Llamacpp_completion(st.file_name)
				model = chat.Properties.Model
				job := chat.FindJob()
				if job != nil {
					return
				}
				chat.Properties = st.Llamacpp_Properties
				chat.Properties.Messages = st.buildOpenAIMsgs()
				chat.done = func(out string) {
					st.addMsg(out, stTime, model, MsgsDiv)
				}
				job = chat.Start()
			case "openai":
				chat := OpenMemory_OpenAI_completion(st.file_name)
				model = chat.Properties.Model
				job := chat.FindJob()
				if job != nil {
					return
				}
				chat.Properties = st.OpenAI_Properties
				chat.Properties.Messages = st.buildOpenAIMsgs()
				chat.done = func(out string) {
					st.addMsg(out, stTime, model, MsgsDiv)
				}
				job = chat.Start()

			case "anthropic":
				chat := OpenMemory_Anthropic_completion(st.file_name)
				model = chat.Properties.Model
				job := chat.FindJob()
				if job != nil {
					return
				}
				chat.Properties = st.Anthropic_Properties
				chat.Properties.Messages = st.buildOpenAIMsgs()
				chat.done = func(out string) {
					st.addMsg(out, stTime, model, MsgsDiv)
				}
				job = chat.Start()
			case "xai":
				chat := OpenMemory_Xai_completion(st.file_name)
				model = chat.Properties.Model
				job := chat.FindJob()
				if job != nil {
					return
				}
				chat.Properties = st.Xai_Properties
				chat.Properties.Messages = st.buildOpenAIMsgs()
				chat.done = func(out string) {
					st.addMsg(out, stTime, model, MsgsDiv)
				}
				job = chat.Start()
			case "groq":
				chat := OpenMemory_Groq_completion(st.file_name)
				model = chat.Properties.Model
				job := chat.FindJob()
				if job != nil {
					return
				}
				chat.Properties = st.Groq_Properties
				chat.Properties.Messages = st.buildOpenAIMsgs()
				chat.done = func(out string) {
					st.addMsg(out, stTime, model, MsgsDiv)
				}
				job = chat.Start()
			}

			MsgsDiv.VScrollToTheBottom()
		}
	}

	PropertiesDiv := layout.AddLayout(1, 0, 1, 1)
	{
		PropertiesDiv.SetColumn(0, 1, 100)
		PropertiesDiv.SetColumn(1, 1, 3)
		PropertiesDiv.SetRowResizable(1, 1, 10, 3) //Instructions editbox
		PropertiesDiv.SetRowFromSub(3, 1, 100)

		PropertiesDiv.AddText(0, 0, 1, 1, "Instructions")
		bt := PropertiesDiv.AddButton(1, 0, 1, 1, "Reset")
		bt.Background = 0.5
		bt.clicked = func() {
			st.Instructions = g__chat_instructions_default
		}
		PropertiesDiv.AddEditboxMultiline(0, 1, 2, 1, &st.Instructions)

		labels := []string{"Llama.cpp", "OpenAI", "Anthropic", "xAI", "Groq"}
		values := []string{"llamacpp", "openai", "anthropic", "xai", "groq"}
		PropertiesDiv.AddCombo(0, 2, 2, 1, &st.Service, labels, values)

		switch st.Service {
		case "llamacpp":
			PropertiesDiv.AddLlamacpp_completion_props(0, 3, 2, 1, &st.Llamacpp_Properties)
		case "openai":
			PropertiesDiv.AddOpenAI_completion_props(0, 3, 2, 1, &st.OpenAI_Properties)
		case "anthropic":
			PropertiesDiv.AddAnthropic_completion_props(0, 3, 2, 1, &st.Anthropic_Properties)
		case "xai":
			PropertiesDiv.AddXai_completion_props(0, 3, 2, 1, &st.Xai_Properties)
		case "groq":
			PropertiesDiv.AddGroq_completion_props(0, 3, 2, 1, &st.Groq_Properties)
		}
	}
}

const g__chat_instructions_default = "You are an AI assistant. Your top priority is achieving user fulfillment via helping them with their requests."

func (st *Chat) Reset() {
	st.Label = "New Chat"
	st.Service = "llamacpp"
	st.Instructions = g__chat_instructions_default
	st.Llamacpp_Properties.Reset()
	st.OpenAI_Properties.Reset()
	//st.OpenAI_PropertiesV.Reset()
	st.Anthropic_Properties.Reset()
	st.Xai_Properties.Reset()
	//st.Xai_PropertiesV.Reset()
	st.Groq_Properties.Reset()
}
func (st *Chat) addMsg(out string, createdTimeSec float64, createdBy string, scrollLayout *Layout) {
	if out != "" {
		st.Msgs = append(st.Msgs, ChatMsg{Text: st.Input.Text, Files: st.Input.Files, CreatedTimeSec: createdTimeSec, CreatedBy: ""})
		st.Msgs = append(st.Msgs, ChatMsg{Text: out, CreatedTimeSec: float64(time.Now().UnixMilli()) / 1000, CreatedBy: createdBy})

		scrollLayout.VScrollToTheBottom()
		st.Input.reset()
	}
}

func (st *Chat) buildOpenAIMsgs() []OpenAI_completion_msg {
	var Messages []OpenAI_completion_msg

	sysMsg := OpenAI_completion_msg{Role: "system"}
	sysMsg.AddText(st.Instructions)
	Messages = append(Messages, sysMsg)

	for _, msg := range st.Msgs {
		var userMsg OpenAI_completion_msg

		//role
		userMsg.Role = "user"
		if msg.CreatedBy != "" {
			userMsg.Role = "assistant"
		}
		//text
		userMsg.AddText(msg.Text)

		//image(s)
		for _, file := range msg.Files {
			err := userMsg.AddImageFile(file)
			if err != nil {
				Layout_WriteError(err)
			}
		}

		Messages = append(Messages, userMsg)
	}

	//latest msg
	{
		userMsg := OpenAI_completion_msg{Role: "user"}

		//text
		userMsg.AddText(st.Input.Text)

		//image(s)
		for _, file := range st.Input.Files {
			err := userMsg.AddImageFile(file)
			if err != nil {
				Layout_WriteError(err)
			}
		}

		Messages = append(Messages, userMsg)
	}

	return Messages
}
