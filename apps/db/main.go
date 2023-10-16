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
	"strings"
)

type Storage struct {
	Tables        []*Table
	SelectedTable int

	renameTable  string
	renameColumn string
	createTable  string
	createColumn string

	showFilterDialog bool

	calendar_page int64
}

var store Storage

type Translations struct {
	NO_TABLES    string
	CREATE_TABLE string
	RENAME       string
	REMOVE       string
	DUPLICATE    string

	ALREADY_EXISTS string
	EMPTY_FIELD    string
	INVALID_NAME   string

	COLUMNS  string
	SHOW_ALL string
	HIDE_ALL string

	FILTER string
	SORT   string

	ENABLE      string
	NAME        string
	ROWS_HEIGHT string

	AND string
	OR  string

	TEXT      string
	INTEGER   string
	REAL      string
	BLOB      string
	CHECK_BOX string
	DATE      string
	PERCENT   string
	RATING    string

	MAX_STARS         string
	DECIMAL_PRECISION string

	HIDE string

	ADD_ROW string

	MIN   string
	MAX   string
	AVG   string
	SUM   string
	COUNT string
}

var trns Translations

type FilterItem struct {
	Column string
	Op     int
	Value  string
}

func (f *FilterItem) GetOpString() string {

	switch f.Op {
	case 0:
		return "="
	case 1:
		return "!="
	case 2:
		return "<="
	case 3:
		return ">="
	case 4:
		return "<"
	case 5:
		return ">"
	}
	return ""
}
func Filter_getOptions() string {
	return "=|<>|<=|>=|<|>"
}

type Filter struct {
	Enable bool
	Items  []*FilterItem
	Rel    int
}

func (f *Filter) UpdateColumn(old string, new string) {
	for _, it := range f.Items {
		if it.Column == old {
			it.Column = new
		}
	}
}

func (f *Filter) Add(columnName string, op int) {
	f.Items = append(f.Items, &FilterItem{Column: columnName, Op: op})
}

func (f *Filter) Check() {
	//toto smaže neplatné, ale možná by se měli přejmenovat na ""
}

type SortItem struct {
	Column string
	Az     int
}
type Sort struct {
	Enable bool
	Items  []*SortItem
}

func (s *Sort) UpdateColumn(old string, new string) {
	for _, it := range s.Items {
		if it.Column == old {
			it.Column = new
		}
	}
}

func (s *Sort) Find(columnName string) *SortItem {
	for _, it := range s.Items {
		if it.Column == columnName {
			return it
		}
	}
	return nil
}
func (s *Sort) Add(columnName string, az int) bool {

	if len(columnName) == 0 || s.Find(columnName) == nil {
		s.Items = append(s.Items, &SortItem{Column: columnName, Az: az})
		return true
	}
	return false
}

func (s *Sort) Check() {

}

type Column struct {
	Name   string
	Type   string
	Show   bool
	Resize float64

	Render string //checkbox, etc.

	Prop_rating_max_stars int
	Prop_percent_floats   int

	StatFunc string
}

func (col *Column) isRowId() bool {
	return col.Name == "rowid"
}

type Table struct {
	Name    string
	Columns []*Column

	Filter  Filter
	Sort    Sort
	RowSize int //0=>1row, 1=2rows

	scrollDown bool
}

func (table *Table) UpdateColumn(old string, new string) {
	table.Filter.UpdateColumn(old, new)
	table.Sort.UpdateColumn(old, new)
}

func GetDbStructure() []*Table {
	var tables []*Table

	qt := SA_SqlRead("", "SELECT name FROM sqlite_master WHERE type='table'")
	var tname string
	for qt.Next(&tname) {

		//table
		table := Table{Name: tname}
		table.Filter.Enable = true
		table.Sort.Enable = true
		// rowid column
		table.Columns = append(table.Columns, &Column{Name: "rowid", Type: "INT"})

		//column
		qc := SA_SqlRead("", "pragma table_info("+tname+");")
		var cid int
		var cname, ctype string
		for qc.Next(&cid, &cname, &ctype) {
			resize := float64(4)
			if cname == "rowid" {
				resize = 1
			}
			table.Columns = append(table.Columns, &Column{Name: cname, Type: ctype, Show: true, Resize: resize})
		}

		tables = append(tables, &table)
	}

	return tables
}

func FindTable(tables []*Table, tname string) *Table {
	for _, tb := range tables {
		if tb.Name == tname {
			return tb
		}
	}
	return nil
}

func (table *Table) FindColumn(cname string) *Column {
	for _, cl := range table.Columns {
		if cl.Name == cname {
			return cl
		}
	}
	return nil
}

func UpdateTables() {

	db := GetDbStructure()

	//add tables
	for _, db_tb := range db {
		if FindTable(store.Tables, db_tb.Name) == nil {
			store.Tables = append(store.Tables, db_tb)
		}
	}

	//add columns
	for _, table := range store.Tables {

		db_tb := FindTable(db, table.Name)
		if db_tb != nil {
			for _, db_cl := range db_tb.Columns {
				column := table.FindColumn(db_cl.Name)
				if column == nil {
					column = db_cl
					table.Columns = append(table.Columns, column)
				}
				column.Type = db_cl.Type
			}
		}
	}

	//remove tables/Columns
	for ti := len(store.Tables) - 1; ti >= 0; ti-- {
		table := store.Tables[ti]

		db_tb := FindTable(db, table.Name)
		if db_tb == nil {
			store.Tables = append(store.Tables[:ti], store.Tables[ti+1:]...) //remove table
			continue
		}

		for ci := len(table.Columns) - 1; ci >= 0; ci-- {
			column := table.Columns[ci]

			db_cl := db_tb.FindColumn(column.Name)
			if db_cl == nil {
				table.Columns = append(table.Columns[:ci], table.Columns[ci+1:]...) //remove column
			}
		}
	}
	if store.SelectedTable >= len(store.Tables) {
		store.SelectedTable = 0
	}

	//fix Columns
	for _, table := range store.Tables {
		for _, column := range table.Columns {
			if column.isRowId() {
				column.Show = true
			}
		}
	}

	//fix filter/short
	for _, table := range store.Tables {
		table.Filter.Check()
		table.Sort.Check()
	}
}

