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
		form := ui.addTable()
		ln := form.addLine()
		ln.addText("Born")             //description
		ln.addEditboxInt(&person.Born) //value

		ln = form.addLine()
		ln.addText("Height")
		ln.addEditboxInt(&person.Height)
	}

	ui.addText(fmt.Sprintf("note: PersonID = %s", st.PersonID)).Align_h = 2 //note which is align to right side

	return nil
}
