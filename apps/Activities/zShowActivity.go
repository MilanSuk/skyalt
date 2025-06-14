package main

import (
	"fmt"
)

// Edit activity attributes
type ShowActivity struct {
	ActivityID string //ID of the activity
}

func (st *ShowActivity) run(caller *ToolCaller, ui *UI) error {
	source_activities, err := NewActivities("")
	if err != nil {
		return err
	}

	activity, found := source_activities.Activities[st.ActivityID]
	if !found {
		return fmt.Errorf("activity '%s' not found", st.ActivityID)
	}

	ui.SetColumn(0, 2, 4)
	ui.SetColumn(1, 3, 16)

	ui.AddTextLabel(0, 0, 2, 1, fmt.Sprintf("Edit - %s", ConvertTextDate(int64(activity.Date))))

	y := 1
	ui.AddText(0, y, 1, 1, "Type")
	cb := ui.AddCombo(1, y, 1, 1, &activity.Type, source_activities.GetTypeLabels(), source_activities.GetTypeValues())
	cb.DialogWidth = 3
	cb.changed = func() error {
		return nil
	}
	y++

	ui.AddText(0, y, 1, 1, "Description")
	ui.AddEditboxString(1, y, 1, 1, &activity.Description)
	y++

	return nil
}
