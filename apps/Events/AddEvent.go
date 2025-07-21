package main

import (
	"strconv"
	"time"
)

// Show form to add new event into the calendar. User will fill the form and send it.
type AddEvent struct {
	//Title is optional. [optional]
	Title string

	//Description is optional. [optional]
	Description string

	//Pathes to attachment files. Files parameter is optional. [optional]
	Files []string

	//Date of event start. Format: YYYY-MM-DD HH:MM. [optional]
	Start string

	//Date of event end. Format: YYYY-MM-DD HH:MM. [optional]
	End string

	//GroupID
	GroupID int64
}

func (st *AddEvent) run(caller *ToolCaller, ui *UI) error {
	source_events, err := NewEvents("")
	if err != nil {
		return err
	}

	var startDate, endDate int64
	tm, err := time.ParseInLocation("2006-01-02 15:04", st.Start, time.Local)
	if err == nil {
		startDate = tm.Unix()
	}
	tm, err = time.ParseInLocation("2006-01-02 15:04", st.End, time.Local)
	if err == nil {
		endDate = tm.Unix()
	}

	if startDate >= endDate {
		endDate = startDate + 30*60
	}

	ui.SetColumn(0, 2, 4)
	ui.SetColumn(1, 3, 16)

	ui.AddTextLabel(0, 0, 2, 1, "Add new event into calendar")

	ui.AddText(0, 1, 1, 1, "Title")
	ui.AddEditboxString(1, 1, 1, 1, &st.Title)

	ui.AddText(0, 2, 1, 1, "Description")
	ui.AddEditboxString(1, 2, 1, 1, &st.Description)

	//start
	ui.AddText(0, 4, 1, 1, "Start")
	ui.AddDatePickerButton(1, 4, 1, 1, &startDate, nil, true)

	ui.AddText(0, 5, 1, 1, "End")
	ui.AddDatePickerButton(1, 5, 1, 1, &endDate, nil, true)

	ui.AddText(0, 7, 1, 1, "Group")
	groupID := strconv.FormatInt(st.GroupID, 10)
	cb := ui.AddDropDown(1, 7, 1, 1, &groupID, source_events.getGroupsLabels(), source_events.getGroupsValues())
	cb.changed = func() error {
		st.GroupID, _ = strconv.ParseInt(groupID, 10, 64)
		return nil
	}

	//files
	ui.AddText(0, 9, 1, 1, "Attachment(s)")
	//....

	bt := ui.AddButton(0, 12, 2, 1, "Add new Event")
	bt.clicked = func() error {
		//checks
		if startDate >= endDate {
			endDate = startDate + 30*60
		}

		//import files
		var files []string
		//for _, f := range st.Files {
		//copy files
		//}

		//update
		source_events.Events[time.Now().UnixNano()] = &EventsItem{Title: st.Title, Description: st.Description, Files: files, Start: startDate, Duration: (endDate - startDate), GroupID: st.GroupID}

		return nil
	}

	return nil
}
