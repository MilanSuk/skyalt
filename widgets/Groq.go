package main

import (
	"strings"
)

type Groq struct {
	Enable  bool
	Api_key string

	ChatModel string
}

func (layout *Layout) AddGroq(x, y, w, h int, props *Groq) *Groq {
	layout._createDiv(x, y, w, h, "Groq", props.Build, nil, nil)
	return props
}

var g_Groq *Groq

func OpenFile_Groq() *Groq {
	if g_Groq == nil {
		g_Groq = &Groq{Enable: true, ChatModel: "llama-3.3-70b-versatile"}
		_read_file("Groq-Groq", g_Groq)
	}
	return g_Groq
}

func (st *Groq) Build(layout *Layout) {

	if st.ChatModel == "" {
		st.ChatModel = "llama-3.3-70b-versatile"
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
	KeyBt.BrowserUrl = "https://console.groq.com/keys"
	y++

	layout.AddText(0, y, 1, 1, "Chat Model")
	layout.AddCombo(1, y, 1, 1, &st.ChatModel, Groq_GetChatModelList(), Groq_GetChatModelList())
	y++
}

func Groq_GetChatModelList() []string {
	return []string{
		"llama-3.3-70b-versatile",
	}
}