func DragAndDropTable(dst int) {
	SA_Div_SetDrag("table", uint64(dst))
	src, pos, done := SA_Div_IsDrop("table", false, true, false)
	if done {
		selTable := store.Tables[store.SelectedTable]
		SA_MoveElement(&store.Tables, &store.Tables, int(src), dst, pos)

		for i, tb := range store.Tables {
			if tb == selTable {
				store.SelectedTable = i
			}
		}
	}
}

func DragAndDropColumn(dst int, table *Table) {
	SA_Div_SetDrag("column", uint64(dst))
	src, pos, done := SA_Div_IsDrop("column", false, true, false)
	if done {
		SA_MoveElement(&table.Columns, &table.Columns, int(src), dst, pos)
	}
}

func TablesList() {
	SA_DivInfoSet("scrollHnarrow", 1)
	SA_DivInfoSet("scrollVshow", 0)

	for x := range store.Tables {
		SA_Col(x, 3)
		SA_ColMax(x, 5)
	}

	for x, table := range store.Tables {
		SA_DivStart(x, 0, 1, 1)
		{
			SA_ColMax(0, 5)

			isSelected := (store.SelectedTable == x)

			if SA_ButtonMenu(table.Name).Icon("app:resources/table.png", 0.13).Pressed(isSelected).Show(0, 0, 1, 1).click {
				store.SelectedTable = x
				if isSelected {
					SA_DialogOpen("TableMenu_"+table.Name, 1)
				}
			}

			DragAndDropTable(x)

			if SA_DialogStart("TableMenu_" + table.Name) {
				SA_ColMax(0, 5)
				SA_Row(2, 0.3)

				if SA_ButtonMenu(trns.RENAME).Show(0, 0, 1, 1).click {
					store.renameTable = table.Name
					SA_DialogClose()
					SA_DialogOpen("RenameTable_"+table.Name, 1)
				}

				if SA_ButtonMenu(trns.DUPLICATE).Show(0, 1, 1, 1).click {
					store.renameTable = table.Name
					SA_DialogClose()
					SA_DialogOpen("DuplicateTable_"+table.Name, 1)
				}

				//space
				SA_RowSpacer(0, 2, 1, 1)

				if SA_ButtonDangerMenu(trns.REMOVE).Show(0, 3, 1, 1).click {
					SA_DialogClose()
					SA_DialogOpen("RemoveTableConfirm_"+table.Name, 1)
				}

				SA_DialogEnd()
			}

			if SA_DialogStart("RenameTable_" + table.Name) {
				RenameTable(table)
				SA_DialogEnd()
			}

			if SA_DialogStart("DuplicateTable_" + table.Name) {
				DuplicateTable(table)
				SA_DialogEnd()
			}

			if SA_DialogStart("RemoveTableConfirm_" + table.Name) {
				if SA_DialogConfirm() {
					SA_SqlWrite("", "DROP TABLE "+table.Name+";")
				}

				SA_DialogEnd()
			}

		}
		SA_DivEnd()
	}

}

func CheckName(name string, alreadyExist bool) error {

	empty := len(name) == 0

	name = strings.ToLower(name)
	invalidName := !empty && (name[0] < 'a' || name[0] > 'z') //first must be a-z

	var err error
	if alreadyExist {
		err = errors.New(trns.ALREADY_EXISTS)
	} else if empty {
		err = errors.New(trns.EMPTY_FIELD)
	} else if invalidName {
		err = errors.New(trns.INVALID_NAME)
	}

	return err
}

func CreateTable() {
	SA_ColMax(0, 9)

	err := CheckName(store.createTable, FindTable(store.Tables, store.createTable) != nil)

	SA_Editbox(&store.createTable).Error(err).TempToValue(true).ShowDescription(0, 0, 1, 1, trns.NAME, 2, nil)

	if SA_Button(trns.CREATE_TABLE).Enable(err == nil).Show(0, 1, 1, 1).click {
		SA_SqlWrite("", "CREATE TABLE "+store.createTable+"(column TEXT DEFAULT '' NOT NULL);")
		SA_DialogClose()
	}
}

func RenameTable(table *Table) {
	SA_ColMax(0, 7)
	SA_ColMax(1, 3)

	err := CheckName(store.renameTable, FindTable(store.Tables, store.renameTable) != nil)

	SA_Editbox(&store.renameTable).Error(err).TempToValue(true).Show(0, 0, 1, 1)

	if SA_Button(trns.RENAME).Enable(err == nil).Show(1, 0, 1, 1).click {
		if table.Name != store.renameTable {
			SA_SqlWrite("", "ALTER TABLE "+table.Name+" RENAME TO "+store.renameTable+";")
		}
		table.Name = store.renameTable
		SA_DialogClose()
	}
}

func DuplicateTable(table *Table) {
	SA_ColMax(0, 7)
	SA_ColMax(1, 3)

	err := CheckName(store.renameTable, FindTable(store.Tables, store.renameTable) != nil)

	SA_Editbox(&store.renameTable).Error(err).TempToValue(true).Show(0, 0, 1, 1)

	if SA_Button(trns.DUPLICATE).Enable(err == nil).Show(1, 0, 1, 1).click {

		SA_SqlWrite("", "CREATE TABLE "+store.renameTable+" AS SELECT * FROM "+table.Name+";")

		//bug: columns,sorts,filters are pointers -> need to do deep copy ...
		//copy := *FindTable(store.Tables, table.Name)
		//store.Tables = append(store.Tables, &copy)

		table.Name = store.renameTable //select
		SA_DialogClose()
	}
}

func TopHeader() {
	SA_ColMax(1, 100)
	SA_Col(2, 2)

	if SA_Button("+").Tooltip(trns.CREATE_TABLE).Show(0, 0, 1, 1).click {
		SA_DialogOpen("CreateTable", 1)
	}
	if SA_DialogStart("CreateTable") {
		CreateTable()
		SA_DialogEnd()
	}

	if len(store.Tables) == 0 {
		SA_Text(trns.NO_TABLES).Show(1, 0, 1, 1)
	} else {
		SA_DivStart(1, 0, 1, 1)
		TablesList()
		SA_DivEnd()
	}

}

