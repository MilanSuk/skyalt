package main

func _skyalt_save() {
	if g_Assistant != nil {
		_write_file("Assistant-Assistant", g_Assistant)
		g_Assistant = nil
	}
	if g_Counter != nil {
		_write_file("Counter-Counter", g_Counter)
		g_Counter = nil
	}
	if g_Env != nil {
		_write_file("Env-Env", g_Env)
		g_Env = nil
	}
	if g_Logs != nil {
		_write_file("Logs-Logs", g_Logs)
		g_Logs = nil
	}
	if g_Microphone != nil {
		_write_file("Microphone-Microphone", g_Microphone)
		g_Microphone = nil
	}
	if g_OpenAI != nil {
		_write_file("OpenAI-OpenAI", g_OpenAI)
		g_OpenAI = nil
	}
	if g_Root != nil {
		_write_file("Root-Root", g_Root)
		g_Root = nil
	}
	if g_Whispercpp != nil {
		_write_file("Whispercpp-Whispercpp", g_Whispercpp)
		g_Whispercpp = nil
	}
	if g_Xai != nil {
		_write_file("Xai-Xai", g_Xai)
		g_Xai = nil
	}
}
