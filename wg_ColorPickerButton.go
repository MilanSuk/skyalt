package main

import (
	"fmt"
	"image/color"
)

type ColorPickerButton struct {
	Cd      *color.RGBA
	changed func()
}

func (layout *Layout) AddColorPickerButton(x, y, w, h int, cd *color.RGBA) *ColorPickerButton {
	props := &ColorPickerButton{Cd: cd}
	layout._createDiv(x, y, w, h, "ColorPickerButton", props.Build, nil, nil)
	return props
}

func (st *ColorPickerButton) Build(layout *Layout) {
	layout.SetColumn(0, 1, 100)
	layout.SetRow(0, 1, 100)

	bt := layout.AddButtonMenu(0, 0, 1, 1, "", "resources/color.png", 0.1)
	bt.Border = true
	bt.Background = 1
	bt.Cd = *st.Cd //set Button
	bt.Tooltip = fmt.Sprintf("RGBA(%d,%d,%d,%d)", int(st.Cd.R), int(st.Cd.G), int(st.Cd.B), int(st.Cd.A))

	cdialog := layout.AddDialog("color_picker_dialog")
	cdialog.Layout.SetColumn(0, 10, 17)
	cdialog.Layout.SetRow(0, 7, 7)
	cld := cdialog.Layout.AddColorPicker(0, 0, 1, 1, st.Cd)
	cld.changed = func() {
		if st.changed != nil {
			st.changed()
		}
	}
	bt.clicked = func() {
		cdialog.OpenRelative(layout.UID)
	}
}
