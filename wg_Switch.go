package main

import (
	"fmt"
	"image/color"
)

type Switch struct {
	Tooltip string

	Label string
	Value *bool

	changed func()
}

func (layout *Layout) AddSwitch(x, y, w, h int, label string, value *bool) *Switch {
	props := &Switch{Label: label, Value: value}
	lay := layout._createDiv(x, y, w, h, "Switch", nil, props.Draw, props.Input)
	lay.fnGetLLMTip = props.getLLMTip
	return props
}

func (st *Switch) getLLMTip(layout *Layout) string {
	return Layout_buildLLMTip("Switch", fmt.Sprintf("%v", *st.Value), false, st.Tooltip)
}

func (st *Switch) Draw(rect Rect, layout *Layout) (paint LayoutPaint) {
	paint.Cursor("hand", rect)
	paint.Tooltip(st.Tooltip, rect)

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

	cd2_over := cd2
	cd2_down := cd2

	cd_text := onB
	cd_text_over := cd_text
	cd_text_down := cd_text

	cd2_over = Color_Aprox(cd2_over, onB, 0.2)
	cd2_down = Color_Aprox(cd2_down, B, 0.4)

	cd_text_over = Color_Aprox(cd_text_over, B, 0.2)
	cd_text_down = Color_Aprox(cd_text_down, onB, 0.4)

	//coord
	rc := rect
	rectLabel := rc
	{
		height := rc.H * 0.9
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
	paint.RectRad(rc, cd, cd, cd, 0, rc.H*layout.getRounding())

	rc = rc.Cut(0.1)
	rc.W /= 2
	in_rounding := rc.H * layout.getRounding()
	if !*st.Value {
		paint.RectRad(rc, cd2, cd2_over, cd2_down, 0, in_rounding)

		//0
		rc = rc.Cut(0.1)
		paint.Line(rc, 0, 0, 1, 1, layout.GetPalette().GetGrey(0.6), 0.03)
		paint.Line(rc, 0, 1, 1, 0, layout.GetPalette().GetGrey(0.6), 0.03)

	} else {
		rc.X += rc.W
		paint.RectRad(rc, cd2, cd2_over, cd2_down, 0, in_rounding)

		//I
		rc = rc.Cut(0.1)
		paint.Line(rc, 1.0/3, 0.9, 0.05, 2.0/3, cd, 0.03)
		paint.Line(rc, 1.0/3, 0.9, 0.95, 1.0/4, cd, 0.03)
	}

	//draw label
	if st.Label != "" {
		paint.Text(rectLabel.Cut(0.1), st.Label, "", cd_text, cd_text_over, cd_text_down, true, false, 0, 1)
	}

	return
}

func (st *Switch) Input(in LayoutInput, layout *Layout) {
	active := in.IsActive
	inside := in.IsInside && (active || !in.IsUse)
	clicked := in.IsUp && active && inside

	if clicked {
		*st.Value = !*st.Value
		if st.changed != nil {
			st.changed()
		}
		layout.RedrawThis()
	}
}
