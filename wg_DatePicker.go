package main

import (
	"strconv"
	"time"
)

type DatePicker struct {
	Value *int64
	Page  *int64

	changed func()
}

func (layout *Layout) AddDatePicker(x, y, w, h int, value *int64, page *int64) *DatePicker {
	props := &DatePicker{Value: value, Page: page}
	layout._createDiv(x, y, w, h, "DatePicker", props.Build, nil, nil)
	return props
}

func (st *DatePicker) Build(layout *Layout) {

	if st.Page == nil {
		st.Page = &layout.ui.datePage
	}

	for x := 0; x < 7; x++ {
		layout.SetColumn(x, 1, 1.4)
	}
	layout.SetRow(0, 1, 1.4)
	layout.SetRow(1, 7*1, 7*1.4)
	layout.SetRow(2, 2, 2)

	//Info
	Today := layout.AddButton(0, 0, 2, 1, "Today")
	Info := layout.AddText(2, 0, 3, 1, "<h2>"+Date_GetMonthYear(*st.Page, layout))
	PageBack := layout.AddButton(5, 0, 1, 1, "<")
	PageForward := layout.AddButton(6, 0, 1, 1, ">")

	Info.Align_h = 1

	Today.Background = 0.5
	PageBack.Background = 0.5
	PageForward.Background = 0.5

	Today.Tooltip = layout.ConvertTextDate(time.Now().Unix())

	Today.clicked = func() {
		st.setPage(time.Now().Unix())
		*st.Value = time.Now().Unix()
		if st.changed != nil {
			st.changed()
		}
	}
	PageBack.clicked = func() {
		st.setPage(Date_AddDate(*st.Page, 0, -1, 0))
		if st.changed != nil {
			st.changed()
		}
	}
	PageForward.clicked = func() {
		st.setPage(Date_AddDate(*st.Page, 0, 1, 0))
		if st.changed != nil {
			st.changed()
		}
	}

	cl := layout.AddCalendar(0, 1, 7, 1, st.Value, st.Page)
	cl.changed = func() {
		if st.changed != nil {
			st.changed()
		}
	}

	//year/month/day
	{
		ymdDiv := layout.AddLayout(0, 2, 7, 1)
		ymdDiv.SetColumn(0, 1, 100)
		ymdDiv.SetColumn(1, 1, 100)
		ymdDiv.SetColumn(2, 1, 100)

		curr := Date_GetDateFromTime(*st.Value)

		ymdDiv.AddText(0, 0, 1, 1, "Year").Align_h = 1
		yearEd := ymdDiv.AddEditbox(0, 1, 1, 1, &curr.Year)
		yearEd.changed = func() {
			*st.Value = Date_GetTimeFromDate(&curr)
			*st.Page = *st.Value
			if st.changed != nil {
				st.changed()
			}
		}

		ymdDiv.AddText(1, 0, 1, 1, "Month").Align_h = 1
		monthEd := ymdDiv.AddEditbox(1, 1, 1, 1, &curr.Month)
		monthEd.changed = yearEd.changed

		ymdDiv.AddText(2, 0, 1, 1, "Day").Align_h = 1
		dayEd := ymdDiv.AddEditbox(2, 1, 1, 1, &curr.Day)
		dayEd.changed = yearEd.changed
	}
}

func (st *DatePicker) setPage(page int64) {
	//fix page(need to start with day 1)
	*st.Page = Date_GetStartMonth(page)
}

func Date_GetMonthYear(unix_sec int64, layout *Layout) string {
	tm := time.Unix(unix_sec, 0)
	return layout.GetMonthText(int(tm.Month())) + " " + strconv.Itoa(tm.Year())
}
