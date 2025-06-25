package main

// Returns map settings(Enable/disable, Tiles url, Copyright url)
type GetMapSettings struct {
	Out_settings MapSettings
}

func (st *GetMapSettings) run(caller *ToolCaller, ui *UI) error {
	source_map, err := NewMapSettings("")
	if err != nil {
		return err
	}

	st.Out_settings = *source_map
	return nil
}
