package main

type Divider struct {
	Horizontal bool
	Label      string
	Width      float64
	Margin     float64
}

func (layout *Layout) AddDivider(x, y, w, h int, horizontal bool) *Divider {
	props := &Divider{Horizontal: horizontal, Width: 0.03, Margin: 0.1}
	layout._createDiv(x, y, w, h, "Divider", props.Build, nil, nil)
	return props
}

func (st *Divider) Build(layout *Layout) {

	if st.Label == "" {
		layout.fnDraw = func(rect Rect, layout *Layout) (paint LayoutPaint) {
			cd := layout.GetPalette().GetGrey(0.6)

			if st.Horizontal {
				paint.Line(rect.CutLeft(st.Margin).CutRight(st.Margin), 0, 0.5, 1, 0.5, cd, st.Width)
			} else {
				paint.Line(rect.CutTop(st.Margin).CutBottom(st.Margin), 0.5, 0, 0.5, 1, cd, st.Width)
			}
			return
		}
	} else {
		if st.Horizontal {
			layout.SetRow(0, 0, 100)

			layout.SetColumn(0, 0, 100)
			layout.SetColumnFromSub(1, 1, 100, true)
			layout.SetColumn(2, 0, 100)
			layout.AddDivider(0, 0, 1, 1, true)
			layout.AddText(1, 0, 1, 1, st.Label)
			layout.AddDivider(2, 0, 1, 1, true)
		} else {
			layout.SetColumn(0, 0, 100)

			layout.SetRow(0, 0, 100)
			layout.SetRowFromSub(1, 1, 100, true)
			layout.SetRow(2, 0, 100)
			layout.AddDivider(0, 0, 1, 1, true)
			layout.AddText(0, 1, 1, 1, st.Label) //rotate ....
			layout.AddDivider(0, 2, 1, 1, true)
		}
	}

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