func Reorder[T any](x, y, w, h int, group string, id int, array []T) {

	SA_DivStart(x, y, w, h)
	{
		SA_Div_SetDrag(group, uint64(id))
		src, pos, done := SA_Div_IsDrop(group, true, false, false)
		if done {
			SA_MoveElement(&array, &array, int(src), id, pos)
		}
		SA_Image("app:resources/reorder.png").Margin(0.15).Show(0, 0, 1, 1)
	}
	SA_DivEnd()
}

func TableView(table *Table) {

	SA_ColMax(0, 100)
	SA_RowMax(1, 100)

	SA_DivStart(0, 0, 1, 1)
	{

		SA_ColMax(0, 5)

		//filter
		SA_Col(1, 0.5)
		SA_ColMax(2, 5)

		//sort
		SA_Col(3, 0.5)
		SA_ColMax(4, 5)

		//rows height
		SA_Col(5, 0.5)
		SA_ColMax(6, 4)

		hidden := false
		for _, col := range table.Columns {
			if !col.Show {
				hidden = true
			}
		}

		if SA_ButtonBorder(trns.COLUMNS).Pressed(hidden).Show(0, 0, 1, 1).click {
			SA_DialogOpen("Columns", 1)
		}

		if SA_DialogStart("Columns") {

			SA_ColMax(0, 5)
			SA_ColMax(1, 5)
			y := 0
			for i, col := range table.Columns {
				if col.isRowId() {
					continue
				}

				SA_DivStart(0, y, 2, 1)
				{
					SA_ColMax(1, 100)

					Reorder(0, 0, 1, 1, "column2", i, table.Columns)
					SA_Checkbox(&col.Show, col.Name).Show(1, 0, 1, 1)

					y++
				}
				SA_DivEnd()
			}

			if SA_Button(trns.SHOW_ALL).Show(0, y, 1, 1).click {
				for _, col := range table.Columns {
					col.Show = true
				}
			}

			if SA_Button(trns.HIDE_ALL).Show(1, y, 1, 1).click {
				for _, col := range table.Columns {
					if !col.isRowId() {
						col.Show = false
					}
				}
			}

			SA_DialogEnd()
		}

		if SA_ButtonBorder(trns.FILTER).Pressed(table.Filter.Enable && len(table.Filter.Items) > 0).Show(2, 0, 1, 1).click || store.showFilterDialog {
			store.showFilterDialog = false
			SA_DialogOpen("Filter", 1)
		}

		if SA_DialogStart("Filter") {

			SA_ColMax(0, 2)
			SA_ColMax(1, 6)
			SA_ColMax(2, 4)

			//enable
			y := 0
			SA_Checkbox(&table.Filter.Enable, trns.ENABLE).Show(0, y, 2, 1)

			//and/or
			SA_Combo(&table.Filter.Rel, trns.AND+"|"+trns.OR).Enable(table.Filter.Enable).Search(true).Show(2, y, 1, 1)
			y++

			for fi, it := range table.Filter.Items {

				SA_DivStart(0, y, 3, 1)
				{
					SA_ColMax(1, 5)
					SA_ColMax(2, 2)
					SA_ColMax(3, 3)

					if table.Filter.Enable {
						Reorder(0, 0, 1, 1, "filter", fi, table.Filter.Items)
					}

					SA_DivStart(1, 0, 1, 1)
					ColumnsCombo(table, &it.Column, table.Filter.Enable)
					SA_DivEnd()

					SA_Combo(&it.Op, Filter_getOptions()).Enable(table.Filter.Enable).Search(true).Show(2, 0, 1, 1)

					SA_Editbox(&it.Value).Enable(table.Filter.Enable).Show(3, 0, 1, 1)

					if SA_Button("X").Enable(table.Filter.Enable).Show(4, 0, 1, 1).click {
						table.Filter.Items = append(table.Filter.Items[:fi], table.Filter.Items[fi+1:]...) //remove
						break
					}
				}
				SA_DivEnd()

				y++
			}

			if SA_ButtonLight("+").Enable(table.Filter.Enable).Show(0, y, 1, 1).click {
				table.Filter.Add("", 0)
			}

			SA_DialogEnd()
		}

		if SA_ButtonBorder(trns.SORT).Pressed(table.Sort.Enable && len(table.Sort.Items) > 0).Show(4, 0, 1, 1).click {
			SA_DialogOpen("Sort", 1)
		}

		if SA_DialogStart("Sort") {

			SA_ColMax(2, 7)

			y := 0
			SA_Checkbox(&table.Sort.Enable, trns.ENABLE).Show(0, y, 3, 1)
			y++

			for si, it := range table.Sort.Items {

				SA_DivStart(0, y, 3, 1)
				{
					SA_ColMax(1, 5)
					SA_ColMax(2, 2)

					if table.Sort.Enable {
						Reorder(0, 0, 1, 1, "sort", si, table.Sort.Items)
					}

					SA_DivStart(1, 0, 1, 1)
					ColumnsCombo(table, &it.Column, table.Sort.Enable)
					SA_DivEnd()

					SA_Combo(&it.Az, "A -> Z|Z -> A").Enable(table.Sort.Enable).Show(2, 0, 1, 1)

					if SA_Button("X").Enable(table.Sort.Enable).Show(3, 0, 1, 1).click {
						table.Sort.Items = append(table.Sort.Items[:si], table.Sort.Items[si+1:]...) //remove
						break
					}
				}
				SA_DivEnd()

				y++
			}

			if SA_ButtonLight("+").Enable(table.Sort.Enable).Show(0, y, 2, 1).click {
				table.Sort.Add("", 0)
			}

			SA_DialogEnd()
		}

		SA_Combo(&table.RowSize, "1|2|3|4").ShowDescription(6, 0, 1, 1, trns.ROWS_HEIGHT, 2.5, nil)

	}
	SA_DivEnd()

	SA_DivStart(0, 1, 1, 1)
	Tablee(table)
	SA_DivEnd()
}

func ColumnsCombo(table *Table, selectedColumn *string, enable bool) {
	SA_ColMax(0, 100)

	pos := -1
	var opts string
	for i, col := range table.Columns {
		opts += col.Name + "|"
		if *selectedColumn == col.Name {
			pos = i
		}
	}
	opts, _ = strings.CutSuffix(opts, "|")

	var err error
	if pos < 0 {
		err = errors.New("Column not exist")
	}
	if SA_Combo(&pos, opts).Search(true).Error(err).Enable(enable).Show(0, 0, 1, 1) {
		*selectedColumn = table.Columns[pos].Name
	}
}

