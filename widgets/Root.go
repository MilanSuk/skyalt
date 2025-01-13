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
	}
}

func (st *Root) OpenApp(folder string, name string) {
	files, _ := GetListOfFiles(folder)

	//exact
	for _, file := range files {
		if file.Name == name {
			st.AppPath = file.GetPath()
		}
	}

	//contain
	for _, file := range files {
		if strings.Contains(file.Name, name) {
			st.AppPath = file.GetPath()
		}
	}
}
