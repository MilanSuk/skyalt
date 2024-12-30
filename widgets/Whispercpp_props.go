package main

import (
	"mime/multipart"
)

type Whispercpp_props struct {
	Model           string
	Temperature     float64
	Response_format string
}

func NewWhispercpp_props() *Whispercpp_props {
	props := &Whispercpp_props{}
	props.Reset()
	props.Response_format = "verbose_json"
	return props
}

func (layout *Layout) AddWhispercpp_props(x, y, w, h int, props *Whispercpp_props) *Whispercpp_props {
	layout._createDiv(x, y, w, h, "Whispercpp_props", props.Build, nil, nil)
	return props
}

func (st *Whispercpp_props) Build(layout *Layout) {

	layout.SetColumn(0, 3, 3)
	layout.SetColumn(1, 1, 100)

	y := 0

	layout.AddText(0, y, 1, 1, "Model")
	_, model_pathes := OpenFile_Whispercpp().GetModelList()
	layout.AddCombo(1, y, 1, 1, &st.Model, model_pathes, model_pathes)
	y++

	res := layout.AddText(0, y, 1, 1, "Response format")
	res.Tooltip = "The output response format."
	formats := []string{"verbose_json", "json", "text", "srt", "vtt"}
	layout.AddCombo(1, y, 1, 1, &st.Response_format, formats, formats)
	y++

	tx := layout.AddText(0, y, 1, 1, "Temperature")
	tx.Tooltip = "The sampling temperature, between 0 and 1. Higher values like 0.8 will make the output more random, while lower values like 0.2 will make it more focused and deterministic. If set to 0, the model will use log probability to automatically increase the temperature until certain thresholds are hit."
	sl := layout.AddSliderEdit(1, y, 1, 1, &st.Temperature, 0, 1, 0.01)
	sl.Description_width = 0
	sl.Edit_width = 2
	y++

	ResetBt := layout.AddButton(0, y, 1, 1, NewButton("Reset"))
	ResetBt.Background = 0.5
	ResetBt.clicked = func() {
		st.Reset()
	}
	y++
}

func (props *Whispercpp_props) Reset() {
	if props.Model == "" {
		props.Model = OpenFile_Whispercpp().Model
	}
	props.Temperature = 0 //default
	props.Response_format = "verbose_json"
}

func (props *Whispercpp_props) Write(w *multipart.Writer) {
	w.WriteField("model", props.Model)
	w.WriteField("response_format", props.Response_format)

	if props.Response_format == "verbose_json" {
		w.WriteField("timestamp_granularities[]", "word")
	}
}
