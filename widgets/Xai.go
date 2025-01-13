package main

import (
	"strings"
)

type Xai struct {
	Enable  bool
	Api_key string

	ChatModel string
}

func (layout *Layout) AddXai(x, y, w, h int, props *Xai) *Xai {
	layout._createDiv(x, y, w, h, "Xai", props.Build, nil, nil)
	return props
}

func (st *Xai) Build(layout *Layout) {

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
	KeyBt.BrowserUrl = "https://console.x.ai"
	y++

	layout.AddText(0, y, 1, 1, "Chat Model")
	layout.AddCombo(1, y, 1, 1, &st.ChatModel, Xai_GetChatModelList(), Xai_GetChatModelList())
	y++
}

func Xai_GetChatModelList() []string {
	return []string{
		"grok-2-vision-1212",
		"grok-2-1212",
		"grok-vision-beta",
		"grok-beta",
	}
}
