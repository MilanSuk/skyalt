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

	ui.addTextH1("Person infomation")

	{
		form := ui.addTable("Person infomation")
		ln := form.addLine(fmt.Sprintf("PersonID = %s", st.PersonID))
		ln.addText("Born", "")                 //description
		ln.addEditboxInt(&person.Born, "Born") //value

		ln = form.addLine(fmt.Sprintf("PersonID = %s", st.PersonID))
		ln.addText("Height", "")
		ln.addEditboxInt(&person.Height, "Height")
	}

	return nil
}
