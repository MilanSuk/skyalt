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
	"os"
)

type ServicesSyncMicrophoneSettings struct {
	Enable      bool
	Sample_rate int
	Channels    int
}

type ServicesSyncDeviceSettings struct {
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
	Image_provider string
	STT_provider   string
}

type ServicesSyncMapSettings struct {
	Enable    bool
	Tiles_url string

	Copyright     string
	Copyright_url string
}

type ServicesSync struct {
	services *Services

	Device      ServicesSyncDeviceSettings
	Map         ServicesSyncMapSettings
	Mic         ServicesSyncMicrophoneSettings
	LLM_xai     LLMxAI
	LLM_mistral LLMMistral
	LLM_openai  LLMOpenai
	LLM_wsp     LLMWhispercpp
	LLM_llama   LLMLlamacpp

	last_dev_storage_change int64
}

func NewServicesSync(services *Services) (*ServicesSync, error) {
	snc := &ServicesSync{services: services, last_dev_storage_change: -1}

	snc._loadFiles()

	return snc, nil
}

func (snc *ServicesSync) Destroy() {
}

func (snc *ServicesSync) _loadFiles() error {
	devJs, err := os.ReadFile("apps/Device/DeviceSettings-DeviceSettings.json")
	if err == nil {
		LogsJsonUnmarshal(devJs, &snc.Device)
	}

	mapJs, err := os.ReadFile("apps/Device/MapSettings-MapSettings.json")
	if err == nil {
		LogsJsonUnmarshal(mapJs, &snc.Map)
	}

	micJs, err := os.ReadFile("apps/Device/MicrophoneSettings-MicrophoneSettings.json")
	if err == nil {
		LogsJsonUnmarshal(micJs, &snc.Mic)
	}

	xaiJs, err := os.ReadFile("apps/Device/LLMxAI-LLMxAI.json")
	if err == nil {
		LogsJsonUnmarshal(xaiJs, &snc.LLM_xai)
	}

	mistralJs, err := os.ReadFile("apps/Device/LLMMistral-LLMMistral.json")
	if err == nil {
		LogsJsonUnmarshal(mistralJs, &snc.LLM_mistral)
	}

	openailJs, err := os.ReadFile("apps/Device/LLMOpenai-LLMOpenai.json")
	if err == nil {
		LogsJsonUnmarshal(openailJs, &snc.LLM_openai)
	}

	wspJs, err := os.ReadFile("apps/Device/LLMWhispercpp-LLMWhispercpp.json")
	if err == nil {
		LogsJsonUnmarshal(wspJs, &snc.LLM_wsp)
	}

	llamas, err := os.ReadFile("apps/Device/LLMLlamacpp-LLMLlamacpp.json")
	if err == nil {
		LogsJsonUnmarshal(llamas, &snc.LLM_llama)
	}

	return nil
}

func (snc *ServicesSync) Tick(devApp_storage_changes int64) bool {
	if snc.last_dev_storage_change != devApp_storage_changes {
		snc.last_dev_storage_change = devApp_storage_changes

		snc._loadFiles()
		return true
	}
	return false
}

func (snc *ServicesSync) Upload_LoadFiles() {
	type SetDPIDefault struct {
		DPI int
	}
	snc.services.CallBuildAsync(0, "Device", "LoadFiles", SetDPIDefault{}, nil, nil)
}

func (snc *ServicesSync) Upload_deviceDefaultDPI() {
	type SetDPIDefault struct {
		DPI int
	}
	snc.services.CallBuildAsync(0, "Device", "SetDeviceDPIDefault", SetDPIDefault{DPI: GetDeviceDPI()}, nil, nil)
}

func (snc *ServicesSync) Upload_deviceDPI(new_dpi int) {
	//DPI
	type SetDPI struct {
		DPI int
	}
	snc.services.CallBuildAsync(0, "Device", "SetDeviceDPI", SetDPI{DPI: new_dpi}, nil, nil)
}

func (snc *ServicesSync) Upload_deviceStats(new_stat bool) {
	//Stats
	type SetStats struct {
		Show bool
	}
	snc.services.CallBuildAsync(0, "Device", "SetDeviceStats", SetStats{Show: new_stat}, nil, nil)
}

func (snc *ServicesSync) Upload_deviceFullscreen(new_fullscreen bool) {
	// Fullscreen
	type SetFullscreen struct {
		Enable bool
	}
	snc.services.CallBuildAsync(0, "Device", "SetDeviceFullscreen", SetFullscreen{Enable: new_fullscreen}, nil, nil)

}

func (snc *ServicesSync) GetPalette() *DevPalette {
	switch snc.Device.Theme {
	case "light":
		return &snc.Device.LightPalette
	case "dark":
		return &snc.Device.DarkPalette
	}
	return &snc.Device.CustomPalette

}
func (snc *ServicesSync) GetStats() bool {
	return snc.Device.Stats
}
func (snc *ServicesSync) GetDateFormat() string {
	return snc.Device.DateFormat
}
func (snc *ServicesSync) GetRounding() float64 {
	return snc.Device.Rounding
}
