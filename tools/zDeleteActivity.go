package main

import (
	"fmt"
	"os"
)

// Deletes activity from the database.
type DeleteActivity struct {
	ActivityID string //ID of the activity to delete
}

func (st *DeleteActivity) run(caller *ToolCaller, ui *UI) error {
	source_activities, err := NewActivities("", caller)
	if err != nil {
		return err
	}

	// Check if activity exists
	_, found := source_activities.Activities[st.ActivityID]
	if !found {
		return fmt.Errorf("activity '%s' not found", st.ActivityID)
	}

	// Delete file
	os.Remove(source_activities.GetFilePath(st.ActivityID))

	// Remove items
	delete(source_activities.Activities, st.ActivityID)

	return nil
}
