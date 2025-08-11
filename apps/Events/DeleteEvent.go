package main

import (
	"fmt"
)

// DeleteEvent represents a tool to remove an event from the storage.
type DeleteEvent struct {
	EventID string // The ID of the event to delete.
}

func (st *DeleteEvent) run(caller *ToolCaller, ui *UI) error {
	events, err := LoadEvents()
	if err != nil {
		return err
	}

	if _, found := events.Items[st.EventID]; !found {
		return fmt.Errorf("EventID %s not found", st.EventID)
	}

	delete(events.Items, st.EventID)
	return nil
}
