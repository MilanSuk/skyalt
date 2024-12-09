package main

func (st *Combo) Build() {

	st.layout.SetColumn(0, 1, 100)

	cdialog := st.layout.AddDialog("dialog")

	//button
	bt := st.layout.AddButton(0, 0, 1, 1, NewButton(st.getValueLabel())) //from scratch every frame => easy to debug. Hold
	bt.Tooltip = *st.Value

	bt.Icon = "resources/arrow.png"
	bt.Icon_margin = 0.1
	bt.Icon_align = 2
	bt.Align = 0
	bt.Background = 0
	bt.Border = true

	bt.clicked = func() {
		cdialog.OpenDialogRelative(st.layout)
	}

	{
		found_val := false
		for _, val := range st.Values {
			if val == *st.Value {
				found_val = true
				break
			}
		}
		if !found_val {
			bt.Cd = st.layout.GetPalette().E //error border
		}
	}

	//dialog
	{
		cdialog.SetColumn(0, 1, st.DialogWidth)
		for i, name := range st.Labels {

			bt := cdialog.AddButton(0, i, 1, 1, NewButton(name))
			bt.Tooltip = st.Values[i]
			if st.Values[i] == *st.Value {
				bt.Background = 1 //selected
			} else {
				bt.Background = 0.25
			}
			bt.Align = 0

			bt.clicked = func() {
				*st.Value = bt.Tooltip
				if st.changed != nil {
					st.changed()
				}
				cdialog.CloseDialog()
			}

		}
	}
}

func (st *Combo) getValueLabel() string {
	for i, it := range st.Values {
		if it == *st.Value {
			if i < len(st.Labels) {
				return st.Labels[i]
			}
			break
		}
	}
	return *st.Value //not found
}
