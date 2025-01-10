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

type Root struct {
}

func (layout *Layout) AddRoot(x, y, w, h int) *Root {
	props := &Root{}
	layout._createDiv(x, y, w, h, "Root", props.Build, nil, nil)
	return props
}

func (st *Root) Build(layout *Layout) {

	layout.SetColumn(0, 1, 100)

	layout.SetRowFromSub(0, 1, 100)
	layout.SetRow(1, 0.1, 0.1)
	layout.SetRow(2, 1, 100)

	header := OpenFile_RootHeader()
	layout.AddRootHeader(0, 0, 1, 1, header)

	layout.AddDivider(0, 1, 1, 1, true).Margin = 0

	AppDiv := layout.AddLayout(0, 2, 1, 1)
	AppDiv.App = true
	AppDiv.SetColumn(0, 1, 100)
	AppDiv.SetRow(0, 1, 100)
	if header.ShowPromptList {
		AppDiv.AddPrompts(0, 0, 1, 1)
	} else {
		//App
		AppDiv.AddShowApp(0, 0, 1, 1)
	}

	//Assistant panel
	/*if OpenFile_Assistant().Show {
		layout.SetColumnResizable(1, 5, 20, 6)
		layout.AddAssistant(1, 0, 1, 2, OpenFile_Assistant())
	}*/
}
