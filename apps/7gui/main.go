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
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Person struct {
	Name    string
	Surname string
}
type Circle struct {
	X, Y, Rad float64
}

type Storage struct {
	Count int

	Celsius    float64
	Fahrenheit float64

	StartDate     int64
	ReturnDate    int64
	ReturnFlight  int
	calendar_page int64

	StartTime float64
	MaxTime   float64

	People         []Person
	SelectedPerson int
	Name           string
	Surname        string
	Search         string

	Circles          []Circle
	snapshots        [][]Circle
	snapshots_pos    int
	circles_selected int

	Cells map[string]string
}

var store Storage

type Translations struct {
	COUNT      string
	CELSIUS    string
	FAHRENHEIT string

	YOU_HAVE_BOOKED string
	ONE_WAY_FLIGHT  string
	RETURN_FLIGHT   string
	FROM            string
	TO              string
	ON              string

	BOOK string

	ELAPSED_TIME string
	DURATION     string
	RESET_TIMER  string

	NAME    string
	SURNAME string

	CREATE string
	UPDATE string
	DELETE string
	SEARCH string

	UNDO string
	REDO string
}

var trns Translations

func Counter() {
	SA_ColMax(0, 100)
	SA_ColMax(1, 100)

	SA_Text("").ValueInt(store.Count).Show(0, 0, 1, 1)
	if SA_Button(trns.COUNT).Show(1, 0, 1, 1).click {
		store.Count++
	}
}

func TemperatureConverter() {
	SA_ColMax(0, 100)
	SA_ColMax(1, 100)
	SA_ColMax(3, 100)
	SA_ColMax(4, 100)

	SA_Editbox(&store.Celsius).TempToValue(true).AsNumber(true).Show(0, 0, 1, 1)
	if SA_Editbox(&store.Fahrenheit).TempToValue(true).AsNumber(true).Show(3, 0, 1, 1).changed {
		store.Celsius = (store.Fahrenheit - 32) * (5 / 9.0)
	} else {
		store.Fahrenheit = store.Celsius*(9/5.0) + 32
	}

	SA_Text(trns.CELSIUS).Show(1, 0, 1, 1)
	SA_Text("=").Show(2, 0, 1, 1)
	SA_Text(trns.FAHRENHEIT).Show(4, 0, 1, 1)
}

func FlightBookerDialog() {
	SA_ColMax(0, 15)

	startDate := Format(store.StartDate)
	returnDate := Format(store.ReturnDate)

	text := trns.YOU_HAVE_BOOKED + " "
	if store.ReturnFlight > 0 {
		text += trns.RETURN_FLIGHT + " " + trns.FROM + " " + startDate + " " + trns.TO + " " + returnDate
	} else {
		text += trns.ONE_WAY_FLIGHT + " " + trns.ON + " " + startDate
	}

	SA_TextCenter(text).Show(0, 0, 1, 1)
}

func FlightBooker() {

	SA_ColMax(0, 100)

	//init
	if store.StartDate == 0 {
		store.StartDate = int64(SA_Time())
	}
	if store.ReturnDate == 0 {
		store.ReturnDate = int64(SA_Time())
	}

	SA_Combo(&store.ReturnFlight, trns.ONE_WAY_FLIGHT+"|"+trns.RETURN_FLIGHT).Show(0, 0, 1, 1)

	SA_DivStart(0, 1, 1, 1)
	SA_ColMax(0, 100)
	CalendarButton("Calendar_Start", &store.StartDate, &store.calendar_page, true)
	SA_DivEnd()

	SA_DivStart(0, 2, 1, 1)
	SA_ColMax(0, 100)
	CalendarButton("Calendar_Return", &store.ReturnDate, &store.calendar_page, store.ReturnFlight > 0)
	SA_DivEnd()

	if SA_Button(trns.BOOK).Show(0, 3, 1, 1).click {
		SA_DialogOpen("FlightBookerDialog", 1)
	}

	if SA_DialogStart("FlightBookerDialog") {
		FlightBookerDialog()
		SA_DialogEnd()
	}
}

