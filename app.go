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
	"errors"
	"fmt"
)

type App struct {
	db        *Db
	app_rowid int
	name      string

	wasm  *AppWasm
	debug *AppDebug

	styles *DivStyles

	reopen bool

	logs []string

	gui *LayoutSave

	images []*Image

	info_string string
}

func NewApp(db *Db, app_rowid int) (*App, error) {
	var app App
	app.db = db
	app.app_rowid = app_rowid
	app.reopen = true

	//get app name
	{
		err := app.db.InitTable()
		if err != nil {
			return nil, fmt.Errorf("InitTable() failed: %w", err)
		}

		app.name, _, err = app.GetAppName()
	}

	//extract
	if !OsFolderExists(app.getPath()) {
		app.db.root.ExtractApp(app.name)
	}

	//load wasm
	var err error
	app.wasm, err = NewAppWasm(&app)
	if err != nil {
		return nil, err
	}

	app.styles = NewDivStyles()

	app.Tick()

	return &app, nil
}
func (app *App) Destroy() {
	app.Save()

	if app.wasm != nil {
		app.wasm.Destroy()
	}
	if app.debug != nil {
		app.debug.Destroy()
	}

	for _, img := range app.images {
		err := img.Destroy()
		if err != nil {
			fmt.Printf("Destroy() failed: %v\n", err)
		}
	}
}

func (db *Db) InitTable() error {

	_, err := db.Write("CREATE TABLE IF NOT EXISTS __skyalt__(label TEXT NOT NULL, sort REAL NOT NULL, app TEXT NOT NULL, storage BLOB, gui BLOB);")
	if err != nil {
		return fmt.Errorf("Write() failed: %w", err)
	}
	err = db.Commit()
	if err != nil {
		return fmt.Errorf("Commit() failed: %w", err)
	}
	return nil
}

