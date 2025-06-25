package main

import "time"

// Returns 7 dates which are ordered list of days in the week including 'Date' parameter.
type GetWeekDays struct {
	Date string //Date format: YYYY-MM-DD HH:MM

	Out_days []string
}

func (st *GetWeekDays) run(caller *ToolCaller, ui *UI) error {
	dt, err := time.ParseInLocation("2006-01-02 15:04", st.Date, time.Local)
	if err != nil {
		return err
	}

	//go to first day of the week
	firstDay := time.Monday
	if UI_GetDateFormat() == "us" {
		firstDay = time.Sunday
	}
	for dt.Weekday() != firstDay {
		dt = dt.AddDate(0, 0, -1)
	}

	st.Out_days = nil
	for i := 0; i < 7; i++ {
		st.Out_days = append(st.Out_days, dt.Format("Mon, 02 Jan 2006"))
		dt = dt.AddDate(0, 0, 1)
	}

	return nil
}
