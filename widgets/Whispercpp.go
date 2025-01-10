package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Whispercpp struct {
	RunProcess bool
	Folder     string

	Addr  string
	Port  int
	Model string
}

func (layout *Layout) AddWhispercpp(x, y, w, h int, props *Whispercpp) *Whispercpp {
	layout._createDiv(x, y, w, h, "Whispercpp", props.Build, nil, nil)
	return props
}

var g_Whispercpp *Whispercpp

func OpenFile_Whispercpp() *Whispercpp {
	if g_Whispercpp == nil {
		g_Whispercpp = &Whispercpp{RunProcess: true, Folder: "services/whisper.cpp", Addr: "127.0.0.1", Port: 8091, Model: "ggml-tiny.en"}
		_read_file("Whispercpp-Whispercpp", g_Whispercpp)
	}
	return g_Whispercpp
}

func (st *Whispercpp) Build(layout *Layout) {

	if st.Folder == "" {
		st.Folder = "services/whisper.cpp"
	}
	if st.Addr == "" {
		st.Addr = "127.0.0.1"
	}
	if st.Port == 0 {
		st.Port = 8091
	}

	layout.SetColumn(0, 1, 4)
	layout.SetColumn(1, 1, 10)
	layout.SetColumn(2, 1, 2)

	y := 0

	layout.AddSwitch(1, y, 1, 1, "Run Whisper.cpp program from folder", &st.RunProcess)
	y++

	layout.AddText(0, y, 1, 1, "Folder")
	layout.AddFilePickerButton(1, y, 2, 1, &st.Folder, false)
	y++

	layout.AddText(0, y, 1, 1, "Address")
	layout.AddEditbox(1, y, 1, 1, &st.Addr)
	layout.AddEditbox(2, y, 1, 1, &st.Port)
	y++

	//models
	{
		model_labels, model_pathes := st.GetModelList()
		layout.AddText(0, y, 1, 1, "Default model")
		layout.AddCombo(1, y, 2, 1, &st.Model, model_labels, model_pathes)
		y++
	}

	//downloader
	{
		downDia := layout.AddDialog("download_whisper")
		downDia.Layout.SetColumn(0, 5, 15)
		downDia.Layout.SetRow(0, 5, 10)
		downDia.Layout.AddWhispercpp_downloader(0, 0, 1, 1)

		btDown, btDownLay := layout.AddButton2(1, y, 2, 1, "Download model(s)")
		btDown.clicked = func() {
			downDia.OpenRelative(btDownLay)
		}
		y++
	}
}

func (wsp *Whispercpp) GetUrlInference() string {
	return fmt.Sprintf("http://%s:%d/inference", wsp.Addr, wsp.Port)
}
func (wsp *Whispercpp) GetUrlLoadModel() string {
	return fmt.Sprintf("http://%s:%d/load", wsp.Addr, wsp.Port)
}

func (wsp *Whispercpp) GetModelList() ([]string, []string) {
	modelsPath := filepath.Join(wsp.Folder, "models")
	files, err := os.ReadDir(modelsPath)
	if err != nil {
		Layout_WriteError(err)
		return nil, nil
	}

	var model_labels []string
	var model_pathes []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		label, found := strings.CutPrefix(file.Name(), "ggml-")
		if !found {
			continue
		}
		label, found = strings.CutSuffix(label, ".bin")
		if !found {
			continue
		}
		path, found := strings.CutSuffix(file.Name(), ".bin")
		if !found {
			continue
		}

		model_labels = append(model_labels, label)
		model_pathes = append(model_pathes, path)
	}
	return model_labels, model_pathes
}
