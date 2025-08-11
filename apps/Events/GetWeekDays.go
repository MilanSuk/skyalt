package main

import (
	"fmt"
	"time"
)

// GetWeekDays represents a tool to retrieve a list of 7 dates based on an input date.
type GetWeekDays struct {
	Date      string   // Input date in format YYYY-MM-DD HH:MM
	Out_Dates []string // Output list of 7 dates in YYYY-MM-DD HH:MM format, ordered starting from the input date and including the next 6 days
}

func (st *GetWeekDays) run(caller *ToolCaller, ui *UI) error {
	// Parse the input date string into a time.Time object
	t, err := time.Parse("2006-01-02 15:04", st.Date)
	if err != nil {
		return fmt.Errorf("failed to parse date: %w", err)
	}

	// Generate the list of 7 dates starting from the input date
	st.Out_Dates = make([]string, 7)
	for i := 0; i < 7; i++ {
		st.Out_Dates[i] = t.AddDate(0, 0, i).Format("2006-01-02 15:04")
	}

	return nil
}
