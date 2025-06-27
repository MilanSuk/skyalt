package main

import "strings"

type Microphone struct {
	Shortcut byte

	Format                     string
	Transcribe                 bool
	Output_onlyTranscript      bool
	Transcribe_response_format string //"verbose_json"

	started  func()
	finished func(audio []byte, transcript string)
}

func (layout *Layout) AddMicrophone(x, y, w, h int) *Microphone {
	props := &Microphone{}
	layout._createDiv(x, y, w, h, "Microphone", props.Build, nil, nil)
	return props
}

func (st *Microphone) Build(layout *Layout) {
	if st.Format == "" {
		st.Format = "wav"
	}
	st.Format = strings.ToLower(st.Format)

	layout.SetColumn(0, 1, 100)
	layout.SetRow(0, 1, 100)

	micBt := layout.AddButton(0, 0, 1, 1, "")
	micBt.Background = 0
	micBt.Icon_margin = 0.15
	micBt.IconPath = "resources/mic.png"
	micBt.Shortcut_key = st.Shortcut
	micBt.Tooltip = "Start recording audio"

	if layout.ui.router.mic.Find(layout.UID) != nil {
		micBt.Background = 1 //active
		micBt.Cd = layout.GetPalette().E
		micBt.Tooltip = "Stop recording audio"
	}

	micBt.clicked = func() {

		if layout.ui.router.mic.Find(layout.UID) == nil {

			_, err := layout.ui.router.mic.Start(layout.UID)
			if err != nil {
				return //err ....
			}

			if st.started != nil {
				st.started()
			}
		} else {

			out_raw, err := layout.ui.router.mic.Finished(layout.UID, false)
			if err != nil {
				return //err ....
			}

			//convert
			Out_bytes, err := FFMpeg_convertIntoFile(&out_raw, st.Format, 16000)
			if err != nil {
				return //err ....
			}

			var transcript string
			if st.Transcribe {
				comp := LLMTranscribe{
					AudioBlob:       Out_bytes,
					BlobFileName:    "blob." + st.Format,
					Temperature:     0,
					Response_format: st.Transcribe_response_format,
				}

				err = layout.ui.router.llms.Transcribe(&comp)
				if err != nil {
					return //err ....
				}
				transcript = string(comp.Out_Output)
			}

			if st.finished != nil {
				if st.Output_onlyTranscript {
					st.finished(nil, transcript)
				} else {
					st.finished(Out_bytes, transcript)
				}
			}
		}
	}
}
