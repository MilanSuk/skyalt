package main

func (st *Microphone) Build() {
	st.lock.Lock()
	defer st.lock.Unlock()

	if st.SampleRate == 0 {
		st.SampleRate = 44100
	}
	if st.Channels == 0 {
		st.Channels = 1
	}

	st.layout.SetColumn(0, 5, 5)
	st.layout.SetColumn(1, 1, 10)

	y := 0
	st.layout.AddSwitch(1, y, 1, 1, "Enable", &st.Enable)
	y++

	st.layout.AddText(0, y, 1, 1, "Sample rate")
	st.layout.AddEditboxInt(1, y, 1, 1, &st.SampleRate)
	y++

	st.layout.AddText(0, y, 1, 1, "Number of Channels")
	st.layout.AddEditboxInt(1, y, 1, 1, &st.Channels)
	y++
}
