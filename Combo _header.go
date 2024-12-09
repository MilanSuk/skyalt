package main

type Combo struct {
	layout *Layout

	Value *string

	Labels []string
	Values []string

	DialogWidth float64

	changed func()
}

func (layout *Layout) AddCombo(x, y, w, h int, value *string, labels []string, values []string) *Combo {
	props := &Combo{DialogWidth: 10, Value: value, Labels: labels, Values: values}
	props.layout = layout._createDiv(x, y, w, h, "Combo", props.Build, nil, nil)
	return props
}
