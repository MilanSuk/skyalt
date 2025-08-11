package main

import (
	"fmt"
)

// ShowAddNewGroup is a tool to display a GUI for adding a new group.
// It allows the user to set the label and color components.
type ShowAddNewGroup struct {
	Label string // Label of the group [optional]
	Red   uint8  // Red component of color (0-255)
	Green uint8  // Green component of color (0-255)
	Blue  uint8  // Blue component of color (0-255)
}

func (st *ShowAddNewGroup) run(caller *ToolCaller, ui *UI) error {
	var label string = st.Label      // Use the provided label if available
	var redInt int = int(st.Red)     // Integer value for red component to bind to UI
	var greenInt int = int(st.Green) // Integer value for green component to bind to UI
	var blueInt int = int(st.Blue)   // Integer value for blue component to bind to UI

	ui.addTextH1("Add New Group")

	form := ui.addTable("AddNewGroupForm") // Table for group addition form

	// Line for label input
	ln := form.addLine("AddNewGroupLabelLine")
	ln.addText("Label", "Label of the group") // Description for label
	ln.addEditboxString(&label, "Label=group label")

	// Line for red input
	ln = form.addLine("AddNewGroupRedLine")
	ln.addText("Red", "Red component (0-255)")
	ln.addEditboxInt(&redInt, "Red=group color red component")

	// Line for green input
	ln = form.addLine("AddNewGroupGreenLine")
	ln.addText("Green", "Green component (0-255)")
	ln.addEditboxInt(&greenInt, "Green=group color green component")

	// Line for blue input
	ln = form.addLine("AddNewGroupBlueLine")
	ln.addText("Blue", "Blue component (0-255)")
	ln.addEditboxInt(&blueInt, "Blue=group color blue component")

	// Line for add button
	ln = form.addLine("AddNewGroupButtonLine")
	button := ln.addButton("Add Group", "AddGroupButton=click to add new group")
	button.clicked = func() error {
		groups, err := LoadGroups()
		if err != nil {
			return err
		}

		newID := fmt.Sprintf("group%d", len(groups.Items)) // Generate a simple new ID based on current count
		newGroup := Group{
			GroupID: newID,
			Label:   label,
			Color: Color{
				R: uint8(redInt),   // Convert back to uint8
				G: uint8(greenInt), // Convert back to uint8
				B: uint8(blueInt),  // Convert back to uint8
				A: 255,             // Assume full opacity as default
			},
		}
		groups.Items[newID] = newGroup // Add the new group to the storage
		// No explicit save needed; it's handled automatically
		return nil
	}

	return nil
}
