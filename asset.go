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
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Asset struct {
	app  *App
	name string

	resourceFiles []string
	translations  *Translations

	wasm  *AssetWasm
	debug *AssetDebug

	sts_rowid int

	styles *DivStyles
}

func (asset *Asset) AddLogErr(err error) bool {
	if err != nil {
		asset.app.AddLog(asset.name + ": " + err.Error())
		return true
	}
	return false
}

func NewAsset(app *App, name string) (*Asset, error) {
	var asset Asset
	asset.app = app
	asset.name = name

	var err error
	asset.sts_rowid, err = app.root.settings.FindOrAdd(asset.app.sts_id, asset.name)
	if err != nil {
		return nil, err
	}

	asset.wasm, err = NewAssetWasm(&asset)
	if err != nil {
		return nil, err
	}

	asset.translations = NewTranslations(asset.getTranslationsPath())

	asset.UpdateResources()

	asset.styles = NewDivStyles(&asset)

	asset.Tick()

	return &asset, nil
}

func (asset *Asset) Destroy() {

	asset.SaveData()

	if asset.wasm != nil {
		asset.wasm.Destroy()
	}
	if asset.debug != nil {
		asset.debug.Destroy()
	}
}

func (asset *Asset) IsReadyToFire() bool {
	if asset.debug != nil {
		return asset.debug.conn != nil
	}
	if asset.wasm != nil {
		return asset.wasm.mod != nil
	}
	return false
}

func (asset *Asset) Call(fnName string, args []byte) (int64, error) {
	var ret int64
	var err error

	if asset.debug != nil {
		ret, err = asset.debug.Call(fnName, args, asset)
	} else if asset.wasm != nil {
		ret, err = asset.wasm.Call(fnName, args)
	} else {
		err = errors.New("no call")
	}

	return ret, err
}

func (asset *Asset) CallSet(js []byte, fnName string) {
	var data []byte
	data = append(data, TpBytes)
	data = binary.LittleEndian.AppendUint64(data, uint64(len(js)))
	data = append(data, js...)

	asset.Call(fnName, data)
}

func (asset *Asset) SaveData() {
	if asset.debug != nil {
		asset.debug.SaveData(asset)
	} else if asset.wasm != nil {
		asset.wasm.SaveData()
	}
}

func (asset *Asset) getPath() string {
	return asset.app.getPath() + "/" + asset.name
}

func (asset *Asset) getResourcesPath() string {
	return asset.getPath() + "/resources"
}

func (asset *Asset) getTranslationsPath() string {
	return asset.getResourcesPath() + "/translations.json"
}

func (asset *Asset) getWasmPath() string {
	return asset.getPath() + "/main.wasm"
}

func (asset *Asset) UpdateResources() {

	asset.resourceFiles = nil

	if !OsFolderExists(asset.getResourcesPath()) {
		return
	}

	dir, err := os.ReadDir(asset.getResourcesPath())
	if err != nil {
		fmt.Printf("ReadDir() failed: %v\n", err)
		return
	}

	for _, file := range dir {
		if !file.IsDir() {
			ext := filepath.Ext(file.Name())
			if !strings.EqualFold(ext, ".go") && !strings.EqualFold(ext, ".wasm") && !strings.EqualFold(ext, "translations.json") {

				asset.resourceFiles = append(asset.resourceFiles, file.Name())
			}
		}
	}

	sort.Strings(asset.resourceFiles)
}

func (asset *Asset) loadData() {
	jsStore, err := asset.app.root.settings.GetContent(asset.sts_rowid)
	if asset.AddLogErr(err) {
		return
	}

	asset.CallSet(jsStore, "_sa_init")
}

func (asset *Asset) Tick() {
	loadData := false
	loadTranslations := asset.translations.Maintenance()

	loadStyles := asset.styles.theme != asset.app.root.ui.io.ini.Theme

	assetDebug := asset.app.root.server.Find(asset.app.sts_id, asset.name)
	if assetDebug != nil && asset.debug != assetDebug {
		if asset.debug != nil {
			asset.debug.Destroy()
		}
		asset.debug = assetDebug
		loadData = true
		loadTranslations = true
		loadStyles = true
		//asset.translations.file_tm = 0 //reload transactions
	}

	//wasm
	if asset.debug != nil {
		//connection lost, go back to wasm
		if asset.debug.conn == nil {
			asset.debug.Destroy()
			asset.debug = nil

			loadStyles = true
			loadTranslations = true
			loadData = true
		}
	} else if asset.wasm != nil {
		changed, err := asset.wasm.Tick()
		if err != nil {
			asset.AddLogErr(err)
			asset.wasm = nil

		} else if changed {
			loadStyles = true
			loadTranslations = true
			loadData = true
		}
	}

	//data(json)
	if loadData {
		asset.loadData()
	}

	if loadTranslations {
		js, err := asset.translations.Read(asset.app.root.ui.io.ini.Languages)
		if err == nil {
			asset.CallSet(js, "_sa_translations")
		}
	}

	if loadStyles {
		asset.CallSet(asset.app.root.stylesJs, "_sa_styles")
		asset.styles.theme = asset.app.root.ui.io.ini.Theme
	}
	//resources
	asset.UpdateResources()
}

func (asset *Asset) SetDebugLine(line string) {

	st := asset.app.root.levels.GetStack()
	if st.stack == nil || st.stack.crop.IsZero() {
		return
	}

	if st.stack.enableInput {
		if st.stack.crop.Inside(asset.app.root.ui.io.touch.pos) {
			asset.app.root.SetDebugLine(line)
		}
	}
}
