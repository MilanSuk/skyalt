package main

func _skyalt_save() {
	if g_About != nil {
		_write_file("About-About", g_About)
		g_About = nil
	}
	if g_Activities != nil {
		_write_file("Activities-Activities", g_Activities)
		g_Activities = nil
	}
	if g_Anthropic != nil {
		_write_file("Anthropic-Anthropic", g_Anthropic)
		g_Anthropic = nil
	}
	if g_AssistantChat != nil {
		_write_file("AssistantChat-AssistantChat", g_AssistantChat)
		g_AssistantChat = nil
	}
	if g_Counter != nil {
		_write_file("Counter-Counter", g_Counter)
		g_Counter = nil
	}
	if g_Events != nil {
		_write_file("Events-Events", g_Events)
		g_Events = nil
	}
	if g_Groq != nil {
		_write_file("Groq-Groq", g_Groq)
		g_Groq = nil
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
	if g_Osm != nil {
		_write_file("Osm-Osm", g_Osm)
		g_Osm = nil
	}
	if g_RootHeader != nil {
		_write_file("RootHeader-RootHeader", g_RootHeader)
		g_RootHeader = nil
	}
	if g_Settings != nil {
		_write_file("Settings-Settings", g_Settings)
		g_Settings = nil
	}
	if g_UserInfo != nil {
		_write_file("UserInfo-UserInfo", g_UserInfo)
		g_UserInfo = nil
	}
	if g_Whispercpp != nil {
		_write_file("Whispercpp-Whispercpp", g_Whispercpp)
		g_Whispercpp = nil
	}
	if g_Xai != nil {
		_write_file("Xai-Xai", g_Xai)
		g_Xai = nil
	}
	if g_test != nil {
		_write_file("test-test", g_test)
		g_test = nil
	}
}
