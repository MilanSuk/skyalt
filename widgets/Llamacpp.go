package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Llamacpp struct {
	RunProcess bool
	Folder     string
	Addr       string
	Port       int
	Model      string
}

func (layout *Layout) AddLlamacpp(x, y, w, h int, props *Llamacpp) *Llamacpp {
	layout._createDiv(x, y, w, h, "Llamacpp", props.Build, nil, nil)
	return props
}

var g_Llamacpp *Llamacpp

func OpenFile_Llamacpp() *Llamacpp {
	if g_Llamacpp == nil {
		g_Llamacpp = &Llamacpp{RunProcess: true, Folder: "services/llama.cpp", Addr: "127.0.0.1", Port: 8090, Model: ""}
		_read_file("Llamacpp-Llamacpp", g_Llamacpp)
	}
	return g_Llamacpp
}

func (st *Llamacpp) Build(layout *Layout) {

	if st.Folder == "" {
		st.Folder = "services/llama.cpp"
	}
	if st.Addr == "" {
		st.Addr = "127.0.0.1"
	}
	if st.Port == 0 {
		st.Port = 8090
	}

	layout.SetColumn(0, 1, 4)
	layout.SetColumn(1, 1, 10)
	layout.SetColumn(2, 1, 2)

	y := 0

	layout.AddSwitch(1, y, 1, 1, "Run Llama.cpp program from folder", &st.RunProcess)
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
		models := st.GetModelList()
		layout.AddText(0, y, 1, 1, "Default model")
		layout.AddCombo(1, y, 2, 1, &st.Model, models, models)
		y++
	}

	//downloader
	{
		downDia := layout.AddDialog("download_llama")
		downDia.Layout.SetColumn(0, 5, 25)
		downDia.Layout.SetRow(0, 5, 10)
		downDia.Layout.AddLlamacpp_downloader(0, 0, 1, 1)

		btDown, btDownLay := layout.AddButton2(1, y, 2, 1, "Download model(s)")
		btDown.clicked = func() {
			downDia.OpenRelative(btDownLay)
		}
		y++
	}
}

func (lma *Llamacpp) GetUrCompletion() string {
	return fmt.Sprintf("http://%s:%d/v1/chat/completions", lma.Addr, lma.Port)
}
func (lma *Llamacpp) GetUrlHealth() string {
	return fmt.Sprintf("http://%s:%d/health", lma.Addr, lma.Port)
}

func (lma *Llamacpp) GetModelList() []string {
	modelsPath := filepath.Join(lma.Folder, "models")
	files, err := os.ReadDir(modelsPath)
	if err != nil {
		Layout_WriteError(err)
		return nil
	}

	var models []string

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if strings.HasPrefix(file.Name(), "ggml-") {
			continue
		}

		name, found := strings.CutSuffix(file.Name(), ".gguf")
		if !found {
			continue
		}
		models = append(models, name)
	}

	return models
}
