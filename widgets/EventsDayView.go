package main

import (
	"strconv"
)

type EventsDayView struct {
	Events *Events

	Date *int64

	openDay func()
}

func (layout *Layout) AddEventsDayView(x, y, w, h int) *EventsDayView {
	props := &EventsDayView{}
	layout._createDiv(x, y, w, h, "EventsDayView", props.Build, nil, nil)
	return props
}

func (st *EventsDayView) Build(layout *Layout) {

	format := Layout_GetDateFormat()

	layout.SetColumn(0, 1, 1) //time
	layout.SetColumn(1, 1, 100)
	layout.SetRow(24*2+1, 0.1, 0.1) //bottom

	todayD := Date_GetDateFromTime(GetTime())
	d := Date_GetDateFromTime(*st.Date)

	//header
	tx := layout.AddText(1, 0, 1, 1, "<h2>"+strconv.Itoa(d.Day)+". "+Layout_GetDayTextFull(d.GetWeekDay(format)+1))
	tx.Align_h = 1

	//days
	{
		//time
		for y := 0; y < 25; y++ {
			tx := layout.AddText(0, y*2, 1, 2, "<small>"+strconv.Itoa(y)+":00")
			tx.Align_h = 1
		}

		//grid
		for y := 0; y < 25; y++ {
			div := layout.AddLayout(1, y*2, 1, 2)
			div.EnableTouch = false //only touch, no visual fade
			div.fnDraw = func(rect Rect, layout *Layout) (paint LayoutPaint) {
				paint.Line(rect, 0, 0.5, 1, 0.5, Paint_GetPalette().GetGrey(0.75), 0.03)
				return
			}
		}

		//events
		{
			div := layout.AddLayoutWithName(1, 1, 1, 24*2, "events")
			st.Events.DayEvent(*st.Date, div, 24*2)
		}

		//time-line
		if d.CmpYMD(&todayD) { //is today
			now := GetTime()
			d := Date_GetDateFromTime(now)
			//hour, minute := GetHM(now)
			h := (float64(d.Hour) + (float64(d.Minute) / 60)) / 24
			div := layout.AddLayoutWithName(1, 1, 1, 24*2, "timeline") //same grid coord as Events div
			{
				div.EnableTouch = false //only touch, no visual fade
				div.fnDraw = func(rect Rect, layout *Layout) (paint LayoutPaint) {
					cd := Paint_GetPalette().OnB

					paint.Line(rect, 0, h, 1, h, cd, 0.03)
					paint.CircleRad(rect, 0, h, 0.3, cd, cd, cd, 0)
					return
				}
			}
		}
	}
}
