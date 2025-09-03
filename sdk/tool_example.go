package main

import (
	"fmt"
)

// Shows person year of born and height. Properties are editable.
type ShowPersonInfo struct {
	PersonID string //ID of person to show
}

func (tool *ShowPersonInfo) run(caller *ToolCaller, ui *UI) error {
	people, err := LoadPeople() //LoadPeople() is implemented in other file
	if err != nil {
		return err
	}

	person, found := people.people[tool.PersonID]
	if !found {
		return fmt.Errorf("%d PersonID not found", tool.PersonID)
	}

	ui.addTextH1("Person infomation")

	{
		form := ui.addTable("Person infomation")
		ln := form.addLine(fmt.Sprintf("PersonID = %s", tool.PersonID))
		ln.addText("Born", "") //description
		ln.addEditboxInt(person.Born, func(newValue int, self *UIEditbox) {
			person.Born = newValue
		}, fmt.Sprintf("Year of born for PersonID = %s", tool.PersonID)) //value

		ln = form.addLine(fmt.Sprintf("PersonID = %s", tool.PersonID))
		ln.addText("Height", "") //description
		ln.addEditboxInt(person.Height, func(newValue int, self *UIEditbox) {
			person.Height = newValue
		}, fmt.Sprintf("Height for PersonID = %s", tool.PersonID)) //value
	}

	return nil
}
