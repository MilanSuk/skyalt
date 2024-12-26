package main

type About struct {
}

func (layout *Layout) AddAbout(x, y, w, h int, props *About) *About {
	layout._createDiv(x, y, w, h, "About", props.Build, nil, nil)
	return props
}

var g_About *About

func NewFile_About() *About {
	if g_About == nil {
		g_About = &About{}
		_read_file("About-About", g_About)
	}
	return g_About
}

func (st *About) Build(layout *Layout) {
	layout.SetColumn(0, 1, 100)
	layout.SetColumn(1, 7, 15)
	layout.SetColumn(2, 1, 100)
	layout.SetRow(1, 2, 4)

	layout.AddImage(1, 1, 1, 1, "resources/logo.png")

	Version := layout.AddText(1, 3, 1, 1, "v0.1")
	Version.Align_h = 1

	Url := layout.AddButton(1, 4, 1, 1, NewButton("github.com/milansuk/skyalt/"))
	Url.Background = 0
	Url.BrowserUrl = "https://github.com/milansuk/skyalt/"

	License := layout.AddText(1, 5, 1, 1, "This program is distributed under the terms of Apache License, Version 2.0.")
	License.Align_h = 1

	Copyright := layout.AddText(1, 6, 1, 1, "This program comes with absolutely no warranty.")
	Copyright.Align_h = 1

	Warranty := layout.AddText(1, 7, 1, 1, "Copyright © 2023 - SkyAlt team")
	Warranty.Align_h = 1
}
