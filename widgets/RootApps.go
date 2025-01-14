package main

import "path/filepath"

type RootApps struct {
	Search string
}

func (layout *Layout) AddRootApps(x, y, w, h int) *RootApps {
	props := &RootApps{}
	layout._createDiv(x, y, w, h, "RootApps", props.Build, nil, nil)
	return props
}

func (st *RootApps) Build(layout *Layout) {

	//center
	layout.SetColumn(0, 1, 100)
	layout.SetColumn(1, 1, 20)
	layout.SetColumn(2, 1, 100)
	layout.SetRow(1, 1, 100)

	searchDiv := layout.AddLayout(1, 0, 1, 1)
	searchDiv.SetColumn(0, 1, 100)
	searchDiv.SetColumn(1, 1, 10)
	searchDiv.SetColumn(2, 1, 100)
	search := searchDiv.AddSearch(1, 0, 1, 1, &st.Search, "")

	appDiv := layout.AddLayout(1, 1, 1, 1)
	appDiv.SetColumn(0, 1, 100)
	appDiv.SetRow(0, 1, 100)

	MidDiv := appDiv.AddLayout(0, 0, 1, 1)
	MidDiv.SetColumn(0, 1, 100)

	MidDiv.AddText(0, 0, 1, 1, "Default apps").Align_h = 1
	y := 1
	var firstFolder string
	var firstApp string
	st.addFiles("", MidDiv, &y, &firstFolder, &firstApp)

	search.enter = func() {
		if firstApp != "" {
			OpenFile_Root().OpenApp(firstFolder, firstApp)
		}
	}

}

func (st *RootApps) addFiles(folder string, layout *Layout, y *int, firstFolder *string, firstApp *string) {

	searchWords := Search_Prepare(st.Search)

	files, dirs := GetListOfFiles(folder)

	list := layout.AddLayoutList(0, *y, 1, 1, true)
	//layout.SetRow(*y, 5, 5)
	layout.SetRowFromSub(*y, 1, 10*2) //2=item size
	(*y)++
	for _, file := range files {
		if !Search_Find(file.Name, searchWords) {
			continue
		}

		sub := list.AddListSubItem()
		item_sz := 3.0
		sub.SetColumn(0, item_sz, item_sz)
		sub.SetRow(0, item_sz, item_sz)

		bt := sub.AddButton(0, 0, 1, 1, file.Name)
		bt.Background = 0.5
		bt.clicked = func() {
			OpenFile_Root().OpenApp(folder, file.Name)
		}

		if *firstApp == "" {
			*firstFolder = folder
			*firstApp = file.Name
		}
	}

	for _, dir := range dirs {
		layout.AddDivider(0, *y, 1, 1, true)
		(*y)++
		layout.AddText(0, *y, 1, 1, dir).Align_h = 1
		(*y)++

		st.addFiles(filepath.Join(folder, dir), layout, y, firstFolder, firstApp)
	}
}
