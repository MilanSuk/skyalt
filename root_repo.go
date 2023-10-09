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

func (root *Root) _createGoApp(appFolder string) bool {

	vsCodeFolder := appFolder + "/.vscode"
	err := os.Mkdir(vsCodeFolder, 0700)
	if err != nil {
		fmt.Printf("Mkdir(%s) failed: %v", vsCodeFolder, err)
		return false
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
			fmt.Printf("Mkdir(%s) failed: %v", path, err)
			return false
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
			fmt.Printf("Mkdir(%s) failed: %v", path, err)
			return false
		}
	}

	{
		str := `#!/usr/bin/env bash
go build -tags=debug`
		path := appFolder + "/build_debug"
		err = os.WriteFile(path, []byte(str), 0644)
		if err != nil {
			fmt.Printf("Mkdir(%s) failed: %v", path, err)
			return false
		}
	}

	{
		str := `#!/usr/bin/env bash
tinygo build -o main.wasm -target=wasi`
		path := appFolder + "/build_wasm"
		err = os.WriteFile(path, []byte(str), 0644)
		if err != nil {
			fmt.Printf("Mkdir(%s) failed: %v", path, err)
			return false
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
			fmt.Printf("Mkdir(%s) failed: %v", path, err)
			return false
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
			fmt.Printf("Mkdir(%s) failed: %v", path, err)
			return false
		}
	}

	//copy SDK
	{
		src := "sdk/sdk.go"
		str, err := os.ReadFile(src)
		if err != nil {
			fmt.Printf("Mkdir(%s) failed: %v", src, err)
			return false
		}
		path := appFolder + "/sdk.go"
		err = os.WriteFile(path, []byte(str), 0644)
		if err != nil {
			fmt.Printf("Mkdir(%s) failed: %v", path, err)
			return false
		}

		src = "sdk/sdk_debug.go"
		str, err = os.ReadFile(src)
		if err != nil {
			fmt.Printf("Mkdir(%s) failed: %v", src, err)
			return false
		}
		path = appFolder + "/sdk_debug.go"
		err = os.WriteFile(path, []byte(str), 0644)
		if err != nil {
			fmt.Printf("Mkdir(%s) failed: %v", path, err)
			return false
		}

		src = "sdk/sdk_wasi.go"
		str, err = os.ReadFile(src)
		if err != nil {
			fmt.Printf("Mkdir(%s) failed: %v", src, err)
			return false
		}
		path = appFolder + "/sdk_wasi.go"
		err = os.WriteFile(path, []byte(str), 0644)
		if err != nil {
			fmt.Printf("Mkdir(%s) failed: %v", path, err)
			return false
		}
	}

	return true
}

func (root *Root) CreateApp(name string, lang string) bool {

	appFolder := root.folderApps + "/" + name
	err := os.Mkdir(appFolder, 0700)
	if err != nil {
		fmt.Printf("Mkdir(%s) failed: %v", appFolder, err)
		return false
	}

	if strings.ToLower(lang) == "go" {
		return root._createGoApp(appFolder)
	}

	return false
}

func (root *Root) PackageApp(name string) bool {

	//...
	return true
}

func (root *Root) ExtractApp(name string) bool {

	//...
	return true
}
