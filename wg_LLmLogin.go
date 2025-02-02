package main

type Model struct {
	Name         string
	Input_price  float64
	Cached_price float64
	Output_price float64
}

type LLMLogin struct {
	Label string

	Anthropic_completion_url string
	OpenAI_completion_url    string

	Api_key_id string

	ChatModels []Model `json:",omitempty"`
	STTModels  []Model `json:",omitempty"`
	TTSModels  []Model `json:",omitempty"`
	TTSVoices  []Model `json:",omitempty"`
}

func (layout *Layout) AddLLmLogin(x, y, w, h int, props *LLMLogin) *LLMLogin {
	layout._createDiv(x, y, w, h, "LLMLogin", props.Build, nil, nil)
	return props
}

func (st *LLMLogin) Build(layout *Layout) {

	layout.SetColumn(0, 1, 5)
	layout.SetColumn(1, 1, 20)

	y := 0
	layout.AddText(0, y, 1, 1, "Anthropic completion url")
	layout.AddEditbox(1, y, 1, 1, &st.Anthropic_completion_url)
	y++

	layout.AddText(0, y, 1, 1, "OpenAI completion url")
	layout.AddEditbox(1, y, 1, 1, &st.OpenAI_completion_url)
	y++

	//api key
	{
		layout.AddText(0, y, 1, 1, "API key")
		KeyEd := layout.AddEditbox(1, y, 1, 1, &st.Api_key_id)
		KeyEd.Formating = false
		y++
	}

	KeyBt := layout.AddButton(1, y, 1, 1, "Get API key")
	KeyBt.Align = 2
	KeyBt.Background = 0
	KeyBt.BrowserUrl = "https://console.x.ai"
	y++

	layout.AddText(0, y, 1, 1, "Chat Models")
	for _, md := range st.ChatModels {
		layout.AddText(1, y, 1, 1, md.Name) //price? ....
		y++
	}
}
