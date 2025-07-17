package main

import "image/color"

type ButtonConfirm struct {
	Tooltip string

	Value string //label

	Align int

	Background float64
	Border     bool

	Color color.RGBA

	IconPath    string
	IconBlob    []byte
	Icon_align  int
	Icon_margin float64

	Shortcut_key byte

	Question string

	confirmed func()
}

func (layout *Layout) AddButtonConfirm(x, y, w, h int, label string, question string) *ButtonConfirm {
	props := &ButtonConfirm{Value: label, Align: 1, Background: 1, Question: question}
	layout._createDiv(x, y, w, h, "ButtonConfirm", props.Build, nil, nil)
	return props
}

func (layout *Layout) AddButtonConfirmMenu(x, y, w, h int, label string, icon_path string, icon_margin float64, question string) *ButtonConfirm {
	props := &ButtonConfirm{Value: label, IconPath: icon_path, Icon_margin: icon_margin, Align: 0, Background: 0.25, Question: question}
	layout._createDiv(x, y, w, h, "ButtonConfirm", props.Build, nil, nil)
	return props
}

func (st *ButtonConfirm) Build(layout *Layout) {
	layout.Back_rounding = true //for disable

	layout.SetColumn(0, 1, 100)
	layout.SetRow(0, 1, 100)

	dia := layout.AddDialog("confirm")
	{
		dia.Layout.SetColumnFromSub(0, 6, 20, true)

		tx := dia.Layout.AddText(0, 0, 1, 1, st.Question)
		tx.Align_h = 1

		footerDiv := dia.Layout.AddLayout(0, 1, 1, 1)

		footerDiv.SetColumn(0, 3, 100)
		footerDiv.SetColumn(1, 0.5, 0.5)
		footerDiv.SetColumn(2, 3, 100)

		yes := footerDiv.AddButtonDanger(0, 0, 1, 1, "Yes")
		yes.clicked = func() {
			if st.confirmed != nil {
				st.confirmed()
			}
			dia.Close()
		}

		no := footerDiv.AddButton(2, 0, 1, 1, "No")
		no.clicked = func() {
			dia.Close()
		}
	}

	bt, btL := layout.AddButton2(0, 0, 1, 1, st.Value)
	bt.Tooltip = st.Tooltip
	bt.Align = st.Align
	bt.Background = st.Background
	bt.Border = st.Border
	bt.IconPath = st.IconPath
	bt.IconBlob = st.IconBlob
	bt.Icon_align = st.Icon_align
	bt.Icon_margin = st.Icon_margin
	bt.Cd = st.Color
	bt.ErrorColor = true
	bt.Shortcut_key = st.Shortcut_key

	bt.clicked = func() {
		dia.OpenRelative(btL.UID)
	}
}
