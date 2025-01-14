package main

type ChatMsg struct {
	Text           string
	Files          []string
	CreatedTimeSec float64
	CreatedBy      string //name of service(AI), empty = user wrote it
}

func (layout *Layout) AddChatMsg(x, y, w, h int, props *ChatMsg) *ChatMsg {
	layout._createDiv(x, y, w, h, "ChatMsg", props.Build, nil, nil)
	return props
}

func (st *ChatMsg) Build(layout *Layout) {
	layout.SetColumn(0, 1, 100)
	layout.SetColumn(1, 3, 3) //date
	layout.SetRowFromSub(1, 1, 100)

	sender := "User"
	if st.CreatedBy != "" {
		sender = st.CreatedBy
		layout.Back_cd = Paint_GetPalette().GetGrey(0.8)
	}
	layout.AddText(0, 0, 1, 1, "<b>"+sender)
	date := layout.AddText(1, 0, 1, 1, "<small>"+Layout_ConvertTextDateTime(int64(st.CreatedTimeSec)))
	date.Align_h = 2

	//text
	layout.AddTextMultiline(0, 1, 2, 1, st.Text)

	//image(s)
	if len(st.Files) > 0 {
		layout.SetRowFromSub(2, 1, 2)
		ImgsList := layout.AddLayoutList(0, 2, 2, 1, true)

		for _, file := range st.Files {
			ImgDia := layout.AddDialog("image_" + file)
			ImgDia.Layout.SetColumn(0, 5, 15)
			ImgDia.Layout.SetRow(0, 5, 15)
			ImgDia.Layout.AddImage(0, 0, 1, 1, file)

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
	}

}
