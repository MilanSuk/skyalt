package main

import (
	"strconv"
)

type Settings struct {
	DateFormat  string
	DatePage    int64
	Volume      float64
	Dpi         int
	Dpi_default int

	Fullscreen bool
	Stats      bool

	Theme             string
	ThemePalette      LayoutPalette
	CustomPalette     LayoutPalette
	UseDarkTheme      bool
	UseDarkThemeStart int //hours from midnight
	UseDarkThemeEnd   int
}

func (layout *Layout) AddSettings(x, y, w, h int, props *Settings) *Settings {
	layout._createDiv(x, y, w, h, "Settings", props.Build, nil, nil)
	return props
}

var g_Settings *Settings

func OpenFile_Settings() *Settings {
	if g_Settings == nil {
		g_Settings = &Settings{Volume: 0.5, Theme: "light"}

		_read_file("Settings-Settings", g_Settings)
	}
	return g_Settings
}

func (st *Settings) Build(layout *Layout) {
	layout.SetColumn(0, 1, 100)
	layout.SetColumn(1, 4, 4)
	layout.SetColumn(2, 12, 12)
	layout.SetColumn(3, 1, 100)

	layout.SetRowFromSub(5)

	//Date format
	{
		layout.AddText(1, 1, 1, 1, "Date format")
		df_labels := []string{"EU(31/5/2019)", "US(5/31/2019)", "ISO(2019-5-31)", "Text(May 31 2019)"}
		df_values := []string{"eu", "us", "iso", "text"} //"2base"
		layout.AddCombo(2, 1, 1, 1, &st.DateFormat, df_labels, df_values)
	}

	//DPI, Zoom
	{
		layout.AddText(1, 3, 1, 1, "Zoom(Dots per inch)")

		ZoomDiv := layout.AddLayout(2, 3, 1, 1)
		ZoomDiv.SetColumn(0, 2, 2)
		ZoomDiv.SetColumn(3, 1.5, 1.5)
		ZoomDiv.SetColumn(6, 2, 2)

		//dpi
		DPI := ZoomDiv.AddEditboxInt(0, 0, 1, 1, &st.Dpi)
		DPI.Tooltip = "DPI(Dots per inch)"

		//+
		Add := ZoomDiv.AddButton(2, 0, 1, 1, "+")
		Add.Background = 0.5
		Add.clicked = func() {
			st.Dpi += 3
		}

		//%
		procV := int(float64(st.Dpi) / float64(st.Dpi_default) * 100)
		Info := ZoomDiv.AddText(3, 0, 1, 1, strconv.Itoa(procV)+"%")
		Info.Align_h = 1

		//-
		Sub := ZoomDiv.AddButton(4, 0, 1, 1, "-")
		Sub.Background = 0.5
		Sub.clicked = func() {
			st.Dpi -= 3
		}

		//Reset
		Reset := ZoomDiv.AddButton(6, 0, 1, 1, "Reset")
		Reset.Background = 0.5
		Reset.clicked = func() {
			st.Dpi = st.Dpi_default
		}
	}

	// Theme
	{
		layout.AddText(1, 5, 1, 1, "Theme")

		ThemeDiv := layout.AddLayout(2, 5, 1, 1)
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
				slayout.SetColumn(4, 1, 100)

				slayout.AddText(0, 0, 1, 1, "Primary").Align_h = 1
				slayout.AddText(1, 0, 1, 1, "Secondary").Align_h = 1
				slayout.AddText(2, 0, 1, 1, "Tertiary").Align_h = 1
				slayout.AddText(3, 0, 1, 1, "Background").Align_h = 1
				slayout.AddText(4, 0, 1, 1, "Error").Align_h = 1

				slayout.AddColorPickerButton(0, 1, 1, 1, &st.CustomPalette.P)
				slayout.AddColorPickerButton(1, 1, 1, 1, &st.CustomPalette.S)
				slayout.AddColorPickerButton(2, 1, 1, 1, &st.CustomPalette.T)
				slayout.AddColorPickerButton(3, 1, 1, 1, &st.CustomPalette.B)
				slayout.AddColorPickerButton(4, 1, 1, 1, &st.CustomPalette.E)

				slayout.AddColorPickerButton(0, 2, 1, 1, &st.CustomPalette.OnP)
				slayout.AddColorPickerButton(1, 2, 1, 1, &st.CustomPalette.OnS)
				slayout.AddColorPickerButton(2, 2, 1, 1, &st.CustomPalette.OnT)
				slayout.AddColorPickerButton(3, 2, 1, 1, &st.CustomPalette.OnB)
				slayout.AddColorPickerButton(4, 2, 1, 1, &st.CustomPalette.OnE)
			}
		}
	}

	//Volume
	{
		layout.AddText(1, 7, 1, 1, "Volume")
		volume := st.Volume * 100
		sl := layout.AddSlider(2, 7, 1, 1, &volume, 0, 100, 5)
		sl.changed = func() {
			st.Volume = volume / 100
		}
	}

	layout.AddSwitch(1, 9, 2, 1, "FullScreen(F11)", &st.Fullscreen)

	layout.AddSwitch(1, 10, 2, 1, "Show statistics(F2)", &st.Stats)
}
