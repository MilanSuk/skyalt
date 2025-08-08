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
	"image/color"
	"strings"
)

type LayoutPromptColor struct {
	Label string
	Cd    color.RGBA
}

var g_prompt_colors = []LayoutPromptColor{
	{Label: "red", Cd: color.RGBA{255, 0, 0, 255}},
	{Label: "green", Cd: color.RGBA{0, 255, 0, 255}},
	{Label: "blue", Cd: color.RGBA{0, 0, 255, 255}},

	{Label: "orange", Cd: color.RGBA{255, 165, 0, 255}},
	{Label: "pink", Cd: color.RGBA{255, 192, 203, 255}},
	{Label: "yellow", Cd: color.RGBA{200, 200, 0, 255}},

	{Label: "aqua", Cd: color.RGBA{0, 255, 255, 255}},
	{Label: "fuchsia", Cd: color.RGBA{255, 0, 255, 255}},
	{Label: "olive", Cd: color.RGBA{128, 128, 0, 255}},
	{Label: "teal", Cd: color.RGBA{0, 128, 128, 255}},
	{Label: "purple", Cd: color.RGBA{128, 0, 128, 255}},
	{Label: "navy", Cd: color.RGBA{0, 0, 128, 255}},
	{Label: "marron", Cd: color.RGBA{128, 0, 0, 255}},
	{Label: "lime", Cd: color.RGBA{0, 255, 0, 255}},
	{Label: "brown", Cd: color.RGBA{165, 42, 42, 255}},
	{Label: "grey", Cd: color.RGBA{128, 128, 128, 255}},
}

func Layout3_Get_prompt_color(i int) LayoutPromptColor {
	return g_prompt_colors[i%len(g_prompt_colors)]
}

type LayoutPick struct {
	//Cd     LayoutPromptColor
	LLMTip string
	Points []OsV2
}

func (a *LayoutPick) Cmp(b *LayoutPick) bool {
	return a.LLMTip == b.LLMTip
}

type UiSelection struct {
	active *LayoutPick
}

func UiSelection_Thick(ui *Ui) int {
	return ui.CellWidth(0.3)
}

func (s *UiSelection) Draw(buff *WinPaintBuff, ui *Ui) {
	if s.active != nil {
		backupDepth := buff.depth
		buff.depth = 900

		buff.AddCrop(ui.mainLayout.CropWithScroll())

		buff.AddBrush(OsV2{}, s.active.Points, ui.GetPalette().P, UiSelection_Thick(ui), true)

		buff.depth = backupDepth
	}

}

func (s *UiSelection) UpdateBrush(ui *Ui) *LayoutPick {

	//start
	if ui.GetWin().io.Keys.Ctrl && !ui.GetWin().io.Keys.Shift {
		if ui.GetWin().io.Touch.Start {
			//s.active = &LayoutPick{Cd: Layout3_Get_prompt_color(ui.mainLayout.numBrushes())}
			s.active = &LayoutPick{}
		}
	}

	if s.active == nil {
		return nil
	}

	//add new point
	{
		pos := ui.GetWin().io.Touch.Pos
		if len(s.active.Points) == 0 || !s.active.Points[len(s.active.Points)-1].Cmp(pos) {
			s.active.Points = append(s.active.Points, pos)
		}
	}

	//end
	if ui.GetWin().io.Touch.End {

		cq := s.getRect() //need s.active ! nil
		ret := s.active
		s.active = nil

		if len(ret.Points) > 0 {

			cq = cq.Crop(ui.Cell() / 3)
			cq.Size = cq.Size.Max(OsV2{1, 1})
			cqArea := float64(cq.Area())
			if cqArea > 0 {

				min_area := 0 //pixel ....

				var tip strings.Builder
				ui.mainLayout.buildTips(cq, cqArea, min_area, 0, &tip)
				ret.LLMTip = _UiText_RemoveFormating(tip.String())

				return ret
			}
		}
	}

	return nil
}

func (s *UiSelection) getRect() OsV4 {
	if s.active == nil {
		return OsV4{}
	}

	points := s.active.Points

	if len(points) == 0 {
		return OsV4{}
	}

	min := points[0]
	max := points[0]

	for _, pt := range points {
		min.X = OsMin(min.X, pt.X)
		min.Y = OsMin(min.Y, pt.Y)
		max.X = OsMax(max.X, pt.X)
		max.Y = OsMax(max.Y, pt.Y)
	}
	return InitOsV4ab(min, max)
}

func (layout *Layout) buildTips(cq OsV4, cqArea float64, min_area int, depth int, outStr *strings.Builder) {

	if layout.touch { //layout.touchDia
		inArea := layout.crop.GetIntersect(cq).Area()

		if inArea > min_area {
			var tp string
			if layout.fnGetLLMTip != nil {
				tp = layout.fnGetLLMTip(layout)
			}
			if tp != "" {
				if outStr.Len() > 0 {
					outStr.WriteString("\n")
				}
				for range depth {
					outStr.WriteString("\t")
				}
				if depth > 0 {
					outStr.WriteString("- ")
				}

				//add tip
				outStr.WriteString(tp)

				depth++
			}
		}
	}

	for _, it := range layout.childs {
		if it.IsShown() {
			it.buildTips(cq, cqArea, min_area, depth, outStr)
		}
	}

	if layout.dialog != nil {
		layout.dialog.buildTips(cq, cqArea, min_area, 0, outStr)
	}
}

/*func (layout *Layout) findSelection(cq OsV4, cqArea float64, min_area int) *Layout {
	if !layout.touch {
		return nil
	}

	if layout.touchDia {
		inArea := layout.crop.GetIntersect(cq).Area()

		for _, it := range layout.childs {
			if it.IsShown() {
				foundLay := it.findSelection(cq, cqArea, min_area)
				if foundLay != nil {
					return foundLay
				}
			}
		}

		if inArea > min_area {
			var tip string
			if layout.fnGetLLMTip != nil {
				tip = layout.fnGetLLMTip(layout)
			}
			if tip != "" {
				return layout
			}

		}
	}

	if layout.dialog != nil {
		foundLay := layout.dialog.findSelection(cq, cqArea, min_area)
		if foundLay != nil {
			return foundLay
		}
	}

	return nil
}*/
