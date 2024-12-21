package main

type OpenAI_chat_msg struct {
	Role    string `json:"role"` //"system", "user", "assistant"
	Content string `json:"content"`
}

type OpenAI_chat_format struct {
	Type string `json:"type"` //json_object
	//Json_schema ...
}

func (layout *Layout) AddOpenAI_chat_props(x, y, w, h int, props *OpenAI_chat_props) *OpenAI_chat_props {
	layout._createDiv(x, y, w, h, "OpenAI_chat_props", props.Build, nil, nil)
	return props
}

type OpenAI_chat_props struct {
	Model    string            `json:"model"`
	Messages []OpenAI_chat_msg `json:"messages"`
	Stream   bool              `json:"stream"`

	Temperature       float64 `json:"temperature"`       //1.0
	Max_tokens        int     `json:"max_tokens"`        //
	Top_p             float64 `json:"top_p"`             //1.0
	Frequency_penalty float64 `json:"frequency_penalty"` //0
	Presence_penalty  float64 `json:"presence_penalty"`  //0

	Response_format OpenAI_chat_format `json:"response_format"`
}

func (st *OpenAI_chat_props) Build(layout *Layout) {

	layout.SetColumn(0, 2, 3.5)
	layout.SetColumn(1, 1, 100)

	y := 0

	layout.AddText(0, y, 1, 1, "Model")
	layout.AddCombo(1, y, 1, 1, &st.Model, OpenAI_GetChatModelList(), OpenAI_GetChatModelList())
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

	sl = layout.AddSliderEdit(0, y, 2, 1, &st.Top_p, 0, 1, 0.1)
	sl.ValuePrec = 1
	sl.Description = "Top P"
	sl.Tooltip = "An alternative to sampling with temperature, called nucleus sampling, where the model considers the results of the tokens with top_p probability mass. So 0.1 means only the tokens comprising the top 10% probability mass are considered."
	sl.Description_width = 3.5
	sl.Edit_width = 2
	sl.Legend = true
	y++

	sl = layout.AddSliderEdit(0, y, 2, 1, &st.Frequency_penalty, -2, 2, 0.1)
	sl.ValuePrec = 1
	sl.Description = "Frequency Penalty"
	sl.Tooltip = "Number between -2.0 and 2.0. Positive values penalize new tokens based on their existing frequency in the text so far, decreasing the model's likelihood to repeat the same line verbatim."
	sl.Description_width = 3.5
	sl.Edit_width = 2
	sl.Legend = true
	y++

	sl = layout.AddSliderEdit(0, y, 2, 1, &st.Presence_penalty, -2, 2, 0.1)
	sl.ValuePrec = 1
	sl.Description = "Presence Penalty"
	sl.Tooltip = "Number between -2.0 and 2.0. Positive values penalize new tokens based on whether they appear in the text so far, increasing the model's likelihood to talk about new topics."
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

func (props *OpenAI_chat_props) Reset() {
	if props.Model == "" {
		props.Model = NewFile_OpenAI().ChatModel
	}
	props.Stream = true
	props.Temperature = 1.0
	props.Max_tokens = 4046
	props.Top_p = 0.7 //1.0
	props.Frequency_penalty = 0
	props.Presence_penalty = 0
	//props.Seed = -1
	//props.Cache_prompt = false
	//props.Stop = []string{"</s>", "<|im_start|>", "<|im_end|>", "Llama:", "User:"}
}
