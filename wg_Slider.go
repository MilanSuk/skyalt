package main

import (
	"fmt"
	"math"
	"strconv"
)

type Slider struct {
	Tooltip      string
	ValuePointer interface{} //*int, *float64

	Min  float64
	Max  float64
	Step float64

	Legend    bool
	DrawSteps bool

	changed func()
}

func (layout *Layout) AddSlider(x, y, w, h int, valuePointer interface{}, min, max, step float64) *Slider {
	props := &Slider{ValuePointer: valuePointer, Min: min, Max: max, Step: step}
	lay := layout._createDiv(x, y, w, h, "Slider", nil, props.Draw, props.Input)
	lay.fnGetLLMTip = props.getLLMTip
	return props
}
func (layout *Layout) AddSliderInt(x, y, w, h int, value *int, min, max, step float64) *Slider {
	val := float64(*value)
	props := layout.AddSlider(x, y, w, h, &val, min, max, step)
	props.changed = func() {
		*value = int(val)
	}
	return props
}

func (st *Slider) getLLMTip(layout *Layout) string {
	return fmt.Sprintf("Type: Slider. Value: %f. Tooltip: %s", st.getValue(), st.Tooltip)
}

func (st *Slider) Draw(rect Rect, layout *Layout) (paint LayoutPaint) {

	paint.Cursor("hand", rect)
	paint.Tooltip(st.Tooltip, rect)

	//colors
	B := layout.GetPalette().GetGrey(0.5)
	onB := layout.GetPalette().OnB
	cd := layout.GetPalette().P
	cd2 := Color_Aprox(cd, layout.GetPalette().B, 0.6) //cd.SetAlpha(100)

	cdThumb := cd
	cdThumb_over := cd
	cdThumb_down := cd
	cdThumb_over = Color_Aprox(cdThumb, B, 0.2)
	cdThumb_down = Color_Aprox(cdThumb, onB, 0.4)

	value := st.getValue()
	rad, coord, cqA, cqB := st._getCoords(rect, value)

	//steps
	if st.DrawSteps {
		sz := rad * 1.2
		for i := st.Min + st.Step; i < st.Max; i++ {
			t := (i - st.Min) / (st.Max - st.Min)
			rc := coord
			rc.X = coord.X + coord.W*t - sz/2
			rc.Y = coord.H/2 - sz/2
			rc.W = sz
			rc.H = sz

			cdd := cd
			if i >= value {
				cdd = cd2
			}

			paint.Circle(rc, cdd, cdd, cdd, 0)
		}
	}

	//track(2x lines)
	paint.Rect(cqA, cd, cd, cd, 0)
	paint.Rect(cqB, cd2, cd2, cd2, 0)

	//thumb(sphere)
	{
		var cqT Rect
		cqT.X = cqB.X - rad
		cqT.Y = coord.Y + coord.H/2 - rad
		cqT.W = rad * 2
		cqT.H = rad * 2
		paint.Circle(cqT, cdThumb, cdThumb_over, cdThumb_down, 0)

		//label
		str := strconv.FormatFloat(value, 'f', 2, 64)
		paint.TooltipEx(cqT.Cut(-rad/2), str, true)
	}

	//legend
	if st.Legend {
		frontCd := layout.GetPalette().GetGrey(0.2)

		rc := rect
		rc = rc.CutLeft(0.2)
		rc = rc.CutRight(0.2)
		rc.Y += 0.5
		paint.Text(rc, "<small>"+strconv.FormatFloat(st.Min, 'f', -1, 64), "", frontCd, frontCd, frontCd, false, false, 0, 0)

		paint.Text(rc, "<small>"+strconv.FormatFloat(st.Max, 'f', -1, 64), "", frontCd, frontCd, frontCd, false, false, 2, 0)

	}

	return
}

