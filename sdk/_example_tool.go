package main

import (
	"fmt"
)

type ShowPersonInfo struct {
	PersonID string
}

// Shows person year of born and height. Properties are editable.
func (st *ShowPersonInfo) run(caller *ToolCaller, ui *UI) error {
	people := LoadPeople()

	person, found := people.people[st.PersonID]
	if !found {
		return fmt.Errorf("%d PersonID not found", st.PersonID)
	}

	desc_width := 5 //description column width

	r := ui.addRow()                  //add new row
	r.addText(desc_width, "Born")     //add description 'Born'
	r.addEditboxInt(-1, &person.Born) //add editable text field, width(-1) is automatic

	r = ui.addRow()                     //new line
	r.addText(desc_width, "Height")     //add description 'Height'
	r.addEditboxInt(-1, &person.Height) //add editable text field, width(-1) is automatic

	return nil
}
