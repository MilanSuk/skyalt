package main

type Anthropic struct {
	Enable     bool
	Api_key_id string

	ChatModel string
}

func (layout *Layout) AddAnthropic(x, y, w, h int, props *Anthropic) *Anthropic {
	layout._createDiv(x, y, w, h, "Anthropic", props.Build, nil, nil)
	return props
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
		KeyEd := layout.AddEditbox(1, y, 1, 1, &st.Api_key_id)
		KeyEd.Formating = false
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
