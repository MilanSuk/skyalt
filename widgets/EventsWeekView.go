package main

import (
	"strconv"
)

type EventsWeekView struct {
	Events  *Events
	Date    *int64
	openDay func()
}

func (layout *Layout) AddEventsWeekView(x, y, w, h int) *EventsWeekView {
	props := &EventsWeekView{}
	layout._createDiv(x, y, w, h, "EventsWeekView", props.Build, nil, nil)
	return props
}

func (st *EventsWeekView) Build(layout *Layout) {
	format := Layout_GetDateFormat()

	layout.SetColumn(0, 1, 1) //time

	for i := 1; i < 8; i++ {
		layout.SetColumn(i, 1, 100) //days
	}

	layout.SetRow(24*2+1, 0.1, 0.1) //bottom

	//header
	{
		dt := Date_GetStartWeek(*st.Date, format)
		d := Date_GetDateFromTime(dt)

		if format == "us" {
			//"us"
			this_dt := dt
			bt := layout.AddButton(1, 0, 1, 1, NewButton("<h2>"+strconv.Itoa(d.Day)+". "+Layout_GetDayTextShort(7)))
			bt.Background = 0
			bt.clicked = func() {
				*st.Date = this_dt
				if st.openDay != nil {
					st.openDay()
				}
			}
			dt = Date_AddDate(dt, 0, 0, 1) //add day
			d = Date_GetDateFromTime(dt)

			for x := 1; x < 7; x++ {
				this_dt := dt
				bt := layout.AddButton(1+x, 0, 1, 1, NewButton("<h2>"+strconv.Itoa(d.Day)+". "+Layout_GetDayTextShort(x)))
				bt.Background = 0
				bt.clicked = func() {
					*st.Date = this_dt
					if st.openDay != nil {
						st.openDay()
					}
				}

				dt = Date_AddDate(dt, 0, 0, 1) //add day
				d = Date_GetDateFromTime(dt)
			}
		} else {
			for x := 1; x < 8; x++ {

				this_dt := dt
				bt := layout.AddButton(x, 0, 1, 1, NewButton("<h2>"+strconv.Itoa(d.Day)+". "+Layout_GetDayTextShort(x)))
				bt.Background = 0
				bt.clicked = func() {
					*st.Date = this_dt
					if st.openDay != nil {
						st.openDay()
					}
				}

				dt = Date_AddDate(dt, 0, 0, 1) //add day
				d = Date_GetDateFromTime(dt)
			}
		}
	}

	//days
	{
		//time
		for y := 0; y < 25; y++ {
			tx := layout.AddText(0, y*2, 1, 2, "<small>"+strconv.Itoa(y)+":00")
			tx.Align_h = 1
		}

		//grid
		for y := 0; y < 25; y++ {
			div := layout.AddLayout(1, y*2, 7, 2)
			div.EnableTouch = false //only touch, no visual fade
			div.fnDraw = func(rect Rect, layout *Layout) (paint LayoutPaint) {
				paint.Line(rect, 0, 0.5, 1, 0.5, Paint_GetPalette().GetGrey(0.75), 0.03)
				return
			}
		}

		for x := 1; x < 7; x++ {
			div := layout.AddLayout(1+x, 0, 1, 24*2+1)
			div.EnableTouch = false //only touch, no visual fade
			div.fnDraw = func(rect Rect, layout *Layout) (paint LayoutPaint) {
				paint.Line(rect, 0, 0, 0, 1, Paint_GetPalette().GetGrey(0.75), 0.03)
				return
			}
		}

		//events
		dt := Date_GetStartWeek(*st.Date, format)
		for x := 0; x < 7; x++ {
			div := layout.AddLayout(1+x, 1, 1, 24*2)
			st.Events.DayEvent(dt, div, 24*2)

			dt = Date_AddDate(dt, 0, 0, 1) //add day
		}

		//time-line
		today := GetTime()
		w1 := Date_GetStartWeek(today, format)
		w2 := Date_GetStartWeek(*st.Date, format)
		if w1 == w2 { //today is in current week

			h := float64(today-Date_GetStartDay(today)) / (24 * 3600)

			div := layout.AddLayout(1, 1, 7, 24*2)
			{
				div.EnableTouch = false //only touch, no visual fade
				d := Date_GetDateFromTime(today)

				div.fnDraw = func(rect Rect, layout *Layout) (paint LayoutPaint) {
					cd := Paint_GetPalette().OnB

					paint.Line(rect, 0, h, 1, h, cd, 0.03)
					paint.CircleRad(rect, float64(d.GetWeekDay(format))/7, h, 0.3, cd, cd, cd, 0)
					return
				}
			}
		}
	}
}
