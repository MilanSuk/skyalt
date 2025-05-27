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

type UiSyncDeviceSettingsMicrophone struct {
	Enable      bool
	Sample_rate int
	Channels    int
}

type UiSyncDeviceSettings struct {
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

	Microphone UiSyncDeviceSettingsMicrophone
}

type UiSyncMicrophoneSettings struct {
	Enable      bool
	Sample_rate int
	Channels    int
}
type UiSyncMapSettings struct {
	Enable    bool
	Tiles_url string

	Copyright     string
	Copyright_url string
}

type UiSync struct {
	Device UiSyncDeviceSettings
	Map    UiSyncMapSettings

	last_dev_storage_change int64
}

func NewUiSync(router *ToolsRouter) (*UiSync, error) {
	snc := &UiSync{}

	//"pre-init"
	snc.Device = UiSyncDeviceSettings{
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
		Microphone: UiSyncDeviceSettingsMicrophone{Enable: true, Sample_rate: 44100, Channels: 1},
	}

	//"pre-init"
	snc.Map = UiSyncMapSettings{Enable: false}

	//send default DPI
	{
		type SetDPIDefault struct {
			DPI int
		}
		router.CallAsync(0, "Root", "SetDeviceDPIDefault", SetDPIDefault{DPI: GetDeviceDPI()}, nil, nil)
	}

	return snc, nil
}

func (snc *UiSync) Destroy() {
}

func (snc *UiSync) Tick(router *ToolsRouter) bool {
	devApp := router.FindApp("Root")
	if devApp != nil {
		if devApp.storage_changes != snc.last_dev_storage_change {

			devApp.storage_changes = snc.last_dev_storage_change

			devJs, err := os.ReadFile("apps/Root/DeviceSettings-DeviceSettings.json")
			if err == nil {
				json.Unmarshal(devJs, &snc.Device) //err ....
			}

			mapJs, err := os.ReadFile("apps/Root/MapSettings-MapSettings.json")
			if err == nil {
				json.Unmarshal(mapJs, &snc.Map) //err ....
			}

			router.CallUpdateDev()

			return true
		}
	}

	return false
}

func (snc *UiSync) Upload_deviceDPI(new_dpi int, router *ToolsRouter) {
	//DPI
	type SetDPI struct {
		DPI int
	}
	router.CallAsync(0, "Root", "SetDeviceDPI", SetDPI{DPI: new_dpi}, nil, nil)
}

func (snc *UiSync) Upload_deviceStats(new_stat bool, router *ToolsRouter) {
	//Stats
	type SetStats struct {
		Show bool
	}
	router.CallAsync(0, "Root", "SetDeviceStats", SetStats{Show: new_stat}, nil, nil)
}

func (snc *UiSync) Upload_deviceFullscreen(new_fullscreen bool, router *ToolsRouter) {
	// Fullscreen
	type SetFullscreen struct {
		Enable bool
	}
	router.CallAsync(0, "Root", "SetDeviceFullscreen", SetFullscreen{Enable: new_fullscreen}, nil, nil)

}

func (snc *UiSync) GetPalette() *DevPalette {
	switch snc.Device.Theme {
	case "light":
		return &snc.Device.LightPalette
	case "dark":
		return &snc.Device.DarkPalette
	}
	return &snc.Device.CustomPalette

}
func (snc *UiSync) GetStats() bool {
	return snc.Device.Stats
}
func (snc *UiSync) GetDateFormat() string {
	return snc.Device.DateFormat
}
func (snc *UiSync) GetRounding() float64 {
	return snc.Device.Rounding
}
func (snc *UiSync) IsMicrophoneEnabled() bool {
	return snc.Device.Microphone.Enable
}
