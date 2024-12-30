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

var g_Microphone *Microphone

func OpenFile_Microphone() *Microphone {
	if g_Microphone == nil {
		g_Microphone = &Microphone{Enable: true, SampleRate: 44100, Channels: 1}
		_read_file("Microphone-Microphone", g_Microphone)
	}
	return g_Microphone
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
	layout.AddEditboxInt(1, y, 1, 1, &st.SampleRate)
	y++

	layout.AddText(0, y, 1, 1, "Number of Channels")
	layout.AddEditboxInt(1, y, 1, 1, &st.Channels)
	y++
}
