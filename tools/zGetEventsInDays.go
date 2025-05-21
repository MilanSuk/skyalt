package main

import (
	"time"
)

// Returns list of EventsIDs in specified days.
type GetEventsInDays struct {
	Days   []string //Specific dates of days to filter. Date format: YYYY-MM-DD
	Groups []int    //filter only specific calendar groups. Empty=Show all groups. [optional]

	Out_events_ids []Day
}

type Day struct {
	Day       string
	EventsIDs []int64
}

func (st *GetEventsInDays) run(caller *ToolCaller, ui *UI) error {
	source_events, err := NewEvents("", caller)
	if err != nil {
		return err
	}

	st.Out_events_ids = nil

	for _, date := range st.Days {
		tm, err := time.ParseInLocation("2006-01-02", date, time.Local)
		if err != nil {
			return err
		}
		dayStart := time.Date(tm.Year(), tm.Month(), tm.Day(), 0, 0, 0, 0, time.Local)

		events_ids := source_events.filterEvents(dayStart.Unix(), dayStart.Unix()+(24*3600), st.Groups)

		st.Out_events_ids = append(st.Out_events_ids, Day{Day: date, EventsIDs: events_ids})
	}

	return nil
}
