package main

import (
	"strings"
)

type Anthropic struct {
	Enable  bool
	Api_key string

	ChatModel string
}

func (layout *Layout) AddAnthropic(x, y, w, h int, props *Anthropic) *Anthropic {
	layout._createDiv(x, y, w, h, "Anthropic", props.Build, nil, nil)
	return props
}

var g_Anthropic *Anthropic

func OpenFile_Anthropic() *Anthropic {
	if g_Anthropic == nil {
		g_Anthropic = &Anthropic{Enable: true, ChatModel: "claude-3-5-haiku-latest"}
		_read_file("Anthropic-Anthropic", g_Anthropic)
	}
	return g_Anthropic
}

func (st *Anthropic) Build(layout *Layout) {

	if st.ChatModel == "" {
		st.ChatModel = "grok-beta"
	}

	layout.SetColumn(0, 1, 4)
	layout.SetColumn(1, 1, 20)

	y := 0

	//enable
	layout.AddSwitch(1, y, 1, 1, "Enable", &st.Enable)
	y++

	//api key
	{
		layout.AddText(0, y, 1, 1, "API key")
		//encode
		key := ""
		if len(st.Api_key) > 6 {
			key = (st.Api_key)[:3] //first 3
			for i := 3; i < len(st.Api_key)-3; i++ {
				key += "*"
			}
			key += (st.Api_key)[len(st.Api_key)-3:] //last 3
		}
		KeyEd := layout.AddEditbox(1, y, 1, 1, &key)
		KeyEd.Formating = false
		KeyEd.changed = func() {
			if !strings.Contains(key, "*") {
				st.Api_key = key
			}
		}
		y++
	}

	KeyBt := layout.AddButton(1, y, 1, 1, "Get API key")
	KeyBt.Align = 0
	KeyBt.Background = 0
	KeyBt.BrowserUrl = "https://console.anthropic.com/settings/keys"
	y++

	layout.AddText(0, y, 1, 1, "Chat Model")
	layout.AddCombo(1, y, 1, 1, &st.ChatModel, Anthropic_GetChatModelList(), Anthropic_GetChatModelList())
	y++
}

func Anthropic_GetChatModelList() []string {
	return []string{
		"claude-3-5-haiku-latest",
		"claude-3-5-sonnet-latest",
		"claude-3-opus-latest"}
}
