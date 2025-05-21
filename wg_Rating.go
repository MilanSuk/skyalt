package main

import (
	"image/color"
	"strconv"
)

type Rating struct {
	Value *float64
	Max   float64

	changed func()
}

func (layout *Layout) AddRating(x, y, w, h int, value *float64, max float64) *Rating {
	props := &Rating{Value: value, Max: max}
	layout._createDiv(x, y, w, h, "Rating", props.Build, nil, nil)
	return props
}

func (st *Rating) Build(layout *Layout) {
	layout.scrollH.Narrow = true
	layout.SetColumn(0, st.Max, st.Max) //show scroll
}

func (st *Rating) Draw(rect Rect, layout *Layout) (paint LayoutPaint) {

	paint.Cursor("hand", rect)

	rect.W = rect.W / float64(int(st.Max))
	if rect.W < 1 {
		rect.W = 1
	}

	cd := color.RGBA{0, 0, 0, 255}

	for i := 0; i < int(st.Max); i++ {
		icon := "resources/star_full.png"
		if i >= int(*st.Value) {
			icon = "resources/star_empty.png"
		}

		paint.File(rect, InitWinImagePath_file(icon), cd, cd, cd, 1, 1)

		rect.X += rect.W
	}

	//tooltip
	{
		rad := 0.2
		rc := rect
		rc.X = *st.Value - rad/2
		rc.W = rad
		paint.TooltipEx(rc, strconv.FormatFloat(*st.Value, 'f', -1, 64), true)
	}
	return
}

func (st *Rating) Input(in LayoutInput, layout *Layout) {
	active := in.IsActive

	val := *st.Value
	changed := false
	if active {
		rc := in.Rect
		touch_x := (in.X - rc.X) / rc.W
		if touch_x < 0 {
			touch_x = 0
		}
		if touch_x > 1 {
			touch_x = 1
		}

		vv := touch_x * st.Max
		if vv < 0.3 {
			val = 0 //round first to zero
		} else {
			val = float64(int(vv)) + 1 //roundDown(vv) + 1
		}
		changed = true
	}

	//check
	if val < 0 {
		val = 0
	}
	if val > st.Max {
		val = st.Max
	}

	if changed && *st.Value != val {
		*st.Value = val
		if st.changed != nil {
			st.changed()
		}
		layout.RedrawThis()
	}
}
