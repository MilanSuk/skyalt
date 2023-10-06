/*
Copyright 2023 Milan Suk

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this db except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Log struct {
	Time int64
	Text string
}

type File struct {
	Name         string
	Expand       bool
	initAppTable bool
	id           int
}

func FindInArray(arr []string, name string) int {
	for i, it := range arr {
		if it == name {
			return i
		}
	}
	return -1
}

func FindSelectedFile() *File {

	if store.SelectedFile < 0 {
		store.SelectedFile = 0
	}

	if store.SelectedFile >= len(store.Files) {
		store.SelectedFile = len(store.Files) - 1 //= -1
	}

	if store.SelectedFile >= 0 {
		return store.Files[store.SelectedFile]
	}

	return nil
}

func FindFile(name string) *File {
	for _, f := range store.Files {
		if f.Name == name {
			return f
		}
	}
	return nil
}

type Storage struct {
	Files []*File

	SelectedFile int
	SelectedApp  int

	SearchFiles string
	SearchApp   string

	createFile    string
	duplicateName string

	last_file_id int
}

var store Storage

type Translations struct {
	SAVE            string
	SETTINGS        string
	ZOOM            string
	WINDOW_MODE     string
	FULLSCREEN_MODE string
	ABOUT           string
	QUIT            string
	SEARCH          string

	COPYRIGHT string
	WARRANTY  string

	DATE_FORMAT      string
	DATE_FORMAT_EU   string
	DATE_FORMAT_US   string
	DATE_FORMAT_ISO  string
	DATE_FORMAT_TEXT string

	THEME       string
	THEME_OCEAN string
	THEME_RED   string
	THEME_BLUE  string
	THEME_GREEN string
	THEME_GREY  string

	DPI        string
	SHOW_STATS string
	SHOW_GRID  string
	LANGUAGES  string

	NAME        string
	REMOVE      string
	RENAME      string
	DUPLICATE   string
	CREATE_FILE string

	ALREADY_EXISTS string
	EMPTY_FIELD    string
	INVALID_NAME   string

	IN_USE string

	ADD_APP   string
	CREATE_DB string

	LOGS string
}

var trns Translations

// https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes
const g_langs = "|English|Chinese(中文)|Hindi(हिंदी)|Spanish(Español)|Russian(Руштина)|Czech(Česky)"

var g_lang_codes = []string{"", "en", "zh", "hi", "es", "ru", "cs"}

func FindLangCode(lng string) int {
	for ii, cd := range g_lang_codes {
		if cd == lng {
			return ii
		}
	}
	return 0
}

func Settings() {
	SA_ColMax(1, 12)
	SA_ColMax(2, 1)

	y := 0

	SA_TextCenter(trns.SETTINGS).Show(1, 0, 1, 1)
	y++

	//languages
	{
		SA_TextCenter(trns.LANGUAGES).Show(1, y, 1, 1)
		y++

		inf_langs := SA_Info("languages")
		var langs []string
		if len(inf_langs) > 0 {
			langs = strings.Split(inf_langs, "/")
		}
		for i, lng := range langs {

			lang_id := FindLangCode(lng)

			SA_DivStart(1, y, 1, 1)
			{
				SA_ColMax(2, 100)
				changed := false

				SA_TextCenter(strconv.Itoa(i+1)+".").Show(0, 0, 1, 1)

				SA_DivStart(1, 0, 1, 1)
				{
					SA_Div_SetDrag("lang", uint64(i))
					src, pos, done := SA_Div_IsDrop("lang", true, false, false)
					if done {
						SA_MoveElement(&langs, &langs, int(src), i, pos)
						changed = true
					}
					SA_Image("app:resources/reorder.png").Margin(0.15).Show(0, 0, 1, 1)
				}
				SA_DivEnd()

				if SA_Combo(&lang_id, g_langs).Search(true).Show(2, 0, 1, 1) {
					langs[i] = g_lang_codes[lang_id]
					changed = true
				}

				if SA_ButtonLight("X").Enable(len(langs) > 1 || i > 0).Show(3, 0, 1, 1).click {
					langs = append(langs[:i], langs[i+1:]...)
					changed = true
				}

				if changed {
					ll := ""
					for _, lng := range langs {
						ll += lng + "/"
					}
					SA_InfoSet("languages", strings.TrimSuffix(ll, "/"))

					SA_DivEnd() //!
					break
				}
			}
			SA_DivEnd()
			i++
			y++
		}

		SA_DivStart(1, y, 1, 1)
		if SA_ButtonLight("+").Show(0, 0, 1, 1).click {
			SA_InfoSet("languages", SA_Info("languages")+"/")
		}
		y++
		SA_DivEnd()

		y++ //space
	}

	date := int(SA_InfoFloat("date"))
	if SA_Combo(&date, trns.DATE_FORMAT_EU+"|"+trns.DATE_FORMAT_US+"|"+trns.DATE_FORMAT_ISO+"|"+trns.DATE_FORMAT_TEXT).Search(true).ShowDescription(1, y, 1, 2, trns.DATE_FORMAT, 0, nil) {
		SA_InfoSetFloat("date", float64(date))
	}
	y += 3

	theme := int(SA_InfoFloat("theme"))
	if SA_Combo(&theme, trns.THEME_OCEAN+"|"+trns.THEME_RED+"|"+trns.THEME_BLUE+"|"+trns.THEME_GREEN+"|"+trns.THEME_GREY).Search(true).ShowDescription(1, y, 1, 2, trns.THEME, 0, nil) {
		SA_InfoSetFloat("theme", float64(theme))
	}
	y += 3

	dpi := strconv.Itoa(int(SA_InfoFloat("dpi")))
	if SA_Editbox(&dpi).ShowDescription(1, y, 1, 2, trns.DPI, 4, nil).finished {
		dpiV, err := strconv.Atoi(dpi)
		if err == nil {
			SA_InfoSetFloat("dpi", float64(dpiV))
		}
	}
	y += 2

	{
		stats := false
		if SA_InfoFloat("stats") > 0 {
			stats = true
		}
		if SA_Checkbox(&stats, trns.SHOW_STATS).Show(1, y, 1, 1) {
			statsV := 0
			if stats {
				statsV = 1
			}
			SA_InfoSetFloat("stats", float64(statsV))
		}
	}
	y++

	{
		grid := false
		if SA_InfoFloat("grid") > 0 {
			grid = true
		}
		if SA_Checkbox(&grid, trns.SHOW_GRID).Show(1, y, 1, 1) {
			gridV := 0
			if grid {
				gridV = 1
			}
			SA_InfoSetFloat("grid", float64(gridV))
		}
	}
	y++

}

func About() {
	SA_ColMax(0, 15)
	SA_Row(1, 3)

	SA_TextCenter(trns.ABOUT).Show(0, 0, 1, 1)

	SA_Image("app:resources/logo.png").Show(0, 1, 1, 1)

	SA_TextCenter("v0.3").Show(0, 2, 1, 1)

	SA_ButtonAlpha("www.skyalt.com").Url("https://www.skyalt.com").Show(0, 3, 1, 1)
	SA_ButtonAlpha("github.com/milansuk/skyalt/").Url("https://github.com/milansuk/skyalt/").Show(0, 4, 1, 1)

	SA_TextCenter(trns.COPYRIGHT).Show(0, 5, 1, 1)
	SA_TextCenter(trns.WARRANTY).Show(0, 6, 1, 1)
}

func Menu() {
	SA_ColMax(0, 8)
	SA_Row(1, 0.2)
	SA_Row(3, 0.2)
	SA_Row(5, 0.2)
	SA_Row(7, 0.2)
	SA_Row(9, 0.2)

	//save
	if SA_ButtonMenu(trns.SAVE).Show(0, 0, 1, 1).click {
		SA_InfoSetFloat("save", 1)
		SA_DialogClose()
	}

	SA_RowSpacer(0, 1, 1, 1)

	//settings
	if SA_ButtonMenu(trns.SETTINGS).Show(0, 2, 1, 1).click {
		SA_DialogClose()
		SA_DialogOpen("Settings", 0)
	}

	SA_RowSpacer(0, 3, 1, 1)

	//zoom
	SA_DivStart(0, 4, 1, 1)
	{
		SA_ColMax(0, 100)
		SA_ColMax(2, 2)

		SA_Text(trns.ZOOM).Show(0, 0, 1, 1)

		dpi := SA_InfoFloat("dpi")
		dpi_default := SA_InfoFloat("dpi_default")
		if SA_ButtonBorder("+").Show(1, 0, 1, 1).click {
			SA_InfoSetFloat("dpi", dpi+3)
		}
		dpiV := int(dpi / dpi_default * 100)
		SA_TextCenter(strconv.Itoa(dpiV)+"%").Show(2, 0, 1, 1)
		if SA_ButtonBorder("-").Show(3, 0, 1, 1).click {
			SA_InfoSetFloat("dpi", dpi-3)
		}
	}
	SA_DivEnd()

	SA_RowSpacer(0, 5, 1, 1)

	//window/fullscreen switch
	{
		isFullscreen := SA_InfoFloat("fullscreen")
		ff := trns.WINDOW_MODE
		if isFullscreen == 0 {
			ff = trns.FULLSCREEN_MODE
		}
		if SA_ButtonMenu(ff).Show(0, 6, 1, 1).click {
			if isFullscreen > 0 {
				isFullscreen = 0
			} else {
				isFullscreen = 1
			}
			SA_InfoSetFloat("fullscreen", isFullscreen)
		}
	}

	SA_RowSpacer(0, 7, 1, 1)

	if SA_ButtonMenu(trns.ABOUT).Show(0, 8, 1, 1).click {
		SA_DialogClose()
		SA_DialogOpen("About", 0)
	}

	SA_RowSpacer(0, 9, 1, 1)

	if SA_ButtonMenu(trns.QUIT).Show(0, 10, 1, 1).click {
		SA_InfoSetFloat("exit", 1)
		SA_DialogClose()
	}

}

func GetFileApps(file *File) []string {
	var list []string

	q := SA_SqlRead("dbs:"+file.Name, "SELECT app FROM __skyalt__  WHERE label != '__default__'")
	var app string
	for q.Next(&app) {
		list = append(list, app)
	}
	return list
}
func GetFileNumApps(file *File) int {
	var num int
	q := SA_SqlRead("dbs:"+file.Name, "SELECT COUNT(*) FROM __skyalt__ WHERE label != '__default__'")
	q.Next(&num)
	return num
}

func WriteApp(file *File, query string, commit bool) int {
	insRow := SA_SqlWrite("dbs:"+file.Name, query)

	if commit {
		SA_SqlCommit("dbs:" + file.Name)
	}

	return int(insRow)
}

func FindOrAddDefaultApp(file *File) int {

	//SA_SqlCommit("dbs:" + file.Name)

	//find
	q := SA_SqlRead("dbs:"+file.Name, "SELECT rowid FROM __skyalt__ WHERE label='__default__' AND app='db'")
	var rowid int
	if q.Next(&rowid) {
		return rowid
	}

	//insert
	return WriteApp(file, "INSERT INTO __skyalt__(label, app, sort) VALUES('__default__','db', -1);", true)
}

func RefreshSort(file *File) {

	q := SA_SqlRead("dbs:"+file.Name, "SELECT rowid FROM __skyalt__ ORDER BY sort")
	var rowid int
	i := 1.0
	for q.Next(&rowid) {
		WriteApp(file, fmt.Sprintf("UPDATE __skyalt__ SET sort=%f WHERE rowid=%d", i, rowid), false)
		i++
	}

	SA_SqlCommit("dbs:" + file.Name)
}

func MoveApp(src_file *File, dst_file *File, src int, dst int, pos SA_Drop_POS) {

	dst_sort := 1000.0
	if pos != SA_Drop_INSIDE {
		q := SA_SqlRead("dbs:"+dst_file.Name, fmt.Sprintf("SELECT sort FROM __skyalt__ WHERE rowid=%d", dst))
		q.Next(&dst_sort)
		if pos == SA_Drop_V_LEFT || pos == SA_Drop_H_LEFT {
			dst_sort -= 0.5
		} else {
			dst_sort += 0.5
		}
	}

	if src_file == dst_file {
		//update
		WriteApp(src_file, fmt.Sprintf("UPDATE __skyalt__ SET sort=%f WHERE rowid=%d", dst_sort, src), true)

		//refresh
		RefreshSort(src_file)
	} else {

		//backup
		q := SA_SqlRead("dbs:"+src_file.Name, fmt.Sprintf("SELECT label, app, storage FROM __skyalt__ WHERE rowid=%d;", src))
		var app_label string
		var app_name string
		var app_storage []byte
		q.Next(&app_label, &app_name, &app_storage)

		//remove
		WriteApp(src_file, fmt.Sprintf("DELETE FROM __skyalt__ WHERE rowid=%d;", src), true)

		//add
		WriteApp(dst_file, fmt.Sprintf("INSERT INTO __skyalt__(label, app, sort, storage) VALUES('%s','%s', %f, '%s');", app_label, app_name, dst_sort, string(app_storage)), true)

		//refresh
		RefreshSort(src_file)
		RefreshSort(dst_file)
	}

	dst_file.Expand = true
}

func AppList(file *File, file_i int) {
	SA_ColMax(0, 7)

	y := 0
	SA_Editbox(&store.SearchApp).TempToValue(true).Ghost(trns.SEARCH).Show(0, 0, 1, 1)
	y++

	appsInUse := GetFileApps(file)

	inf_apps := SA_Info("apps")
	var apps []string
	if len(inf_apps) > 0 {
		apps = strings.Split(inf_apps, "/")
	}
	for _, app := range apps {

		if len(store.SearchApp) > 0 {
			if !strings.Contains(strings.ToLower(app), strings.ToLower(store.SearchApp)) {
				continue
			}
		}

		nm := app
		for _, dapp := range appsInUse {
			if dapp == app {
				nm += "(" + trns.IN_USE + ")"
				break
			}
		}

		if SA_ButtonAlpha(nm).Show(0, y, 1, 1).click {
			WriteApp(file, fmt.Sprintf("INSERT INTO __skyalt__(label, app, sort) VALUES('%s','%s',%d);", app, app, 1000), true)
			RefreshSort(file)
			file.Expand = true
			SA_DialogClose()
		}
		y++
	}

}

func ProjectFiles() {
	inf_files := SA_Info("files")
	var files []string
	if len(inf_files) > 0 {
		files = strings.Split(inf_files, "/")
	}

	//add
	for _, nm := range files {
		if FindFile(nm) == nil {
			store.Files = append(store.Files, &File{Name: nm, Expand: true, id: store.last_file_id})
			store.SelectedFile = len(store.Files) - 1
			store.last_file_id++
		}
	}
	//remove
	for i := len(store.Files) - 1; i >= 0; i-- {
		f := store.Files[i]
		if FindInArray(files, f.Name) < 0 {
			store.Files = append(store.Files[:i], store.Files[i+1:]...) //remove
		}
	}

	//check
	for _, file := range store.Files {
		if !file.initAppTable {
			WriteApp(file, "CREATE TABLE IF NOT EXISTS __skyalt__(label TEXT NOT NULL, sort REAL NOT NULL, app TEXT NOT NULL, storage BLOB, gui BLOB);", true)
			file.initAppTable = true
		}

		if file.id == 0 {
			file.id = store.last_file_id
			store.last_file_id++
		}
	}
}

func Files() {

	ProjectFiles()

	SA_ColMax(0, 100)
	y := 0
	for file_i, file := range store.Files {

		if file.Name == "base.sqlite" {
			continue
		}

		if len(store.SearchFiles) > 0 {
			if !strings.Contains(strings.ToLower(file.Name), strings.ToLower(store.SearchFiles)) {
				continue
			}
		}

		SA_DivStart(0, y, 1, 1)
		{
			SA_Col(0, 0.8)
			SA_ColMax(1, 100)

			isSelected := (file_i == store.SelectedFile && store.SelectedApp < 0)
			if isSelected {
				SAPaint_Rect(0, 0, 1, 1, 0, SA_ThemeCd().Aprox(SA_ThemeWhite(), 0.8), 0)
			}

			num_apps := GetFileNumApps(file)
			if num_apps == 0 {
				file.Expand = false
			}
			iconName := "tree_hide.png"
			if !file.Expand {
				iconName = "tree_show.png"
			}
			if SA_ButtonAlpha("").Enable(num_apps > 0).Icon("app:resources/"+iconName, 0.05).Show(0, 0, 1, 1).click {
				file.Expand = !file.Expand
			}

			//drop app on file
			SA_DivStart(1, 0, 1, 1)
			{
				src, _, done := SA_Div_IsDrop("app", false, false, true)
				if done {
					src_file := store.Files[uint32(src>>32)]
					src_app_rowid := uint32(src)

					MoveApp(src_file, file, int(src_app_rowid), 1000, SA_Drop_INSIDE) //1000=end

					file.Expand = true
				}
			}
			SA_DivEnd()

			//name
			SA_DivStart(1, 0, 1, 1)
			{
				//cut .ext
				nm := file.Name
				if strings.HasSuffix(nm, ".sqlite") {
					nm = nm[:len(nm)-7]
				}

				SA_ColMax(0, 100)
				if SA_ButtonMenu(nm).Pressed(isSelected).Show(0, 0, 1, 1).click {
					store.SelectedFile = file_i
					store.SelectedApp = -1

					if SA_DivInfoPos("touchClicks", 0, 0) > 1 { //double click
						SA_DialogOpen("RenameFile_"+file.Name, 1)
					}
				}
				SA_Div_SetDrag("file", uint64(file_i))
				src, pos, done := SA_Div_IsDrop("file", true, false, false)
				if done {
					SA_MoveElement(&store.Files, &store.Files, int(src), file_i, pos)
				}
			}
			SA_DivEnd()

			//add app
			if SA_ButtonStyle("+", &g_ButtonAddApp).Tooltip(trns.ADD_APP).Show(2, 0, 1, 1).click {
				SA_DialogOpen("apps_"+file.Name, 1)
			}
			if SA_DialogStart("apps_" + file.Name) {
				AppList(file, file_i)
				SA_DialogEnd()
			}

			//context
			if SA_ButtonAlpha("").Icon("app:resources/context.png", 0.13).Show(3, 0, 1, 1).click {
				SA_DialogOpen("fileContext_"+file.Name, 1)
			}

			if SA_DialogStart("fileContext_" + file.Name) {
				SA_ColMax(0, 5)

				if SA_ButtonMenu(trns.RENAME).Show(0, 0, 1, 1).click {
					SA_DialogClose()
					SA_DialogOpen("RenameFile_"+file.Name, 1)
				}

				if SA_ButtonMenu(trns.DUPLICATE).Show(0, 1, 1, 1).click {
					SA_DialogClose()
					SA_DialogOpen("DuplicateFile_"+file.Name, 1)

					if strings.HasSuffix(file.Name, ".sqlite") {
						store.duplicateName = file.Name[:len(file.Name)-7] + "_2.sqlite"
					} else {
						store.duplicateName += "_2"
					}

				}

				if SA_ButtonMenu(trns.REMOVE).Show(0, 2, 1, 1).click {
					SA_DialogClose()
					SA_DialogOpen("RemoveFileConfirm_"+file.Name, 1)
				}

				SA_DialogEnd()
			}

			if SA_DialogStart("RenameFile_" + file.Name) {

				SA_ColMax(0, 7)

				newName := file.Name
				if SA_Editbox(&newName).Error(nil).Show(0, 0, 1, 1).finished { //check if file name exist ...
					if file.Name != newName && SA_InfoSet("rename_file", file.Name+"/"+newName) {
						file.Name = newName
					}
					SA_DialogClose()
				}

				SA_DialogEnd()
			}

			if SA_DialogStart("DuplicateFile_" + file.Name) {

				SA_ColMax(0, 7)

				SA_Editbox(&store.duplicateName).Error(nil).Show(0, 0, 1, 1)
				if SA_Button(trns.DUPLICATE).Enable(len(store.duplicateName) > 0).Show(0, 1, 1, 1).click { //check if file name exist ...
					if SA_InfoSet("duplicate_file", file.Name+"/"+store.duplicateName) {
						file.Name = store.duplicateName
					}
					SA_DialogClose()
				}

				SA_DialogEnd()
			}

			if SA_DialogStart("RemoveFileConfirm_" + file.Name) {
				if SA_DialogConfirm() {
					if store.SelectedFile == file_i {
						store.SelectedFile = -1
						store.SelectedApp = -1
					}
					SA_InfoSet("remove_file", file.Name)
				}
				SA_DialogEnd()
			}
		}
		SA_DivEnd()

		y++

		//apps
		if file.Expand {
			q := SA_SqlRead("dbs:"+file.Name, "SELECT rowid, label, app, sort FROM __skyalt__ ORDER BY sort")
			var app_rowid int
			var app_label string
			var app_name string
			var app_sort float64
			for q.Next(&app_rowid, &app_label, &app_name, &app_sort) {

				if app_label == "__default__" && app_name == "db" {
					continue
				}

				SA_DivStart(0, y, 1, 1)
				{
					SA_Col(0, 1)
					SA_ColMax(1, 100)

					isSelected := (file_i == store.SelectedFile && app_rowid == store.SelectedApp)
					if isSelected {
						SAPaint_Rect(0, 0, 1, 1, 0, SA_ThemeCd().Aprox(SA_ThemeWhite(), 0.8), 0)
					}

					//name
					SA_DivStart(1, 0, 1, 1)
					{
						SA_ColMax(0, 100)

						if SA_ButtonMenu(app_label).Pressed(isSelected).Tooltip("app: "+app_name).Show(0, 0, 1, 1).click {
							store.SelectedFile = file_i
							store.SelectedApp = app_rowid

							if SA_DivInfoPos("touchClicks", 0, 0) > 1 { //double click
								SA_DialogOpen("RenameApp_"+file.Name+"_"+strconv.Itoa(app_rowid), 1)
							}
						}

						id := (uint64(file_i) << uint64(32)) | uint64(app_rowid)
						SA_Div_SetDrag("app", id)
						src, pos, done := SA_Div_IsDrop("app", true, false, false)
						if done {
							src_file := store.Files[uint32(src>>32)]
							src_app_rowid := uint32(src)
							MoveApp(src_file, file, int(src_app_rowid), app_rowid, pos)
						}

						//logs
						/*{
							log := SA_Info("log_" + strconv.Itoa(app_i))
							if len(log) > 0 {
								app.logs = append(app.logs, Log{Text: log, Time: int64(SA_Time())})
								app.logs_showtime = SA_Time()
							}
							if app.logs_showtime+5 > SA_Time() {
								if SA_ButtonAlpha("").Icon("app:resources/warning.png", 0.0).Show(1, 0, 1, 1).click {
									SA_DialogOpen("log_"+strconv.Itoa(app_i), 0)
								}
							}
							if SA_DialogStart("log_" + strconv.Itoa(app_i)) {
								SA_ColMax(0, 20)
								SA_RowMax(1, 20)
								SA_TextCenter(trns.LOGS).Show(0, 0, 1, 1)

								SA_DivStart(0, 1, 1, 1)
								{
									SA_ColMax(0, 4)
									SA_ColMax(1, 100)
									for i, l := range app.logs {
										dt := time.Unix(l.Time, 0)
										SA_Text(dt.Format("2006-01-02 15:04:05")).Show(0, i, 1, 1)
										SA_Text(l.Text).Show(1, i, 1, 1)
									}
								}
								SA_DivEnd()

								SA_DialogEnd()
							}
						}*/
					}
					SA_DivEnd()

					//context
					if SA_ButtonAlpha("").Icon("app:resources/context.png", 0.13).Show(2, 0, 1, 1).click {
						SA_DialogOpen("appContext_"+file.Name+"_"+strconv.Itoa(app_rowid), 1)
					}

					if SA_DialogStart("appContext_" + file.Name + "_" + strconv.Itoa(app_rowid)) {
						SA_ColMax(0, 5)

						if SA_ButtonMenu(trns.RENAME).Show(0, 0, 1, 1).click {
							SA_DialogClose()
							SA_DialogOpen("RenameApp_"+file.Name+"_"+strconv.Itoa(app_rowid), 1)
							store.duplicateName = app_label
						}

						if SA_ButtonMenu(trns.DUPLICATE).Show(0, 1, 1, 1).click {
							SA_DialogClose()
							SA_DialogOpen("DuplicateApp_"+file.Name+"_"+strconv.Itoa(app_rowid), 1)
							store.duplicateName = app_label + "_2"
						}

						if SA_ButtonMenu(trns.REMOVE).Show(0, 2, 1, 1).click {
							SA_DialogClose()
							SA_DialogOpen("RemoveAppConfirm_"+file.Name+"_"+strconv.Itoa(app_rowid), 1)

						}
						SA_DialogEnd()
					}

					if SA_DialogStart("RenameApp_" + file.Name + "_" + strconv.Itoa(app_rowid)) {
						SA_ColMax(0, 7)
						if SA_Editbox(&store.duplicateName).Show(0, 0, 1, 1).finished {
							if len(store.duplicateName) > 0 {
								WriteApp(file, fmt.Sprintf("UPDATE __skyalt__ SET label='%s' WHERE rowid=%d", store.duplicateName, app_rowid), true)
							}
							SA_DialogClose()
						}
						SA_DialogEnd()
					}

					if SA_DialogStart("DuplicateApp_" + file.Name + "_" + strconv.Itoa(app_rowid)) {
						SA_ColMax(0, 7)

						SA_Editbox(&store.duplicateName).Error(nil).Show(0, 0, 1, 1)
						if SA_Button(trns.DUPLICATE).Enable(len(store.duplicateName) > 0).Show(0, 1, 1, 1).click { //check if file name exist ...

							if len(store.duplicateName) > 0 {
								q := SA_SqlRead("dbs:"+file.Name, fmt.Sprintf("SELECT storage FROM __skyalt__ WHERE rowid=%d;", app_rowid))
								var app_storage []byte
								q.Next(&app_storage)

								//add
								WriteApp(file, fmt.Sprintf("INSERT INTO __skyalt__(label, app, sort, storage) VALUES('%s','%s', %f, '%s');", store.duplicateName, app_name, app_sort+0.5, string(app_storage)), true)
								RefreshSort(file)
							}
							SA_DialogClose()
						}

						SA_DialogEnd()
					}

					if SA_DialogStart("RemoveAppConfirm_" + file.Name + "_" + strconv.Itoa(app_rowid)) {
						if SA_DialogConfirm() {
							if store.SelectedFile == file_i && store.SelectedApp == app_rowid {
								store.SelectedApp = -1
							}

							WriteApp(file, fmt.Sprintf("DELETE FROM __skyalt__ WHERE rowid=%d;", app_rowid), true)
							SA_DialogEnd() //!
							SA_DivEnd()    //!
							break
						}
						SA_DialogEnd()
					}

					y++

				}
				SA_DivEnd()
			}
		}
	}

	//new database
	SA_DivStart(0, y, 1, 1)
	{
		if SA_Button("+").Tooltip(trns.CREATE_DB).Show(0, 0, 1, 1).click {
			SA_DialogOpen("newFile", 1)
			store.createFile = "" //empty
		}
		if SA_DialogStart("newFile") {

			fnm := store.createFile
			if !strings.HasSuffix(fnm, ".sqlite") {
				fnm += ".sqlite"
			}

			SA_ColMax(0, 9)
			err := CheckFileName(store.createFile, FindFile(fnm) != nil)

			SA_Editbox(&store.createFile).Error(err).TempToValue(true).ShowDescription(0, 0, 1, 1, trns.NAME, 2, nil)

			if SA_Button(trns.CREATE_FILE).Enable(err == nil).Show(0, 1, 1, 1).click {
				SA_InfoSet("new_file", fnm)
				SA_DialogClose()
			}

			SA_DialogEnd()
		}
	}
	SA_DivEnd()
}

