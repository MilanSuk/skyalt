/*
Copyright 2025 Milan Suk

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
	"encoding/json"
	"os"
)

type ToolsSyncMicrophoneSettings struct {
	Enable      bool
	Sample_rate int
	Channels    int
}

type ToolsSyncDeviceSettings struct {
	DateFormat  string
	Rounding    float64
	ScrollThick float64
	Volume      float64
	Dpi         int
	Dpi_default int

	Fullscreen bool
	Stats      bool

	Theme         string //light, dark, custom
	LightPalette  DevPalette
	DarkPalette   DevPalette
	CustomPalette DevPalette

	Chat_provider  string
	Chat_smarter   bool
	Chat_faster    bool
	Image_provider string
	STT_provider   string
}

type ToolsSyncMapSettings struct {
	Enable    bool
	Tiles_url string

	Copyright     string
	Copyright_url string
}

type ToolsSync struct {
	router *ToolsRouter

	Device      ToolsSyncDeviceSettings
	Map         ToolsSyncMapSettings
	Mic         ToolsSyncMicrophoneSettings
	LLM_xai     LLMxAI
	LLM_mistral LLMMistral
	LLM_openai  LLMOpenai
	LLM_wsp     LLMWhispercpp
	LLM_llama   LLMLlamacpp

	last_dev_storage_change int64
}

func NewToolsSync(router *ToolsRouter) (*ToolsSync, error) {
	snc := &ToolsSync{router: router, last_dev_storage_change: -1}

	snc._loadFiles()

	return snc, nil
}

func (snc *ToolsSync) Destroy() {
}

func (snc *ToolsSync) _loadFiles() error {
	devJs, err := os.ReadFile("apps/Device/DeviceSettings-DeviceSettings.json")
	if err == nil {
		err = json.Unmarshal(devJs, &snc.Device)
		snc.router.log.Error(err)
	}

	mapJs, err := os.ReadFile("apps/Device/MapSettings-MapSettings.json")
	if err == nil {
		err = json.Unmarshal(mapJs, &snc.Map)
		snc.router.log.Error(err)
	}

	micJs, err := os.ReadFile("apps/Device/MicrophoneSettings-MicrophoneSettings.json")
	if err == nil {
		err = json.Unmarshal(micJs, &snc.Mic)
		snc.router.log.Error(err)
	}

	xaiJs, err := os.ReadFile("apps/Device/LLMxAI-LLMxAI.json")
	if err == nil {
		err = json.Unmarshal(xaiJs, &snc.LLM_xai)
		snc.router.log.Error(err)
	}

	mistralJs, err := os.ReadFile("apps/Device/LLMMistral-LLMMistral.json")
	if err == nil {
		err = json.Unmarshal(mistralJs, &snc.LLM_mistral)
		snc.router.log.Error(err)
	}

	openailJs, err := os.ReadFile("apps/Device/LLMOpenai-LLMOpenai.json")
	if err == nil {
		err = json.Unmarshal(openailJs, &snc.LLM_openai)
		snc.router.log.Error(err)
	}

	wspJs, err := os.ReadFile("apps/Device/LLMWhispercpp-LLMWhispercpp.json")
	if err == nil {
		err = json.Unmarshal(wspJs, &snc.LLM_wsp)
		snc.router.log.Error(err)
	}

	llamas, err := os.ReadFile("apps/Device/LLMLlamacpp-LLMLlamacpp.json")
	if err == nil {
		err = json.Unmarshal(llamas, &snc.LLM_llama)
		snc.router.log.Error(err)
	}

	return nil
}

func (snc *ToolsSync) Tick() bool {
	devApp := snc.router.FindApp("Device")
	if devApp != nil {
		if devApp.storage_changes != snc.last_dev_storage_change {

			devApp.storage_changes = snc.last_dev_storage_change

			snc._loadFiles()

			snc.router.CallUpdateDev()

			return true
		}
	}

	return false
}

func (snc *ToolsSync) Upload_LoadFiles() {
	type SetDPIDefault struct {
		DPI int
	}
	snc.router.CallBuildAsync(0, "Device", "LoadFiles", SetDPIDefault{}, nil, nil)
}

func (snc *ToolsSync) Upload_deviceDefaultDPI() {
	type SetDPIDefault struct {
		DPI int
	}
	snc.router.CallBuildAsync(0, "Device", "SetDeviceDPIDefault", SetDPIDefault{DPI: GetDeviceDPI()}, nil, nil)
}

func (snc *ToolsSync) Upload_deviceDPI(new_dpi int) {
	//DPI
	type SetDPI struct {
		DPI int
	}
	snc.router.CallBuildAsync(0, "Device", "SetDeviceDPI", SetDPI{DPI: new_dpi}, nil, nil)
}

func (snc *ToolsSync) Upload_deviceStats(new_stat bool) {
	//Stats
	type SetStats struct {
		Show bool
	}
	snc.router.CallBuildAsync(0, "Device", "SetDeviceStats", SetStats{Show: new_stat}, nil, nil)
}

func (snc *ToolsSync) Upload_deviceFullscreen(new_fullscreen bool) {
	// Fullscreen
	type SetFullscreen struct {
		Enable bool
	}
	snc.router.CallBuildAsync(0, "Device", "SetDeviceFullscreen", SetFullscreen{Enable: new_fullscreen}, nil, nil)

}

func (snc *ToolsSync) GetPalette() *DevPalette {
	switch snc.Device.Theme {
	case "light":
		return &snc.Device.LightPalette
	case "dark":
		return &snc.Device.DarkPalette
	}
	return &snc.Device.CustomPalette

}
func (snc *ToolsSync) GetStats() bool {
	return snc.Device.Stats
}
func (snc *ToolsSync) GetDateFormat() string {
	return snc.Device.DateFormat
}
func (snc *ToolsSync) GetRounding() float64 {
	return snc.Device.Rounding
}
