package main

import "sync"

type Microphone struct {
	layout *Layout
	lock   sync.Mutex

	Enable bool

	SampleRate int
	Channels   int
}

func (layout *Layout) AddMicrophone(x, y, w, h int) *Microphone {
	props := &Microphone{}
	props.layout = layout._createDiv(x, y, w, h, "Microphone", props.Build, nil, nil)
	return props
}

var g_Microphone *Microphone

func NewFile_Microphone() *Microphone {
	if g_Microphone == nil {
		g_Microphone = &Microphone{Enable: true, SampleRate: 44100, Channels: 1}
		_read_file("Microphone-Microphone", g_Microphone)
	}
	return g_Microphone
}
