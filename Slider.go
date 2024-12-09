package main

import (
	"math"
	"strconv"
)

func (st *Slider) _getCoords(rect Rect) (float64, Rect, Rect, Rect) {

	rad := 0.2

	coord := rect
	coord = coord.CutLeft(rad)
	coord = coord.CutRight(rad)

	t := 0.0
	if st.Max != st.Min {
		t = (*st.Value - st.Min) / (st.Max - st.Min)
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

func (st *Slider) Draw(rect Rect) {
	st.lock.Lock()
	defer st.lock.Unlock()

	layout := st.layout

	layout.Paint_cursor("hand", rect)

	//colors
	S := layout.GetPalette().S
	onS := layout.GetPalette().OnS
	cd := layout.GetPalette().P
	cd2 := Color_Aprox(cd, layout.GetPalette().B, 0.6) //cd.SetAlpha(100)

	cdThumb := cd
	cdThumb_over := cd
	cdThumb_down := cd
	cdThumb_over = Color_Aprox(cdThumb, S, 0.2)
	cdThumb_down = Color_Aprox(cdThumb, onS, 0.4)

	rad, coord, cqA, cqB := st._getCoords(rect)

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
			if st.Value != nil && i >= *st.Value {
				cdd = cd2
			}

			layout.Paint_circle(rc, cdd, cdd, cdd, 0)
		}
	}

	//track(2x lines)
	layout.Paint_rect(cqA, cd, cd, cd, 0)
	layout.Paint_rect(cqB, cd2, cd2, cd2, 0)

	//thumb(sphere)
	{
		var cqT Rect
		cqT.X = cqB.X - rad
		cqT.Y = coord.Y + coord.H/2 - rad
		cqT.W = rad * 2
		cqT.H = rad * 2
		layout.Paint_circle(cqT, cdThumb, cdThumb_over, cdThumb_down, 0)
	}

	//legend
	if st.Legend {
		frontCd := layout.GetPalette().GetGrey(0.2)

		rc := rect
		rc = rc.CutLeft(0.2)
		rc = rc.CutRight(0.2)
		rc.Y += 0.5
		layout.Paint_text(rc, "<small>"+strconv.FormatFloat(st.Min, 'f', -1, 64), "", frontCd, frontCd, frontCd, false, false, 0, 0, true, false, false, 0)

		layout.Paint_text(rc, "<small>"+strconv.FormatFloat(st.Max, 'f', -1, 64), "", frontCd, frontCd, frontCd, false, false, 2, 0, true, false, false, 0)

	}

	//label
	if st.Value != nil {
		Value := *st.Value
		cqB.Y -= rad //move up
		cqB.W = rad * 2
		cqB.H = rad * 2

		str := strconv.FormatFloat(Value, 'f', 2, 64)
		layout.Paint_tooltipEx(cqB, str, true)
	}
}

func (st *Slider) Input(in LayoutInput) {
	st.lock.Lock()
	defer st.lock.Unlock()

	if st.Value == nil {
		return
	}

	active := in.IsActive
	inside := in.IsInside && (active || !in.IsActive)

	val := *st.Value
	changed := false
	if active {
		_, coord, _, _ := st._getCoords(in.Rect)

		touch_x := (in.X - coord.X) / coord.W
		//clamp
		if touch_x < 0 {
			touch_x = 0
		}
		if touch_x > 1 {
			touch_x = 1
		}

		val = st.Min + (st.Max-st.Min)*touch_x

		changed = true
	}

	if !active && inside && in.Wheel != 0 {
		val += st.Step * float64(-in.Wheel)
		changed = true
	}

	//check & round
	{
		tt := 0.0
		if st.Step != 0 {
			tt = math.Round((val - st.Min) / st.Step)
		}
		val = st.Min + tt*st.Step

		//clamp
		if val < st.Min {
			val = st.Min
		}
		if val > st.Max {
			val = st.Max
		}
	}

	if changed && *st.Value != val {
		*st.Value = val
		if st.changed != nil {
			st.changed()
		}
	}
}
