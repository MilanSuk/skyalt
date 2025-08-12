package main

import (
	"image/color"
)

//Notes:
//When ui.addText(), ui.addButton() and other ui.add...() functions are called, every GUI components is added to new line.
//To add more components on same line, use ui.addTable(). All components in table are aligned by columns which makes tables very usefull for creating forms.
//Use setRowHeight() only when user prompt require that row has specific height.
//Some functions may have 'llmtip' argument which describes what UI component represents. If component shows value from storage which has ID, the ID should be part of llmtip - few examples: "Year of born for PersonID=123", "Image with GalleryID='path/to/image'".

//Code inside callbacks(UIButton.clicked, UIEditbox.changed, LLMCompletion.update, etc.), should not write into UIs structures, it should write only into storage or tool's arguments.

//If button triggers LLM completion, use addLLMCompletionButton() instead of addButton().

type ToolCaller struct {
}

type UI struct { //also known as Layout
	Tooltip string

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

func (ui *UI) addCenteredUI() *UI //return new UI which has width=20 and empty spaces on left and right side

func (ui *UI) setRowHeight(min, max float64) //Set height of current row/line. Must be call before first ui.Add...() is called for the row.

type UITable struct {
	layout *UI
}

func (ui *UI) addTable(llmtip string) *UITable
func (table *UITable) addLine(llmtip string) *UI
func (table *UITable) addDivider() //whole line is horizontal line. Use to separate header from data.

type UIText struct {
	layout *UI
	Label  string

	Align_h int //0=left, 1=center, 2=right
	Align_v int //0=top, 1=center, 2=bottom

	Cd color.RGBA

	Selection    bool
	Formating    bool
	Multiline    bool
	Linewrapping bool

	EnableDropFile bool

	EnableCodeFormating bool

	dropFile func(pathes []string) error
}

func (ed *UIText) setMultilined() //Enable multi-line & Line-wrapping

func (ui *UI) addTextH1(label string) *UIText //add main heading
func (ui *UI) addTextH2(label string) *UIText //add secondary heading
func (ui *UI) addText(label string, llmtip string) *UIText

type UIEditbox struct {
	layout     *UI
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

	AutoSave bool

	changed func() error //called after .Value or .ValueFloat or .ValueInt has been changed
	enter   func() error
}

func (ed *UIEditbox) setMultilined() //Enable multi-line & Line-wrapping

func (ui *UI) addEditboxString(value *string, llmtip string) *UIEditbox
func (ui *UI) addEditboxInt(value *int, llmtip string) *UIEditbox
func (ui *UI) addEditboxFloat(value *float64, precision int, llmtip string) *UIEditbox

type UIButton struct {
	layout *UI
	Label  string
	Align  int

	Shortcut byte

	Background  float64
	Border      bool
	IconBlob    []byte
	IconPath    string
	Icon_margin float64
	BrowserUrl  string
	Cd          color.RGBA

	ConfirmQuestion string

	Drag_group              string
	Drop_group              string
	Drag_source             string
	Drag_index              int
	Drop_h, Drop_v, Drop_in bool

	clicked func() error //called after button is triggered

	dropMove func(src_i, dst_i int, aim_i int, src_source, dst_source string) error
}

func (ui *UI) addButton(label string, llmtip string) *UIButton

type UIDropDownIcon struct {
	Path   string
	Blob   []byte
	Margin float64
}
type UIDropDown struct {
	layout *UI
	Value  *string
	Labels []string
	Values []string
	Icons  []UIDropDownIcon

	changed func() error //called after .Value has been changed
}

func (ui *UI) addDropDown(value *string, labels []string, values []string, llmtip string) *UIDropDown

type UIPromptMenuIcon struct {
	Path   string
	Blob   []byte
	Margin float64
}

type UIPromptMenu struct {
	layout  *UI
	Prompts []string
	Icons   []UIPromptMenuIcon
}

func (ui *UI) addPromptMenu(prompts []string, llmtip string) *UIPromptMenu

type UISwitch struct {
	layout *UI
	Label  string
	Value  *bool

	changed func() error //called after .Value has been changed
}

func (ui *UI) addSwitch(label string, value *bool, llmtip string) *UISwitch

type UICheckbox struct {
	layout *UI
	Label  string
	Value  *float64

	changed func() error //called after .Value has been changed
}

func (ui *UI) addCheckbox(label string, value *float64, llmtip string) *UICheckbox

type UISlider struct {
	layout *UI
	Value  *float64
	Min    float64
	Max    float64
	Step   float64

	changed func() error //called after .Value has been changed
}

func (ui *UI) addSlider(value *float64, min, max, step float64, llmtip string) *UISlider

type UIDivider struct {
	layout     *UI
	Horizontal bool
}

func (ui *UI) addDivider(horizontal bool) *UIDivider

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

func (ui *UI) addMap(lon, lat, zoom *float64, llmtip string) *UIMap
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

func (ui *UI) addYearCalendar(Year int, llmtip string) *UIYearCalendar

type UIMonthCalendar struct {
	layout *UI
	Year   int
	Month  int //1=January, 2=February, etc.

	Events []UICalendarEvent
}

func (ui *UI) addMonthCalendar(Year int, Month int, Events []UICalendarEvent, llmtip string) *UIMonthCalendar

type UIDayCalendar struct {
	layout *UI
	Days   []int64
	Events []UICalendarEvent
}

func (ui *UI) addDayCalendar(Days []int64, Events []UICalendarEvent, llmtip string) *UIDayCalendar

type UIFilePickerButton struct {
	layout      *UI
	Path        *string
	Preview     bool
	OnlyFolders bool

	changed func() error
}

func (ui *UI) addFilePickerButton(path *string, preview bool, onlyFolders bool, llmtip string) *UIFilePickerButton

type UIDatePickerButton struct {
	layout   *UI
	Date     *int64
	Page     *int64
	ShowTime bool
	changed  func() error //called after .Date has been changed
}

func (ui *UI) addDatePickerButton(date *int64, page *int64, showTime bool, llmtip string) *UIDatePickerButton

type UIColorPickerButton struct {
	layout  *UI
	Cd      *color.RGBA
	changed func() error //called after .Cd has been changed
}

func (ui *UI) addColorPickerButton(cd *color.RGBA, llmtip string) *UIColorPickerButton

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

func (ui *UI) addChartLines(Lines []UIChartLine, llmtip string) *UIChartLines

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

func (ui *UI) addChartColumns(columns []UIChartColumn, x_labels []string, llmtip string) *UIChartColumns

// accepts unix time and returns date(formated by user)
func SdkGetDate(unix_sec int64) string

// accepts unix time and returns date(formated by user) + hour:minute:second
func SdkGetDateTime(unix_sec int64) string

type LLMCompletion struct {
	UID string

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

// Show Button to trigger LLM completion. When it's running, it shows "Stop" button and returns answer(full answer so far). After it's finished, it calls done callback with complete answer.
func (ui *UI) addLLMCompletionButton(buttonLabel string, comp *LLMCompletion, done func(answer string), caller *ToolCaller) (running bool, answer string)

// UID: unique ID for completion
func NewLLMCompletion(UID string, systemMessage string, userMessage string) *LLMCompletion {
	return &LLMCompletion{Temperature: 0.2, Max_tokens: 16384, Top_p: 0.95, SystemMessage: systemMessage, UserMessage: userMessage}
}
func (comp *LLMCompletion) Run(caller *ToolCaller) error

// Use this to check If LLM is running. If it's running you can show answer(full answer so far) to user.
func LLMCompletion_find(UID string, caller *ToolCaller) (running bool, answer string)

func LLMCompletion_stop(UID string, caller *ToolCaller)
