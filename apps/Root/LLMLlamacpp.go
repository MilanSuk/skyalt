package main

import (
	"fmt"
	"sync"
)

type LLMLlamacppMsgStats struct {
	Function       string
	CreatedTimeSec float64
	Model          string

	Time             float64
	TimeToFirstToken float64

	Usage LLMMsgUsage
}

// Llama.cpp settings.
type LLMLlamacpp struct {
	lock sync.Mutex

	Address string
	Port    int

	Stats []LLMLlamacppMsgStats
}

func NewLLMLlamacpp(file string) (*LLMLlamacpp, error) {
	st := &LLMLlamacpp{}

	st.Address = "http://localhost"
	st.Port = 8070

	return LoadFile(file, "LLMLlamacpp", "json", st, true)
}

func (wsp *LLMLlamacpp) Check() error {
	if wsp.Address == "" {
		return fmt.Errorf("llama.cpp address is empty")
	}

	return nil
}

func (wsp *LLMLlamacpp) GetUrlHealth() string {
	return fmt.Sprintf("%s:%d/health", wsp.Address, wsp.Port)
}
