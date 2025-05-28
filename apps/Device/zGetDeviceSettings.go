package main

// Returns device settings(DPI, Date format, Fullscreen, Enable/disable show stats, Volume, Theme, Palette, Microphone).
type GetDeviceSettings struct {
	Out_DateFormat  string
	Out_Rounding    float64
	Out_ScrollThick float64
	Out_Volume      float64
	Out_Dpi         int
	Out_Dpi_default int

	Out_Fullscreen bool
	Out_Stats      bool

	Out_Theme string

	Out_Palette DeviceSettingsPalette
}

func (st *GetDeviceSettings) run(caller *ToolCaller, ui *UI) error {
	source_dev, err := NewDeviceSettings("", caller)
	if err != nil {
		return err
	}

	st.Out_DateFormat = source_dev.DateFormat
	st.Out_Rounding = source_dev.Rounding
	st.Out_ScrollThick = source_dev.ScrollThick
	st.Out_Volume = source_dev.Volume
	st.Out_Dpi = source_dev.Dpi
	st.Out_Dpi_default = source_dev.Dpi_default
	st.Out_Fullscreen = source_dev.Fullscreen
	st.Out_Stats = source_dev.Stats
	st.Out_Theme = source_dev.Theme
	st.Out_Palette = *source_dev.GetPalette()

	return nil
}