func IsInteger(tp string) bool {
	return tp == "INT" || tp == "INTEGER" || tp == "TINYINT" || tp == "SMALLINT" || tp == "MEDIUMINT" || tp == "BIGINT" || tp == "INT2" || tp == "INT8"
}
func IsFloat(tp string) bool {
	return tp == "REAL" || tp == "DOUBLE" || tp == "DOUBLE PRECISION" || tp == "FLOAT"
}
func IsText(tp string) bool {
	if tp == "TEXT" || tp == "CLOB" {
		return true
	}
	if strings.Index(tp, "CHARACTER") == 0 || strings.Index(tp, "VARCHAR") == 0 || strings.Index(tp, "NCHAR") == 0 || strings.Index(tp, "NVARCHAR") == 0 {
		return true
	}
	return false
}

func IsBlob(tp string) bool {
	return tp == "BLOB"
}

func IsDate(tp string) bool {
	return tp == "DATE" || tp == "DATETIME"
}

func GetColumnDefault(tp string) string {

	if IsInteger(tp) {
		return "0"
	}
	if IsFloat(tp) {
		return "0"
	}
	if IsText(tp) {
		return "''"
	}
	return "" //blob
}

func (column *Column) GetColumnName() string {
	nm := column.Type
	if IsInteger(column.Type) || IsFloat(column.Type) {
		switch column.Render {
		case "":
			if IsInteger(column.Type) {
				nm = trns.INTEGER
			} else if IsFloat(column.Type) {
				nm = trns.REAL
			}
		case "PERCENT":
			nm = trns.PERCENT
		case "CHECK_BOX":
			nm = trns.CHECK_BOX
		case "DATE":
			nm = trns.DATE
		case "RATING":
			nm = trns.RATING
		}
	} else {
		switch nm {
		case "TEXT":
			nm = trns.TEXT
		case "BLOB":
			nm = trns.BLOB
		default:
			if IsDate(nm) {
				nm = trns.DATE
			}
		}

	}
	return nm
}

func (column *Column) Convert(table *Table, colType string, renderType string) {

	addCol := fmt.Sprintf("ALTER TABLE %s ADD COLUMN __skyalt_temp_column__ %s;", table.Name, colType)
	copyCol := fmt.Sprintf("UPDATE %s SET __skyalt_temp_column__ = CAST(%s as %s);", table.Name, column.Name, colType)
	delCol := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;", table.Name, column.Name)
	renameCol := fmt.Sprintf("ALTER TABLE %s RENAME COLUMN __skyalt_temp_column__ TO %s;", table.Name, column.Name)

	SA_SqlWrite("", addCol+copyCol+delCol+renameCol)

	column.Type = colType
	column.Render = renderType
}

