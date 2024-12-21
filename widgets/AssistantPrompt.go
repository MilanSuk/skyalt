package main

type AssistantPrompt struct {
	Properties OpenAI_chat_props
	//chat *OpenAI_chat

	SystemPrompt string
	UserPrompt   string

	ChatInput  string
	ChatOutput string

	Error error
}

func (layout *Layout) AddAssistantPrompt(x, y, w, h int, props *AssistantPrompt) *AssistantPrompt {
	layout._createDiv(x, y, w, h, "AssistantPrompt", props.Build, nil, nil)
	return props
}

func (st *AssistantPrompt) Build(layout *Layout) {

	layout.SetColumn(0, 1, 5)
	layout.SetColumn(1, 1, 20)

	y := 0
	space := 0.1

	layout.AddText(0, y, 1, 1, "Error")
	if st.Error != nil {
		tx := layout.AddText(1, y, 1, 1, st.Error.Error())
		tx.Cd = Paint_GetPalette().E
	} else {
		layout.AddText(1, y, 1, 1, "-")
	}
	y++

	layout.SetRow(y, space, space)
	layout.AddDivider(0, y, 2, 1, true)
	y++

	layout.SetRow(y, 2, 2)
	layout.AddText(0, y, 1, 1, "System Prompt")
	layout.AddTextMultiline(1, y, 1, 1, st.SystemPrompt)
	y++

	layout.SetRow(y, space, space)
	layout.AddDivider(0, y, 2, 1, true)
	y++

	layout.SetRow(y, 2, 5)
	layout.AddText(0, y, 1, 1, "User Prompt")
	layout.AddTextMultiline(1, y, 1, 1, st.UserPrompt)
	y++

	layout.SetRow(y, space, space)
	layout.AddDivider(0, y, 2, 1, true)
	y++

	layout.SetRow(y, 2, 5)
	layout.AddText(0, y, 1, 1, "Chat")

	layout.AddTextMultiline(1, y, 1, 1, st.ChatInput)
	y++
	/*chat := NewGlobal_OpenAI_chat("AssistantPrompt")
	chat.UID = "AssistantPrompt"
	chat.Properties = st.Properties
	if !chat.IsRunning() {
		layout.AddTextMultiline(1, y, 1, 1, st.ChatInput)
	} else {
		ch := layout.AddOpenAI_chat(1, y, 1, 1, chat)
		ch.layout.Back_cd = Paint_GetPalette().GetGrey(0.8)

		layout.VScrollToTheBottom()
	}*/

	layout.SetRow(y, space, space)
	layout.AddDivider(0, y, 2, 1, true)
	y++

	layout.SetRow(y, 2, 5)
	layout.AddText(0, y, 1, 1, "Chat Output")
	layout.AddTextMultiline(1, y, 1, 1, st.ChatOutput)
	y++
}
