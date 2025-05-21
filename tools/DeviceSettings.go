package main

import (
	"image/color"
	"time"
)

type DeviceSettingsMicrophone struct {
	Enable      bool
	Sample_rate int
	Channels    int
}

type DeviceSettings struct {
	DateFormat string
	Rounding   float64
	Volume     float64

	Microphone DeviceSettingsMicrophone

	Dpi         int
	Dpi_default int

	Fullscreen bool
	Stats      bool

	Theme string //light, dark, custom

	LightPalette  DevPalette
	DarkPalette   DevPalette
	CustomPalette DevPalette
}

func NewDeviceSettings(file string, caller *ToolCaller) (*DeviceSettings, error) {
	st := &DeviceSettings{}

	//DPI
	st.Dpi = 100
	st.Dpi_default = 100

	//UI rounding
	st.Rounding = 0.2

	//Speaker
	st.Volume = 0.5

	//Theme
	st.Theme = "light"

	st.LightPalette = DevPalette{
		P:   color.RGBA{37, 100, 120, 255},
		OnP: color.RGBA{255, 255, 255, 255},

		S:   color.RGBA{170, 200, 170, 255},
		OnS: color.RGBA{255, 255, 255, 255},

		E:   color.RGBA{180, 40, 30, 255},
		OnE: color.RGBA{255, 255, 255, 255},

		B:   color.RGBA{250, 250, 250, 255},
		OnB: color.RGBA{25, 27, 30, 255},
	}

	st.DarkPalette = DevPalette{
		P:   color.RGBA{150, 205, 225, 255},
		OnP: color.RGBA{0, 50, 65, 255},

		S:   color.RGBA{190, 200, 205, 255},
		OnS: color.RGBA{40, 50, 55, 255},

		E:   color.RGBA{240, 185, 180, 255},
		OnE: color.RGBA{45, 45, 65, 255},

		B:   color.RGBA{25, 30, 30, 255},
		OnB: color.RGBA{230, 230, 230, 255},
	}

	st.CustomPalette = DevPalette{
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
			st.DateFormat = "us"
		} else {
			st.DateFormat = "eu"
		}
	}

	//Microphone
	st.Microphone = DeviceSettingsMicrophone{Enable: true, Sample_rate: 44100, Channels: 1}

	return _loadInstance(file, "DeviceSettings", "json", st, true, caller)
}

func (st *DeviceSettings) GetPalette() *DevPalette {
	switch st.Theme {
	case "light":
		return &st.LightPalette
	case "dark":
		return &st.DarkPalette
	}
	return &st.CustomPalette
}

func (st *DeviceSettings) SetDPI(dpi int) {
	//check range
	if dpi < 30 {
		dpi = 30
	}
	if dpi > 600 {
		dpi = 600
	}
	st.Dpi = dpi
}
