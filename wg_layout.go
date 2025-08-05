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
	"strings"
	"time"
)

func (layout *Layout) SetClipboardText(text string) {
	layout.ui.win.SetClipboardText(text)
}
func (layout *Layout) CopyText() {
	layout.ui.edit.KeyCopy = true
}
func (layout *Layout) SelectAllText() {
	layout.ui.edit.KeySelectAll = true
}
func (layout *Layout) CutText() {
	layout.ui.edit.KeyCut = true
}
func (layout *Layout) PasteText() {
	layout.ui.edit.KeyPaste = true
}
func (layout *Layout) RecordText() {
	layout.ui.edit.KeyRecord = true
}

func (layout *Layout) _createDiv(x, y, w, h int, name string, fnBuild func(layout *Layout), fnDraw func(rect Rect, layout *Layout) LayoutPaint, fnInput func(in LayoutInput, layout *Layout)) *Layout {

	var lay *Layout
	//find
	for _, it := range layout.childs {
		if it.X == x && it.Y == y && it.W == w && it.H == h && it.Name == name {
			lay = it
			break
		}
	}

	//add
	if lay == nil {
		lay = _newLayout(x, y, w, h, name, layout)
		layout.childs = append(layout.childs, lay)
	}

	//set
	lay.fnBuild = fnBuild
	lay.fnDraw = fnDraw
	lay.fnInput = fnInput

	return lay
}

func (layout *Layout) AddLayoutWithName(x, y, w, h int, name string) *Layout {
	return layout._createDiv(x, y, w, h, "_layout_"+name, nil, nil, nil)
}
func (layout *Layout) AddLayout(x, y, w, h int) *Layout {
	lay := layout._createDiv(x, y, w, h, "_layout", nil, nil, nil)
	lay.fnGetLLMTip = func(layout *Layout) string {
		return Layout_buildLLMTip("Layout", "", false, layout.Tooltip)
	}
	return lay
}

func (layout *Layout) AddLayoutCards(x, y, w, h int, autoSpacing bool) *Layout {
	lay := layout._createDiv(x, y, w, h, "_cards", nil, nil, nil)
	lay.Cards_autoSpacing = autoSpacing
	return lay
}

func (layout *Layout) IsTypeCards() bool {
	return layout.Name == "_cards"
}
func (layout *Layout) IsTypeLayout() bool {
	return strings.HasPrefix(layout.Name, "_layout")
}

func (layout *Layout) AddCardsSubItem() *Layout {
	return layout.AddLayout(0, len(layout.childs), 1, 1)
}

func (dia *LayoutDialog) OpenCentered() {
	dia.Layout.ui.settings.OpenDialog(dia.Layout.UID, 0, OsV2{})
}
func (dia *LayoutDialog) OpenRelative(parent_uid uint64) {
	if parent_uid > 0 {
		dia.Layout.ui.settings.OpenDialog(dia.Layout.UID, parent_uid, OsV2{})
	} else {
		dia.OpenCentered()
	}
}
func (dia *LayoutDialog) OpenOnTouch() {
	dia.Layout.ui.settings.OpenDialog(dia.Layout.UID, 0, dia.Layout.ui.GetWin().io.Touch.Pos)
}
func (dia *LayoutDialog) Close() {
	dd := dia.Layout.ui.settings.FindDialog(dia.Layout.UID)
	if dd != nil {
		dia.Layout.ui.settings.CloseDialog(dd)
	}
}

func (layout *Layout) FindDialog(name string) *LayoutDialog {
	for _, it := range layout.dialogs {
		if it.Layout != nil && it.Layout.Name == name {
			return it
		}
	}
	return nil
}
func (layout *Layout) AddDialog(name string) *LayoutDialog {

	dia := layout.FindDialog(name)
	if dia == nil {
		dia = &LayoutDialog{Layout: _newLayout(0, 0, 0, 0, name, layout)}
		layout.dialogs = append(layout.dialogs, dia)
	} else {
		fmt.Println("Dialog already exist")
	}

	return dia
}

func (layout *Layout) AddDialogBorder(name string, title string, width float64) (*LayoutDialog, *Layout) {
	dia := layout.AddDialog(name)
	lay := dia.Layout
	if width > 0 {
		lay.SetColumn(1, 1, width)
	} else {
		lay.SetColumnFromSub(1, 1, 100, true)
	}
	lay.SetRowFromSub(1, 1, 100, true)
	lay.SetColumn(2, 1, 1)
	lay.SetRow(2, 1, 1)

	tx := lay.AddText(0, 0, 3, 1, title)
	tx.Align_h = 1

	return dia, lay.AddLayout(1, 1, 1, 1)
}

