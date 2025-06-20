package main

// Show Microphone settings. User can enable/disable microphone and change number of channels and sample rate.
type ShowMicrophoneSettings struct {
}

func (st *ShowMicrophoneSettings) run(caller *ToolCaller, ui *UI) error {
	source_mic, err := NewMicrophoneSettings("")
	if err != nil {
		return err
	}

	ui.SetColumn(0, 1, 5)
	ui.SetColumn(1, 1, 20)

	//title
	ui.AddTextLabel(0, 0, 2, 1, "Microphone Settings")

	y := 1

	ui.AddSwitch(1, y, 1, 1, "", &source_mic.Enable)
	y++

	ui.AddText(0, y, 1, 1, "Channels")
	ui.AddEditboxInt(1, y, 1, 1, &source_mic.Channels)
	y++

	ui.AddText(0, y, 1, 1, "Sample rate")
	ui.AddEditboxInt(1, y, 1, 1, &source_mic.Sample_rate)
	y++

	return nil
}
