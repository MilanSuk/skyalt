package main

type Combo struct {
	Value *string

	Labels []string
	Values []string

	DialogWidth float64

	changed func()
}

func (layout *Layout) AddCombo(x, y, w, h int, value *string, labels []string, values []string) *Combo {
	props := &Combo{DialogWidth: 10, Value: value, Labels: labels, Values: values}
	layout._createDiv(x, y, w, h, "Combo", props.Build, nil, nil)
	return props
}

func (st *Combo) Build(layout *Layout) {

	layout.SetColumn(0, 1, 100)

	cdialog := layout.AddDialog("dialog")

	//button
	bt := layout.AddButton(0, 0, 1, 1, NewButton(st.getValueLabel())) //from scratch every frame => easy to debug. Hold
	bt.Tooltip = *st.Value

	bt.Icon = "resources/arrow.png"
	bt.Icon_margin = 0.1
	bt.Icon_align = 2
	bt.Align = 0
	bt.Background = 0
	bt.Border = true

	bt.clicked = func() {
		cdialog.OpenRelative(layout)
	}

	{
		found_val := false
		for _, val := range st.Values {
			if val == *st.Value {
				found_val = true
				break
			}
		}
		if !found_val {
			bt.Cd = Paint_GetPalette().E //error border
		}
	}

	//dialog
	{
		cdialog.Layout.SetColumn(0, 1, st.DialogWidth)
		for i, name := range st.Labels {

			bt := cdialog.Layout.AddButton(0, i, 1, 1, NewButton(name))
			bt.Tooltip = st.Values[i]
			if st.Values[i] == *st.Value {
				bt.Background = 1 //selected
			} else {
				bt.Background = 0.25
			}
			bt.Align = 0

			bt.clicked = func() {
				*st.Value = bt.Tooltip
				if st.changed != nil {
					st.changed()
				}
				cdialog.Close()
			}
		}
	}
}

func (st *Combo) getValueLabel() string {
	for i, it := range st.Values {
		if it == *st.Value {
			if i < len(st.Labels) {
				return st.Labels[i]
			}
			break
		}
	}
	return *st.Value //not found
}
