package main

// Set device default DPI(dots per inch). [ignore]
type SetDeviceDPIDefault struct {
	DPI int
}

func (st *SetDeviceDPIDefault) run(caller *ToolCaller, ui *UI) error {
	source_dev, err := NewDeviceSettings("", caller)
	if err != nil {
		return err
	}

	source_dev.Dpi_default = st.DPI
	if source_dev.Dpi <= 0 {
		source_dev.Dpi = source_dev.Dpi_default
	}

	return nil
}
