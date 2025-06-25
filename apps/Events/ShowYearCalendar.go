package main

// Show calendar's Year view.
type ShowYearCalendar struct {
	Year int
}

func (st *ShowYearCalendar) run(caller *ToolCaller, ui *UI) error {
	ui.SetColumn(0, 1, 100)
	ui.SetRow(0, 1, 100)

	ui.AddYearCalendar(0, 0, 1, 1, st.Year)

	return nil
}
