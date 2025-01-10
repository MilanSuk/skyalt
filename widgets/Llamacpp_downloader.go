package main

import (
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/pbnjay/memory"
)

type Llamacpp_downloader struct {
}

func (layout *Layout) AddLlamacpp_downloader(x, y, w, h int) *Llamacpp_downloader {
	props := &Llamacpp_downloader{}
	layout._createDiv(x, y, w, h, "Llamacpp_downloader", props.Build, nil, nil)
	return props
}

func (st *Llamacpp_downloader) Build(layout *Layout) {

	max_ram := memory.TotalMemory()

	layout.SetColumn(0, 1, 100)
	layout.SetRow(0, 1, 100)

	Div := layout.AddLayout(0, 0, 1, 1)
	Div.SetColumn(0, 1, 10)
	Div.SetColumn(1, 2, 2)
	Div.SetColumn(2, 2, 2)
	Div.SetColumn(3, 4, 100)

	type ChatModel struct {
		Name string
		Url  string
		Size int //GB
		RAM  int //MB
	}
	models := []ChatModel{
		{Name: "tinyllama-1.1b-chat-v1.0.Q4_K_M", Size: 669, RAM: 3300, Url: "https://huggingface.co/TheBloke/TinyLlama-1.1B-Chat-v1.0-GGUF/resolve/main"},

		//RAM ...
		{Name: "Phi-3-mini-4k-instruct-q4", Size: 2390, RAM: 0, Url: "https://huggingface.co/microsoft/Phi-3-mini-4k-instruct-gguf/resolve/main"},

		{Name: "gemma-2-2b-it-Q4_K_M", Size: 1751, RAM: 0, Url: "https://huggingface.co/bartowski/gemma-2-2b-it-GGUF/resolve/main"},
		{Name: "gemma-2-27b-it-Q4_K_M", Size: 16650, RAM: 0, Url: "https://huggingface.co/bartowski/gemma-2-27b-it-GGUF/resolve/main"},

		{Name: "llama-3.2-1b-instruct-q8_0", Size: 1230, RAM: 0, Url: "https://huggingface.co/hugging-quants/Llama-3.2-1B-Instruct-Q8_0-GGUF/resolve/main"},
		{Name: "llama-3.2-3b-instruct-q4_k_m", Size: 3210, RAM: 0, Url: "https://huggingface.co/hugging-quants/Llama-3.2-3B-Instruct-Q4_K_M-GGUF/resolve/main"},

		{Name: "Meta-Llama-3.1-8B-Instruct-Q4_K_M", Size: 4900, RAM: 0, Url: "https://huggingface.co/bartowski/Meta-Llama-3.1-8B-Instruct-GGUF/resolve/main"},
		{Name: "Meta-Llama-3.1-70B-Instruct-Q4_K_M", Size: 42500, RAM: 0, Url: "https://huggingface.co/bartowski/Meta-Llama-3.1-70B-Instruct-GGUF/resolve/main"},
	}

	for i := range models {
		model := &models[i]

		Div.AddText(0, i, 1, 1, model.Name)

		u, err := Llamacpp_GetUrl(model.Url, model.Name)
		if err != nil {
			continue
		}
		path := Llamacpp_GetPath(model.Name)

		disk := Div.AddText(1, i, 1, 1, fmt.Sprintf("%dMB", model.Size))
		disk.Tooltip = "File size on disk"

		amount := float64(model.RAM*1024*1024) / float64(max_ram)
		ram := Div.AddText(2, i, 1, 1, fmt.Sprintf("%.0f%%", amount*100))
		ram.Tooltip = fmt.Sprintf("Percentage of total RAM: %dMB of %dMB", model.RAM, int(max_ram/1024/1024))
		if amount > 0.9 {
			ram.Cd = Paint_GetPalette().E
		}

		Div.AddButtonDownload(3, i, 1, 1, path, u.String())
	}

	layout.AddText(0, 1, 1, 1, "<i>You can manually download .bin model and copy it into '<llama.cpp>/models' folder.")
}

func Llamacpp_GetUrl(Url string, Name string) (*url.URL, error) {
	u, err := url.Parse(Url)
	if err != nil {
		return nil, err
	}
	u.Path = filepath.Join(u.Path, Name+".gguf")
	return u, nil
}

func Llamacpp_GetPath(model string) string {
	return filepath.Join(OpenFile_Llamacpp().Folder, "models/"+model+".bin")
}
