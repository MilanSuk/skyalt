package main

import "strings"

// Show LLMs settings.
type ShowLLMsSettings struct {
}

func (st *ShowLLMsSettings) run(caller *ToolCaller, ui *UI) error {
	source_dev, err := NewDeviceSettings("")
	if err != nil {
		return err
	}

	chatProviders := []string{"", "xAI", "Mistral", "OpenAI", "Llama.cpp"}
	imageProviders := []string{"", "xAI", "OpenAI"}
	sttProviders := []string{"", "Whisper.cpp", "OpenAI"}

	fnProvider := func(ChatDiv *UI, provider string) {
		ChatDia := ChatDiv.AddDialog(provider + "_settings")
		ChatDia.UI.SetColumn(0, 1, 20)
		ChatDia.UI.SetRowFromSub(0, 1, 100)
		found := true
		switch strings.ToLower(provider) {
		case "xai":
			ChatDia.UI.AddTool(0, 0, 1, 1, (&ShowLLMxAISettings{}).run, caller)
		case "mistral":
			ChatDia.UI.AddTool(0, 0, 1, 1, (&ShowLLMMistralSettings{}).run, caller)
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

	//Chat
	{
		ui.SetRowFromSub(y, 1, 100)
		ChatDiv := ui.AddLayout(0, y, 1, 1)
		ChatDiv.SetColumn(0, 1, 4)
		ChatDiv.SetColumn(1, 1, 100)

		tx := ChatDiv.AddText(0, 0, 2, 1, "Chat")
		tx.Align_h = 1

		ChatDiv.AddCombo(0, 1, 1, 1, &source_dev.Chat_provider, chatProviders, chatProviders)
		fnProvider(ChatDiv, source_dev.Chat_provider)

		smarterSw := ChatDiv.AddSwitch(0, 2, 2, 1, "Smarter", &source_dev.Chat_smarter)
		smarterSw.layout.Enable = (source_dev.Chat_provider != "")
	}
	y++
	ui.AddDivider(0, y, 1, 1, true)
	y++ //space

	//Image
	{
		ui.SetRowFromSub(y, 1, 100)
		ChatDiv := ui.AddLayout(0, y, 1, 1)
		ChatDiv.SetColumn(0, 1, 4)
		ChatDiv.SetColumn(1, 1, 100)

		tx := ChatDiv.AddText(0, 0, 2, 1, "Image generation")
		tx.Align_h = 1

		ChatDiv.AddCombo(0, 1, 1, 1, &source_dev.Image_provider, imageProviders, imageProviders)
		fnProvider(ChatDiv, source_dev.Image_provider)
	}
	y++
	ui.AddDivider(0, y, 1, 1, true)
	y++ //space

	//STT
	{
		ui.SetRowFromSub(y, 1, 100)
		ChatDiv := ui.AddLayout(0, y, 1, 1)
		ChatDiv.SetColumn(0, 1, 4)
		ChatDiv.SetColumn(1, 1, 100)

		tx := ChatDiv.AddText(0, 0, 2, 1, "Speech to text")
		tx.Align_h = 1

		ChatDiv.AddCombo(0, 1, 1, 1, &source_dev.STT_provider, sttProviders, sttProviders)
		fnProvider(ChatDiv, source_dev.STT_provider)
	}
	y++
	//y++ //space

	//number of tries to fix error ....
	return nil
}
