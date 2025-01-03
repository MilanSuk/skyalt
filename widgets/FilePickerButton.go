package main

import (
	"os"
)

type FilePickerButton struct {
	Path *string

	SelectFile   bool
	ErrWhenEmpty bool
	changed      func()
}

func (layout *Layout) AddFilePickerButton(x, y, w, h int, path *string, selectFile bool) *FilePickerButton {
	props := &FilePickerButton{Path: path, SelectFile: selectFile}
	layout._createDiv(x, y, w, h, "FilePickerButton", props.Build, nil, nil)
	return props
}

var g_file_picker_temp_path = ""

func (st *FilePickerButton) Build(layout *Layout) {

	layout.SetColumn(0, 1, 100)
	layout.SetRow(0, 1, 100)

	layout.dropFile = func(path string) {
		*st.Path = path
		if st.changed != nil {
			st.changed()
		}
	}

	cdialog := layout.AddDialog("dialog")
	{
		cdialog.Layout.SetColumn(0, 5, 20)
		cdialog.Layout.SetRow(1, 4, 12)

		//header
		{
			hDiv := cdialog.Layout.AddLayout(0, 0, 1, 1)
			hDiv.SetColumn(0, 1, 3)
			hDiv.SetColumn(1, 1, 100)
			hDiv.SetColumn(2, 1, 3)

			closeBt := hDiv.AddButton(0, 0, 1, 1, "Cancel")
			closeBt.clicked = func() {
				cdialog.Close()
			}

			nm := "Open Folder"
			if st.SelectFile {
				nm = "Open File"
			}
			tx := hDiv.AddText(1, 0, 1, 1, nm)
			tx.Align_h = 1

			openBt := hDiv.AddButton(2, 0, 1, 1, "Open")
			openBt.clicked = func() {
				*st.Path = g_file_picker_temp_path
				if st.changed != nil {
					st.changed()
				}
				cdialog.Close()
			}
		}

		picker := cdialog.Layout.AddFilePicker(0, 1, 1, 1, &g_file_picker_temp_path, st.SelectFile)
		picker.changed = func(close bool) {
			if close {
				cdialog.Close()
			}
		}
	}

	var exist bool
	if st.SelectFile {
		exist = FilePicker_FileExists(*st.Path)
	} else {
		exist = FilePicker_FolderExists(*st.Path)
	}

	icon := "resources/folder.png"
	if st.SelectFile {
		icon = "resources/file.png"
	}
	bt := layout.AddButtonMenu(0, 0, 1, 1, *st.Path, icon, 0.1)
	bt.Border = true
	bt.clicked = func() {
		if *st.Path == "" {
			dir, err := os.Getwd()
			if err == nil {
				g_file_picker_temp_path = dir
			}
		} else {
			g_file_picker_temp_path = *st.Path
		}
		cdialog.OpenRelative(layout)
	}

	if *st.Path == "" {
		if st.SelectFile {
			bt.Value = "< Select File >"
		} else {
			bt.Value = "< Select Folder >"
		}
		if !st.ErrWhenEmpty {
			exist = true
		}
	} else {
		if !exist {
			bt.Value = *st.Path + " not found!"
			bt.Cd = Paint_GetPalette().E
		}
	}
}
