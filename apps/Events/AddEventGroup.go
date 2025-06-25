package main

import (
	"fmt"
	"image/color"
	"time"
)

// Show GUI to add new group. User can set label and color.
type AddEventGroup struct {
	Label string //Label is optional. [optional]

	Cd_red   int //color's red channel in range <0-255>
	Cd_green int //color's green channel in range <0-255>
	Cd_blue  int //color's blue channel in range <0-255>
}

func (st *AddEventGroup) run(caller *ToolCaller, ui *UI) error {
	source_events, err := NewEvents("")
	if err != nil {
		return err
	}

	ui.SetColumn(0, 2, 4)
	ui.SetColumn(1, 3, 16)

	ui.AddTextLabel(0, 0, 2, 1, "Add new calendar Group")

	ui.AddText(0, 1, 1, 1, "Label")
	lbl := ui.AddEditboxString(1, 0, 1, 1, &st.Label)

	ui.AddText(0, 2, 1, 1, "Color")
	cd := color.RGBA{uint8(st.Cd_red), uint8(st.Cd_green), uint8(st.Cd_blue), 255}
	cdd := ui.AddColorPickerButton(1, 1, 1, 1, &cd)
	cdd.changed = func() error {
		st.Cd_red = int(cd.R)
		st.Cd_green = int(cd.R)
		st.Cd_blue = int(cd.R)
		return nil
	}

	bt := ui.AddButton(0, 4, 2, 1, "Add new Group")
	bt.clicked = func() error {
		//checks
		if st.Label == "" {
			lbl.Error = "Empty field"
		}

		if lbl.Error != "" {
			return fmt.Errorf("invalid input(s)")
		}

		//update
		source_events.Groups[time.Now().UnixNano()] = &EventsGroup{Label: st.Label, Color: color.RGBA{uint8(st.Cd_red), uint8(st.Cd_green), uint8(st.Cd_blue), 255}}

		return nil
	}

	return nil
}
