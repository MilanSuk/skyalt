package main

import "time"

// Show calendar's Days view.
type ShowDayCalendar struct {
	Days []string //Specific dates of days to show. Date format: YYYY-MM-DD

	Groups []int //filter only specific calendar groups. Empty=Show all groups. [optional]
}

func (st *ShowDayCalendar) run(caller *ToolCaller, ui *UI) error {
	source_events, err := NewEvents("")
	if err != nil {
		return err
	}

	var days []int64
	var events []UICalendarEvent

	for _, date := range st.Days {
		tm, err := time.ParseInLocation("2006-01-02", date, time.Local)
		if err != nil {
			return err
		}
		dayStart := time.Date(tm.Year(), tm.Month(), tm.Day(), 0, 0, 0, 0, time.Local)

		days = append(days, dayStart.Unix())

		events_ids := source_events.filterEvents(dayStart.Unix(), dayStart.Unix()+(24*3600), st.Groups)

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
	}

	ui.SetColumn(0, 1, Layout_MAX_SIZE)
	ui.SetRowFromSub(0, 5, Layout_MAX_SIZE, true)

	ui.AddDayCalendar(0, 0, 1, 1, days, events)

	return nil
}
