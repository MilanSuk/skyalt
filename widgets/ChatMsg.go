package main

type ChatMsg struct {
	Text           string
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

	layout.AddTextMultiline(0, 1, 2, 1, st.Text)

}
