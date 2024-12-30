package main

type Anthropic_completion_props struct {
	Model string `json:"model"`

	System   string                  `json:"system"`
	Messages []OpenAI_completion_msg `json:"messages"`
	Stream   bool                    `json:"stream"`

	Temperature float64 `json:"temperature"` //1.0
	Max_tokens  int     `json:"max_tokens"`

	//Response_format *OpenAI_completion_format `json:"response_format"`
}

func (layout *Layout) AddAnthropic_completion_props(x, y, w, h int, props *Anthropic_completion_props) *Anthropic_completion_props {
	layout._createDiv(x, y, w, h, "Anthropic_completion_props", props.Build, nil, nil)
	return props
}

func (st *Anthropic_completion_props) Build(layout *Layout) {
	layout.SetColumn(0, 2, 3.5)
	layout.SetColumn(1, 1, 100)

	y := 0

	layout.AddText(0, y, 1, 1, "Model")
	layout.AddCombo(1, y, 1, 1, &st.Model, Anthropic_GetChatModelList(), Anthropic_GetChatModelList())
	y++

	sl := layout.AddSliderEdit(0, y, 2, 1, &st.Temperature, 0, 2, 0.1)
	sl.ValuePrec = 1
	sl.Description = "Temperature"
	sl.Tooltip = "The sampling temperature, between 0 and 1. Higher values like 0.8 will make the output more random, while lower values like 0.2 will make it more focused and deterministic. If set to 0, the model will use log probability to automatically increase the temperature until certain thresholds are hit."
	sl.Description_width = 3.5
	sl.Edit_width = 2
	sl.Legend = true
	y++

	sl = layout.AddSliderEditInt(0, y, 2, 1, &st.Max_tokens, 128, 4096, 1)
	sl.ValuePrec = 0
	sl.Description = "Max Tokens"
	sl.Tooltip = "The maximum number of tokens that can be generated in the chat completion. The total length of input tokens and generated tokens is limited by the model's context length."
	sl.Description_width = 3.5
	sl.Edit_width = 2
	sl.Legend = true
	y++

	ResetBt := layout.AddButton(0, y, 1, 1, NewButton("Reset"))
	ResetBt.Background = 0.5
	ResetBt.clicked = func() {
		st.Reset()
	}
	y++
}

func (props *Anthropic_completion_props) Reset() {
	if props.Model == "" {
		props.Model = OpenFile_Anthropic().ChatModel
	}
	//props.Stream = true
	props.Temperature = 1.0
	props.Max_tokens = 4046
	//props.Top_p = 0.7 //1.0
	//props.Frequency_penalty = 0
	//props.Presence_penalty = 0
}
