package main

func (st *ButtonConfirm) Build() {
	st.layout.SetColumn(0, 1, 100)
	st.layout.SetRow(0, 1, 100)

	dia := st.layout.AddDialog("confirm")
	{
		dia.SetColumn(0, 3, 3)
		dia.SetColumn(1, 0.5, 0.5)
		dia.SetColumn(2, 2, 3)

		tx := dia.AddText(0, 0, 3, 1, st.Question)
		tx.Align_h = 1

		yes := dia.AddButton(0, 1, 1, 1, NewButtonDanger("Yes", st.layout.GetPalette()))
		yes.clicked = func() {
			if st.Confirmed != nil {
				st.Confirmed()
			}
			dia.CloseDialog()
		}

		no := dia.AddButton(2, 1, 1, 1, NewButton("No"))
		no.clicked = func() {
			dia.CloseDialog()
		}
	}

	bt := st.layout.AddButton(0, 0, 1, 1, NewButton(st.Value))
	bt.Tooltip = st.Tooltip
	bt.Align = st.Align
	bt.Background = st.Draw_back
	bt.Border = st.Draw_border
	bt.Icon = st.Icon
	bt.Icon_align = st.Icon_align
	bt.Icon_margin = st.Icon_margin

	bt.clicked = func() {
		dia.OpenDialogRelative(bt.layout)
	}
}
