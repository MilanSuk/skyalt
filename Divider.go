package main

type Divider struct {
	Horizontal bool
	Width      float64
}

func (layout *Layout) AddDivider(x, y, w, h int, horizontal bool) *Divider {
	props := &Divider{Horizontal: horizontal}
	layout._createDiv(x, y, w, h, "Divider", nil, props.Draw, nil)
	return props
}

func (st *Divider) Draw(rect Rect, layout *Layout) (paint LayoutPaint) {
	cd := Paint_GetPalette().GetGrey(0.5)
	cut := 0.1
	if st.Horizontal {
		paint.Line(rect.CutLeft(cut).CutRight(cut), 0, 0.5, 1, 0.5, cd, st.Width)
	} else {
		paint.Line(rect.CutTop(cut).CutBottom(cut), 0.5, 0, 0.5, 1, cd, st.Width)
	}
	return
}
