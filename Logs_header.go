package main

import (
	"sync"
)

type Error struct {
	Time_unix   int64
	Layout_hash uint64
	Error       string
}

type Logs struct {
	layout *Layout
	lock   sync.Mutex

	Items []Error
}

func (layout *Layout) AddLogs(x, y, w, h int, props *Logs) *Logs {
	props.layout = layout._createDiv(x, y, w, h, "Logs", props.Build, nil, nil)
	return props
}

var g_Logs *Logs

func NewFile_Logs() *Logs {
	if g_Logs == nil {
		g_Logs = &Logs{}
		_read_file("Logs-Logs", g_Logs)
	}
	return g_Logs
}
