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
	"os"
	"strings"
)

type App struct {
	root   *Root
	name   string
	assets []*Asset

	baseAsset *Asset
	db_name   string

	sts_id int

	fn2Return  []byte
	fn2Returns []byte

	logs []string
}

func NewApp(root *Root, name string, db_name string, sts_id int) (*App, error) {
	var app App
	app.root = root
	app.name = name
	app.db_name = db_name
	app.sts_id = sts_id

	//load assets
	dir, err := os.ReadDir(app.getPath())
	if err != nil {
		return nil, fmt.Errorf("ReadDir(%s) failed: %w", app.getPath(), err)
	}
	for _, fld := range dir {
		if fld.IsDir() {
			asset, err := app.AddAsset(fld.Name())
			if err != nil {
				return nil, err
			}
			if strings.EqualFold(fld.Name(), "main") {
				app.baseAsset = asset
			}
		}
	}

	return &app, nil
}
func (app *App) Destroy() {
	for _, asset := range app.assets {
		asset.Destroy()
	}
}

func (app *App) SaveData() {
	for _, asset := range app.assets {
		asset.SaveData()
	}
}

func (app *App) AddLog(err string) {
	//print
	fmt.Printf("Error(%s): %s\n", app.getPath(), err)

	//add
	app.logs = append(app.logs, err)
}
func (app *App) GetLog() string {
	var ret string
	if len(app.logs) > 0 {
		ret = app.logs[0]
		app.logs = app.logs[1:] //cut
	}
	return ret
}

func (app *App) FindAsset(name string) *Asset {
	for _, asset := range app.assets {
		if asset.name == name {
			return asset
		}
	}
	return nil
}
func (app *App) AddAsset(name string) (*Asset, error) {
	//find
	asset := app.FindAsset(name)
	if asset != nil {
		return asset, nil //ok
	}

	//add
	var err error
	asset, err = NewAsset(app, name)
	if err != nil {
		return nil, err
	}

	app.assets = append(app.assets, asset)
	return asset, nil
}

func (app *App) getPath() string {
	return app.root.folderApps + "/" + app.name
}

func (app *App) Tick() {
	for _, asset := range app.assets {
		asset.Tick()
	}
}

func (app *App) IsReadyToFire() bool {
	for _, asset := range app.assets {
		if !asset.IsReadyToFire() {
			return false
		}
	}
	return true
}

func (app *App) ReloadTranslations() {
	for _, asset := range app.assets {
		asset.translations.file_tm = 0
	}
	app.root.last_ticks = 0
}

func (app *App) Render(startIt bool) {

	if startIt {
		app.baseAsset.renderStart()
	}
	if app.IsReadyToFire() {
		_, err := app.baseAsset.Call("_sa_render", nil)
		if err != nil {
			fmt.Print(err)
		}
	} else {
		app.baseAsset.paint_text(0, 0, 1, 1, nil, "Error: 'Main.wasm' is missing or corrupted", "", false, false, true)
	}

	if app.baseAsset.debug != nil {
		//draw blue rectangle, when debug mode is active
		blue := OsCd{50, 50, 255, 180}

		style := app.root.styles.Text
		style.Main.Font_alignH = 0
		style.Main.Color = blue

		app.baseAsset.paint_rect(0, 0, 1, 1, 0.06, blue, 0.03)
		app.baseAsset.paint_text(0, 0, 1, 1, &style, "DEBUG ON", "", false, false, true)
	}
	if startIt {
		app.baseAsset.renderEnd(true)
	}
}
