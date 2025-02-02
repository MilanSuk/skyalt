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

func (dom *Layout3) _renderResizeDraw(i int, cd color.RGBA, vertical bool) {

	cell := dom.Cell()

	layoutScreen := dom.canvas

	rspace := LayoutArray_resizerSize(cell)
	if vertical {
		layoutScreen.Start.X += dom.cols.GetResizerPos(i, cell)
		layoutScreen.Size.X = rspace
	} else {
		layoutScreen.Start.Y += dom.rows.GetResizerPos(i, cell)
		layoutScreen.Size.Y = rspace
	}

	dom.ui.GetWin().buff.AddRect(layoutScreen.Crop(4), cd, 0)
}

func (dom *Layout3) _setResizer(i int, value float64, isCol bool) bool {

	diff := false

	value = math.Max(0.3, value/float64(dom.Cell()))

	var ind int
	var arr *UiLayoutArray
	if isCol {
		ind = dom.cols.GetResizeIndex(i)
		arr = &dom.cols
	} else {
		ind = dom.rows.GetResizeIndex(i)
		arr = &dom.rows
	}

	if ind >= 0 {
		diff = (arr.inputs[ind].resize.value != float32(value))
		//arr.inputs[ind].resize.value = float32(value)
		if diff {

			if isCol {
				dom.GetSettings().SetCol(dom.props.Hash, ind, value) //save
				for i, it := range dom.props.UserCols {
					if it.Pos == ind {
						dom.props.UserCols[i].Resize_value = value //temporary(mouse down) update
					}
				}

			} else {
				dom.GetSettings().SetRow(dom.props.Hash, ind, value) //save
				for i, it := range dom.props.UserRows {
					if it.Pos == ind {
						dom.props.UserRows[i].Resize_value = value //temporary(mouse down) update
					}
				}
			}

			dom.ui.SetRelayout()
			dom.ui.SetRedrawBuffer()
		}
	}

	return diff
}

func (dom *Layout3) updateResizer() {
	enableInput := dom.CanTouch()

	cell := dom.Cell()
	tpos := dom.ui.GetWin().io.Touch.Pos.Sub(dom.canvas.Start)

	dom.touchResizeHighlightCol = false
	dom.touchResizeHighlightRow = false
	col := -1
	row := -1
	if enableInput && dom.IsTouchPosInside() {
		col = dom.cols.IsResizerTouch((tpos.X), cell)
		row = dom.rows.IsResizerTouch((tpos.Y), cell)

		dom.touchResizeHighlightCol = (col >= 0)
		dom.touchResizeHighlightRow = (row >= 0)

		// start
		if dom.IsMouseButtonDownStart() && (dom.touchResizeHighlightCol || dom.touchResizeHighlightRow) {
			if dom.touchResizeHighlightCol || dom.touchResizeHighlightRow {
				dom.GetUis().touch.Set(0, 0, 0, dom.props.Hash)
			}

			if dom.touchResizeHighlightCol {
				dom.touchResizeIndex = col
				dom.touchResizeIsActive = true
			}
			if dom.touchResizeHighlightRow {
				dom.touchResizeIndex = row
				dom.touchResizeIsActive = false
			}
		}

		if dom.GetUis().touch.IsActive() {
			dom.touchResizeHighlightCol = false
			dom.touchResizeHighlightRow = false
			//active = true
		}
	}

	// resize
	if dom.GetUis().touch.IsFnMove(0, 0, 0, dom.props.Hash) {

		r := 1.0
		if dom.touchResizeIsActive {
			col = dom.touchResizeIndex
			dom.touchResizeHighlightCol = true

			if dom.cols.IsLastResizeValid() && int(col) == dom.cols.NumInputs()-2 {
				r = float64(dom.canvas.Size.X - tpos.X) // last
			} else {
				r = float64(tpos.X - dom.cols.GetResizerPos(int(col)-1, cell))
			}

			dom._setResizer(int(col), r, true)
		} else {
			row = dom.touchResizeIndex
			dom.touchResizeHighlightRow = true

			if dom.rows.IsLastResizeValid() && int(row) == dom.rows.NumInputs()-2 {
				r = float64(dom.canvas.Size.Y - tpos.Y) // last
			} else {
				r = float64(tpos.Y - (dom.rows.GetResizerPos(int(row)-1, cell)))
			}

			dom._setResizer(int(row), r, false)
		}

		if dom.ui.GetWin().io.Touch.End {
			dom.ui.SetRefresh()
		}
	}

	// cursor
	if enableInput {
		if dom.touchResizeHighlightCol {
			dom.ui.GetWin().PaintCursor("res_col")
		}
		if dom.touchResizeHighlightRow {
			dom.ui.GetWin().PaintCursor("res_row")
		}
	}
}

func (dom *Layout3) drawResizer() {
	activeCd := dom.GetPalette().P
	activeCd.A = 150

	defaultCd := dom.GetPalette().GetGrey(0.5)

	for i := 0; i < dom.cols.NumInputs(); i++ {
		if dom.cols.GetResizeIndex(i) >= 0 {
			if dom.touchResizeHighlightCol && i == dom.touchResizeIndex {
				dom._renderResizeDraw(i, activeCd, true)
			} else {
				dom._renderResizeDraw(i, defaultCd, true)
			}
		}
	}

	for i := 0; i < dom.rows.NumInputs(); i++ {
		if dom.rows.GetResizeIndex(i) >= 0 {
			if dom.touchResizeHighlightRow && i == dom.touchResizeIndex {
				dom._renderResizeDraw(i, activeCd, false)
			} else {
				dom._renderResizeDraw(i, defaultCd, false)
			}
		}
	}
}
