package main

import (
	"fmt"
	"strconv"
	"time"
)

type EventsMonthView struct {
	Events  *Events
	Date    *int64
	openDay func()
}

func (layout *Layout) AddEventsMonthView(x, y, w, h int) *EventsMonthView {
	props := &EventsMonthView{}
	layout._createDiv(x, y, w, h, "EventsMonthView", props.Build, nil, nil)
	return props
}

func (st *EventsMonthView) Build(layout *Layout) {
	format := Layout_GetDateFormat()

	for x := 0; x < 7; x++ {
		layout.SetColumn(x, 1, 100)
	}

	for y := 0; y < 6; y++ {
		layout.SetRow(1+y, 1, 100)
	}

	//--Day names(short)--
	if format == "us" {
		//"us"
		tx := layout.AddText(0, 0, 1, 1, Layout_GetDayTextShort(7))
		tx.Align_h = 1 //center
		for x := 1; x < 7; x++ {
			tx := layout.AddText(x, 0, 1, 1, Layout_GetDayTextShort(x))
			tx.Align_h = 1 //center
		}
	} else {
		for x := 1; x < 8; x++ {
			tx := layout.AddText(x-1, 0, 1, 1, Layout_GetDayTextShort(x))
			tx.Align_h = 1 //center
		}
	}

	//grid
	for y := 0; y < 6; y++ {
		div := layout.AddLayout(0, 1+y, 7, 1)
		div.fnDraw = func(rect Rect, layout *Layout) (paint LayoutPaint) {
			paint.Line(rect, 0, 0, 1, 0, Paint_GetPalette().GetGrey(0.75), 0.03)
			return
		}
	}

	for x := 0; x < 6; x++ {
		div := layout.AddLayout(1+x, 1, 1, 6)
		div.fnDraw = func(rect Rect, layout *Layout) (paint LayoutPaint) {
			paint.Line(rect, 0, 0, 0, 1, Paint_GetPalette().GetGrey(0.75), 0.03)
			return
		}
	}

	{
		today := Date_GetStartDay(GetTime())
		dt := Date_GetStartWeek(Date_GetStartMonth(*st.Date), format)
		month := Date_GetDateFromTime(Date_GetStartMonth(*st.Date)).Month

		for y := 0; y < 6; y++ {
			for x := 0; x < 7; x++ {

				this_dt := dt
				stT := Date_GetStartDay(dt)
				enT := Date_AddDate(stT, 0, 0, 1)

				Div := layout.AddLayout(x, 1+y, 1, 1)
				{
					Div.SetColumn(0, 1, 100)
					Div.SetRow(1, 1, 100)

					if today == dt { //isToday
						Div.Back_cd = Color_Aprox(Paint_GetPalette().B, Paint_GetPalette().P, 0.3)
					}

					d := Date_GetDateFromTime(dt)

					bt := Div.AddButton(0, 0, 1, 1, "<h2>"+strconv.Itoa(d.Day)+".")
					bt.Background = 0
					bt.Align = 0
					bt.Cd_fade = (d.Month != month) //is day in current month
					bt.clicked = func() {
						*st.Date = this_dt
						if st.openDay != nil {
							st.openDay()
						}
					}

					Div2 := Div.AddLayout(0, 1, 1, 1)
					{
						events := st.Events.ExtractDayEvents(stT, enT)

						Div2.ScrollV.Narrow = true
						Div2.SetColumn(0, 1, 100)
						for i := 0; i < len(events); i++ {
							Div2.SetRow(i, 0.7, 0.7)
						}

						y := 0
						for _, event_id := range events {
							this_event_id := event_id
							event := &st.Events.Events[event_id]

							ShowDia, ShowLay := layout.AddDialogBorder("show_event_"+strconv.Itoa(event_id), "Event", 14)
							ShowLay.SetColumn(0, 1, 100)
							ShowLay.SetRowFromSub(0)
							sh := ShowLay.AddEventShow(0, 0, 1, 1)
							sh.Event = event
							sh.deleted = func() {
								st.Events.Events = append(st.Events.Events[:this_event_id], st.Events.Events[this_event_id+1:]...) //remove
							}

							toolTip := event.Title + ": " + Date_GetTextTime(event.Start) + " - " + Date_GetTextTime(event.Start+event.Duration)
							bt := Div2.AddButton(0, y, 1, 1, event.Title)
							bt.Tooltip = toolTip
							//bt.Draw_back = 2
							bt.Cd = event.Color
							bt.clicked = func() {
								ShowDia.OpenCentered()
							}

							y++
						}
					}
				}

				dt = enT //Date_AddDate(dt, 0, 0, 1) //add day
			}
		}
	}
}

func Date_GetTextTime(unix_sec int64) string {
	tm := time.Unix(unix_sec, 0)
	return fmt.Sprintf("%.02d:%.02d", tm.Hour(), tm.Minute())
}
