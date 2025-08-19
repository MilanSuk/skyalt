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
	Label   string
	Align_h int
	Align_v int
	Cd      color.RGBA

	Selection    bool
	Formating    bool
	Multiline    bool
	Linewrapping bool

	EnableDropFile bool

	EnableCodeFormating bool
}
type UIEditbox struct {
	Name       string
	Value      *string
	ValueFloat *float64
	ValueInt   *int
	Precision  int
	Ghost      string
	Password   bool

	Align_h int //0=left, 1=center, 2=right
	Align_v int //0=top, 1=center, 2=bottom

	Formating    bool
	Multiline    bool
	Linewrapping bool
}
type UISlider struct {
	Label string
	Value *float64
	Min   float64
	Max   float64
	Step  float64

	ShowLegend      bool
	ShowDrawSteps   bool
	ShowRecommend   bool
	Recommend_value float64
}
type UIButton struct {
	Label string
	Align int

	Shortcut rune

	Background  float64
	Border      bool
	IconBlob    []byte
	IconPath    string
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
	Tooltip     string
	Path        *string
	Preview     bool
	OnlyFolders bool
}
type UIDatePickerButton struct {
	Tooltip  string
	Date     *int64
	Page     *int64
	ShowTime bool
}
type UIColorPickerButton struct {
	Cd *color.RGBA
}

type UIDropDown struct {
	Value  *string
	Labels []string
	Values []string
	Icons  []DropDownIcon
}

type UIPromptMenu struct {
	Prompts []string
	Icons   []PromptMenuIcon
}

type UISwitch struct {
	Label string
	Value *bool
}

type UICheckbox struct {
	Label string
	Value *float64
}

type UIMicrophone struct {
	Shortcut                   rune
	Format                     string
	Transcribe                 bool
	Transcribe_response_format string //"verbose_json"
	Output_onlyTranscript      bool
}

type UIDivider struct {
	Horizontal bool
	Label      string
}

