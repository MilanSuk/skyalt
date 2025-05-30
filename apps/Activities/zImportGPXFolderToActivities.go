package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Form to import .gpx files from folder into database.
type ImportGPXFolderToActivities struct {
	// Path to folder. Optional, default is "". [optional]
	FolderPath string
}

func (st *ImportGPXFolderToActivities) run(caller *ToolCaller, ui *UI) error {
	source_activities, err := NewActivities("")
	if err != nil {
		return err
	}

	ui.SetColumn(0, 2, 4)
	ui.SetColumn(1, 3, 16)

	ui.AddTextLabel(0, 0, 2, 1, "Import activities from folder")

	ui.AddText(0, 1, 1, 1, "Folder")
	fpck := ui.AddFilePickerButton(1, 0, 1, 1, &st.FolderPath, true, true)

	bt := ui.AddButton(0, 3, 2, 1, "Import")
	bt.clicked = func() error {
		//checks
		if st.FolderPath == "" {
			fpck.Error = "Empty field"
			return fmt.Errorf("invalid input(s)")
		}

		//copy
		files, err := os.ReadDir(st.FolderPath)
		if err != nil {
			return err
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}
			ext := filepath.Ext(file.Name())
			if strings.ToLower(ext) != ".gpx" {
				continue
			}
			_, err := source_activities._importGPXFile(filepath.Join(st.FolderPath, file.Name()), "", "")
			if err != nil {
				return err
			}
		}

		return nil
	}

	return nil
}
