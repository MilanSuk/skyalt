package main

type Xai struct {
	Enable     bool
	Api_key_id string

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
		KeyEd := layout.AddEditbox(1, y, 1, 1, &st.Api_key_id)
		KeyEd.Formating = false
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