type UIMap struct {
	Lon, Lat, Zoom *float64
	Locators       []MapLocators
	Routes         []MapRoute
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

type UIMedia struct {
	Blob []byte
	Path string

	Cd          color.RGBA
	Draw_border bool

	Margin  float64
	Align_h int
	Align_v int

	Translate_x, Translate_y float64
	Scale_x, Scale_y         float64
}

type UICards struct {
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

	SetFromChild     bool `json:",omitempty"`
	SetFromChild_fix bool `json:",omitempty"`
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
	ToolName string

	UID        uint64
	X, Y, W, H int

	Tooltip       string
	TooltipGroups []LayoutTooltip

	Cols  []UIGridSize
	Rows  []UIGridSize
	Items []*UI

	Dialogs []*UIDialog

	Enable        bool
	EnableTouch   bool
	EnableBrush   bool
	Back_cd       color.RGBA
	Back_margin   float64
	Back_rounding bool
	Border_cd     color.RGBA
	ScrollV       LayoutScroll
	ScrollH       LayoutScroll

	Cards             *UICards
	Text              *UIText
	Editbox           *UIEditbox
	Button            *UIButton
	Slider            *UISlider
	FilePickerButton  *UIFilePickerButton
	DatePickerButton  *UIDatePickerButton
	ColorPickerButton *UIColorPickerButton
	DropDown          *UIDropDown
	PromptMenu        *UIPromptMenu
	Switch            *UISwitch
	Checkbox          *UICheckbox
	Microphone        *UIMicrophone
	Divider           *UIDivider
	Map               *UIMap
	ChartLines        *UIChartLines
	ChartColumns      *UIChartColumns
	Media             *UIMedia
	YearCalendar      *UIYearCalendar
	MonthCalendar     *UIMonthCalendar
	DayCalendar       *UIDayCalendar

	Paint []UIPaint

	HasUpdate bool

	App bool
}

func (ui *UI) Is() bool {
	return len(ui.Items) > 0
}

func (ui *UI) addDialogs(layout *Layout, appName string, toolName string, parent_UID uint64, fnProgress func(cmdsGob [][]byte, err error, start_time float64), fnDone func(dataJs []byte, uiGob []byte, cmdsGob []byte, err error, start_time float64)) {
	for _, dia := range ui.Dialogs {
		d := layout.AddDialog(dia.UID)
		dia.UI.addLayout(d.Layout, appName, toolName, parent_UID, fnProgress, fnDone)
	}
}

type ToolCmd struct {
	Dialog_Open_UID     uint64 `json:",omitempty"`
	Dialog_Relative_UID uint64 `json:",omitempty"`
	Dialog_OnTouch      bool   `json:",omitempty"`

	Dialog_Close_UID uint64 `json:",omitempty"`

	Editbox_Activate uint64 `json:",omitempty"`

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

	if cmd.Editbox_Activate > 0 {
		editDom := ui.mainLayout.FindUID(cmd.Editbox_Activate)
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

func (ui *UI) addLayout(layout *Layout, appName string, toolName string, parent_UID uint64, fnProgress func(cmds [][]byte, err error, start_time float64), fnDone func(dataJs []byte, uiGob []byte, cmdsGob []byte, err error, start_time float64)) {

	if layout.UID != parent_UID {
		layout.setLayoutFromUI(ui)
	}

	if ui.HasUpdate {
		layout.fnUpdate = func() {
			fnDone := func(dataJs []byte, uiGob []byte, cmdsGob []byte, err error, start_time float64) {
				if LogsError(err) != nil {
					return
				}

				var subUI UI
				err = LogsGobUnmarshal(uiGob, &subUI)
				if err != nil {
					return
				}

				layout.parent.addLayoutComp(&subUI, appName, toolName, parent_UID, layout.ui._addLayout_FnProgress, layout.ui._addLayout_FnIODone)

				layout._build()
				layout._relayout()
				layout._draw()
				layout.ui.SetRedrawBuffer()

				var cmds []ToolCmd
				err = LogsGobUnmarshal(cmdsGob, &cmds)
				if err != nil {
					return
				}
				layout.ui.temp_cmds = append(layout.ui.temp_cmds, cmds...)

				fmt.Printf("_updated(): %.4fsec\n", OsTime()-start_time)
			}

			layout.ui.router.CallUpdateAsync(parent_UID, layout.UID, appName, toolName, layout.ui._addLayout_FnProgress, fnDone)
		}
	}

	if ui.AppName != "" {
		pre_appName := appName
		pre_fnDone := fnDone
		pre_parent_UID := parent_UID
		fnDone = func(dataJs []byte, uiGob []byte, cmdsGob []byte, err error, start_time float64) {
			if LogsError(err) != nil {
				return
			}
			var cmds []ToolCmd
			err = LogsGobUnmarshal(cmdsGob, &cmds)
			if err != nil {
				return
			}
			layout.ui.temp_cmds = append(layout.ui.temp_cmds, cmds...)

			layout.ui.router.CallChangeAsync(pre_parent_UID, pre_appName, "back", ToolsSdkChange{UID: ui.UID, ValueBytes: dataJs}, fnProgress, pre_fnDone)
		}
		appName = ui.AppName
		toolName = ui.ToolName
		parent_UID = ui.UID

		layout.AppName = ui.AppName
		layout.ToolName = ui.ToolName
	}

	for _, col := range ui.Cols {
		if col.SetFromChild {
			layout.SetColumnFromSub(col.Pos, col.Min, col.Max, col.SetFromChild_fix)
		} else {
			if col.Default_resize > 0 {
				layout.SetColumnResizable(col.Pos, col.Min, col.Max, col.Default_resize)
			} else {
				layout.SetColumn(col.Pos, col.Min, col.Max)
			}
		}

	}
	for _, row := range ui.Rows {
		if row.SetFromChild {
			layout.SetRowFromSub(row.Pos, row.Min, row.Max, row.SetFromChild_fix)
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
		layout.addLayoutComp(it, appName, toolName, parent_UID, fnProgress, fnDone)
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
	ui.addDialogs(layout, appName, toolName, parent_UID, fnProgress, fnDone)
}

func (layout *Layout) addLayoutComp(it *UI, appName string, toolName string, parent_UID uint64, fnProgress func(cmds [][]byte, err error, start_time float64), fnDone func(dataJs []byte, uiGob []byte, cmdsGob []byte, err error, start_time float64)) {

	//var tooltip_value string

	if it.Cards != nil {
		cards := layout.AddLayoutCards(it.X, it.Y, it.W, it.H, it.Cards.AutoSpacing)
		for _, itt := range it.Items {
			cardsItem := cards.AddCardsSubItem()
			itt.addLayout(cardsItem, appName, toolName, parent_UID, fnProgress, fnDone)
		}
	} else if it.Text != nil {
		label := it.Text.Label
		if it.Text.EnableCodeFormating {
			label = _UiText_FormatAsCode(label, layout.GetPalette())
		}

		tx := layout.AddText(it.X, it.Y, it.W, it.H, label)
		tx.Tooltip = it.Tooltip
		tx.Align_h = it.Text.Align_h
		tx.Align_v = it.Text.Align_v
		tx.Multiline = it.Text.Multiline
		tx.Linewrapping = it.Text.Linewrapping
		tx.Formating = it.Text.Formating
		tx.Selection = it.Text.Selection

		if it.Text.EnableDropFile {
			txLay := layout.FindGrid(it.X, it.Y, it.W, it.H)
			txLay.dropFile = func(path string) {
				pathes := []string{path}
				pathesJs, err := LogsJsonMarshalIndent(pathes)
				if err == nil {
					layout.ui.router.CallChangeAsync(parent_UID, appName, toolName, ToolsSdkChange{UID: it.UID, ValueBytes: pathesJs}, fnProgress, fnDone)
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
			//tooltip_value = *it.Editbox.Value
		} else if it.Editbox.ValueInt != nil {
			val = it.Editbox.ValueInt
			//tooltip_value = strconv.Itoa(*it.Editbox.ValueInt)
		} else if it.Editbox.ValueFloat != nil {
			val = it.Editbox.ValueFloat
			//tooltip_value = strconv.FormatFloat(*it.Editbox.ValueFloat, 'f', -1, 64)
		} else {
			tempStr := ""
			it.Editbox.Value = &tempStr
			val = it.Editbox.Value
		}

		ed := layout.AddEditbox(it.X, it.Y, it.W, it.H, val)
		ed.Tooltip = it.Tooltip
		ed.Align_h = it.Editbox.Align_h
		ed.Align_v = it.Editbox.Align_v
		ed.ValueFloatPrec = it.Editbox.Precision
		ed.Ghost = it.Editbox.Ghost
		ed.Password = it.Editbox.Password
		ed.Multiline = it.Editbox.Multiline
		ed.Linewrapping = it.Editbox.Linewrapping
		ed.Formating = it.Editbox.Formating

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
			layout.ui.router.CallChangeAsync(parent_UID, appName, toolName, change, fnProgress, fnDone)
		}

		ed.enter = func() {
			change := createChange()
			change.ValueBool = true
			layout.ui.router.CallChangeAsync(parent_UID, appName, toolName, change, fnProgress, fnDone)
		}

	} else if it.Slider != nil {
		if it.Slider.Value == nil {
			tempVal := 0.0
			it.Slider.Value = &tempVal
		}

		sl := layout.AddSlider(it.X, it.Y, it.W, it.H, it.Slider.Value, it.Slider.Min, it.Slider.Max, it.Slider.Step)
		sl.Tooltip = it.Tooltip
		sl.ShowLegend = it.Slider.ShowLegend
		sl.ShowDrawSteps = it.Slider.ShowDrawSteps
		sl.ShowRecommend = it.Slider.ShowRecommend
		sl.Recommend_value = it.Slider.Recommend_value

		//tooltip_value = strconv.FormatFloat(*it.Slider.Value, 'f', -1, 64)
		sl.changed = func() {
			layout.ui.router.CallChangeAsync(parent_UID, appName, toolName, ToolsSdkChange{UID: it.UID, ValueFloat: *it.Slider.Value}, fnProgress, fnDone)
		}

	} else if it.FilePickerButton != nil {
		if it.FilePickerButton.Path == nil {
			tempVal := ""
			it.FilePickerButton.Path = &tempVal
		}

		bt := layout.AddFilePickerButton(it.X, it.Y, it.W, it.H, it.FilePickerButton.Path, it.FilePickerButton.Preview, it.FilePickerButton.OnlyFolders)
		bt.Tooltip = it.Tooltip
		if it.FilePickerButton.Path != nil {
			//tooltip_value = "File: " + *it.FilePickerButton.Path
		}
		bt.changed = func() {
			layout.ui.router.CallChangeAsync(parent_UID, appName, toolName, ToolsSdkChange{UID: it.UID, ValueString: *it.FilePickerButton.Path}, fnProgress, fnDone)
		}
	} else if it.DatePickerButton != nil {
		if it.DatePickerButton.Date == nil {
			tempVal := int64(0)
			it.DatePickerButton.Date = &tempVal
		}
		bt := layout.AddDatePickerButton(it.X, it.Y, it.W, it.H, it.DatePickerButton.Date, it.DatePickerButton.Page, it.DatePickerButton.ShowTime)
		bt.Tooltip = it.Tooltip
		if it.DatePickerButton.Date != nil {
			//tooltip_value = "Date(YY-MM-DD HH:MM): " + time.Unix(*it.DatePickerButton.Date, 0).Format("01-02-2006 15:04")
		}
		bt.changed = func() {
			layout.ui.router.CallChangeAsync(parent_UID, appName, toolName, ToolsSdkChange{UID: it.UID, ValueInt: *it.DatePickerButton.Date}, fnProgress, fnDone)
		}

	} else if it.ColorPickerButton != nil {
		if it.ColorPickerButton.Cd == nil {
			tempVal := color.RGBA{}
			it.ColorPickerButton.Cd = &tempVal
		}
		bt := layout.AddColorPickerButton(it.X, it.Y, it.W, it.H, it.ColorPickerButton.Cd)
		bt.Tooltip = it.Tooltip
		//tooltip_value = fmt.Sprintf("Color: R:%d, G:%d, B:%d, A:%d", it.ColorPickerButton.Cd.R, it.ColorPickerButton.Cd.G, it.ColorPickerButton.Cd.B, it.ColorPickerButton.Cd.A)
		bt.changed = func() {
			layout.ui.router.CallChangeAsync(parent_UID, appName, toolName, ToolsSdkChange{UID: it.UID, ValueString: fmt.Sprintf("%d %d %d %d", it.ColorPickerButton.Cd.R, it.ColorPickerButton.Cd.G, it.ColorPickerButton.Cd.B, it.ColorPickerButton.Cd.A)}, fnProgress, fnDone)
		}

	} else if it.DropDown != nil {
		if it.DropDown.Value == nil {
			tempVal := ""
			it.DropDown.Value = &tempVal
		}
		cb := layout.AddDropDown(it.X, it.Y, it.W, it.H, it.DropDown.Value, it.DropDown.Labels, it.DropDown.Values)
		cb.Tooltip = it.Tooltip
		cb.Icons = it.DropDown.Icons
		cb.changed = func() {
			layout.ui.router.CallChangeAsync(parent_UID, appName, toolName, ToolsSdkChange{UID: it.UID, ValueString: *it.DropDown.Value}, fnProgress, fnDone)
		}

		if it.DropDown.Value != nil {
			//tooltip_value = fmt.Sprintf("Value: %s, Label: %s", *it.DropDown.Value, DropDown_getValueLabel(*it.DropDown.Value, it.DropDown.Labels, it.DropDown.Values))
		}

	} else if it.PromptMenu != nil {
		cb := layout.AddPromptMenu(it.X, it.Y, it.W, it.H, it.PromptMenu.Prompts)
		cb.Tooltip = it.Tooltip
		cb.Icons = it.PromptMenu.Icons
		cb.changed = func() {
			layout.ui.router.CallChangeAsync(parent_UID, appName, toolName, ToolsSdkChange{UID: it.UID}, fnProgress, fnDone)
		}

	} else if it.Switch != nil {
		if it.Switch.Value == nil {
			tempVal := false
			it.Switch.Value = &tempVal
		}
		sw := layout.AddSwitch(it.X, it.Y, it.W, it.H, it.Switch.Label, it.Switch.Value)
		sw.Tooltip = it.Tooltip
		sw.changed = func() {
			layout.ui.router.CallChangeAsync(parent_UID, appName, toolName, ToolsSdkChange{UID: it.UID, ValueBool: *it.Switch.Value}, fnProgress, fnDone)
		}

		if it.Switch.Value != nil {
			//tooltip_value = fmt.Sprintf("Value: %v, Label: %s", *it.Switch.Value, it.Switch.Label)
		}
	} else if it.Checkbox != nil {
		if it.Checkbox.Value == nil {
			tempVal := 0.0
			it.Checkbox.Value = &tempVal
		}
		che := layout.AddCheckbox(it.X, it.Y, it.W, it.H, it.Checkbox.Label, it.Checkbox.Value)
		che.Tooltip = it.Tooltip
		che.changed = func() {
			layout.ui.router.CallChangeAsync(parent_UID, appName, toolName, ToolsSdkChange{UID: it.UID, ValueFloat: *it.Checkbox.Value}, fnProgress, fnDone)
		}

		if it.Checkbox.Value != nil {
			//tooltip_value = fmt.Sprintf("Value: %f, Label: %s", *it.Checkbox.Value, it.Checkbox.Label)
		}
	} else if it.Microphone != nil {
		mic := layout.AddMicrophone(it.X, it.Y, it.W, it.H)
		mic.Shortcut = it.Microphone.Shortcut
		mic.Format = it.Microphone.Format
		mic.Transcribe = it.Microphone.Transcribe
		mic.Transcribe_response_format = it.Microphone.Transcribe_response_format
		mic.Output_onlyTranscript = it.Microphone.Output_onlyTranscript

		mic.started = func() {
			layout.ui.router.CallChangeAsync(parent_UID, appName, toolName, ToolsSdkChange{UID: it.UID, ValueBool: true}, fnProgress, fnDone)
		}
		mic.finished = func(audio []byte, transcript string) {
			layout.ui.router.CallChangeAsync(parent_UID, appName, toolName, ToolsSdkChange{UID: it.UID, ValueBool: false, ValueBytes: audio, ValueString: transcript}, fnProgress, fnDone)
		}
	} else if it.Divider != nil {
		d := layout.AddDivider(it.X, it.Y, it.W, it.H, it.Divider.Horizontal)
		d.Label = it.Divider.Label

	} else if it.Map != nil {
		if it.Map.Lon == nil {
			tempVal := 0.0
			it.Map.Lon = &tempVal
		}
		if it.Map.Lat == nil {
			tempVal := 0.0
			it.Map.Lat = &tempVal
		}
		if it.Map.Zoom == nil {
			tempVal := 0.0
			it.Map.Zoom = &tempVal
		}
		mp := layout.AddMap(it.X, it.Y, it.W, it.H, &MapCam{Lon: *it.Map.Lon, Lat: *it.Map.Lat, Zoom: *it.Map.Zoom})
		mp.Tooltip = it.Tooltip
		mp.Locators = it.Map.Locators
		mp.Routes = it.Map.Routes

		mp.changed = func() {
			layout.ui.router.CallChangeAsync(parent_UID, appName, toolName, ToolsSdkChange{UID: it.UID, ValueString: fmt.Sprintf("%f %f %f", mp.Cam.Lon, mp.Cam.Lat, mp.Cam.Zoom)}, fnProgress, fnDone)
		}

		//itemLLMTip ....
	} else if it.ChartLines != nil {
		ch := layout.AddChartLines(it.X, it.Y, it.W, it.H, it.ChartLines.Lines)
		ch.Tooltip = it.Tooltip
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
		ch.Tooltip = it.Tooltip
		ch.X_unit = it.ChartColumns.X_unit
		ch.Y_unit = it.ChartColumns.Y_unit
		ch.Bound_y0 = it.ChartColumns.Bound_y0
		ch.Y_as_time = it.ChartColumns.Y_as_time
		ch.ColumnMargin = it.ChartColumns.ColumnMargin

		//itemLLMTip ....
	} else if it.Media != nil {

		tp, _ := layout.ui.router.services.media.Type(it.Media.Path)

		switch tp {
		case 0:
			img := layout.AddImage(it.X, it.Y, it.W, it.H, it.Media.Path, it.Media.Blob)
			img.Cd = it.Media.Cd
			img.Draw_border = it.Media.Draw_border
			img.Margin = it.Media.Margin
			img.Align_h = it.Media.Align_h
			img.Align_v = it.Media.Align_v

			img.Translate_x = it.Media.Translate_x
			img.Translate_y = it.Media.Translate_y
			img.Scale_x = it.Media.Scale_x
			img.Scale_y = it.Media.Scale_y
		case 1:
			vid := layout.AddVideo(it.X, it.Y, it.W, it.H, it.Media.Path)
			vid.Cd = it.Media.Cd
			vid.Draw_border = it.Media.Draw_border
			vid.Margin = it.Media.Margin
			vid.Align_h = it.Media.Align_h
			vid.Align_v = it.Media.Align_v
		case 2:
			layout.AddAudio(it.X, it.Y, it.W, it.H, it.Media.Path)
		}

		//itemLLMTip ....
	} else if it.Button != nil {

		if it.Button.ConfirmQuestion != "" {
			bt := layout.AddButtonConfirm(it.X, it.Y, it.W, it.H, it.Button.Label, it.Button.ConfirmQuestion)
			bt.confirmed = func() {
				layout.ui.router.CallChangeAsync(parent_UID, appName, toolName, ToolsSdkChange{UID: it.UID}, fnProgress, fnDone)
			}
			bt.Tooltip = it.Tooltip
			bt.Align = it.Button.Align
			bt.Background = it.Button.Background
			bt.Border = it.Button.Border
			bt.IconPath = it.Button.IconPath
			bt.IconBlob = it.Button.IconBlob
			bt.Icon_margin = it.Button.Icon_margin
			bt.Color = it.Button.Cd
			bt.Shortcut_key = it.Button.Shortcut
		} else {
			bt := layout.AddButton(it.X, it.Y, it.W, it.H, it.Button.Label)
			bt.clicked = func() {
				layout.ui.router.CallChangeAsync(parent_UID, appName, toolName, ToolsSdkChange{UID: it.UID}, fnProgress, fnDone)
			}
			bt.Tooltip = it.Tooltip
			bt.Align = it.Button.Align
			bt.Background = it.Button.Background
			bt.Border = it.Button.Border
			bt.IconPath = it.Button.IconPath
			bt.IconBlob = it.Button.IconBlob
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
		btLay.dropMove = func(src_i, dst_i int, aim_i int, src_source, dst_source string) {
			layout.ui.router.CallChangeAsync(parent_UID, appName, toolName, ToolsSdkChange{UID: it.UID, ValueString: fmt.Sprintf("%d %d %d %s %s", src_i, dst_i, aim_i, src_source, dst_source)}, fnProgress, fnDone)
		}

		//tooltip_value = "Button with label: " + it.Button.Label

	} else if it.YearCalendar != nil {
		layout.AddYearCalendar(it.X, it.Y, it.W, it.H, it.YearCalendar.Year)

	} else if it.MonthCalendar != nil {
		layout.AddMonthCalendar(it.X, it.Y, it.W, it.H, it.MonthCalendar.Year, it.MonthCalendar.Month, it.MonthCalendar.Events)

	} else if it.DayCalendar != nil {
		layout.AddDayCalendar(it.X, it.Y, it.W, it.H, it.DayCalendar.Days, it.DayCalendar.Events)
	} else {
		layout.AddLayout(it.X, it.Y, it.W, it.H)
	}

	it2 := layout.FindGrid(it.X, it.Y, it.W, it.H)
	if it2 != nil {
		it.addLayout(it2, appName, toolName, parent_UID, fnProgress, fnDone)
	}
}

func (layout *Layout) setLayoutFromUI(ui *UI) {
	layout.Enable = ui.Enable
	layout.EnableTouch = ui.EnableTouch
	layout.EnableBrush = ui.EnableBrush
	layout.Back_cd = ui.Back_cd
	layout.Back_margin = ui.Back_margin
	layout.Back_rounding = ui.Back_rounding
	layout.Border_cd = ui.Border_cd

	layout.scrollV.Show = !ui.ScrollV.Hide
	layout.scrollH.Show = !ui.ScrollH.Hide
	layout.scrollV.Narrow = ui.ScrollV.Narrow
	layout.scrollH.Narrow = ui.ScrollH.Narrow

	if layout.Tooltip == "" {
		layout.Tooltip = ui.Tooltip
	}
	layout.TooltipGroups = ui.TooltipGroups

	layout.App = ui.App

	layout.UID = ui.UID
}
