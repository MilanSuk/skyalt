package main

import (
	"fmt"
)

type ShowUserBodyMeasurements struct {
	//No arguments required
}

func (st *ShowUserBodyMeasurements) run(caller *ToolCaller, ui *UI) error {
	body, err := LoadUserBodyMeasurements()
	if err != nil {
		return err
	}

	ui.addTextH1("User Body Measurements")

	form := ui.addTable("")

	{
		ln := form.addLine("Gender = " + body.Gender)
		ln.addText("Gender", "")
		ln.addDropDown(&body.Gender, []string{"Man", "Woman"}, []string{"man", "woman"}, "Gender = "+body.Gender)
	}

	{
		ln := form.addLine("BirthYear = " + fmt.Sprintf("%d", body.BirthYear))
		ln.addText("Year of birth", "")
		ln.addEditboxInt(&body.BirthYear, "BirthYear = "+fmt.Sprintf("%d", body.BirthYear))
	}

	{
		ln := form.addLine("Height = " + fmt.Sprintf("%.2f", body.Height))
		ln.addText("Height (m)", "")
		ln.addEditboxFloat(&body.Height, 2, "Height = "+fmt.Sprintf("%.2f", body.Height))
	}

	{
		ln := form.addLine("Weight = " + fmt.Sprintf("%.1f", body.Weight))
		ln.addText("Weight (kg)", "")
		ln.addEditboxFloat(&body.Weight, 1, "Weight = "+fmt.Sprintf("%.1f", body.Weight))
	}

	return nil
}
