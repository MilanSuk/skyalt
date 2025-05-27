package main

import (
	"encoding/json"
	"fmt"
)

// Change event's GroupID attribute
type ChangeEventGroup struct {
	EventID json.Number //Event ID
	GroupID json.Number //Group ID. Use function GetListOfGroups() to get list of groups ids.
}

func (st *ChangeEventGroup) run(caller *ToolCaller, ui *UI) error {
	source_events, err := NewEvents("", caller)
	if err != nil {
		return err
	}

	ID, err := st.EventID.Int64()
	if err != nil {
		return err
	}

	_, found := source_events.Events[ID]
	if !found {
		return fmt.Errorf("event '%d' not found", ID)
	}

	//update
	source_events.Events[ID].GroupID, _ = st.GroupID.Int64()

	return nil
}
