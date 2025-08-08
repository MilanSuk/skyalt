package main

import (
	"fmt"
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

	y := 0
	ui.SetColumn(0, 1, Layout_MAX_SIZE)

	//App
	{
		ui.SetRowFromSub(y, 1, Layout_MAX_SIZE, true)
		AppDiv := ui.AddLayout(0, y, 1, 1)
		AppDiv.SetColumn(0, 1, 4)
		AppDiv.SetColumn(1, 1, Layout_MAX_SIZE)

		tx := AppDiv.AddText(0, 0, 2, 1, "App")
		tx.Align_h = 1

		AppDiv.AddDropDown(0, 1, 1, 1, &source_dev.App_provider, DeviceSettings_getAppProviders(), DeviceSettings_getAppProviders())
		source_dev.BuildProvider(AppDiv, source_dev.App_provider, caller)

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
		ui.SetRowFromSub(y, 1, Layout_MAX_SIZE, true)
		ui.AddTool(0, y, 1, 1, "llm", (&ShowLLMsCodeSettings{}).run, caller)
	}
	y++
	ui.AddDivider(0, y, 1, 1, true)
	y++ //space

	//Image
	/*{
		ui.SetRowFromSub(y, 1, Layout_MAX_SIZE, true)
		ChatDiv := ui.AddLayout(0, y, 1, 1)
		ChatDiv.SetColumn(0, 1, 4)
		ChatDiv.SetColumn(1, 1, Layout_MAX_SIZE)

		tx := ChatDiv.AddText(0, 0, 2, 1, "Image generation")
		tx.Align_h = 1

		ChatDiv.AddDropDown(0, 1, 1, 1, &source_dev.Image_provider, DeviceSettings_getImageProviders(), DeviceSettings_getImageProviders())
		fnProvider(ChatDiv, source_dev.Image_provider)
	}
	y++
	ui.AddDivider(0, y, 1, 1, true)
	y++ //space
	*/

	//STT
	{
		ui.SetRowFromSub(y, 1, Layout_MAX_SIZE, true)
		ChatDiv := ui.AddLayout(0, y, 1, 1)
		ChatDiv.SetColumn(0, 1, 4)
		ChatDiv.SetColumn(1, 1, Layout_MAX_SIZE)

		tx := ChatDiv.AddText(0, 0, 2, 1, "Speech to text")
		tx.Align_h = 1

		ChatDiv.AddDropDown(0, 1, 1, 1, &source_dev.STT_provider, DeviceSettings_getSTTProviders(), DeviceSettings_getSTTProviders())
		source_dev.BuildProvider(ChatDiv, source_dev.STT_provider, caller)
	}
	y++
	//y++ //space

	//number of tries to fix error ....
	return nil
}
