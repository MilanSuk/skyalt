package main

import (
	"fmt"
	"image/color"
)

type Color struct {
	Type    string
	Cd      *color.RGBA
	Quality float64
	Tooltip string

	changed func()
}

func (layout *Layout) AddColor(x, y, w, h int, cd *color.RGBA) *Color {
	props := &Color{Type: "rainbow", Cd: cd, Quality: 0.7}
	layout._createDiv(x, y, w, h, "Color", nil, props.Draw, props.Input)
	return props
}

func (st *Color) Draw(rect Rect, layout *Layout) (paint LayoutPaint) {

	switch st.Type {
	case "rainbow":
		st._drawRainbow(&paint, rect)
	}

	return paint
}

func (st *Color) Input(in LayoutInput, layout *Layout) {

	NewCd := *st.Cd

	active := in.IsActive
	if active {
		x := _Color_clampFloat(float64(in.X/in.Rect.W), 0, 1)

		_, s, l := RGBtoHSL(*st.Cd)
		s = _Color_clampFloat(s, 0.1, 0.9)
		l = _Color_clampFloat(l, 0.1, 0.9)
		NewCd = HSLtoRGB(360*x, s, l, st.Cd.A)
	}

	if NewCd != *st.Cd {
		*st.Cd = NewCd

		if st.changed != nil {
			st.changed()
		}
		layout.RedrawThis()
	}
}

func (st *Color) _drawRainbow(paint *LayoutPaint, rect Rect) {
	//draw rainbow
	{
		rc := rect
		startX := rc.X
		endX := rc.X + rc.W
		rc.W = _Color_clampFloat(0.5*(1-st.Quality), 0.03, 0.5)
		for rc.X < endX {
			p := (rc.X - startX) / (endX - startX) //<0, 1>

			cd := HSLtoRGB(360*p, 1, 0.5, 255)
			rr := rc
			rr.W += 0.03
			paint.Rect(rr, cd, cd, cd, 0)
			rc.X += rc.W
		}
	}

	//selected position
	H, _, _ := RGBtoHSL(*st.Cd)
	p := float64(H) / 360
	paint.Line(rect, p, 0, p, 1, color.RGBA{0, 0, 0, 255}, 0.06)

	paint.Cursor("hand", rect)
	paint.Tooltip(st.Tooltip, rect)

	cq := rect
	cq.X += cq.W*p - 0.25
	cq.W = 0.5
	paint.TooltipEx(cq, fmt.Sprintf("Hue: %d", int(H)), true)
}

func _Color_clampFloat(v, min, max float64) float64 {
	if v < min {
		v = min
	}
	if v > max {
		v = max
	}
	return v
}
