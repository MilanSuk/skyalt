package main

import (
	"fmt"
	"path/filepath"
	"sync"
)

// Whisper.cpp settings.
type LLMWhispercpp struct {
	lock sync.Mutex

	Address string
	Port    int
}

func NewLLMWhispercpp(file string) (*LLMWhispercpp, error) {
	st := &LLMWhispercpp{}

	st.Address = "http://localhost"
	st.Port = 8090

	return LoadFile(file, "LLMWhispercpp", "json", st, true)
}

func (wsp *LLMWhispercpp) Check() error {
	if wsp.Address == "" {
		return fmt.Errorf("whisper.cpp address is empty")
	}

	return nil
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
