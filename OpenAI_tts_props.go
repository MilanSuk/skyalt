package main

// Text To Speech
type OpenAI_tts_props struct {
	Input string `json:"input"`
	Model string `json:"model"`
	Voice string `json:"voice"`
}

func (layout *Layout) AddOpenAI_tts_props(x, y, w, h int) *OpenAI_tts_props {
	props := &OpenAI_tts_props{}
	layout._createDiv(x, y, w, h, "OpenAI_tts_props", props.Build, nil, nil)
	return props
}

func (st *OpenAI_tts_props) Build(layout *Layout) {

	layout.SetColumn(0, 3, 3)
	layout.SetColumn(1, 1, 100)

	y := 0

	layout.AddText(0, y, 1, 1, "Model")
	layout.AddCombo(1, y, 1, 1, &st.Model, OpenAI_GetTTSModelList(), OpenAI_GetTTSModelList())
	y++

	layout.AddText(0, y, 1, 1, "Voice")
	layout.AddCombo(1, y, 1, 1, &st.Voice, OpenAI_GetTTVoiceList(), OpenAI_GetTTVoiceList())
	y++
}

func (props *OpenAI_tts_props) Reset() {
	//...
}
