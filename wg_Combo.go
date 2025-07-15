package main

import "slices"

type Combo struct {
	Value *string

	Labels []string
	Values []string

	changed func()
}

func (layout *Layout) AddCombo(x, y, w, h int, value *string, labels []string, values []string) *Combo {
	props := &Combo{Value: value, Labels: labels, Values: values}
	layout._createDiv(x, y, w, h, "Combo", props.Build, nil, nil)
	return props
}

func (st *Combo) Build(layout *Layout) {

	layout.SetColumnFromSub(0, 1, 100, false)

	cdialog := layout.AddDialog("dialog")

	//button
	bt := layout.AddButton(0, 0, 1, 1, Combo_getValueLabel(*st.Value, st.Values, st.Labels))
	bt.Tooltip = *st.Value
	bt.IconPath = "resources/arrow_down.png"
	bt.Icon_margin = 0.1
	bt.Icon_align = 2
	bt.Align = 0
	bt.Background = 0
	bt.Border = true
	bt.clicked = func() {
		cdialog.OpenRelative(layout.UID)
	}

	{
		found_val := slices.Contains(st.Values, *st.Value)
		if !found_val {
			bt.Cd = layout.GetPalette().E //error border
		}
	}

	//dialog
	{
		cdialog.Layout.SetColumnFromSub(0, 1, 15, true)
		for i, name := range st.Labels {

			bt := cdialog.Layout.AddButton(0, i, 1, 1, name)
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

func Combo_getValueLabel(value string, Values, Labels []string) string {
	for i, it := range Values {
		if it == value {
			if i < len(Labels) {
				return Labels[i]
			}
			break
		}
	}
	return value //not found
}
