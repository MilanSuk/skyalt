/*
Copyright 2024 Milan Suk

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
	"strings"
)

type Root struct {
	AppPath string
}

func (st *Root) Build(layout *Layout) {
	layout.SetColumn(0, 1, 100)

	layout.SetRowFromSub(0, 1, 100)
	layout.SetRow(1, 0.1, 0.1)
	layout.SetRow(2, 1, 100)

	header := OpenFile_RootHeader()
	layout.AddRootHeader(0, 0, 1, 1, header)

	layout.AddDivider(0, 1, 1, 1, true).Margin = 0

	appLay := layout.AddApp(0, 2, 1, 1, st.AppPath)
	if appLay != nil {
		appLay.App = true
	} else {
		tx := layout.AddText(0, 2, 1, 1, fmt.Sprintf("App '%s' not found", st.AppPath))
		tx.Align_h = 1
		tx.Cd = Paint_GetPalette().E
	}
}

func (st *Root) OpenApp(folder string, name string) bool {
	name = strings.ToLower(name)

	files, _ := GetListOfFiles(folder)

	//find - exact
	for _, file := range files {
		if strings.ToLower(file.Name) == name {
			st.AppPath = file.GetPath()
			return true
		}
	}

	//find - contain
	for _, file := range files {
		if strings.Contains(strings.ToLower(file.Name), name) {
			st.AppPath = file.GetPath()
			return true
		}
	}

	Layout_WriteError(fmt.Errorf("App %s not found", name))
	return false
}

func (st *Root) OpenAppForce(folder string, tp string, name string) {
	f := _File{Folder: folder, Type: tp, Name: name}
	st.AppPath = f.GetPath()
}
