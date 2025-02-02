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

type LLMList struct {
	Model      *string
	Login_path *string
}

func (layout *Layout) AddLLMList(x, y, w, h int, model *string) *LLMList {
	props := &LLMList{Model: model}
	layout._createDiv(x, y, w, h, "LLMList", props.Build, nil, nil)
	return props
}

func (st *LLMList) Build(layout *Layout) {

	icon_margin := 0.2

	layout.SetColumn(0, 1, 100)

	//search ....

	y := 0
	logins, _ := OpenDir_llms_logins() //err ....
	for _, login_path := range logins {

		login := OpenFile_LLMLogin(login_path)

		//settings
		dia, diaLay := layout.AddDialogBorder(login_path, login.Label, 22)
		diaLay.SetColumn(0, 1, 100)
		diaLay.SetRowFromSub(0, 1, 100)
		srv := diaLay.AddLLmLogin(0, 0, 1, 1, login)
		enable := srv.Api_key_id != ""

		//service
		layout.AddText(0, y, 1, 1, "<b>"+login.Label).Align_h = 1
		bt := layout.AddButtonMenu(1, y, 1, 1, "", "resources/settings.png", icon_margin)
		bt.clicked = func() {
			dia.OpenCentered()
			//dia.OpenRelative(layout)
		}
		y++

		//models
		for _, m := range login.ChatModels {
			st._addModelTTT(layout, y, m.Name, enable)
			y++
		}

		layout.AddDivider(0, y, 2, 1, true)
		layout.SetRow(y, 0.1, 0.1)
		y++
	}
}

func (st *LLMList) _addModelTTT(layout *Layout, y int, model string, enable bool) {
	b, bLay := layout.AddButtonMenu2(0, y, 2, 1, model, "", 0)
	bLay.Enable = enable
	if *st.Model == model {
		b.Background = 1
	}
	b.clicked = func() {
		*st.Model = model
	}
}
