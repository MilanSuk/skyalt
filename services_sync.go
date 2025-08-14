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
	"image/color"
	"os"
	"time"
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

	App_provider string
	App_smarter  bool
	App_model    string

	Code_provider string
	Code_smarter  bool
	Code_model    string

	Image_provider string
	Image_model    string

	STT_provider string
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
	LLM_groq    LLMGroq
	LLM_wsp     LLMWhispercpp
	LLM_llama   LLMLlamacpp

	last_dev_storage_change int64
}

func NewServicesSync(services *Services) (*ServicesSync, error) {
	snc := &ServicesSync{services: services, last_dev_storage_change: -1}

	snc._readOrInitFiles()

	return snc, nil
}

func (snc *ServicesSync) Destroy() {
}

func (snc *ServicesSync) _readOrInitFiles() error {
	var path string

	path = "apps/Device/DeviceSettings-DeviceSettings.json"
	devJs, err := os.ReadFile(path)
	if err != nil {
		//DPI
		snc.Device.Dpi = GetDeviceDPI()
		snc.Device.Dpi_default = GetDeviceDPI()

		//UI rounding
		snc.Device.Rounding = 0.2

		//Scroll
		snc.Device.ScrollThick = 0.5

		//Speaker
		snc.Device.Volume = 0.5

		//Theme
		snc.Device.Theme = "light"

		snc.Device.LightPalette = DevPalette{
			P:   color.RGBA{37, 100, 120, 255},
			OnP: color.RGBA{255, 255, 255, 255},

			S:   color.RGBA{170, 200, 170, 255},
			OnS: color.RGBA{255, 255, 255, 255},

			E:   color.RGBA{180, 40, 30, 255},
			OnE: color.RGBA{255, 255, 255, 255},

			B:   color.RGBA{250, 250, 250, 255},
			OnB: color.RGBA{25, 27, 30, 255},
		}

		snc.Device.DarkPalette = DevPalette{
			P:   color.RGBA{150, 205, 225, 255},
			OnP: color.RGBA{0, 50, 65, 255},

			S:   color.RGBA{190, 200, 205, 255},
			OnS: color.RGBA{40, 50, 55, 255},

			E:   color.RGBA{240, 185, 180, 255},
			OnE: color.RGBA{45, 45, 65, 255},

			B:   color.RGBA{25, 30, 30, 255},
			OnB: color.RGBA{230, 230, 230, 255},
		}

		snc.Device.CustomPalette = DevPalette{
			P:   color.RGBA{37, 100, 120, 255},
			OnP: color.RGBA{255, 255, 255, 255},

			S:   color.RGBA{170, 200, 170, 255},
			OnS: color.RGBA{255, 255, 255, 255},

			E:   color.RGBA{180, 40, 30, 255},
			OnE: color.RGBA{255, 255, 255, 255},

			B:   color.RGBA{250, 250, 250, 255},
			OnB: color.RGBA{25, 27, 30, 255},
		}

		//Date format
		{
			_, zn := time.Now().Zone()
			zn = zn / 3600
			if zn <= -3 && zn >= -10 {
				snc.Device.DateFormat = "us"
			} else {
				snc.Device.DateFormat = "eu"
			}
		}

		Tools_WriteJSONFile(path, &snc.Device)
	} else {
		LogsJsonUnmarshal(devJs, &snc.Device)
	}

	path = "apps/Device/MapSettings-MapSettings.json"
	mapJs, err := os.ReadFile(path)
	if err != nil {
		snc.Map.Enable = true
		snc.Map.Tiles_url = "https://tile.openstreetmap.org/{z}/{x}/{y}.png"
		snc.Map.Copyright = "(c)OpenStreetMap contributors"
		snc.Map.Copyright_url = "https://www.openstreetmap.org/copyright"

		Tools_WriteJSONFile(path, &snc.Map)
	} else {
		LogsJsonUnmarshal(mapJs, &snc.Map)
	}

	path = "apps/Device/MicrophoneSettings-MicrophoneSettings.json"
	micJs, err := os.ReadFile(path)
	if err != nil {
		snc.Mic.Enable = true
		snc.Mic.Sample_rate = 44100
		snc.Mic.Channels = 1

		Tools_WriteJSONFile(path, &snc.Mic)
	} else {
		LogsJsonUnmarshal(micJs, &snc.Mic)
	}

	path = "apps/Device/LLMxAI-LLMxAI.json"
	xaiJs, err := os.ReadFile(path)
	if err != nil {
		snc.LLM_xai.Provider = "xAI"
		snc.LLM_xai.OpenAI_url = "https://api.x.ai/v1"
		snc.LLM_xai.DevUrl = "https://console.x.ai"
		Tools_WriteJSONFile(path, &snc.LLM_xai)

	} else {
		LogsJsonUnmarshal(xaiJs, &snc.LLM_xai)
	}

	path = "apps/Device/LLMMistral-LLMMistral.json"
	mistralJs, err := os.ReadFile(path)
	if err != nil {
		snc.LLM_mistral.Provider = "Mistral"
		snc.LLM_mistral.OpenAI_url = "https://api.mistral.ai/v1"
		snc.LLM_mistral.DevUrl = "https://console.mistral.ai"

		Tools_WriteJSONFile(path, &snc.LLM_mistral)
	} else {
		LogsJsonUnmarshal(mistralJs, &snc.LLM_mistral)
	}

	path = "apps/Device/LLMOpenai-LLMOpenai.json"
	openaiJs, err := os.ReadFile(path)
	if err != nil {
		snc.LLM_openai.Provider = "OpenAI"
		snc.LLM_openai.OpenAI_url = "https://api.openai.com/v1"
		snc.LLM_openai.DevUrl = "https://platform.openai.com/"
		Tools_WriteJSONFile(path, &snc.LLM_openai)
	} else {
		LogsJsonUnmarshal(openaiJs, &snc.LLM_openai)
	}

	path = "apps/Device/LLMGroq-LLMGroq.json"
	groqJs, err := os.ReadFile(path)
	if err != nil {
		snc.LLM_groq.Provider = "Groq"
		snc.LLM_groq.OpenAI_url = "https://api.groq.com/openai/v1"
		snc.LLM_groq.DevUrl = "https://console.groq.com/keys"

		Tools_WriteJSONFile(path, &snc.LLM_groq)
	} else {
		LogsJsonUnmarshal(groqJs, &snc.LLM_groq)
	}

	wspJs, err := os.ReadFile("apps/Device/LLMWhispercpp-LLMWhispercpp.json")
	if err != nil {
		snc.LLM_wsp.Address = "http://localhost"
		snc.LLM_wsp.Port = 8090
		Tools_WriteJSONFile("apps/Device/LLMWhispercpp-LLMWhispercpp.json", &snc.Map)
	} else {
		LogsJsonUnmarshal(wspJs, &snc.LLM_wsp)
	}

	path = "apps/Device/LLMLlamacpp-LLMLlamacpp.json"
	llamas, err := os.ReadFile(path)
	if err != nil {
		snc.LLM_llama.Address = "http://localhost"
		snc.LLM_llama.Port = 8070
		Tools_WriteJSONFile(path, &snc.Map)
	} else {
		LogsJsonUnmarshal(llamas, &snc.LLM_llama)
	}

	return nil
}

func (snc *ServicesSync) Tick(devApp_storage_changes int64) bool {
	if snc.last_dev_storage_change != devApp_storage_changes {
		snc.last_dev_storage_change = devApp_storage_changes

		snc._readOrInitFiles()
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
