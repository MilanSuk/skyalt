package main

import (
	"fmt"
)

// Shows person year of born and height. Properties are editable.
type ShowPersonInfo struct {
	PersonID string //ID of person to show
}

func (tool *ShowPersonInfo) run(caller *ToolCaller, ui *UI) error {
	people := LoadPeople()

	person, found := people.people[tool.PersonID]
	if !found {
		return fmt.Errorf("%d PersonID not found", tool.PersonID)
	}

	ui.addTextH1("Person infomation")

	{
		form := ui.addTable("Person infomation")
		ln := form.addLine(fmt.Sprintf("PersonID = %s", tool.PersonID))
		ln.addText("Born", "")                                                                     //description
		ln.addEditboxInt(&person.Born, fmt.Sprintf("Year of born for PersonID = %s", st.PersonID)) //value

		ln = form.addLine(fmt.Sprintf("PersonID = %s", tool.PersonID))
		ln.addText("Height", "")                                                                 //description
		ln.addEditboxInt(&person.Height, fmt.Sprintf("Height for PersonID = %s", tool.PersonID)) //value
	}

	return nil
}
