package main

type Microphone struct {
	Enable bool

	Sample_rate int
	Channels    int
}

func (layout *Layout) AddMicrophone(x, y, w, h int, props *Microphone) *Microphone {
	layout._createDiv(x, y, w, h, "Microphone", props.Build, nil, nil)
	return props
}

func (st *Microphone) Build(layout *Layout) {

	if st.Sample_rate == 0 {
		st.Sample_rate = 44100
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
	layout.AddEditbox(1, y, 1, 1, &st.Sample_rate)
	y++

	layout.AddText(0, y, 1, 1, "Number of Channels")
	layout.AddEditbox(1, y, 1, 1, &st.Channels)
	y++
}
