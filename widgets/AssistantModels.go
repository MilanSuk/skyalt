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

type AssistantModelsServices struct {
	Service         string
	AssistantModels []string
}

// note: make sure that different services don't have model with same name!
var g__models_ttt = []AssistantModelsServices{
	{Service: "xai", AssistantModels: []string{"grok-2-1212", "grok-beta"}},
	{Service: "openai", AssistantModels: []string{"gpt-4", "gpt-4-turbo", "gpt-4o"}},
	{Service: "anthropic", AssistantModels: []string{"claude-3-5-haiku-latest", "claude-3-5-sonnet-latest"}},
	{Service: "groq", AssistantModels: []string{"llama-3.3-70b-specdec", "llama-3.3-70b-versatile"}},
	{Service: "llamacpp", AssistantModels: nil},
}
var g__models_stt = []AssistantModelsServices{
	{Service: "openai", AssistantModels: OpenAI_GetSTTModelList()},
	{Service: "whispercpp", AssistantModels: nil},
	{Service: "groq", AssistantModels: Groq_GetSTTModelList()},
}

type AssistantModels struct {
	TextModel string
	STTModel  string
}

func (layout *Layout) AddAssistantModels(x, y, w, h int, props *AssistantModels) *AssistantModels {
	layout._createDiv(x, y, w, h, "AssistantModels", props.Build, nil, nil)
	return props
}