func Timer() {

	SA_ColMax(0, 5)
	SA_ColMax(1, 100)

	//init
	if store.StartTime == 0 {
		store.StartTime = SA_Time()
	}
	if store.MaxTime == 0 {
		store.MaxTime = 30
	}

	//time diff
	dt := SA_Time() - store.StartTime
	if dt > store.MaxTime {
		dt = store.MaxTime
	}
	if dt < store.MaxTime {
		SA_InfoSet("nosleep", "", "", "")
	}

	SA_Text(trns.ELAPSED_TIME).Show(0, 0, 1, 1)
	SA_Text(trns.DURATION).Show(0, 2, 1, 1)

	SA_Progress(float64(dt)/float64(store.MaxTime), 1).Show(1, 0, 1, 1)
	SA_Text(strconv.FormatFloat(dt, 'f', 1, 64)+"sec").Show(1, 1, 1, 1)
	SA_Slider(&store.MaxTime).Min(0.1).Max(120).Jump(1).Show(1, 2, 1, 1)

	if SA_Button(trns.RESET_TIMER).Show(1, 3, 1, 1).click {
		store.StartTime = SA_Time()
	}
}

func CRUDList(search string) {
	SA_ColMax(0, 100)
	y := 0

	for i, person := range store.People {
		if len(search) == 0 || strings.Contains(strings.ToLower(person.Name), strings.ToLower(search)) || strings.Contains(strings.ToLower(person.Surname), strings.ToLower(search)) {
			if SA_ButtonMenu(person.Surname+", "+person.Name).Pressed(i == store.SelectedPerson).Show(0, y, 1, 1).click {
				store.SelectedPerson = i
				store.Surname = person.Surname
				store.Name = person.Name
			}
			y += 1
		}
	}

}

func CRUDChange() {
	SA_ColMax(0, 100)

	SA_Text(store.Surname+", "+store.Name).Show(0, 0, 1, 1)
	SA_Editbox(&store.Name).TempToValue(true).ShowDescription(0, 1, 1, 1, trns.NAME, 2, nil)
	SA_Editbox(&store.Surname).TempToValue(true).ShowDescription(0, 2, 1, 1, trns.SURNAME, 2, nil)
}

func CRUDButtons() {
	SA_ColMax(0, 100)
	SA_ColMax(1, 100)
	SA_ColMax(2, 100)

	if SA_Button(trns.CREATE).Show(0, 0, 1, 1).click {
		store.People = append(store.People, Person{})
		store.SelectedPerson = len(store.People) - 1
		store.Name = ""
		store.Surname = ""
	}

	if SA_Button(trns.UPDATE).Enable(store.SelectedPerson >= 0 && (store.Surname != store.People[store.SelectedPerson].Surname || store.Name != store.People[store.SelectedPerson].Name)).Show(1, 0, 1, 1).click {
		selected := &store.People[store.SelectedPerson]
		selected.Surname = store.Surname
		selected.Name = store.Name
	}

	if SA_Button(trns.DELETE).Enable(store.SelectedPerson >= 0).Show(2, 0, 1, 1).click {
		store.People = append(store.People[:store.SelectedPerson], store.People[store.SelectedPerson+1:]...)
		store.SelectedPerson = -1
	}

}

func CRUD() {
	SA_ColMax(0, 100)
	SA_ColMax(1, 100)
	SA_RowMax(1, 100)

	SA_Editbox(&store.Search).Ghost(trns.SEARCH).TempToValue(true).Show(0, 0, 3, 1)

	SA_DivStart(0, 1, 1, 1)
	CRUDList(store.Search)
	SA_DivEnd()

	if store.SelectedPerson >= len(store.People) {
		store.SelectedPerson = -1
	}

	if store.SelectedPerson >= 0 {
		SA_DivStart(1, 1, 1, 1)
		CRUDChange()
		SA_DivEnd()
	}

	SA_DivStart(0, 2, 3, 1)
	CRUDButtons()
	SA_DivEnd()

}

func CircleDrawerAddSnapShot() {
	//cut later
	if store.snapshots_pos < len(store.snapshots) {
		store.snapshots = store.snapshots[:store.snapshots_pos+1]
	}

	//copy all circles
	tmp := make([]Circle, len(store.Circles))
	copy(tmp, store.Circles)

	//append copy into snapshots array
	store.snapshots = append(store.snapshots, tmp)
	store.snapshots_pos++
}