func (st *Slider) Input(in LayoutInput, layout *Layout) {
	active := in.IsActive
	inside := in.IsInside && (active || !in.IsActive)

	value := st.getValue()

	if active {
		_, coord, _, _ := st._getCoords(in.Rect, value)

		touch_x := (in.X - coord.X) / coord.W
		//clamp
		if touch_x < 0 {
			touch_x = 0
		}
		if touch_x > 1 {
			touch_x = 1
		}

		value = st.Min + (st.Max-st.Min)*touch_x
	}

	if !active && inside && in.Wheel != 0 && layout.findParentScroll() == nil {
		value += st.Step * float64(-in.Wheel)
	}

	//check & round
	{
		tt := 0.0
		if st.Step != 0 {
			tt = math.Round((value - st.Min) / st.Step)
		}
		value = st.Min + tt*st.Step

		//clamp
		if value < st.Min {
			value = st.Min
		}
		if value > st.Max {
			value = st.Max
		}
	}

	if st.setValue(value) {
		if st.changed != nil {
			st.changed()
		}
		layout.RedrawThis()
	}
}

func (st *Slider) _getCoords(rect Rect, value float64) (float64, Rect, Rect, Rect) {

	rad := 0.2

	coord := rect
	coord = coord.CutLeft(rad)
	coord = coord.CutRight(rad)

	t := 0.0
	if st.Max != st.Min {
		t = (value - st.Min) / (st.Max - st.Min)
	}
	cqA := coord
	cqB := coord
	cqA.W = cqA.W * t
	cqB.X += cqA.W
	cqB.W -= cqA.W

	h_rad := rad / 2
	cqA.Y = cqA.H/2 - h_rad/2
	cqB.Y = cqB.H/2 - h_rad/2
	cqA.H = h_rad
	cqB.H = h_rad

	return rad, coord, cqA, cqB
}

func (st *Slider) setValue(value float64) bool {
	diff := false
	switch v := st.ValuePointer.(type) {
	case *string:
		val := strconv.FormatFloat(value, 'f', -1, 64)
		diff = (*v != val)
		*v = val

	case *int:
		diff = (*v != int(value))
		*v = int(value)
	case *int64:
		diff = (*v != int64(value))
		*v = int64(value)
	case *int32: //also rune
		diff = (*v != int32(value))
		*v = int32(value)
	case *int16:
		diff = (*v != int16(value))
		*v = int16(value)
	case *int8:
		diff = (*v != int8(value))
		*v = int8(value)

	case *uint:
		diff = (*v != uint(value))
		*v = uint(value)
	case *uint64:
		diff = (*v != uint64(value))
		*v = uint64(value)
	case *uint32: //also rune
		diff = (*v != uint32(value))
		*v = uint32(value)
	case *uint16:
		diff = (*v != uint16(value))
		*v = uint16(value)
	case *uint8:
		diff = (*v != uint8(value))
		*v = uint8(value)

	case *float64:
		diff = (*v != float64(value))
		*v = float64(value)
	case *float32:
		diff = (*v != float32(value))
		*v = float32(value)

	case *bool:
		val := (value != 0)
		diff = (*v != val)
		*v = val
	}

	return diff
}

func (st *Slider) getValue() float64 {
	switch v := st.ValuePointer.(type) {
	case *string:
		val, _ := strconv.ParseFloat(*v, 64)
		return val

	case *int:
		return float64(*v)
	case *int64:
		return float64(*v)
	case *int32: //also rune
		return float64(*v)
	case *int16:
		return float64(*v)
	case *int8:
		return float64(*v)

	case *uint:
		return float64(*v)
	case *uint64:
		return float64(*v)
	case *uint32: //also rune
		return float64(*v)
	case *uint16:
		return float64(*v)
	case *uint8:
		return float64(*v)

	case *float64:
		return *v
	case *float32:
		return float64(*v)

	case *bool:
		if *v {
			return 1
		} else {
			return 0
		}
	}

	return 0
}