func (layout *Layout) SetColumn(grid_x int, min_size, max_size float64) {
	newItem := LayoutCR{Pos: grid_x, Min: min_size, Max: max_size}

	for i := range layout.UserCols {
		if layout.UserCols[i].Pos == grid_x {
			layout.UserCols[i] = newItem
			return
		}
	}

	layout.UserCols = append(layout.UserCols, newItem)
}

func (layout *Layout) SetRow(grid_y int, min_size, max_size float64) {
	newItem := LayoutCR{Pos: grid_y, Min: min_size, Max: max_size}

	for i := range layout.UserRows {
		if layout.UserRows[i].Pos == grid_y {
			layout.UserRows[i] = newItem
			return
		}
	}

	layout.UserRows = append(layout.UserRows, newItem)
}

func (layout *Layout) SetColumnFromSub(grid_x int, min_size, max_size float64, fix bool) {
	newItem := LayoutCR{Pos: grid_x, SetFromChild_min: min_size, SetFromChild_max: max_size, SetFromChild_fix: fix}

	for i := range layout.UserCols {
		if layout.UserCols[i].Pos == grid_x {
			layout.UserCols[i] = newItem
			return
		}
	}

	layout.UserCols = append(layout.UserCols, newItem)
}

func (layout *Layout) SetRowFromSub(grid_y int, min_size, max_size float64, fix bool) {
	newItem := LayoutCR{Pos: grid_y, SetFromChild_min: min_size, SetFromChild_max: max_size, SetFromChild_fix: fix}

	for i := range layout.UserRows {
		if layout.UserRows[i].Pos == grid_y {
			layout.UserRows[i] = newItem
			return
		}
	}

	layout.UserRows = append(layout.UserRows, newItem)
}

func (layout *Layout) SetColumnResizable(grid_x int, min_size, max_size, default_size float64) {
	layout.UserCols = append(layout.UserCols, LayoutCR{Pos: grid_x, Min: min_size, Max: max_size, Resize_value: default_size})

}
func (layout *Layout) SetRowResizable(grid_y int, min_size, max_size, default_size float64) {
	layout.UserRows = append(layout.UserRows, LayoutCR{Pos: grid_y, Min: min_size, Max: max_size, Resize_value: default_size})
}

func (paint *LayoutPaint) Rect(rect Rect, cd, cd_over, cd_down color.RGBA, borderWidth float64) *LayoutDrawRect {
	return paint.RectRad(rect, cd, cd_over, cd_down, borderWidth, 0)
}
func (paint *LayoutPaint) RectRad(rect Rect, cd, cd_over, cd_down color.RGBA, borderWidth float64, radius float64) *LayoutDrawRect {
	prim := &LayoutDrawRect{Cd: cd, Cd_over: cd_over, Cd_down: cd_down, Border: borderWidth, Radius: radius}
	paint.buffer = append(paint.buffer, LayoutDrawPrim{Rect: rect, Rectangle: prim})
	return prim
}
func (paint *LayoutPaint) Circle(rect Rect, cd, cd_over, cd_down color.RGBA, borderWidth float64) *LayoutDrawRect {
	prim := &LayoutDrawRect{Cd: cd, Cd_over: cd_over, Cd_down: cd_down, Border: borderWidth}
	paint.buffer = append(paint.buffer, LayoutDrawPrim{Rect: rect, Circle: prim})
	return prim
}

func (paint *LayoutPaint) CircleRad(rect Rect, x, y float64, rad_cells float64, cd, cd_over, cd_down color.RGBA, borderWidth float64) *LayoutDrawRect {
	if rad_cells <= 0 {
		return &LayoutDrawRect{}
	}

	rect.X += rect.W * x
	rect.Y += rect.H * y
	rect.W = rad_cells
	rect.H = rad_cells

	//move
	rect.X -= rect.W / 2
	rect.Y -= rect.H / 2

	return paint.Circle(rect, cd, cd, cd, 0)
}

func (paint *LayoutPaint) File(rect Rect, image_path WinImagePath, cd, cd_over, cd_down color.RGBA, align_h, align_v uint8) *LayoutDrawFile {
	prim := &LayoutDrawFile{Cd: cd, Cd_over: cd_over, Cd_down: cd_down, ImagePath: image_path, Align_h: align_h, Align_v: align_v}
	paint.buffer = append(paint.buffer, LayoutDrawPrim{Rect: rect, File: prim})
	return prim
}

func (paint *LayoutPaint) Line(rect Rect, sx, sy, ex, ey float64, cd color.RGBA, width float64) *LayoutDrawLine {
	prim := &LayoutDrawLine{Cd: cd, Cd_over: cd, Cd_down: cd, Border: width, Sx: sx, Sy: sy, Ex: ex, Ey: ey}
	paint.buffer = append(paint.buffer, LayoutDrawPrim{Rect: rect, Line: prim})
	return prim
}