func (st *AssistantModels) Build(layout *Layout) {
	layout.SetColumn(0, 4, 100)
	layout.SetColumn(1, 0.1, 1) //divider
	layout.SetColumn(2, 4, 100)
	layout.SetRowFromSub(1, 1, 100)
	//layout.SetRow(1, 9, 9)

	icon_margin := 0.2

	layout.AddText(0, 0, 1, 1, "<b>Text -> Text").Align_h = 1
	layout.AddText(2, 0, 1, 1, "<b>Speech -> Text").Align_h = 1

	layout.AddDivider(1, 0, 1, 2, false)

	//text to text
	{
		layTTT := layout.AddLayout(0, 1, 1, 1)
		layTTT.SetColumn(0, 1, 100)

		y := 0
		//xAI
		{
			//settings
			dia, diaLay := layTTT.AddDialogBorder("xai", "xAI", 15)
			diaLay.SetColumn(0, 1, 100)
			diaLay.SetRowFromSub(0, 1, 100)
			srv := diaLay.AddXai(0, 0, 1, 1, OpenFile_Xai())

			//service
			layTTT.AddText(0, y, 1, 1, "xAI").Align_h = 1
			bt := layTTT.AddButtonMenu(1, y, 1, 1, "", "resources/settings.png", icon_margin)
			bt.clicked = func() {
				dia.OpenCentered()
				//dia.OpenRelative(layout)
			}
			y++

			//models
			enable := srv.Enable && srv.Api_key != ""
			for _, it := range g__models_ttt {
				if it.Service == "xai" {
					for _, m := range it.AssistantModels {
						st._addModelTTT(layTTT, y, m, enable)
						y++
					}
				}
			}
		}

		layTTT.AddDivider(0, y, 2, 1, true)
		layTTT.SetRow(y, 0.1, 0.1)
		y++

		//OpenAI
		{
			//settings
			dia, diaLay := layTTT.AddDialogBorder("openai", "OpenAI", 15)
			diaLay.SetColumn(0, 1, 100)
			diaLay.SetRowFromSub(0, 1, 100)
			srv := diaLay.AddOpenAI(0, 0, 1, 1, OpenFile_OpenAI())

			//service
			layTTT.AddText(0, y, 1, 1, "OpenAI").Align_h = 1
			bt := layTTT.AddButtonMenu(1, y, 1, 1, "", "resources/settings.png", icon_margin)
			bt.clicked = func() {
				dia.OpenCentered()
			}
			y++

			//models
			enable := srv.Enable && srv.Api_key != ""
			for _, it := range g__models_ttt {
				if it.Service == "openai" {
					for _, m := range it.AssistantModels {
						st._addModelTTT(layTTT, y, m, enable)
						y++
					}
				}
			}
		}

		layTTT.AddDivider(0, y, 2, 1, true)
		layTTT.SetRow(y, 0.1, 0.1)
		y++

		//Anthropic
		{
			//settings
			dia, diaLay := layTTT.AddDialogBorder("anthropic", "Anthropic", 15)
			diaLay.SetColumn(0, 1, 100)
			diaLay.SetRowFromSub(0, 1, 100)
			srv := diaLay.AddAnthropic(0, 0, 1, 1, OpenFile_Anthropic())

			//service
			layTTT.AddText(0, y, 1, 1, "Anthropic").Align_h = 1
			bt := layTTT.AddButtonMenu(1, y, 1, 1, "", "resources/settings.png", icon_margin)
			bt.clicked = func() {
				dia.OpenCentered()
			}
			y++

			//models
			enable := srv.Enable && srv.Api_key != ""
			for _, it := range g__models_ttt {
				if it.Service == "anthropic" {
					for _, m := range it.AssistantModels {
						st._addModelTTT(layTTT, y, m, enable)
						y++
					}
				}
			}
		}

		layTTT.AddDivider(0, y, 2, 1, true)
		layTTT.SetRow(y, 0.1, 0.1)
		y++

		//Groq
		{
			//settings
			dia, diaLay := layTTT.AddDialogBorder("groq", "Groq", 15)
			diaLay.SetColumn(0, 1, 100)
			diaLay.SetRowFromSub(0, 1, 100)
			srv := diaLay.AddGroq(0, 0, 1, 1, OpenFile_Groq())

			//service
			layTTT.AddText(0, y, 1, 1, "Groq").Align_h = 1
			bt := layTTT.AddButtonMenu(1, y, 1, 1, "", "resources/settings.png", icon_margin)
			bt.clicked = func() {
				dia.OpenCentered()
			}
			y++

			//models
			enable := srv.Enable && srv.Api_key != ""
			for _, it := range g__models_ttt {
				if it.Service == "groq" {
					for _, m := range it.AssistantModels {
						st._addModelTTT(layTTT, y, m, enable)
						y++
					}
				}
			}
		}

		layTTT.AddDivider(0, y, 2, 1, true)
		layTTT.SetRow(y, 0.1, 0.1)
		y++

		//Llama.cpp
		{
			//settings
			dia, diaLay := layTTT.AddDialogBorder("llama.cpp", "Llama.cpp", 15)
			diaLay.SetColumn(0, 1, 100)
			diaLay.SetRowFromSub(0, 1, 100)
			srv := diaLay.AddLlamacpp(0, 0, 1, 1, OpenFile_Llamacpp())

			//service
			layTTT.AddText(0, y, 1, 1, "Llama.cpp").Align_h = 1
			bt := layTTT.AddButtonMenu(1, y, 1, 1, "", "resources/settings.png", icon_margin)
			bt.clicked = func() {
				dia.OpenCentered()
			}
			y++

			//models
			for si, it := range g__models_ttt {
				if it.Service == "llamacpp" {
					models := srv.GetModelList()
					g__models_ttt[si].AssistantModels = models //update
					for _, mi := range models {
						st._addModelTTT(layTTT, y, mi, true)
						y++
					}
				}
			}
		}
	}

	//save attr + load in Assistant ........

	//speech to text
	{
		laySTT := layout.AddLayout(2, 1, 1, 1)
		laySTT.SetColumn(0, 1, 100)
		y := 0
		//OpenAI
		{
			//settings
			dia, diaLay := laySTT.AddDialogBorder("openai", "OpenAI", 15)
			diaLay.SetColumn(0, 1, 100)
			diaLay.SetRowFromSub(0, 1, 100)
			srv := diaLay.AddOpenAI(0, 0, 1, 1, OpenFile_OpenAI())

			//service
			laySTT.AddText(0, y, 1, 1, "OpenAI").Align_h = 1
			bt := laySTT.AddButtonMenu(1, y, 1, 1, "", "resources/settings.png", icon_margin)
			bt.clicked = func() {
				dia.OpenCentered()
			}
			y++

			//models
			enable := srv.Enable && srv.Api_key != ""
			for _, it := range g__models_stt {
				if it.Service == "openai" {
					for _, m := range it.AssistantModels {
						st._addModelSTT(laySTT, y, m, m, enable)
						y++
					}
				}
			}
		}

		//Groq - right now it does not suppport 'timestamp_granularities', which are need for precisi word timestamps!
		/*{
			//settings
			dia, diaLay := laySTT.AddDialogBorder("groq", "Groq", 15)
			diaLay.SetColumn(0, 1, 100)
			diaLay.SetRowFromSub(0, 1, 100)
			srv := diaLay.AddGroq(0, 0, 1, 1, OpenFile_Groq())

			//service
			laySTT.AddText(0, y, 1, 1, "Groq").Align_h = 1
			bt := laySTT.AddButtonMenu(1, y, 1, 1, "", "resources/settings.png", icon_margin)
			bt.clicked = func() {
				dia.OpenCentered()
			}
			y++

			//models
			enable := srv.Enable && srv.Api_key != ""
			for _, it := range g__models_stt {
				if it.Service == "groq" {
					for _, m := range it.AssistantModels {
						st._addModelSTT(laySTT, y, m, m, enable)
						y++
					}
				}
			}
		}*/

		laySTT.AddDivider(0, y, 2, 1, true)
		laySTT.SetRow(y, 0.1, 0.1)
		y++

		//whisper.cpp
		{
			//settings
			dia, diaLay := laySTT.AddDialogBorder("whisper.cpp", "Whisper.cpp", 15)
			diaLay.SetColumn(0, 1, 100)
			diaLay.SetRowFromSub(0, 1, 100)
			srv := diaLay.AddWhispercpp(0, 0, 1, 1, OpenFile_Whispercpp())

			//service
			laySTT.AddText(0, y, 1, 1, "Whisper.cpp").Align_h = 1
			bt := laySTT.AddButtonMenu(1, y, 1, 1, "", "resources/settings.png", icon_margin)
			bt.clicked = func() {
				dia.OpenCentered()
			}
			y++

			//models
			for si, it := range g__models_stt {
				if it.Service == "whispercpp" {
					modelNames, modelFiles := srv.GetModelList()
					g__models_stt[si].AssistantModels = modelFiles //update
					for mi := range modelNames {
						st._addModelSTT(laySTT, y, modelNames[mi], modelFiles[mi], true)
						y++
					}
				}
			}
		}
	}
}

func (st *AssistantModels) _addModelTTT(layout *Layout, y int, model string, enable bool) {
	b, bLay := layout.AddButtonMenu2(0, y, 2, 1, model, "", 0)
	bLay.Enable = enable
	if st.TextModel == model {
		b.Background = 1
	}
	b.clicked = func() {
		st.TextModel = model
	}
}
func (st *AssistantModels) _addModelSTT(layout *Layout, y int, name string, model string, enable bool) {
	b, bLay := layout.AddButtonMenu2(0, y, 2, 1, name, "", 0)
	b.Tooltip = model
	bLay.Enable = enable
	if st.STTModel == model {
		b.Background = 1
	}
	b.clicked = func() {
		st.STTModel = model
	}
}

func (st *AssistantModels) GetTTService() string {
	for _, it := range g__models_ttt {
		for _, m := range it.AssistantModels {
			if m == st.TextModel {
				return it.Service
			}
		}
	}
	return ""
}

func (st *AssistantModels) GetSTTService() string {
	for _, it := range g__models_stt {
		for _, m := range it.AssistantModels {
			if m == st.STTModel {
				return it.Service
			}
		}
	}
	return ""
}
