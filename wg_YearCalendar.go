package main

import (
	"fmt"
	"strconv"
	"time"
)

type YearCalendar struct {
	Year int
}

func (layout *Layout) AddYearCalendar(x, y, w, h int, Year int) *YearCalendar {
	props := &YearCalendar{Year: Year}
	layout._createDiv(x, y, w, h, "YearCalendar", props.Build, nil, nil)
	return props
}

func (st *YearCalendar) Build(layout *Layout) {

	layout.SetColumn(0, 1, 100)
	layout.SetRowFromSub(1, 5, 100)

	layout.AddText(0, 0, 1, 1, fmt.Sprintf("<b>Year: %d</b>", st.Year)).Align_h = 1

	list := layout.AddLayoutList(0, 1, 1, 1, true)

	today := time.Now()

	for i := 0; i < 12; i++ {
		//this_i := i
		Item := list.AddListSubItem()
		space := 0.5
		Item.LLMTip = fmt.Sprintf("%s, %d", layout.GetMonthText(1+i), st.Year)
		Item.SetColumn(0, 0.2, space)
		Item.SetColumn(1, 1, 6.5)
		Item.SetColumn(2, 0.2, space)
		Item.SetRow(0, 0.2, space)
		Item.SetRow(1, 1, 7.5)
		Item.SetRow(2, 0.2, space)

		Div := Item.AddLayout(1, 1, 1, 1)
		Div.SetColumn(0, 1, 100)
		Div.SetRow(1, 1, 100)

		Month := Div.AddText(0, 0, 1, 1, "<h2>"+layout.GetMonthText(1+i))
		if int(today.Month()) == i+1 {
			Month.Value = "<b>" + Month.Value
		}

		_formMonthCalendar(st.Year, i+1, Div.AddLayout(0, 1, 1, 1))
	}
}

func _formMonthCalendar(year, month int, layout *Layout) {

	layout.LLMTip = fmt.Sprintf("%s, %d", layout.GetMonthText(month), year)

	for x := 0; x < 7; x++ {
		layout.SetColumn(x, 0.9, 100)
	}
	for y := 0; y < 7; y++ {
		layout.SetRow(y, 0.9, 100)
	}

	// Day names(short)
	if layout.ui.sync.GetDateFormat() == "us" {
		//"us"
		txt, txtLay := layout.AddText2(0, 0, 1, 1, layout.GetDayTextShort(7))
		txt.Align_h = 1
		for x := 1; x < 7; x++ {
			txt := layout.AddText(x, 0, 1, 1, layout.GetDayTextShort(x))
			txt.Align_h = 1
			txtLay.scrollH.Show = false
		}
		txtLay.scrollH.Show = false

	} else {
		for x := 1; x < 8; x++ {
			txt, txtLay := layout.AddText2(x-1, 0, 1, 1, layout.GetDayTextShort(x))
			txt.Align_h = 1
			txtLay.scrollH.Show = false
		}
	}

	// Week Days
	today := time.Now()
	dt := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	{
		firstDay := time.Monday
		if layout.ui.sync.GetDateFormat() == "us" {
			firstDay = time.Sunday
		}
		for dt.Weekday() != firstDay {
			dt = dt.AddDate(0, 0, -1)
		}
	}

	for y := range 6 {
		for x := range 7 {

			Day, DayLay := layout.AddText2(x, 1+y, 1, 1, strconv.Itoa(dt.Day()))
			DayLay.LLMTip = dt.Format("Mon, 02 Jan 2006 15:04")
			Day.Align_h = 1

			fade := int(dt.Month()) != month
			if fade { //is day in current month
				Day.Cd.A = 100 //fade
			}

			if !fade && today.Year() == dt.Year() && today.Month() == dt.Month() && today.Day() == dt.Day() {
				//highlight today
				Day.Value = "<b>" + Day.Value
				DayLay.Back_cd = layout.GetPalette().S
			}

			dt = dt.AddDate(0, 0, 1) //add day
		}
	}
}
