package main

import (
	"encoding/json"
)

// GetListOfGroups is a tool that retrieves and returns a JSON representation of all groups, including their ID and attributes.
type GetListOfGroups struct {
	Out_GroupsJSON string // The JSON string containing the list of groups, including GroupID, Label, and Color attributes
}

func (st *GetListOfGroups) run(caller *ToolCaller, ui *UI) error {
	gs, err := LoadGroups()
	if err != nil {
		return err
	}

	var groupsList []Group
	for _, group := range gs.Items {
		groupsList = append(groupsList, group)
	}

	jsonData, err := json.Marshal(groupsList)
	if err != nil {
		return err
	}

	st.Out_GroupsJSON = string(jsonData)

	return nil
}
