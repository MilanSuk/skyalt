package main

import (
	"path/filepath"
	"strings"
)

func OpenFile_About() *About {
	return OpenFilePath_About("")
}
func OpenFilePath_About(path string) *About {
	props := OpenFile[About](path)
	return props
}
func OpenFile_Activities() *Activities {
	return OpenFilePath_Activities("")
}
func OpenFilePath_Activities(path string) *Activities {
	props := OpenFile[Activities](path)
	return props
}
func OpenFile_Anthropic() *Anthropic {
	return OpenFilePath_Anthropic("")
}
func OpenFilePath_Anthropic(path string) *Anthropic {
	props := OpenFile[Anthropic](path)
	return props
}
func OpenFile_AssistantChat() *AssistantChat {
	return OpenFilePath_AssistantChat("")
}
func OpenFilePath_AssistantChat(path string) *AssistantChat {
	props := OpenFile[AssistantChat](path)
	return props
}
func OpenFile_Chat() *Chat {
	return OpenFilePath_Chat("")
}
func OpenFilePath_Chat(path string) *Chat {
	props := OpenFile[Chat](path)
	return props
}
func OpenFile_Chats() *Chats {
	return OpenFilePath_Chats("")
}
func OpenFilePath_Chats(path string) *Chats {
	props := OpenFile[Chats](path)
	return props
}
func OpenFile_Counter() *Counter {
	return OpenFilePath_Counter("")
}
func OpenFilePath_Counter(path string) *Counter {
	props := OpenFile[Counter](path)
	return props
}
func OpenFile_Events() *Events {
	return OpenFilePath_Events("")
}
func OpenFilePath_Events(path string) *Events {
	props := OpenFile[Events](path)
	return props
}
func OpenFile_Groq() *Groq {
	return OpenFilePath_Groq("")
}
func OpenFilePath_Groq(path string) *Groq {
	props := OpenFile[Groq](path)
	return props
}
func OpenFile_Llamacpp() *Llamacpp {
	return OpenFilePath_Llamacpp("")
}
func OpenFilePath_Llamacpp(path string) *Llamacpp {
	props := OpenFile[Llamacpp](path)
	return props
}
func OpenFile_Logs() *Logs {
	return OpenFilePath_Logs("")
}
func OpenFilePath_Logs(path string) *Logs {
	props := OpenFile[Logs](path)
	return props
}
func OpenFile_Microphone() *Microphone {
	return OpenFilePath_Microphone("")
}
func OpenFilePath_Microphone(path string) *Microphone {
	props := OpenFile[Microphone](path)
	return props
}
func OpenFile_OpenAI() *OpenAI {
	return OpenFilePath_OpenAI("")
}
func OpenFilePath_OpenAI(path string) *OpenAI {
	props := OpenFile[OpenAI](path)
	return props
}
func OpenFile_Osm() *Osm {
	return OpenFilePath_Osm("")
}
func OpenFilePath_Osm(path string) *Osm {
	props := OpenFile[Osm](path)
	return props
}
func OpenFile_Root() *Root {
	return OpenFilePath_Root("")
}
func OpenFilePath_Root(path string) *Root {
	props := OpenFile[Root](path)
	return props
}
func OpenFile_RootApps() *RootApps {
	return OpenFilePath_RootApps("")
}
func OpenFilePath_RootApps(path string) *RootApps {
	props := OpenFile[RootApps](path)
	return props
}
func OpenFile_RootHeader() *RootHeader {
	return OpenFilePath_RootHeader("")
}
func OpenFilePath_RootHeader(path string) *RootHeader {
	props := OpenFile[RootHeader](path)
	return props
}
func OpenFile_Settings() *Settings {
	return OpenFilePath_Settings("")
}
func OpenFilePath_Settings(path string) *Settings {
	props := OpenFile[Settings](path)
	return props
}
func OpenFile_Test() *Test {
	return OpenFilePath_Test("")
}
func OpenFilePath_Test(path string) *Test {
	props := OpenFile[Test](path)
	return props
}
func OpenFile_UserInfo() *UserInfo {
	return OpenFilePath_UserInfo("")
}
func OpenFilePath_UserInfo(path string) *UserInfo {
	props := OpenFile[UserInfo](path)
	return props
}
func OpenFile_Whispercpp() *Whispercpp {
	return OpenFilePath_Whispercpp("")
}
func OpenFilePath_Whispercpp(path string) *Whispercpp {
	props := OpenFile[Whispercpp](path)
	return props
}
func OpenFile_Xai() *Xai {
	return OpenFilePath_Xai("")
}
func OpenFilePath_Xai(path string) *Xai {
	props := OpenFile[Xai](path)
	return props
}
func (layout *Layout) AddApp(x, y, w, h int, path string) *Layout {
	name := filepath.Base(path)
	d := strings.IndexByte(name, '-')
	if d <= 0 {
		return nil //? ...
	}
	var lay *Layout
	switch name[:d] {
	case "About":
		props := OpenFilePath_About(path)
		lay = layout._createDiv(x, y, w, h, "About", props.Build, nil, nil)
	case "Activities":
		props := OpenFilePath_Activities(path)
		lay = layout._createDiv(x, y, w, h, "Activities", props.Build, nil, nil)
	case "Anthropic":
		props := OpenFilePath_Anthropic(path)
		lay = layout._createDiv(x, y, w, h, "Anthropic", props.Build, nil, nil)
	case "AssistantChat":
		props := OpenFilePath_AssistantChat(path)
		lay = layout._createDiv(x, y, w, h, "AssistantChat", props.Build, nil, nil)
	case "Chat":
		props := OpenFilePath_Chat(path)
		lay = layout._createDiv(x, y, w, h, "Chat", props.Build, nil, nil)
	case "Chats":
		props := OpenFilePath_Chats(path)
		lay = layout._createDiv(x, y, w, h, "Chats", props.Build, nil, nil)
	case "Counter":
		props := OpenFilePath_Counter(path)
		lay = layout._createDiv(x, y, w, h, "Counter", props.Build, nil, nil)
	case "Events":
		props := OpenFilePath_Events(path)
		lay = layout._createDiv(x, y, w, h, "Events", props.Build, nil, nil)
	case "Groq":
		props := OpenFilePath_Groq(path)
		lay = layout._createDiv(x, y, w, h, "Groq", props.Build, nil, nil)
	case "Llamacpp":
		props := OpenFilePath_Llamacpp(path)
		lay = layout._createDiv(x, y, w, h, "Llamacpp", props.Build, nil, nil)
	case "Logs":
		props := OpenFilePath_Logs(path)
		lay = layout._createDiv(x, y, w, h, "Logs", props.Build, nil, nil)
	case "Microphone":
		props := OpenFilePath_Microphone(path)
		lay = layout._createDiv(x, y, w, h, "Microphone", props.Build, nil, nil)
	case "OpenAI":
		props := OpenFilePath_OpenAI(path)
		lay = layout._createDiv(x, y, w, h, "OpenAI", props.Build, nil, nil)
	case "Osm":
		props := OpenFilePath_Osm(path)
		lay = layout._createDiv(x, y, w, h, "Osm", props.Build, nil, nil)
	case "Root":
		props := OpenFilePath_Root(path)
		lay = layout._createDiv(x, y, w, h, "Root", props.Build, nil, nil)
	case "RootApps":
		props := OpenFilePath_RootApps(path)
		lay = layout._createDiv(x, y, w, h, "RootApps", props.Build, nil, nil)
	case "RootHeader":
		props := OpenFilePath_RootHeader(path)
		lay = layout._createDiv(x, y, w, h, "RootHeader", props.Build, nil, nil)
	case "Settings":
		props := OpenFilePath_Settings(path)
		lay = layout._createDiv(x, y, w, h, "Settings", props.Build, nil, nil)
	case "Test":
		props := OpenFilePath_Test(path)
		lay = layout._createDiv(x, y, w, h, "Test", props.Build, nil, nil)
	case "UserInfo":
		props := OpenFilePath_UserInfo(path)
		lay = layout._createDiv(x, y, w, h, "UserInfo", props.Build, nil, nil)
	case "Whispercpp":
		props := OpenFilePath_Whispercpp(path)
		lay = layout._createDiv(x, y, w, h, "Whispercpp", props.Build, nil, nil)
	case "Xai":
		props := OpenFilePath_Xai(path)
		lay = layout._createDiv(x, y, w, h, "Xai", props.Build, nil, nil)
	}
	return lay
}
