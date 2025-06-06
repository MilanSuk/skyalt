package main

type Checkbox struct {
	Label   string
	Tooltip string
	Value   *float64

	changed func()
}

func (layout *Layout) AddCheckbox(x, y, w, h int, label string, value *float64) *Checkbox {
	props := &Checkbox{Label: label, Value: value}
	layout._createDiv(x, y, w, h, "Checkbox", nil, props.Draw, props.Input)
	return props
}

func (st *Checkbox) Draw(rect Rect, layout *Layout) (paint LayoutPaint) {

	paint.Cursor("hand", rect)
	paint.Tooltip(st.Tooltip, rect)

	//colors
	P := layout.GetPalette().P
	B := layout.GetPalette().B
	onB := layout.GetPalette().OnB

	cd := P

	cd_over := cd
	cd_down := cd

	cd_text := onB
	cd_text_over := cd_text
	cd_text_down := cd_text

	cd_over = Color_Aprox(cd_over, B, 0.2)
	cd_down = Color_Aprox(cd_down, onB, 0.4)
	cd_text_over = Color_Aprox(cd_text_over, B, 0.2)
	cd_text_down = Color_Aprox(cd_text_down, onB, 0.4)

	//coord
	rc := rect
	rectLabel := rc
	{
		height := rc.H * 0.5
		if height > 0.6 {
			height = 0.6
		}

		width := height
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

	//draw checkbox
	paint.RectRad(rc, cd, cd_over, cd_down, 0.03, rc.H*layout.ui.router.sync.GetRounding())

	if *st.Value >= 1 {
		//check
		rc = rc.Cut(0.1)
		paint.Line(rc, 1.0/3, 0.9, 0.05, 2.0/3, cd, 0.05)
		paint.Line(rc, 1.0/3, 0.9, 0.95, 1.0/4, cd, 0.05)
	} else if *st.Value > 0 {
		//[-]
		rc = rc.Cut(0.1)
		paint.Line(rc, 0, 0.5, 1, 0.5, cd, 0.05)
	}

	//draw label
	if st.Label != "" {
		paint.Text(rectLabel.Cut(0.1), st.Label, "", cd_text, cd_text_over, cd_text_down, true, false, 0, 1)
	}

	return
}

func (st *Checkbox) Input(in LayoutInput, layout *Layout) {
	clicked := false
	active := in.IsActive
	inside := in.IsInside && (active || !in.IsUse)
	clicked = in.IsUp && active && inside

	if clicked {
		if *st.Value > 0 {
			*st.Value = 0.0
		} else {
			*st.Value = 1.0
		}
		if st.changed != nil {
			st.changed()
		}
		layout.RedrawThis()
	}
}
