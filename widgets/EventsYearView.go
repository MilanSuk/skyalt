package main

import (
	"time"
)

type EventsYearView struct {
	//Year          int
	Date *int64

	openMonth func()
	openDay   func()
}

func (layout *Layout) AddEventsYearView(x, y, w, h int) *EventsYearView {
	props := &EventsYearView{}
	layout._createDiv(x, y, w, h, "EventsYearView", props.Build, nil, nil)
	return props
}

func (st *EventsYearView) Build(layout *Layout) {

	layout.SetColumn(0, 1, 100)
	layout.SetRow(0, 1, 100)
	list := layout.AddLayoutList(0, 0, 1, 1, true)

	for i := 0; i < 12; i++ {
		//this_i := i
		Item := list.AddListSubItem()
		space := 0.5
		Item.SetColumn(0, 0.2, space)
		Item.SetColumn(1, 1, 6.5)
		Item.SetColumn(2, 0.2, space)
		Item.SetRow(0, 0.2, space)
		Item.SetRow(1, 1, 7.5)
		Item.SetRow(2, 0.2, space)

		Div := Item.AddLayout(1, 1, 1, 1)
		Div.SetColumn(0, 1, 100)
		Div.SetRow(1, 1, 100)

		bt := Div.AddButton(0, 0, 1, 1, NewButtonMenu("<h2>"+Layout_GetMonthText(1+i), "", 0))
		bt.clicked = func() {
			if st.openMonth != nil {
				st.openMonth() //time.Month(1 + this_i))
			}
		}

		year := Date_GetDateFromTime(*st.Date).Year
		page_unix := time.Date(year, time.Month(i+1), 1, 0, 0, 0, 0, time.Local).Unix()
		cal := Div.AddCalendar(0, 1, 1, 1, st.Date, &page_unix)
		cal.changed = func() {
			if st.openDay != nil {
				st.openDay()
			}
		}
	}
}
