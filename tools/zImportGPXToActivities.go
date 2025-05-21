package main

import "fmt"

// Form to import .gpx file into database.
type ImportGPXToActivities struct {
	FilePath string // Path to .gpx file. Optional, default is "". [optional]

	Type string //Type of the activity [optional] [options: "run", "ride", "swim", "walk"]

	Description string // Description of the activity. Optional, default is "". [optional]
}

func (st *ImportGPXToActivities) run(caller *ToolCaller, ui *UI) error {
	source_activities, err := NewActivities("", caller)
	if err != nil {
		return err
	}

	ui.SetColumn(0, 2, 4)
	ui.SetColumn(1, 3, 16)

	ui.AddTextLabel(0, 0, 2, 1, "Import new activity")

	ui.AddText(0, 1, 1, 1, "File(.gpx)")
	fpck := ui.AddFilePickerButton(1, 0, 1, 1, &st.FilePath, true, false)

	ui.AddText(0, 2, 1, 1, "Type")
	ui.AddCombo(1, 2, 1, 1, &st.Type, source_activities.GetTypeLabels(), source_activities.GetTypeValues())

	ui.AddText(0, 3, 1, 1, "Description")
	ui.AddEditboxString(1, 3, 1, 1, &st.Description)

	bt := ui.AddButton(0, 5, 2, 1, "Import")
	bt.clicked = func() error {
		//checks
		if st.FilePath == "" {
			fpck.Error = "Empty field"
			return fmt.Errorf("invalid input(s)")
		}

		_, err := source_activities._importGPXFile(st.FilePath, st.Type, st.Description, caller)
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}
