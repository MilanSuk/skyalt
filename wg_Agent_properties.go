package main

import "fmt"

type Agent_properties struct {
	Model             string
	Stream            bool
	Temperature       float64 //1.0
	Max_tokens        int     //
	Top_p             float64 //1.0
	Frequency_penalty float64 //0
	Presence_penalty  float64 //0
}

func (layout *Layout) AddAgents_properties(x, y, w, h int, props *Agent_properties) *Agent_properties {
	layout._createDiv(x, y, w, h, "Agent_properties", props.Build, nil, nil)
	return props
}

func (st *Agent_properties) Build(layout *Layout) {

	layout.SetColumn(0, 2, 3.5)
	layout.SetColumn(1, 1, 100)

	y := 0

	layout.AddText(0, y, 1, 1, "Model")

	ModelsDia := layout.AddDialog("models")
	ModelsDia.Layout.SetColumn(0, 1, 8)
	ModelsDia.Layout.SetRowFromSub(0, 1, 100)
	ModelsDia.Layout.AddLLMList(0, 0, 1, 1, &st.Model)
	modelBt, modelBtLay := layout.AddButton2(1, y, 1, 1, st.Model)
	modelBt.Background = 0
	modelBt.Align = 0
	modelBt.Icon = "resources/arrow_down.png"
	modelBt.Icon_align = 2
	modelBt.Icon_margin = 0.1
	modelBt.Border = true
	modelBt.clicked = func() {
		ModelsDia.OpenRelative(modelBtLay)
	}

	y++

	login, login_path := FindLoginChatModel(st.Model)
	if login == nil {
		return
	}

	{
		dia, diaLay := layout.AddDialogBorder(login_path, login.Label, 22)
		diaLay.SetColumn(0, 1, 100)
		diaLay.SetRowFromSub(0, 1, 100)
		diaLay.AddLLmLogin(0, 0, 1, 1, login)

		openBt := layout.AddButton(1, y, 1, 1, fmt.Sprintf("%s settings", login.Label))
		openBt.Background = 0.5
		openBt.clicked = func() {
			dia.OpenCentered()
		}
		y++
	}

	tx := layout.AddText(0, y, 1, 1, "Streaming")
	tx.Tooltip = "See result as is generated."
	layout.AddSwitch(1, y, 1, 1, "", &st.Stream)
	y++

	sl := layout.AddSliderEdit(0, y, 2, 1, &st.Temperature, 0, 2, 0.1)
	sl.ValuePointerPrec = 1
	sl.Description = "Temperature"
	sl.Tooltip = "The sampling temperature, between 0 and 1. Higher values like 0.8 will make the output more random, while lower values like 0.2 will make it more focused and deterministic. If set to 0, the model will use log probability to automatically increase the temperature until certain thresholds are hit."
	sl.Description_width = 3.5
	sl.Edit_width = 2
	sl.Legend = true
	y++

	sl = layout.AddSliderEditInt(0, y, 2, 1, &st.Max_tokens, 128, 4096, 1)
	sl.ValuePointerPrec = 0
	sl.Description = "Max Tokens"
	sl.Tooltip = "The maximum number of tokens that can be generated in the chat completion. The total length of input tokens and generated tokens is limited by the model's context length."
	sl.Description_width = 3.5
	sl.Edit_width = 2
	sl.Legend = true
	y++

	if login.Anthropic_completion_url == "" {
		//openai extra

		sl = layout.AddSliderEdit(0, y, 2, 1, &st.Top_p, 0, 1, 0.1)
		sl.ValuePointerPrec = 1
		sl.Description = "Top P"
		sl.Tooltip = "An alternative to sampling with temperature, called nucleus sampling, where the model considers the results of the tokens with top_p probability mass. So 0.1 means only the tokens comprising the top 10% probability mass are considered."
		sl.Description_width = 3.5
		sl.Edit_width = 2
		sl.Legend = true
		y++

		sl = layout.AddSliderEdit(0, y, 2, 1, &st.Frequency_penalty, -2, 2, 0.1)
		sl.ValuePointerPrec = 1
		sl.Description = "Frequency Penalty"
		sl.Tooltip = "Number between -2.0 and 2.0. Positive values penalize new tokens based on their existing frequency in the text so far, decreasing the model's likelihood to repeat the same line verbatim."
		sl.Description_width = 3.5
		sl.Edit_width = 2
		sl.Legend = true
		y++

		sl = layout.AddSliderEdit(0, y, 2, 1, &st.Presence_penalty, -2, 2, 0.1)
		sl.ValuePointerPrec = 1
		sl.Description = "Presence Penalty"
		sl.Tooltip = "Number between -2.0 and 2.0. Positive values penalize new tokens based on whether they appear in the text so far, increasing the model's likelihood to talk about new topics."
		sl.Description_width = 3.5
		sl.Edit_width = 2
		sl.Legend = true
		y++
	}

	ResetBt := layout.AddButton(0, y, 1, 1, "Reset")
	ResetBt.Background = 0.5
	ResetBt.clicked = func() {
		st.Model = "grok-2"
		st.Stream = false
		st.Temperature = 0.2
		st.Max_tokens = 4046
		st.Top_p = 0.7 //1.0
		st.Frequency_penalty = 0
		st.Presence_penalty = 0
	}
	y++
}
