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
	"encoding/json"
	"fmt"
	"image/color"
	"strconv"
	"time"
)

type UIText struct {
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

	EnableCodeFormating bool
}
type UIEditbox struct {
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

	AutoSave bool
}
type UISlider struct {
	Error string
	Label string
	Value *float64
	Min   float64
	Max   float64
	Step  float64
}
type UIButton struct {
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

	Drag_group              string
	Drop_group              string
	Drag_source             string
	Drag_index              int
	Drop_h, Drop_v, Drop_in bool

	ConfirmQuestion string

	StartStopMode bool
}

type UIFilePickerButton struct {
	Error       string
	Path        *string
	Preview     bool
	OnlyFolders bool
}
type UIDatePickerButton struct {
	Error    string
	Date     *int64
	Page     *int64
	ShowTime bool
}
type UIColorPickerButton struct {
	Error string
	Cd    *color.RGBA
}

type UICombo struct {
	Error       string
	Value       *string
	Labels      []string
	Values      []string
	DialogWidth float64
}
type UISwitch struct {
	Error   string
	Label   string
	Tooltip string
	Value   *bool
}

type UICheckbox struct {
	Error   string
	Label   string
	Tooltip string
	Value   *float64
}

type UIDivider struct {
	Horizontal bool
}

type UIOsmMap struct {
	Lon, Lat, Zoom *float64
	Locators       []OsmMapLocators
	Routes         []OsmMapRoute
}

type UIChartLines struct {
	Lines []ChartLine

	X_unit, Y_unit        string
	Bound_x0, Bound_y0    bool
	Point_rad, Line_thick float64
	Draw_XHelpLines       bool
	Draw_YHelpLines       bool
}

type UIChartColumns struct {
	X_unit, Y_unit string
	Bound_y0       bool
	Y_as_time      bool
	Columns        []ChartColumn
	X_Labels       []string
	ColumnMargin   float64
}

