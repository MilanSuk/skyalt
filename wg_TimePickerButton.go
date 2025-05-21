package main

import "time"

type TimePickerButton struct {
	Value *int64
	Page  *int64

	changed func()
}

func (layout *Layout) AddTimePickerButton(x, y, w, h int, value *int64, page *int64) *TimePickerButton {
	props := &TimePickerButton{Value: value, Page: page}
	layout._createDiv(x, y, w, h, "TimePickerButton", props.Build, nil, nil)
	return props
}

func (st *TimePickerButton) Build(layout *Layout) {
	if st.Page == nil {
		st.Page = &layout.ui.datePage
	}

	layout.SetColumn(0, 1, 100)
	layout.SetColumn(1, 0.5, 0.5) //space
	layout.SetColumn(3, 0.5, 0.5)
	layout.SetRow(0, 1, 100)

	btDate := layout.AddButton(0, 0, 1, 1, layout.ConvertTextDate(*st.Value))
	btDate.Tooltip = "Date"
	btDate.IconPath = "resources/calendar.png"
	btDate.Align = 0
	btDate.Icon_margin = 0.1
	btDate.Border = true
	btDate.Background = 0

	date := time.Unix(*st.Value, 0)

	hour := date.Hour()
	edHour := layout.AddEditbox(2, 0, 1, 1, &hour)
	edHour.Tooltip = "Hour"
	edHour.changed = func() {
		*st.Value = time.Date(date.Year(), time.Month(date.Month()), date.Day(), hour, date.Minute(), 0, 0, time.Local).Unix()
		if st.changed != nil {
			st.changed()
		}
	}

	layout.AddText(3, 0, 1, 1, ":").Align_h = 1

	min := date.Minute()
	edMin := layout.AddEditbox(4, 0, 1, 1, &min)
	edMin.Tooltip = "Minute"
	edMin.changed = func() {
		*st.Value = time.Date(date.Year(), time.Month(date.Month()), date.Day(), date.Hour(), min, 0, 0, time.Local).Unix()
		if st.changed != nil {
			st.changed()
		}
	}
	layout.AddText(5, 0, 1, 1, "H:M")

	cdialog := layout.AddDialog("dialog")
	cdialog.Layout.SetColumn(0, 7, 9)
	cdialog.Layout.SetRow(0, 8, 11)
	cld := cdialog.Layout.AddDatePicker(0, 0, 1, 1, st.Value, st.Page)
	cld.changed = func() {
		if st.changed != nil {
			st.changed()
		}
	}
	btDate.clicked = func() {
		cdialog.OpenRelative(layout.UID)
	}
}
