package main

import (
	"fmt"
)

// ShowYearCalendar displays a calendar for the specified year.
type ShowYearCalendar struct {
	Year int // Year to display the calendar for
}

func (st *ShowYearCalendar) run(caller *ToolCaller, ui *UI) error {
	ui.addTextH1(fmt.Sprintf("Calendar for year %d", st.Year))

	ui.addYearCalendar(st.Year)

	return nil
}
