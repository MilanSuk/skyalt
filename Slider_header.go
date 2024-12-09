package main

import (
	"sync"
)

type Slider struct {
	layout *Layout
	lock   sync.Mutex

	Value *float64
	Min   float64
	Max   float64
	Step  float64

	Legend bool

	DrawSteps bool

	changed func()
}

func (layout *Layout) AddSlider(x, y, w, h int, value *float64, min, max, step float64) *Slider {
	props := &Slider{Value: value, Min: min, Max: max, Step: step}
	props.layout = layout._createDiv(x, y, w, h, "Slider", nil, props.Draw, props.Input)
	return props
}
