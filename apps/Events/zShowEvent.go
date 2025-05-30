package main

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Show form to edit event. User will fill the form and send it.
type ShowEvent struct {
	ID json.Number //Event ID
}

func (st *ShowEvent) run(caller *ToolCaller, ui *UI) error {
	source_events, err := NewEvents("")
	if err != nil {
		return err
	}

	ID, err := st.ID.Int64()
	if err != nil {
		return err
	}

	event, found := source_events.Events[ID]
	if !found {
		return fmt.Errorf("activity '%d' not found", ID)
	}

	ui.SetColumn(0, 2, 4)
	ui.SetColumn(1, 3, 16)

	ui.AddTextLabel(0, 0, 2, 1, "Edit event")

	ui.AddText(0, 1, 1, 1, "Title")
	tlt := ui.AddEditboxString(1, 1, 1, 1, &event.Title)
	tlt.changed = func() error {
		if event.Title == "" {
			tlt.Error = "Empty field"
		}
		return nil
	}

	ui.AddText(0, 2, 1, 1, "Description")
	desc := ui.AddEditboxString(1, 2, 1, 1, &event.Description)
	desc.changed = func() error {
		return nil
	}

	//start
	ui.AddText(0, 4, 1, 1, "Start")
	start := ui.AddDatePickerButton(1, 4, 1, 1, &event.Start, nil, true)
	start.changed = func() error {
		return nil
	}

	ui.AddText(0, 5, 1, 1, "Duration(h:m)")
	dur := ui.AddLayout(1, 5, 1, 1)
	dur.SetColumn(1, 0.5, 0.5)

	hours := int(event.Duration / 3600)
	mins := int((event.Duration % 3600) / 60)
	h := dur.AddEditboxInt(0, 0, 1, 1, &hours)
	h.Tooltip = "Hours"
	dur.AddText(1, 0, 1, 1, ":").Align_h = 1
	m := dur.AddEditboxInt(2, 0, 1, 1, &mins)
	m.Tooltip = "Minutes"
	h.changed = func() error {
		event.Duration = int64(hours*3600 + mins*60)
		return nil
	}
	m.changed = h.changed

	ui.AddText(0, 7, 1, 1, "Color")
	groupID := strconv.FormatInt(event.GroupID, 10)
	cb := ui.AddCombo(1, 7, 1, 1, &groupID, source_events.getGroupsLabels(), source_events.getGroupsValues())
	cb.changed = func() error {
		event.GroupID, _ = strconv.ParseInt(groupID, 10, 64)
		return nil
	}

	//files
	ui.AddText(0, 9, 1, 1, "Attachment(s)")
	//....
	//import new files
	//var files []string
	//for _, f := range st.Files {
	//copy files
	//}

	return nil
}
