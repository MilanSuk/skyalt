package main

import "fmt"

type ShowGroups struct {
	// No arguments required for this tool.
}

func (st *ShowGroups) run(caller *ToolCaller, ui *UI) error {
	gs, err := LoadGroups()
	if err != nil {
		return err
	}

	ui.addTextH1("List of Groups")

	table := ui.addTable("Groups") // llmtip: Represents the list of groups

	for id, group := range gs.Items {
		ln := table.addLine(fmt.Sprintf("GroupID = %s", id))         // llmtip: Identifies the group entry
		ln.addText(group.Label, fmt.Sprintf("GroupID=%s Label", id)) // llmtip: Label for the specific group
	}

	return nil
}
