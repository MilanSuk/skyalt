package main

import (
	"strconv"
)

type Env struct {
	DateFormat  string
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

func (layout *Layout) AddEnv(x, y, w, h int, props *Env) *Env {
	layout._createDiv(x, y, w, h, "Env", props.Build, nil, nil)
	return props
}

var g_Env *Env

func NewFile_Env() *Env {
	if g_Env == nil {
		g_Env = &Env{Volume: 0.5, Theme: "light"}

		_read_file("Env-Env", g_Env)
	}
	return g_Env
}

func (st *Env) Build(layout *Layout) {

	layout.SetColumn(0, 1, 100)
	layout.SetColumn(1, 4, 4)
	layout.SetColumn(2, 12, 12)
	layout.SetColumn(3, 1, 100)

	y := 0
	y++

	//Date format

	{
		layout.AddText(1, y, 1, 1, "Date format")
		df_labels := []string{"EU(31/5/2019)", "US(5/31/2019)", "ISO(2019-5-31)", "Text(May 31 2019)"}
		df_values := []string{"eu", "us", "iso", "text"} //"2base"
		layout.AddCombo(2, y, 1, 1, &st.DateFormat, df_labels, df_values)
		y += 2
	}

	//DPI, Zoom

	{
		layout.AddText(1, y, 1, 1, "Zoom(Dots per inch)")

		Zoom := layout.AddLayout(2, y, 1, 1)
		Zoom.SetColumn(0, 2, 2)
		Zoom.SetColumn(3, 1.5, 1.5)
		Zoom.SetColumn(6, 2, 2)

		//dpi
		DPI := Zoom.AddEditboxInt(0, 0, 1, 1, &st.Dpi)
		DPI.Tooltip = "DPI(Dots per inch)"

		//+
		Add := Zoom.AddButton(2, 0, 1, 1, NewButton("+"))
		Add.Background = 0.5
		Add.clicked = func() {
			st.Dpi += 3
		}

		//%
		procV := int(float64(st.Dpi) / float64(st.Dpi_default) * 100)
		Info := Zoom.AddText(3, 0, 1, 1, strconv.Itoa(procV)+"%")
		Info.Align_h = 1

		//-
		Sub := Zoom.AddButton(4, 0, 1, 1, NewButton("-"))
		Sub.Background = 0.5
		Sub.clicked = func() {
			st.Dpi -= 3
		}

		//Reset
		Reset := Zoom.AddButton(6, 0, 1, 1, NewButton("Reset"))
		Reset.Background = 0.5
		Reset.clicked = func() {
			st.Dpi = st.Dpi_default
		}

		y += 2
	}

	// Theme
	{
		layout.AddText(1, y, 1, 1, "Theme")

		Theme := layout.AddLayout(2, y, 1, 1)
		{
			Theme.SetColumn(0, 1, 100)

			Theme.AddCombo(0, 0, 1, 1, &st.Theme, []string{"Light", "Dark", "Custom"}, []string{"light", "dark", "custom"})

			if st.Theme == "custom" {
				layout.SetRow(y, 4, 4)

				Theme.SetRow(1, 3, 3)
				slayout := Theme.AddLayout(0, 1, 1, 1)

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

		y += 2
	}

	//Volume

	{
		layout.AddText(1, y, 1, 1, "Volume")
		volume := st.Volume * 100
		sl := layout.AddSlider(2, y, 1, 1, &volume, 0, 100, 5)
		sl.changed = func() {
			st.Volume = volume / 100
		}

		y += 2
	}

	layout.AddSwitch(1, y, 2, 1, "FullScreen(F11)", &st.Fullscreen)
	y++

	layout.AddSwitch(1, y, 2, 1, "Show statistics(F2)", &st.Stats)
	y++
}
