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
	"path/filepath"
	"strings"
)

func (root *Root) _createGoApp(appFolder string) error {

	vsCodeFolder := appFolder + "/.vscode"
	err := os.Mkdir(vsCodeFolder, 0700)
	if err != nil {
		return fmt.Errorf("Mkdir(%s) failed: %w", vsCodeFolder, err)
	}

	{
		str := `{
	// Use IntelliSense to learn about possible attributes.
	// Hover to view descriptions of existing attributes.
	// For more information, visit: https://go.microsoft.com/fwlink/?linkid=830387
	"version": "0.2.0",
	"configurations": [
		{
			"name": "Launch DEBUG",
			"type": "go",
			"request": "launch",
			"mode": "auto",
			"program": "${fileDirname}",
			"buildFlags": "-tags=debug"
		}
	]
}`
		path := vsCodeFolder + "/launch.json"
		err = os.WriteFile(path, []byte(str), 0644)
		if err != nil {
			return fmt.Errorf("Mkdir(%s) failed: %w", path, err)
		}
	}
	{
		str := `{
	"version": "2.0.0",
	"tasks": [
		{
			"type": "go",
			"label": "go: build workspace",
			"command": "build",
			"args": [
				"./..."
			],
			"problemMatcher": [
				"$go"
			],
			"group": "build",
		}
	]
}`
		path := vsCodeFolder + "/tasks.json"
		err = os.WriteFile(path, []byte(str), 0644)
		if err != nil {
			return fmt.Errorf("Mkdir(%s) failed: %w", path, err)
		}
	}

	{
		str := `#!/usr/bin/env bash
go build -tags=debug`
		path := appFolder + "/build_debug"
		err = os.WriteFile(path, []byte(str), 0644)
		if err != nil {
			return fmt.Errorf("Mkdir(%s) failed: %w", path, err)
		}
	}

	{
		str := `#!/usr/bin/env bash
tinygo build -o main.wasm -target=wasi`
		path := appFolder + "/build_wasm"
		err = os.WriteFile(path, []byte(str), 0644)
		if err != nil {
			return fmt.Errorf("Mkdir(%s) failed: %w", path, err)
		}
	}

	//translations
	{
		str := `{
	"COUNT.en": "Count",
	"COUNT.cs": "Přidej"
}`
		path := appFolder + "/translations.json"
		err = os.WriteFile(path, []byte(str), 0644)
		if err != nil {
			return fmt.Errorf("Mkdir(%s) failed: %w", path, err)
		}
	}

	//main.go
	{
		str := `package main

import (
	"encoding/json"
)

type Storage struct {
	Count int
}
var store Storage

type Translations struct {
	COUNT      string
}
var trns Translations

func Render() {
	SA_ColMax(0, 2)
	SA_ColMax(1, 4)

	SA_Text("").ValueInt(store.Count).Show(0, 0, 1, 1)
	if SA_Button(trns.COUNT).Show(1, 0, 1, 1).click {
		store.Count++
	}
}

var styles SA_Styles

func Init() {
	//default
	json.Unmarshal(SA_File("storage_json"), &store)
	json.Unmarshal(SA_File("translations_json:app:translations.json"), &trns)
	json.Unmarshal(SA_File("styles_json"), &styles)
}

func Save() []byte {
	js, _ := json.MarshalIndent(&store, "", "")
	return js
}`
		path := appFolder + "/main.go"
		err = os.WriteFile(path, []byte(str), 0644)
		if err != nil {
			return fmt.Errorf("Mkdir(%s) failed: %w", path, err)
		}
	}

	//copy SDK
	{
		src := "sdk/sdk.go"
		str, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("Mkdir(%s) failed: %w", src, err)
		}
		path := appFolder + "/sdk.go"
		err = os.WriteFile(path, []byte(str), 0644)
		if err != nil {
			return fmt.Errorf("Mkdir(%s) failed: %w", path, err)
		}

		src = "sdk/sdk_debug.go"
		str, err = os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("Mkdir(%s) failed: %w", src, err)
		}
		path = appFolder + "/sdk_debug.go"
		err = os.WriteFile(path, []byte(str), 0644)
		if err != nil {
			return fmt.Errorf("Mkdir(%s) failed: %w", path, err)
		}

		src = "sdk/sdk_wasi.go"
		str, err = os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("Mkdir(%s) failed: %w", src, err)
		}
		path = appFolder + "/sdk_wasi.go"
		err = os.WriteFile(path, []byte(str), 0644)
		if err != nil {
			return fmt.Errorf("Mkdir(%s) failed: %w", path, err)
		}
	}

	return nil
}

