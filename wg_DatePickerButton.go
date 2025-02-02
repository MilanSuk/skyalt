package main

type DatePickerButton struct {
	Value *int64
	Page  *int64

	changed func()
}

func (layout *Layout) AddDatePickerButton(x, y, w, h int, value *int64, page *int64) *DatePickerButton {
	props := &DatePickerButton{Value: value, Page: page}
	layout._createDiv(x, y, w, h, "DatePickerButton", props.Build, nil, nil)
	return props
}

func (st *DatePickerButton) Build(layout *Layout) {

	if st.Page == nil {
		st.Page = Layout_GetDatePage()
	}

	layout.SetColumn(0, 1, 100)
	layout.SetRow(0, 1, 100)

	bt := layout.AddButton(0, 0, 1, 1, Layout_ConvertTextDate(*st.Value))
	bt.Icon = "resources/calendar.png"
	bt.Align = 0
	bt.Icon_margin = 0.1
	bt.Border = true
	bt.Background = 0

	cdialog := layout.AddDialog("dialog")
	cdialog.Layout.SetColumn(0, 7, 9)
	cdialog.Layout.SetRow(0, 8, 11)
	cld := cdialog.Layout.AddDatePicker(0, 0, 1, 1, st.Value, st.Page)
	cld.changed = func() {
		if st.changed != nil {
			st.changed()
		}
	}
	bt.clicked = func() {
		cdialog.OpenRelative(layout)
	}
}
