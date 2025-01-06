package main

type ShowApp struct {
}

func (layout *Layout) AddShowApp(x, y, w, h int) *ShowApp {
	props := &ShowApp{}
	layout._createDiv(x, y, w, h, "ShowApp", props.Build, nil, nil)
	return props
}

func (st *ShowApp) Build(layout *Layout) {
	layout.SetColumn(0, 1, 100)
	layout.SetRow(0, 1, 100)

	layout.AddSettings(0, 0, 1, 1, OpenFile_Settings())
}