func CheckFileName(name string, alreadyExist bool) error {

	empty := len(name) == 0

	name = strings.ToLower(name)

	var err error
	if alreadyExist {
		err = errors.New(trns.ALREADY_EXISTS)
	} else if empty {
		err = errors.New(trns.EMPTY_FIELD)
	}

	return err
}

func Render() {
	SA_Col(0, 4.5) //min
	SA_ColResize(0, 7)
	SA_ColMax(1, 100)
	SA_RowMax(1, 100)

	SA_DivStart(0, 0, 1, 1)
	{
		SA_Col(0, 2)
		SA_ColMax(1, 100)

		//Menu + dialogs
		if SA_ButtonStyle("", &styles.ButtonLogo).Icon("app:resources/logo_small.png", 0).Show(0, 0, 1, 1).click {
			SA_DialogOpen("Menu", 1)
		}
		if SA_DialogStart("Menu") {
			Menu()
			SA_DialogEnd()
		}

		if SA_DialogStart("Settings") {
			Settings()
			SA_DialogEnd()
		}
		if SA_DialogStart("About") {
			About()
			SA_DialogEnd()
		}

		//Search
		SA_Editbox(&store.SearchFiles).TempToValue(true).Ghost(trns.SEARCH).Highlight(len(store.SearchFiles) > 0, &styles.EditboxYellow).Show(1, 0, 1, 1)

	}
	SA_DivEnd()

	SA_DivStart(0, 1, 1, 1)
	Files()
	SA_DivEnd()

	file := FindSelectedFile()
	//app := FindSelectedApp()	//fix if not exist ...
	if store.SelectedApp > 0 {
		SA_DivStartName(1, 0, 1, 2, fmt.Sprintf("%d_%d", file.id, store.SelectedApp))
		SA_RenderApp("dbs:"+file.Name, store.SelectedApp)
		SA_DivEnd()
	} else if file != nil {
		app_rowid := FindOrAddDefaultApp(file)
		SA_DivStartName(1, 0, 1, 2, fmt.Sprintf("%d_%d", file.id, app_rowid))
		SA_RenderApp("dbs:"+file.Name, app_rowid)
		SA_DivEnd()
	}
}

var styles SA_Styles
var g_ButtonAddApp _SA_Style

func Init() {
	store.SelectedFile = -1
	store.SelectedApp = -1
	store.last_file_id = 1

	//default
	json.Unmarshal(SA_File("storage_json"), &store)
	json.Unmarshal(SA_File("translations_json:app:resources/translations.json"), &trns)
	json.Unmarshal(SA_File("styles_json"), &styles)

	//styles
	g_ButtonAddApp = styles.ButtonBorder
	g_ButtonAddApp.Margin(0.17)
}
func Save() []byte {
	js, _ := json.MarshalIndent(&store, "", "")
	return js
}
