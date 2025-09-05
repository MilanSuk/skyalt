package main

import (
	"fmt"
)

// Llamacpp LLM settings.
type LLMLlamacpp struct {
	Address string
	Port    int

	Stats []LLMMsgStats
}

func (llama *LLMLlamacpp) Check() error {
	if llama.Address == "" {
		return LogsErrorf("llama.cpp address is empty")
	}

	return nil
}

func (llama *LLMLlamacpp) Complete(st *LLMComplete, app_port int, tools []*ToolsOpenAI_completion_tool, msg *AppsRouterMsg) error {
	err := llama.Check()
	if err != nil {
		return err
	}

	stats, err := OpenAI_Complete("llamacpp", fmt.Sprintf("%s:%d/v1", llama.Address, llama.Port), "", st, app_port, tools, msg, nil)
	llama.Stats = append(llama.Stats, stats...)
	return err
}
