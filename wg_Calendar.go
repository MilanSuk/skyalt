package main

import (
	"strconv"
	"time"
)

type Calendar struct {
	Value *int64
	Page  *int64

	changed func()
}

func (layout *Layout) AddCalendar(x, y, w, h int, value *int64, page *int64) *Calendar {
	props := &Calendar{Value: value, Page: page}
	layout._createDiv(x, y, w, h, "Calendar", props.Build, nil, nil)
	return props
}

func (st *Calendar) Build(layout *Layout) {

	format := layout.ui.sync.GetDateFormat()

	if st.Page == nil {
		st.Page = &layout.ui.datePage
	}

	st.setPage(*st.Page)

	for x := 0; x < 7; x++ {
		layout.SetColumn(x, 0.9, 100)
	}
	for y := 0; y < 7; y++ {
		layout.SetRow(y, 0.9, 100)
	}

	//Day names(short)
	if format == "us" {
		//"us"
		txt := layout.AddText(0, 0, 1, 1, layout.GetDayTextShort(7))
		txt.Align_h = 1
		for x := 1; x < 7; x++ {
			txt := layout.AddText(x, 0, 1, 1, layout.GetDayTextShort(x))
			txt.Align_h = 1
		}
	} else {
		for x := 1; x < 8; x++ {
			txt := layout.AddText(x-1, 0, 1, 1, layout.GetDayTextShort(x))
			txt.Align_h = 1
		}
	}

	//Week Days
	today := Date_GetDateFromTime(time.Now().Unix())
	value_dtt := Date_GetDateFromTime(*st.Value)
	curr_month := Date_GetDateFromTime(*st.Page).Month
	dt := Date_GetStartWeek(*st.Page, format)

	for y := range 6 {
		for x := range 7 {

			showBack := false
			fade := false //default

			dtt := Date_GetDateFromTime(dt)
			isDayToday := today.CmpYMD(&dtt)
			isDaySelected := value_dtt.CmpYMD(&dtt)
			isDayInMonth := (dtt.Month == curr_month)

			if isDaySelected && isDayInMonth { //selected day
				showBack = true
			}
			if !isDayInMonth { //is day in current month
				fade = true
			}

			Day := layout.AddButton(x, 1+y, 1, 1, strconv.Itoa(dtt.Day))
			dtClicked := dt
			Day.clicked = func() {
				*st.Value = dtClicked
				st.setPage(*st.Value)

				if st.changed != nil {
					st.changed()
				}
			}

			if showBack {
				Day.Background = 1
			} else {
				Day.Background = 0.1
			}
			Day.Cd_fade = fade
			Day.Border = isDayToday

			dt = Date_AddDate(dt, 0, 0, 1) //add day
		}
	}
}

func (st *Calendar) setPage(page int64) {
	//fix page(need to start with day 1)
	*st.Page = Date_GetStartMonth(page)
}

type SADate struct {
	Year, Month, Day     int
	Hour, Minute, Second int

	WeekDay        int //sun=0, mon=1, etc.
	ZoneOffset_sec int
	ZoneName       string
	IsDST          bool
}

func (a *SADate) CmpYMD(b *SADate) bool {
	return a.Year == b.Year && a.Month == b.Month && a.Day == b.Day
}

func (d *SADate) GetWeekDay(dateFormat string) int {
	week := d.WeekDay
	if dateFormat != "us" {
		//not "us"
		week -= 1
		if week < 0 {
			week = 6
		}
	}
	return week
}

func Date_GetDateFromTime(unixTime int64) SADate {
	tm := time.Unix(unixTime, 0)
	zoneName, zoneOffset_sec := tm.Zone()

	var d SADate
	d.Year = tm.Year()
	d.Month = int(tm.Month())
	d.Day = tm.Day()
	d.Hour = tm.Hour()
	d.Minute = tm.Minute()
	d.Second = tm.Second()

	d.WeekDay = int(tm.Weekday())

	d.ZoneOffset_sec = zoneOffset_sec
	d.ZoneName = zoneName
	d.IsDST = tm.IsDST()

	return d
}

func Date_GetTimeFromDate(date *SADate) int64 {
	tm := time.Date(date.Year, time.Month(date.Month), date.Day, date.Hour, date.Minute, date.Second, 0, time.Local)
	return tm.Unix()
}

func Date_GetStartMonth(unix_sec int64) int64 {
	d := Date_GetDateFromTime(unix_sec)
	d.Day = 1
	return Date_GetTimeFromDate(&d)
}

func Date_GetStartWeek(unix_sec int64, dateFormat string) int64 {
	unix_sec = Date_GetStartDay(unix_sec)

	d := Date_GetDateFromTime(unix_sec)
	weekDay := d.GetWeekDay(dateFormat) //možná dát dateFormat, také do SADate? ....

	return Date_AddDate(unix_sec, 0, 0, -weekDay)
}

func Date_GetStartDay(unix_sec int64) int64 {
	d := Date_GetDateFromTime(unix_sec)
	return unix_sec - int64(d.Hour)*3600 - int64(d.Minute)*60 - int64(d.Second)
}

func Date_AddDate(unixTime int64, add_years, add_months, add_days int) int64 {
	tm := time.Unix(unixTime, 0)
	return tm.AddDate(add_years, add_months, add_days).Unix()
}
