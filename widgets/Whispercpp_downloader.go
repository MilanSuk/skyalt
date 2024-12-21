package main

import (
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/pbnjay/memory"
)

type Whispercpp_downloader struct {
}

func (layout *Layout) AddWhispercpp_downloader(x, y, w, h int) *Whispercpp_downloader {
	props := &Whispercpp_downloader{}
	layout._createDiv(x, y, w, h, "Whispercpp_downloader", props.Build, nil, nil)
	return props
}

func (st *Whispercpp_downloader) Build(layout *Layout) {

	max_ram := memory.TotalMemory()

	layout.SetColumn(0, 1, 100)
	layout.SetRow(0, 1, 100)

	Div := layout.AddLayout(0, 0, 1, 1)
	Div.SetColumn(0, 1, 4)
	Div.SetColumn(1, 2, 2)
	Div.SetColumn(2, 2, 2)
	Div.SetColumn(3, 4, 100)

	var model_names = []string{"ggml-tiny.en", "ggml-tiny", "ggml-base.en", "ggml-base", "ggml-small.en", "ggml-small", "ggml-medium.en", "ggml-medium", "ggml-large-v1", "ggml-large-v2", "ggml-large-v3"}
	var model_sizes = []int{75, 75, 142, 142, 466, 466, 1500, 1500, 2900, 2900, 2900}    //MB
	var model_rams = []int{390, 390, 500, 500, 1000, 1000, 2600, 2600, 4700, 4700, 4700} //MB

	for i := range model_names {
		Div.AddText(0, i, 1, 1, model_names[i])

		u, err := Whispercpp_GetUrl(model_names[i])
		if err != nil {
			continue
		}
		path := Whispercpp_GetPath(model_names[i])

		disk := Div.AddText(1, i, 1, 1, fmt.Sprintf("%dMB", model_sizes[i]))
		disk.Tooltip = "File size on disk"

		amount := float64(model_rams[i]*1024*1024) / float64(max_ram)
		ram := Div.AddText(2, i, 1, 1, fmt.Sprintf("%.0f%%", amount*100))
		ram.Tooltip = fmt.Sprintf("Percentage of total RAM: %dMB of %dMB", model_rams[i], int(max_ram/1024/1024))
		if amount > 0.9 {
			ram.Cd = Paint_GetPalette().E
		}

		Div.AddButtonDownload(3, i, 1, 1, path, u.String())
	}

	layout.AddText(0, 1, 1, 1, "<i>You can manually download .bin model and copy it into '<whisper.cpp>/models' folder.")
}

func Whispercpp_GetUrl(model string) (*url.URL, error) {
	u, err := url.Parse("https://huggingface.co/ggerganov/whisper.cpp/resolve/main/")
	if err != nil {
		return nil, err
	}
	u.Path = filepath.Join(u.Path, model+".bin")
	return u, nil
}
func Whispercpp_GetPath(model string) string {
	return filepath.Join(NewFile_Whispercpp().Folder, "models/"+model+".bin")
}
