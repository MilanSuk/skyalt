package main

import "fmt"

// Shows UI to filter events(date range) and export them into .ics file.
type ExportEventsIntoICS struct {
	DateStart string //Select only runs after this data. Ignore = empty string. Date format: YYYY-MM-DD HH:MM
	DateEnd   string //Select only runs before this data. Ignore = empty string. Date format: YYYY-MM-DD HH:MM

	FilePath string // Path to .ics file.  Optional, default is "". [optional]
}

func (st *ExportEventsIntoICS) run(caller *ToolCaller, ui *UI) error {
	source_events, err := NewEvents("", caller)
	if err != nil {
		return err
	}

	ui.SetColumn(0, 2, 3)
	ui.SetColumn(1, 3, 16)

	ui.AddTextLabel(0, 0, 2, 1, "Export .ics file")

	ui.AddText(0, 1, 1, 1, "File(.ics)")
	fpck := ui.AddFilePickerButton(1, 1, 1, 1, &st.FilePath, true, false)

	events, err := source_events.Filter(st.DateStart, st.DateEnd)
	if err != nil {
		return err
	}

	ui.AddText(1, 2, 1, 1, fmt.Sprintf("%d events selected between %s -> %s", len(events), st.DateStart, st.DateEnd))

	bt := ui.AddButton(0, 4, 2, 1, "Export")
	bt.clicked = func() error {
		//checks
		if st.FilePath == "" {
			fpck.Error = "Empty field"
			return fmt.Errorf("invalid input(s)")
		}

		err := source_events.ExportICS(events, st.FilePath)
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}