func (paint *LayoutPaint) Text(rect Rect, text string, ghost string, frontCd, frontCd_over, frontCd_down color.RGBA, selection, editable bool, align_h uint8, align_v uint8) *LayoutDrawText {
	prim := &LayoutDrawText{
		Cd: frontCd, Cd_over: frontCd_over, Cd_down: frontCd_down,
		Text:         text,
		Ghost:        ghost,
		Align_h:      align_h,
		Align_v:      align_v,
		Formating:    true,
		Multiline:    false,
		Linewrapping: false,
		Selection:    selection,
		Editable:     editable,
		Refresh:      false,
	}

	paint.buffer = append(paint.buffer, LayoutDrawPrim{Rect: rect, Text: prim})
	return prim
}

func (paint *LayoutPaint) Brush(cd color.RGBA, points []OsV2) *LayoutDrawBrush {

	prim := &LayoutDrawBrush{
		Cd:     cd,
		Points: points,
	}

	paint.buffer = append(paint.buffer, LayoutDrawPrim{Brush: prim})
	return prim
}

func (paint *LayoutPaint) CursorEx(rect Rect, name string) *LayoutDrawCursor {
	prim := &LayoutDrawCursor{Name: name}
	paint.buffer = append(paint.buffer, LayoutDrawPrim{Rect: rect, Cursor: prim})
	return prim
}
func (paint *LayoutPaint) Cursor(name string, rect Rect) {
	paint.CursorEx(rect, name)
}

func (paint *LayoutPaint) TooltipEx(rect Rect, description string, force bool) *LayoutDrawTooltip {
	if description != "" {
		prim := &LayoutDrawTooltip{Description: description, Force: force}
		paint.buffer = append(paint.buffer, LayoutDrawPrim{Rect: rect, Tooltip: prim})
		return prim
	}
	return &LayoutDrawTooltip{}
}
func (paint *LayoutPaint) Tooltip(text string, rect Rect) {
	paint.TooltipEx(rect, text, false)
}

func (layout *Layout) GetMonthText(month int) string {
	switch month {
	case 1:
		return "January"
	case 2:
		return "February"
	case 3:
		return "March"
	case 4:
		return "April"
	case 5:
		return "May"
	case 6:
		return "June"
	case 7:
		return "July"
	case 8:
		return "August"
	case 9:
		return "September"
	case 10:
		return "October"
	case 11:
		return "November"
	case 12:
		return "December"
	}
	return ""
}

func (layout *Layout) GetDayTextFull(day int) string {
	switch day {
	case 1:
		return "Monday"
	case 2:
		return "Tuesday"
	case 3:
		return "Wednesday"
	case 4:
		return "Thursday"
	case 5:
		return "Friday"
	case 6:
		return "Saturday"
	case 7:
		return "Sunday"
	}
	return ""
}

func (layout *Layout) GetDayTextShort(day int) string {
	switch day {
	case 1:
		return "Mon"
	case 2:
		return "Tue"
	case 3:
		return "Wed"
	case 4:
		return "Thu"
	case 5:
		return "Fri"
	case 6:
		return "Sat"
	case 7:
		return "Sun"
	}
	return ""
}

func (layout *Layout) ConvertTextTime(unix_sec int64) string {
	tm := time.Unix(unix_sec, 0)
	return fmt.Sprintf("%.02d:%.02d", tm.Hour(), tm.Minute())
}

func (layout *Layout) ConvertTextDate(unix_sec int64) string {
	tm := time.Unix(unix_sec, 0)

	switch layout.ui.router.services.sync.GetDateFormat() {
	case "eu":
		return fmt.Sprintf("%d/%d/%d", tm.Day(), int(tm.Month()), tm.Year())

	case "us":
		return fmt.Sprintf("%d/%d/%d", int(tm.Month()), tm.Day(), tm.Year())

	case "iso":
		return fmt.Sprintf("%d-%02d-%02d", tm.Year(), int(tm.Month()), tm.Day())

	case "text":
		return fmt.Sprintf("%s %d, %d", layout.GetMonthText(int(tm.Month())), tm.Day(), tm.Year())

	case "2base":
		return fmt.Sprintf("%d %d-%d", tm.Year(), int(tm.Month()), tm.Day())
	}

	return ""
}
func (layout *Layout) ConvertTextDateTime(unix_sec int64) string {
	return layout.ConvertTextDate(unix_sec) + " " + layout.ConvertTextTime(unix_sec)
}
