package main

// [ignore]
type LoadFiles struct {
}

func (st *LoadFiles) run(caller *ToolCaller, ui *UI) error {

	NewDeviceSettings("")
	NewLLMxAI("")
	NewLLMMistral("")
	NewLLMOpenai("")
	NewLLMWhispercpp("")
	NewLLMLlamacpp("")
	NewMapSettings("")
	NewMicrophoneSettings("")

	return nil
}
