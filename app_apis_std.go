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
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (app *App) hasRootPriviliges(cmd string) bool {
	if app.name != "base" {
		app.AddLogErr(fmt.Errorf("command '%s' denied, low privilages", cmd))
		return false
	}
	return true
}

func (app *App) _info_get_prepare(cmd string, prm1 string, prm2 string, onlyLen bool) (string, int64) {
	switch strings.ToLower(cmd) {
	case "theme":
		return strconv.Itoa(app.db.root.ui.io.ini.Theme), 1

	case "date":
		return strconv.Itoa(app.db.root.ui.io.ini.Date), 1

	case "timezone":
		return strconv.Itoa(app.db.root.ui.io.ini.TimeZone), 1

	case "time_utc":
		return strconv.FormatFloat(float64(time.Now().UnixMicro())/1000000, 'f', -1, 64), 1 //seconds

	case "time":
		tm := time.Now()
		return strconv.FormatFloat(float64(tm.UnixMicro())/1000000+float64(app.db.root.ui.io.ini.TimeZone*3600), 'f', -1, 64), 1 //seconds
	}

	if !app.hasRootPriviliges(cmd) {
		return "", 0
	}

	switch strings.ToLower(cmd) {
	case "dpi":
		return strconv.Itoa(app.db.root.ui.io.ini.Dpi), 1

	case "dpi_default":
		return strconv.Itoa(app.db.root.ui.io.ini.Dpi_default), 1

	case "fullscreen":
		return strconv.Itoa(OsTrn(app.db.root.ui.io.ini.Fullscreen, 1, 0)), 1

	case "stats":
		return strconv.Itoa(OsTrn(app.db.root.ui.io.ini.Stats, 1, 0)), 1

	case "grid":
		return strconv.Itoa(OsTrn(app.db.root.ui.io.ini.Grid, 1, 0)), 1

	case "file_size":
		db := app.db.root.FindDb(app.db.root.folderDatabases + "/" + prm1)
		if db != nil {
			return strconv.Itoa(int(db.bytes)), 1
		}

	case "log":
		app2 := app.GetAppFromValue(prm1, prm2)
		if app2 != nil {
			return app2.GetLog(!onlyLen), 1
		}
		return "", 0

	case "files":
		return app.db.root.dbsList, 1

	case "apps":
		list := app.db.root.GetAppsList()
		js, err := json.MarshalIndent(&list, "", "")
		if err == nil {
			return string(js), 1
		}

	case "languages":
		lngs := ""
		for _, lng := range app.db.root.ui.io.ini.Languages {
			lngs += lng + "/"
		}
		return strings.TrimSuffix(lngs, "/"), 1

	default:
		fmt.Println("info_string(): Unknown key: ", cmd)

	}

	return "", -1
}

func (app *App) info_set(cmd string, prm1 string, prm2 string, prm3 string) int64 {

	switch strings.ToLower(cmd) {
	case "nosleep":
		app.db.root.ui.SetNoSleep()
		return 1
	}

	if !app.hasRootPriviliges(cmd) {
		return -1
	}

	switch strings.ToLower(cmd) {
	case "theme":
		v, err := strconv.Atoi(prm1)
		if err == nil {
			app.db.root.ui.io.ini.Theme = v
			app.db.root.ReopenApps()
			return 1
		}
		fmt.Printf("Atoi() failed: %v\n", err)
	case "date":
		v, err := strconv.Atoi(prm1)
		if err == nil {
			app.db.root.ui.io.ini.Date = v
			return 1
		}
		fmt.Printf("Atoi() failed: %v\n", err)
	case "timezone":
		v, err := strconv.Atoi(prm1)
		if err == nil {
			app.db.root.ui.io.ini.TimeZone = v
			return 1
		}
		fmt.Printf("Atoi() failed: %v\n", err)
	case "dpi":
		v, err := strconv.Atoi(prm1)
		if err == nil {
			app.db.root.ui.io.ini.Dpi = v
			return 1
		}
		fmt.Printf("Atoi() failed: %v\n", err)
	case "fullscreen":
		v, err := strconv.Atoi(prm1)
		if err == nil {
			app.db.root.ui.io.ini.Fullscreen = (v > 0)
			return 1
		}
		fmt.Printf("Atoi() failed: %v\n", err)
	case "stats":
		v, err := strconv.Atoi(prm1)
		if err == nil {
			app.db.root.ui.io.ini.Stats = (v > 0)
			return 1
		}
		fmt.Printf("Atoi() failed: %v\n", err)
	case "grid":
		v, err := strconv.Atoi(prm1)
		if err == nil {
			app.db.root.ui.io.ini.Grid = (v > 0)
			return 1
		}
		fmt.Printf("Atoi() failed: %v\n", err)
	case "nosleep":
		app.db.root.ui.SetNoSleep()
		return 1

	case "vacuum":
		app.db.root.Vacuum()
		return 1

	case "save":
		app.db.root.save = true //call app.SaveData() after tick
		return 1

	case "exit":
		app.db.root.exit = true
		return 1

	case "languages":
		if len(prm1) > 0 {
			app.db.root.ui.io.ini.Languages = strings.Split(prm1, "/")
		} else {
			app.db.root.ui.io.ini.Languages = nil
		}
		app.db.root.ReopenApps()
		return 1

	case "new_file":
		if app.db.root.CreateDb(prm1) {
			return 1
		}
		return -1

	case "rename_file":
		if app.db.root.RenameDb(prm1, prm2) {
			return 1
		}

		return -1

	case "duplicate_file":
		if app.db.root.DuplicateDb(prm1, prm2) {
			return 1
		}
		return -1

	case "remove_file":
		if app.db.root.RemoveDb(app.db.root.folderDatabases + "/" + prm1) {
			return 1
		}
		return -1

	case "vacuum_file":
		db := app.db.root.FindDb(app.db.root.folderDatabases + "/" + prm1)
		if db != nil {
			db.Vacuum()
			return 1
		}
		return -1

	case "new_app":
		err := app.db.root.CreateApp(prm1, prm2)
		if !app.AddLogErr(err) {
			return 1
		}
		return -1

	case "package_app":
		err := app.db.root.PackageApp(prm1)
		if !app.AddLogErr(err) {
			return 1
		}
		return -1

	case "extract_app":
		err := app.db.root.ExtractApp(prm1)
		if !app.AddLogErr(err) {
			return 1
		}
		return -1

	case "save_app":
		app2 := app.GetAppFromValue(prm1, prm2)
		if app2 != nil {
			app2.Save()
		}

	case "setup_db":
		app2 := app.GetAppFromValue(prm1, prm2)
		if app2 != nil {
			app2.SetupDB()
		}

	default:
		fmt.Println("info_setFloat(): Unknown key: ", cmd)
	}

	return -1
}

