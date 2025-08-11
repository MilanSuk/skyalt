package main

import "fmt"

type ChangeEventGroup struct {
	EventID string // The ID of the event to change.
	GroupID string // The new GroupID to assign to the event.
}

func (st *ChangeEventGroup) run(caller *ToolCaller, ui *UI) error {
	events, err := LoadEvents()
	if err != nil {
		return err
	}

	if event, found := events.Items[st.EventID]; found {
		event.GroupID = st.GroupID
		events.Items[st.EventID] = event
		return nil
	}

	return fmt.Errorf("EventID %s not found", st.EventID)
}
