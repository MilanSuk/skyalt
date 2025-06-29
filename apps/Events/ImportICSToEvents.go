package main

import (
	"encoding/json"
	"time"
)

// Form to import .ics file into calendar.
type ImportICSToEvents struct {
	FilePath string // Path to .ics file. Optional, default is "". [optional]

	GroupID json.Number //Group ID. Use function GetListOfGroups() to get list of groups ids.
}

func (st *ImportICSToEvents) run(caller *ToolCaller, ui *UI) error {
	source_events, err := NewEvents("")
	if err != nil {
		return err
	}

	ui.SetColumn(0, 2, 4)
	ui.SetColumn(1, 3, 16)

	ui.AddTextLabel(0, 0, 2, 1, "Import .ics file")

	ui.AddText(0, 1, 1, 1, "File(.ics)")
	ui.AddFilePickerButton(1, 1, 1, 1, &st.FilePath, true, false)

	bt := ui.AddButton(0, 3, 2, 1, "Import")
	bt.clicked = func() error {

		groupID, _ := st.GroupID.Int64()

		events, err := source_events.ImportICS(st.FilePath, groupID)
		if err != nil {
			return err
		}

		for _, ev := range events {
			source_events.Events[time.Now().UnixNano()] = ev
		}

		return nil
	}

	return nil
}
