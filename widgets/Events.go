package main

import (
	"cmp"
	"image/color"
	"slices"
	"strconv"
	"time"
)

type EventsEvent struct {
	Title       string
	Description string
	File        string //attachment //multiple? ...

	Start    int64 //unix time
	Duration int64 //seconds

	Color color.RGBA
}

func (st *EventsEvent) SetNow(cd color.RGBA) {
	*st = EventsEvent{}

	st.Start = time.Now().Unix()
	st.Duration = 3600 / 2 //+30minutes
	st.Color = cd

}

type Events struct {
	HideSide bool
	Mode     string

	Events []EventsEvent

	NewEvent EventsEvent

	Selected_date int64 //unix time
	Selected_page int64 //unix time
}

func (layout *Layout) AddEvents(x, y, w, h int, props *Events) *Events {
	layout._createDiv(x, y, w, h, "Events", props.Build, nil, nil)
	return props
}

var g_Events *Events

func OpenFile_Events() *Events {
	if g_Events == nil {
		g_Events = &Events{Mode: "year"}
		_read_file("Events-Events", g_Events)
	}
	return g_Events
}

func (st *Events) Build(layout *Layout) {

	if st.Mode == "" {
		st.Mode = "year"
	}

	if st.Selected_date == 0 {
		st.Selected_date = GetTime()
	}
	if st.Selected_page == 0 {
		st.Selected_page = GetTime()
	}

	if !st.HideSide {
		layout.SetColumn(1, 6.3, 6.3)
	}
	layout.SetColumn(0, 1, 100)
	layout.SetRow(0, 1, 100)

	mode := layout.AddLayout(0, 0, 1, 1)
	st.ModePanel(mode)

	side := layout.AddLayout(1, 0, 1, 1)
	st.Side(side)

}

func GetTime() int64 {
	return time.Now().Unix()
}

func (st *Events) ModeYear(layout *Layout) {
	layout.SetColumn(0, 1, 100)
	layout.SetRow(0, 1, 100)
	view := layout.AddEventsYearView(0, 0, 1, 1)
	view.Date = &st.Selected_date
	view.openMonth = func() {
		st.Selected_page = st.Selected_date
		st.Mode = "month"
	}
	view.openDay = func() {
		st.Selected_page = st.Selected_date
		st.Mode = "day"
	}
}

func (st *Events) ModeMonth(layout *Layout) {
	layout.SetColumn(0, 1, 100)
	layout.SetRow(0, 1, 100)
	view := layout.AddEventsMonthView(0, 0, 1, 1)
	view.Events = st
	view.Date = &st.Selected_date
	view.openDay = func() {
		st.Selected_page = st.Selected_date
		st.Mode = "day"
	}
}

func (st *Events) ModeWeek(layout *Layout) {
	layout.SetColumn(0, 1, 100)
	layout.SetRow(0, 1, 100)
	view := layout.AddEventsWeekView(0, 0, 1, 1)
	view.Events = st
	view.Date = &st.Selected_date
	view.openDay = func() {
		st.Selected_page = st.Selected_date
		st.Mode = "day"
	}
}

func (st *Events) ModeDay(layout *Layout) {
	layout.SetColumn(0, 1, 100)
	layout.SetRow(0, 1, 100)
	view := layout.AddEventsDayView(0, 0, 1, 1)
	view.Events = st
	view.Date = &st.Selected_date
	view.openDay = func() {
		st.Selected_page = st.Selected_date
		st.Mode = "day"
	}
}

