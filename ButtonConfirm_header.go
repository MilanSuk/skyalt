package main

type ButtonConfirm struct {
	layout *Layout

	Value string //label

	Tooltip string
	Align   int

	Draw_back   float64
	Draw_border bool

	Icon        string
	Icon_align  int
	Icon_margin float64

	Question string

	Confirmed func() `json:"-"`
}

func (layout *Layout) AddButtonConfirm(x, y, w, h int, label string, question string) *ButtonConfirm {
	props := &ButtonConfirm{Value: label, Align: 1, Draw_back: 1, Question: question}
	props.layout = layout._createDiv(x, y, w, h, "ButtonConfirm", props.Build, nil, nil)
	return props
}

func (layout *Layout) AddButtonConfirmMenu(x, y, w, h int, label string, icon_path string, icon_margin float64, question string) *ButtonConfirm {
	props := &ButtonConfirm{Value: label, Icon: icon_path, Icon_margin: icon_margin, Align: 0, Draw_back: 0.25, Question: question}
	props.layout = layout._createDiv(x, y, w, h, "ButtonConfirm", props.Build, nil, nil)
	return props
}
