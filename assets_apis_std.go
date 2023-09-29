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
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (asset *Asset) print(str string) {
	fmt.Println(str)
}
func (asset *Asset) _sa_print(strMem uint64) {

	str, err := asset.ptrToString(strMem)
	if asset.AddLogErr(err) {
		return
	}

	asset.print(str)
}
func (asset *Asset) _sa_print_float(val float64) {
	fmt.Println(val)
}

func (asset *Asset) info_float(key string) float64 {
	switch strings.ToLower(key) {
	case "theme":
		return float64(asset.app.root.ui.io.ini.Theme)

	case "date":
		return float64(asset.app.root.ui.io.ini.Date)

	case "time_zone":
		_, o := time.Now().Zone()
		return float64(o) / 3600

	case "time_utc":
		return float64(time.Now().UnixMicro()) / 1000000 //seconds

	case "time":
		tm := time.Now()
		_, zone_sec := tm.Zone()
		return (float64(tm.UnixMicro()) / 1000000) + float64(zone_sec) //seconds

	case "dpi":
		return float64(asset.app.root.ui.io.ini.Dpi)

	case "dpi_default":
		return float64(asset.app.root.ui.io.ini.Dpi_default)

	case "fullscreen":
		return OsTrnFloat(asset.app.root.ui.io.ini.Fullscreen, 1, 0)

	case "stats":
		return OsTrnFloat(asset.app.root.ui.io.ini.Stats, 1, 0)

	case "grid":
		return OsTrnFloat(asset.app.root.ui.io.ini.Grid, 1, 0)

	case "sts_uid":
		return float64(asset.app.root.settings.AddSts_uid())

	default:
		fmt.Println("info_float(): Unknown key: ", key)
	}

	return -1
}

func (asset *Asset) info_setFloat(key string, v float64) int64 {
	switch strings.ToLower(key) {
	case "theme":
		asset.app.root.ui.io.ini.Theme = int(v)
		asset.app.root.ReloadStyles()
		return 1
	case "date":
		asset.app.root.ui.io.ini.Date = int(v)
		return 1

	case "dpi":
		asset.app.root.ui.io.ini.Dpi = int(v)
		return 1
	case "fullscreen":
		asset.app.root.ui.io.ini.Fullscreen = (v > 0)
		return 1

	case "stats":
		asset.app.root.ui.io.ini.Stats = (v > 0)
		return 1

	case "grid":
		asset.app.root.ui.io.ini.Grid = (v > 0)
		return 1

	case "nosleep":
		asset.app.root.ui.SetNoSleep()
		return 1

	case "save":
		if v > 0 {
			asset.app.root.save = true //call app.SaveData() after tick
			return 1
		}
		return 0

	case "exit":
		if v > 0 {
			asset.app.root.exit = true
			return 1
		}
		return 0

	default:
		fmt.Println("info_setFloat(): Unknown key: ", key)

	}

	return -1
}

func (asset *Asset) _sa_info_float(keyMem uint64) float64 {

	key, err := asset.ptrToString(keyMem)
	if asset.AddLogErr(err) {
		return -1
	}

	return asset.info_float(key)
}

func (asset *Asset) _sa_info_setFloat(keyMem uint64, v float64) int64 {

	key, err := asset.ptrToString(keyMem)
	if asset.AddLogErr(err) {
		return -1
	}

	return asset.info_setFloat(key, v)
}

func (asset *Asset) info_string(key string) (string, int64) {

	logSts_Id, found := strings.CutPrefix(key, "log_")
	if found {
		sts_id, err := strconv.Atoi(logSts_Id)
		if err == nil {
			app := asset.app.root.FindAppId(sts_id)
			if app != nil {
				return app.GetLog(), 1
			}
		}
		return "", 1
	}

	switch strings.ToLower(key) {
	case "files":
		return asset.app.root.dbsList, 1

	case "apps":
		return asset.app.root.appsList, 1

	case "languages":
		lngs := ""
		for _, lng := range asset.app.root.ui.io.ini.Languages {
			lngs += lng + "/"
		}
		return strings.TrimSuffix(lngs, "/"), 1

	default:
		fmt.Println("info_string(): Unknown key: ", key)

	}
	return "", -1
}
func (asset *Asset) info_string_len(key string) int64 {

	dst, ret := asset.info_string(key)
	if ret > 0 {
		return int64(len(dst))
	}
	return -1
}

