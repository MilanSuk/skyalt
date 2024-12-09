package main

type Switch struct {
	layout *Layout

	Label   string
	Tooltip string
	Value   *bool

	changed func()
}

func (layout *Layout) AddSwitch(x, y, w, h int, label string, value *bool) *Switch {
	props := &Switch{Label: label, Value: value}
	props.layout = layout._createDiv(x, y, w, h, "Switch", nil, props.Draw, props.Input)
	return props
}
