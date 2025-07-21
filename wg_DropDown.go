package main

import (
	"slices"
)

type DropDown struct {
	Tooltip string
	Value   *string

	Labels []string
	Values []string

	changed func()
}

func (layout *Layout) AddDropDown(x, y, w, h int, value *string, labels []string, values []string) *DropDown {
	props := &DropDown{Value: value, Labels: labels, Values: values}
	lay := layout._createDiv(x, y, w, h, "drop_down", props.Build, nil, nil)
	lay.fnGetLLMTip = props.getLLMTip
	return props
}

func (st *DropDown) getLLMTip(layout *Layout) string {
	return Layout_buildLLMTip("DropDown", *st.Value, true, st.Tooltip)
}

func (st *DropDown) Build(layout *Layout) {

	layout.SetColumnFromSub(0, 1, 100, false)

	cdialog := layout.AddDialog("dialog")

	//button
	bt := layout.AddButton(0, 0, 1, 1, DropDown_getValueLabel(*st.Value, st.Values, st.Labels))
	bt.Tooltip = st.Tooltip
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

func DropDown_getValueLabel(value string, Values, Labels []string) string {
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
