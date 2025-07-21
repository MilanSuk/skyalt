package main

// Shows settings for user's body measurements.
type ShowUserBodyMeasurements struct {
}

func (st *ShowUserBodyMeasurements) run(caller *ToolCaller, ui *UI) error {
	source_body, err := NewUserBodyMeasurements("")
	if err != nil {
		return err
	}

	ui.SetColumn(0, 2, 4)
	ui.SetColumn(1, 3, 16)

	ui.AddTextLabel(0, 0, 2, 1, "Edit body measurements")

	ui.AddText(0, 1, 1, 1, "Gender")
	var gender string
	if source_body.Female {
		gender = "female"
	} else {
		gender = "male"
	}

	gen := ui.AddDropDown(1, 1, 1, 1, &gender, []string{"Male", "Female"}, []string{"male", "female"})
	//gen.DialogWidth = 4
	gen.changed = func() error {
		source_body.Female = (gender == "female")
		return nil
	}

	year := source_body.BornYear
	ui.AddText(0, 2, 1, 1, "Year of birth")
	ed := ui.AddEditboxInt(1, 1, 1, 1, &year)
	ed.Ghost = "2000"
	ed.changed = func() error {
		if year < 1 {
			year = 1
		}
		source_body.BornYear = year
		return nil
	}

	height := source_body.Height
	ui.AddText(0, 3, 1, 1, "Height(meters)")
	ed = ui.AddEditboxFloat(1, 2, 1, 1, &height, 2)
	ed.Ghost = "1.7m"
	ed.changed = func() error {
		if height < 1 {
			height = 1
		}
		source_body.Height = height
		return nil
	}

	weight := source_body.Weight
	ui.AddText(0, 4, 1, 1, "Weight(kg)")
	ed = ui.AddEditboxFloat(1, 3, 1, 1, &weight, 2)
	ed.Ghost = "60 kg"
	ed.changed = func() error {
		if weight < 1 {
			weight = 1
		}
		source_body.Weight = weight
		return nil
	}

	return nil
}