func ColumnDetail(table *Table, column *Column) {

	SA_ColMax(0, 10)
	SA_Row(1, 0.5)
	SA_Row(3, 0.5)
	SA_Row(7, 0.5)

	SA_DivStart(0, 0, 1, 1)
	{
		SA_ColMax(0, 100)
		SA_ColMax(1, 4)

		//rename
		origName := column.Name
		if SA_Editbox(&column.Name).ShowDescription(0, 0, 1, 1, trns.NAME, 2, nil).finished {
			if origName != column.Name {
				SA_SqlWrite("", "ALTER TABLE "+table.Name+" RENAME COLUMN "+origName+" TO "+column.Name+";")
			}

			//update filter/short
			table.UpdateColumn(origName, column.Name)
		}

		//convert type
		if SA_ButtonStyle(column.GetColumnName(), &g_ButtonLeft).Icon("app:resources/"+_getColumnIcon(column.Type, column.Render), 0.13).Enable(!IsBlob(column.Type)).Show(1, 0, 1, 1).click {
			SA_DialogOpen("changeType", 1)
		}
		if SA_DialogStart("changeType") {

			SA_ColMax(0, 5)

			if IsText(column.Type) {
				y := 0

				if SA_ButtonMenu(trns.INTEGER).Icon("app:resources/"+_getColumnIcon("INT", ""), 0.13).Enable(column.Render != "PERCENT").Show(0, y, 1, 1).click {
					column.Convert(table, "INT", "")
				}
				y++

				if SA_ButtonMenu(trns.REAL).Icon("app:resources/"+_getColumnIcon("REAL", ""), 0.13).Enable(column.Render != "PERCENT").Show(0, y, 1, 1).click {
					column.Convert(table, "REAL", "")
				}
				y++

			} else if IsFloat(column.Type) {
				y := 0

				if SA_ButtonMenu(trns.REAL).Icon("app:resources/"+_getColumnIcon(column.Type, ""), 0.13).Enable(column.Render != "").Show(0, y, 1, 1).click {
					column.Render = ""
				}
				y++

				if SA_ButtonMenu(trns.PERCENT).Icon("app:resources/"+_getColumnIcon(column.Type, "PERCENT"), 0.13).Enable(column.Render != "PERCENT").Show(0, y, 1, 1).click {
					column.Render = "PERCENT"
				}
				y++

				SA_RowSpacer(0, y, 1, 1)
				y++

				if SA_ButtonMenu(trns.INTEGER).Icon("app:resources/"+_getColumnIcon("INT", ""), 0.13).Enable(column.Render != "PERCENT").Show(0, y, 1, 1).click {
					column.Convert(table, "INT", "")
				}
				y++

				if SA_ButtonMenu(trns.TEXT).Icon("app:resources/"+_getColumnIcon("TEXT", ""), 0.13).Enable(column.Render != "PERCENT").Show(0, y, 1, 1).click {
					column.Convert(table, "TEXT", "")
				}
				y++

			} else if IsInteger(column.Type) {
				y := 0

				if SA_ButtonMenu(trns.INTEGER).Icon("app:resources/"+_getColumnIcon(column.Type, ""), 0.13).Enable(column.Render != "").Show(0, y, 1, 1).click {
					column.Render = ""
				}
				y++

				if SA_ButtonMenu(trns.CHECK_BOX).Icon("app:resources/"+_getColumnIcon(column.Type, "CHECK_BOX"), 0.13).Enable(column.Render != "CHECK_BOX").Show(0, y, 1, 1).click {
					column.Render = "CHECK_BOX"
				}
				y++

				if SA_ButtonMenu(trns.DATE).Icon("app:resources/"+_getColumnIcon(column.Type, "DATE"), 0.13).Enable(column.Render != "DATE").Show(0, y, 1, 1).click {
					column.Render = "DATE"
				}
				y++

				if SA_ButtonMenu(trns.RATING).Icon("app:resources/"+_getColumnIcon(column.Type, "RATING"), 0.13).Enable(column.Render != "RATING").Show(0, y, 1, 1).click {
					column.Render = "RATING"
					if column.Prop_rating_max_stars == 0 {
						column.Prop_rating_max_stars = 5
					}
				}
				y++

				SA_RowSpacer(0, y, 1, 1)
				y++

				if SA_ButtonMenu(trns.REAL).Icon("app:resources/"+_getColumnIcon("REAL", ""), 0.13).Enable(column.Render != "PERCENT").Show(0, y, 1, 1).click {
					column.Convert(table, "REAL", "")
				}
				y++

				if SA_ButtonMenu(trns.TEXT).Icon("app:resources/"+_getColumnIcon("TEXT", ""), 0.13).Enable(column.Render != "PERCENT").Show(0, y, 1, 1).click {
					column.Convert(table, "TEXT", "")
				}
				y++

			}
			SA_DialogEnd()
		}

	}
	SA_DivEnd()

	//sort/filter
	SA_DivStart(0, 2, 1, 1)
	{
		SA_ColMax(0, 100)
		SA_ColMax(1, 100)
		SA_ColMax(2, 100)

		SA_RowMax(0, 100)

		//filter
		if SA_ButtonStyle(trns.FILTER, &g_ButtonLightLeft).Icon("app:resources/filter.png", 0.13).Show(0, 0, 1, 1).click {
			table.Sort.Add(column.Name, 0)

			table.Filter.Add(column.Name, 0)

			store.showFilterDialog = true
			SA_DialogClose()
		}

		//sort
		sort_notUse := table.Sort.Find(column.Name) == nil
		if SA_ButtonStyle(trns.SORT, &g_ButtonLightLeft).Icon("app:resources/sort_az.png", 0.13).Enable(sort_notUse).Show(1, 0, 1, 1).click {
			table.Sort.Add(column.Name, 0)
		}

		//sort
		if SA_ButtonStyle(trns.SORT, &g_ButtonLightLeft).Icon("app:resources/sort_za.png", 0.13).Enable(sort_notUse).Show(2, 0, 1, 1).click {
			table.Sort.Add(column.Name, 1)
		}
	}
	SA_DivEnd()

	//properties
	SA_DivStart(0, 4, 1, 3)
	{
		if column.Render == "RATING" {
			SA_ColMax(0, 100)
			SA_Editbox(&column.Prop_rating_max_stars).ShowDescription(0, 0, 1, 1, trns.MAX_STARS, 4, nil)
		}
		if column.Render == "PERCENT" {
			SA_ColMax(0, 100)
			SA_Editbox(&column.Prop_percent_floats).ShowDescription(0, 0, 1, 1, trns.DECIMAL_PRECISION, 4, nil)
		}
	}
	SA_DivEnd()

	SA_DivStart(0, 8, 1, 1)
	{
		SA_ColMax(0, 100)
		SA_ColMax(1, 100)
		SA_ColMax(2, 100)

		//hide
		if SA_ButtonLight(trns.HIDE).Show(0, 0, 1, 1).click {
			column.Show = false
			SA_DialogClose()
		}

		//duplicate
		if SA_ButtonLight(trns.DUPLICATE).Show(1, 0, 1, 1).click {
			store.renameColumn = column.Name
			SA_DialogOpen("DuplicateColumn"+column.Name, 1)
		}

		//remove
		if SA_ButtonDanger(trns.REMOVE).Show(2, 0, 1, 1).click {
			SA_DialogOpen("RemoveColumnConfirm", 1)
		}

		if SA_DialogStart("DuplicateColumn" + column.Name) {
			SA_ColMax(0, 7)
			SA_ColMax(1, 3)

			SA_Editbox(&store.renameColumn).Show(0, 0, 1, 1) //err ...
			if SA_Button(trns.DUPLICATE).Show(1, 0, 1, 1).click {

				var add_def string
				defValue := GetColumnDefault(column.Type)
				if len(defValue) > 0 {
					add_def = "DEFAULT " + defValue + " NOT NULL"
				}

				SA_SqlWrite("", "ALTER TABLE "+table.Name+" ADD "+store.renameColumn+" "+column.Type+" "+add_def+";")
				SA_SqlWrite("", "UPDATE "+table.Name+" SET "+store.renameColumn+" = "+column.Name+";")

				copy := *column
				copy.Name = store.renameColumn
				table.Columns = append(table.Columns, &copy)
				SA_DialogClose()
			}

			SA_DialogEnd()
		}

		if SA_DialogStart("RemoveColumnConfirm") {
			if SA_DialogConfirm() {
				SA_SqlWrite("", "ALTER TABLE "+table.Name+" DROP COLUMN "+column.Name+";")
				SA_DialogClose()
			}
			SA_DialogEnd()
		}
	}
	SA_DivEnd()
}

func Tablee(table *Table) {

	sumWidth := 1.5 //"+"
	for _, col := range table.Columns {
		if col.Show {
			sumWidth += float64(col.Resize)
		}
	}

	SA_Col(0, sumWidth)
	SA_RowMax(1, 100)

	//columns header
	SA_DivStart(0, 0, 1, 1)
	TableColumns(table)
	SA_DivEnd()

	//rows
	SA_DivStart(0, 1, 1, 1)
	TableRows(table)
	SA_DivEnd()

	// add row + column stats
	SA_DivStart(0, 2, 1, 1)
	TableStats(table)
	SA_DivEnd()
}

func _getColumnIcon(tp string, render string) string {
	if IsText(tp) {
		return "type_text.png"
	}
	if IsInteger(tp) {
		switch render {
		case "":
			return "type_number.png"
		case "CHECK_BOX":
			return "type_checkbox.png"
		case "DATE":
			return "type_date.png"
		case "RATING":
			return "type_rating.png"
		}
	}
	if IsFloat(tp) {
		switch render {
		case "":
			return "type_number.png"
		case "PERCENT":
			return "type_percent.png"
		}
	}
	if IsBlob(tp) {
		return "type_blob.png"
	}

	return ""
}

