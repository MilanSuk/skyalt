/*
Copyright 2024 Milan Suk

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
	"fmt"
	"image/color"
	"os"
	"strings"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

type WinFontFace struct {
	face        font.Face
	lastUseTick int64
}

func NewWinFontFace(prop *WinFontProps) *WinFontFace {

	var name string
	switch prop.weight {
	case 100:
		name = OsTrnString(!prop.italic, "Inter-Thin", "Inter-ThinItalic")
	case 200:
		name = OsTrnString(!prop.italic, "Inter-ExtraLight", "Inter-ExtraLightItalic")
	case 300:
		name = OsTrnString(!prop.italic, "Inter-Light", "Inter-LightItalic")
	case 400:
		name = OsTrnString(!prop.italic, "Inter-Regular", "Inter-Italic") //default
	case 500:
		name = OsTrnString(!prop.italic, "Inter-Medium", "Inter-MediumItalic")
	case 600:
		name = OsTrnString(!prop.italic, "Inter-SemiBold", "Inter-SemiBoldItalic")
	case 700:
		name = OsTrnString(!prop.italic, "Inter-Bold", "Inter-BoldItalic")
	case 800:
		name = OsTrnString(!prop.italic, "Inter-ExtraBold", "Inter-ExtraBoldItalic")
	case 900:
		name = OsTrnString(!prop.italic, "Inter-Black", "Inter-BlackItalic")
	}

	if name == "" {
		fmt.Printf("Unknown wieght %d\n", prop.weight)
		return nil
	}

	path := "resources/Inter/" + name + ".ttf"

	fl, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("ReadFile() failed: %v\n", err)
		return nil
	}

	ft, err := truetype.Parse(fl)
	if err != nil {
		fmt.Printf("truetype.Parse() failed: %v\n", err)
		return nil
	}

	face := truetype.NewFace(ft, &truetype.Options{Size: float64(prop.textH)}) //Hinting: font.HintingFull

	return &WinFontFace{face: face}
}
func (ff *WinFontFace) Destroy() {
	ff.face.Close()
}

func (ff *WinFontFace) UpdateTick() {
	ff.lastUseTick = OsTicks()
}
func (ff *WinFontFace) IsUsed() bool {
	return OsIsTicksIn(ff.lastUseTick, 30000) //30 sec
}

type WinFont struct {
	faces      [9]*WinFontFace
	faces_ital [9]*WinFontFace
}

func (ft *WinFont) Destroy() {
	for i, f := range ft.faces {
		if f != nil {
			f.Destroy()
			ft.faces[i] = nil
		}
	}
	for i, f := range ft.faces_ital {
		if f != nil {
			f.Destroy()
			ft.faces_ital[i] = nil
		}
	}
}

func (ft *WinFont) GetFace(prop *WinFontProps) *WinFontFace {

	var ret *WinFontFace

	w := (prop.weight / 100) - 1
	w = OsClamp(w, 0, 8)

	if prop.italic {
		if ft.faces_ital[w] == nil {
			ft.faces_ital[w] = NewWinFontFace(prop)
		}
		ret = ft.faces_ital[w]
	} else {
		if ft.faces[w] == nil {
			ft.faces[w] = NewWinFontFace(prop)
		}
		ret = ft.faces[w]
	}

	ret.UpdateTick()

	return ret
}

func (ft *WinFont) Maintenance() {
	for i := len(ft.faces) - 1; i >= 0; i-- {
		if ft.faces[i] != nil && !ft.faces[i].IsUsed() {
			ft.faces[i].Destroy()
			ft.faces[i] = nil
		}
	}

	for i := len(ft.faces_ital) - 1; i >= 0; i-- {
		if ft.faces_ital[i] != nil && !ft.faces_ital[i].IsUsed() {
			ft.faces_ital[i].Destroy()
			ft.faces_ital[i] = nil
		}
	}
}

type WinFontProps struct {
	textH int
	lineH int

	weight int
	italic bool

	switch_formating_when_edit bool
	formating                  bool

	fixed_width bool
}

func WinFontProps_GetDefaultTextH() float64 {
	return 0.37
}

func WinFontProps_GetDefaultLineH() float64 {
	return 0.6
}

// textH & lineH are in <0-1> range
func InitWinFontProps(weight int, textH, lineH float64, italic bool, formating bool, cell int) WinFontProps {
	if weight <= 0 {
		weight = 400
	}

	if textH <= 0 {
		textH = WinFontProps_GetDefaultTextH()
	}
	tPx := int(float64(cell) * textH)

	if lineH <= 0 {
		lineH = WinFontProps_GetDefaultLineH()
	}
	lPx := int(float64(cell) * lineH)

	return WinFontProps{weight: weight, textH: tPx, lineH: lPx, italic: italic, formating: formating}
}

func InitWinFontPropsDef(cell int) WinFontProps {
	return InitWinFontProps(0, 0, 0, false, true, cell)
}

func (a *WinFontProps) Cmp(b *WinFontProps) bool {
	return a.weight == b.weight &&
		a.textH == b.textH &&
		a.lineH == b.lineH &&
		a.italic == b.italic &&
		a.formating == b.formating
}

func (orig_prop *WinFontProps) processLetter(text string, prop *WinFontProps, frontCd *color.RGBA, skip *int) bool {

	if *skip > 0 {
		*skip -= 1
		return false
	}

	//new line = reset
	if strings.HasPrefix(text, "\n") {
		prop.weight = orig_prop.weight
		prop.italic = false
		prop.textH = orig_prop.textH
	}

	//bold
	if strings.HasPrefix(text, "<b>") {
		prop.weight = orig_prop.weight * 3 / 2
		*skip = 2
		return false
	}
	if strings.HasPrefix(text, "</b>") {
		prop.weight = orig_prop.weight
		*skip = 3
		return false
	}

	//italic
	if strings.HasPrefix(text, "<i>") {
		prop.italic = true
		*skip = 2
		return false
	}
	if strings.HasPrefix(text, "</i>") {
		prop.italic = false
		*skip = 3
		return false
	}

	//small
	if strings.HasPrefix(text, "<small>") {
		prop.textH = int(float64(orig_prop.textH) * 0.9)
		*skip = 6
		return false
	}
	if strings.HasPrefix(text, "</small>") {
		prop.textH = orig_prop.textH
		*skip = 7
		return false
	}

	//taller
	if strings.HasPrefix(text, "<h1>") {
		prop.textH = int(float64(orig_prop.textH) * 1.5)
		*skip = 3
		return false
	}
	if strings.HasPrefix(text, "</h1>") {
		prop.textH = orig_prop.textH
		*skip = 4
		return false
	}
	if strings.HasPrefix(text, "<h2>") {
		prop.textH = int(float64(orig_prop.textH) * 1.2)
		*skip = 3
		return false
	}
	if strings.HasPrefix(text, "</h2>") {
		prop.textH = orig_prop.textH
		*skip = 4
		return false
	}

	if strings.HasPrefix(text, "</rgba>") {
		*frontCd = color.RGBA{}
		*skip = 6
		return false
	}

	//color
	if strings.HasPrefix(text, "<rgba") {
		p := strings.IndexByte(text, '>')
		if p > 0 {
			text = text[5:p]

			frontCd.A = 0 //orig
			fmt.Sscanf(text, "%d,%d,%d,%d", &frontCd.R, &frontCd.G, &frontCd.B, &frontCd.A)

			*skip = p
		}
		return false
	}

	return true
}
