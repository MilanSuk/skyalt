package main

import (
	"image/color"
)

//Note:
// Tip for 'width' or 'height' argument for all below 'func (ui *UI) add...(width/height int, ...)'.
// Negative width/height value means auto-resize based on content(text, etc.). In most of cases, you will use negative value(-1).
// Positive width/height value is cell size. One cell can fit 5 letters, for example "Abcde".

type ToolCaller struct {
}

type UI struct {
	LLMTip string

	Enable        bool
	EnableTouch   bool
	Back_cd       color.RGBA
	Back_margin   float64
	Back_rounding bool
	Border_cd     color.RGBA
	ScrollV       UIScroll
	ScrollH       UIScroll

	changed func(newParams []byte) error
}

func (ui *UI) addRow(height float64) *UI

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

	EnableCodeFormating bool

	dropFile func(pathes []string) error
}

func (ui *UI) addText(width float64, label string) *UIText

type UIEditbox struct {
	layout     *UI
	Name       string
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

	changed func() error
	enter   func() error
}

func (ui *UI) addUI(width float64) *UI

// 'height: -1' => auto-resize based on row content
func (ui *UI) addRow(height float64) *UI

func (ui *UI) addEditboxString(width float64, value *string) *UIEditbox
func (ui *UI) addEditboxInt(width float64, value *int) *UIEditbox
func (ui *UI) addEditboxFloat(width float64, value *float64, precision int) *UIEditbox

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

func (ui *UI) addButton(width float64, label string) *UIButton

type UICombo struct {
	layout *UI
	Value  *string
	Labels []string
	Values []string

	changed func() error
}

func (ui *UI) addCombo(width float64, value *string, labels []string, values []string) *UICombo

type UISwitch struct {
	layout  *UI
	Label   string
	Tooltip string
	Value   *bool

	changed func() error
}

func (ui *UI) addSwitch(width float64, label string, value *bool) *UISwitch

type UICheckbox struct {
	layout  *UI
	Label   string
	Tooltip string
	Value   *float64

	changed func() error
}

func (ui *UI) addCheckbox(width float64, label string, value *float64) *UICheckbox

type UISlider struct {
	layout *UI
	Value  *float64
	Min    float64
	Max    float64
	Step   float64

	changed func() error
}

func (ui *UI) addSlider(width float64, value *float64, min, max, step float64) *UISlider

type UIDivider struct {
	layout     *UI
	Horizontal bool
}

func (ui *UI) addDivider(width float64, horizontal bool) *UIDivider

type UIMapLoc struct {
	Lon   float64
	Lat   float64
	Label string
}
type UIMapLocators struct {
	Locators []UIMapLoc
	clicked  func(i int, caller *ToolCaller)
}

type UIMapSegmentTrk struct {
	Lon  float64
	Lat  float64
	Ele  float64
	Time string
	Cd   color.RGBA
}
type UIMapSegment struct {
	Trkpts []UIMapSegmentTrk
	Label  string
}
type UIMapRoute struct {
	Segments []UIMapSegment
}
type UIMap struct {
	layout         *UI
	Lon, Lat, Zoom *float64
	Locators       []UIMapLocators
	Routes         []UIMapRoute
}

func (ui *UI) addMap(width float64, lon, lat, zoom *float64) *UIMap
func (mp *UIMap) addLocators(loc UIMapLocators)
func (mp *UIMap) addRoute(route UIMapRoute)

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

func (ui *UI) addYearCalendar(width float64, Year int) *UIYearCalendar

type UIMonthCalendar struct {
	layout *UI
	Year   int
	Month  int //1=January, 2=February, etc.

	Events []UICalendarEvent
}

func (ui *UI) addMonthCalendar(width float64, Year int, Month int, Events []UICalendarEvent) *UIMonthCalendar

type UIDayCalendar struct {
	layout *UI
	Days   []int64
	Events []UICalendarEvent
}

func (ui *UI) addDayCalendar(width float64, Days []int64, Events []UICalendarEvent) *UIDayCalendar

type UIFilePickerButton struct {
	layout      *UI
	Path        *string
	Preview     bool
	OnlyFolders bool

	changed func() error
}

func (ui *UI) addFilePickerButton(width float64, path *string, preview bool, onlyFolders bool) *UIFilePickerButton

type UIDatePickerButton struct {
	layout   *UI
	Date     *int64
	Page     *int64
	ShowTime bool
	changed  func() error
}

func (ui *UI) addDatePickerButton(width float64, date *int64, page *int64, showTime bool) *UIDatePickerButton

type UIColorPickerButton struct {
	layout  *UI
	Cd      *color.RGBA
	changed func() error
}

func (ui *UI) addColorPickerButton(width float64, cd *color.RGBA) *UIColorPickerButton

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

func (ui *UI) addChartLines(width float64, Lines []UIChartLine) *UIChartLines

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

func (ui *UI) addChartColumns(width float64, columns []UIChartColumn, x_labels []string) *UIChartColumns

// accepts unix time and returns date(formated by user)
func SdkGetDate(unix_sec int64) string

// accepts unix time and returns date(formated by user) + hour:minute:second
func SdkGetDateTime(unix_sec int64) string

type LLMCompletion struct {
	Temperature       float64
	Top_p             float64
	Max_tokens        int
	Frequency_penalty float64
	Presence_penalty  float64
	Reasoning_effort  string //"low", "medium", "high"

	SystemMessage string
	UserMessage   string
	UserFiles     []string

	Response_format string //"", "json_object"

	Out_answer    string
	Out_reasoning string
}

func NewLLMCompletion(systemMessage string, userMessage string) *LLMCompletion {
	return &LLMCompletion{Temperature: 0.2, Max_tokens: 16384, Top_p: 0.95, SystemMessage: systemMessage, UserMessage: userMessage}
}

func (comp *LLMCompletion) Run(caller *ToolCaller) error
