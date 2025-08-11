package main

import (
	"fmt"
	"time"
)

type ChangeEventDate struct {
	EventID string // The ID of the event to change
	Start   string // New start date in YYYY-MM-DD HH:MM format
	End     string // New end date in YYYY-MM-DD HH:MM format
}

// Changes the start time and duration of an event based on the provided inputs.
func (st *ChangeEventDate) run(caller *ToolCaller, ui *UI) error {
	ui.addTextH1("Change Event Date")

	form := ui.addTable("Form for changing event date")

	ln := form.addLine("Event ID")
	ln.addText("Event ID", "")                  // Label for Event ID input
	ln.addEditboxString(&st.EventID, "EventID") // Input for the event ID

	ln = form.addLine("New Start Date")
	ln.addText("New Start Date", "")            // Label for new start date input
	ln.addEditboxString(&st.Start, "New Start") // Input for the new start date

	ln = form.addLine("New End Date")
	ln.addText("New End Date", "")          // Label for new end date input
	ln.addEditboxString(&st.End, "New End") // Input for the new end date

	btn := ui.addButton("Apply Changes", "Button to apply the date changes")
	btn.clicked = func() error {
		events, err := LoadEvents()
		if err != nil {
			return fmt.Errorf("Failed to load events: %v", err)
		}

		event, found := events.Items[st.EventID]
		if !found {
			return fmt.Errorf("EventID %s not found", st.EventID)
		}

		startTime, err := time.Parse("2006-01-02 15:04", st.Start)
		if err != nil {
			return fmt.Errorf("Invalid Start format: %v", err)
		}

		endTime, err := time.Parse("2006-01-02 15:04", st.End)
		if err != nil {
			return fmt.Errorf("Invalid End format: %v", err)
		}

		event.Start = startTime.Unix() // Set new start time in Unix seconds

		durationSeconds := endTime.Unix() - startTime.Unix()
		if durationSeconds < 0 {
			return fmt.Errorf("End time must be after Start time")
		}
		event.Duration = durationSeconds // Set new duration in seconds

		events.Items[st.EventID] = event // Update the event in storage

		return nil
	}

	return nil
}
