package main

type ChatInput struct {
	Text string

	Files          []string
	FilePickerPath string

	sended func()
}

func (layout *Layout) AddChatInput(x, y, w, h int, props *ChatInput) *ChatInput {
	layout._createDiv(x, y, w, h, "ChatInput", props.Build, nil, nil)
	return props
}

func (st *ChatInput) Build(layout *Layout) {
	layout.SetColumn(0, 1, 100)
	layout.SetColumn(1, 3, 3)
	layout.SetRowFromSub(0, 1, 5)
	layout.SetRowFromSub(1, 1, 2)

	NewMsg := layout.AddEditboxMultiline(0, 0, 2, 1, &st.Text)
	NewMsg.enter = func() {
		if st.Text != "" && st.sended != nil {
			st.sended()
		}
	}

	//image(s)
	{
		FileDia := layout.AddDialog("file")
		FileDia.Layout.SetColumn(0, 5, 20)
		FileDia.Layout.SetRow(0, 5, 10)
		pk := FileDia.Layout.AddFilePicker(0, 0, 1, 1, &st.FilePickerPath, true)
		pk.changed = func(close bool) {
			if close {
				st.Files = append(st.Files, st.FilePickerPath)
				st.FilePickerPath = "" //reset
				FileDia.Close()
			}
		}

		ImgsList := layout.AddLayoutList(0, 1, 1, 1, true)
		ImgsList.dropFile = func(path string) {
			st.Files = append(st.Files, path)
		}

		for fi, file := range st.Files {
			ImgDia := layout.AddDialog("image_" + file)
			ImgDia.Layout.SetColumn(0, 5, 12)
			ImgDia.Layout.SetColumn(1, 3, 3)
			ImgDia.Layout.SetRow(1, 5, 15)
			ImgDia.Layout.AddImage(0, 1, 2, 1, file)
			RemoveBt := ImgDia.Layout.AddButton(1, 0, 1, 1, "Remove")
			RemoveBt.clicked = func() {
				st.Files = append(st.Files[:fi], st.Files[fi+1:]...) //remove
				ImgDia.Close()
			}

			imgLay := ImgsList.AddListSubItem()
			imgLay.SetColumn(0, 2, 2)
			imgLay.SetRow(0, 2, 2)
			imgBt := imgLay.AddButtonIcon(0, 0, 1, 1, file, 0, file)
			imgBt.Background = 0
			imgBt.Cd = Paint_GetPalette().B
			imgBt.Border = true
			imgBt.clicked = func() {
				ImgDia.OpenRelative(imgLay)
			}
		}

		addImgLay := ImgsList.AddListSubItem()
		AddImgBt := addImgLay.AddButton(0, 0, 1, 1, "+")
		AddImgBt.clicked = func() {
			FileDia.OpenRelative(addImgLay)
		}
	}

	SendBt := layout.AddButton(1, 1, 1, 1, "Send")
	SendBt.clicked = NewMsg.enter
}

func (st *ChatInput) reset() {
	st.Text = ""
	st.Files = nil
}
