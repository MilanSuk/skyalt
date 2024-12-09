package main

import (
	"sync"
)

type Counter struct {
	layout *Layout
	lock   sync.Mutex
	Count  int
}

func (layout *Layout) AddCounter(x, y, w, h int, props *Counter) *Counter {
	props.layout = layout._createDiv(x, y, w, h, "Counter", props.Build, nil, nil)
	return props
}

var g_Counter *Counter

func NewFile_Counter() *Counter {
	if g_Counter == nil {
		g_Counter = &Counter{Count: 7}
		_read_file("Counter-Counter", g_Counter)
	}
	return g_Counter
}