func (app *App) GetAppName() (string, bool, error) {
	rows, err := app.db.db.Query("SELECT app FROM __skyalt__ WHERE rowid=?", app.app_rowid)
	if err != nil {
		return "", false, fmt.Errorf("query SELECT failed: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return "", false, nil
	}

	var name string
	err = rows.Scan(&name)
	if err != nil {
		return "", false, fmt.Errorf("Scan() failed: %w", err)
	}

	return name, true, nil //true = row exist
}

func (db *Db) AddFirstRowId(appName string) error {

	err := db.InitTable()
	if err != nil {
		return fmt.Errorf("InitTable() failed: %w", err)
	}

	//check 0 rows
	rows, err := db.db.Query("SELECT app FROM __skyalt__")
	if err != nil {
		return fmt.Errorf("SELECT() failed: %w", err)
	}
	defer rows.Close()
	if rows.Next() {
		return nil //ok
	}

	//insert
	_, err = db.Write("INSERT INTO __skyalt__(label, app, sort) VALUES(?,?,?);", appName, appName, 0)
	if err != nil {
		return fmt.Errorf("Write() failed: %w", err)
	}
	err = db.Commit()
	if err != nil {
		return fmt.Errorf("Commit() failed: %w", err)
	}

	return nil
}

func (app *App) Save() {
	if app.debug != nil {
		app.debug.SaveData(app)
	} else if app.wasm != nil {
		app.wasm.SaveData()
	}

	err := app.saveGui()
	app.AddLogErr(err)
}

func (app *App) AddLogErr(err error) bool {
	if err != nil {
		//print
		fmt.Printf("Error(%s): %v\n", app.getPath(), err)

		//add
		app.logs = append(app.logs, err.Error())
		return true
	}
	return false
}

func (app *App) GetLog(remove bool) string {
	var ret string
	if len(app.logs) > 0 {
		ret = app.logs[0]
		if remove {
			app.logs = app.logs[1:] //cut
		}
	}
	return ret
}

func (app *App) GetStorage() ([]byte, error) {

	row := app.db.db.QueryRow("SELECT storage FROM __skyalt__ WHERE rowid=?", app.app_rowid)

	var js []byte
	err := row.Scan(&js)
	if err != nil {
		return nil, fmt.Errorf("Scan() failed: %w", err)
	}
	return js, nil
}

func (app *App) SetStorage(js []byte) error {

	_, err := app.db.Write("UPDATE __skyalt__ SET storage=? WHERE rowid=?;", js, app.app_rowid)
	if err != nil {
		return fmt.Errorf("Write() failed: %w", err)
	}

	app.db.Commit()
	return nil
}

func (app *App) GetGui() ([]byte, error) {

	row := app.db.db.QueryRow("SELECT gui FROM __skyalt__ WHERE rowid=?", app.app_rowid)

	var js []byte
	err := row.Scan(&js)
	if err != nil {
		return nil, fmt.Errorf("Scan() failed: %w", err)
	}
	return js, nil
}

func (app *App) SetGui(js []byte) error {

	_, err := app.db.Write("UPDATE __skyalt__ SET gui=? WHERE rowid=?;", js, app.app_rowid)
	if err != nil {
		return fmt.Errorf("Write() failed: %w", err)
	}

	err = app.db.Commit()
	if err != nil {
		return fmt.Errorf("Commit() failed: %w", err)
	}

	return nil
}

func (app *App) getPath() string {
	return app.db.root.folderApps + "/" + app.name
}

func (app *App) getWasmPath() string {
	return app.getPath() + "/main.wasm"
}

func (app *App) IsReadyToFire() bool {
	if app.debug != nil {
		return app.debug.conn != nil
	}
	if app.wasm != nil {
		return app.wasm.mod != nil
	}
	return false
}

func (app *App) Call(fnName string) (int64, error) {
	var ret int64
	var err error

	if app.debug != nil {
		ret, err = app.debug.Call(fnName, app)
	} else if app.wasm != nil {
		ret, err = app.wasm.Call(fnName)
	} else {
		err = errors.New("no call")
	}

	return ret, err
}

func (app *App) Render(startIt bool) {

	if startIt {
		app.renderStart()
	}
	if app.IsReadyToFire() {
		_, err := app.Call("_sa_render")
		if err != nil {
			fmt.Print(err)
		}

		app.TryConnectDebug()
	} else {
		app.db.root.styles.TextCenter.Paint(app.getCoord(0, 0, 1, 1, 0, 0, 0), "Error: 'Main.wasm' is missing or corrupted", "", false, false, "", 0, false, app)
	}

	if app.debug != nil {
		//draw blue rectangle, when debug mode is active
		blue := OsCd{50, 50, 255, 180}

		style := app.db.root.styles.Text
		style.Margin(0.06)
		style.BorderCd(blue)
		style.Border(0.03)
		style.FontAlignV(2)
		style.FontAlignH(2)
		style.Color(blue)
		style.Cursor("")

		style.Paint(app.getCoord(0, 0, 1, 1, 0, 0, 0), "DEBUG ON    ", "", false, false, "", 0, false, app)
	}
	if startIt {
		app.renderEnd(true)
	}
}

func (app *App) CheckDebug() {
	if app.debug != nil {
		if app.debug.conn == nil {
			app.debug.Destroy()
			app.debug = nil
		}
	}
}

func (app *App) TryConnectDebug() {
	if app.debug == nil {
		appDebug := app.db.root.server.Get(app.name)
		if appDebug != nil {
			app.Save()
			app.debug = appDebug
			app.CallOpen()
		}
	}

	app.CheckDebug()
}

func (app *App) SetupDB() {
	app.Call("_sa_setup_db")
}

func (app *App) CallOpen() {
	app.Call("_sa_open")

	//if db is empty, call SetupDB()
	{
		rows, err := app.db.db.Query("SELECT name FROM sqlite_master WHERE type='table'")
		if err != nil {
			fmt.Printf("Query() failed: %v\n", err)
			return
		}
		defer rows.Close()

		num_tables := 0
		for rows.Next() {
			var nm string
			err = rows.Scan(&nm)
			if err != nil {
				fmt.Printf("Scan() failed: %v\n", err)
				return
			}

			if nm != "__skyalt__" {
				num_tables++
			}
		}
		if num_tables == 0 {
			app.SetupDB()
		}
	}
}

func (app *App) Tick() bool {

	for i := len(app.images) - 1; i >= 0; i-- {
		ok, _ := app.images[i].Maintenance(app.db.root.ui.render)
		if !ok {
			app.images = append(app.images[:i], app.images[i+1:]...)
		}
	}

	if app.debug != nil {
		app.CheckDebug()
	} else if app.wasm != nil {
		changed, err := app.wasm.Tick()
		if err != nil {
			app.AddLogErr(err)
			app.wasm = nil
		} else if changed {
			app.reopen = true
		}
	}

	{
		nm, exist, err := app.GetAppName()
		if err != nil {
			fmt.Printf("GetAppName() failed: %v\n", err)
			return true
		}

		if !exist {
			return false //remove from db
		}

		if nm != app.name {
			app.Save()
			app.images = nil

			app.name = nm //this change wasm path => app.wasm.Tick() will have different time => reload wasm + reopen
		}
	}

	//init
	if app.reopen {
		app.CallOpen()
		app.reopen = false
	}

	return true
}

func (app *App) SetDebugLine(line string) {

	st := app.db.root.levels.GetStack()
	if st.stack == nil || st.stack.crop.IsZero() {
		return
	}

	if st.stack.enableInput {
		if st.stack.crop.Inside(app.db.root.ui.io.touch.pos) {
			app.db.root.SetDebugLine(line)
		}
	}
}

func (app *App) CheckLoadGui() error {

	if app.gui != nil {
		return nil
	}

	js, err := app.GetGui()
	if err != nil {
		return fmt.Errorf("GetGui() failed: %w", err)
	}

	app.gui, err = NewRS_LScroll(js)
	if err != nil {
		return fmt.Errorf("NewRS_LScroll() failed: %w", err)
	}
	return nil
}

func (app *App) FindGlobalScrollHash(hash uint64) *LayoutSaveItem {

	err := app.CheckLoadGui()
	if err != nil {
		fmt.Printf("CheckLoadGui() failed: %v\n", err)
		return nil
	}
	return app.gui.FindGlobalScrollHash(hash)
}

func (app *App) AddGlobalScrollHash(hash uint64) *LayoutSaveItem {

	err := app.CheckLoadGui()
	if err != nil {
		fmt.Printf("CheckLoadGui() failed: %v\n", err)
		return nil
	}
	return app.gui.AddGlobalScrollHash(hash)
}

func (app *App) saveGui() error {

	err := app.CheckLoadGui()
	if err != nil {
		return err
	}

	js, err := app.GetGui()
	if err != nil {
		return fmt.Errorf("GetGui() failed: %w", err)
	}

	//save
	app.db.root.levels.Save()

	//convert into json
	js, err = app.gui.Save(js)
	if err != nil {
		return fmt.Errorf("Save() failed: %w", err)
	}

	//write into db
	err = app.SetGui(js)
	if err != nil {
		return err
	}

	return nil
}
