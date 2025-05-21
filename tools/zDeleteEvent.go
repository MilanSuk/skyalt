package main

import (
	"encoding/json"
)

// Delete event
type DeleteEvent struct {
	EventID json.Number //ID of the event to delete
}

func (st *DeleteEvent) run(caller *ToolCaller, ui *UI) error {
	source_events, err := NewEvents("", caller)
	if err != nil {
		return err
	}

	ID, err := st.EventID.Int64()
	if err != nil {
		return err
	}

	delete(source_events.Events, ID)

	return nil
}
