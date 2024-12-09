package main

import (
	"sync"
)

type Env struct {
	layout *Layout
	lock   sync.Mutex

	DateFormat  string
	Volume      float64
	Dpi         int
	Dpi_default int

	Fullscreen bool
	Stats      bool

	Theme             string
	ThemePalette      LayoutPalette
	CustomPalette     LayoutPalette
	UseDarkTheme      bool
	UseDarkThemeStart int //hours from midnight
	UseDarkThemeEnd   int
}

func (layout *Layout) AddEnv(x, y, w, h int, props *Env) *Env {
	props.layout = layout._createDiv(x, y, w, h, "Env", props.Build, nil, nil)
	return props
}

var g_Env *Env

func NewFile_Env() *Env {
	if g_Env == nil {
		g_Env = &Env{Volume: 0.5, Theme: "light"}

		_read_file("Env-Env", g_Env)
	}
	return g_Env
}
