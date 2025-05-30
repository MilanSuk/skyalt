package main

// Changed device fullscreen mode.
type SetDeviceFullscreen struct {
	Enable bool //Enable fullscreen mode
}

func (st *SetDeviceFullscreen) run(caller *ToolCaller, ui *UI) error {
	source_dev, err := NewDeviceSettings("")
	if err != nil {
		return err
	}

	source_dev.Fullscreen = st.Enable

	return nil
}
