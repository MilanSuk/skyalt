package main

import (
	"fmt"
	"slices"
	"strconv"
	"time"
)

type DayCalendar struct {
	Days   []int64
	Events []CalendarEvent
}

func (layout *Layout) AddDayCalendar(x, y, w, h int, Days []int64, Events []CalendarEvent) *DayCalendar {
	props := &DayCalendar{Days: Days, Events: Events}
	layout._createDiv(x, y, w, h, "DayCalendar", props.Build, nil, nil)
	return props
}

func (st *DayCalendar) Build(layout *Layout) {
	slices.Sort(st.Days)

	layout.SetColumn(0, 1.2, 1.2) //time
	for i := range st.Days {
		layout.SetColumn(i+1, 1, Layout_MAX_SIZE) //days
	}
	layout.SetRow(24*2+2, 0.1, 0.1) //bottom

	layout.AddText(0, 0, 1+len(st.Days), 1, "<b>Day(s) view</b>").Align_h = 1

	layout.Tooltip = fmt.Sprintf("Days(format YYYY-MM-DD): %v", st.Days)

	//header
	{
		x := 0
		for _, date := range st.Days {
			layout.AddText(1+x, 1, 1, 1, "<h2>"+layout.ConvertTextDate(date)).Align_h = 1
			x++
		}
	}

	//time
	for y := 0; y < 25; y++ {
		tx := layout.AddText(0, 1+y*2, 1, 2, "<small>"+strconv.Itoa(y)+":00")
		tx.Align_h = 1
	}

	//grid
	for y := range 25 {
		div := layout.AddLayout(1, 1+y*2, len(st.Days), 2)
		div.EnableTouch = false
		div.fnDraw = func(rect Rect, l *Layout) (paint LayoutPaint) {
			paint.Line(rect, 0, 0.5, 1, 0.5, layout.GetPalette().GetGrey(0.25), 0.03)
			return
		}
	}

	for x := 1; x < len(st.Days); x++ {
		div := layout.AddLayout(1+x, 1, 1, 24*2+1)
		div.Back_cd.A = 0
		div.EnableTouch = false
		div.fnDraw = func(rect Rect, l *Layout) (paint LayoutPaint) {
			paint.Line(rect, 0, 0, 0, 1, layout.GetPalette().GetGrey(0.25), 0.03)
			return
		}
	}
	//days
	{
		//events
		{
			today := time.Now()
			x := 0
			for _, date := range st.Days {
				tm := time.Unix(date, 0)

				dayStart := time.Date(tm.Year(), tm.Month(), tm.Day(), 0, 0, 0, 0, time.Local)

				div := layout.AddLayout(1+x, 2, 1, 24*2)
				div.Tooltip = fmt.Sprintf("Day(YYYY-MM-DD): %d-%d-%d", tm.Year(), tm.Month(), tm.Day())

				_DayCalendar_showDaysView_DayEvent(st.Events, dayStart.Unix(), div, 24*2)

				//time-line
				if today.Year() == tm.Year() && today.Month() == tm.Month() && today.Day() == tm.Day() {

					h := float64(today.Unix()-dayStart.Unix()) / (24 * 3600)
					div := layout.AddLayout(1, 2, len(st.Days), 24*2)
					div.Tooltip = fmt.Sprintf("Timeline with time(YY-MM-DD HH:MM): %s", today.Format("01-02-2006 15:04"))
					{
						div.EnableTouch = false
						div.fnDraw = func(rect Rect, l *Layout) (paint LayoutPaint) {
							cd := layout.GetPalette().S
							paint.Line(rect, 0, h, 1, h, cd, 0.06)
							paint.CircleRad(rect, 0, h, 0.3, cd, cd, cd, 0)
							return
						}
					}
				}

				x++
			}
		}
	}
}

type EventItem struct {
	event_i   int
	endVisual int64 //(end-start) can be too short, so endVisial is at least 15min different
}

func _DayCalendar_showDaysView_HasCover(Events []CalendarEvent, a, b EventItem) bool {
	//b is before
	if Events[b.event_i].Start < Events[a.event_i].Start && b.endVisual <= Events[a.event_i].Start {
		return false
	}

	//b is after
	if Events[b.event_i].Start >= a.endVisual && b.endVisual > a.endVisual {
		return false
	}

	return true
}

func _DayCalendar_showDaysView_DayEvent(Events []CalendarEvent, dayStart int64, layout *Layout, layout_height float64) {
	var cols [][]EventItem
	for event_i, event := range Events {

		if !event.isInTimeRange(dayStart, dayStart+(24*3600)) {
			continue
		}

		var item EventItem
		item.event_i = event_i
		item.endVisual = event.Start + event.Duration
		if event.Duration < 3600/4 {
			item.endVisual = event.Start + 3600/4 //too short time => 15min(3600/4) at least
		}

		//find column
		fcol := 0
		for c := range cols {
			found := false
			for _, it := range cols[c] {
				if _DayCalendar_showDaysView_HasCover(Events, it, item) {
					fcol++
					found = true
					break
				}
			}
			if !found {
				break
			}
		}

		//add
		if fcol == len(cols) {
			cols = append(cols, []EventItem{})
		}
		cols[fcol] = append(cols[fcol], item)
	}

	layout.SetRow(0, 1, Layout_MAX_SIZE)

	for c := range cols {
		layout.SetColumn(c, 1, Layout_MAX_SIZE)
	}

	for c := range cols {
		Div := layout.AddLayout(c, 0, 1, 1)
		Div.SetColumn(0, 1, Layout_MAX_SIZE)

		last_end := float64(0)
		for i, it := range cols[c] {

			start := float64(Events[it.event_i].Start-dayStart) / float64(24*3600)
			end := float64(it.endVisual-dayStart) / float64(24*3600)
			if end > 1 {
				end = 1
			}

			start *= layout_height
			end *= layout_height

			Div.SetRow(i*2+0, float64(start-last_end), float64(start-last_end)) //empty space to last
			Div.SetRow(i*2+1, float64(end-start), float64(end-start))           //this event
			last_end = end
		}

		for i, it := range cols[c] {
			event := Events[it.event_i]
			tx, txLay := Div.AddText2(0, i*2+1, 1, 1, event.Title)
			tx.Tooltip = fmt.Sprintf("Event ID: %d, Title: %s, Start: %s, End: %s, GroupID: %d", event.EventID, event.Title, layout.ConvertTextDateTime(event.Start), layout.ConvertTextDateTime(event.Start+event.Duration), event.GroupID)
			//tx.Tooltip = layout.ConvertTextTime(event.Start) + " - " + layout.ConvertTextTime(event.Start+event.Duration) + "\n" + event.Title
			txLay.Back_cd = event.Color
			if txLay.Back_cd.A == 0 {
				txLay.Back_cd = layout.GetPalette().P
			}
			txLay.Back_margin = 0.03
		}
	}
}
