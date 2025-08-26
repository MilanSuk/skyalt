package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type FilePicker struct {
	Path *string

	SelectFile bool

	changed func(close bool)
}

func (layout *Layout) AddFilePicker(x, y, w, h int, path *string, selectFile bool) *FilePicker {
	props := &FilePicker{Path: path, SelectFile: selectFile}
	layout._createDiv(x, y, w, h, "FilePicker", props.Build, nil, nil)
	return props
}

var g_FilePicker2_search string

func (st *FilePicker) Build(layout *Layout) {

	//check
	var isFile bool
	{
		if *st.Path == "" {
			*st.Path = "/"
		}
		isFile = FilePicker_FileExists(*st.Path)
		isFolder := FilePicker_FolderExists(*st.Path)

		if !isFile && !isFolder {
			//get skyalt dir
			ex, err := os.Executable()
			if err == nil {
				*st.Path = filepath.Dir(ex)
			}
		}
	}

	folder := *st.Path
	if isFile {
		folder = filepath.Dir(*st.Path)
	}

	layout.SetColumn(4, 2, 2)
	layout.SetColumn(5, 1, Layout_MAX_SIZE)

	layout.SetRow(1, 2, 10)

	Root := layout.AddButton(0, 0, 1, 1, "/")
	CurDir := layout.AddButton(1, 0, 1, 1, "SA")
	Home := layout.AddButtonIcon(2, 0, 1, 1, "resources/home.png", 0.25, "Home directory")
	Levelup, LevelupL := layout.AddButtonIcon2(3, 0, 1, 1, "resources/levelup.png", 0.25, "Jump into parent directory")
	layout.AddEditbox(4, 0, 2, 1, st.Path) //path editbox

	CreateFile, CreateFileL := layout.AddButton2(0, 2, 3, 1, "Create File")
	CreateFolder := layout.AddButton(3, 2, 2, 1, "CreateFolder")
	EditSearch := layout.AddSearch(5, 2, 1, 1, &g_FilePicker2_search, "Search files")

	Root.Background = 0
	Home.Background = 0
	CurDir.Background = 0
	Levelup.Background = 0

	Root.Tooltip = "Root directory"
	CurDir.Tooltip = "Skyalt directory"

	EditSearch.Ghost = "Search"

	LevelupL.Enable = (*st.Path != "/")

	CreateFile.Background = 0.5
	CreateFolder.Background = 0.5
	CreateFileL.Enable = st.SelectFile

	Root.clicked = func() {
		*st.Path = "/"
	}
	Home.clicked = func() {
		dir, err := os.UserHomeDir()
		if err == nil {
			*st.Path = dir
		}
	}
	CurDir.clicked = func() {
		dir, err := os.Getwd()
		if err == nil {
			*st.Path = dir
		}
	}
	Levelup.clicked = func() {
		*st.Path = filepath.Dir(folder)
	}

	//create file
	{
		createDia := layout.AddDialog("create_file")
		createDia.Layout.SetColumn(0, 3, 7)
		createDia.Layout.SetColumn(1, 3, 3)
		var file_name string
		createDia.Layout.AddEditbox(0, 0, 1, 1, &file_name)
		bt := createDia.Layout.AddButton(1, 0, 1, 1, "Create")
		bt.clicked = func() {
			newPath := folder + "/" + file_name
			file, err := os.Create(newPath)
			if err == nil {
				*st.Path = newPath
				createDia.Close(layout.ui)
			}
			defer file.Close()
		}
		CreateFile.clicked = func() {
			createDia.OpenRelative(CreateFileL.UID)
		}
	}

	//create folder
	{
		createDia := layout.AddDialog("create_folder")
		createDia.Layout.SetColumn(0, 3, 7)
		createDia.Layout.SetColumn(1, 3, 3)
		var file_name string
		createDia.Layout.AddEditbox(0, 0, 1, 1, &file_name)
		bt := createDia.Layout.AddButton(1, 0, 1, 1, "Create")
		bt.clicked = func() {
			newPath := folder + "/" + file_name
			err := os.MkdirAll(newPath, os.ModePerm)
			if err == nil {
				*st.Path = newPath
				createDia.Close(layout.ui)
			}
		}
		CreateFolder.clicked = func() {
			createDia.OpenRelative(CreateFileL.UID)
		}
	}

	List := layout.AddLayout(0, 1, 6, 1)
	List.SetColumn(0, 3, Layout_MAX_SIZE)
	List.SetColumn(1, 2, 2)
	List.SetColumn(2, 3, 3.5)

	{
		folder := *st.Path
		if isFile {
			folder = filepath.Dir(*st.Path)
		}
		dir, err := os.ReadDir(folder)
		if err != nil {
			//....
			return
		}

		EditSearch.RefreshValueFromTemp()

		searchWords := Search_Prepare(g_FilePicker2_search)
		y := 0
		for _, f := range dir {
			fileName := f.Name()
			isDir := f.IsDir()
			enable := st.SelectFile || isDir //show both, but enable only what can be selected

			if !Search_Find(fileName, searchWords) {
				continue
			}

			iconFile := "file.png"
			if isDir {
				iconFile = "folder.png"
			}

			inf, _ := f.Info()

			selected := (folder+"/"+fileName == *st.Path)

			bt, btL := List.AddButton2(0, y, 1, 1, fileName)
			bt.Align = 0
			btL.Enable = enable
			bt.IconPath = "resources/" + iconFile
			bt.Icon_margin = 0.17
			if selected {
				bt.Background = 1
			} else {
				bt.Background = 0.25
			}

			bt.clickedEx = func(numClicks int, altClick bool) {
				path := filepath.Join(folder, fileName)
				isDir := FilePicker_FolderExists(path)

				if isDir {
					*st.Path = path
					g_FilePicker2_search = "" //reset
				} else {
					*st.Path = path
				}

				if st.changed != nil {
					st.changed(numClicks > 1) //+close dialog
				}
			}

			List.AddText(1, y, 1, 1, FilePicker_ConvertBytesToString(int(inf.Size()))) //size
			List.AddText(2, y, 1, 1, layout.ConvertTextDateTime(inf.ModTime().Unix())) //date

			y++
		}
	}
	//slow: for folder with many files? ....
}

func FilePicker_FileExists(fileName string) bool {
	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func FilePicker_FolderExists(fileName string) bool {
	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func FilePicker_ConvertBytesToString(bytes int) string {
	var str string
	if bytes >= 1000*1000*1000 {
		str = fmt.Sprintf("%.1f GB", float64(bytes)/(1000*1000*1000))
	} else if bytes >= 1000*1000 {
		str = fmt.Sprintf("%.1f MB", float64(bytes)/(1000*1000))
	} else if bytes >= 1000 {
		str = fmt.Sprintf("%.1f KB", float64(bytes)/(1000))
	} else {
		str = fmt.Sprintf("%d B", bytes)
	}
	return str
}
