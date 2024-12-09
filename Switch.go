package main

import "image/color"

func (st *Switch) Draw(rect Rect) {
	layout := st.layout

	layout.Paint_cursor("hand", rect)
	layout.Paint_tooltip(st.Tooltip, rect)

	//colors
	P := layout.GetPalette().P
	B := layout.GetPalette().B
	onB := layout.GetPalette().OnB

	var cd, cd2 color.RGBA
	if *st.Value {
		cd = P
		cd2 = B
	} else {
		cd = layout.GetPalette().GetGrey(0.3)
		cd2 = B
	}

	cd_over := cd
	cd_down := cd
	cd2_over := cd2
	cd2_down := cd2

	cd_text := onB
	cd_text_over := cd_text
	cd_text_down := cd_text

	cd_over = Color_Aprox(cd_over, B, 0.2)
	cd_down = Color_Aprox(cd_down, onB, 0.4)

	cd2_over = Color_Aprox(cd2_over, onB, 0.2)
	cd2_down = Color_Aprox(cd2_down, B, 0.4)

	cd_text_over = Color_Aprox(cd_text_over, B, 0.2)
	cd_text_down = Color_Aprox(cd_text_down, onB, 0.4)

	//coord
	rc := rect
	rectLabel := rc
	{
		height := rc.H * 0.9 //OsMinFloat(rc.H*0.9, 0.7)
		if height > 0.7 {
			height = 0.7
		}

		width := height * 3 / 2
		if st.Label != "" {

			x := rc.X
			rc = Rect_centerFull(rc, width, height)
			rc.X = x

			rectLabel = rectLabel.CutLeft(rc.W + 0.1)
		} else {
			//center
			rc = Rect_centerFull(rc, width, height)
		}
	}

	//draw switch
	layout.Paint_rect(rc, cd, cd_over, cd_down, 0)

	rc = rc.Cut(0.1)
	rc.W /= 2
	if !*st.Value {
		layout.Paint_rect(rc, cd2, cd2_over, cd2_down, 0)

		//0
		rc = rc.Cut(0.1)
		layout.Paint_line(rc, 0, 0, 1, 1, cd, 0.05)
		layout.Paint_line(rc, 0, 1, 1, 0, cd, 0.05)

	} else {
		rc.X += rc.W
		layout.Paint_rect(rc, cd2, cd2_over, cd2_down, 0)

		//I
		rc = rc.Cut(0.1)
		layout.Paint_line(rc, 1.0/3, 0.9, 0.05, 2.0/3, cd, 0.05)
		layout.Paint_line(rc, 1.0/3, 0.9, 0.95, 1.0/4, cd, 0.05)
	}

	//draw label
	if st.Label != "" {
		layout.Paint_text(rectLabel, st.Label, "", cd_text, cd_text_over, cd_text_down, true, false, 0, 1, true, false, false, 0.1)
	}
}

func (st *Switch) Input(in LayoutInput) {
	clicked := false

	active := in.IsActive
	inside := in.IsInside && (active || !in.IsUse)
	clicked = in.IsUp && active && inside

	if clicked {
		*st.Value = !*st.Value
		if st.changed != nil {
			st.changed()
		}
	}
}
