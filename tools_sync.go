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
	"image/color"
	"os"
)

type ToolsSyncDeviceSettingsMicrophone struct {
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

	Microphone ToolsSyncDeviceSettingsMicrophone
}

type ToolsSyncMicrophoneSettings struct {
	Enable      bool
	Sample_rate int
	Channels    int
}
type ToolsSyncMapSettings struct {
	Enable    bool
	Tiles_url string

	Copyright     string
	Copyright_url string
}

type ToolsSync struct {
	router *ToolsRouter

	Device ToolsSyncDeviceSettings
	Map    ToolsSyncMapSettings

	LLM_xai LLMxAI

	last_dev_storage_change int64
}

func NewToolsSync(router *ToolsRouter) (*ToolsSync, error) {
	snc := &ToolsSync{router: router, last_dev_storage_change: -1}

	//"pre-init"
	snc.Device = ToolsSyncDeviceSettings{
		Dpi:        GetDeviceDPI(),
		DateFormat: "us",
		Rounding:   0.2,
		Fullscreen: false,
		Stats:      false,
		Theme:      "light",
		LightPalette: DevPalette{
			P:   color.RGBA{37, 100, 120, 255},
			OnP: color.RGBA{255, 255, 255, 255},

			S:   color.RGBA{170, 200, 170, 255},
			OnS: color.RGBA{255, 255, 255, 255},

			E:   color.RGBA{180, 40, 30, 255},
			OnE: color.RGBA{255, 255, 255, 255},

			B:   color.RGBA{250, 250, 250, 255},
			OnB: color.RGBA{25, 27, 30, 255},
		},
		Microphone: ToolsSyncDeviceSettingsMicrophone{Enable: true, Sample_rate: 44100, Channels: 1},
	}

	snc.LLM_xai = InitLLMxAI()

	//"pre-init"
	snc.Map = ToolsSyncMapSettings{Enable: false}

	snc._loadFiles()

	return snc, nil
}

func (snc *ToolsSync) Destroy() {
}

func (snc *ToolsSync) _loadFiles() error {
	devJs, err := os.ReadFile("apps/Device/DeviceSettings-DeviceSettings.json")
	if err == nil {
		json.Unmarshal(devJs, &snc.Device) //err ....
	}

	mapJs, err := os.ReadFile("apps/Device/MapSettings-MapSettings.json")
	if err == nil {
		json.Unmarshal(mapJs, &snc.Map) //err ....
	}

	xaiJs, err := os.ReadFile("apps/Root/LLMxAI-LLMxAI.json") //move to apps/Device ....
	if err == nil {
		json.Unmarshal(xaiJs, &snc.LLM_xai) //err ....
	}

	return nil
}

func (snc *ToolsSync) Tick() bool {
	devApp := snc.router.FindApp("apps/Device")
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

func (snc *ToolsSync) Upload_deviceDefaultDPI() {
	type SetDPIDefault struct {
		DPI int
	}
	snc.router.CallAsync(0, "apps/Device", "SetDeviceDPIDefault", SetDPIDefault{DPI: GetDeviceDPI()}, nil, nil)
}

func (snc *ToolsSync) Upload_deviceDPI(new_dpi int) {
	//DPI
	type SetDPI struct {
		DPI int
	}
	snc.router.CallAsync(0, "apps/Device", "SetDeviceDPI", SetDPI{DPI: new_dpi}, nil, nil)
}

func (snc *ToolsSync) Upload_deviceStats(new_stat bool) {
	//Stats
	type SetStats struct {
		Show bool
	}
	snc.router.CallAsync(0, "apps/Device", "SetDeviceStats", SetStats{Show: new_stat}, nil, nil)
}

func (snc *ToolsSync) Upload_deviceFullscreen(new_fullscreen bool) {
	// Fullscreen
	type SetFullscreen struct {
		Enable bool
	}
	snc.router.CallAsync(0, "apps/Device", "SetDeviceFullscreen", SetFullscreen{Enable: new_fullscreen}, nil, nil)

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
func (snc *ToolsSync) IsMicrophoneEnabled() bool {
	return snc.Device.Microphone.Enable
}
