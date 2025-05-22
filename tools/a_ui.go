/*
Copyright 2025 Milan Suk

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

type UIText struct {
	layout  *UI
	Label   string
	Align_h int
	Align_v int
	Cd      color.RGBA
	Tooltip string

	Selection    bool
	Formating    bool
	Multiline    bool
	Linewrapping bool

	EnableDropFile bool
	dropFile       func(pathes []string) error
}
type UIEditbox struct {
	layout     *UI
	Name       string
	Error      string
	Value      *string
	ValueFloat *float64
	ValueInt   *int
	Precision  int
	Ghost      string
	Tooltip    string
	Password   bool

	Align_h int //0=left, 1=center, 2=right
	Align_v int //0=top, 1=center, 2=bottom

	Formating    bool
	Multiline    bool
	Linewrapping bool

	changed func() error
	enter   func() error
}
type UISlider struct {
	layout *UI
	Error  string
	Value  *float64
	Min    float64
	Max    float64
	Step   float64

	changed func() error
}
type UIButton struct {
	layout  *UI
	Label   string
	Tooltip string
	Align   int

	Shortcut byte

	Background  float64
	Border      bool
	IconBlob    []byte
	IconPath    string
	Icon_align  int
	Icon_margin float64
	BrowserUrl  string
	Cd          color.RGBA

	ConfirmQuestion string

	Drag_group              string
	Drop_group              string
	Drag_source             string
	Drag_index              int
	Drop_h, Drop_v, Drop_in bool

	clicked func() error

	dropMove func(src_i, dst_i int, src_source, dst_source string) error
}

type UIFilePickerButton struct {
	layout      *UI
	Error       string
	Path        *string
	Preview     bool
	OnlyFolders bool

	changed func() error
}
type UIDatePickerButton struct {
	layout   *UI
	Error    string
	Date     *int64
	Page     *int64
	ShowTime bool
	changed  func() error
}
type UIColorPickerButton struct {
	layout  *UI
	Error   string
	Cd      *color.RGBA
	changed func() error
}

type UICombo struct {
	layout      *UI
	Error       string
	Value       *string
	Labels      []string
	Values      []string
	DialogWidth float64

	changed func() error
}

type UISwitch struct {
	layout  *UI
	Error   string
	Label   string
	Tooltip string
	Value   *bool

	changed func() error
}

type UICheckbox struct {
	layout  *UI
	Error   string
	Label   string
	Tooltip string
	Value   *float64

	changed func() error
}

type UIDivider struct {
	layout     *UI
	Horizontal bool
}

type UIOsmMapLoc struct {
	Lon   float64
	Lat   float64
	Label string
}
type UIOsmMapLocators struct {
	Locators []UIOsmMapLoc
	clicked  func(i int, caller *ToolCaller)
}

type UIOsmMapSegmentTrk struct {
	Lon  float64
	Lat  float64
	Ele  float64
	Time string
	Cd   color.RGBA
}
type UIOsmMapSegment struct {
	Trkpts []UIOsmMapSegmentTrk
	Label  string
}
type UIOsmMapRoute struct {
	Segments []UIOsmMapSegment
}
type UIOsmMap struct {
	layout         *UI
	Lon, Lat, Zoom *float64
	Locators       []UIOsmMapLocators
	Routes         []UIOsmMapRoute
}

type UIChartPoint struct {
	X  float64
	Y  float64
	Cd color.RGBA
}

type UIChartLine struct {
	Points []UIChartPoint
	Label  string
	Cd     color.RGBA
}

type UIChartLines struct {
	layout *UI

	Lines []UIChartLine

	X_unit, Y_unit        string
	Bound_x0, Bound_y0    bool
	Point_rad, Line_thick float64
	Draw_XHelpLines       bool
	Draw_YHelpLines       bool
}

type UIChartColumnValue struct {
	Value float64
	Label string
	Cd    color.RGBA
}

type UIChartColumn struct {
	Values []UIChartColumnValue
}

type UIChartColumns struct {
	layout *UI

	X_unit, Y_unit string
	Bound_y0       bool
	Y_as_time      bool
	Columns        []UIChartColumn
	X_Labels       []string
	ColumnMargin   float64
}

type UIImage struct {
	layout *UI

	Blob    []byte
	Path    string
	Tooltip string

	Cd          color.RGBA
	Draw_border bool

	Margin  float64
	Align_h int
	Align_v int

	Translate_x, Translate_y float64
	Scale_x, Scale_y         float64
}

type UIList struct {
	layout *UI

	AutoSpacing bool
	//Items       []*UI
}

type UIGridSize struct {
	Pos int
	Min float64
	Max float64

	Default_resize float64

	SetFromChild_min float64 `json:",omitempty"`
	SetFromChild_max float64 `json:",omitempty"`
}
type UIScroll struct {
	Hide   bool
	Narrow bool
}
type UIPaint_Rect struct {
	Cd, Cd_over, Cd_down color.RGBA
	Width                float64
	X, Y, W, H           float64
}
type UIPaint_Circle struct {
	Cd, Cd_over, Cd_down color.RGBA
	Rad                  float64
	Width                float64
	X, Y                 float64
}
type UIPaint_Line struct {
	Cd, Cd_over, Cd_down color.RGBA
	Width                float64
	Sx, Sy, Ex, Ey       float64
}
type UIPaint_Text struct {
	Text                 string
	Cd, Cd_over, Cd_down color.RGBA
	Align_v              int
	Align_h              int
	Formating            bool
	Multiline            bool
	Linewrapping         bool
	Sx, Sy, Ex, Ey       float64
}
type UIPaintBrushPoint struct {
	X int
	Y int
}
type UIPaint_Brush struct {
	Cd     color.RGBA
	Points []UIPaintBrushPoint
}

type UIPaint struct {
	Rectangle *UIPaint_Rect   `json:",omitempty"`
	Circle    *UIPaint_Circle `json:",omitempty"`
	Line      *UIPaint_Line   `json:",omitempty"`
	Text      *UIPaint_Text   `json:",omitempty"`
	Brush     *UIPaint_Brush  `json:",omitempty"`
}

type UICalendarEvent struct {
	EventID int64
	GroupID int64

	Title string

	Start    int64 //unix time
	Duration int64 //seconds

	Color color.RGBA
}

type UIYearCalendar struct {
	layout *UI
	Year   int
}
type UIMonthCalendar struct {
	layout *UI
	Year   int
	Month  int //1=January, 2=February, etc.

	Events []UICalendarEvent
}
type UIDayCalendar struct {
	layout *UI
	Days   []int64
	Events []UICalendarEvent
}

type ToolCmd struct {
	Dialog_Open_UID     uint64 `json:",omitempty"`
	Dialog_Relative_UID uint64 `json:",omitempty"`
	Dialog_OnTouch      bool   `json:",omitempty"`

	Dialog_Close_UID uint64 `json:",omitempty"`

	Editbox_Activate string `json:",omitempty"`

	VScrollToTheTop      uint64 `json:",omitempty"`
	VScrollToTheBottom   uint64 `json:",omitempty"`
	VScrollToTheBottomIf uint64 `json:",omitempty"`
	HScrollToTheLeft     uint64 `json:",omitempty"`
	HScrollToTheRight    uint64 `json:",omitempty"`

	SetClipboardText string `json:",omitempty"`
}

type UIDialog struct {
	UID string
	UI  UI
}

func (dia *UIDialog) OpenCentered(caller *ToolCaller) {
	caller._addCmd(ToolCmd{Dialog_Open_UID: dia.UI.UID})
}
func (dia *UIDialog) OpenRelative(relative *UI, caller *ToolCaller) {
	caller._addCmd(ToolCmd{Dialog_Open_UID: dia.UI.UID, Dialog_Relative_UID: relative.UID})
}
func (dia *UIDialog) OpenOnTouch(caller *ToolCaller) {
	caller._addCmd(ToolCmd{Dialog_Open_UID: dia.UI.UID, Dialog_OnTouch: true})
}
func (dia *UIDialog) Close(caller *ToolCaller) {
	caller._addCmd(ToolCmd{Dialog_Close_UID: dia.UI.UID})
}

func (ui *UI) ActivateEditbox(editbox_name string, caller *ToolCaller) {
	caller._addCmd(ToolCmd{Editbox_Activate: editbox_name})
}

func (ui *UI) VScrollToTheTop(caller *ToolCaller) {
	caller._addCmd(ToolCmd{VScrollToTheTop: ui.UID})
}
func (ui *UI) VScrollToTheBottom(onlyWhenAtBottom bool, caller *ToolCaller) {
	if onlyWhenAtBottom {
		caller._addCmd(ToolCmd{VScrollToTheBottomIf: ui.UID})
	} else {
		caller._addCmd(ToolCmd{VScrollToTheBottom: ui.UID})
	}
}
func (ui *UI) HScrollToTheLeft(caller *ToolCaller) {
	caller._addCmd(ToolCmd{HScrollToTheLeft: ui.UID})
}
func (ui *UI) HScrollToTheRight(caller *ToolCaller) {
	caller._addCmd(ToolCmd{HScrollToTheRight: ui.UID})
}

func (caller *ToolCaller) SetClipboardText(text string) {
	caller._addCmd(ToolCmd{SetClipboardText: text})
}

func (ui *UI) SetColumn(pos int, min, max float64) {
	for i := range ui.Cols {
		if ui.Cols[i].Pos == pos {
			ui.Cols[i].Min = min
			ui.Cols[i].Max = max
			return
		}
	}
	ui.Cols = append(ui.Cols, UIGridSize{Pos: pos, Min: min, Max: max})
}
func (ui *UI) SetRow(pos int, min, max float64) {
	for i := range ui.Rows {
		if ui.Rows[i].Pos == pos {
			ui.Rows[i].Min = min
			ui.Rows[i].Max = max
			return
		}
	}
	ui.Rows = append(ui.Rows, UIGridSize{Pos: pos, Min: min, Max: max})
}

func (ui *UI) SetColumnResizable(pos int, min, max, default_size float64) {
	for i := range ui.Cols {
		if ui.Cols[i].Pos == pos {
			ui.Cols[i].Min = min
			ui.Cols[i].Max = max
			ui.Cols[i].Default_resize = default_size
			return
		}
	}
	ui.Cols = append(ui.Cols, UIGridSize{Pos: pos, Min: min, Max: max, Default_resize: default_size})
}
func (ui *UI) SetRowResizable(pos int, min, max, default_size float64) {
	for i := range ui.Rows {
		if ui.Rows[i].Pos == pos {
			ui.Rows[i].Min = min
			ui.Rows[i].Max = max
			ui.Rows[i].Default_resize = default_size
			return
		}
	}
	ui.Rows = append(ui.Rows, UIGridSize{Pos: pos, Min: min, Max: max, Default_resize: default_size})
}

func (ui *UI) SetColumnFromSub(grid_y int, min_size, max_size float64) {
	newItem := UIGridSize{Pos: grid_y, SetFromChild_min: min_size, SetFromChild_max: max_size}

	for i := range ui.Cols {
		if ui.Cols[i].Pos == grid_y {
			ui.Cols[i] = newItem
			return
		}
	}
	ui.Cols = append(ui.Cols, newItem)
}

func (ui *UI) SetRowFromSub(grid_y int, min_size, max_size float64) {
	newItem := UIGridSize{Pos: grid_y, SetFromChild_min: min_size, SetFromChild_max: max_size}

	for i := range ui.Rows {
		if ui.Rows[i].Pos == grid_y {
			ui.Rows[i] = newItem
			return
		}
	}
	ui.Rows = append(ui.Rows, newItem)
}

func (ui *UI) AddText(x, y, w, h int, label string) *UIText {
	item := &UIText{Label: label, Align_h: 0, Align_v: 1, Selection: true, Formating: true, Multiline: true, Linewrapping: true, layout: _newUIItem(x, y, w, h)}
	item.layout.Text = item
	ui._addUISub(item.layout, "")
	return item
}
func (ui *UI) AddTextLabel(x, y, w, h int, value string) *UIText {
	txt := ui.AddText(x, y, w, h, "<b>"+value+"</b>")
	txt.Align_h = 1
	return txt
}

func (ui *UI) AddEditboxString(x, y, w, h int, value *string) *UIEditbox {
	item := &UIEditbox{Value: value, Align_v: 1, Formating: true, layout: _newUIItem(x, y, w, h)}
	item.layout.Editbox = item
	ui._addUISub(item.layout, "")
	return item
}
func (ui *UI) AddEditboxInt(x, y, w, h int, value *int) *UIEditbox {
	item := &UIEditbox{ValueInt: value, Align_v: 1, Formating: true, layout: _newUIItem(x, y, w, h)}
	item.layout.Editbox = item
	ui._addUISub(item.layout, "")
	return item
}
func (ui *UI) AddEditboxFloat(x, y, w, h int, value *float64, precision int) *UIEditbox {
	item := &UIEditbox{ValueFloat: value, Align_v: 1, Precision: precision, Formating: true, layout: _newUIItem(x, y, w, h)}
	item.layout.Editbox = item
	ui._addUISub(item.layout, "")
	return item
}

func (ui *UI) AddSlider(x, y, w, h int, value *float64, min, max, step float64) *UISlider {
	item := &UISlider{Value: value, Min: min, Max: max, Step: step, layout: _newUIItem(x, y, w, h)}
	item.layout.Slider = item
	ui._addUISub(item.layout, "")
	return item
}

func (ui *UI) AddButton(x, y, w, h int, label string) *UIButton {
	item := &UIButton{Label: label, Background: 1, Align: 1, layout: _newUIItem(x, y, w, h)}
	item.layout.Button = item
	ui._addUISub(item.layout, "")
	return item
}

func (ui *UI) AddYearCalendar(x, y, w, h int, Year int) *UIYearCalendar {
	item := &UIYearCalendar{Year: Year, layout: _newUIItem(x, y, w, h)}
	item.layout.YearCalendar = item
	ui._addUISub(item.layout, "")
	return item
}

func (ui *UI) AddMonthCalendar(x, y, w, h int, Year int, Month int, Events []UICalendarEvent) *UIMonthCalendar {
	item := &UIMonthCalendar{Year: Year, Month: Month, Events: Events, layout: _newUIItem(x, y, w, h)}
	item.layout.MonthCalendar = item
	ui._addUISub(item.layout, "")
	return item
}
func (ui *UI) AddDayCalendar(x, y, w, h int, Days []int64, Events []UICalendarEvent) *UIDayCalendar {
	item := &UIDayCalendar{Days: Days, Events: Events, layout: _newUIItem(x, y, w, h)}
	item.layout.DayCalendar = item
	ui._addUISub(item.layout, "")
	return item
}

func (ui *UI) AddFilePickerButton(x, y, w, h int, path *string, preview bool, onlyFolders bool) *UIFilePickerButton {
	item := &UIFilePickerButton{Path: path, Preview: preview, OnlyFolders: onlyFolders, layout: _newUIItem(x, y, w, h)}
	item.layout.FilePickerButton = item
	ui._addUISub(item.layout, "")
	return item
}
func (ui *UI) AddDatePickerButton(x, y, w, h int, date *int64, page *int64, showTime bool) *UIDatePickerButton {
	item := &UIDatePickerButton{Date: date, Page: page, ShowTime: showTime, layout: _newUIItem(x, y, w, h)}
	item.layout.DatePickerButton = item
	ui._addUISub(item.layout, "")
	return item
}
func (ui *UI) AddColorPickerButton(x, y, w, h int, cd *color.RGBA) *UIColorPickerButton {
	item := &UIColorPickerButton{Cd: cd, layout: _newUIItem(x, y, w, h)}
	item.layout.ColorPickerButton = item
	ui._addUISub(item.layout, "")
	return item
}

func (ui *UI) AddCombo(x, y, w, h int, value *string, labels []string, values []string) *UICombo {
	item := &UICombo{Value: value, Labels: labels, Values: values, DialogWidth: 5, layout: _newUIItem(x, y, w, h)}
	item.layout.Combo = item
	ui._addUISub(item.layout, "")
	return item
}

func (ui *UI) AddSwitch(x, y, w, h int, label string, value *bool) *UISwitch {
	item := &UISwitch{Label: label, Value: value, layout: _newUIItem(x, y, w, h)}
	item.layout.Switch = item
	ui._addUISub(item.layout, "")
	return item
}
func (ui *UI) AddCheckbox(x, y, w, h int, label string, value *float64) *UICheckbox {
	item := &UICheckbox{Label: label, Value: value, layout: _newUIItem(x, y, w, h)}
	item.layout.Checkbox = item
	ui._addUISub(item.layout, "")
	return item
}
func (ui *UI) AddDivider(x, y, w, h int, horizontal bool) *UIDivider {
	item := &UIDivider{Horizontal: horizontal, layout: _newUIItem(x, y, w, h)}
	item.layout.Divider = item
	ui._addUISub(item.layout, "")
	return item
}

func (ui *UI) AddOsmMap(x, y, w, h int, lon, lat, zoom *float64) *UIOsmMap {
	item := &UIOsmMap{Lon: lon, Lat: lat, Zoom: zoom, layout: _newUIItem(x, y, w, h)}
	item.layout.OsmMap = item
	ui._addUISub(item.layout, "")
	return item
}
func (mp *UIOsmMap) AddLocators(loc UIOsmMapLocators) {
	mp.Locators = append(mp.Locators, loc)
}
func (mp *UIOsmMap) AddRoute(route UIOsmMapRoute) {
	mp.Routes = append(mp.Routes, route)
}

func (ui *UI) AddChartLines(x, y, w, h int, Lines []UIChartLine) *UIChartLines {
	item := &UIChartLines{Lines: Lines, Point_rad: 0.2, Line_thick: 0.03, layout: _newUIItem(x, y, w, h)}
	item.layout.ChartLines = item
	ui._addUISub(item.layout, "")
	return item
}
func (ui *UI) AddChartColumns(x, y, w, h int, columns []UIChartColumn, x_labels []string) *UIChartColumns {
	item := &UIChartColumns{Columns: columns, X_Labels: x_labels, ColumnMargin: 0.2, layout: _newUIItem(x, y, w, h)}
	item.layout.ChartColumns = item
	ui._addUISub(item.layout, "")
	return item
}

func (ui *UI) _addImage(x, y, w, h int, path string, blob []byte, cd color.RGBA) *UIImage {
	item := &UIImage{Path: path, Blob: blob, Align_h: 1, Align_v: 1, Margin: 0.1, Cd: cd, layout: _newUIItem(x, y, w, h)}
	item.layout.Image = item
	ui._addUISub(item.layout, "")
	return item
}
func (ui *UI) AddImagePath(x, y, w, h int, path string) *UIImage {
	return ui._addImage(x, y, w, h, path, nil, color.RGBA{255, 255, 255, 255})
}
func (ui *UI) AddImageBlob(x, y, w, h int, blob []byte) *UIImage {
	return ui._addImage(x, y, w, h, "", blob, color.RGBA{255, 255, 255, 255})
}

func (ui *UI) AddLayoutList(x, y, w, h int, autoSpacing bool) *UIList {
	item := &UIList{AutoSpacing: autoSpacing, layout: _newUIItem(x, y, w, h)}
	item.layout.List = item

	ui._addUISub(item.layout, "")
	return item
}
func (list *UIList) AddItem() *UI {
	item := _newUIItem(0, len(list.layout.Items), 1, 1)
	list.layout._addUISub(item, "")
	return item
}

func (ui *UI) AddLayout(x, y, w, h int) *UI {
	item := _newUIItem(x, y, w, h)
	ui._addUISub(item, "")
	return item
}
func (ui *UI) AddLayoutWithName(x, y, w, h int, name string) *UI {
	item := _newUIItem(x, y, w, h)
	ui._addUISub(item, name)
	return item
}

func (ui *UI) FindDialog(name string) *UIDialog {
	for _, dia := range ui.Dialogs {
		if dia.UID == name {
			return dia
		}
	}
	return nil
}
func (ui *UI) AddDialog(uid string) *UIDialog {
	dia := ui.FindDialog(uid)
	if dia == nil {
		dia = &UIDialog{UID: uid, UI: *_newUIItem(0, 0, 0, 0)}
		ui.Dialogs = append(ui.Dialogs, dia)
		dia.UI._computeUID(ui, uid)
	} else {
		fmt.Println("Dialog already exist")
	}
	return dia
}

func (ui *UI) AddDialogBorder(name string, title string) (*UIDialog, *UI) {
	dia := ui.AddDialog(name)
	lay := dia.UI
	lay.SetColumnFromSub(1, 1, 100)
	lay.SetRowFromSub(1, 1, 100)
	lay.SetColumn(2, 1, 1)
	lay.SetRow(2, 1, 1)

	tx := lay.AddText(0, 0, 3, 1, title)
	tx.Align_h = 1

	return dia, lay.AddLayout(1, 1, 1, 1)
}

func (ui *UI) AddTool(x, y, w, h int, fnRun func(caller *ToolCaller, ui *UI) error, caller *ToolCaller) (*UI, error) {
	return ui._addTool(x, y, w, h, "", fnRun, caller)
}

func (ui *UI) AddToolByName(x, y, w, h int, funcName string, jsParams []byte, caller *ToolCaller) (*UI, interface{}, error) {
	fnRun, st := FindToolRunFunc(funcName, jsParams)
	out_ui, out_err := ui._addTool(x, y, w, h, funcName, fnRun, caller)
	return out_ui, st, out_err
}

func (ui *UI) Paint_Rect(x, y, w, h float64, cd, cd_over, cd_down color.RGBA, width float64) {
	ui.Paint = append(ui.Paint, UIPaint{Rectangle: &UIPaint_Rect{
		Cd: cd, Cd_over: cd, Cd_down: cd,
		Width: width,
		X:     x,
		Y:     y,
		W:     w,
		H:     h,
	}})
}

func (ui *UI) Paint_CircleOnPos(x, y float64, rad float64, cd, cd_over, cd_down color.RGBA, width float64) {
	ui.Paint = append(ui.Paint, UIPaint{Circle: &UIPaint_Circle{
		Cd: cd, Cd_over: cd_over, Cd_down: cd_down,
		Rad:   rad,
		Width: width,
		X:     x,
		Y:     y,
	}})
}

func (ui *UI) Paint_Line(sx, sy, ex, ey float64, cd color.RGBA, width float64) {
	ui.Paint = append(ui.Paint, UIPaint{Line: &UIPaint_Line{
		Cd: cd, Cd_over: cd, Cd_down: cd,
		Width: width,
		Sx:    sx,
		Sy:    sy,
		Ex:    ex,
		Ey:    ey,
	}})
}

func (ui *UI) Paint_Text(sx, sy, ex, ey float64, Text string, cd color.RGBA, Align_v int, Align_h int, Formating bool, Multiline bool, Linewrapping bool) {
	ui.Paint = append(ui.Paint, UIPaint{Text: &UIPaint_Text{
		Cd: cd, Cd_over: cd, Cd_down: cd,
		Text:         Text,
		Align_v:      Align_v,
		Align_h:      Align_h,
		Formating:    Formating,
		Multiline:    Multiline,
		Linewrapping: Linewrapping,
		Sx:           sx,
		Sy:           sy,
		Ex:           ex,
		Ey:           ey,
	}})
}

func (ui *UI) Paint_Brush(cd color.RGBA, pts []UIPaintBrushPoint) {
	ui.Paint = append(ui.Paint, UIPaint{Brush: &UIPaint_Brush{
		Cd:     cd,
		Points: pts,
	}})
}
