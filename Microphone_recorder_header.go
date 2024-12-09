package main

import (
	"sync"

	"github.com/go-audio/audio"
)

type Microphone_recorder struct {
	layout *Layout
	lock   sync.Mutex

	Label        string
	Tooltip      string
	Shortcut_key byte
	Background   float64

	start func()
	done  func(buff audio.IntBuffer)
}

func (layout *Layout) AddMicrophone_recorder(x, y, w, h int) *Microphone_recorder {
	props := &Microphone_recorder{}
	props.layout = layout._createDiv(x, y, w, h, "Microphone_recorder", props.Build, nil, nil)
	return props
}
