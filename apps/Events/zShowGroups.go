package main

import (
	"fmt"
	"time"
)

// Show list of groups. User can rename group or change color. User can also add new group or delete group.
type ShowGroups struct {
}

func (st *ShowGroups) run(caller *ToolCaller, ui *UI) error {
	source_events, err := NewEvents("")
	if err != nil {
		return err
	}

	ui.SetColumn(0, 1, 15)

	ui.AddTextLabel(0, 0, 1, 1, "Calendar Groups")

	sorted := source_events.getSortedGroupIDs()
	y := 1
	for _, id := range sorted {
		group := source_events.Groups[id]
		GroupDiv := ui.AddLayout(0, y, 1, 1)
		GroupDiv.SetColumn(0, 2, 100)
		GroupDiv.SetColumn(1, 3, 4)
		GroupDiv.SetColumn(3, 3, 3)

		GroupDiv.LLMTip = fmt.Sprintf("Calendar GroupID: %d, Label: %s, Color(RGB): %d,%d,%d", id, group.Label, group.Color.R, group.Color.G, group.Color.B)

		ed := GroupDiv.AddEditboxString(0, 0, 1, 1, &group.Label)
		ed.changed = func() error {
			return nil
		}
		cd := GroupDiv.AddColorPickerButton(1, 0, 1, 1, &group.Color)
		cd.changed = func() error {
			return nil
		}

		bt := GroupDiv.AddButton(3, 0, 1, 1, "Remove")
		bt.ConfirmQuestion = "Are you sure?"
		bt.clicked = func() error {
			delete(source_events.Groups, id)
			return nil
		}

		y++
	}

	y++ //space

	bt := ui.AddButton(0, y, 1, 1, "Add new Group")
	bt.clicked = func() error {
		//new
		source_events.Groups[time.Now().UnixNano()] = &EventsGroup{Label: "New Group", Color: UI_GetPalette().P}
		return nil
	}

	return nil
}
