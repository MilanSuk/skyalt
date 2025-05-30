package main

// Show/Hide device statistics. [ignore]
type SetDeviceStats struct {
	Show bool
}

func (st *SetDeviceStats) run(caller *ToolCaller, ui *UI) error {
	source_dev, err := NewDeviceSettings("")
	if err != nil {
		return err
	}

	source_dev.Stats = st.Show

	return nil
}
