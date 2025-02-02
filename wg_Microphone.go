package main

type Microphone struct {
	Enable bool

	SampleRate int
	Channels   int
}

func (layout *Layout) AddMicrophone(x, y, w, h int) *Microphone {
	props := &Microphone{}
	layout._createDiv(x, y, w, h, "Microphone", props.Build, nil, nil)
	return props
}

func (st *Microphone) Build(layout *Layout) {

	if st.SampleRate == 0 {
		st.SampleRate = 44100
	}
	if st.Channels == 0 {
		st.Channels = 1
	}

	layout.SetColumn(0, 5, 5)
	layout.SetColumn(1, 1, 10)

	y := 0
	layout.AddSwitch(1, y, 1, 1, "Enable", &st.Enable)
	y++

	layout.AddText(0, y, 1, 1, "Sample rate")
	layout.AddEditbox(1, y, 1, 1, &st.SampleRate)
	y++

	layout.AddText(0, y, 1, 1, "Number of Channels")
	layout.AddEditbox(1, y, 1, 1, &st.Channels)
	y++
}
