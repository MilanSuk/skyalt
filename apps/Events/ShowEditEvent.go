package main

import (
	"fmt"
	"strings"
)

// ShowEditEvent displays a form to edit an existing event based on the provided EventID.
// The form allows editing of the event's attributes: Title, Description, Files (as a comma-separated list),
// Start (as a timestamp), Duration (in seconds), and GroupID (selected from available groups).
type ShowEditEvent struct {
	EventID string // The ID of the event to edit
}

func (st *ShowEditEvent) run(caller *ToolCaller, ui *UI) error {
	events, err := LoadEvents()
	if err != nil {
		return fmt.Errorf("failed to load events: %w", err)
	}

	event, found := events.Items[st.EventID]
	if !found {
		return fmt.Errorf("EventID %s not found", st.EventID)
	}

	groups, err := LoadGroups()
	if err != nil {
		return fmt.Errorf("failed to load groups: %w", err)
	}

	// Prepare data for GroupID dropdown
	var groupLabels []string
	var groupValues []string
	for _, group := range groups.Items {
		groupLabels = append(groupLabels, group.Label)
		groupValues = append(groupValues, group.GroupID)
	}

	ui.addTextH1("Edit Event")

	form := ui.addTable(fmt.Sprintf("Event information for EventID = %s", st.EventID))

	// Line for Title
	ln := form.addLine("")
	ln.addText("Title", "")                    // Description of the field
	ln.addEditboxString(&event.Title, "Title") // Editable string value for the event title

	// Line for Description
	ln = form.addLine("")
	ln.addText("Description", "")                                             // Description of the field
	editDescription := ln.addEditboxString(&event.Description, "Description") // Editable string value for the event description
	editDescription.setMultilined()                                           // Enables multi-line editing for longer descriptions

	// Line for Files
	ln = form.addLine("")
	ln.addText("Files", "")                                // Description of the field; comma-separated list of file paths
	editFiles := ln.addEditboxString(new(string), "Files") // Temporary string for editing; will be parsed back to []string
	*editFiles.Value = strings.Join(event.Files, ",")      // Initialize with current files as comma-separated string
	editFiles.setMultilined()                              // Enables multi-line for easier editing of lists

	// Line for Start
	ln = form.addLine("")
	ln.addText("Start", "")                                         // Description of the field; Unix timestamp for the event start time
	ln.addDatePickerButton(&event.Start, new(int64), true, "Start") // Editable date with time; page is optional and set to a new int64

	// Line for Duration
	ln = form.addLine("")
	ln.addText("Duration", "")                 // Description of the field; Duration in seconds
	var durationInt int = int(event.Duration)  // Temporary int value for editing; converted from int64
	ln.addEditboxInt(&durationInt, "Duration") // Editable integer value for the event duration in seconds

	// Line for GroupID
	ln = form.addLine("")
	ln.addText("GroupID", "")                                           // Description of the field; ID of the group the event belongs to [options: values from loaded groups]
	ln.addDropDown(&event.GroupID, groupLabels, groupValues, "GroupID") // Dropdown for selecting from available groups

	form.addDivider() // Add a divider to separate the form from any potential buttons or notes

	return nil
}
