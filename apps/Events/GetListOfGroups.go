package main

// Returns list of groups(GroupID, Label, Color) as JSON.
type GetListOfGroups struct {
	Out_groups map[int64]*EventsGroup
}

func (st *GetListOfGroups) run(caller *ToolCaller, ui *UI) error {
	source_events, err := NewEvents("")
	if err != nil {
		return err
	}

	st.Out_groups = source_events.Groups
	return nil
}