func CircleDrawerCanvas() {

	touch_x := SA_DivInfo("touchX")
	touch_y := SA_DivInfo("touchY")
	width := SA_DivInfo("layoutWidth")
	height := SA_DivInfo("layoutHeight")

	touch_clicked := SA_DivInfo("touchInside") > 0 && SA_DivInfo("touchEnd") > 0

	//find circle inside radius
	closest_i := -1
	minDist := 10000000.0
	for i, it := range store.Circles {
		rx := (it.X - touch_x) * width
		ry := (it.Y - touch_y) * height
		dist := math.Sqrt(rx*rx + ry*ry)
		if dist < it.Rad && dist < minDist {
			closest_i = i
			minDist = dist
		}
	}

	//draw circles + open dialog if clicked
	for i, it := range store.Circles {
		if closest_i == i {
			//highlight
			selectCd := SA_ThemeGrey(0.5)
			selectCd.A = 120
			SAPaint_Circle(it.X, it.Y, it.Rad, selectCd, 0) //select
			SAPaint_Cursor("hand")

			if touch_clicked {
				store.circles_selected = i
				SA_DialogOpen("CirclesDialog", 2) //if I use name as prefix, I don't need 'store.circles_selected' ...
			}
		}
		SAPaint_Circle(it.X, it.Y, it.Rad, SA_ThemeBlack(), 0.03)
	}

	//dialog(resize circle)
	if SA_DialogStart("CirclesDialog") {
		circle := &store.Circles[store.circles_selected]
		SA_ColMax(0, 5)
		if SA_Slider(&circle.Rad).Min(0.1).Max(3).Jump(0.1).Show(0, 0, 1, 1).finished {
			CircleDrawerAddSnapShot()
		}
		SA_TextCenter("").ValueFloat(circle.Rad, 1).Show(1, 0, 1, 1)

		SA_DialogEnd()
	}

	//create new one
	if touch_clicked && closest_i < 0 {
		store.Circles = append(store.Circles, Circle{X: touch_x, Y: touch_y, Rad: 0.75})
		CircleDrawerAddSnapShot()
	}

	//border
	SAPaint_Rect(0, 0, 1, 1, 0, SA_ThemeBlack(), 0.03)

}

func CircleDrawer() {
	SA_ColMax(0, 100)
	SA_ColMax(1, 4)
	SA_ColMax(3, 4)
	SA_ColMax(4, 100)
	SA_RowMax(1, 5)

	if len(store.snapshots) == 0 {
		CircleDrawerAddSnapShot()
		store.snapshots_pos = 0
	}

	if SA_Button(trns.UNDO).Enable(store.snapshots_pos != 0).Show(1, 0, 1, 1).click {
		store.snapshots_pos--
		store.Circles = store.snapshots[store.snapshots_pos]
	}

	SA_TextCenter(strconv.Itoa(store.snapshots_pos+1)+"/"+strconv.Itoa(len(store.snapshots))).Show(2, 0, 1, 1)

	if SA_Button(trns.REDO).Enable(store.snapshots_pos+1 != len(store.snapshots)).Show(3, 0, 1, 1).click {
		store.snapshots_pos++
		store.Circles = store.snapshots[store.snapshots_pos]
	}

	SA_DivStart(0, 1, 5, 1)
	CircleDrawerCanvas()
	SA_DivEnd()
}

