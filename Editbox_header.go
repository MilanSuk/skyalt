package main

import (
	"image/color"
	"sync"
)

type Editbox struct {
	layout *Layout
	lock   sync.Mutex

	Cd color.RGBA

	Value      *string
	ValueFloat *float64
	ValueInt   *int
	FloatPrec  int //for 'ValueFloat'

	Ghost   string
	Tooltip string

	Align_h int
	Align_v int

	Icon        string
	Icon_margin float64

	Formating    bool
	Multiline    bool
	Linewrapping bool

	changed func()
	enter   func()
}

func (layout *Layout) AddEditbox(x, y, w, h int, value *string) *Editbox {
	props := &Editbox{Value: value, Align_v: 1, Formating: true}
	props.layout = layout._createDiv(x, y, w, h, "Editbox", props.Build, props.Draw, props.Input)
	return props
}

func (layout *Layout) AddEditboxInt(x, y, w, h int, value *int) *Editbox {
	props := &Editbox{ValueInt: value, Align_v: 1, Formating: true}
	props.layout = layout._createDiv(x, y, w, h, "Editbox", props.Build, props.Draw, props.Input)
	return props
}

func (layout *Layout) AddEditboxFloat(x, y, w, h int, value *float64, prec int) *Editbox {
	props := &Editbox{ValueFloat: value, FloatPrec: prec, Align_v: 1, Formating: true}
	props.layout = layout._createDiv(x, y, w, h, "Editbox", props.Build, props.Draw, props.Input)
	return props
}