func (cal *Events) ModePanel(layout *Layout) {
	layout.SetColumn(0, 1, 100)
	layout.SetRow(1, 1, 100)

	var title string
	if cal.Mode == "year" {
		title = Date_GetYear(cal.Selected_date)
		modeDiv := layout.AddLayoutWithName(0, 1, 1, 1, "year")
		cal.ModeYear(modeDiv)
	} else if cal.Mode == "month" {
		title = Date_GetMonthYear(cal.Selected_date)
		modeDiv := layout.AddLayoutWithName(0, 1, 1, 1, "month")
		cal.ModeMonth(modeDiv)
	} else if cal.Mode == "week" {
		title = Date_GetMonthYear(cal.Selected_date)
		modeDiv := layout.AddLayoutWithName(0, 1, 1, 1, "week")
		cal.ModeWeek(modeDiv)
	} else if cal.Mode == "day" {
		title = Layout_ConvertTextDate(cal.Selected_date)
		modeDiv := layout.AddLayoutWithName(0, 1, 1, 1, "day")
		cal.ModeDay(modeDiv)
	}

	Div := layout.AddLayout(0, 0, 1, 1)
	{
		Div.SetColumn(0, 1, 2)
		Div.SetColumn(3, 1, 100)
		Div.SetColumn(4, 1, 8)

		//today
		TodayBt := Div.AddButton(0, 0, 1, 1, NewButton("Today"))
		TodayBt.Background = 0.5
		TodayBt.Tooltip = Layout_ConvertTextDate(GetTime())
		TodayBt.clicked = func() {
			cal.Selected_date = GetTime()
			cal.Selected_page = GetTime()
		}

		//arrows
		btL := Div.AddButton(1, 0, 1, 1, NewButton("<"))
		btR := Div.AddButton(2, 0, 1, 1, NewButton(">"))
		btL.Background = 0.5
		btR.Background = 0.5

		btL.clicked = func() {
			if cal.Mode == "year" {
				cal.Selected_date = Date_AddDate(cal.Selected_date, -1, 0, 0)
				cal.Selected_page = cal.Selected_date
			} else if cal.Mode == "month" {
				cal.Selected_date = Date_AddDate(cal.Selected_date, 0, -1, 0)
				cal.Selected_page = cal.Selected_date
			} else if cal.Mode == "week" {
				cal.Selected_date = Date_AddDate(cal.Selected_date, 0, 0, -7)
				cal.Selected_page = cal.Selected_date
			} else if cal.Mode == "day" {
				cal.Selected_date = Date_AddDate(cal.Selected_date, 0, 0, -1)
				cal.Selected_page = cal.Selected_date
			}
		}
		btR.clicked = func() {
			if cal.Mode == "year" {
				cal.Selected_date = Date_AddDate(cal.Selected_date, 1, 0, 0)
				cal.Selected_page = cal.Selected_date
			} else if cal.Mode == "month" {
				cal.Selected_date = Date_AddDate(cal.Selected_date, 0, 1, 0)
				cal.Selected_page = cal.Selected_date
			} else if cal.Mode == "week" {
				cal.Selected_date = Date_AddDate(cal.Selected_date, 0, 0, 7)
				cal.Selected_page = cal.Selected_date
			} else if cal.Mode == "day" {
				cal.Selected_date = Date_AddDate(cal.Selected_date, 0, 0, 1)
				cal.Selected_page = cal.Selected_date
			}
		}

		//title
		Title := Div.AddText(3, 0, 1, 1, "<h2>"+title)
		Title.Align_h = 1

		//Modes
		Div.AddTabs(4, 0, 1, 1, &cal.Mode, []string{"Day", "Week", "Month", "Year"}, []string{"day", "week", "month", "year"})
	}
}

func (cal *Events) Side(layout *Layout) {

	layout.SetColumn(0, 1, 100)
	layout.Back_cd = Paint_GetPalette().GetGrey(0.9)

	if !cal.HideSide {
		layout.SetRow(1, 0.3, 0.3)
		layout.SetRow(2, 1.2, 1.2)
		layout.SetRow(3, 6.5, 6.5)
		layout.SetRow(4, 0.3, 0.3)
		layout.SetRow(5, 1, 100)

		//create Event
		NewEventDia, NewEventDiaLay := layout.AddDialogBorder("new_event", "Event", 14)
		NewEventDiaLay.SetColumn(0, 1, 100)
		NewEventDiaLay.SetRowFromSub(0)
		ed := NewEventDiaLay.AddEventEdit(0, 0, 1, 1)
		ed.Event = &cal.NewEvent
		ed.ChangedButtonTitle = "Create"
		ed.changed = func() {
			cal.Events = append(cal.Events, cal.NewEvent)
			NewEventDia.Close()
		}
		ed.deleted = func() {
			NewEventDia.Close()
		}

		NewEventBt := layout.AddButton(0, 0, 1, 1, NewButton("New Event"))
		NewEventBt.clicked = func() {
			cal.NewEvent.SetNow(Paint_GetPalette().P)
			NewEventDia.OpenCentered()
		}

		layout.AddDivider(0, 1, 1, 1, true)

		PageDiv := layout.AddLayout(0, 2, 1, 1)
		{
			PageDiv.SetColumn(0, 1, 100)
			PageDiv.AddText(0, 0, 1, 1, "<h2>"+Date_GetMonthYear(cal.Selected_page))

			btL := PageDiv.AddButton(1, 0, 1, 1, NewButton("<"))
			btL.Background = 0.5
			btL.clicked = func() {
				cal.Selected_page = Date_AddDate(cal.Selected_page, 0, -1, 0)
			}

			btR := PageDiv.AddButton(2, 0, 1, 1, NewButton(">"))
			btR.Background = 0.5
			btR.clicked = func() {
				cal.Selected_page = Date_AddDate(cal.Selected_page, 0, 1, 0)
			}
		}

		layout.AddCalendar(0, 3, 1, 1, &cal.Selected_date, &cal.Selected_page)

		layout.AddDivider(0, 4, 1, 1, true)

		ShowSideDiv := layout.AddLayout(0, 6, 1, 1)
		{
			bt := ShowSideDiv.AddButton(0, 0, 1, 1, NewButton(">>"))
			bt.Background = 0.5
			bt.clicked = func() {
				cal.HideSide = true
			}
		}
	} else {
		layout.SetRow(0, 1, 100)

		bt := layout.AddButton(0, 1, 1, 1, NewButton("<<"))
		bt.Background = 0.5
		bt.clicked = func() {
			cal.HideSide = false
		}
	}
}

