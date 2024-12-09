package main

import (
	"image/color"
	"sync"
)

type Button struct {
	layout *Layout
	lock   sync.Mutex

	Value string //label

	Tooltip string
	Align   int

	Background float64
	Border     bool

	Icon        string
	Icon_align  int
	Icon_margin float64

	BrowserUrl string

	Cd      color.RGBA
	Cd_fade bool

	clicked   func()
	clickedEx func(numClicks int, altClick bool)
}

func NewButton(label string) *Button {
	return &Button{Value: label, Align: 1, Background: 1}
}

func NewButtonDanger(label string, palette *LayoutPalette) *Button {
	bt := NewButton(label)
	bt.Cd = palette.E // background has error(red) color

	return bt
}
func NewButtonIcon(icon_path string, icon_margin float64, Tooltip string) *Button {
	return &Button{Icon: icon_path, Icon_align: 1, Icon_margin: icon_margin, Tooltip: Tooltip, Background: 1}
}
func NewButtonMenu(label string, icon_path string, icon_margin float64) *Button {
	return &Button{Value: label, Icon: icon_path, Icon_margin: icon_margin, Align: 0, Background: 0.25}
}

func (layout *Layout) AddButton(x, y, w, h int, props *Button) *Button {
	props.layout = layout._createDiv(x, y, w, h, "Button", nil, props.Draw, props.Input)
	return props
}
