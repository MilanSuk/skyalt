package main

type ButtonConfirm struct {
	Value string //label

	Tooltip string
	Align   int

	Draw_back   float64
	Draw_border bool

	Icon        string
	Icon_align  int
	Icon_margin float64

	Question string

	confirmed func()
}

func (layout *Layout) AddButtonConfirm(x, y, w, h int, label string, question string) *ButtonConfirm {
	props := &ButtonConfirm{Value: label, Align: 1, Draw_back: 1, Question: question}
	layout._createDiv(x, y, w, h, "ButtonConfirm", props.Build, nil, nil)
	return props
}

func (layout *Layout) AddButtonConfirmMenu(x, y, w, h int, label string, icon_path string, icon_margin float64, question string) *ButtonConfirm {
	props := &ButtonConfirm{Value: label, Icon: icon_path, Icon_margin: icon_margin, Align: 0, Draw_back: 0.25, Question: question}
	layout._createDiv(x, y, w, h, "ButtonConfirm", props.Build, nil, nil)
	return props
}

func (st *ButtonConfirm) Build(layout *Layout) {

	layout.SetColumn(0, 1, 100)
	layout.SetRow(0, 1, 100)

	dia := layout.AddDialog("confirm")
	{
		dia.Layout.SetColumn(0, 3, 5)
		dia.Layout.SetColumn(1, 0.5, 0.5)
		dia.Layout.SetColumn(2, 2, 5)

		tx := dia.Layout.AddText(0, 0, 3, 1, st.Question)
		tx.Align_h = 1

		yes := dia.Layout.AddButtonDanger(0, 1, 1, 1, "Yes")
		yes.clicked = func() {
			if st.confirmed != nil {
				st.confirmed()
			}
			dia.Close()
		}

		no := dia.Layout.AddButton(2, 1, 1, 1, "No")
		no.clicked = func() {
			dia.Close()
		}
	}

	bt, btL := layout.AddButton2(0, 0, 1, 1, st.Value)
	bt.Tooltip = st.Tooltip
	bt.Align = st.Align
	bt.Background = st.Draw_back
	bt.Border = st.Draw_border
	bt.Icon = st.Icon
	bt.Icon_align = st.Icon_align
	bt.Icon_margin = st.Icon_margin
	bt.Cd = Paint_GetPalette().E

	bt.clicked = func() {
		dia.OpenRelative(btL)
	}
}
