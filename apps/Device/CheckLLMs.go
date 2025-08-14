package main

// [ignore]
type CheckLLMs struct {
	Out_AppProvider_error   string
	Out_CodeProvider_error  string
	Out_ImageProvider_error string
	Out_STTProvider_error   string
}

func (st *CheckLLMs) run(caller *ToolCaller, ui *UI) error {
	source_dev, err := NewDeviceSettings("")
	if err != nil {
		return err
	}

	providerErr := source_dev.CheckProvider(source_dev.App_provider)
	if providerErr != nil {
		st.Out_AppProvider_error = providerErr.Error()
	}

	providerErr = source_dev.CheckProvider(source_dev.Code_provider)
	if providerErr != nil {
		st.Out_CodeProvider_error = providerErr.Error()
	}

	providerErr = source_dev.CheckProvider(source_dev.Image_provider)
	if providerErr != nil {
		st.Out_ImageProvider_error = providerErr.Error()
	}

	providerErr = source_dev.CheckProvider(source_dev.STT_provider)
	if providerErr != nil {
		st.Out_STTProvider_error = providerErr.Error()
	}

	return nil
}
