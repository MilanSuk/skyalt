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
)

type UiSelectionBrush struct {
	Points []OsV2
	Cd     LayoutPromptColor
}

type UiSelection struct {
	active bool

	appName string

	brushes []UiSelectionBrush

	backup_edit_hash uint64
}

func (s *UiSelection) IsActive() bool {
	return s.active
}

func (s *UiSelection) Reset() {
	s.brushes = nil
}

func (s *UiSelection) Draw(buff *WinPaintBuff, ui *Ui) {

	if ui.parent.win.io.Keys.Ctrl {
		n := 0
		ui.dom.postDraw(s.appName, 0, &n) //dialogs? ....
	}

	buff.AddCrop(ui.dom.CropWithScroll())

	thick := ui.CellWidth(0.3)
	for _, it := range s.brushes {
		buff.AddLines(it.Points, it.Cd.Cd, thick, true)
	}
}

func (s *UiSelection) UpdateComp(ui *Ui) {

	//start
	if ui.GetWin().io.Keys.Ctrl {
		if ui.GetWin().io.Touch.Start {
			s.active = true
			s.brushes = append(s.brushes, UiSelectionBrush{Cd: Layout3_Get_prompt_color(len(s.brushes))})

			s.backup_edit_hash = ui.parent.edit.hash
		}
	}

	if !s.IsActive() {
		return
	}

	//add point
	s.brushes[len(s.brushes)-1].Points = append(s.brushes[len(s.brushes)-1].Points, ui.GetWin().io.Touch.Pos)

	//end
	if ui.GetWin().io.Touch.End {
		s.active = false

		appLay := ui.dom.FindFirstName(s.appName)
		if appLay != nil {

			cq := s.getRect()
			cq = cq.Crop(ui.Cell() / 3)
			cq.Size = cq.Size.Max(OsV2{1, 1})
			cqArea := float64(cq.Area())
			if cqArea > 0 {

				best_layout := appLay
				best_area := 0.0
				appLay.findSelection(cq, cqArea, &best_area, &best_layout, s.appName)

				st_rel := cq.Start.Sub(best_layout.canvas.Start)
				en_rel := cq.End().Sub(best_layout.canvas.Start)
				stCol := best_layout.cols.GetCellPos(st_rel.X, ui.Cell())
				stRow := best_layout.rows.GetCellPos(st_rel.Y, ui.Cell())
				enCol := best_layout.cols.GetCellPos(en_rel.X, ui.Cell())
				enRow := best_layout.rows.GetCellPos(en_rel.Y, ui.Cell())

				grid := InitOsV4ab(OsV2{stCol, stRow}, OsV2{enCol, enRow})
				grid.Size.X++
				grid.Size.Y++

				var pick LayoutPick
				pick.X = grid.Start.X
				pick.Y = grid.Start.Y
				pick.W = grid.Size.X
				pick.H = grid.Size.Y
				pick.Cd = s.brushes[len(s.brushes)-1].Cd.Cd
				pick.Label = s.brushes[len(s.brushes)-1].Cd.Label

				fmt.Println("--pick", pick.X, pick.Y, pick.W, pick.H)

				if best_layout == appLay {
					//get Build() pos
					wf, err := Compile_getWidgetFile(s.appName)
					if err != nil {
						fmt.Println("Error:", err)
						return
					}
					if wf.Build < 0 {
						fmt.Println("Error 1456")
						return
					}
					pick.Line = wf.Build

				} else {
					pick.Line = best_layout.props.Caller_line
				}

				in := LayoutInput{Pick: pick, PickApp: s.appName}
				ui.parent.CallInput(&ui.dom.props, &in)
			}
		}

		//recover editbox
		ui.SetRefresh()
		ui._refresh()
		ui.parent.edit.reload_hash = s.backup_edit_hash
		s.backup_edit_hash = 0
	}
}

func (s *UiSelection) getRect() OsV4 {

	points := s.brushes[len(s.brushes)-1].Points

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

func (dom *Layout3) findSelection(cq OsV4, cqArea float64, best_area *float64, best_layout **Layout3, appName string) {
	if dom.touchDia && /*dom.props.Caller_file == appName+".go" &&*/ (dom.props.Name == appName || dom.props.Name == "_layout") {
		inArea := float64(dom.crop.GetIntersect(cq).Area()) / cqArea
		if inArea >= *best_area {
			*best_area = inArea
			*best_layout = dom
		}
	}

	for _, it := range dom.childs {
		if it.IsShown() {
			it.findSelection(cq, cqArea, best_area, best_layout, appName)
		}
	}

	if dom.dialog != nil {
		dom.dialog.findSelection(cq, cqArea, best_area, best_layout, appName)
	}
}

type LayoutPromptColor struct {
	Label string
	Cd    color.RGBA
}

var g_prompt_colors = []LayoutPromptColor{
	{Label: "red", Cd: color.RGBA{255, 0, 0, 255}},
	{Label: "green", Cd: color.RGBA{0, 255, 0, 255}},
	{Label: "blue", Cd: color.RGBA{0, 0, 255, 255}},

	{Label: "yellow", Cd: color.RGBA{200, 200, 0, 255}},
	{Label: "aqua", Cd: color.RGBA{0, 255, 255, 255}},
	{Label: "fuchsia", Cd: color.RGBA{255, 0, 255, 255}},

	{Label: "olive", Cd: color.RGBA{128, 128, 0, 255}},
	{Label: "teal", Cd: color.RGBA{0, 128, 128, 255}},
	{Label: "purple", Cd: color.RGBA{128, 0, 128, 255}},

	{Label: "navy", Cd: color.RGBA{0, 0, 128, 255}},
	{Label: "marron", Cd: color.RGBA{128, 0, 0, 255}},
}

func Layout3_Get_prompt_color(i int) LayoutPromptColor {
	return g_prompt_colors[i%len(g_prompt_colors)]
}

func (dom *Layout3) postDraw(name string, depth int, num_cds *int) {

	if dom.props.Name == name || (dom.props.Caller_file == name+".go" && dom.props.Name == "_layout") {

		cd := Layout3_Get_prompt_color(*num_cds)
		cd.Cd.A = 200

		dom.drawGrid(cd.Cd, 0.03, depth)

		(*num_cds)++
		depth++
	}

	//subs
	for _, it := range dom.childs {
		if it.IsShown() {
			it.postDraw(name, depth, num_cds)
		}
	}
}