func Cells() {

	n_cols := int(('Z' - 'A') + 1)
	n_rows := 100

	for c := 0; c <= n_cols; c++ {
		SA_ColResize(1+c, 2)
	}

	for r := 0; r <= n_rows; r++ {
		SA_RowResize(r+1, 1)
	}

	//columns header
	for c := 0; c < n_cols; c++ {
		SA_TextStyle(string(rune('A'+c)), &g_TextCellHeader).Show(1+c, 0, 1, 1)
	}

	//rows header
	stRow := int(SA_DivInfo("startRow"))
	enRow := int(SA_DivInfo("endRow"))
	if enRow > n_rows {
		enRow = n_rows
	}
	for r := stRow; r <= enRow; r++ {
		if r > 0 {
			SA_TextStyle(strconv.Itoa(r-1), &g_TextCellHeader).Show(0, r, 1, 1)
		}
	}

	//content
	for r := stRow; r <= enRow; r++ {
		if r > 0 {
			for c := 0; c <= n_cols; c++ {
				if c > 0 {
					id := fmt.Sprintf("%d %d", c-1, r-1)
					v := store.Cells[id]
					if SA_EditboxStyle(&v, &g_EditboxNoBorder).Show(c, r, 1, 1).finished {
						if len(v) > 0 {
							store.Cells[id] = v
						} else {
							delete(store.Cells, id)
						}
					}
				}
			}
		}
	}

	//bug: header scroll ouside of screen ...
	//bug: resizer should be only in header(not content) ...
	//todo: formulas ...
}

func Render() {

	SA_ColMax(0, 100)
	SA_Col(1, 10)
	SA_ColMax(1, 100)
	SA_ColMax(2, 100)

	y := 1
	SA_Row(0, 0.5)

	n := 1
	SA_Text("Counter()").Show(0, y, 1, 1)
	SA_DivStart(1, y, 1, n)
	Counter()
	SA_DivEnd()
	SA_RowSpacer(0, y+n, 3, 1)
	y += n + 1

	n = 1
	SA_Text("TemperatureConverter()").Show(0, y, 1, 1)
	SA_DivStart(1, y, 1, n)
	TemperatureConverter()
	SA_DivEnd()
	SA_RowSpacer(0, y+n, 3, 1)
	y += n + 1

	n = 4
	SA_Text("FlightBooker()").Show(0, y, 1, 1)
	SA_DivStart(1, y, 1, n)
	FlightBooker()
	SA_DivEnd()
	SA_RowSpacer(0, y+n, 3, 1)
	y += n + 1

	n = 4
	SA_Text("Timer()").Show(0, y, 1, 1)
	SA_DivStart(1, y, 1, n)
	Timer()
	SA_DivEnd()
	SA_RowSpacer(0, y+n, 3, 1)
	y += n + 1

	n = 6
	SA_Text("CRUD()").Show(0, y, 1, 1)
	SA_DivStart(1, y, 1, n)
	CRUD()
	SA_DivEnd()
	SA_RowSpacer(0, y+n, 3, 1)
	y += n + 1

	n = 6
	SA_Text("CircleDrawer()").Show(0, y, 1, 1)
	SA_DivStart(1, y, 1, n)
	CircleDrawer()
	SA_DivEnd()
	SA_RowSpacer(0, y+n, 3, 1)
	y += n + 1

	n = 8
	SA_Text("Cells()").Show(0, y, 1, 1)
	SA_DivStart(1, y, 1, n)
	Cells()
	SA_DivEnd()
	//SA_RowSpacer(0, y+n, 3, 1)
	//y += n + 1
}

var styles SA_Styles
var g_TextCellHeader _SA_Style
var g_EditboxNoBorder _SA_Style

func Open() {
	store.Cells = make(map[string]string)
	store.SelectedPerson = -1

	//default
	json.Unmarshal(SA_File("storage_json"), &store)
	json.Unmarshal(SA_File("translations_json:app:resources/translations.json"), &trns)
	json.Unmarshal(SA_File("styles_json"), &styles)

	//styles
	g_TextCellHeader = styles.Text
	g_TextCellHeader.ContentColor(SA_ThemeGrey(0.9))

	g_EditboxNoBorder = styles.Editbox
	g_EditboxNoBorder.Border(0)

	//storage
	if len(store.People) == 0 {
		store.People = append(store.People, Person{Surname: "Emil", Name: "Hans"})
		store.People = append(store.People, Person{Surname: "Mustermann", Name: "Max"})
		store.People = append(store.People, Person{Surname: "Tisch", Name: "Roman"})
		store.People = append(store.People, Person{Surname: "Romba", Name: "John"})
	}

	InitCalendar()
}

func SetupDB() {
}

func Save() []byte {
	js, _ := json.MarshalIndent(&store, "", "")
	return js
}
