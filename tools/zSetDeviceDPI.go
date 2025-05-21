package main

// Set device DPI(dots per inch).
type SetDeviceDPI struct {
	DPI int //new DPI value
}

func (st *SetDeviceDPI) run(caller *ToolCaller, ui *UI) error {
	source_dev, err := NewDeviceSettings("", caller)
	if err != nil {
		return err
	}

	source_dev.SetDPI(st.DPI)

	return nil
}
