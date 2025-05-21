package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type LLMWhispercppModel struct {
	Label string
	Path  string
}
type LLMWhispercpp struct {
	lock sync.Mutex

	Folder  string
	Address string
	Port    int

	Models []LLMWhispercppModel

	ModelName string
}

func NewLLMWhispercpp_wsp(file string, caller *ToolCaller) (*LLMWhispercpp, error) {
	st := &LLMWhispercpp{}

	st.Folder = "services/whisper.cpp"
	st.Address = "http://localhost"
	st.Port = 8091

	return _loadInstance(file, "LLMWhispercpp", "json", st, true, caller)
}

func (wsp *LLMWhispercpp) Check() error {
	if wsp.Address == "" {
		return fmt.Errorf("Whisper address is empty")

	}
	if wsp.Folder == "" {
		return fmt.Errorf("Whisper folder is empty")

	}
	if !wsp.IsFolderExists(wsp.Folder) {
		return fmt.Errorf("Whisper folder not found")
	}

	err := wsp.ReloadModels()
	if err != nil {
		return err
	}

	return nil
}
func (wsp *LLMWhispercpp) IsFolderExists(fileName string) bool {
	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func (wsp *LLMWhispercpp) getModelPath(model_name string) string {
	return filepath.Join("models/", model_name+".bin")
}

func (wsp *LLMWhispercpp) GetUrlInference() string {
	return fmt.Sprintf("%s:%d/inference", wsp.Address, wsp.Port)
}
func (wsp *LLMWhispercpp) GetUrlLoadModel() string {
	return fmt.Sprintf("%s:%d/load", wsp.Address, wsp.Port)
}

func (wsp *LLMWhispercpp) ReloadModels() error {

	wsp.Models = nil //reset

	modelsFolder := filepath.Join(wsp.Folder, "models")
	files, err := os.ReadDir(modelsFolder)
	if err != nil {
		return err
	}

	last_name := ""
	found_path := false
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

		wsp.Models = append(wsp.Models, LLMWhispercppModel{Label: label, Path: path})

		if wsp.ModelName == file.Name() {
			found_path = true
		}
		last_name = file.Name()
	}

	if !found_path {
		wsp.ModelName = ""
	}

	if wsp.ModelName == "" {
		wsp.ModelName = last_name
	}

	return nil
}
