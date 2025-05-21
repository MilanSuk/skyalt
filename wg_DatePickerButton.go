package main

import "strconv"

type DatePickerButton struct {
	Value    *int64
	Page     *int64
	ShowTime bool
	changed  func()
}

func (layout *Layout) AddDatePickerButton(x, y, w, h int, value *int64, page *int64, showTime bool) *DatePickerButton {
	props := &DatePickerButton{Value: value, Page: page, ShowTime: showTime}
	layout._createDiv(x, y, w, h, "DatePickerButton", props.Build, nil, nil)
	return props
}

func (st *DatePickerButton) Build(layout *Layout) {
	if st.Page == nil {
		st.Page = &layout.ui.datePage
	}
	layout.SetRow(0, 1, 100)

	x := 0
	layout.SetColumn(x, 1, 4)
	bt := layout.AddButtonIcon(x, 0, 1, 1, "resources/calendar.png", 0.1, layout.ConvertTextDate(*st.Value))
	x++
	bt.Background = 0.5

	curr := Date_GetDateFromTime(*st.Value)
	changeEdit := func() {
		*st.Value = Date_GetTimeFromDate(&curr)
		*st.Page = *st.Value
		if st.changed != nil {
			st.changed()
		}
	}

	cdialog := layout.AddDialog("dialog")
	cdialog.Layout.SetColumn(0, 7, 9)
	cdialog.Layout.SetRow(0, 8, 11)
	cld := cdialog.Layout.AddDatePicker(0, 0, 1, 1, st.Value, st.Page)
	cld.changed = func() {

		next := Date_GetDateFromTime(*st.Value)
		curr.Year = next.Year
		curr.Month = next.Month
		curr.Day = next.Day

		changeEdit()
	}
	bt.clicked = func() {
		cdialog.OpenRelative(layout.UID)
	}

	layout.SetColumn(x, 1, 1)
	d := layout.AddEditbox(x, 0, 1, 1, &curr.Day)
	x++
	d.Ghost = "Day"
	d.Tooltip = "Day"
	d.changed = changeEdit

	layout.SetColumn(x, 1, 3)
	var month_labels []string
	var month_values []string
	for i := range 12 {
		month_labels = append(month_labels, layout.GetMonthText(1+i))
		month_values = append(month_values, strconv.Itoa(i+1))
	}
	monthValue := strconv.Itoa(curr.Month)
	m := layout.AddCombo(x, 0, 1, 1, &monthValue, month_labels, month_values)
	m.DialogWidth = 3
	x++
	m.changed = func() {
		curr.Month, _ = strconv.Atoi(monthValue)
		changeEdit()
	}

	layout.SetColumn(x, 1, 2)
	y := layout.AddEditbox(x, 0, 1, 1, &curr.Year)
	x++
	y.Ghost = "Year"
	y.Tooltip = "Year"
	y.changed = changeEdit

	if st.ShowTime {
		x++ //space

		layout.SetColumn(x, 1, 1)
		h := layout.AddEditbox(x, 0, 1, 1, &curr.Hour)
		x++
		h.Ghost = "Hour"
		h.Tooltip = "Hour"
		h.changed = changeEdit

		layout.SetColumn(x, 0.3, 0.3)
		layout.AddText(x, 0, 1, 1, ":").Align_h = 1
		x++ //space

		layout.SetColumn(x, 1, 1)
		mm := layout.AddEditbox(x, 0, 1, 1, &curr.Minute)
		x++
		mm.Ghost = "Minute"
		mm.Tooltip = "Minute"
		mm.changed = changeEdit
	}
}
