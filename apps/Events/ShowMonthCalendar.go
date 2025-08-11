package main

import (
	"image/color"
	"strconv"
	"time"
)

// Year int     // The year for which to show the calendar.
// Month int    // The month for which to show the calendar (1=January, etc.).
// GroupIDs []string  // List of Group IDs to filter events by. If empty, show all events. [optional]
type ShowMonthCalendar struct {
	Year     int
	Month    int
	GroupIDs []string
}

func (st *ShowMonthCalendar) run(caller *ToolCaller, ui *UI) error {
	eventsData, err := LoadEvents()
	if err != nil {
		return err
	}

	groupsData, err := LoadGroups()
	if err != nil {
		return err
	}

	// Calculate start and end times for the month
	firstOfMonth := time.Date(st.Year, time.Month(st.Month), 1, 0, 0, 0, 0, time.UTC)
	startTime := firstOfMonth.Unix()
	endOfMonth := firstOfMonth.AddDate(0, 1, 0) // First day of next month
	endTime := endOfMonth.Unix() - 1            // Last second of the current month

	eventIDs, err := FilterEvents(startTime, endTime, st.GroupIDs)
	if err != nil {
		return err
	}

	var uiEvents []UICalendarEvent
	for _, eventIDStr := range eventIDs {
		if event, found := eventsData.Items[eventIDStr]; found {
			eventID, err := strconv.ParseInt(event.EventID, 10, 64)
			if err != nil {
				continue // Skip if EventID can't be parsed
			}
			groupID, err := strconv.ParseInt(event.GroupID, 10, 64)
			if err != nil {
				continue // Skip if GroupID can't be parsed
			}

			uiEvent := UICalendarEvent{
				EventID:  eventID,
				GroupID:  groupID,
				Title:    event.Title,
				Start:    event.Start,
				Duration: event.Duration,
			}

			if group, found := groupsData.Items[event.GroupID]; found {
				uiEvent.Color = color.RGBA{group.Color.R, group.Color.G, group.Color.B, group.Color.A}
			}

			uiEvents = append(uiEvents, uiEvent)
		}
	}

	ui.addMonthCalendar(st.Year, st.Month, uiEvents)

	return nil
}
