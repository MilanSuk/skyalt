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

type UiSelection struct {
	activeComp bool
	activeGrid bool

	points []OsV2

	draw_hash uint64
	draw_grid OsV4

	ShowGrid bool
}

func (s *UiSelection) IsActive() bool {
	return s.activeComp || s.activeGrid
}

func (s *UiSelection) Draw(buff *WinPaintBuff, ui *Ui) {
	if s.draw_hash == 0 {
		return
	}

	lay := ui.dom.FindHash(s.draw_hash)
	if lay == nil {
		return
	}

	cd := ui.GetPalette().S
	thick := ui.CellWidth(0.3)

	buff.AddLines(s.points, cd, thick, true)
	buff.AddRect(lay.crop.Crop(thick/3), ui.GetPalette().T, thick/3)

	if s.activeGrid {
		//grid
		cell := float64(ui.Cell())

		sx := lay.cols.GetSumOutput(s.draw_grid.Start.X)
		ex := lay.cols.GetSumOutput(s.draw_grid.End().X)
		sy := lay.rows.GetSumOutput(s.draw_grid.Start.Y)
		ey := lay.rows.GetSumOutput(s.draw_grid.End().Y)

		rc := lay._getRect()
		rc.X += float64(sx) / cell
		rc.Y += float64(sy) / cell
		rc.W = float64(ex-sx) / cell
		rc.H = float64(ey-sy) / cell

		rc = rc.Cut(0.2)
		cd := cd
		cd.A = 180

		buff.AddRect(lay.getCanvasPx(rc), cd, 0)
	}
}

func (dom *Layout3) findSelection(cq OsV4, cqArea float64, best_area *float64, best_layout **Layout3) {
	if dom.touchDia {
		inArea := float64(dom.crop.GetIntersect(cq).Area()) / cqArea
		if inArea >= *best_area {
			*best_area = inArea
			*best_layout = dom
		}
	}

	for _, it := range dom.childs {
		if it.IsShown() {
			it.findSelection(cq, cqArea, best_area, best_layout)
		}
	}

	if dom.dialog != nil {
		dom.dialog.findSelection(cq, cqArea, best_area, best_layout)
	}
}

func (s *UiSelection) getRect() OsV4 {
	if len(s.points) == 0 {
		return OsV4{}
	}

	min := s.points[0]
	max := s.points[0]

	for _, pt := range s.points {
		min.X = OsMin(min.X, pt.X)
		min.Y = OsMin(min.Y, pt.Y)
		max.X = OsMax(max.X, pt.X)
		max.Y = OsMax(max.Y, pt.Y)
	}
	return InitOsV4ab(min, max)
}

func (s *UiSelection) UpdateComp(ui *Ui) {

	//start
	if ui.GetWin().io.Keys.Ctrl {
		if ui.GetWin().io.Touch.Start {
			if !ui.GetWin().io.Touch.Rm {
				s.activeComp = true
			} else {
				s.activeGrid = true
			}
		}
	}

	if !s.IsActive() {
		return
	}

	//add point
	s.points = append(s.points, ui.GetWin().io.Touch.Pos)

	best_layout := ui.dom
	{
		//update
		cq := s.getRect()
		cq = cq.Crop(ui.Cell() / 3)
		cq.Size = cq.Size.Max(OsV2{1, 1})
		cqArea := float64(cq.Area())
		if cqArea > 0 {

			best_area := 0.0
			ui.dom.findSelection(cq, cqArea, &best_area, &best_layout)

			if s.activeGrid {
				if best_layout.parent != nil && best_layout.props.Name != "_layout" && !best_layout.props.App && len(best_layout.childs) == 0 {
					best_layout = best_layout.parent
				}

				st_rel := cq.Start.Sub(best_layout.canvas.Start)
				en_rel := cq.End().Sub(best_layout.canvas.Start)
				stCol := best_layout.cols.GetCloseCell(st_rel.X)
				stRow := best_layout.rows.GetCloseCell(st_rel.Y)
				enCol := best_layout.cols.GetCloseCell(en_rel.X)
				enRow := best_layout.rows.GetCloseCell(en_rel.Y)

				s.draw_grid = InitOsV4ab(OsV2{stCol, stRow}, OsV2{enCol, enRow})
				s.draw_grid.Size.X++
				s.draw_grid.Size.Y++
			}

			s.draw_hash = best_layout.props.Hash
		}
	}

	if ui.GetWin().io.Touch.End {
		if s.activeComp {
			in := LayoutInput{Pick: InitLayoutPick(best_layout.props.Caller_file, best_layout.props.Caller_line, OsV4{}, best_layout.props.LLMTip)}
			ui.parent.CallInput(&ui.dom.props, &in)

		} else {
			Caller_file := best_layout.props.Caller_file
			Caller_line := best_layout.props.Caller_line

			in := LayoutInput{Pick: InitLayoutPick(Caller_file, Caller_line, s.draw_grid, best_layout.props.LLMTip)}
			ui.parent.CallInput(&ui.dom.props, &in)
		}

		//reset
		s.draw_hash = 0
		s.activeGrid = false
		s.activeComp = false
		s.points = nil
	}
}
