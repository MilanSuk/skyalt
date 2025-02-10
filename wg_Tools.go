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
	"os"
	"path/filepath"
	"sort"
)

type Tools struct {
	//Search   string
	Selected string
}

func (layout *Layout) AddTools(x, y, w, h int, props *Tools) *Tools {
	layout._createDiv(x, y, w, h, "Tools", props.Build, nil, nil)
	return props
}

func (st *Tools) Build(layout *Layout) {
	layout.SetColumn(0, 1, 100)
	layout.SetColumnResizable(1, 3, 10, 5)
	layout.SetRow(0, 1, 100)

	//code
	fl, err := os.ReadFile(filepath.Join(st.Selected, "tool.go"))
	if err == nil {
		code := layout.AddLayoutWithName(0, 0, 1, 1, st.Selected)
		code.SetColumn(0, 1, 100)
		code.SetRow(0, 1, 100)
		code.AddTextMultiline(0, 0, 1, 1, string(fl))
	}

	//list
	folder := "tools"
	st.showTool(folder, st.getToolList(folder), layout.AddLayout(1, 0, 1, 1))

}

func (st *Tools) getToolList(folder string) []string {
	dir, _ := os.ReadDir(folder)

	var tools []string
	for _, fl := range dir {
		if !fl.IsDir() {
			continue
		}
		tools = append(tools, fl.Name())
	}
	sort.Strings(tools)

	return tools

}
func (st *Tools) showTool(folder string, tools []string, layout *Layout) {
	layout.SetColumn(0, 0.5, 0.5)
	layout.SetColumn(1, 1, 100)

	y := 0
	for _, fl := range tools {

		subFolder := filepath.Join(folder, fl)

		bt := layout.AddButton(0, y, 2, 1, fl)
		if subFolder == st.Selected {
			bt.Background = 1
		} else {
			bt.Background = 0.25
		}
		bt.Align = 0
		bt.clicked = func() {
			st.Selected = subFolder
		}
		y++

		subTools := st.getToolList(subFolder)
		if len(subTools) > 0 {
			layout.SetRowFromSub(y, 1, 100)

			subLayout := layout.AddLayout(1, y, 1, 1)
			st.showTool(subFolder, subTools, subLayout)
			y++
		}
	}
}
