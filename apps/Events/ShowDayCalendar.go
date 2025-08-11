package main

import (
	"image/color"
	"strconv"
	"time"
)

// ShowDayCalendar displays a calendar for the specified dates.
// It loads events from storage and filters them to include only those that occur on the provided dates.
type ShowDayCalendar struct {
	Dates []string // List of dates in format YYYY-MM-DD
}

func (st *ShowDayCalendar) run(caller *ToolCaller, ui *UI) error {
	// Convert date strings to Unix timestamps
	var days []int64
	for _, dateStr := range st.Dates {
		t, err := time.Parse("2006-01-02", dateStr) // Parse YYYY-MM-DD format
		if err != nil {
			return err
		}
		days = append(days, t.Unix()) // Add midnight timestamp for the day
	}

	// Load events and groups from storage
	eventsData, err := LoadEvents()
	if err != nil {
		return err
	}
	groupsData, err := LoadGroups()
	if err != nil {
		return err
	}

	// Filter and convert events to UICalendarEvent format
	var uiEvents []UICalendarEvent
	for _, event := range eventsData.Items {
		for _, dayTimestamp := range days {
			dayStart := dayTimestamp   // Start of the day
			dayEnd := dayStart + 86400 // End of the day (next midnight)
			if event.Start >= dayStart && event.Start < dayEnd {
				if group, found := groupsData.Items[event.GroupID]; found {
					eventIDInt, err1 := strconv.ParseInt(event.EventID, 10, 64) // Convert string EventID to int64
					if err1 == nil {
						groupIDInt, err2 := strconv.ParseInt(event.GroupID, 10, 64) // Convert string GroupID to int64
						if err2 == nil {
							uiEvents = append(uiEvents, UICalendarEvent{
								EventID:  eventIDInt,
								Title:    event.Title,
								Start:    event.Start,
								Duration: event.Duration,
								GroupID:  groupIDInt,
								Color:    color.RGBA{R: group.Color.R, G: group.Color.G, B: group.Color.B, A: 255},
							})
						}
					}
				}
			}
		}
	}

	// Add the day calendar to the UI
	ui.addDayCalendar(days, uiEvents)

	return nil
}
