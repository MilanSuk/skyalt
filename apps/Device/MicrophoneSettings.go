package main

type MicrophoneSettings struct {
	Enable      bool
	Sample_rate int
	Channels    int
}

func NewMicrophoneSettings(file string) (*MicrophoneSettings, error) {
	st := &MicrophoneSettings{}

	st.Enable = true
	st.Sample_rate = 44100
	st.Channels = 1

	return LoadFile(file, "MicrophoneSettings", "json", st, true)
}