func TableColumns(table *Table) {
	x := 0
	for _, col := range table.Columns {
		if !col.Show {
			continue
		}
		SA_Col(x, 1.5) //minimum
		col.Resize = SA_ColResizeName(x, col.Name, col.Resize)
		x++
	}
	SA_Col(x, 1) //"+"

	x = 0
	for _, col := range table.Columns {
		if !col.Show {
			continue
		}

		nm := col.Name
		if col.isRowId() {
			nm = "#"
		}

		SA_DivStart(x, 0, 1, 1)
		{
			SA_ColMax(0, 100)

			if col.isRowId() {
				SA_TextCenter(nm).Show(0, 0, 1, 1)
			} else {
				if SA_ButtonStyle(nm, &g_ButtonColumnHeader).Icon("app:resources/"+_getColumnIcon(col.Type, col.Render), 0.13).Show(0, 0, 1, 1).click && !col.isRowId() {
					SA_DialogOpen("columnDetail_"+nm, 1)
				}

				DragAndDropColumn(x, table)
			}
		}
		SA_DivEnd()

		if SA_DialogStart("columnDetail_" + nm) {
			ColumnDetail(table, col)
			SA_DialogEnd()
		}

		x++
	}

	//create column
	if SA_ButtonLight("+").Show(x, 0, 1, 1).click {
		SA_DialogOpen("createColumn", 1)
	}

	if SA_DialogStart("createColumn") {

		SA_ColMax(0, 5)
		y := 0
		add_type := ""
		//defValue := ""
		render := ""

		//name
		err := CheckName(store.createColumn, table.FindColumn(store.createColumn) != nil)
		SA_Editbox(&store.createColumn).Error(err).TempToValue(true).ShowDescription(0, y, 1, 1, trns.NAME, 2, nil)
		y++

		//types
		if SA_ButtonMenu(trns.TEXT).Icon("app:resources/"+_getColumnIcon("TEXT", ""), 0.13).Enable(err == nil).Show(0, y, 1, 1).click {
			add_type = "TEXT"
		}
		y++

		if SA_ButtonMenu(trns.INTEGER).Icon("app:resources/"+_getColumnIcon("INT", ""), 0.13).Enable(err == nil).Show(0, y, 1, 1).click {
			add_type = "INT"
		}
		y++

		if SA_ButtonMenu(trns.REAL).Icon("app:resources/"+_getColumnIcon("REAL", ""), 0.13).Enable(err == nil).Show(0, y, 1, 1).click {
			add_type = "REAL"
		}
		y++

		if SA_ButtonMenu(trns.BLOB).Icon("app:resources/"+_getColumnIcon("BLOB", ""), 0.13).Enable(err == nil).Show(0, y, 1, 1).click {
			add_type = "BLOB"
		}
		y++

		if SA_ButtonMenu(trns.CHECK_BOX).Icon("app:resources/"+_getColumnIcon("INT", "CHECK_BOX"), 0.13).Enable(err == nil).Show(0, y, 1, 1).click {
			add_type = "INT"
			render = "CHECK_BOX"
		}
		y++

		if SA_ButtonMenu(trns.DATE).Icon("app:resources/"+_getColumnIcon("INT", "DATE"), 0.13).Enable(err == nil).Show(0, y, 1, 1).click {
			add_type = "INT"
			render = "DATE"
		}
		y++

		if SA_ButtonMenu(trns.PERCENT).Icon("app:resources/"+_getColumnIcon("REAL", "PERCENT"), 0.13).Enable(err == nil).Show(0, y, 1, 1).click {
			add_type = "REAL"
			render = "PERCENT"
		}
		y++

		if SA_ButtonMenu(trns.RATING).Icon("app:resources/"+_getColumnIcon("INT", "RATING"), 0.13).Enable(err == nil).Show(0, y, 1, 1).click {
			add_type = "INT"
			render = "RATING"

		}
		y++

		if len(add_type) > 0 {
			var add_def string
			defValue := GetColumnDefault(add_type)
			if len(defValue) > 0 {
				add_def = "DEFAULT " + defValue + " NOT NULL"
			}

			SA_SqlWrite("", "ALTER TABLE "+table.Name+" ADD "+store.createColumn+" "+add_type+" "+add_def+";")

			if len(render) > 0 {
				column := &Column{Name: store.createColumn, Type: add_type, Show: true, Resize: 4, Render: render}
				table.Columns = append(table.Columns, column)
				//others will copy 'render' from here

				if render == "RATING" {
					column.Prop_rating_max_stars = 5
				}

				if render == "PERCENT" {
					column.Prop_percent_floats = 2
				}

			}

			store.createColumn = ""
			SA_DialogClose()
		}

		SA_DialogEnd()
	}

}

