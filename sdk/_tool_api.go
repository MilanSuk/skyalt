package main

import (
	"image/color"
)

//Notes:
//When ui.addText(), ui.addButton() and other ui.add...() functions are called, every GUI components is added to new line.
//To add more components on same line, use ui.addTable(). All components in table are aligned by columns which makes tables very usefull for creating forms.
//Use setRowHeight() only when user prompt require that row has specific height.

//Some functions may have 'llmtip' argument which describes what UI component represents. If component or table line show value from storage which has ID, the format should be <storage path>=<ID>. Few examples: "Year of born for PersonID=123", "Image with GalleryID='path/to/image'".

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

func (ui *UI) addLeftSideUI(resizable bool) (*UI, *UI)                            //add and return Left side panel and Content UIs. Left side can be set resizable.
func (ui *UI) addRightSideUI(resizable bool) (*UI, *UI)                           //add and return Content and Right side panel UIs. Right side can be set resizable.
func (ui *UI) addBothSideUI(resizable_left, resizable_right bool) (*UI, *UI, *UI) //add and return Left side panel, Content and Right side panel UIs. Left and Right side can be set resizable.

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
	layout           *UI
	Value            *string
	ValueFloat       *float64
	ValueInt         *int
	Precision        int
	Ghost            string
	Password         bool
	ShowLineNumbers  bool
	ActivateOnCreate bool

	Align_h int //0=left, 1=center, 2=right
	Align_v int //0=top, 1=center, 2=bottom

	Formating    bool
	Multiline    bool
	Linewrapping bool

	enter func() error //called after editbox is finished by pressing enter key
}

func (ed *UIEditbox) setMultilined()              //Enable multi-line & Line-wrapping
func (ed *UIEditbox) Activate(caller *ToolCaller) //Activate editbox.

func (ui *UI) addEditboxString(value string, changed func(newValue string, self *UIEditbox), tooltip string) *UIEditbox
func (ui *UI) addEditboxInt(value *int, changed func(newValue int), tooltip string, self *UIEditbox) *UIEditbox
func (ui *UI) addEditboxFloat(value *float64, changed func(newValue float64, self *UIEditbox), precision int, tooltip string)

type UIButton struct {
	layout *UI
	Label  string
	Align  int //Default is 1

	Shortcut rune

	Background  float64 //0=transparent, 0.5=light back, 1.0=full color. Default is 1.0.
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
}

// 'labels' should start with Upper letter and have spaces(if multiple words).
func (ui *UI) addDropDown(value string, changed func(newValue string, self *UIDropDow), labels []string, values []string, tooltip string) *UIDropDow

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
}

func (ui *UI) addSwitch(label string, value bool, changed func(newValue bool, self *UISwitch), tooltip string) *UISwitch

type UICheckbox struct {
	layout *UI
	Label  string
	Value  *float64
}

func (ui *UI) addCheckbox(label string, value float64, changed func(newValue float64, self *UICheckbox), tooltip string) *UICheckbox

type UISlider struct {
	layout *UI
	Value  *float64
	Min    float64
	Max    float64
	Step   float64
}

func (ui *UI) addSlider(value float64, changed func(newValue float64, self *UISlider), min, max, step float64, tooltip string) *UISlider

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

	Color color.RGBA //optional, default is {0, 0, 0, 0}
}

type UIYearCalendar struct {
	layout *UI
	Year   int
}

func (ui *UI) addYearCalendar(Year int) *UIYearCalendar

type UIMonthCalendar struct {
	layout *UI
	Year   int
	Month  int //1=January, 2=February, etc.

	Events []UICalendarEvent
}

func (ui *UI) addMonthCalendar(Year int, Month int, Events []UICalendarEvent) *UIMonthCalendar

type UIDayCalendar struct {
	layout *UI
	Days   []int64
	Events []UICalendarEvent
}

func (ui *UI) addDayCalendar(Days []int64, Events []UICalendarEvent) *UIDayCalendar

type UIFilePickerButton struct {
	layout      *UI
	Path        *string
	Preview     bool
	OnlyFolders bool
}

func (ui *UI) addFilePickerButton(path string, changed func(newPath string, self *UIFilePickerButton), preview bool, onlyFolders bool, tooltip string) *UIFilePickerButton

type UIDatePickerButton struct {
	layout   *UI
	Date     *int64
	Page     *int64
	ShowTime bool
}

func (ui *UI) addDatePickerButton(date int64, changed func(newDate int64, self *UIDatePickerButton), page *int64, showTime bool, tooltip string) *UIDatePickerButton

type UIColorPickerButton struct {
	layout *UI
	Cd     *color.RGBA
}

func (ui *UI) addColorPickerButton(cd color.RGBA, changed func(newCd color.RGBA, self *UIColorPickerButton), tooltip string) *UIColorPickerButton

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

func (ui *UI) addMediaFilePath(path string, tooltip string) *UIMedia

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

// Show Button to trigger LLM completion. When it's already running, it shows "Stop" button and returns work-in-progress answer so you can show it on screen. After it's finished, it calls done() callback with complete answer.
func (ui *UI) addLLMCompletionButton(buttonLabel string, comp *LLMCompletion, done func(answer string), caller *ToolCaller) (running bool, work_in_progress_answer string)

// UID: unique ID for completion
func NewLLMCompletion(UID string, systemMessage string, userMessage string) *LLMCompletion {
	return &LLMCompletion{UID: UID, Temperature: 0.2, Max_tokens: 16384, Top_p: 0.95, SystemMessage: systemMessage, UserMessage: userMessage}
}

func (comp *LLMCompletion) Run(caller *ToolCaller) error

// Use this to check If LLM is running. If it's running you can show answer(full answer so far) to user.
func (comp *LLMCompletion) Find(caller *ToolCaller) (running bool, answer string)

func (comp *LLMCompletion) Stop(caller *ToolCaller) error

// Close tool
func (caller *ToolCaller) CloseTool()
