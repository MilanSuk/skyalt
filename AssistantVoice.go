package main

import (
	"time"
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
	Mic := layout.AddMicrophone_recorder(0, 0, 1, 1, NewGlobal_Microphone_recorder("AssistantVoice"))
	Mic.Shortcut_key = st.Shortcut
	Mic.Tooltip = "Start/Stop AI Assistant audio recording"
	Mic.Background = st.Button_background
	Mic.start = func() {
		st.VoiceStart_sec = float64(time.Now().UnixMilli()) / 1000
	}
	if Mic.IsRunning() {
		layout.SetColumn(0, 1, 1)
		layout.SetColumn(1, 2, 2)
		cancelBt := layout.AddButton(1, 0, 1, 1, NewButton("Cancel"))
		cancelBt.Cd = Paint_GetPalette().E
		cancelBt.clicked = func() {
			Mic.Cancel()
		}
	}

	//STT
	SttDia := layout.AddDialog("Transcribe")
	SttDia.Layout.SetColumn(0, 1, 20)
	SttDia.Layout.SetRowFromSub(0)
	whisp := NewGlobal_Whispercpp_stt("AssistantVoice")
	stt := SttDia.Layout.AddWhispercpp_stt(0, 0, 1, 1, whisp)

	//Chat
	ChatDia := layout.AddDialog("prompt")
	ChatDia.Layout.SetColumn(0, 1, 20)
	ChatDia.Layout.SetRowFromSub(0)
	ChatDia.Layout.AddAssistantPrompt(0, 0, 1, 1, &NewFile_Assistant().TempPrompt)

	stt.done = func() {
		SttDia.Close()

		NewFile_Assistant().SetVoice([]byte(stt.Out), st.VoiceStart_sec)
		if st.AutoSend > 0 {
			NewFile_Assistant().Send(ChatDia.Layout)
		}
	}

	Mic.done = func() {
		stt.Input_Data = Mic.Out
		SttDia.OpenCentered()
		whisp.Start()
	}

}
