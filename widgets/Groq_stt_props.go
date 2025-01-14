package main

import (
	"mime/multipart"
)

type Groq_stt_props struct {
	Model           string
	Temperature     float64
	Response_format string
}

func (layout *Layout) AddGroq_stt_props(x, y, w, h int) *Groq_stt_props {
	props := &Groq_stt_props{}
	layout._createDiv(x, y, w, h, "Groq_stt_props", props.Build, nil, nil)
	return props
}

func (st *Groq_stt_props) Build(layout *Layout) {

	layout.SetColumn(0, 3, 3)
	layout.SetColumn(1, 1, 100)

	y := 0

	layout.AddText(0, y, 1, 1, "Model")
	layout.AddCombo(1, y, 1, 1, &st.Model, Groq_GetSTTModelList(), Groq_GetSTTModelList())
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

	ResetBt := layout.AddButton(0, y, 1, 1, "Reset")
	ResetBt.Background = 0.5
	ResetBt.clicked = func() {
		st.Reset()
	}
	y++
}

func (props *Groq_stt_props) Reset() {
	props.Model = OpenFile_Groq().STTModel
	props.Temperature = 0                  //default
	props.Response_format = "verbose_json" //"text"
}

func (props *Groq_stt_props) Write(w *multipart.Writer) {
	w.WriteField("model", props.Model)
	w.WriteField("response_format", props.Response_format)

	if props.Response_format == "verbose_json" {
		//w.WriteField("timestamp_granularities[]", "word")
	}
}
