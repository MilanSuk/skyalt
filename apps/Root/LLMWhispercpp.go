package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type LLMWhispercppModel struct {
	Label string
	Path  string
}

// Whisper.cpp settings.
type LLMWhispercpp struct {
	lock sync.Mutex

	Address string
	Port    int
}

func NewLLMWhispercpp_wsp(file string) (*LLMWhispercpp, error) {
	st := &LLMWhispercpp{}

	st.Address = "http://localhost"
	st.Port = 8090

	return LoadFile(file, "LLMWhispercpp", "json", st, true)
}

func (wsp *LLMWhispercpp) Check() error {
	if wsp.Address == "" {
		return fmt.Errorf("Whisper address is empty")
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
	return filepath.Join("models", model_name+".bin")
}

func (wsp *LLMWhispercpp) GetUrlInference() string {
	return fmt.Sprintf("%s:%d/inference", wsp.Address, wsp.Port)
}
func (wsp *LLMWhispercpp) GetUrlLoadModel() string {
	return fmt.Sprintf("%s:%d/load", wsp.Address, wsp.Port)
}
