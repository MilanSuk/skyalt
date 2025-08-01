package main

import (
	"fmt"
	"strings"
)

// Show LLMs settings.
type ShowLLMsSettings struct {
}

func (st *ShowLLMsSettings) run(caller *ToolCaller, ui *UI) error {
	source_dev, err := NewDeviceSettings("")
	if err != nil {
		return err
	}

	source_dev.UpdateModels()

	appProviders := []string{"", "xAI", "Mistral", "OpenAI", "Groq", "Llama.cpp"}
	//imageProviders := []string{"", "xAI", "OpenAI"}
	sttProviders := []string{"", "Whisper.cpp", "OpenAI"}

	fnProvider := func(ChatDiv *UI, provider string) {
		ChatDia := ChatDiv.AddDialog(provider + "_settings")
		ChatDia.UI.SetColumn(0, 1, 20)
		ChatDia.UI.SetRowFromSub(0, 1, 100, true)
		found := true
		switch strings.ToLower(provider) {
		case "xai":
			ChatDia.UI.AddTool(0, 0, 1, 1, (&ShowLLMxAISettings{}).run, caller)
		case "mistral":
			ChatDia.UI.AddTool(0, 0, 1, 1, (&ShowLLMMistralSettings{}).run, caller)
		case "groq":
			ChatDia.UI.AddTool(0, 0, 1, 1, (&ShowLLMGroqSettings{}).run, caller)
		case "openai":
			ChatDia.UI.AddTool(0, 0, 1, 1, (&ShowLLMOpenAISettings{}).run, caller)
		case "llama.cpp":
			ChatDia.UI.AddTool(0, 0, 1, 1, (&ShowLLMLlamacppSettings{}).run, caller)
		case "whisper.cpp":
			ChatDia.UI.AddTool(0, 0, 1, 1, (&ShowLLMWhispercppSettings{}).run, caller)
		default:
			found = false
		}

		if found {
			ChatProvider := ChatDiv.AddButton(1, 1, 1, 1, provider+" Settings")
			ChatProvider.Background = 0.5

			providerErr := source_dev.CheckProvider(provider)
			if providerErr != nil {
				ChatProvider.Cd = UI_GetPalette().E
				ChatProvider.layout.Tooltip = providerErr.Error()
			}
			ChatProvider.clicked = func() error {
				ChatDia.OpenCentered(caller)
				return nil
			}
		} else {
			errTx := ChatDiv.AddText(1, 1, 1, 1, "Disabled")
			errTx.Cd = UI_GetPalette().E
		}
	}

	y := 0
	ui.SetColumn(0, 1, 100)

	//App
	{
		ui.SetRowFromSub(y, 1, 100, true)
		AppDiv := ui.AddLayout(0, y, 1, 1)
		AppDiv.SetColumn(0, 1, 4)
		AppDiv.SetColumn(1, 1, 100)

		tx := AppDiv.AddText(0, 0, 2, 1, "App")
		tx.Align_h = 1

		AppDiv.AddDropDown(0, 1, 1, 1, &source_dev.App_provider, appProviders, appProviders)
		fnProvider(AppDiv, source_dev.App_provider)

		smarterSw := AppDiv.AddSwitch(0, 2, 1, 1, "Smarter", &source_dev.App_smarter)
		smarterSw.layout.Enable = (source_dev.App_provider != "")

		mdl := AppDiv.AddText(1, 2, 1, 1, source_dev.App_model+fmt.Sprintf(" (<i>%s</i>)", source_dev.GetPricingString(source_dev.App_provider, source_dev.App_model)))
		mdl.layout.Tooltip = DeviceSettings_GetPricingStringTooltip()
	}
	y++
	ui.AddDivider(0, y, 1, 1, true)
	y++ //space

	//Coding
	{
		ui.SetRowFromSub(y, 1, 100, true)
		CodeDiv := ui.AddLayout(0, y, 1, 1)
		CodeDiv.SetColumn(0, 1, 4)
		CodeDiv.SetColumn(1, 1, 100)

		tx := CodeDiv.AddText(0, 0, 2, 1, "Coding")
		tx.Align_h = 1

		CodeDiv.AddDropDown(0, 1, 1, 1, &source_dev.Code_provider, appProviders, appProviders)
		fnProvider(CodeDiv, source_dev.Code_provider)

		smarterSw := CodeDiv.AddSwitch(0, 2, 1, 1, "Smarter", &source_dev.Code_smarter)
		smarterSw.layout.Enable = (source_dev.Code_provider != "")

		mdl := CodeDiv.AddText(1, 2, 1, 1, source_dev.Code_model+fmt.Sprintf(" (<i>%s</i>)", source_dev.GetPricingString(source_dev.Code_provider, source_dev.Code_model)))
		mdl.layout.Tooltip = DeviceSettings_GetPricingStringTooltip()
	}
	y++
	ui.AddDivider(0, y, 1, 1, true)
	y++ //space

	//Image
	/*{
		ui.SetRowFromSub(y, 1, 100, true)
		ChatDiv := ui.AddLayout(0, y, 1, 1)
		ChatDiv.SetColumn(0, 1, 4)
		ChatDiv.SetColumn(1, 1, 100)

		tx := ChatDiv.AddText(0, 0, 2, 1, "Image generation")
		tx.Align_h = 1

		ChatDiv.AddDropDown(0, 1, 1, 1, &source_dev.Image_provider, imageProviders, imageProviders)
		fnProvider(ChatDiv, source_dev.Image_provider)
	}
	y++
	ui.AddDivider(0, y, 1, 1, true)
	y++ //space
	*/

	//STT
	{
		ui.SetRowFromSub(y, 1, 100, true)
		ChatDiv := ui.AddLayout(0, y, 1, 1)
		ChatDiv.SetColumn(0, 1, 4)
		ChatDiv.SetColumn(1, 1, 100)

		tx := ChatDiv.AddText(0, 0, 2, 1, "Speech to text")
		tx.Align_h = 1

		ChatDiv.AddDropDown(0, 1, 1, 1, &source_dev.STT_provider, sttProviders, sttProviders)
		fnProvider(ChatDiv, source_dev.STT_provider)
	}
	y++
	//y++ //space

	//number of tries to fix error ....
	return nil
}