func (app *App) info_get_prepare(cmd string, prm1 string, prm2 string) int64 {

	var ret int64
	app.info_string, ret = app._info_get_prepare(cmd, prm1, prm2, true)
	if ret > 0 {
		return int64(len(app.info_string))
	}
	return -1
}

func (app *App) _sa_info_get_prepare(keyMem uint64, prm1Mem uint64, prm2Mem uint64) int64 {
	key, err := app.ptrToString(keyMem)
	if app.AddLogErr(err) {
		return -1
	}
	prm1, err := app.ptrToString(prm1Mem)
	if app.AddLogErr(err) {
		return -1
	}
	prm2, err := app.ptrToString(prm2Mem)
	if app.AddLogErr(err) {
		return -1
	}

	return app.info_get_prepare(key, prm1, prm2)
}

func (app *App) _sa_info_get(dstMem uint64) int64 {
	err := app.stringToPtr(app.info_string, dstMem)
	if app.AddLogErr(err) {
		return -1
	}
	return 1
}

func (app *App) _sa_info_set(keyMem uint64, prm1Mem uint64, prm2Mem uint64, prm3Mem uint64) int64 {

	key, err := app.ptrToString(keyMem)
	if app.AddLogErr(err) {
		return -1
	}
	prm1, err := app.ptrToString(prm1Mem)
	if app.AddLogErr(err) {
		return -1
	}
	prm2, err := app.ptrToString(prm2Mem)
	if app.AddLogErr(err) {
		return -1
	}
	prm3, err := app.ptrToString(prm3Mem)
	if app.AddLogErr(err) {
		return -1
	}

	return app.info_set(key, prm1, prm2, prm3)
}

// prm1 = <file.sqlite>
// prm2 = <approw_id>
func (app *App) GetAppFromValue(prm1 string, prm2 string) *App {
	db := app.db.root.FindDb(app.db.root.folderDatabases + "/" + prm1)
	if db != nil {
		app_rowid, err := strconv.Atoi(prm2)
		if err == nil {
			return db.FindApp(app_rowid)
		} else {
			fmt.Printf("Atoi() failed: %v\n", err)
		}
	}
	return nil
}

func (app *App) _getResource(url string) ([]byte, error) {

	if strings.EqualFold(url, "storage_json") {
		js, err := app.GetStorage()
		if err != nil {
			return nil, fmt.Errorf("GetStorage() failed: %w", err)
		}
		return js, nil
	}

	if strings.EqualFold(url, "styles_json") {
		return app.db.root.stylesJs, nil
	}

	var isTrns bool
	url, isTrns = strings.CutPrefix(url, "translations_json:")
	//protobuff, csv, etc.? ...

	res, err := MediaParseUrl(url, app)
	if err != nil {
		return nil, fmt.Errorf("MediaParseUrl() failed: %w", err)
	}

	data, _, err := res.GetBlob()
	if err != nil {
		return nil, fmt.Errorf("GetBlob() failed: %w", err)
	}

	if isTrns {
		data, err = TranslateJson(data, app.db.root.ui.io.ini.Languages)
		if err != nil {
			return nil, fmt.Errorf("GetBlob() failed: %w", err)
		}
		return data, nil
	}

	return data, nil
}

func (app *App) resource(path string) ([]byte, int64, error) {

	data, err := app._getResource(path)
	if err != nil {
		return nil, -1, err
	}
	return data, 1, nil
}

func (app *App) resource_len(path string) (int64, error) {

	data, err := app._getResource(path)
	if err != nil {
		return -1, err
	}
	return int64(len(data)), nil
}

func (app *App) _sa_blob(pathMem uint64, dstMem uint64) int64 {

	path, err := app.ptrToString(pathMem)
	if app.AddLogErr(err) {
		return -1
	}

	data, err := app._getResource(path)
	app.AddLogErr(err)
	if err != nil {
		return -1
	}

	err = app.bytesToPtr(data, dstMem)
	app.AddLogErr(err)
	if err != nil {
		return -1
	}

	return 1
}

func (app *App) _sa_blob_len(pathMem uint64) int64 {

	path, err := app.ptrToString(pathMem)
	if app.AddLogErr(err) {
		return -1
	}

	ret, err := app.resource_len(path)
	app.AddLogErr(err)
	return ret
}

func (app *App) print(str string) {
	fmt.Println(str)
}
func (app *App) _sa_print(strMem uint64) {

	str, err := app.ptrToString(strMem)
	if app.AddLogErr(err) {
		return
	}

	app.print(str)
}
func (app *App) _sa_print_float(val float64) {
	fmt.Println(val)
}
