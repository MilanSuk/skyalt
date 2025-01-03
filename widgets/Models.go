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

type ModelsServices struct {
	Service string
	Models  []string
}

var g_models_services = []ModelsServices{
	{Service: "xai", Models: []string{"grok-2-1212", "grok-beta"}},
	{Service: "openai", Models: []string{"gpt-4", "gpt-4-turbo", "gpt-4o"}},
	{Service: "anthropic", Models: []string{"claude-3-5-haiku-latest", "claude-3-5-sonnet-latest"}},
	{Service: "groq", Models: []string{"llama-3.3-70b-specdec", "llama-3.3-70b-versatile"}},
}

type Models struct {
	Model string
}

func (layout *Layout) AddModels(x, y, w, h int, props *Models) *Models {
	layout._createDiv(x, y, w, h, "Models", props.Build, nil, nil)
	return props
}

func (st *Models) Build(layout *Layout) {
	layout.SetColumn(0, 4, 100)
	layout.SetColumn(1, 1, 1)

	icon_margin := 0.2

	y := 0

	//XAI
	{
		//settings
		dia, diaLay := layout.AddDialogBorder("xai", "XAI", 15)
		diaLay.SetColumn(0, 1, 100)
		diaLay.SetRowFromSub(0)
		srv := diaLay.AddXai(0, 0, 1, 1, OpenFile_Xai())

		//service
		layout.AddText(0, y, 1, 1, "XAI").Align_h = 1
		bt := layout.AddButtonMenu(1, y, 1, 1, "", "resources/settings.png", icon_margin)
		bt.clicked = func() {
			dia.OpenCentered()
			//dia.OpenRelative(layout)
		}
		y++

		// models
		enable := srv.Enable && srv.Api_key != ""
		for _, it := range g_models_services {
			if it.Service == "xai" {
				for _, m := range it.Models {
					st._addModel(layout, y, m, enable)
					y++
				}
			}
		}
	}

	layout.AddDivider(0, y, 2, 1, true)
	layout.SetRow(y, 0.1, 0.1)
	y++

	//OpenAI
	{
		//settings
		dia, diaLay := layout.AddDialogBorder("openai", "OpenAI", 15)
		diaLay.SetColumn(0, 1, 100)
		diaLay.SetRowFromSub(0)
		srv := diaLay.AddOpenAI(0, 0, 1, 1, OpenFile_OpenAI())

		//service
		layout.AddText(0, y, 1, 1, "OpenAI").Align_h = 1
		bt := layout.AddButtonMenu(1, y, 1, 1, "", "resources/settings.png", icon_margin)
		bt.clicked = func() {
			dia.OpenCentered()
		}
		y++

		// models
		enable := srv.Enable && srv.Api_key != ""
		for _, it := range g_models_services {
			if it.Service == "openai" {
				for _, m := range it.Models {
					st._addModel(layout, y, m, enable)
					y++
				}
			}
		}
	}

	layout.AddDivider(0, y, 2, 1, true)
	layout.SetRow(y, 0.1, 0.1)
	y++

	//Anthropic
	{
		//settings
		dia, diaLay := layout.AddDialogBorder("anthropic", "Anthropic", 15)
		diaLay.SetColumn(0, 1, 100)
		diaLay.SetRowFromSub(0)
		srv := diaLay.AddAnthropic(0, 0, 1, 1, OpenFile_Anthropic())

		//service
		layout.AddText(0, y, 1, 1, "Anthropic").Align_h = 1
		bt := layout.AddButtonMenu(1, y, 1, 1, "", "resources/settings.png", icon_margin)
		bt.clicked = func() {
			dia.OpenCentered()
		}
		y++

		// models
		enable := srv.Enable && srv.Api_key != ""
		for _, it := range g_models_services {
			if it.Service == "anthropic" {
				for _, m := range it.Models {
					st._addModel(layout, y, m, enable)
					y++
				}
			}
		}
	}

	layout.AddDivider(0, y, 2, 1, true)
	layout.SetRow(y, 0.1, 0.1)
	y++

	//Groq
	{
		//settings
		dia, diaLay := layout.AddDialogBorder("groq", "Groq", 15)
		diaLay.SetColumn(0, 1, 100)
		diaLay.SetRowFromSub(0)
		srv := diaLay.AddGroq(0, 0, 1, 1, OpenFile_Groq())

		//service
		layout.AddText(0, y, 1, 1, "Groq").Align_h = 1
		bt := layout.AddButtonMenu(1, y, 1, 1, "", "resources/settings.png", icon_margin)
		bt.clicked = func() {
			dia.OpenCentered()
		}
		y++

		// models
		enable := srv.Enable && srv.Api_key != ""
		for _, it := range g_models_services {
			if it.Service == "groq" {
				for _, m := range it.Models {
					st._addModel(layout, y, m, enable)
					y++
				}
			}
		}
	}

	//save attr + load in Assistant ........

}

func (st *Models) _addModel(layout *Layout, y int, model string, enable bool) {
	b, bLay := layout.AddButtonMenu2(0, y, 2, 1, model, "", 0)
	bLay.Enable = enable
	if st.Model == model {
		b.Background = 1
	}
	b.clicked = func() {
		st.Model = model
	}
}

func (st *Models) GetService() string {
	for _, it := range g_models_services {
		for _, m := range it.Models {
			if m == st.Model {
				return it.Service
			}
		}
	}
	return ""
}
