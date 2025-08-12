package main

import (
	"fmt"
	"slices"
)

type DropDownIcon struct {
	Path   string
	Blob   []byte
	Margin float64
}

type DropDown struct {
	Tooltip string
	Value   *string

	Labels []string
	Values []string
	Icons  []DropDownIcon

	changed func()
}

func (layout *Layout) AddDropDown(x, y, w, h int, value *string, labels []string, values []string) *DropDown {
	props := &DropDown{Value: value, Labels: labels, Values: values}
	lay := layout._createDiv(x, y, w, h, "drop_down", props.Build, nil, nil)
	lay.fnGetLLMTip = props.getLLMTip
	return props
}

func (st *DropDown) getLLMTip(layout *Layout) string {
	if st.Value != nil {
		label := *st.Value
		i := st.getValueLabelPos()
		if i >= 0 {
			label = st.Labels[i]
		}

		return Layout_buildLLMTip("DropDown", "with", fmt.Sprintf("value=\"%s\" and label=\"%s\"", *st.Value, label), st.Tooltip)
	}
	return "Error: Value == nil"
}

func (st *DropDown) Build(layout *Layout) {
	layout.scrollH.Narrow = true
	layout.scrollV.Show = false

	layout.SetColumnFromSub(0, 1, Layout_MAX_SIZE, false)

	if st.Value == nil {
		layout.AddText(0, 0, 1, 1, "Error: Value == nil")
		return
	}

	cdialog := layout.AddDialog("dialog")

	//button
	bt := layout.AddButton(0, 0, 1, 1, "")
	bt.Tooltip = st.Tooltip
	bt.Icon2Path = "resources/arrow_down.png"
	bt.Icon2_margin = 0.1
	bt.Align = 0
	bt.Background = 0
	bt.Border = true
	layout.FindGrid(0, 0, 1, 1).fnGetLLMTip = nil //DropDown already has LLMTip

	//set Label and Icon
	{
		i := st.getValueLabelPos()
		if i >= 0 {
			bt.Value = st.Labels[i]
			if i < len(st.Icons) {
				bt.IconBlob = st.Icons[i].Blob
				bt.IconPath = st.Icons[i].Path
				bt.Icon_margin = st.Icons[i].Margin
			}
		}
	}

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

			if i < len(st.Icons) {
				bt.IconBlob = st.Icons[i].Blob
				bt.IconPath = st.Icons[i].Path
				bt.Icon_margin = st.Icons[i].Margin
			}

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

func (st *DropDown) getValueLabelPos() int {
	for i, it := range st.Values {
		if it == *st.Value {
			if i < len(st.Labels) {
				return i
			}
		}
	}
	return -1
}
