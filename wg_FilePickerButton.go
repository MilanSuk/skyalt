package main

import (
	"fmt"
	"os"
)

type FilePickerButton struct {
	Tooltip     string
	Path        *string
	Preview     bool
	OnlyFolders bool

	changed func()
}

func (layout *Layout) AddFilePickerButton(x, y, w, h int, path *string, preview bool, onlyFolders bool) *FilePickerButton {
	props := &FilePickerButton{Path: path, Preview: preview, OnlyFolders: onlyFolders}
	lay := layout._createDiv(x, y, w, h, "FilePickerButton", props.Build, nil, nil)
	lay.fnGetLLMTip = props.getLLMTip
	return props
}

func (st *FilePickerButton) getLLMTip(layout *Layout) string {
	return fmt.Sprintf("Type: FilePickerButton. Value: %s. Tooltip: %s", *st.Path, st.Tooltip)
}

func (st *FilePickerButton) Build(layout *Layout) {
	layout.SetColumn(0, 1, 100)
	layout.SetRow(0, 1, 100)

	layout.dropFile = func(path string) {
		*st.Path = path
		if st.changed != nil {
			st.changed()
		}
	}

	dropLabel := OsTrnString(st.OnlyFolders, "Drop folder(s) here", "Drop file(s) here")

	cdialog := layout.AddDialog("dialog")
	{
		cdialog.Layout.SetColumn(0, 3, 7)
		cdialog.Layout.SetRow(0, 3, 7)

		cdialog.Layout.AddText(0, 0, 1, 1, dropLabel).Align_h = 1
		cdialog.Layout.dropFile = func(path string) {
			*st.Path = path
			if st.changed != nil {
				st.changed()
			}
			//cdialog.Close()	//can't close, because dropFile() can be called multiple times for multiple files
		}
	}

	var exist bool
	if st.Path != nil {
		_, err := os.Stat(*st.Path)
		exist = (err == nil && !os.IsNotExist(err))

	}

	label := "< empty >"
	if *st.Path != "" {
		label = *st.Path
	}

	bt := layout.AddButtonMenu(0, 0, 1, 1, OsTrnString(st.Preview, label, ""), "resources/attachment.png", 0.15)
	//bt.Border = true
	bt.Tooltip = dropLabel

	if st.Preview && label != "" && !exist {
		bt.Cd = layout.GetPalette().E
		bt.Tooltip = "Not found!"
	}

	bt.clicked = func() {
		cdialog.OpenRelative(layout.UID)
	}
}
