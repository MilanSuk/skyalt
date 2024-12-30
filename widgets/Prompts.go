package main

type Prompts struct {
}

func (layout *Layout) AddPrompts(x, y, w, h int) *Prompts {
	props := &Prompts{}
	layout._createDiv(x, y, w, h, "Prompts", props.Build, nil, nil)
	return props
}

func (st *Prompts) Build(layout *Layout) {

	layout.SetColumn(0, 1, 100)
	layout.SetColumn(1, 10, 10)
	layout.SetColumn(2, 1, 100)

	layout.AddText(1, 0, 1, 1, "Prompt examples").Align_h = 1

	y := 1
	st.AddButton("Show Chats app", &y, layout)
	st.AddButton("Show Events app", &y, layout)
	st.AddButton("Show Activities app", &y, layout)
	st.AddButton("Show Gallery app", &y, layout)

	st.AddButton("Show Services app", &y, layout)
	st.AddButton("Show Settings app", &y, layout)
	st.AddButton("Show Counter app", &y, layout)

	st.AddButton("Show About app", &y, layout)

}

func (st *Prompts) AddButton(prompt string, y *int, layout *Layout) {
	ast := NewFile_AssistantChat()

	bt := layout.AddButton(1, *y, 1, 1, NewButtonMenu(prompt, "", 0))
	bt.Background = 0.5
	bt.clicked = func() {
		ast.Prompt = prompt
		ast.Send()

		NewFile_Root().ShowPromptList = false
	}
	(*y)++
}