func TableRows(table *Table) {
	var count int
	{
		query := GetQueryCount(table)
		q := SA_SqlRead("", query)
		q.Next(&count)
	}

	SA_DivInfoSet("scrollOnScreen", 1)
	if table.scrollDown {
		SA_DivInfoSet("scrollVpos", 100000000)
		table.scrollDown = false
	}

	rowSize := table.RowSize + 1

	SA_ColMax(0, 100)
	SA_Row(0, float64(count*rowSize))
	SA_DivStart(0, 0, 1, 1)
	{
		SA_ColMax(0, 100)

		st, en := SA_DivRangeVer(float64(rowSize))

		query, ncols := GetQueryBasic(table)
		var stat *SA_Sql
		if len(query) > 0 {
			stat = SA_SqlRead("", query)
			if stat != nil {
				stat.row_i = uint64(st)
			}
		}
		values := make([]string, ncols)
		args := make([]interface{}, ncols)
		for i := range values {
			args[i] = &values[i]
		}

		for st < en && stat.Next(args...) {

			SA_DivStart(0, st*rowSize, 1, rowSize)
			{
				//columns sizes
				x := 0
				for _, col := range table.Columns {
					if col.Show {
						SA_Col(x, col.Resize)
						x++
					}
				}

				x = 0
				for _, col := range table.Columns {
					if !col.Show {
						continue
					}

					writeCell := false
					if col.isRowId() {

						if SA_ButtonLight(values[x]).Show(0, 0, 1, rowSize).click {
							SA_DialogOpen("RowId_"+values[x], 1)
						}

						if SA_DialogStart("RowId_" + values[x]) {
							SA_ColMax(0, 5)
							SA_Row(1, 0.5)

							if SA_ButtonMenu(trns.DUPLICATE).Show(0, 0, 1, 1).click {

								SA_SqlWrite("", fmt.Sprintf("INSERT INTO %s SELECT * FROM %s WHERE rowId=%s;", table.Name, table.Name, values[x]))

								SA_DialogClose()
							}

							SA_RowSpacer(0, 1, 1, 1)

							if SA_ButtonDangerMenu(trns.REMOVE).Show(0, 2, 1, 1).click {
								SA_SqlWrite("", "DELETE FROM "+table.Name+" WHERE rowid="+values[x]+";")
								SA_DialogClose()
							}

							SA_DialogEnd()
						}

					} else if IsBlob(col.Type) {

						r, err := strconv.Atoi(values[x])
						if r > 0 && err == nil {

							res := fmt.Sprintf("dbs::%s/%s/%d", table.Name, col.Name, r)
							SA_DivStart(x, 0, 1, rowSize)
							{
								SAPaint_File(0, 0, 1, 1, res, "", 0.03, 0, 0, SA_ThemeWhite(), 1, 1, false)
								SAPaint_Rect(0, 0, 1, 1, 0, SA_ThemeGrey(0.3), 0.03)

								inside := SA_DivInfoGet("touchInside") > 0
								end := SA_DivInfoGet("touchEnd") > 0
								if r > 0 && inside {
									SAPaint_Cursor("hand")
								}
								if r > 0 && inside && end {
									SA_DialogOpen("Image_"+values[x], 1)
								}

								if SA_DialogStart("Image_" + values[x]) {
									SA_ColMax(0, 15)
									SA_RowMax(0, 15)
									SAPaint_File(0, 0, 1, 1, res, "", 0.03, 0, 0, SA_ThemeWhite(), 1, 1, false)
									if SA_DivInfoGet("touchInside") > 0 && SA_DivInfoGet("touchEnd") > 0 {
										SA_DialogClose()
									}
									SA_DialogEnd()
								}
							}
							SA_DivEnd()
						}
					} else if IsFloat(col.Type) {

						switch col.Render {
						case "":
							if SA_Editbox(&values[x]).Show(x, 0, 1, rowSize).finished {
								writeCell = true
							}
						case "PERCENT":
							v, _ := strconv.ParseFloat(values[x], 64)
							value := strconv.FormatFloat(v*100, 'f', col.Prop_percent_floats, 64) + "%"
							if SA_Editbox(&value).ValueOrig(values[x]).Show(x, 0, 1, rowSize).finished {
								values[x] = value
								writeCell = true
							}
						}
					} else if IsInteger(col.Type) {

						switch col.Render {
						case "":
							if SA_Editbox(&values[x]).Show(x, 0, 1, rowSize).finished {
								writeCell = true
							}
						case "CHECK_BOX":
							bv := values[x] != "0"
							if SA_Checkbox(&bv, "").Show(x, 0, 1, rowSize) {
								if bv {
									values[x] = "1"
								} else {
									values[x] = "0"
								}
								writeCell = true
							}
						case "DATE":
							SA_DivStart(x, 0, 1, rowSize)
							SA_ColMax(0, 100)
							datt, _ := strconv.Atoi(values[x])
							date := int64(datt)
							if CalendarButton(fmt.Sprint("Calendar_%s_%s_%d_%d", table.Name, col.Name, st, x), &date, &store.calendar_page, true) {
								values[x] = strconv.Itoa(int(date))
								writeCell = true
							}
							SA_DivEnd()

						case "RATING":
							if SA_DivStart(x, 0, 1, rowSize) {
								act, _ := strconv.Atoi(values[x])
								act, writeCell = SA_Rating(act, col.Prop_rating_max_stars, SA_ThemeCd(), SA_ThemeGrey(0.8), "app:resources/star.png")
								if writeCell {
									values[x] = strconv.Itoa(act)
								}
							}
							SA_DivEnd()
						}

					} else if IsText(col.Type) {
						if SA_Editbox(&values[x]).Show(x, 0, 1, rowSize).finished {
							writeCell = true
						}
					} else {
						SA_TextError("Error: Unknown type").Show(x, 0, 1, rowSize)
					}

					if writeCell {
						v := values[x]
						if IsText(col.Type) {
							v = "'" + v + "'"
						}
						SA_SqlWrite("", fmt.Sprintf("UPDATE %s SET %s=%s WHERE rowid=%s;", table.Name, col.Name, v, values[0]))
					}
					x++
				}
			}
			SA_DivEnd()

			st++
		}
	}
	SA_DivEnd()
}
func TableStats(table *Table) {

	//columns sizes
	{
		x := 0
		for _, col := range table.Columns {
			if col.Show {
				SA_Col(x, col.Resize)
				x++
			}
		}
	}

	var stat *SA_Sql
	q, num_cols := GetQueryStats(table)
	values := make([]string, num_cols)
	if len(q) > 0 {
		stat = SA_SqlRead("", q)

		args := make([]interface{}, num_cols)
		for i := range values {
			args[i] = &values[i]
		}
		stat.Next(args...)
	}

	stat_i := 0
	x := 0
	for _, col := range table.Columns {
		if !col.Show {
			continue
		}

		if col.isRowId() {
			//add row
			if SA_ButtonLight("+").Tooltip(trns.ADD_ROW).Show(x, 0, 1, 1).click {
				SA_SqlWrite("", "INSERT INTO "+table.Name+" DEFAULT VALUES;")
				table.scrollDown = true
			}
		} else {
			//column stat
			text := ""
			if len(col.StatFunc) > 0 {
				text = col.StatFunc + ": " + values[stat_i]
				stat_i++
			}
			if SA_ButtonStyle(text, &g_ButtonStat).Show(x, 0, 1, 1).click { //show result
				SA_DialogOpen("Stat_"+strconv.Itoa(x), 1)
			}
			if SA_DialogStart("Stat_" + strconv.Itoa(x)) {

				SA_ColMax(0, 5)
				y := 0
				if IsInteger(col.Type) || IsFloat(col.Type) {

					if SA_ButtonMenu(trns.MIN).Show(0, y, 1, 1).click {
						col.StatFunc = "min"
						SA_DialogClose()
					}
					y++

					if SA_ButtonMenu(trns.MAX).Show(0, y, 1, 1).click {
						col.StatFunc = "max"
						SA_DialogClose()
					}
					y++

					if SA_ButtonMenu(trns.AVG).Show(0, y, 1, 1).click {
						col.StatFunc = "avg"
						SA_DialogClose()
					}
					y++

					if SA_ButtonMenu(trns.SUM).Show(0, y, 1, 1).click {
						col.StatFunc = "sum"
						SA_DialogClose()
					}
					y++

					if SA_ButtonMenu(trns.COUNT).Show(0, y, 1, 1).click {
						col.StatFunc = "count"
						SA_DialogClose()
					}
					y++

				}

				SA_DialogEnd()
			}
		}
		x++
	}

}

