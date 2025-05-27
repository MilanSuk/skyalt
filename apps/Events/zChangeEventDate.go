package main

import (
	"encoding/json"
	"fmt"
	"time"
)

// Change event's Start or Duration. For example to move event to different day/time.
type ChangeEventDate struct {
	ID json.Number //Event ID

	Start string //Date of event start. Format: YYYY-MM-DD HH:MM.
	End   string //Date of event end. Format: YYYY-MM-DD HH:MM.

}

func (st *ChangeEventDate) run(caller *ToolCaller, ui *UI) error {
	source_events, err := NewEvents("", caller)
	if err != nil {
		return err
	}

	ts, err := time.ParseInLocation("2006-01-02 15:04", st.Start, time.Local)
	if err != nil {
		return err
	}
	te, err := time.ParseInLocation("2006-01-02 15:04", st.End, time.Local)
	if err != nil {
		return err
	}

	startDate := ts.Unix()
	endDate := te.Unix()

	if startDate >= endDate {
		endDate = startDate + 30*60
	}

	ID, err := st.ID.Int64()
	if err != nil {
		return err
	}

	event, found := source_events.Events[ID]
	if !found {
		return fmt.Errorf("event '%d' not found", ID)
	}

	//update
	event.Start = startDate
	event.Duration = endDate - startDate
	source_events.Events[ID] = event

	return nil
}
