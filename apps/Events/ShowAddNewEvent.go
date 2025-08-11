package main

import (
	"strconv"
	"strings"
)

// Out_EventID is the ID of the newly created event.
type ShowAddNewEvent struct {
	Out_EventID string // The ID of the newly created event
}

func (st *ShowAddNewEvent) run(caller *ToolCaller, ui *UI) error {
	events, err := LoadEvents()
	if err != nil {
		return err
	}

	groups, err := LoadGroups()
	if err != nil {
		return err
	}

	newEvent := Event{}    // Temporary Event struct that will be populated via the UI form
	var groupID string     // Selected Group ID from the dropdown
	var filesString string // Comma-separated list of files
	var startInt int       // Local variable for Start time (int for UI compatibility)
	var durationInt int    // Local variable for Duration (int for UI compatibility)

	ui.addTextH1("Add New Event")

	form := ui.addTable("Form to add a new event")

	// Line for Title
	ln := form.addLine("Title field")
	ln.addText("Title:", "") // Description for the title field
	ln.addEditboxString(&newEvent.Title, "Title of the event")

	// Line for Description
	ln = form.addLine("Description field")
	ln.addText("Description:", "") // Description for the description field
	ed := ln.addEditboxString(&newEvent.Description, "Description of the event")
	ed.setMultilined() // Enable multi-line input for description

	// Line for Start Time
	ln = form.addLine("Start time field")
	ln.addText("Start (Unix timestamp):", "") // Description for the start time field
	ln.addEditboxInt(&startInt, "Start time in Unix timestamp")

	// Line for Duration
	ln = form.addLine("Duration field")
	ln.addText("Duration (seconds):", "") // Description for the duration field
	ln.addEditboxInt(&durationInt, "Duration in seconds")

	// Line for Group ID
	ln = form.addLine("Group ID field")
	ln.addText("Group ID:", "") // Description for the group ID field
	groupLabels := []string{}   // List of group IDs for the dropdown
	for id := range groups.Items {
		groupLabels = append(groupLabels, id)
	}
	ln.addDropDown(&groupID, groupLabels, groupLabels, "Group ID to associate with the event [options: "+strings.Join(groupLabels, ", ")+"]")

	// Line for Files
	ln = form.addLine("Files field")
	ln.addText("Files (comma-separated):", "") // Description for the files field
	ln.addEditboxString(&filesString, "Comma-separated list of files for the event")

	// Add button to submit the form and create the event
	btn := ui.addButton("Add Event", "Button to create the new event")
	btn.clicked = func() error {
		newID := "event" + strconv.Itoa(len(events.Items)+1) // Simple ID generation based on current length
		newEvent.EventID = newID
		newEvent.GroupID = groupID                                          // Set from the dropdown
		newEvent.Start = int64(startInt)                                    // Convert to int64 for storage
		newEvent.Duration = int64(durationInt)                              // Convert to int64 for storage
		newEvent.Files = strings.Split(strings.TrimSpace(filesString), ",") // Split the string into a slice

		events.Items[newID] = newEvent // Add the new event to the storage
		st.Out_EventID = newID         // Set the output argument
		return nil
	}

	return nil
}
