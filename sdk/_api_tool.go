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
