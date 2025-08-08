package main

import (
	"time"
)

// Show calendar's Month view.
type ShowMonthCalendar struct {
	Year  int
	Month int //1=January, 2=February, etc.

	Groups []int //filter only specific calendar groups. Empty=Show all groups. [optional]
}

func (st *ShowMonthCalendar) run(caller *ToolCaller, ui *UI) error {
	source_events, err := NewEvents("")
	if err != nil {
		return err
	}

	// Week Days
	dt := time.Date(st.Year, time.Month(st.Month), 1, 0, 0, 0, 0, time.Local)
	{
		firstDay := time.Monday
		if UI_GetDateFormat() == "us" {
			firstDay = time.Sunday
		}
		for dt.Weekday() != firstDay {
			dt = dt.AddDate(0, 0, -1)
		}
	}
	var events []UICalendarEvent

	events_ids := source_events.filterEvents(dt.Unix(), dt.Unix()+(31*24*3600), st.Groups)
	for _, event_id := range events_ids {
		event := source_events.Events[event_id]

		events = append(events, UICalendarEvent{
			EventID:  event_id,
			GroupID:  event.GroupID,
			Title:    event.Title,
			Start:    event.Start,
			Duration: event.Duration,
			Color:    source_events.findGroupColorOrDefault(event.GroupID, caller)})

	}

	ui.SetColumn(0, 1, Layout_MAX_SIZE)
	ui.SetRowFromSub(0, 5, Layout_MAX_SIZE, true)

	ui.AddMonthCalendar(0, 0, 1, 1, st.Year, st.Month, events)

	return nil
}
