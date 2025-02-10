package main

type ModelText struct {
	Name         string
	Input_price  float64
	Cached_price float64
	Output_price float64
}

type ModelSTT struct {
	Name  string
	Price float64 //per hour
}
type ModelTTS struct {
	Name   string
	Price  float64 //per 1M characters
	Voices []string
}

type LLMLogin struct {
	Label string

	Anthropic_completion_url string
	OpenAI_completion_url    string

	Api_key_id string

	ChatModels []ModelText `json:",omitempty"`
	STTModels  []ModelSTT  `json:",omitempty"`
	TTSModels  []ModelTTS  `json:",omitempty"`
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

	if len(st.ChatModels) > 0 {
		layout.AddText(0, y, 1, 1, "Chat")
		for _, md := range st.ChatModels {
			layout.AddText(1, y, 1, 1, md.Name) //price? ....
			y++
		}
	}

	if len(st.STTModels) > 0 {
		layout.AddText(0, y, 1, 1, "Speech to text")
		for _, md := range st.STTModels {
			layout.AddText(1, y, 1, 1, md.Name) //price? ....
			y++
		}
	}

	if len(st.TTSModels) > 0 {
		layout.AddText(0, y, 1, 1, "Text to speech")
		for _, md := range st.TTSModels {
			layout.AddText(1, y, 1, 1, md.Name) //price? ....
			y++
		}
	}

}
