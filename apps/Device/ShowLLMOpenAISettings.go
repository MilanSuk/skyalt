package main

import (
	"fmt"
)

// Show OpenAI settings. User can change API key and chat/image/STT/TTS models.
type ShowLLMOpenAISettings struct {
}

func (st *ShowLLMOpenAISettings) run(caller *ToolCaller, ui *UI) error {
	source_llm, err := NewLLMOpenai("")
	if err != nil {
		return err
	}

	source_llm.Check()

	ui.SetColumn(0, 1, 5)
	ui.SetColumn(1, 1, 20)

	//title
	ui.AddTextLabel(0, 0, 2, 1, source_llm.Provider)

	y := 1

	y++ //space

	ui.AddText(0, y, 1, 1, "OpenAI API endpoint")
	ui.AddText(1, y, 1, 1, source_llm.OpenAI_url)
	y++

	//api key
	{
		tx := ui.AddText(0, y, 1, 1, "API key")
		if source_llm.API_key == "" {
			tx.Cd = UI_GetPalette().E
		}

		KeyEd := ui.AddEditboxString(1, y, 1, 1, &source_llm.API_key)
		KeyEd.Password = true
		KeyEd.changed = func() error {
			return nil
		}

		y++
	}

	KeyBt := ui.AddButton(1, y, 1, 1, "Get API key")
	KeyBt.Align = 0
	KeyBt.Background = 0
	KeyBt.BrowserUrl = source_llm.DevUrl
	y++

	y++ //space

	//Models
	ui.SetRowFromSub(y, 1, 100, true)
	ModelsDiv := ui.AddLayout(0, y, 2, 1)
	y++
	ModelsDiv.SetColumn(0, 5, 5)
	ModelsDiv.SetColumn(1, 1, 100)
	ModelsDiv.SetColumn(2, 1, 100)
	my := 0

	ModelsDiv.AddText(0, my, 2, 1, "Language models")
	for _, it := range source_llm.LanguageModels {
		ModelsDiv.AddText(1, my, 1, 1, it.Id)
		tx := ModelsDiv.AddText(2, my, 1, 1, fmt.Sprintf("<i>%s</i>", source_llm.GetPricingString(it.Id)))
		tx.layout.Tooltip = "Price of Input_text/Input_image/Input_cached/Output per 1M tokens"
		my++
	}

	return nil
}
