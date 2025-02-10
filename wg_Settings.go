package main

import (
	"strconv"
	"time"
)

type DeviceSettings struct {
	DateFormat  string
	DatePage    int64
	Volume      float64
	Dpi         int
	Dpi_default int

	Fullscreen bool
	Stats      bool

	Theme             string
	ThemePalette      WinCdPalette
	CustomPalette     WinCdPalette
	UseDarkTheme      bool
	UseDarkThemeStart int //hours from midnight
	UseDarkThemeEnd   int
}

func (layout *Layout) AddSettings(x, y, w, h int, props *DeviceSettings) *DeviceSettings {
	layout._createDiv(x, y, w, h, "Settings", props.Build, nil, nil)
	return props
}

func (st *DeviceSettings) Build(layout *Layout) {
	layout.SetColumn(0, 3, 3)
	layout.SetColumn(1, 5, 100)

	y := 1

	//DPI, Zoom
	{
		layout.AddText(0, y, 1, 1, "Zoom")
		ZoomDiv := layout.AddLayout(1, y, 1, 1)
		y++

		//+
		x := 0
		Add := ZoomDiv.AddButton(x, 0, 1, 1, "+")
		Add.Background = 0.5
		Add.clicked = func() {
			st.Dpi += 3
		}
		x++

		//%
		procV := int(float64(st.Dpi) / float64(st.Dpi_default) * 100)
		ZoomDiv.SetColumn(x, 1.5, 1.5)
		Info := ZoomDiv.AddText(x, 0, 1, 1, strconv.Itoa(procV)+"%")
		Info.Align_h = 1
		x++

		//-
		Sub := ZoomDiv.AddButton(x, 0, 1, 1, "-")
		Sub.Background = 0.5
		Sub.clicked = func() {
			st.Dpi -= 3
		}
		x++

		x++ //space

		//Reset
		ZoomDiv.SetColumn(x, 2, 3)
		Reset := ZoomDiv.AddButton(x, 0, 1, 1, "Reset")
		Reset.Background = 0.5
		Reset.clicked = func() {
			st.Dpi = st.Dpi_default
		}
		x++

		x++ //space

		//dpi
		ZoomDiv.SetColumn(x, 1, 1)
		ZoomDiv.AddEditbox(x, 0, 1, 1, &st.Dpi)
		x++

		ZoomDiv.AddText(x, 0, 1, 1, "DPI")
		x++
	}

	y++

	//Date format
	{
		layout.AddText(0, y, 1, 1, "Date format")
		df_labels := []string{"EU(31/5/2019)", "US(5/31/2019)", "ISO(2019-5-31)", "Text(May 31 2019)"}
		df_values := []string{"eu", "us", "iso", "text"} //"2base"
		layout.AddCombo(1, y, 1, 1, &st.DateFormat, df_labels, df_values)
		y++
	}

	y++

	// Theme
	{
		layout.SetRowFromSub(5, 1, 100)
		layout.AddText(0, y, 1, 1, "Theme")
		ThemeDiv := layout.AddLayout(1, y, 1, 1)
		y++
		{
			ThemeDiv.SetColumn(0, 1, 100)

			ThemeDiv.AddCombo(0, 0, 1, 1, &st.Theme, []string{"Light", "Dark", "Custom"}, []string{"light", "dark", "custom"})

			if st.Theme == "custom" {

				ThemeDiv.SetRow(1, 3, 3)
				slayout := ThemeDiv.AddLayout(0, 1, 1, 1)

				slayout.SetColumn(0, 1, 100)
				slayout.SetColumn(1, 1, 100)
				slayout.SetColumn(2, 1, 100)
				slayout.SetColumn(3, 1, 100)

				slayout.AddText(0, 0, 1, 1, "Primary").Align_h = 1
				slayout.AddText(1, 0, 1, 1, "Secondary").Align_h = 1
				slayout.AddText(2, 0, 1, 1, "Background").Align_h = 1
				slayout.AddText(3, 0, 1, 1, "Error").Align_h = 1

				slayout.AddColorPickerButton(0, 1, 1, 1, &st.CustomPalette.P)
				slayout.AddColorPickerButton(1, 1, 1, 1, &st.CustomPalette.S)
				slayout.AddColorPickerButton(2, 1, 1, 1, &st.CustomPalette.B)
				slayout.AddColorPickerButton(3, 1, 1, 1, &st.CustomPalette.E)

				slayout.AddColorPickerButton(0, 2, 1, 1, &st.CustomPalette.OnP)
				slayout.AddColorPickerButton(1, 2, 1, 1, &st.CustomPalette.OnS)
				slayout.AddColorPickerButton(2, 2, 1, 1, &st.CustomPalette.OnB)
				slayout.AddColorPickerButton(3, 2, 1, 1, &st.CustomPalette.OnE)
			}
		}
	}

	y++

	//Volume
	{
		layout.AddText(0, y, 1, 1, "Volume")
		volume := st.Volume * 100
		sl := layout.AddSlider(1, y, 1, 1, &volume, 0, 100, 5)
		sl.changed = func() {
			st.Volume = volume / 100
		}
		y++
	}

	y++

	layout.AddSwitch(0, y, 2, 1, "FullScreen(F11)", &st.Fullscreen)
	y++
	layout.AddSwitch(0, y, 2, 1, "Show statistics(F2)", &st.Stats)
	y++
}

func (env *DeviceSettings) Check() bool {

	backup := *env

	//date
	if env.DateFormat == "" {
		_, zn := time.Now().Zone()
		zn = zn / 3600

		if zn <= -3 && zn >= -10 {
			env.DateFormat = "us"
		} else {
			env.DateFormat = "eu"
		}
	}

	//dpi
	if env.Dpi <= 0 {
		env.Dpi = GetDeviceDPI()
	}
	if env.Dpi < 30 {
		env.Dpi = 30
	}
	if env.Dpi > 5000 {
		env.Dpi = 5000
	}

	if env.Dpi_default != GetDeviceDPI() {
		env.Dpi_default = GetDeviceDPI()
	}

	//theme
	if env.Theme == "" {
		env.Theme = "light"
	}
	if env.CustomPalette.P.A == 0 {
		env.CustomPalette = g_theme_light
	}

	return *env == backup
}
