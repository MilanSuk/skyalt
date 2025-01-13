package main

import (
	"github.com/go-audio/audio"
)

type AssistantVoice struct {
	Shortcut          byte
	Button_background float64
	Button_margin     float64
	AutoSend          float64

	VoiceStart_sec float64
}

func (layout *Layout) AddAssistantVoice(x, y, w, h int) *AssistantVoice {
	props := &AssistantVoice{}
	layout._createDiv(x, y, w, h, "AssistantVoice", props.Build, nil, nil)
	return props
}

func (st *AssistantVoice) Build(layout *Layout) {

	layout.SetColumn(0, 1, 1)
	layout.SetRow(0, 1, 100)

	//Recorder
	Mic := layout.AddMicrophone_recorder(0, 0, 1, 1, OpenMemory_Microphone_recorder("AssistantVoice"))
	Mic.Shortcut_key = st.Shortcut
	Mic.Tooltip = "Start/Stop AI Assistant audio recording"
	Mic.Background = st.Button_background
	if Mic.FindJob() != nil {
		layout.SetColumn(0, 1, 1)
		layout.SetColumn(1, 2, 2)
		cancelBt := layout.AddButton(1, 0, 1, 1, "Cancel")
		cancelBt.Cd = Paint_GetPalette().E
		cancelBt.clicked = func() {
			Mic.Cancel()
		}
	}

	service := OpenFile_AssistantChat().Model.GetSTTService()

	done := func(out string) {
		OpenFile_AssistantChat().SetVoice([]byte(out), Mic.Out_startUnixTime, service)
		if st.AutoSend > 0 {
			OpenFile_AssistantChat().Send()
			layout.Redraw()
		}
	}

	switch service {
	case "whispercpp":
		stt := OpenMemory_Whispercpp_stt("AssistantVoice")
		stt.Properties.Model = OpenFile_AssistantChat().Model.STTModel
		stt.done = done
		Mic.done = func(out audio.IntBuffer) {
			stt.Input_Data = out
			stt.Start()
		}

	case "openai":
		stt := OpenMemory_OpenAI_stt("AssistantVoice")
		stt.Properties.Model = OpenFile_AssistantChat().Model.STTModel
		stt.done = done
		Mic.done = func(out audio.IntBuffer) {
			stt.Input_Data = out
			stt.Start()
		}

	case "groq":
		stt := OpenMemory_Groq_stt("AssistantVoice")
		stt.Properties.Model = OpenFile_AssistantChat().Model.STTModel
		stt.done = done
		Mic.done = func(out audio.IntBuffer) {
			stt.Input_Data = out
			stt.Start()
		}
	}

}
