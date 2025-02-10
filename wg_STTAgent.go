package main

type STTAgent struct {
	Model           string  `json:"model"`
	Temperature     float64 `json:"temperature"`     //0
	Response_format string  `json:"response_format"` //"verbose_json", "json", "text", "srt", "vtt"
}
