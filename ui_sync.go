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
	"fmt"
	"image/color"
)

type UiSyncDeviceSettingsMicrophone struct {
	Enable      bool
	Sample_rate int
	Channels    int
}

type UiSyncDeviceSettings struct {
	DateFormat  string
	Rounding    float64
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
	device_settings UiSyncDeviceSettings
	map_settings    UiSyncMapSettings

	DeviceSettings_path       string
	DeviceSettings_time_stamp int64

	MapSettings_path       string
	MapSettings_time_stamp int64
}

func NewUiSync(router *ToolsRouter) (*UiSync, error) {
	snc := &UiSync{}

	snc.DeviceSettings_path = "DeviceSettings-DeviceSettings.json"
	snc.MapSettings_path = "MapSettings-MapSettings.json"

	//"pre-init"
	snc.device_settings = UiSyncDeviceSettings{
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
	snc.map_settings = UiSyncMapSettings{Enable: false}

	//send default DPI
	{
		type SetDPIDefault struct {
			DPI int
		}
		_, err := router.CallAsync(0, "SetDeviceDPIDefault", SetDPIDefault{DPI: GetDeviceDPI()}, nil, nil)
		if err != nil {
			return nil, err
		}
	}

	return snc, nil
}

func (snc *UiSync) Destroy() {
}

func (snc *UiSync) NeedRefresh(router *ToolsRouter) bool {

	changed := false

	//Device
	{
		dev, _ := router.files.GetFile(snc.DeviceSettings_path)
		if dev != nil {
			if dev.time_stamp != snc.DeviceSettings_time_stamp {
				snc.DeviceSettings_time_stamp = dev.time_stamp
				changed = true
			}
		}
	}

	//Maps
	{
		mp, _ := router.files.GetFile(snc.MapSettings_path)
		if mp != nil {
			if mp.time_stamp != snc.MapSettings_time_stamp {
				snc.MapSettings_time_stamp = mp.time_stamp
				changed = true
			}
		}
	}

	return changed
}

func (snc *UiSync) ReloadSettings(router *ToolsRouter) error {

	//Device
	{
		dev, dev_data := router.files.GetFile(snc.DeviceSettings_path)
		if dev == nil {
			return fmt.Errorf("'%s' not found", snc.DeviceSettings_path)
		}
		err := json.Unmarshal(dev_data, &snc.device_settings)
		if err != nil {
			return err
		}
	}

	//Maps
	{
		mps, mps_data := router.files.GetFile(snc.MapSettings_path)
		if mps == nil {
			return fmt.Errorf("'%s' not found", snc.MapSettings_path)
		}
		err := json.Unmarshal(mps_data, &snc.map_settings)
		if err != nil {
			return err
		}
	}

	snc.device_settings.Dpi = OsClamp(snc.device_settings.Dpi, 30, 600)

	return nil
}

func (snc *UiSync) Upload_deviceDPI(new_dpi int, router *ToolsRouter) {
	//DPI
	type SetDPI struct {
		DPI int
	}
	router.CallAsync(0, "SetDeviceDPI", SetDPI{DPI: new_dpi}, nil, nil)
}

func (snc *UiSync) Upload_deviceStats(new_stat bool, router *ToolsRouter) {
	//Stats
	type SetStats struct {
		Show bool
	}
	router.CallAsync(0, "SetDeviceStats", SetStats{Show: new_stat}, nil, nil)
}

func (snc *UiSync) Upload_deviceFullscreen(new_fullscreen bool, router *ToolsRouter) {
	// Fullscreen
	type SetFullscreen struct {
		Enable bool
	}
	router.CallAsync(0, "SetDeviceFullscreen", SetFullscreen{Enable: new_fullscreen}, nil, nil)

}

func (snc *UiSync) GetPalette() *DevPalette {

	switch snc.device_settings.Theme {
	case "light":
		return &snc.device_settings.LightPalette
	case "dark":
		return &snc.device_settings.DarkPalette
	}
	return &snc.device_settings.CustomPalette

}
func (snc *UiSync) GetStats() bool {
	return snc.device_settings.Stats
}
func (snc *UiSync) GetDateFormat() string {
	return snc.device_settings.DateFormat
}
func (snc *UiSync) GetRounding() float64 {
	return snc.device_settings.Rounding
}
func (snc *UiSync) IsMicrophoneEnabled() bool {
	return snc.device_settings.Microphone.Enable
}
