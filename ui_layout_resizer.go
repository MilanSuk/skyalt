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
	"math"
)

func (layout *Layout) _renderResizeDraw(i int, cd color.RGBA, vertical bool) {

	cell := layout.Cell()

	layoutScreen := layout.canvas

	rspace := LayoutArray_resizerSize(cell)
	if vertical {
		layoutScreen.Start.X += layout.cols.GetResizerPos(i, cell)
		layoutScreen.Size.X = rspace
	} else {
		layoutScreen.Start.Y += layout.rows.GetResizerPos(i, cell)
		layoutScreen.Size.Y = rspace
	}

	layout.ui.GetWin().buff.AddRect(layoutScreen.Crop(4), cd, 0)
}

func (layout *Layout) _setResizer(i int, value float64, isCol bool) bool {

	diff := false

	value = math.Max(0.3, value/float64(layout.Cell()))

	var ind int
	var arr *UiLayoutArray
	if isCol {
		ind = layout.cols.GetResizeIndex(i)
		arr = &layout.cols
	} else {
		ind = layout.rows.GetResizeIndex(i)
		arr = &layout.rows
	}

	if ind >= 0 {
		diff = (arr.inputs[ind].resize.value != value)
		//arr.inputs[ind].resize.value = value
		if diff {

			if isCol {
				layout.GetSettings().SetCol(layout.UID, ind, value) //save
				for i, it := range layout.UserCols {
					if it.Pos == ind {
						layout.UserCols[i].Resize_value = value //temporary(mouse down) update
					}
				}

			} else {
				layout.GetSettings().SetRow(layout.UID, ind, value) //save
				for i, it := range layout.UserRows {
					if it.Pos == ind {
						layout.UserRows[i].Resize_value = value //temporary(mouse down) update
					}
				}
			}

			layout.ui.SetRelayoutHard()
		}
	}

	return diff
}

func (layout *Layout) updateResizer() {
	enableInput := layout.CanTouch()

	cell := layout.Cell()
	tpos := layout.ui.GetWin().io.Touch.Pos.Sub(layout.canvas.Start)

	layout.touchResizeHighlightCol = false
	layout.touchResizeHighlightRow = false
	col := -1
	row := -1
	if enableInput && layout.IsTouchPosInside() {
		col = layout.cols.IsResizerTouch((tpos.X), cell)
		row = layout.rows.IsResizerTouch((tpos.Y), cell)

		layout.touchResizeHighlightCol = (col >= 0)
		layout.touchResizeHighlightRow = (row >= 0)

		// start
		if layout.IsMouseButtonDownStart() && (layout.touchResizeHighlightCol || layout.touchResizeHighlightRow) {
			if layout.touchResizeHighlightCol || layout.touchResizeHighlightRow {
				layout.ui.touch.Set(0, 0, 0, layout.UID)
			}

			if layout.touchResizeHighlightCol {
				layout.touchResizeIndex = col
				layout.touchResizeIsActive = true
			}
			if layout.touchResizeHighlightRow {
				layout.touchResizeIndex = row
				layout.touchResizeIsActive = false
			}
		}

		if layout.ui.touch.IsActive() {
			layout.touchResizeHighlightCol = false
			layout.touchResizeHighlightRow = false
			//active = true
		}
	}

	// resize
	if layout.ui.touch.IsFnMove(0, 0, 0, layout.UID) {

		r := 1.0
		if layout.touchResizeIsActive {
			col = layout.touchResizeIndex
			layout.touchResizeHighlightCol = true

			if layout.cols.IsLastResizeValid() && int(col) == layout.cols.NumInputs()-2 {
				r = float64(layout.canvas.Size.X - tpos.X) // last
			} else {
				r = float64(tpos.X - layout.cols.GetResizerPos(int(col)-1, cell))
			}

			layout._setResizer(int(col), r, true)
		} else {
			row = layout.touchResizeIndex
			layout.touchResizeHighlightRow = true

			if layout.rows.IsLastResizeValid() && int(row) == layout.rows.NumInputs()-2 {
				r = float64(layout.canvas.Size.Y - tpos.Y) // last
			} else {
				r = float64(tpos.Y - (layout.rows.GetResizerPos(int(row)-1, cell)))
			}

			layout._setResizer(int(row), r, false)
		}

		if layout.ui.GetWin().io.Touch.End {
			layout.ui.SetRefresh()
		}
	}

	// cursor
	if enableInput {
		if layout.touchResizeHighlightCol {
			layout.ui.GetWin().PaintCursor("res_col")
		}
		if layout.touchResizeHighlightRow {
			layout.ui.GetWin().PaintCursor("res_row")
		}
	}
}

func (layout *Layout) drawResizer() {
	activeCd := layout.GetPalette().P
	activeCd.A = 150

	defaultCd := layout.GetPalette().GetGrey(0.5)

	for i := 0; i < layout.cols.NumInputs(); i++ {
		if layout.cols.GetResizeIndex(i) >= 0 {
			if layout.touchResizeHighlightCol && i == layout.touchResizeIndex {
				layout._renderResizeDraw(i, activeCd, true)
			} else {
				layout._renderResizeDraw(i, defaultCd, true)
			}
		}
	}

	for i := 0; i < layout.rows.NumInputs(); i++ {
		if layout.rows.GetResizeIndex(i) >= 0 {
			if layout.touchResizeHighlightRow && i == layout.touchResizeIndex {
				layout._renderResizeDraw(i, activeCd, false)
			} else {
				layout._renderResizeDraw(i, defaultCd, false)
			}
		}
	}
}
