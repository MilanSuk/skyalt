package main

import (
	"fmt"
	"image/color"
)

// ShowEditGroup is a tool to display and edit the attributes of a specific group.
// It loads the group based on the provided GroupID and presents a form for editing
// the Label and Color attributes. Changes are saved automatically after the tool runs.
type ShowEditGroup struct {
	// GroupID is the unique identifier of the group to edit.
	GroupID string
}

func (st *ShowEditGroup) run(caller *ToolCaller, ui *UI) error {
	groups, err := LoadGroups()
	if err != nil {
		return fmt.Errorf("failed to load groups: %w", err)
	}

	group, found := groups.Items[st.GroupID]
	if !found {
		return fmt.Errorf("GroupID %s not found", st.GroupID)
	}

	ui.addTextH1("Edit Group")

	form := ui.addTable(fmt.Sprintf("Group information for GroupID=%s", st.GroupID))

	// Line for Label
	ln := form.addLine(fmt.Sprintf("Label for GroupID=%s", st.GroupID))
	ln.addText("Label", "")                    // Description for the label field
	ln.addEditboxString(&group.Label, "Label") // Editable field for the group's label

	// Line for Color
	var colorPicker color.RGBA = color.RGBA{group.Color.R, group.Color.G, group.Color.B, group.Color.A}
	ln = form.addLine(fmt.Sprintf("Color for GroupID=%s", st.GroupID))
	ln.addText("Color", "")                                  // Description for the color field
	picker := ln.addColorPickerButton(&colorPicker, "Color") // Color picker for the group's color
	picker.changed = func() error {
		// Update the group's color when the picker changes
		group.Color.R = colorPicker.R
		group.Color.G = colorPicker.G
		group.Color.B = colorPicker.B
		group.Color.A = colorPicker.A
		return nil
	}

	return nil
}
