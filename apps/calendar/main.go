/*
Copyright 2023 Milan Suk

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this db except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

type Storage struct {
	ShowSide bool
	Mode     string

	Small_date int64
	Small_page int64

	event_page       int64
	event_start_date int64
	event_start_hour int
	event_start_min  int

	event_end_date int64
	event_end_hour int
	event_end_min  int

	event_title       string
	event_description string
	event_file        string
}

var store Storage

type Translations struct {
	YEAR      string
	MONTH     string
	DAY       string
	WEEK      string
	JANUARY   string
	FEBRUARY  string
	MARCH     string
	APRIL     string
	MAY       string
	JUNE      string
	JULY      string
	AUGUST    string
	SEPTEMBER string
	OCTOBER   string
	NOVEMBER  string
	DECEMBER  string
	MON       string
	TUE       string
	WED       string
	THU       string
	FRI       string
	SAT       string
	SUN       string

	MONDAY    string
	TUESDAY   string
	WEDNESDAY string
	THURSDAY  string
	FRIDAY    string
	SATURDAY  string
	SUNDAY    string

	NEW_EVENT string
	OK        string
	TODAY     string
	BETWEEN   string

	TITLE       string
	DESCRIPTION string
	FILE        string
	ADD_EVENT   string
	CANCEL      string

	BEGIN  string
	FINISH string

	EMPTY  string
	EDIT   string
	DELETE string
}

var trns Translations

func MonthText(month int) string {
	switch month {
	case 1:
		return trns.JANUARY
	case 2:
		return trns.FEBRUARY
	case 3:
		return trns.MARCH
	case 4:
		return trns.APRIL
	case 5:
		return trns.MAY
	case 6:
		return trns.JUNE
	case 7:
		return trns.JULY
	case 8:
		return trns.AUGUST
	case 9:
		return trns.SEPTEMBER
	case 10:
		return trns.OCTOBER
	case 11:
		return trns.NOVEMBER
	case 12:
		return trns.DECEMBER
	}
	return ""
}

func DayTextFull(day int) string {

	switch day {
	case 1:
		return trns.MONDAY
	case 2:
		return trns.TUESDAY
	case 3:
		return trns.WEDNESDAY
	case 4:
		return trns.THURSDAY
	case 5:
		return trns.FRIDAY
	case 6:
		return trns.SATURDAY
	case 7:
		return trns.SUNDAY
	}
	return ""
}

func DayTextShort(day int) string {

	switch day {
	case 1:
		return trns.MON
	case 2:
		return trns.TUE
	case 3:
		return trns.WED
	case 4:
		return trns.THU
	case 5:
		return trns.FRI
	case 6:
		return trns.SAT
	case 7:
		return trns.SUN
	}
	return ""
}

func GetWeekDayPure(unix_sec int64) int {
	date := time.Unix(unix_sec, 0)

	week := int(date.Weekday()) //sun=0, mon=1, etc.
	if week == 0 {
		week = 7
	}
	return week
}

func GetWeekDay(unix_sec int64, format float64) int {
	date := time.Unix(unix_sec, 0)

	week := int(date.Weekday()) //sun=0, mon=1, etc.
	if format != 1 {
		//not "us"
		week -= 1
		if week < 0 {
			week = 6
		}
	}
	return week
}

func GetStartWeekDay(unix_sec int64, format float64) time.Time {
	weekDay := GetWeekDay(unix_sec, format)

	date := time.Unix(unix_sec, 0)
	return date.AddDate(0, 0, -weekDay)
}

func GetYMD(unix_sec int64) (int, int, int) {
	date := time.Unix(unix_sec, 0)
	return date.Year(), int(date.Month()), date.Day()
}

func GetHM(unix_sec int64) (int, int) {
	date := time.Unix(unix_sec, 0)
	return date.Hour(), date.Minute()
}

func GetTextTime(unix_sec int64) string {
	hour, min := GetHM(unix_sec)
	return fmt.Sprintf("%d:%d", hour, min)
}

func GetTextDateTime(unix_sec int64) string {
	return GetTextDate(unix_sec) + " " + GetTextTime(unix_sec)
}

func GetMonthYear(unix_sec int64) string {
	tm := time.Unix(unix_sec, 0)
	return MonthText(int(tm.Month())) + " " + strconv.Itoa(tm.Year())
}

func GetYear(unix_sec int64) string {
	tm := time.Unix(unix_sec, 0)
	return strconv.Itoa(tm.Year())
}

func CmpDates(a int64, b int64) bool {
	ta := time.Unix(a, 0)
	tb := time.Unix(b, 0)

	return ta.Year() == tb.Year() && ta.Month() == tb.Month() && ta.Day() == tb.Day()
}

func GetTextDate(unix_sec int64) string {

	tm := time.Unix(unix_sec, 0)

	d := strconv.Itoa(tm.Day())
	m := strconv.Itoa(int(tm.Month()))
	y := strconv.Itoa(tm.Year())

	switch SA_InfoGetFloat("date", "", "") {
	case 0: //eu
		return d + "/" + m + "/" + y

	case 1: //us
		return m + "/" + d + "/" + y

	case 2: //iso
		return y + "-" + fmt.Sprintf("%02d", m) + "-" + fmt.Sprintf("%02d", d)

	case 3: //text
		//return month + " " + d + "," + y	//...

	case 4: //2base
		return y + fmt.Sprintf("%02d", m) + fmt.Sprintf("%02d", d)
	}

	return ""
}

func Calendar(value *int64, page *int64) {
	format := SA_InfoGetFloat("date", "", "")

	for x := 0; x < 7; x++ {
		SA_Col(x, 0.9)
		SA_ColMax(x, 2)
	}
	for y := 0; y < 6; y++ {
		SA_Row(y, 0.9)
		SA_RowMax(y, 2)
	}

	//fix page(need to start with day 1)
	{
		dtt := time.Unix(*page, 0)
		*page = dtt.AddDate(0, 0, -(dtt.Day() - 1)).Unix()
	}

	//--Day names(short)--
	if format == 1 {
		//"us"
		SA_TextCenter(DayTextShort(7)).Show(0, 0, 1, 1)
		for x := 1; x < 7; x++ {
			SA_TextCenter(DayTextShort(x)).Show(x, 0, 1, 1)
		}
	} else {
		for x := 1; x < 8; x++ {
			SA_TextCenter(DayTextShort(x)).Show(x-1, 0, 1, 1)
		}
	}

	//--Week days--
	now := int64(SA_Time())
	orig_month := time.Unix(*page, 0).Month()
	dtt := GetStartWeekDay(*page, format)
	for y := 0; y < 6; y++ {
		for x := 0; x < 7; x++ {
			//alpha := float64(1)
			//backCd := SA_ThemeCd()
			//frontCd := SA_ThemeBlack()

			isDayToday := CmpDates(dtt.Unix(), now)
			isDaySelected := CmpDates(dtt.Unix(), *value)
			isDayInMonth := dtt.Month() == orig_month

			style := &styles.ButtonAlpha

			if isDayToday {
				style = &g_ButtonToday
				//frontCd = SA_ThemeCd()
			}

			if isDaySelected && isDayInMonth { //selected day
				//alpha = 0 //show back
				//frontCd = SA_ThemeWhite()
				//backCd = SA_ThemeGrey(0.4)
				style = &g_ButtonSelect

				if isDayToday {
					style = &styles.Button
				}
			}

			if !isDayInMonth { //is day in current month
				//frontCd = SA_ThemeGrey(0.7)
				if isDaySelected {
					style = &g_ButtonOutsideMonthSelect
				} else {
					style = &g_ButtonOutsideMonth
				}
			}

			if SA_ButtonStyle(strconv.Itoa(dtt.Day()), style).Show(x, 1+y, 1, 1).click {
				*value = dtt.Unix()
				*page = *value
			}

			dtt = dtt.AddDate(0, 0, 1) //add day
		}
	}
}

func DateTimePicker(name string, date *int64, hour *int, minute *int) bool {

	SA_ColMax(0, 3)
	SA_ColMax(1, 15)

	SA_Text(name).Show(0, 0, 1, 1)

	//date
	if SA_Button(GetTextDate(*date)).Show(1, 0, 1, 1).click {
		SA_DialogOpen("DateTimePicker_"+name, 1)
		store.event_page = int64(SA_Time())
	}

	if SA_DialogStart("DateTimePicker_" + name) {
		Calendar(date, &store.event_page)
		SA_DialogEnd()
	}

	//time
	var err1 error
	var err2 error

	if *hour < 0 || *hour > 23 {
		err1 = errors.New(trns.BETWEEN + " 0 - 23")
	}
	if SA_Editbox(hour).TempToValue(true).Error(err1).Show(3, 0, 1, 1).finished {
		if *hour < 0 {
			*hour = 0
		}
		if *hour > 23 {
			*hour = 23
		}
	}

	SA_TextCenter(":").Show(4, 0, 1, 1)

	if *minute < 0 || *minute > 59 {
		err2 = errors.New(trns.BETWEEN + " 0 - 59")
	}
	if SA_Editbox(minute).TempToValue(true).Error(err2).Show(5, 0, 1, 1).finished {
		if *minute < 0 {
			*minute = 0
		}
		if *minute > 59 {
			*minute = 59
		}
	}

	return err1 == nil && err2 == nil
}

func EditEvent(rowid int64) {
	SA_ColMax(0, 15)

	//start date
	SA_DivStart(0, 0, 1, 1)
	startOk := DateTimePicker(trns.BEGIN, &store.event_start_date, &store.event_start_hour, &store.event_start_min)
	SA_DivEnd()

	//end date
	SA_DivStart(0, 1, 1, 1)
	endOk := DateTimePicker(trns.FINISH, &store.event_end_date, &store.event_end_hour, &store.event_end_min)
	SA_DivEnd()

	var errTitle error
	if len(store.event_title) == 0 {
		errTitle = errors.New(trns.EMPTY)
	}
	SA_Editbox(&store.event_title).TempToValue(true).Error(errTitle).ShowDescription(0, 2, 1, 1, trns.TITLE, 3, nil)
	SA_Editbox(&store.event_description).ShowDescription(0, 3, 1, 1, trns.DESCRIPTION, 3, nil)
	//SA_Editbox(&store.new_event_file).ShowDescription(0, 4, 1, 1, trns.FILE, 3, 0) //drag & drop ...

	SA_DivStart(0, 5, 1, 1)
	{
		SA_ColMax(0, 100)
		SA_ColMax(1, 100)

		bNm := trns.ADD_EVENT
		if rowid >= 0 {
			bNm = trns.EDIT
		}
		if SA_Button(bNm).Enable(startOk && endOk && errTitle == nil).Show(0, 0, 1, 1).click {
			store.event_start_date -= store.event_start_date % (24 * 3600) //round to begin of day
			store.event_end_date -= store.event_end_date % (24 * 3600)     //round to begin of day

			start := store.event_start_date + (int64(store.event_start_hour) * 3600) + (int64(store.event_start_min) * 60)
			end := store.event_end_date + (int64(store.event_end_hour) * 3600) + (int64(store.event_end_min) * 60)

			//send file into db - maybe hex()? ...
			if rowid >= 0 {
				SA_SqlWrite("", fmt.Sprintf("UPDATE events SET start=%d, end=%d, title='%s', description='%s' WHERE rowid=%d;", start, end, store.event_title, store.event_description, rowid))
			} else {
				SA_SqlWrite("", fmt.Sprintf("INSERT INTO events(start, end, title, description) VALUES(%d, %d, '%s', '%s');", start, end, store.event_title, store.event_description))
			}
			SA_DialogClose()
		}
		if SA_Button(trns.CANCEL).Show(1, 0, 1, 1).click {
			store.event_start_date = 0
			store.event_end_date = 0
			store.event_title = ""
			store.event_description = ""
			store.event_file = ""
			SA_DialogClose()
		}
	}
	SA_DivEnd()
}

func ShowEvent(rowid int64) {

	query := SA_SqlRead("", fmt.Sprintf("SELECT start, end, title, description FROM events WHERE rowid=%d", rowid))
	var start, end int64
	var title, description string
	if query.Next(&start, &end, &title, &description) {

		SA_ColMax(0, 10)

		SA_Text(GetTextDateTime(start)).ShowDescription(0, 0, 1, 1, trns.BEGIN, 3, nil)
		SA_Text(GetTextDateTime(end)).ShowDescription(0, 1, 1, 1, trns.FINISH, 3, nil)
		SA_Text(title).ShowDescription(0, 2, 1, 1, trns.TITLE, 3, nil)
		SA_Text(description).ShowDescription(0, 3, 1, 1, trns.DESCRIPTION, 3, nil)

		SA_DivStart(0, 4, 1, 1)
		{
			SA_ColMax(0, 4)
			SA_ColMax(1, 100)
			SA_ColMax(2, 2)
			if SA_Button(trns.EDIT).Show(0, 0, 1, 1).click {
				SA_DialogClose()
				SA_DialogOpen(fmt.Sprintf("eventEdit_%d", rowid), 1)
				store.event_start_date = start
				store.event_end_date = end
				store.event_start_hour = int(start%(24*3600)) / 3600
				store.event_start_min = int(start%3600) / 60
				store.event_end_hour = int(end%(24*3600)) / 3600
				store.event_end_min = int(end%3600) / 60

				store.event_title = title
				store.event_description = description
			}
			if SA_Button(trns.DELETE).Show(2, 0, 1, 1).click {
				SA_DialogClose()
				SA_DialogOpen(fmt.Sprintf("eventRemove_%d", rowid), 1)
			}
		}
		SA_DivEnd()
	}
}

func Side() {

	SA_ColMax(0, 100)

	SAPaint_Rect(0, 0, 1, 1, 0, SA_ThemeGrey(0.97), 0) //paintRect(color: themeGrey(0.97))

	if store.ShowSide {
		SA_Row(1, 0.3)
		SA_Row(2, 1.2)
		SA_Row(3, 6.5)
		SA_Row(4, 0.3)
		SA_RowMax(5, 100)

		//newEvent := false
		if SA_Button(trns.NEW_EVENT).Show(0, 0, 1, 1).click {
			//newEvent = true
			SA_DialogOpen("NewEvent", 0)

			//init
			store.event_start_date = int64(SA_Time())
			store.event_end_date = int64(SA_Time())
			store.event_start_hour, store.event_start_min = GetHM(int64(SA_Time()))
			store.event_end_hour = store.event_start_hour
			store.event_end_min = store.event_start_min
		}

		if SA_DialogStart("NewEvent") {
			EditEvent(-1)
			SA_DialogEnd()
		}

		SA_RowSpacer(0, 1, 1, 1)

		SA_DivStart(0, 2, 1, 1)
		{
			SA_ColMax(0, 100)
			SA_Text("##"+GetMonthYear(store.Small_page)).Show(0, 0, 1, 1)

			if SA_ButtonLight("<").Show(1, 0, 1, 1).click {
				tm := time.Unix(store.Small_page, 0)
				store.Small_page = tm.AddDate(0, -1, 0).Unix()
			}
			if SA_ButtonLight(">").Show(2, 0, 1, 1).click {
				tm := time.Unix(store.Small_page, 0)
				store.Small_page = tm.AddDate(0, 1, 0).Unix()
			}
		}
		SA_DivEnd()

		SA_DivStart(0, 3, 1, 1)
		Calendar(&store.Small_date, &store.Small_page)
		SA_DivEnd()

		SA_RowSpacer(0, 4, 1, 1)

		SA_DivStart(0, 6, 1, 1)
		{
			if SA_Button(">>").Show(0, 0, 2, 1).click {
				store.ShowSide = false
			}
		}
		SA_DivEnd()
	} else {
		SA_RowMax(0, 100)
		if SA_Button("<<").Show(0, 1, 1, 1).click {
			store.ShowSide = true
		}
	}
}

func ModeYear() {

	w := int(SA_DivInfo("screenWidth") / 7) //7cells is needed for one calendar
	if w == 0 {
		return
	}
	h := 12 / w
	if 12%w > 0 {
		h++
	}

	for x := 0; x < w*2; x += 2 {
		SA_Col(x, 6.5)

		SA_Col(x+1, 0.2)
		SA_ColMax(x+1, 100)
	}
	for y := 0; y < h*2; y += 2 {
		SA_Row(y, 7.5)
		SA_Row(y+1, 0.2)
	}

	year := time.Unix(store.Small_date, 0).Year()
	i := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if i < 12 {
				if SA_DivStart(x*2, y*2, 1, 1) {

					page := time.Date(year, time.Month(1+i), 1, 0, 0, 0, 0, time.Now().Location()).Unix()

					SA_ColMax(0, 100)
					SA_RowMax(1, 100)

					if SA_ButtonMenu("##"+MonthText(1+i)).Show(0, 0, 1, 1).click {

						//change month = i+1
						t := time.Unix(store.Small_date, 0)
						store.Small_date = time.Date(t.Year(), time.Month(i+1), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
						store.Mode = "month"
					}

					if SA_DivStart(0, 1, 1, 1) {
						Calendar(&store.Small_date, &page)
					}
					SA_DivEnd()

				}
				SA_DivEnd()
				i += 1
			}
		}
	}
}

func ModeMonth() {

	format := SA_InfoGetFloat("date", "", "")

	for x := 0; x < 7; x++ {
		SA_ColMax(x, 100)
	}
	for y := 0; y < 6; y++ {
		SA_RowMax(1+y, 100)
	}

	//--Day names(short)--
	if format == 1 {
		//"us"
		SA_TextCenter(DayTextShort(7)).Show(0, 0, 1, 1)
		for x := 1; x < 7; x++ {
			SA_TextCenter(DayTextShort(x)).Show(x, 0, 1, 1)
		}
	} else {
		for x := 1; x < 8; x++ {
			SA_TextCenter(DayTextShort(x)).Show(x-1, 0, 1, 1)
		}
	}

	//grid
	for y := 0; y < 6; y++ {
		SA_DivStart(0, 1+y, 7, 1)
		SAPaint_Line(0, 0, 1, 0, SA_ThemeGrey(0.75), 0.03)
		//paintLine(sy: 0, ey: 0, width: 0.03, color: themeGrey(0.75))
		SA_DivEnd()
	}

	for x := 0; x < 6; x++ {
		SA_DivStart(1+x, 1, 1, 6)
		SAPaint_Line(0, 0, 0, 1, SA_ThemeGrey(0.75), 0.03)
		//paintLine(sx: 0, ex: 0, width: 0.03, color: themeGrey(0.75))
		SA_DivEnd()
	}

	{

		orig_month := time.Unix(store.Small_date, 0).Month()

		//fix page(need to start with day 1)
		dtt := time.Unix(store.Small_date, 0)
		page := dtt.AddDate(0, 0, -(dtt.Day() - 1)).Unix()
		dtt = GetStartWeekDay(page, format)

		for y := 0; y < 6; y++ {
			for x := 0; x < 7; x++ {

				SA_DivStart(x, 1+y, 1, 1)
				{
					SA_ColMax(0, 100)
					SA_RowMax(1, 100)

					isToday := CmpDates(dtt.Unix(), int64(SA_Time()))
					if isToday {
						SAPaint_Rect(0, 0, 1, 1, 0.03, SA_ThemeWhite().Aprox(SA_ThemeCd(), 0.3), 0)
					}

					style := &styles.ButtonMenu
					if dtt.Month() != orig_month { //is day out of current month
						style = &styles.ButtonMenuSelected
					}
					if SA_ButtonStyle("##"+strconv.Itoa(dtt.Day())+".", style).Show(0, 0, 1, 1).click {
						store.Small_date = dtt.Unix()
						store.Mode = "day"
					}

					SA_DivStart(0, 1, 1, 1)
					{

						//fmt.Print(dtt)
						t := dtt.Unix()
						t -= t % (24 * 3600)
						query := SA_SqlRead("", fmt.Sprintf("SELECT rowid, start, end, title FROM events WHERE start >= %d AND start < %d ORDER BY start", t, t+(24*3600)))

						SA_DivSetInfo("scrollVnarrow", 1)
						//paintRect(borderWidth:0.03, margin: 0.1, color: themeGrey())
						SA_ColMax(0, 100)
						for i := int64(0); i < query.row_count; i++ {
							SA_Row(int(i), 0.7)
						}

						y := 0
						var rowid, start, end int64
						var title string
						for query.Next(&rowid, &start, &end, &title) {

							if SA_ButtonStyle(title, &g_ButtonEvent).Tooltip(title+": "+GetTextTime(start)+" - "+GetTextTime(end)).Show(0, y, 1, 1).click {
								SA_DialogOpen(fmt.Sprintf("eventDetail_%d", rowid), 1)
							}

							eventDialogs(rowid)

							y++
						}
					}
					SA_DivEnd()

				}
				SA_DivEnd()

				dtt = dtt.AddDate(0, 0, 1) //add day
			}
		}
	}
}

func eventDialogs(rowid int64) {

	//details
	if SA_DialogStart(fmt.Sprintf("eventDetail_%d", rowid)) {
		ShowEvent(rowid)
		SA_DialogEnd()
	}

	//edit
	if SA_DialogStart(fmt.Sprintf("eventEdit_%d", rowid)) {
		EditEvent(rowid)
		SA_DialogEnd()
	}

	//remove
	if SA_DialogStart(fmt.Sprintf("eventRemove_%d", rowid)) {
		if SA_DialogConfirm() {
			SA_SqlWrite("", fmt.Sprintf("DELETE FROM events WHERE rowid=%d;", rowid))
		}
		SA_DialogEnd()
	}
}

func ModeWeek() {

	format := SA_InfoGetFloat("date", "", "")

	SA_ColMax(0, 100)
	SA_RowMax(1, 100)

	//header
	dtt := GetStartWeekDay(store.Small_date, format)
	SA_DivStart(0, 0, 1, 1)
	{
		SA_Col(0, 1.5) //time
		for i := 1; i < 8; i++ {
			SA_ColMax(i, 100)
		}

		changeDay := -1
		if format == 1 {
			//"us"
			if SA_ButtonAlpha("##"+strconv.Itoa(dtt.Day())+". "+DayTextShort(7)).Show(1, 0, 1, 1).click {
				changeDay = dtt.Day()
			}

			dtt = dtt.AddDate(0, 0, 1) //add day

			for x := 1; x < 7; x++ {
				if SA_ButtonAlpha("##"+strconv.Itoa(dtt.Day())+". "+DayTextShort(x)).Show(1+x, 0, 1, 1).click {
					changeDay = dtt.Day()
				}
				dtt = dtt.AddDate(0, 0, 1) //add day
			}
		} else {
			for x := 1; x < 8; x++ {
				if SA_ButtonAlpha("##"+strconv.Itoa(dtt.Day())+". "+DayTextShort(x)).Show(x, 0, 1, 1).click {
					changeDay = dtt.Day()
				}
				dtt = dtt.AddDate(0, 0, 1) //add day
			}
		}

		if changeDay >= 0 {
			//change day = changeDay
			t := time.Unix(store.Small_date, 0)
			store.Small_date = time.Date(t.Year(), t.Month(), changeDay, 0, 0, 0, 0, t.Location()).Unix()

			store.Mode = "day"
		}

	}
	SA_DivEnd()

	//days
	SA_DivStart(0, 1, 1, 1)
	{
		SA_Col(0, 1.5) //time
		for i := 1; i < 8; i++ {
			SA_ColMax(i, 100)
		}
		for i := 0; i < 25; i++ {
			SA_Row(i*2, 0.3)
			SA_Row(i*2+1, 1.5)
		}

		//time
		for y := 0; y < 25; y++ {
			SA_Text("###"+strconv.Itoa(y)+":00").Show(0, y*2, 1, 1) //AlignV(0)
		}

		//grid
		for y := 0; y < 25; y++ {
			SA_DivStart(1, y*2, 7, 1)
			SAPaint_Line(0, 0.5, 1, 0.5, SA_ThemeGrey(0.75), 0.03)
			SA_DivEnd()
		}

		for x := 1; x < 7; x++ {
			SA_DivStart(1+x, 0, 1, 24*2+1)
			SAPaint_Line(0, 0, 0, 1, SA_ThemeGrey(0.75), 0.03)
			SA_DivEnd()
		}

		//events
		dtt = GetStartWeekDay(store.Small_date, format)
		for x := 0; x < 7; x++ {
			SA_DivStart(1+x, 0, 1, 24*2+1)
			DayEvent(dtt.Unix())
			SA_DivEnd()

			dtt = dtt.AddDate(0, 0, 1) //add day
		}

		//time-line
		w1 := GetStartWeekDay(int64(SA_Time()), format).Unix()
		w2 := GetStartWeekDay(store.Small_date, format).Unix()
		if CmpDates(w1, w2) { //today is in current week

			now := int64(SA_Time())
			hour, minute := GetHM(now)
			h := (float64(hour) + (float64(minute) / 60)) / 24
			week := GetWeekDay(now, format)

			SA_DivStart(1, 0, 7, 24*2+1)
			{
				SA_DivSetInfo("touch_enable", 0)
				SAPaint_Line(0, h, 1, h, SA_ThemeEdit(), 0.03)
				SAPaint_Circle(float64(week)/7, h, 0.1, SA_ThemeEdit(), 0)
			}
			SA_DivEnd()
		}

	}
	SA_DivEnd()
}

type EventItem struct {
	rowid, start, end, endVisual int64
	title                        string
}

func (a EventItem) HasCover(b EventItem) bool {

	//b is before
	if b.start < a.start && b.endVisual < a.start {
		return false
	}

	//b is after
	if b.start > a.endVisual && b.endVisual > a.endVisual {
		return false
	}

	return true
}

func DayEvent(t int64) {
	stT := t - (t % (24 * 3600))
	enT := stT + (24 * 3600)

	var cols [][]EventItem
	query := SA_SqlRead("", fmt.Sprintf("SELECT rowid, start, end, title FROM events WHERE start >= %d AND start < %d ORDER BY start", stT, enT))
	var item EventItem
	for query.Next(&item.rowid, &item.start, &item.end, &item.title) {

		item.endVisual = item.end
		if (item.end - item.start) < 3600/4 {
			item.endVisual = item.start + 3600/4
		}

		//find column
		fcol := 0
		for c := 0; c < len(cols); c++ {
			found := false
			for _, it := range cols[c] {
				if it.HasCover(item) {
					fcol++
					found = true
					break
				}
			}
			if !found {
				break
			}
		}

		//add
		if fcol == len(cols) {
			cols = append(cols, []EventItem{})
		}
		cols[fcol] = append(cols[fcol], item)
	}

	SA_RowMax(0, 100)
	for c := 0; c < len(cols); c++ {
		SA_ColMax(c, 100)
	}

	height := SA_DivInfo("layoutHeight") - 0.15
	for c := 0; c < len(cols); c++ {
		SA_DivStart(c, 0, 1, 1)

		{
			SA_ColMax(0, 100)
			last_end := float64(0)
			for i, it := range cols[c] {

				daySec := int64(24 * 3600)
				start := float64(it.start%daySec) / float64(daySec)
				end := float64(it.endVisual%daySec) / float64(daySec)
				if it.endVisual/daySec > it.start/daySec {
					end = 1
				}

				start *= height
				end *= height
				start += 0.15
				end += 0.15

				SA_Row(i*2+0, float64(start-last_end))
				SA_Row(i*2+1, float64(end-start))
				last_end = end
			}

			for i, it := range cols[c] {
				if SA_Button(it.title).Tooltip(it.title+": "+GetTextTime(it.start)+" - "+GetTextTime(it.end)).Show(0, i*2+1, 1, 1).click {
					SA_DialogOpen(fmt.Sprintf("eventDetail_%d", it.rowid), 1)
				}

				eventDialogs(it.rowid)
			}
		}
		SA_DivEnd()
	}

}

func ModeDay() {
	SA_ColMax(0, 100)
	SA_RowMax(1, 100)

	//header
	SA_DivStart(0, 0, 1, 1)
	{
		SA_Col(0, 1.5) //time
		SA_ColMax(1, 100)

		dtt := time.Unix(store.Small_date, 0)
		SA_Text("##"+strconv.Itoa(dtt.Day())+". "+DayTextFull(GetWeekDayPure(store.Small_date))).Show(1, 0, 1, 1)
	}
	SA_DivEnd()

	//days
	SA_DivStart(0, 1, 1, 1)
	{
		SA_Col(0, 1.5) //time
		SA_ColMax(1, 100)

		for y := 0; y < 25; y++ {
			SA_Row(y*2, 0.3)
			SA_Row(y*2+1, 1.5)
		}

		//time
		for y := 0; y < 25; y++ {
			SA_TextCenter("###"+strconv.Itoa(y)+":00").Show(0, y*2, 1, 1) //.AlignV(0)
		}

		//grid
		for y := 0; y < 25; y++ {
			SA_DivStart(1, y*2, 1, 1)
			SAPaint_Line(0, 0.5, 1, 0.5, SA_ThemeGrey(0.75), 0.03)
			SA_DivEnd()
		}

		//events
		SA_DivStart(1, 0, 1, 24*2+1)
		dtt := time.Unix(store.Small_date, 0)
		DayEvent(dtt.Unix())
		SA_DivEnd()

		//time-line
		if CmpDates(int64(SA_Time()), store.Small_date) { //today == day

			now := int64(SA_Time())
			hour, minute := GetHM(now)
			h := (float64(hour) + (float64(minute) / 60)) / 24
			SA_DivStart(1, 0, 1, 24*2+1)
			{
				SA_DivSetInfo("touch_enable", 0)
				SAPaint_Line(0, h, 1, h, SA_ThemeEdit(), 0.03)
				SAPaint_Circle(0, h, 0.1, SA_ThemeEdit(), 0)
			}
			SA_DivEnd()
		}

	}
	SA_DivEnd()
}

func ModePanel() {
	SA_ColMax(0, 100)
	SA_RowMax(1, 100)

	var title string
	if store.Mode == "year" {
		title = GetYear(store.Small_date)
		SA_DivStartName(0, 1, 1, 1, "year")
		ModeYear()
		SA_DivEnd()
	} else if store.Mode == "month" {
		title = GetMonthYear(store.Small_date)
		SA_DivStartName(0, 1, 1, 1, "month")
		ModeMonth()
		SA_DivEnd()
	} else if store.Mode == "week" {
		title = GetMonthYear(store.Small_date)
		SA_DivStartName(0, 1, 1, 1, "week")
		ModeWeek()
		SA_DivEnd()
	} else if store.Mode == "day" {
		title = GetTextDate(store.Small_date)
		SA_DivStartName(0, 1, 1, 1, "day")
		ModeDay()
		SA_DivEnd()
	}

	SA_DivStart(0, 0, 1, 1)
	{
		SA_ColMax(0, 2)
		SA_ColMax(3, 100)
		SA_ColMax(4, 8)

		//today
		if SA_ButtonLight(trns.TODAY).Tooltip(GetTextDate(int64(SA_Time()))).Show(0, 0, 1, 1).click {
			store.Small_date = int64(SA_Time())
			store.Small_page = int64(SA_Time())
		}

		//arrows
		if SA_ButtonLight("<").Show(1, 0, 1, 1).click {
			tm := time.Unix(store.Small_date, 0)
			if store.Mode == "year" {
				store.Small_date = tm.AddDate(-1, 0, 0).Unix()
			} else if store.Mode == "month" {
				store.Small_date = tm.AddDate(0, -1, 0).Unix()
			} else if store.Mode == "week" {
				store.Small_date = tm.AddDate(0, 0, -7).Unix()
			} else if store.Mode == "day" {
				store.Small_date = tm.AddDate(0, 0, -1).Unix()
			}
		}
		if SA_ButtonLight(">").Show(2, 0, 1, 1).click {
			tm := time.Unix(store.Small_date, 0)
			if store.Mode == "year" {
				store.Small_date = tm.AddDate(1, 0, 0).Unix()
			} else if store.Mode == "month" {
				store.Small_date = tm.AddDate(0, 1, 0).Unix()
			} else if store.Mode == "week" {
				store.Small_date = tm.AddDate(0, 0, 7).Unix()
			} else if store.Mode == "day" {
				store.Small_date = tm.AddDate(0, 0, 1).Unix()
			}
		}

		//title
		SA_TextCenter("##"+title).Show(3, 0, 1, 1)

		//Modes
		SA_DivStart(4, 0, 1, 1)
		{
			SAPaint_Rect(0, 0, 1, 1, 0, SA_ThemeGrey(0.9), 0)

			for i := 0; i < 4; i++ {
				SA_ColMax(i*2, 2)
				SA_Col(i*2+1, 0.2)
			}
			for i := 0; i < 4; i++ {
				if i < 3 {
					SA_ColSpacer(i*2+1, 0, 1, 1)
				}
			}

			if SA_ButtonAlpha(trns.DAY).Pressed(store.Mode == "day").Show(0, 0, 1, 1).click {
				store.Mode = "day"
			}
			if SA_ButtonAlpha(trns.WEEK).Pressed(store.Mode == "week").Show(2, 0, 1, 1).click {
				store.Mode = "week"
			}
			if SA_ButtonAlpha(trns.MONTH).Pressed(store.Mode == "month").Show(4, 0, 1, 1).click {
				store.Mode = "month"
			}
			if SA_ButtonAlpha(trns.YEAR).Pressed(store.Mode == "year").Show(6, 0, 1, 1).click {
				store.Mode = "year"
			}

			//paintRect(color: themeGrey(0.3), borderWidth: 0.03)
		}
		SA_DivEnd()

	}
	SA_DivEnd()

}

func Render() {

	if store.ShowSide {
		SA_Col(1, 6.3)
	}
	SA_ColMax(0, 100)
	SA_RowMax(0, 100)

	SA_DivStart(0, 0, 1, 1)
	ModePanel()
	SA_DivEnd()

	SA_DivStart(1, 0, 1, 1)
	Side()
	SA_DivEnd()

}

var styles SA_Styles
var g_ButtonSelect _SA_Style
var g_ButtonToday _SA_Style
var g_ButtonOutsideMonth _SA_Style
var g_ButtonOutsideMonthSelect _SA_Style
var g_ButtonEvent _SA_Style

func Open() {
	store.ShowSide = true
	store.Small_date = int64(SA_Time())
	store.Small_page = int64(SA_Time())

	//default
	json.Unmarshal(SA_File("storage_json"), &store)
	json.Unmarshal(SA_File("translations_json:app:resources/translations.json"), &trns)
	json.Unmarshal(SA_File("styles_json"), &styles)

	//styles
	g_ButtonSelect = styles.Button
	g_ButtonSelect.Main.Color = SA_ThemeWhite()
	g_ButtonSelect.Main.Content_color = SA_ThemeGrey(0.4)
	g_ButtonSelect.Id = 0

	g_ButtonToday = styles.ButtonAlpha
	g_ButtonToday.Main.Color = SA_ThemeCd()
	g_ButtonToday.Id = 0

	g_ButtonOutsideMonth = styles.ButtonAlpha
	g_ButtonOutsideMonth.Main.Color = SA_ThemeGrey(0.7)
	g_ButtonOutsideMonth.Id = 0

	g_ButtonOutsideMonthSelect = styles.Button
	g_ButtonOutsideMonthSelect.Main.Color = SA_ThemeGrey(0.7)
	g_ButtonOutsideMonthSelect.Id = 0

	g_ButtonEvent = styles.ButtonLight
	g_ButtonEvent.FontAlignH(0)
	g_ButtonEvent.Id = 0
}

func SetupDB() {
	SA_SqlWrite("", "CREATE TABLE IF NOT EXISTS events(title TEXT, description TEXT, start INTEGER, end INTEGER);")
}

func Save() []byte {
	js, _ := json.MarshalIndent(&store, "", "")
	return js
}