type UIImage struct {
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
	AutoSpacing bool
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
type UIPaintBrush struct {
	Cd     color.RGBA
	Points []OsV2
}
type UIPaint struct {
	Rectangle *UIPaint_Rect   `json:",omitempty"`
	Circle    *UIPaint_Circle `json:",omitempty"`
	Line      *UIPaint_Line   `json:",omitempty"`
	Text      *UIPaint_Text   `json:",omitempty"`
	Brush     *UIPaintBrush   `json:",omitempty"`
}
type UIGridSize struct {
	Pos int
	Min float64
	Max float64

	Default_resize float64

	SetFromChild_min float64 `json:",omitempty"`
	SetFromChild_max float64 `json:",omitempty"`
}

type UIYearCalendar struct {
	Year int
}
type UIMonthCalendar struct {
	Year  int
	Month int //1=January, 2=February, etc.

	Events []CalendarEvent
}
type UIDayCalendar struct {
	Days   []int64
	Events []CalendarEvent
}

type UIDialog struct {
	UID string
	UI  UI
}
type UI struct {
	AppName  string
	FuncName string

	UID        uint64
	X, Y, W, H int
	LLMTip     string

	Cols  []UIGridSize
	Rows  []UIGridSize
	Items []*UI

	Dialogs []*UIDialog

	Enable        bool
	EnableTouch   bool
	Back_cd       color.RGBA
	Back_margin   float64
	Back_rounding bool
	Border_cd     color.RGBA
	ScrollV       LayoutScroll
	ScrollH       LayoutScroll

	List              *UIList
	Text              *UIText
	Editbox           *UIEditbox
	Button            *UIButton
	Slider            *UISlider
	FilePickerButton  *UIFilePickerButton
	DatePickerButton  *UIDatePickerButton
	ColorPickerButton *UIColorPickerButton
	Combo             *UICombo
	Switch            *UISwitch
	Checkbox          *UICheckbox
	Divider           *UIDivider
	OsmMap            *UIOsmMap
	ChartLines        *UIChartLines
	ChartColumns      *UIChartColumns
	Image             *UIImage
	YearCalendar      *UIYearCalendar
	MonthCalendar     *UIMonthCalendar
	DayCalendar       *UIDayCalendar

	Paint []UIPaint
}

func (ui *UI) Is() bool {
	return len(ui.Items) > 0
}

func (ui *UI) addDialogs(layout *Layout, appName string, funcName string, parent_UID uint64, fnProgress func(cmdsJs [][]byte, err error, start_time float64), fnDone func(dataJs []byte, uiJs []byte, cmdsJs []byte, err error, start_time float64)) {
	for _, dia := range ui.Dialogs {
		d := layout.AddDialog(dia.UID)
		dia.UI.addLayout(d.Layout, appName, funcName, parent_UID, fnProgress, fnDone)
	}
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

func (cmd *ToolCmd) Exe(ui *Ui) bool {
	found := false

	if cmd.Dialog_Open_UID > 0 {

		pos := OsV2{}
		if cmd.Dialog_OnTouch {
			pos = ui.GetWin().io.Touch.Pos
		}

		ui.settings.OpenDialog(cmd.Dialog_Open_UID, cmd.Dialog_Relative_UID, pos)
		found = true
	}

	if cmd.Dialog_Close_UID > 0 {
		dd := ui.settings.FindDialog(cmd.Dialog_Close_UID)
		if dd != nil {
			ui.settings.CloseDialog(dd)
			found = true
		}
	}

	if cmd.Editbox_Activate != "" {
		editDom := ui.mainLayout.FindEditbox(cmd.Editbox_Activate)
		if editDom != nil {
			ui.edit.SetActivate(editDom.UID)
			found = true
		}
	}

	if cmd.VScrollToTheTop > 0 {
		dom := ui.mainLayout.FindUID(cmd.VScrollToTheTop)
		if dom != nil {
			dom.VScrollToTheTop()
			found = true
		}
	}
	if cmd.VScrollToTheBottom > 0 {
		dom := ui.mainLayout.FindUID(cmd.VScrollToTheBottom)
		if dom != nil {
			dom.VScrollToTheBottom()
			found = true
		}
	}
	if cmd.VScrollToTheBottomIf > 0 {
		dom := ui.mainLayout.FindUID(cmd.VScrollToTheBottomIf)
		if dom != nil {
			dom.VScrollToTheBottomIf()
			found = true
		}
	}

	if cmd.HScrollToTheLeft > 0 {
		dom := ui.mainLayout.FindUID(cmd.HScrollToTheLeft)
		if dom != nil {
			dom.HScrollToTheLeft()
			found = true
		}
	}
	if cmd.HScrollToTheRight > 0 {
		dom := ui.mainLayout.FindUID(cmd.HScrollToTheRight)
		if dom != nil {
			dom.HScrollToTheRight()
			found = true
		}
	}

	if cmd.SetClipboardText != "" {
		ui.GetWin().SetClipboardText(cmd.SetClipboardText)
	}

	return found
}

func (ui *UI) addLayout(layout *Layout, appName string, funcName string, parent_UID uint64, fnProgress func(cmds [][]byte, err error, start_time float64), fnDone func(dataJs []byte, uiJs []byte, cmdsJs []byte, err error, start_time float64)) {

	if layout.UID != parent_UID {
		layout.setLayoutFromUI(ui)
	}

	if ui.AppName != "" {
		pre_appName := appName
		pre_fnDone := fnDone
		pre_parent_UID := parent_UID
		fnDone = func(dataJs []byte, uiJs []byte, cmdsJs []byte, err error, start_time float64) {
			if err != nil {
				return
			}
			var cmds []ToolCmd
			err = json.Unmarshal(cmdsJs, &cmds)
			if err == nil {
				layout.ui.temp_cmds = append(layout.ui.temp_cmds, cmds...)
			}
			layout.ui.router.CallChangeAsync(pre_parent_UID, pre_appName, "back", ToolsSdkChange{UID: ui.UID, ValueBytes: dataJs}, fnProgress, pre_fnDone)
		}
		appName = ui.AppName
		funcName = ui.FuncName
		parent_UID = ui.UID
	}

	for _, col := range ui.Cols {
		if col.SetFromChild_min > 0 || col.SetFromChild_max > 0 {
			layout.SetColumnFromSub(col.Pos, col.SetFromChild_min, col.SetFromChild_max)
		} else {
			if col.Default_resize > 0 {
				layout.SetColumnResizable(col.Pos, col.Min, col.Max, col.Default_resize)
			} else {
				layout.SetColumn(col.Pos, col.Min, col.Max)
			}
		}

	}
	for _, row := range ui.Rows {
		if row.SetFromChild_min > 0 || row.SetFromChild_max > 0 {
			layout.SetRowFromSub(row.Pos, row.SetFromChild_min, row.SetFromChild_max)
		} else {
			if row.Default_resize > 0 {
				layout.SetRowResizable(row.Pos, row.Min, row.Max, row.Default_resize)
			} else {
				layout.SetRow(row.Pos, row.Min, row.Max)
			}

		}
	}

	//SubItems
	for _, it := range ui.Items {

		glay := layout
		gx, gy, gw, gh := it.X, it.Y, it.W, it.H

		var itemLLMTip string //it.Label

		if it.List != nil {
			list := layout.AddLayoutList(it.X, it.Y, it.W, it.H, it.List.AutoSpacing)
			for _, itt := range it.Items {
				listItem := list.AddListSubItem()
				itt.addLayout(listItem, appName, funcName, parent_UID, fnProgress, fnDone)
			}
		} else if it.Text != nil {

			label := it.Text.Label
			if it.Text.EnableCodeFormating {
				label = _UiText_FormatAsCode(label, layout.GetPalette())
			}

			tx := layout.AddText(it.X, it.Y, it.W, it.H, label)
			tx.Align_h = it.Text.Align_h
			tx.Align_v = it.Text.Align_v
			tx.Tooltip = it.Text.Tooltip
			tx.Multiline = it.Text.Multiline
			tx.Linewrapping = it.Text.Linewrapping
			tx.Formating = it.Text.Formating
			tx.Selection = it.Text.Selection
			itemLLMTip = it.Text.Label

			if it.Text.EnableDropFile {
				txLay := glay.FindGrid(it.X, it.Y, it.W, it.H)
				txLay.dropFile = func(path string) {
					pathes := []string{path}
					pathesJs, err := json.Marshal(pathes)
					if err == nil {
						layout.ui.router.CallChangeAsync(parent_UID, appName, funcName, ToolsSdkChange{UID: it.UID, ValueBytes: pathesJs}, fnProgress, fnDone)
					}
				}
			}

			tx.Cd = it.Text.Cd
			if tx.Cd.A == 0 {
				tx.Cd = layout.GetPalette().OnB
			}

		} else if it.Editbox != nil {
			var val interface{}
			if it.Editbox.Value != nil {
				val = it.Editbox.Value
				itemLLMTip = *it.Editbox.Value
			} else if it.Editbox.ValueInt != nil {
				val = it.Editbox.ValueInt
				itemLLMTip = strconv.Itoa(*it.Editbox.ValueInt)
			} else if it.Editbox.ValueFloat != nil {
				val = it.Editbox.ValueFloat
				itemLLMTip = strconv.FormatFloat(*it.Editbox.ValueFloat, 'f', -1, 64)
			}

			var ed *Editbox
			var edLay *Layout
			if it.Editbox.Error != "" {
				errDiv := layout.AddLayout(it.X, it.Y, it.W, it.H)
				errDiv.SetColumn(0, 1, 100)
				errDiv.SetColumn(1, 1, 4)

				ed, edLay = errDiv.AddEditbox2(0, 0, 1, 1, val)
				glay = errDiv
				gx, gy, gw, gh = 0, 0, 1, 1

				//error
				tx := errDiv.AddText(1, 0, 1, 1, it.Editbox.Error)
				tx.Cd = layout.GetPalette().E
			} else {
				ed, edLay = layout.AddEditbox2(it.X, it.Y, it.W, it.H, val)
			}
			ed.Align_h = it.Editbox.Align_h
			ed.Align_v = it.Editbox.Align_v
			ed.ValueFloatPrec = it.Editbox.Precision
			ed.Ghost = it.Editbox.Ghost
			ed.Tooltip = it.Editbox.Tooltip
			ed.Password = it.Editbox.Password
			ed.Multiline = it.Editbox.Multiline
			ed.Linewrapping = it.Editbox.Linewrapping
			ed.Formating = it.Editbox.Formating
			ed.Refresh = it.Editbox.AutoSave

			createChange := func() ToolsSdkChange {
				change := ToolsSdkChange{UID: it.UID}
				if it.Editbox.Value != nil {
					change.ValueString = *it.Editbox.Value
				} else if it.Editbox.ValueInt != nil {
					change.ValueInt = int64(*it.Editbox.ValueInt)
				} else if it.Editbox.ValueFloat != nil {
					change.ValueFloat = *it.Editbox.ValueFloat
				}
				return change
			}

			ed.changed = func() {
				change := createChange()
				layout.ui.router.CallChangeAsync(parent_UID, appName, funcName, change, fnProgress, fnDone)
			}

			ed.enter = func() {
				change := createChange()
				change.ValueBool = true
				layout.ui.router.CallChangeAsync(parent_UID, appName, funcName, change, fnProgress, fnDone)
			}

			edLay.Editbox_name = it.Editbox.Name

		} else if it.Slider != nil {
			var sl *Slider
			if it.Slider.Error != "" {
				errDiv := layout.AddLayout(it.X, it.Y, it.W, it.H)
				errDiv.SetColumn(0, 1, 100)
				errDiv.SetColumn(1, 1, 4)

				sl = errDiv.AddSlider(0, 0, 1, 1, it.Slider.Value, it.Slider.Min, it.Slider.Max, it.Slider.Step)
				glay = errDiv
				gx, gy, gw, gh = 0, 0, 1, 1

				//error
				tx := errDiv.AddText(1, 0, 1, 1, it.Slider.Error)
				tx.Cd = layout.GetPalette().E
			} else {
				sl = layout.AddSlider(it.X, it.Y, it.W, it.H, it.Slider.Value, it.Slider.Min, it.Slider.Max, it.Slider.Step)
			}
			itemLLMTip = strconv.FormatFloat(*it.Slider.Value, 'f', -1, 64)
			sl.changed = func() {
				layout.ui.router.CallChangeAsync(parent_UID, appName, funcName, ToolsSdkChange{UID: it.UID, ValueFloat: *it.Slider.Value}, fnProgress, fnDone)
			}

		} else if it.FilePickerButton != nil {
			var bt *FilePickerButton
			if it.FilePickerButton.Error != "" {
				errDiv := layout.AddLayout(it.X, it.Y, it.W, it.H)
				errDiv.SetColumn(0, 1, 100)
				errDiv.SetColumn(1, 1, 4)

				bt = errDiv.AddFilePickerButton(0, 0, 1, 1, it.FilePickerButton.Path, it.FilePickerButton.Preview, it.FilePickerButton.OnlyFolders)
				glay = errDiv
				gx, gy, gw, gh = 0, 0, 1, 1

				//error
				tx := errDiv.AddText(1, 0, 1, 1, it.FilePickerButton.Error)
				tx.Cd = layout.GetPalette().E
			} else {
				bt = layout.AddFilePickerButton(it.X, it.Y, it.W, it.H, it.FilePickerButton.Path, it.FilePickerButton.Preview, it.FilePickerButton.OnlyFolders)
			}
			if it.FilePickerButton.Path != nil {
				itemLLMTip = "File: " + *it.FilePickerButton.Path
			}
			bt.changed = func() {
				layout.ui.router.CallChangeAsync(parent_UID, appName, funcName, ToolsSdkChange{UID: it.UID, ValueString: *it.FilePickerButton.Path}, fnProgress, fnDone)
			}
		} else if it.DatePickerButton != nil {
			var bt *DatePickerButton
			if it.DatePickerButton.Error != "" {
				errDiv := layout.AddLayout(it.X, it.Y, it.W, it.H)
				errDiv.SetColumn(0, 1, 100)
				errDiv.SetColumn(1, 1, 4)

				bt = errDiv.AddDatePickerButton(0, 0, 1, 1, it.DatePickerButton.Date, it.DatePickerButton.Page, it.DatePickerButton.ShowTime)
				glay = errDiv
				gx, gy, gw, gh = 0, 0, 1, 1

				//error
				tx := errDiv.AddText(1, 0, 1, 1, it.DatePickerButton.Error)
				tx.Cd = layout.GetPalette().E
			} else {
				bt = layout.AddDatePickerButton(it.X, it.Y, it.W, it.H, it.DatePickerButton.Date, it.DatePickerButton.Page, it.DatePickerButton.ShowTime)
			}
			if it.DatePickerButton.Date != nil {
				itemLLMTip = "Date(YY-MM-DD HH:MM): " + time.Unix(*it.DatePickerButton.Date, 0).Format("01-02-2006 15:04")
			}
			bt.changed = func() {
				layout.ui.router.CallChangeAsync(parent_UID, appName, funcName, ToolsSdkChange{UID: it.UID, ValueInt: *it.DatePickerButton.Date}, fnProgress, fnDone)
			}

		} else if it.ColorPickerButton != nil {
			var bt *ColorPickerButton
			if it.ColorPickerButton.Error != "" {
				errDiv := layout.AddLayout(it.X, it.Y, it.W, it.H)
				errDiv.SetColumn(0, 1, 100)
				errDiv.SetColumn(1, 1, 4)

				bt = errDiv.AddColorPickerButton(0, 0, 1, 1, it.ColorPickerButton.Cd)
				glay = errDiv
				gx, gy, gw, gh = 0, 0, 1, 1

				//error
				tx := errDiv.AddText(1, 0, 1, 1, it.ColorPickerButton.Error)
				tx.Cd = layout.GetPalette().E
			} else {
				bt = layout.AddColorPickerButton(it.X, it.Y, it.W, it.H, it.ColorPickerButton.Cd)
			}
			itemLLMTip = fmt.Sprintf("Color: R:%d, G:%d, B:%d, A:%d", it.ColorPickerButton.Cd.R, it.ColorPickerButton.Cd.G, it.ColorPickerButton.Cd.B, it.ColorPickerButton.Cd.A)
			bt.changed = func() {
				layout.ui.router.CallChangeAsync(parent_UID, appName, funcName, ToolsSdkChange{UID: it.UID, ValueString: fmt.Sprintf("%d %d %d %d", it.ColorPickerButton.Cd.R, it.ColorPickerButton.Cd.G, it.ColorPickerButton.Cd.B, it.ColorPickerButton.Cd.A)}, fnProgress, fnDone)
			}

		} else if it.Combo != nil {
			var cb *Combo
			if it.Combo.Error != "" {
				errDiv := layout.AddLayout(it.X, it.Y, it.W, it.H)
				errDiv.SetColumn(0, 1, 100)
				errDiv.SetColumn(1, 1, 4)

				cb = errDiv.AddCombo(it.X, it.Y, it.W, it.H, it.Combo.Value, it.Combo.Labels, it.Combo.Values)
				cb.DialogWidth = it.Combo.DialogWidth

				glay = errDiv
				gx, gy, gw, gh = 0, 0, 1, 1

				//error
				tx := errDiv.AddText(1, 0, 1, 1, it.Combo.Error)
				tx.Cd = layout.GetPalette().E
			} else {
				cb = layout.AddCombo(it.X, it.Y, it.W, it.H, it.Combo.Value, it.Combo.Labels, it.Combo.Values)
				cb.DialogWidth = it.Combo.DialogWidth
			}
			cb.changed = func() {
				layout.ui.router.CallChangeAsync(parent_UID, appName, funcName, ToolsSdkChange{UID: it.UID, ValueString: *it.Combo.Value}, fnProgress, fnDone)
			}

			if it.Combo.Value != nil {
				itemLLMTip = fmt.Sprintf("Value: %s, Label: %s", *it.Combo.Value, Combo_getValueLabel(*it.Combo.Value, it.Combo.Labels, it.Combo.Values))
			}
		} else if it.Switch != nil {
			var sw *Switch
			if it.Switch.Error != "" {
				errDiv := layout.AddLayout(it.X, it.Y, it.W, it.H)
				errDiv.SetColumn(0, 1, 100)
				errDiv.SetColumn(1, 1, 4)

				sw = errDiv.AddSwitch(it.X, it.Y, it.W, it.H, it.Switch.Label, it.Switch.Value)
				sw.Tooltip = it.Switch.Tooltip

				glay = errDiv
				gx, gy, gw, gh = 0, 0, 1, 1

				//error
				tx := errDiv.AddText(1, 0, 1, 1, it.Switch.Error)
				tx.Cd = layout.GetPalette().E
			} else {
				sw = layout.AddSwitch(it.X, it.Y, it.W, it.H, it.Switch.Label, it.Switch.Value)
				sw.Tooltip = it.Switch.Tooltip
			}
			sw.changed = func() {
				layout.ui.router.CallChangeAsync(parent_UID, appName, funcName, ToolsSdkChange{UID: it.UID, ValueBool: *it.Switch.Value}, fnProgress, fnDone)
			}

			if it.Switch.Value != nil {
				itemLLMTip = fmt.Sprintf("Value: %v, Label: %s", *it.Switch.Value, it.Switch.Label)
			}
		} else if it.Checkbox != nil {
			var che *Checkbox
			if it.Checkbox.Error != "" {
				errDiv := layout.AddLayout(it.X, it.Y, it.W, it.H)
				errDiv.SetColumn(0, 1, 100)
				errDiv.SetColumn(1, 1, 4)

				che = errDiv.AddCheckbox(it.X, it.Y, it.W, it.H, it.Checkbox.Label, it.Checkbox.Value)
				che.Tooltip = it.Checkbox.Tooltip

				glay = errDiv
				gx, gy, gw, gh = 0, 0, 1, 1

				//error
				tx := errDiv.AddText(1, 0, 1, 1, it.Checkbox.Error)
				tx.Cd = layout.GetPalette().E
			} else {
				che = layout.AddCheckbox(it.X, it.Y, it.W, it.H, it.Checkbox.Label, it.Checkbox.Value)
				che.Tooltip = it.Checkbox.Tooltip
			}
			che.changed = func() {
				layout.ui.router.CallChangeAsync(parent_UID, appName, funcName, ToolsSdkChange{UID: it.UID, ValueFloat: *it.Checkbox.Value}, fnProgress, fnDone)
			}

			if it.Checkbox.Value != nil {
				itemLLMTip = fmt.Sprintf("Value: %f, Label: %s", *it.Checkbox.Value, it.Checkbox.Label)
			}
		} else if it.Divider != nil {
			layout.AddDivider(it.X, it.Y, it.W, it.H, it.Divider.Horizontal)

		} else if it.OsmMap != nil {
			mp := layout.AddOsmMap(it.X, it.Y, it.W, it.H, &OsmMapCam{Lon: *it.OsmMap.Lon, Lat: *it.OsmMap.Lat, Zoom: *it.OsmMap.Zoom})
			mp.Locators = it.OsmMap.Locators
			mp.Routes = it.OsmMap.Routes

			mp.changed = func() {
				layout.ui.router.CallChangeAsync(parent_UID, appName, funcName, ToolsSdkChange{UID: it.UID, ValueString: fmt.Sprintf("%f %f %f", mp.Cam.Lon, mp.Cam.Lat, mp.Cam.Zoom)}, fnProgress, fnDone)
			}

			//itemLLMTip ....
		} else if it.ChartLines != nil {
			ch := layout.AddChartLines(it.X, it.Y, it.W, it.H, it.ChartLines.Lines)

			ch.X_unit = it.ChartLines.X_unit
			ch.Y_unit = it.ChartLines.Y_unit
			ch.Bound_x0 = it.ChartLines.Bound_x0
			ch.Bound_y0 = it.ChartLines.Bound_y0
			ch.Point_rad = it.ChartLines.Point_rad
			ch.Line_thick = it.ChartLines.Line_thick
			ch.Draw_XHelpLines = it.ChartLines.Draw_XHelpLines
			ch.Draw_YHelpLines = it.ChartLines.Draw_YHelpLines

			//itemLLMTip ....
		} else if it.ChartColumns != nil {
			ch := layout.AddChartColumns(it.X, it.Y, it.W, it.H, it.ChartColumns.Columns, it.ChartColumns.X_Labels)
			ch.X_unit = it.ChartColumns.X_unit
			ch.Y_unit = it.ChartColumns.Y_unit
			ch.Bound_y0 = it.ChartColumns.Bound_y0
			ch.Y_as_time = it.ChartColumns.Y_as_time
			ch.ColumnMargin = it.ChartColumns.ColumnMargin

			//itemLLMTip ....
		} else if it.Image != nil {
			img := layout.AddImageCd(it.X, it.Y, it.W, it.H, it.Image.Path, it.Image.Blob, it.Image.Cd)
			img.Tooltip = it.Image.Tooltip
			img.Draw_border = it.Image.Draw_border
			img.Margin = it.Image.Margin
			img.Align_h = it.Image.Align_h
			img.Align_v = it.Image.Align_v

			img.Translate_x = it.Image.Translate_x
			img.Translate_y = it.Image.Translate_y
			img.Scale_x = it.Image.Scale_x
			img.Scale_y = it.Image.Scale_y

			//itemLLMTip ....
		} else if it.Button != nil {

			if it.Button.ConfirmQuestion != "" {
				bt := layout.AddButtonConfirm(it.X, it.Y, it.W, it.H, it.Button.Label, it.Button.ConfirmQuestion)
				bt.confirmed = func() {
					layout.ui.router.CallChangeAsync(parent_UID, appName, funcName, ToolsSdkChange{UID: it.UID}, fnProgress, fnDone)
				}
				bt.Tooltip = it.Button.Tooltip
				bt.Align = it.Button.Align
				bt.Background = it.Button.Background
				bt.Border = it.Button.Border
				bt.IconPath = it.Button.IconPath
				bt.IconBlob = it.Button.IconBlob
				bt.Icon_align = it.Button.Icon_align
				bt.Icon_margin = it.Button.Icon_margin
				bt.Color = it.Button.Cd
				bt.Shortcut_key = it.Button.Shortcut
			} else {
				bt := layout.AddButton(it.X, it.Y, it.W, it.H, it.Button.Label)
				bt.clicked = func() {
					layout.ui.router.CallChangeAsync(parent_UID, appName, funcName, ToolsSdkChange{UID: it.UID}, fnProgress, fnDone)
				}
				bt.Tooltip = it.Button.Tooltip
				bt.Align = it.Button.Align
				bt.Background = it.Button.Background
				bt.Border = it.Button.Border
				bt.IconPath = it.Button.IconPath
				bt.IconBlob = it.Button.IconBlob
				bt.Icon_align = it.Button.Icon_align
				bt.Icon_margin = it.Button.Icon_margin
				bt.BrowserUrl = it.Button.BrowserUrl
				bt.Cd = it.Button.Cd
				bt.Shortcut_key = it.Button.Shortcut
			}

			btLay := layout.FindGrid(it.X, it.Y, it.W, it.H)

			btLay.Drag_group = it.Button.Drag_group
			btLay.Drop_group = it.Button.Drop_group
			btLay.Drag_source = it.Button.Drag_source
			btLay.Drag_index = it.Button.Drag_index
			btLay.Drop_h = it.Button.Drop_h
			btLay.Drop_v = it.Button.Drop_v
			btLay.Drop_in = it.Button.Drop_in
			btLay.dropMove = func(src_i, dst_i int, src_source, dst_source string) {
				layout.ui.router.CallChangeAsync(parent_UID, appName, funcName, ToolsSdkChange{UID: it.UID, ValueString: fmt.Sprintf("%d %d %s %s", src_i, dst_i, src_source, dst_source)}, fnProgress, fnDone)
			}

			itemLLMTip = "Button with label: " + it.Button.Label

		} else if it.YearCalendar != nil {
			layout.AddYearCalendar(it.X, it.Y, it.W, it.H, it.YearCalendar.Year)

		} else if it.MonthCalendar != nil {
			layout.AddMonthCalendar(it.X, it.Y, it.W, it.H, it.MonthCalendar.Year, it.MonthCalendar.Month, it.MonthCalendar.Events)

		} else if it.DayCalendar != nil {
			layout.AddDayCalendar(it.X, it.Y, it.W, it.H, it.DayCalendar.Days, it.DayCalendar.Events)
		} else {
			layout.AddLayout(it.X, it.Y, it.W, it.H)
		}

		it2 := glay.FindGrid(gx, gy, gw, gh)

		if it2 != nil && it2.LLMTip == "" {
			it2.LLMTip = _UiText_RemoveFormating(itemLLMTip)
		}

		it.addLayout(it2, appName, funcName, parent_UID, fnProgress, fnDone)
	}

	if len(ui.Paint) > 0 {
		layout.fnDraw = func(rect Rect, l *Layout) (paint LayoutPaint) {
			for _, pt := range ui.Paint {
				if pt.Rectangle != nil {
					rc := rect
					rc.X += rc.W * pt.Rectangle.X
					rc.Y += rc.H * pt.Rectangle.Y
					rc.W *= pt.Rectangle.W
					rc.H *= pt.Rectangle.H

					paint.Rect(rc, pt.Rectangle.Cd, pt.Rectangle.Cd_down, pt.Rectangle.Cd_over, pt.Rectangle.Width)
				}

				if pt.Circle != nil {
					paint.CircleRad(rect, pt.Circle.X, pt.Circle.Y, pt.Circle.Rad, pt.Circle.Cd, pt.Circle.Cd_down, pt.Circle.Cd_over, pt.Circle.Width)
				}

				if pt.Line != nil {
					paint.Line(rect, pt.Line.Sx, pt.Line.Sy, pt.Line.Ex, pt.Line.Ey, pt.Line.Cd, pt.Line.Width)
				}

				if pt.Text != nil {
					t := paint.Text(rect, pt.Text.Text, "", pt.Text.Cd, pt.Text.Cd_over, pt.Text.Cd_down, false, false, uint8(pt.Text.Align_h), uint8(pt.Text.Align_v))
					t.Formating = pt.Text.Formating
					t.Multiline = pt.Text.Multiline
					t.Linewrapping = pt.Text.Linewrapping
				}

				if pt.Brush != nil {
					paint.Brush(pt.Brush.Cd, pt.Brush.Points)
				}
			}
			return
		}
	}

	//must be after subs, because of relative!
	ui.addDialogs(layout, appName, funcName, parent_UID, fnProgress, fnDone)
}

func (layout *Layout) setLayoutFromUI(item *UI) {
	layout.Enable = item.Enable
	layout.EnableTouch = item.EnableTouch
	layout.Back_cd = item.Back_cd
	layout.Back_margin = item.Back_margin
	layout.Back_rounding = item.Back_rounding
	layout.Border_cd = item.Border_cd

	layout.scrollV.Show = !item.ScrollV.Hide
	layout.scrollH.Show = !item.ScrollH.Hide
	layout.scrollV.Narrow = item.ScrollV.Narrow
	layout.scrollH.Narrow = item.ScrollH.Narrow

	layout.LLMTip = item.LLMTip

	layout.UID = item.UID

}
