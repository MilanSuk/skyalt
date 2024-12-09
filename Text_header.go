package main

import (
	"image/color"
	"sync"
)

type Text struct {
	layout *Layout
	lock   sync.Mutex

	Cd color.RGBA

	Value   string
	Tooltip string

	Align_h int
	Align_v int

	Icon        string
	Icon_margin float64

	Selection    bool
	Formating    bool
	Multiline    bool
	Linewrapping bool
}

func (layout *Layout) AddText(x, y, w, h int, value string) *Text {
	props := &Text{Value: value, Align_v: 1, Selection: true, Formating: true}
	props.layout = layout._createDiv(x, y, w, h, "Text", props.Build, props.Draw, props.Input)
	return props
}