func (root *Root) CreateApp(name string, lang string) error {

	if len(name) == 0 {
		return fmt.Errorf("name is empty")
	}

	appFolder := root.folderApps + "/" + name
	err := os.Mkdir(appFolder, 0700)
	if err != nil {
		return fmt.Errorf("Mkdir(%s) failed: %w", appFolder, err)
	}

	if strings.ToLower(lang) == "go" {
		return root._createGoApp(appFolder)
	}

	return fmt.Errorf("unknown language(%s)", lang)
}

func _packageAppImport(db *Db, path string, sortPath string) error {

	dir, err := os.ReadDir(path)
	if err == nil {
		for _, file := range dir {

			fpath := path + "/" + file.Name()
			fspath := sortPath + OsTrnString(len(sortPath) > 0, "/", "") + file.Name()

			if file.IsDir() {
				err = _packageAppImport(db, path+"/"+file.Name(), fspath)
				if err != nil {
					return err
				}
			} else {
				data, err := os.ReadFile(fpath)
				if err != nil {
					return fmt.Errorf("ReadFile(%s) failed: %w", path, err)
				}

				_, err = db.Write("INSERT INTO __skyalt__(path, file) VALUES(?,?);", fspath, data)
				if err != nil {
					return fmt.Errorf("Write(INSERT INTO ...) failed: %w", err)
				}
			}
		}
	}
	return nil
}

func (root *Root) PackageApp(name string) error {

	//cut ext
	name, _ = strings.CutSuffix(name, ".sqlite")
	if len(name) == 0 {
		return fmt.Errorf("name is empty")
	}

	folderPath := root.folderApps + "/" + name
	packagePath := root.folderApps + "/" + name + ".sqlite"

	//check for 'main.wasm'
	if !OsFileExists(folderPath + "/main.wasm") {
		return fmt.Errorf("'main.wasm' is not in app folder(%s)", folderPath)
	}

	//remove old
	if OsFileExists(packagePath) {
		err := OsFileRemove(packagePath)
		if err != nil {
			return fmt.Errorf("OsFileRemove(%s) failed: %w", packagePath, err)
		}
	}

	//open db
	db, err := NewDb(root, packagePath)
	if err != nil {
		return fmt.Errorf("NewDb(%s) failed: %w", packagePath, err)
	}
	defer db.Destroy()

	_, err = db.Write("CREATE TABLE IF NOT EXISTS __skyalt__(path TEXT NOT NULL, file BLOB);")
	if err != nil {
		return fmt.Errorf("Write(CREATE TABLE ...) failed: %w", err)
	}

	err = _packageAppImport(db, folderPath, "")
	if err != nil {
		return fmt.Errorf("_packageAppImport() failed: %w", err)
	}

	err = db.Commit()
	if err != nil {
		return fmt.Errorf("Commit() failed: %w", err)
	}
	return nil
}

func (root *Root) PackageAllApps() {
	apps := root.GetAppsList()
	for _, app := range apps {
		err := root.PackageApp(app.Name)
		if err != nil {
			fmt.Printf("PackageApp(%s) failed: %v\n", app.Name, err)
		}
	}
}

func (root *Root) ExtractApp(name string) error {

	//cut ext
	name, _ = strings.CutSuffix(name, ".sqlite")
	if len(name) == 0 {
		return fmt.Errorf("name is empty")
	}

	folderPath := root.folderApps + "/" + name
	packagePath := root.folderApps + "/" + name + ".sqlite"

	//remove old
	if OsFileExists(folderPath) {
		err := OsFolderRemove(folderPath)
		if err != nil {
			return fmt.Errorf("OsFolderRemove(%s) failed: %w", folderPath, err)
		}
	}

	//open db
	db, err := NewDb(root, packagePath)
	if err != nil {
		return fmt.Errorf("NewDb(%s) failed: %w", packagePath, err)
	}
	defer db.Destroy()

	q, err := db.db.Query("SELECT path, file FROM __skyalt__")
	if err != nil {
		return fmt.Errorf("Query(SELECT ...) failed: %w", err)
	}
	for q.Next() {
		var path string
		var data []byte
		err = q.Scan(&path, &data)
		if err != nil {
			return fmt.Errorf("Scan() failed: %w", err)
		}

		path = folderPath + "/" + path

		//create folders
		dir := filepath.Dir(path)
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			return fmt.Errorf("MkdirAll(%s) failed: %w", dir, err)
		}

		//write file
		err = os.WriteFile(path, data, 0644)
		if err != nil {
			return fmt.Errorf("WriteFile(%s) failed: %w", path, err)
		}
	}
	err = q.Close()
	if err != nil {
		return fmt.Errorf("Close() failed: %w", err)
	}

	return nil
}
