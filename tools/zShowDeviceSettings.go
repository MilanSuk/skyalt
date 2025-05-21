package main

import (
	"strconv"
)

// Show Device settings(DPI, data format, theme, volume, fullscreen/window modes).
type ShowDeviceSettings struct {
}

func (st *ShowDeviceSettings) run(caller *ToolCaller, ui *UI) error {
	source_dev, err := NewDeviceSettings("", caller)
	if err != nil {
		return err
	}

	ui.SetColumn(0, 3, 3)
	ui.SetColumn(1, 5, 100)

	y := 0

	//DPI, Zoom
	{
		ui.AddText(0, y, 1, 1, "Zoom")
		ZoomDiv := ui.AddLayout(1, y, 1, 1)
		y++

		//+
		x := 0
		Add := ZoomDiv.AddButton(x, 0, 1, 1, "+")
		Add.Background = 0.5
		Add.Shortcut = '+'
		Add.clicked = func() error {
			source_dev.Dpi += 3
			return nil
		}
		x++

		//%
		procV := int(float64(source_dev.Dpi) / float64(source_dev.Dpi_default) * 100)
		ZoomDiv.SetColumn(x, 1.5, 1.5)
		Info := ZoomDiv.AddText(x, 0, 1, 1, strconv.Itoa(procV)+"%")
		Info.Align_h = 1
		x++

		//-
		Sub := ZoomDiv.AddButton(x, 0, 1, 1, "-")
		Sub.Background = 0.5
		Sub.Shortcut = '-'
		Sub.clicked = func() error {
			source_dev.Dpi -= 3
			return nil
		}
		x++

		x++ //space

		//Reset
		ZoomDiv.SetColumn(x, 2, 3)
		Reset := ZoomDiv.AddButton(x, 0, 1, 1, "Reset")
		Reset.Background = 0.5
		Reset.Shortcut = '0'
		Reset.clicked = func() error {
			source_dev.Dpi = source_dev.Dpi_default
			return nil
		}
		x++

		x++ //space

		//dpi
		ZoomDiv.SetColumn(x, 1, 1)
		dpi := source_dev.Dpi
		edDPI := ZoomDiv.AddEditboxInt(x, 0, 1, 1, &dpi)
		edDPI.changed = func() error {
			source_dev.SetDPI(dpi)
			return nil
		}
		x++

		ZoomDiv.AddText(x, 0, 1, 1, "DPI")
		x++
	}

	y++

	//Date format
	{
		ui.AddText(0, y, 1, 1, "Date format")
		df_labels := []string{"EU(31/5/2019)", "US(5/31/2019)", "ISO(2019-5-31)", "Text(May 31 2019)"}
		df_values := []string{"eu", "us", "iso", "text"} //"2base"
		ui.AddCombo(1, y, 1, 1, &source_dev.DateFormat, df_labels, df_values)
		y++
	}

	y++

	// Theme
	{
		ui.SetRowFromSub(y, 1, 100)
		ui.AddText(0, y, 1, 1, "Theme")
		ThemeDiv := ui.AddLayout(1, y, 1, 1)
		y++
		{
			ThemeDiv.SetColumn(0, 1, 100)

			ThemeDiv.AddCombo(0, 0, 1, 1, &source_dev.Theme, []string{"Light", "Dark", "Custom"}, []string{"light", "dark", "custom"})

			if source_dev.Theme == "custom" {

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

				slayout.AddColorPickerButton(0, 1, 1, 1, &source_dev.CustomPalette.P)

				slayout.AddColorPickerButton(1, 1, 1, 1, &source_dev.CustomPalette.S)
				slayout.AddColorPickerButton(2, 1, 1, 1, &source_dev.CustomPalette.B)
				slayout.AddColorPickerButton(3, 1, 1, 1, &source_dev.CustomPalette.E)

				slayout.AddColorPickerButton(0, 2, 1, 1, &source_dev.CustomPalette.OnP)
				slayout.AddColorPickerButton(1, 2, 1, 1, &source_dev.CustomPalette.OnS)
				slayout.AddColorPickerButton(2, 2, 1, 1, &source_dev.CustomPalette.OnB)
				slayout.AddColorPickerButton(3, 2, 1, 1, &source_dev.CustomPalette.OnE)
			}
		}
	}

	y++

	//UI rounding
	{
		ui.AddText(0, y, 1, 1, "UI rounding")
		rounding := source_dev.Rounding * 100
		sl := ui.AddSlider(1, y, 1, 1, &rounding, 0, 50, 5)
		sl.changed = func() error {
			source_dev.Rounding = rounding / 100
			return nil
		}
		y++
	}

	y++

	//Volume
	{
		ui.AddText(0, y, 1, 1, "Volume")
		volume := source_dev.Volume * 100
		sl := ui.AddSlider(1, y, 1, 1, &volume, 0, 100, 5)
		sl.changed = func() error {
			source_dev.Volume = volume / 100
			return nil
		}
		y++
	}

	y++

	ui.AddSwitch(0, y, 2, 1, "Fullscreen(F11)", &source_dev.Fullscreen)
	y++
	ui.AddSwitch(0, y, 2, 1, "Performance & resources statistics(F2)", &source_dev.Stats)
	y++

	return nil
}