func GetQueryWHERE(table *Table) string {

	var query string

	if table.Filter.Enable {

		nfilters := 0
		for _, f := range table.Filter.Items {
			if f.Column == "" || len(f.GetOpString()) == 0 {
				continue
			}
			nfilters++
		}

		i := 0
		queryFilter := ""
		for _, f := range table.Filter.Items {
			col := table.FindColumn(f.Column)
			if col == nil {
				continue
			}

			op := f.GetOpString()
			val := f.Value

			//convert
			if IsText(col.Type) {
				val = "'" + val + "'" //add quotes
			}
			if IsInteger(col.Type) {
				v, _ := strconv.Atoi(val)
				val = strconv.Itoa(v)
			}
			if IsFloat(col.Type) {
				v, _ := strconv.ParseFloat(val, 64)
				val = strconv.FormatFloat(v, 'f', -1, 64)
			}
			if IsBlob(col.Type) {
				//...
			}

			queryFilter += f.Column + op + val
			if i+1 < nfilters {
				if table.Filter.Rel == 0 {
					queryFilter += " AND "
				} else {
					queryFilter += " OR "
				}
			}
			i++
		}
		if len(queryFilter) > 0 {
			query += " WHERE " + queryFilter
		}
	}

	if table.Sort.Enable {
		nsorts := 0
		for _, s := range table.Sort.Items {
			if s.Column == "" {
				continue
			}
			nsorts++
		}

		i := 0
		querySort := ""
		for _, s := range table.Sort.Items {
			if s.Column == "" {
				continue
			}

			querySort += s.Column
			if s.Az == 0 {
				querySort += " ASC"
			} else {
				querySort += " DESC"
			}
			if i+1 < nsorts {
				querySort += ","
			}

			i++
		}
		if len(querySort) > 0 {
			query += " ORDER BY " + querySort
		}
	}

	return query
}

func GetQueryCount(table *Table) string {
	query := "SELECT COUNT(*) AS COUNT FROM " + table.Name
	query += GetQueryWHERE(table)
	return query
}

func GetQueryBasic(table *Table) (string, int) {
	query := "SELECT "

	//columns
	ncols := 0
	for _, col := range table.Columns {
		if col.Show {
			ncols++
		}
	}

	if ncols == 0 {
		return "", 0
	}

	i := 0
	for _, col := range table.Columns {
		if !col.Show {
			continue
		}

		if IsBlob(col.Type) {
			query += "rowid AS " + col.Name
		} else if IsDate(col.Type) {
			query += "DATE(" + col.Name + ")"
		} else {
			query += col.Name
		}

		if i+1 < ncols {
			query += ","
		}
		i++
	}

	query += " FROM " + table.Name + ""
	query += GetQueryWHERE(table)
	return query, ncols
}

func GetQueryStats(table *Table) (string, int) {
	query := "SELECT "

	//columns
	ncols := 0
	for _, col := range table.Columns {
		if col.Show && len(col.StatFunc) > 0 {
			ncols++
		}
	}

	if ncols == 0 {
		return "", ncols
	}

	i := 0
	for _, col := range table.Columns {
		if !col.Show || len(col.StatFunc) == 0 {
			continue
		}

		query += col.StatFunc + "(" + col.Name + ")"
		if i+1 < ncols {
			query += ","
		}
		i++
	}

	query += " FROM " + table.Name + ""
	query += GetQueryWHERE(table)
	return query, ncols
}

func Render() {

	UpdateTables()
	SA_ColMax(0, 100)
	SA_RowMax(1, 100)

	SA_DivStart(0, 0, 1, 1)
	TopHeader()
	SA_DivEnd()

	var selectedTable *Table
	if len(store.Tables) > 0 {
		selectedTable = store.Tables[store.SelectedTable]
	}

	if selectedTable != nil {
		SA_DivStartName(0, 1, 1, 1, selectedTable.Name)
		{
			SA_ColMax(0, 100)
			SA_RowMax(0, 100)

			//table
			SA_DivStart(0, 0, 1, 1)
			TableView(selectedTable)
			SA_DivEnd()
		}
		SA_DivEnd()
	}
}

var styles SA_Styles
var g_ButtonColumnHeader _SA_Style
var g_ButtonStat _SA_Style
var g_ButtonLeft _SA_Style
var g_ButtonLightLeft _SA_Style

func Open() {
	//default
	json.Unmarshal(SA_File("storage_json"), &store)
	json.Unmarshal(SA_File("translations_json:app:resources/translations.json"), &trns)
	json.Unmarshal(SA_File("styles_json"), &styles)

	//styles
	g_ButtonColumnHeader = styles.ButtonLight
	g_ButtonColumnHeader.FontAlignH(0)
	g_ButtonColumnHeader.Id = 0

	g_ButtonStat = styles.ButtonLight
	g_ButtonStat.FontAlignH(0)
	g_ButtonStat.Id = 0

	g_ButtonLeft = styles.Button
	g_ButtonLeft.FontAlignH(0)

	g_ButtonLightLeft = styles.ButtonLight
	g_ButtonLightLeft.FontAlignH(0)

	//others
	InitCalendar()
}

func SetupDB() {
}

func Save() []byte {
	js, _ := json.MarshalIndent(&store, "", "")
	return js
}
