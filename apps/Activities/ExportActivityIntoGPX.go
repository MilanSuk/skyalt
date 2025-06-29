package main

// Form to export activity from database into .gpx file.
type ExportActivityIntoGPX struct {
	ActivityID string //ID of the activity

	// Path to .gpx file.  Optional, default is "". [optional]
	FilePath string
}

func (st *ExportActivityIntoGPX) run(caller *ToolCaller, ui *UI) error {

	ui.SetColumn(0, 2, 4)
	ui.SetColumn(1, 3, 16)

	ui.AddTextLabel(0, 0, 2, 1, "Export activity")

	ui.AddText(0, 1, 1, 1, "File(.gpx)")
	ui.AddFilePickerButton(1, 0, 1, 1, &st.FilePath, true, false)

	bt := ui.AddButton(0, 3, 2, 1, "Export")
	bt.clicked = func() error {
		source_activities, err := NewActivities("")
		if err != nil {
			return err
		}

		//Save copy
		err = OsCopyFile(st.FilePath, source_activities.GetFilePath(st.ActivityID))
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}
