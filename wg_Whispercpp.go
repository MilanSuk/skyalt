package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Whispercpp struct {
	Folder string

	Address string
	Port    int
	Model   string
}

func (layout *Layout) AddWhispercpp(x, y, w, h int, props *Whispercpp) *Whispercpp {
	layout._createDiv(x, y, w, h, "Whispercpp", props.Build, nil, nil)
	return props
}

func (st *Whispercpp) Build(layout *Layout) {

	if st.Folder == "" {
		st.Folder = "services/whisper.cpp"
	}
	if st.Address == "" {
		st.Address = "http://127.0.0.1"
	}
	if st.Port == 0 {
		st.Port = 8091
	}

	layout.SetColumn(0, 1, 4)
	layout.SetColumn(1, 1, 100)

	y := 0

	layout.AddText(0, y, 1, 1, "Folder")
	layout.AddFilePickerButton(1, y, 1, 1, &st.Folder, false)
	y++

	layout.AddText(0, y, 1, 1, "Address")
	AddrDiv := layout.AddLayout(1, y, 1, 1)
	{
		AddrDiv.SetColumn(0, 1, 100)
		AddrDiv.SetColumn(1, 1, 4)
		AddrDiv.SetColumn(2, 1, 4)
		AddrDiv.SetColumn(3, 0, 4)

		AddrDiv.AddEditbox(0, 0, 1, 1, &st.Address)
		AddrDiv.AddEditbox(1, 0, 1, 1, &st.Port)

		TestOKDia := layout.AddDialog("test_ok")
		TestOKDia.Layout.SetColumn(0, 1, 5)
		tx := TestOKDia.Layout.AddText(0, 0, 1, 1, "Success")
		tx.Align_h = 1

		TestErrDia := layout.AddDialog("test_err")
		TestErrDia.Layout.SetColumn(0, 1, 5)
		tx = TestErrDia.Layout.AddText(0, 0, 1, 1, "Error")
		tx.Align_h = 1
		tx.Cd = Paint_GetPalette().E

		TestBt := AddrDiv.AddButton(2, 0, 1, 1, "Test")
		TestBt.clicked = func() {
			status, err := Whispercpp_SetModel(strconv.Itoa(int(time.Now().Unix())), st)
			if err == nil && status == 200 {
				TestOKDia.OpenCentered()
			} else {
				TestErrDia.OpenCentered()
			}
			Layout_RefreshDelayed()
		}

	}
	y++

	model_labels, model_pathes := st.GetModelList()

	//default_model
	layout.AddText(0, y, 1, 1, "Default model")
	layout.AddCombo(1, y, 1, 1, &st.Model, model_labels, model_pathes)
	y++

	//model list
	{
		layout.AddText(0, y, 1, 1, "Models")
		for _, md := range model_labels {
			layout.AddText(1, y, 1, 1, md)
			y++
		}
	}
}

func (wsp *Whispercpp) GetUrlInference() string {
	return fmt.Sprintf("%s:%d/inference", wsp.Address, wsp.Port)
}
func (wsp *Whispercpp) GetUrlLoadModel() string {
	return fmt.Sprintf("%s:%d/load", wsp.Address, wsp.Port)
}

func (wsp *Whispercpp) GetModelList() (model_labels []string, model_pathes []string) {
	modelsPath := filepath.Join(wsp.Folder, "models")
	files, err := os.ReadDir(modelsPath)
	if err != nil {
		//...
	}

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
