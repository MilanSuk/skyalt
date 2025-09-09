package main

import (
	"fmt"
	"image/color"
	"strconv"
	"time"
)

type CalendarEvent struct {
	EventID int64
	GroupID int64

	Title string

	Start    int64 //unix time
	Duration int64 //seconds

	Color color.RGBA
}

func (event *CalendarEvent) isInTimeRange(ST, EN int64) bool {
	itEnd := (event.Start + event.Duration)
	return (event.Start >= ST && event.Start < EN) || (itEnd >= ST && itEnd < EN) || (event.Start < ST && itEnd > EN)
}

type MonthCalendar struct {
	Year  int
	Month int //1=January, 2=February, etc.

	Events []CalendarEvent
}

func (layout *Layout) AddMonthCalendar(x, y, w, h int, Year int, Month int, Events []CalendarEvent) *MonthCalendar {
	props := &MonthCalendar{Year: Year, Month: Month, Events: Events}
	layout._createDiv(x, y, w, h, "MonthCalendar", props.Build, nil, nil)
	return props
}

func (st *MonthCalendar) Build(layout *Layout) {

	for x := 0; x < 7; x++ {
		layout.SetColumn(x, 1, Layout_MAX_SIZE)
	}

	for y := 0; y < 6; y++ {
		layout.SetRow(2+y, 3, 3)
	}

	layout.AddText(0, 0, 7, 1, fmt.Sprintf("<b>%s %d</b>", layout.GetMonthText(st.Month), st.Year)).Align_h = 1

	//--Day names(short)--
	if layout.ui.router.services.sync.GetDateFormat() == "us" {
		//"us"
		tx := layout.AddText(0, 1, 1, 1, layout.GetDayTextShort(7))
		tx.Align_h = 1 //center
		for x := 1; x < 7; x++ {
			tx := layout.AddText(x, 1, 1, 1, layout.GetDayTextShort(x))
			tx.Align_h = 1 //center
		}
	} else {
		for x := 1; x < 8; x++ {
			tx := layout.AddText(x-1, 1, 1, 1, layout.GetDayTextShort(x))
			tx.Align_h = 1 //center
		}
	}

	//grid
	{
		for y := range 6 {
			div := layout.AddLayout(0, 2+y, 7, 1)
			div.fnDraw = func(rect Rect, l *Layout) (paint LayoutPaint) {
				paint.Line(rect, 0, 0, 1, 0, layout.GetPalette().GetGrey(0.25), 0.03)
				return
			}
		}

		for x := range 6 {
			div := layout.AddLayout(1+x, 2, 1, 6)
			div.fnDraw = func(rect Rect, l *Layout) (paint LayoutPaint) {
				paint.Line(rect, 0, 0, 0, 1, layout.GetPalette().GetGrey(0.25), 0.03)
				return
			}
		}
	}

	{
		// Week Days
		today := time.Now()
		dt := time.Date(st.Year, time.Month(st.Month), 1, 0, 0, 0, 0, time.Local)
		{
			firstDay := time.Monday
			if layout.ui.router.services.sync.GetDateFormat() == "us" {
				firstDay = time.Sunday
			}
			for dt.Weekday() != firstDay {
				dt = dt.AddDate(0, 0, -1)
			}
		}

		for y := range 6 {
			for x := range 7 {
				Div := layout.AddLayout(x, 2+y, 1, 1)
				Div.Tooltip = dt.Format("Mon, 02 Jan 2006 15:04")
				{
					Div.SetColumn(0, 1, Layout_MAX_SIZE)
					Div.SetRow(1, 1, Layout_MAX_SIZE)

					Day := Div.AddText(0, 0, 1, 1, "<h2>"+strconv.Itoa(dt.Day())+".")
					if int(dt.Month()) != st.Month { //is day in current month
						Day.Cd.A = 100 //fade
					}

					if today.Year() == dt.Year() && today.Month() == dt.Month() && today.Day() == dt.Day() { //isToday
						Div.Back_cd = layout.GetPalette().S
						Day.Value = "<b>" + Day.Value
					}

					{
						DivEvents := Div.AddLayout(0, 1, 1, 1)
						DivEvents.scrollV.Narrow = true

						DivEvents.SetColumn(0, 1, Layout_MAX_SIZE)
						y := 0
						for _, event := range st.Events {
							if !event.isInTimeRange(dt.Unix(), dt.Unix()+(24*3600)) {
								continue
							}

							DivEvents.SetRow(y, 0.7, 0.7)
							y++
						}

						y = 0
						for _, event := range st.Events {
							if !event.isInTimeRange(dt.Unix(), dt.Unix()+(24*3600)) {
								continue
							}

							tx, txLay := DivEvents.AddText2(0, y, 1, 1, event.Title)
							tx.Tooltip = fmt.Sprintf("EventID: %d, Title: %s, Start: %s, End: %s, GroupID: %d", event.EventID, event.Title, layout.ConvertTextDateTime(event.Start), layout.ConvertTextDateTime(event.Start+event.Duration), event.GroupID)
							//tx.Tooltip = layout.ConvertTextTime(event.Start) + " - " + layout.ConvertTextTime(event.Start+event.Duration) + "\n" + event.Title
							txLay.Back_cd = event.Color
							if txLay.Back_cd.A == 0 {
								txLay.Back_cd = layout.GetPalette().P
							}
							txLay.Back_margin = 0.03

							y++
						}
					}
				}

				dt = dt.AddDate(0, 0, 1) //add day
			}
		}
	}

}
