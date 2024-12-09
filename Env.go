package main

import "strconv"

func (st *Env) Build() {
	st.lock.Lock()
	defer st.lock.Unlock()

	st.layout.SetColumn(0, 1, 100)
	st.layout.SetColumn(1, 4, 4)
	st.layout.SetColumn(2, 12, 12)
	st.layout.SetColumn(3, 1, 100)

	y := 0
	y++

	//Date format

	{
		st.layout.AddText(1, y, 1, 1, "Date format")
		df_labels := []string{"EU(31/5/2019)", "US(5/31/2019)", "ISO(2019-5-31)", "Text(May 31 2019)"}
		df_values := []string{"eu", "us", "iso", "text"} //"2base"
		st.layout.AddCombo(2, y, 1, 1, &st.DateFormat, df_labels, df_values)
		y += 2
	}

	//DPI, Zoom

	{
		st.layout.AddText(1, y, 1, 1, "Zoom(Dots per inch)")

		Zoom := st.layout.AddLayout(2, y, 1, 1)
		Zoom.SetColumn(0, 2, 2)
		Zoom.SetColumn(3, 1.5, 1.5)
		Zoom.SetColumn(6, 2, 2)

		//dpi
		DPI := Zoom.AddEditboxInt(0, 0, 1, 1, &st.Dpi)
		DPI.Tooltip = "DPI(Dots per inch)"

		//+
		Add := Zoom.AddButton(2, 0, 1, 1, NewButton("+"))
		//Add.Draw_border = true
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
		//Sub.Draw_border = true
		Sub.Background = 0.5
		Sub.clicked = func() {
			st.Dpi -= 3
		}

		//Reset
		Reset := Zoom.AddButton(6, 0, 1, 1, NewButton("Reset"))
		//Sub.Draw_border = true
		Reset.Background = 0.5
		Reset.clicked = func() {
			st.Dpi = st.Dpi_default
		}

		y += 2
	}

	// Theme
	{
		st.layout.AddText(1, y, 1, 1, "Theme")

		Theme := st.layout.AddLayout(2, y, 1, 1)
		{
			Theme.SetColumn(0, 1, 100)

			Theme.AddCombo(0, 0, 1, 1, &st.Theme, []string{"Light", "Dark", "Custom"}, []string{"light", "dark", "custom"})

			if st.Theme == "custom" {
				st.layout.SetRow(y, 4, 4)

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

				//........
				/*slayout.AddColorPickerButton(0, 1, 1, 1, NewColorPickerButton(&st.CustomPalette.P))
				slayout.AddColorPickerButton(1, 1, 1, 1, NewColorPickerButton(&st.CustomPalette.S))
				slayout.AddColorPickerButton(2, 1, 1, 1, NewColorPickerButton(&st.CustomPalette.T))
				slayout.AddColorPickerButton(3, 1, 1, 1, NewColorPickerButton(&st.CustomPalette.B))
				slayout.AddColorPickerButton(4, 1, 1, 1, NewColorPickerButton(&st.CustomPalette.E))

				slayout.AddColorPickerButton(0, 2, 1, 1, NewColorPickerButton(&st.CustomPalette.OnP))
				slayout.AddColorPickerButton(1, 2, 1, 1, NewColorPickerButton(&st.CustomPalette.OnS))
				slayout.AddColorPickerButton(2, 2, 1, 1, NewColorPickerButton(&st.CustomPalette.OnT))
				slayout.AddColorPickerButton(3, 2, 1, 1, NewColorPickerButton(&st.CustomPalette.OnB))
				slayout.AddColorPickerButton(4, 2, 1, 1, NewColorPickerButton(&st.CustomPalette.OnE))*/
			}
		}

		y += 2
	}

	//Volume

	{
		st.layout.AddText(1, y, 1, 1, "Volume")
		volume := st.Volume * 100
		sl := st.layout.AddSlider(2, y, 1, 1, &volume, 0, 100, 5)
		sl.changed = func() {
			st.Volume = volume / 100
		}

		y += 2
	}

	st.layout.AddSwitch(1, y, 2, 1, "FullScreen(F11)", &st.Fullscreen)
	y++

	st.layout.AddSwitch(1, y, 2, 1, "Show statistics(F2)", &st.Stats)
	y++

}
