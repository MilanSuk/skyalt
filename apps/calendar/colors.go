/*
Copyright 2023 Milan Suk

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this db except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
)

type ColorTranslations struct {
	RED   string
	GREEN string
	BLUE  string

	HUE        string
	SATURATION string
	LIGHTNESS  string
}

var trnsColors ColorTranslations

var g_ButtonCd _SA_Style

func OpenColors() {

	json.Unmarshal(SA_File("translations_json:app:resources/translations_colors.json"), &trnsColors)

	g_ButtonCd = styles.Button
	g_ButtonCd.ContentColor(SACd{}) //transparent
}

func ColorPickerHueRainbow(cd *SACd) bool {

	n := SA_DivInfoGet("layoutWidth") * 5 //5 lines in 1 cell

	//draw rainbow
	st := 1 / n
	last_i := float64(0)
	for i := st; i < 1+st; i += st {
		//p = i / n
		rgb := HSL{H: int(360 * i), S: 1, L: 0.5}.HSLtoRGB()

		SAPaint_Rect(last_i, 0, (i-last_i)+0.06, 1, 0, rgb, 0)
		last_i = i
	}

	//selected position
	cdd := cd.RGBtoHSL()
	p := float64(cdd.H) / 360
	SAPaint_Line(p, 0, p, 1, SA_ThemeBlack(), 0.06)

	//picker
	changed := false
	if SA_DivInfoGet("touchInside") > 0 {
		SAPaint_Cursor("hand")

		if SA_DivInfoGet("touchActive") > 0 {
			x := SA_DivInfoGet("touchX")
			x = float64(Clamp(float32(x), 0, 1))

			*cd = HSL{H: int(360 * x), S: 0.7, L: 0.5}.HSLtoRGB()
			changed = true
		}
	}
	return changed
}

func ButtonColor(x, y, w, h int, value string, tooltip string, cd SACd, rowSize float64) bool {
	var click bool

	SA_DivStart(x, y, w, h)
	{
		SAPaint_Rect(0, 0, 1, 1, 0.06, cd, 0) //background

		SA_ColMax(0, 100)
		if rowSize > 0 {
			SA_Row(0, float64(rowSize))
		} else {
			SA_RowMax(0, 100)
		}
		click = SA_ButtonStyle(value, &g_ButtonCd).Tooltip(tooltip).Show(0, 0, 1, 1).click //transparent, so background is seen
	}
	SA_DivEnd()

	return click
}

func ButtonColorPicker(cd *SACd, description string, description_width float64, dialogName string) bool {
	origCd := *cd
	cd.A = 255

	SA_RowMax(0, 100)

	main_x := 0
	if len(description) > 0 {
		SA_ColMax(0, description_width)
		SA_ColMax(1, 100)
		SA_Text(description).Show(0, 0, 1, 1)
		main_x = 1
	} else {
		SA_ColMax(0, 100)
	}

	if ButtonColor(main_x, 0, 1, 1, "", "", *cd, 0) {
		SA_DialogOpen(dialogName, 1)
	}

	dialogOpen := SA_DialogStart(dialogName)
	if dialogOpen {
		ColorPicker(cd)
		SA_DialogEnd()
	}

	return !origCd.Cmp(*cd)
}

func ColorPicker(cd *SACd) {

	SA_ColMax(0, 7)
	SA_ColMax(1, 7)

	//final color
	SA_DivStart(0, 0, 2, 1)
	SAPaint_Rect(0, 0, 1, 1, 0.06, *cd, 0)
	SA_DivEnd()

	//RGB
	SA_DivStart(0, 1, 1, 3)
	{
		SA_ColMax(0, 100)

		r := float64(cd.R)
		g := float64(cd.G)
		b := float64(cd.B)
		if SA_Slider(&r).Min(0).Max(255).Jump(1).ShowDescription(0, 0, 1, 1, trnsColors.RED, 3, nil).changed {
			cd.R = uint8(r)
		}
		if SA_Slider(&g).Min(0).Max(255).Jump(1).ShowDescription(0, 1, 1, 1, trnsColors.GREEN, 3, nil).changed {
			cd.G = uint8(g)
		}
		if SA_Slider(&b).Min(0).Max(255).Jump(1).ShowDescription(0, 2, 1, 1, trnsColors.BLUE, 3, nil).changed {
			cd.B = uint8(b)
		}
	}
	SA_DivEnd()

	//HSL
	SA_DivStart(1, 1, 1, 3)
	{
		SA_ColMax(0, 100)

		hsl := cd.RGBtoHSL()
		h := float64(hsl.H)
		s := float64(hsl.S)
		l := float64(hsl.L)
		changed := false
		if SA_Slider(&h).Min(0).Max(360).Jump(1).ShowDescription(0, 0, 1, 1, trnsColors.HUE, 3, nil).changed { //decription Hue ...
			changed = true
		}
		if SA_Slider(&s).Min(0).Max(1).Jump(0.01).ShowDescription(0, 1, 1, 1, trnsColors.SATURATION, 3, nil).changed {
			changed = true
		}
		if SA_Slider(&l).Min(0).Max(1).Jump(0.01).ShowDescription(0, 2, 1, 1, trnsColors.LIGHTNESS, 3, nil).changed {
			changed = true
		}
		if changed {
			*cd = HSL{H: int(h), S: float32(s), L: float32(l)}.HSLtoRGB()
		}
	}
	SA_DivEnd()

	//rainbow
	SA_DivStart(0, 5, 2, 1)
	ColorPickerHueRainbow(cd)
	SA_DivEnd()

	//pre-build
	SA_DivStart(0, 6, 2, 1)
	{
		for i := 0; i < 12; i++ {
			SA_ColMax(i, 2)
		}

		//first 8
		for i := 0; i < 8; i++ {
			cdd := HSL{H: int(360 * float32(i) / 8), S: 0.7, L: 0.5}.HSLtoRGB()
			if ButtonColor(i, 0, 1, 1, "", "", cdd, 0) {
				*cd = cdd
			}
		}

		//other 4
		for i := 0; i < 4; i++ {
			cdd := HSL{H: 0, S: 0, L: float32(i) / 4}.HSLtoRGB()
			if ButtonColor(8+i, 0, 1, 1, "", "", cdd, 0) {
				*cd = cdd
			}
		}
	}
	SA_DivEnd()
}

type HSL struct {
	H int
	S float32
	L float32
}

func _hueToRGB(v1, v2, vH float32) float32 {
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

func INTtoRGB(v uint32) SACd {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)

	return SACd{R: b[0], G: b[1], B: b[2], A: b[3]}
}
func (rgb SACd) RGBtoINT() uint32 {
	b := []byte{rgb.R, rgb.G, rgb.B, rgb.A}
	return binary.LittleEndian.Uint32(b)
}

func (hsl HSL) HSLtoRGB() SACd {
	cd := SACd{A: 255}

	if hsl.S == 0 {
		ll := hsl.L * 255
		ll = Clamp(ll, 0, 255)

		cd.R = uint8(ll)
		cd.G = uint8(ll)
		cd.B = uint8(ll)
	} else {
		var v2 float32
		if hsl.L < 0.5 {
			v2 = (hsl.L * (1 + hsl.S))
		} else {
			v2 = ((hsl.L + hsl.S) - (hsl.L * hsl.S))
		}
		v1 := 2*hsl.L - v2

		hue := float32(hsl.H) / 360
		cd.R = uint8(255 * _hueToRGB(v1, v2, hue+(1.0/3)))
		cd.G = uint8(255 * _hueToRGB(v1, v2, hue))
		cd.B = uint8(255 * _hueToRGB(v1, v2, hue-(1.0/3)))
	}

	return cd
}

func (cd SACd) RGBtoHSL() HSL {
	var hsl HSL

	r := float32(cd.R) / 255
	g := float32(cd.G) / 255
	b := float32(cd.B) / 255

	min := Min(Min(r, g), b)
	max := Max(Max(r, g), b)
	delta := max - min

	hsl.L = (max + min) / 2

	if delta == 0 {
		hsl.H = 0
		hsl.S = 0
	} else {
		if hsl.L <= 0.5 {
			hsl.S = delta / (max + min)
		} else {
			hsl.S = delta / (2 - max - min)
		}

		var hue float32
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

		hsl.H = int(hue * 360)
	}

	return hsl
}

func HEXtoRGBwithCheck(hex string, defaultCd SACd) SACd {
	if len(hex) == 6 || (len(hex) == 7 && hex[0] == '#') {
		return HEXtoRGB(hex)
	}
	return defaultCd
}

func HEXtoRGB(hex string) SACd {
	cd := SACd{A: 255}

	if len(hex) == 0 {
		return cd
	}

	if hex[0] == '#' {
		hex = hex[1:] //skip
	}

	if len(hex) < 2 {
		return cd
	}
	r, _ := strconv.ParseInt(hex[:2], 16, 16)
	cd.R = uint8(r)
	hex = hex[2:]

	if len(hex) < 2 {
		return cd
	}
	g, _ := strconv.ParseInt(hex[:2], 16, 16)
	cd.G = uint8(g)
	hex = hex[2:]

	if len(hex) < 2 {
		return cd
	}
	b, _ := strconv.ParseInt(hex[:2], 16, 16)
	cd.B = uint8(b)

	return cd
}

func (cd SACd) RGBtoHEX() string {
	return fmt.Sprintf("#%02x%02x%02x", cd.R, cd.G, cd.B)
}

func Max(x, y float32) float32 {
	if x < y {
		return y
	}
	return x
}
func Min(x, y float32) float32 {
	if x > y {
		return y
	}
	return x
}
func Clamp(v, min, max float32) float32 {
	return Min(Max(v, min), max)
}
