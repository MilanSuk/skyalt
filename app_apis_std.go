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
	"time"
)

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

func (app *App) info_float(key string) float64 {
	switch strings.ToLower(key) {
	case "theme":
		return float64(app.db.root.ui.io.ini.Theme)

	case "date":
		return float64(app.db.root.ui.io.ini.Date)

	case "timezone":
		return float64(app.db.root.ui.io.ini.TimeZone)

	case "time_utc":
		return float64(time.Now().UnixMicro()) / 1000000 //seconds

	case "time":
		tm := time.Now()
		return (float64(tm.UnixMicro()) / 1000000) + float64(app.db.root.ui.io.ini.TimeZone*3600) //seconds

	case "dpi":
		return float64(app.db.root.ui.io.ini.Dpi)

	case "dpi_default":
		return float64(app.db.root.ui.io.ini.Dpi_default)

	case "fullscreen":
		return OsTrnFloat(app.db.root.ui.io.ini.Fullscreen, 1, 0)

	case "stats":
		return OsTrnFloat(app.db.root.ui.io.ini.Stats, 1, 0)

	case "grid":
		return OsTrnFloat(app.db.root.ui.io.ini.Grid, 1, 0)

	default:
		fmt.Println("info_float(): Unknown key: ", key)
	}

	return -1
}

func (app *App) info_setFloat(key string, v float64) int64 {
	switch strings.ToLower(key) {
	case "theme":
		app.db.root.ui.io.ini.Theme = int(v)
		app.db.root.ReloadApps()
		return 1
	case "date":
		app.db.root.ui.io.ini.Date = int(v)
		return 1

	case "timezone":
		app.db.root.ui.io.ini.TimeZone = int(v)
		return 1

	case "dpi":
		app.db.root.ui.io.ini.Dpi = int(v)
		return 1
	case "fullscreen":
		app.db.root.ui.io.ini.Fullscreen = (v > 0)
		return 1

	case "stats":
		app.db.root.ui.io.ini.Stats = (v > 0)
		return 1

	case "grid":
		app.db.root.ui.io.ini.Grid = (v > 0)
		return 1

	case "nosleep":
		app.db.root.ui.SetNoSleep()
		return 1

	case "save":
		if v > 0 {
			app.db.root.save = true //call app.SaveData() after tick
			return 1
		}
		return 0

	case "exit":
		if v > 0 {
			app.db.root.exit = true
			return 1
		}
		return 0

	default:
		fmt.Println("info_setFloat(): Unknown key: ", key)

	}

	return -1
}

func (app *App) _sa_info_float(keyMem uint64) float64 {

	key, err := app.ptrToString(keyMem)
	if app.AddLogErr(err) {
		return -1
	}

	return app.info_float(key)
}

func (app *App) _sa_info_setFloat(keyMem uint64, v float64) int64 {

	key, err := app.ptrToString(keyMem)
	if app.AddLogErr(err) {
		return -1
	}

	return app.info_setFloat(key, v)
}

func (app *App) info_string(key string, onlyLen bool) (string, int64) {

	if app.name != "base" {
		app.AddLogErr(errors.New("access denied, low privilages"))
		return "", 0
	}

	//log
	log, found := strings.CutPrefix(key, "log_") //log_<file.sqlite>/<approw_id>
	if found {
		d := strings.IndexByte(log, '/')
		if d >= 0 {
			db := app.db.root.FindDb(app.db.root.folderDatabases + "/" + log[:d])
			if db != nil {
				app_rowid, err := strconv.Atoi(log[d+1:])
				if err == nil {
					app2 := db.FindApp(app_rowid)
					if app2 != nil {
						return app2.GetLog(!onlyLen), 1
					}
				}
			}
		}
		return "", 0
	}

	switch strings.ToLower(key) {
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
		fmt.Println("info_string(): Unknown key: ", key)

	}
	return "", -1
}
func (app *App) info_string_len(key string) int64 {

	dst, ret := app.info_string(key, true)
	if ret > 0 {
		return int64(len(dst))
	}
	return -1
}

func (app *App) _sa_info_string(keyMem uint64, dstMem uint64) int64 {

	key, err := app.ptrToString(keyMem)
	if app.AddLogErr(err) {
		return -1
	}

	dst, ret := app.info_string(key, false)
	err = app.stringToPtr(dst, dstMem)
	app.AddLogErr(err)
	return ret
}

func (app *App) _sa_info_string_len(keyMem uint64) int64 {

	key, err := app.ptrToString(keyMem)
	if app.AddLogErr(err) {
		return -1
	}

	return app.info_string_len(key)
}

func (app *App) info_setString(key string, value string) int64 {
	switch strings.ToLower(key) {
	case "languages":
		if len(value) > 0 {
			app.db.root.ui.io.ini.Languages = strings.Split(value, "/")
		} else {
			app.db.root.ui.io.ini.Languages = nil
		}
		app.db.root.ReloadApps()
		return 1

	case "new_file":
		if app.db.root.CreateDb(value) {
			return 1
		}
		return -1

	case "rename_file":
		d := strings.IndexByte(value, '/')
		if d > 0 && d < len(value)-1 {
			if app.db.root.RenameDb(value[:d], value[d+1:]) {
				return 1
			}
		}
		return -1

	case "duplicate_file":
		d := strings.IndexByte(value, '/')
		if d > 0 && d < len(value)-1 {
			if app.db.root.DuplicateDb(value[:d], value[d+1:]) {
				return 1
			}
		}
		return -1

	case "remove_file":
		if app.db.root.RemoveDb(value) {
			return 1
		}
		return -1

	case "new_repo":
		d := strings.IndexByte(value, '/')
		if d > 0 && d < len(value)-1 {
			name := value[:d]
			lang := value[d+1:]
			err := app.db.root.CreateApp(name, lang)
			if !app.AddLogErr(err) {
				return 1
			}
		}
		return -1

	case "package_repo":
		err := app.db.root.PackageApp(value)
		if !app.AddLogErr(err) {
			return 1
		}
		return -1

	case "extract_repo":
		err := app.db.root.ExtractApp(value)
		if !app.AddLogErr(err) {
			return 1
		}
		return -1

	default:
		fmt.Println("info_setString(): Unknown key: ", key)
	}

	return -1
}

func (app *App) _sa_info_setString(keyMem uint64, valueMem uint64) int64 {

	key, err := app.ptrToString(keyMem)
	if app.AddLogErr(err) {
		return -1
	}
	value, err := app.ptrToString(valueMem)
	if app.AddLogErr(err) {
		return -1
	}

	return app.info_setString(key, value)
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
