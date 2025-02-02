package main

type OpenAI struct {
	Enable     bool
	Api_key_id string

	ChatModel string
	STTModel  string
	TTSModel  string
	TTSVoice  string
}

func (layout *Layout) AddOpenAI(x, y, w, h int, props *OpenAI) *OpenAI {
	layout._createDiv(x, y, w, h, "OpenAI", props.Build, nil, nil)
	return props
}

func (st *OpenAI) Build(layout *Layout) {

	if st.ChatModel == "" {
		st.ChatModel = "gpt-3.5-turbo"
	}
	if st.STTModel == "" {
		st.STTModel = "whisper-1"
	}
	if st.TTSModel == "" {
		st.TTSModel = "tts-1"
	}
	if st.TTSVoice == "" {
		st.TTSVoice = "alloy"
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
	KeyBt.BrowserUrl = "https://platform.openai.com/account/api-keys"
	y++

	layout.AddText(0, y, 1, 1, "Chat Model")
	layout.AddCombo(1, y, 1, 1, &st.ChatModel, OpenAI_GetChatModelList(), OpenAI_GetChatModelList())
	y++

	layout.AddText(0, y, 1, 1, "STT Model")
	layout.AddCombo(1, y, 1, 1, &st.STTModel, OpenAI_GetSTTModelList(), OpenAI_GetSTTModelList())
	y++

	layout.AddText(0, y, 1, 1, "TTS Model")
	layout.AddCombo(1, y, 1, 1, &st.TTSModel, OpenAI_GetTTSModelList(), OpenAI_GetTTSModelList())
	y++

	layout.AddText(0, y, 1, 1, "TTS Voice")
	layout.AddCombo(1, y, 1, 1, &st.TTSVoice, OpenAI_GetTTVoiceList(), OpenAI_GetTTVoiceList())
	y++
}

func OpenAI_GetChatModelList() []string {
	return []string{"gpt-3.5-turbo", "gpt-4", "gpt-4-turbo", "gpt-4o", "gpt-4o-mini"}
}
func OpenAI_GetSTTModelList() []string {
	return []string{"whisper-1"}
}
func OpenAI_GetTTSModelList() []string {
	return []string{"tts-1", "tts-1-hd"}
}
func OpenAI_GetTTVoiceList() []string {
	return []string{"alloy", "echo", "fable", "onyx", "nova", "shimmer"}
}
