package main

type Divider struct {
	Horizontal bool
	Width      float64
	Margin     float64
}

func (layout *Layout) AddDivider(x, y, w, h int, horizontal bool) *Divider {
	props := &Divider{Horizontal: horizontal, Width: 0.03, Margin: 0.1}
	layout._createDiv(x, y, w, h, "Divider", nil, props.Draw, nil)
	return props
}

func (st *Divider) Draw(rect Rect, layout *Layout) (paint LayoutPaint) {
	cd := layout.GetPalette().GetGrey(0.6)

	if st.Horizontal {
		paint.Line(rect.CutLeft(st.Margin).CutRight(st.Margin), 0, 0.5, 1, 0.5, cd, st.Width)
	} else {
		paint.Line(rect.CutTop(st.Margin).CutBottom(st.Margin), 0.5, 0, 0.5, 1, cd, st.Width)
	}
	return
}