func (cal *Events) ExtractDayEvents(start, end int64) []int {

	var events []int //event_id

	//extract
	for i, ev := range cal.Events {
		if ev.Start >= start && (ev.Start+ev.Duration) < end {
			events = append(events, i)
		}
	}
	//order by .Start
	slices.SortFunc(events, func(a, b int) int {
		return cmp.Compare(cal.Events[a].Start, cal.Events[b].Start)
	})

	return events
}

type EventItem struct {
	event_id  int
	endVisual int64 //(end-start) can be too short, so endVisial is at least 15min different
}

func (ev *EventItem) GetEvent(cal *Events) *EventsEvent {
	return &cal.Events[ev.event_id]
}

func (cal *Events) HasCover(a, b EventItem) bool {

	//b is before
	if b.GetEvent(cal).Start < a.GetEvent(cal).Start && b.endVisual <= a.GetEvent(cal).Start {
		return false
	}

	//b is after
	if b.GetEvent(cal).Start >= a.endVisual && b.endVisual > a.endVisual {
		return false
	}

	return true
}

func (cal *Events) DayEvent(unix_sec int64, layout *Layout, layout_height float64) {

	stT := Date_GetStartDay(unix_sec)
	enT := Date_AddDate(stT, 0, 0, 1)

	events := cal.ExtractDayEvents(stT, enT)

	var cols [][]EventItem
	for _, event_id := range events {

		event := &cal.Events[event_id]

		var item EventItem
		item.event_id = event_id
		item.endVisual = event.Start + event.Duration
		if event.Duration < 3600/4 {
			item.endVisual = event.Start + 3600/4 //too short time => 15min(3600/4) at least
		}

		//find column
		fcol := 0
		for c := 0; c < len(cols); c++ {
			found := false
			for _, it := range cols[c] {
				if cal.HasCover(it, item) {
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

	layout.SetRow(0, 1, 100)

	for c := 0; c < len(cols); c++ {
		layout.SetColumn(c, 1, 100)
	}

	//height := layout.ComputeHeight() //- 0.15

	for c := 0; c < len(cols); c++ {
		Div := layout.AddLayout(c, 0, 1, 1)
		Div.SetColumn(0, 1, 100)

		last_end := float64(0)
		for i, it := range cols[c] {

			start := float64(it.GetEvent(cal).Start-stT) / float64(24*3600)
			end := float64(it.endVisual-stT) / float64(24*3600)
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
			event_i := it.event_id
			event := it.GetEvent(cal)
			tooltip := event.Title + ": " + Date_GetTextTime(event.Start) + " - " + Date_GetTextTime(event.Start+event.Duration)

			ShowDia, ShowLay := layout.AddDialogBorder("show_event_"+strconv.Itoa(it.event_id), "Event", 14)
			ShowLay.SetColumn(0, 1, 100)
			ShowLay.SetRowFromSub(0)
			sh := ShowLay.AddEventShow(0, 0, 1, 1)
			sh.Event = event
			sh.deleted = func() {
				cal.Events = append(cal.Events[:event_i], cal.Events[event_i+1:]...) //remove
			}

			bt := Div.AddButton(0, i*2+1, 1, 1, NewButton(event.Title))
			bt.Tooltip = tooltip
			bt.Cd = event.Color
			bt.Align = 0
			bt.clicked = func() {
				ShowDia.OpenCentered()
			}
		}
	}
}

func Date_GetYear(unix_sec int64) string {
	tm := time.Unix(unix_sec, 0)
	return strconv.Itoa(tm.Year())
}
