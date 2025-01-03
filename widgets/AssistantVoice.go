package main

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
	if Mic.IsRunning() {
		layout.SetColumn(0, 1, 1)
		layout.SetColumn(1, 2, 2)
		cancelBt := layout.AddButton(1, 0, 1, 1, "Cancel")
		cancelBt.Cd = Paint_GetPalette().E
		cancelBt.clicked = func() {
			Mic.Cancel()
		}
	}

	//STT
	whisp := NewGlobal_Whispercpp_stt("AssistantVoice")
	whisp.done = func() {
		OpenFile_AssistantChat().SetVoice([]byte(whisp.Out), Mic.startUnixTime)
		if st.AutoSend > 0 {
			OpenFile_AssistantChat().Send()
		}
	}

	Mic.done = func() {
		whisp.Input_Data = Mic.Out
		//SttDia.OpenCentered()
		whisp.Start()
	}
}