func (asset *Asset) _sa_info_string(keyMem uint64, dstMem uint64) int64 {

	key, err := asset.ptrToString(keyMem)
	if asset.AddLogErr(err) {
		return -1
	}

	dst, ret := asset.info_string(key)
	err = asset.stringToPtr(dst, dstMem)
	asset.AddLogErr(err)
	return ret
}

func (asset *Asset) _sa_info_string_len(keyMem uint64) int64 {

	key, err := asset.ptrToString(keyMem)
	if asset.AddLogErr(err) {
		return -1
	}

	return asset.info_string_len(key)
}

func (asset *Asset) info_setString(key string, value string) int64 {
	switch strings.ToLower(key) {
	case "languages":
		if len(value) > 0 {
			asset.app.root.ui.io.ini.Languages = strings.Split(value, "/")
		} else {
			asset.app.root.ui.io.ini.Languages = nil
		}
		asset.app.root.ReloadTranslations()
		return 1

	case "new_file":
		if asset.app.root.CreateDb(value) {
			return 1
		}
		return -1

	case "rename_file":
		d := strings.IndexByte(value, '/')
		if d > 0 && d < len(value)-1 {
			if asset.app.root.RenameDb(value[:d], value[d+1:]) {
				return 1
			}
		}
		return -1

	case "duplicate_file":
		d := strings.IndexByte(value, '/')
		if d > 0 && d < len(value)-1 {
			if asset.app.root.DuplicateDb(value[:d], value[d+1:]) {
				return 1
			}
		}
		return -1

	case "remove_file":
		if asset.app.root.RemoveDb(value) {
			return 1
		}
		return -1

	case "duplicate_setting":
		srcid, err := strconv.Atoi(value)
		if err != nil {
			asset.AddLogErr(err)
			return -1
		}

		dstid, err := asset.app.root.settings.Duplicate(srcid)
		if err != nil {
			asset.AddLogErr(err)
			return -1
		}
		return int64(dstid)

	default:
		fmt.Println("info_setString(): Unknown key: ", key)
	}

	return -1
}

func (asset *Asset) _sa_info_setString(keyMem uint64, valueMem uint64) int64 {

	key, err := asset.ptrToString(keyMem)
	if asset.AddLogErr(err) {
		return -1
	}
	value, err := asset.ptrToString(valueMem)
	if asset.AddLogErr(err) {
		return -1
	}

	return asset.info_setString(key, value)
}

func (asset *Asset) findAsset(assetName string) *Asset {
	if len(assetName) > 0 {
		return asset.app.FindAsset(assetName)
	}
	return asset
}

func (asset *Asset) _getResource(url string) ([]byte, error) {

	res, err := MediaParseUrl(url, asset)
	if err != nil {
		return nil, fmt.Errorf("MediaParseUrl() failed: %w", err)
	}

	data, err := res.GetBlob()
	if err != nil {
		return nil, fmt.Errorf("GetBlob() failed: %w", err)
	}

	return data, nil
}

func (asset *Asset) resource(path string) ([]byte, int64, error) {

	data, err := asset._getResource(path)
	if err != nil {
		return nil, -1, err
	}
	return data, 1, nil
}

func (asset *Asset) resource_len(path string) (int64, error) {

	data, err := asset._getResource(path)
	if err != nil {
		return -1, err
	}
	return int64(len(data)), nil
}

func (asset *Asset) _sa_blob(pathMem uint64, dstMem uint64) int64 {

	path, err := asset.ptrToString(pathMem)
	if asset.AddLogErr(err) {
		return -1
	}

	data, err := asset._getResource(path)
	asset.AddLogErr(err)
	if err != nil {
		return -1
	}

	err = asset.bytesToPtr(data, dstMem)
	asset.AddLogErr(err)
	if err != nil {
		return -1
	}

	return 1
}

func (asset *Asset) _sa_blob_len(pathMem uint64) int64 {

	path, err := asset.ptrToString(pathMem)
	if asset.AddLogErr(err) {
		return -1
	}

	ret, err := asset.resource_len(path)
	asset.AddLogErr(err)
	return ret
}
