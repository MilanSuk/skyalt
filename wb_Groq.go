package main

type Groq struct {
	Enable     bool
	Api_key_id string

	ChatModel string
	STTModel  string
}

func (layout *Layout) AddGroq(x, y, w, h int, props *Groq) *Groq {
	layout._createDiv(x, y, w, h, "Groq", props.Build, nil, nil)
	return props
}

func (st *Groq) Build(layout *Layout) {

	if st.ChatModel == "" {
		st.ChatModel = "llama-3.3-70b-versatile"
	}
	if st.STTModel == "" {
		st.STTModel = "whisper-large-v3"
	}

	layout.SetColumn(0, 1, 4)
	layout.SetColumn(1, 1, 20)

	y := 0

	//enable
	layout.AddSwitch(1, y, 1, 1, "Enable", &st.Enable)
	y++

	//api key
	{
		layout.AddText(0, y, 1, 1, "API key")
		KeyEd := layout.AddEditbox(1, y, 1, 1, &st.Api_key_id)
		KeyEd.Formating = false
		y++
	}

	KeyBt := layout.AddButton(1, y, 1, 1, "Get API key")
	KeyBt.Align = 0
	KeyBt.Background = 0
	KeyBt.BrowserUrl = "https://console.groq.com/keys"
	y++

	layout.AddText(0, y, 1, 1, "Chat Model")
	layout.AddCombo(1, y, 1, 1, &st.ChatModel, Groq_GetChatModelList(), Groq_GetChatModelList())
	y++

	layout.AddText(0, y, 1, 1, "STT Model")
	layout.AddCombo(1, y, 1, 1, &st.STTModel, Groq_GetSTTModelList(), Groq_GetSTTModelList())
	y++

}

func Groq_GetChatModelList() []string {
	return []string{"llama-3.3-70b-versatile"}
}

func Groq_GetSTTModelList() []string {
	return []string{"whisper-large-v3", "whisper-large-v3-turbo", "distil-whisper-large-v3-en"}
}
