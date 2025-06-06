package main

type ProgressBar struct {
	Value   float64
	Height  float64 //cells
	Tooltip bool
}

func (layout *Layout) AddProgressBar(x, y, w, h int, value float64) *ProgressBar {
	props := &ProgressBar{Value: value, Height: 0.9}
	layout._createDiv(x, y, w, h, "ProgressBar", props.Build, props.Draw, nil)
	return props
}

func (st *ProgressBar) Build(layout *Layout) {

	layout.scrollH.Narrow = true
}

func (st *ProgressBar) Draw(rect Rect, layout *Layout) (paint LayoutPaint) {
	cdDone := layout.GetPalette().P
	cdRest := cdDone
	cdRest.A = 100

	margin_x := 0.03
	margin_y := (rect.H - st.Height) / 2
	if margin_y < 0 {
		margin_y = 0
	}
	rect = rect.CutLeft(margin_x)
	rect = rect.CutRight(margin_x)
	rect = rect.CutTop(margin_y)
	rect = rect.CutBottom(margin_y)

	paint.RectRad(rect, cdRest, cdRest, cdRest, 0, layout.ui.router.sync.GetRounding()) //100%

	rect.W *= st.Value
	paint.RectRad(rect, cdDone, cdDone, cdDone, 0, layout.ui.router.sync.GetRounding()) //done

	return
}
