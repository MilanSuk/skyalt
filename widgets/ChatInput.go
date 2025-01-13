package main

type ChatInput struct {
	UserMsg string
	sended  func()
}

func (layout *Layout) AddChatInput(x, y, w, h int, props *ChatInput) *ChatInput {
	layout._createDiv(x, y, w, h, "ChatInput", props.Build, nil, nil)
	return props
}

func (st *ChatInput) Build(layout *Layout) {
	layout.SetColumn(0, 1, 100)
	layout.SetColumn(1, 3, 3)
	layout.SetRowFromSub(0, 1, 5)

	NewMsg := layout.AddEditboxMultiline(0, 0, 2, 1, &st.UserMsg)
	NewMsg.enter = func() {
		if st.UserMsg != "" && st.sended != nil {
			st.sended()
		}
	}

	{
		ImgsList := layout.AddLayoutList(0, 1, 1, 1, true)
		//already added small preview ...
		//+dialog preview + remove ...

		addImgLay := ImgsList.AddListSubItem()
		AddImgBt := addImgLay.AddButton(0, 0, 1, 1, "+")
		AddImgBt.clicked = func() {
			//relative dialog FilePicker ...
		}
	}

	SendBt := layout.AddButton(1, 1, 1, 1, "Send")
	SendBt.clicked = NewMsg.enter

}
