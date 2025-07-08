package main

import (
	"image/color"
)

type ColorPicker struct {
	Cd      *color.RGBA
	changed func()
}

func (layout *Layout) AddColorPicker(x, y, w, h int, cd *color.RGBA) *ColorPicker {
	props := &ColorPicker{Cd: cd}
	layout._createDiv(x, y, w, h, "ColorPicker", props.Build, nil, nil)
	return props
}

func (st *ColorPicker) Build(layout *Layout) {

	layout.SetColumn(0, 5, 100)
	layout.SetColumn(1, 5, 100)

	layout.SetRow(0, 1.5, 1.5)
	layout.SetRow(4, 1.5, 1.5)

	//Final
	Final := layout.AddImage(0, 0, 2, 1, "", nil)
	Final.Cd = *st.Cd
	Final.Draw_border = true
	Final.Margin = 0.1

	//properties
	r, g, b, a := float64(st.Cd.R), float64(st.Cd.G), float64(st.Cd.B), float64(st.Cd.A)
	h, s, l := RGBtoHSL(*st.Cd)
	R := st._addValue(layout, 0, 1, 1, 1, "Red", 0, &r, 0, 255, 1)
	G := st._addValue(layout, 0, 2, 1, 1, "Green", 0, &g, 0, 255, 1)
	B := st._addValue(layout, 0, 3, 1, 1, "Blue", 0, &b, 0, 255, 1)
	A := st._addValue(layout, 0, 5, 2, 1, "Alpha", 0, &a, 0, 255, 1)
	H := st._addValue(layout, 1, 1, 1, 1, "Hue", 0, &h, 0, 359, 1)
	S := st._addValue(layout, 1, 2, 1, 1, "Saturation", 2, &s, 0, 1, 0.01)
	L := st._addValue(layout, 1, 3, 1, 1, "Lightness", 2, &l, 0, 1, 0.01)

	R.changed = func() {
		st.Cd.R = uint8(r)
		if st.changed != nil {
			st.changed()
		}
	}
	G.changed = func() {
		st.Cd.G = uint8(g)
		if st.changed != nil {
			st.changed()
		}
	}
	B.changed = func() {
		st.Cd.B = uint8(b)
		if st.changed != nil {
			st.changed()
		}
	}
	A.changed = func() {
		st.Cd.A = uint8(a)
		if st.changed != nil {
			st.changed()
		}
	}

	H.changed = func() {
		*st.Cd = HSLtoRGB(h, s, l, st.Cd.A)
		if st.changed != nil {
			st.changed()
		}
	}
	S.changed = func() {
		*st.Cd = HSLtoRGB(h, s, l, st.Cd.A)
		if st.changed != nil {
			st.changed()
		}
	}
	L.changed = func() {
		*st.Cd = HSLtoRGB(h, s, l, st.Cd.A)
		if st.changed != nil {
			st.changed()
		}
	}

	Rainbow := layout.AddColor(0, 4, 2, 1, st.Cd)
	Rainbow.changed = func() {
		if st.changed != nil {
			st.changed()
		}
	}
}

func (st *ColorPicker) _addValue(layout *Layout, x, y, w, h int, Description string, Prec int, Value *float64, Min, Max, Step float64) *SliderEdit {
	it := layout.AddSliderEdit(x, y, w, h, Value, Min, Max, Step)
	it.ValuePointerPrec = Prec
	it.Description_width = 2
	it.Slider_width = 100
	it.Edit_width = 2
	it.Description = Description
	return it
}

func _hueToRGB1(v1, v2, vH float64) float64 {
	if vH < 0 {
		vH++
	}
	if vH > 1 {
		vH--
	}

	if (6 * vH) < 1 {
		return v1 + (v2-v1)*6*vH
	} else if (2 * vH) < 1 {
		return v2
	} else if (3 * vH) < 2 {
		return v1 + (v2-v1)*((2.0/3)-vH)*6
	}

	return v1
}

func _maxFloat(x, y float64) float64 {
	if x < y {
		return y
	}
	return x
}
func _minFloat(x, y float64) float64 {
	if x > y {
		return y
	}
	return x
}
func _clampFloat(v, min, max float64) float64 {
	return _minFloat(_maxFloat(v, min), max)
}

func HSLtoRGB(H, S, L float64, alpha uint8) color.RGBA {
	var dst color.RGBA
	dst.A = alpha

	if S == 0 {
		ll := L * 255
		ll = float64(_clampFloat(float64(ll), 0, 255))

		dst.R = uint8(ll)
		dst.G = uint8(ll)
		dst.B = uint8(ll)
	} else {
		var v2 float64
		if L < 0.5 {
			v2 = (L * (1 + S))
		} else {
			v2 = ((L + S) - (L * S))
		}
		v1 := 2*L - v2

		hue := float64(H) / 360
		dst.R = uint8(255 * _hueToRGB1(v1, v2, hue+(1.0/3)))
		dst.G = uint8(255 * _hueToRGB1(v1, v2, hue))
		dst.B = uint8(255 * _hueToRGB1(v1, v2, hue-(1.0/3)))
	}

	return dst
}

func RGBtoHSL(src color.RGBA) (float64, float64, float64) {
	var H, S, L float64

	r := float64(src.R) / 255
	g := float64(src.G) / 255
	b := float64(src.B) / 255

	min := _minFloat(_minFloat(r, g), b)
	max := _maxFloat(_maxFloat(r, g), b)
	delta := max - min

	L = (max + min) / 2

	if delta == 0 {
		H = 0
		S = 0
	} else {
		if L <= 0.5 {
			S = delta / (max + min)
		} else {
			S = delta / (2 - max - min)
		}

		var hue float64
		if r == max {
			hue = ((g - b) / 6) / delta
		} else if g == max {
			hue = (1.0 / 3) + ((b-r)/6)/delta
		} else {
			hue = (2.0 / 3) + ((r-g)/6)/delta
		}

		if hue < 0 {
			hue += 1
		}
		if hue > 1 {
			hue -= 1
		}

		H = float64(int(hue * 360))
	}

	return H, S, L
}
