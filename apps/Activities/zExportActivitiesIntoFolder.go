package main

import (
	"fmt"
	"path/filepath"
	"time"
)

// Form to filter activities(date range) and export them into folder as .gpx file. Form also shows the list of exported activities.
type ExportActivitiesIntoFolder struct {
	DateStart string //Select only runs after this data. Ignore = empty string. Date format: YYYY-MM-DD HH:MM
	DateEnd   string //Select only runs before this data. Ignore = empty string. Date format: YYYY-MM-DD HH:MM

	// Path to folder. Optional, default is "". [optional]
	FolderPath string
}

func (st *ExportActivitiesIntoFolder) run(caller *ToolCaller, ui *UI) error {
	source_activities, err := NewActivities("")
	if err != nil {
		return err
	}

	sorted, err := source_activities.Filter(st.DateStart, st.DateEnd, "", false, 0)
	if err != nil {
		return err
	}

	ui.SetColumn(0, 2, 4)
	ui.SetColumn(1, 2, 4)
	ui.SetColumn(2, 3, 8)

	ui.AddTextLabel(0, 0, 3, 1, "Export activities into folder")

	ui.AddText(0, 1, 1, 1, "Folder")
	fpck := ui.AddFilePickerButton(1, 0, 2, 1, &st.FolderPath, true, true)

	y := 3
	for _, id := range sorted {
		it := source_activities.Activities[id]
		ui.AddText(0, y, 1, 1, caller.ConvertTextDate(int64(it.Date)))
		ui.AddText(1, y, 1, 1, it.Type)
		ui.AddText(2, y, 1, 1, it.Description)
		y++
	}

	y++ //space

	bt := ui.AddButton(0, y, 3, 1, "Export")
	bt.clicked = func() error {
		//checks
		if st.FolderPath == "" {
			fpck.Error = "Empty field"
			return fmt.Errorf("invalid input(s)")
		}

		//Items
		for _, id := range sorted {
			it := source_activities.Activities[id]

			//Save copy
			err := OsCopyFile(filepath.Join(st.FolderPath, time.Unix(int64(it.Date), 0).Format("Mon, 02 Jan 2006 15:04")+".gpx"), source_activities.GetFilePath(id))
			if err != nil {
				return err
			}
		}

		return nil
	}

	return nil
}
