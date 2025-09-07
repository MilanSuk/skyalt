package main

import (
	"fmt"
)

// Show LLMs settings.
type ShowLLMsCodeSettings struct {
}

func (st *ShowLLMsCodeSettings) run(caller *ToolCaller, ui *UI) error {
	source_dev, err := NewDeviceSettings("")
	if err != nil {
		return err
	}

	source_dev.UpdateModels()

	//Coding
	{
		ui.SetColumn(0, 1, Layout_MAX_SIZE)
		ui.SetRowFromSub(0, 1, Layout_MAX_SIZE, true)
		CodeDiv := ui.AddLayout(0, 0, 1, 1)
		CodeDiv.SetColumn(0, 1, 4)
		CodeDiv.SetColumn(1, 1, Layout_MAX_SIZE)

		tx := CodeDiv.AddText(0, 0, 2, 1, "Coding")
		tx.Align_h = 1

		CodeDiv.AddDropDown(0, 1, 1, 1, &source_dev.Code_provider, DeviceSettings_getAppProviders(), DeviceSettings_getAppProviders())
		source_dev.BuildProvider(CodeDiv, source_dev.Code_provider, caller)

		smarterSw := CodeDiv.AddSwitch(0, 2, 1, 1, "Smarter", &source_dev.Code_smarter)
		smarterSw.layout.Enable = (source_dev.Code_provider != "")

		pricing, tooltip := source_dev.GetPricingString(source_dev.Code_provider, source_dev.Code_model)
		mdl := CodeDiv.AddText(1, 2, 1, 1, source_dev.Code_model+fmt.Sprintf(" (<i>%s</i>)", pricing))
		mdl.layout.Tooltip = tooltip
	}

	return nil
}
