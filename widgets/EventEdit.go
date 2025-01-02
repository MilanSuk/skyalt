package main

import (
	"errors"
)

type EventEdit struct {
	ChangedButtonTitle string //"Create", "Edit"
	Event              *EventsEvent
	Date               int64
	changed            func()
	deleted            func()
}

func (layout *Layout) AddEventEdit(x, y, w, h int) *EventEdit {
	props := &EventEdit{}
	layout._createDiv(x, y, w, h, "EventEdit", props.Build, nil, nil)
	return props
}

func (st *EventEdit) Build(layout *Layout) {

	layout.SetColumn(0, 1, 100)

	//start date
	{
		BeginDiv := layout.AddLayout(0, 0, 1, 1)
		BeginDiv.SetColumn(0, 3, 3)
		BeginDiv.SetColumn(1, 1, 100)
		BeginDiv.AddText(0, 0, 1, 1, "Begin")
		BeginDiv.AddTimePickerButton(1, 0, 1, 1, &st.Event.Start, &st.Date)
	}

	//end date
	{
		FinishDiv := layout.AddLayout(0, 1, 1, 1)
		FinishDiv.SetColumn(0, 3, 3)
		FinishDiv.SetColumn(1, 1, 100)
		FinishDiv.AddText(0, 0, 1, 1, "Finish")
		end := st.Event.Start + st.Event.Duration
		pk := FinishDiv.AddTimePickerButton(1, 0, 1, 1, &end, &st.Date)
		pk.changed = func() {
			st.Event.Duration = end - st.Event.Start
		}
	}

	var errTitle error
	if len(st.Event.Title) == 0 {
		errTitle = errors.New("Empty")
	}

	{
		TitleDiv := layout.AddLayout(0, 2, 1, 1)
		TitleDiv.SetColumn(0, 3, 3)
		TitleDiv.SetColumn(1, 1, 100)
		TitleDiv.AddText(0, 0, 1, 1, "Title")
		ed := TitleDiv.AddEditbox(1, 0, 1, 1, &st.Event.Title)
		ed.RefreshDelaySec = 1
		//errTitle ...
	}

	{
		DescDiv := layout.AddLayout(0, 3, 1, 1)
		DescDiv.SetColumn(0, 3, 3)
		DescDiv.SetColumn(1, 1, 100)
		DescDiv.AddText(0, 0, 1, 1, "Description")
		ed := DescDiv.AddEditbox(1, 0, 1, 1, &st.Event.Description)
		ed.RefreshDelaySec = 1
	}

	{
		FileDiv := layout.AddLayout(0, 4, 1, 1)
		FileDiv.SetColumn(0, 3, 3)
		FileDiv.SetColumn(1, 1, 100)
		FileDiv.AddText(0, 0, 1, 1, "File")
		FileDiv.AddFilePickerButton(1, 0, 1, 1, &st.Event.File, true) //drag & drop? ...
	}

	{
		CdDiv := layout.AddLayout(0, 5, 1, 1)
		CdDiv.SetColumn(0, 3, 3)
		CdDiv.SetColumn(1, 1, 100)
		CdDiv.AddText(0, 0, 1, 1, "Color")
		CdDiv.AddColorPickerButton(1, 0, 1, 1, &st.Event.Color)
	}

	var errOrder error
	{
		if st.Event.Duration == 0 {
			errOrder = errors.New("finish" + " > " + "begin")
			layout.AddText(0, 6, 1, 1, errOrder.Error())
		}
	}

	Div := layout.AddLayout(0, 7, 1, 1)
	{
		Div.SetColumn(0, 1, 100)

		bt, btLay := Div.AddButton2(0, 0, 1, 1, NewButton(st.ChangedButtonTitle))
		btLay.Enable = (errTitle == nil && errOrder == nil)
		bt.clicked = func() {
			if st.changed != nil {
				st.changed()
			}
		}

		if st.deleted != nil {
			Div.SetColumn(1, 1, 100)
			Div.SetColumn(2, 1, 100)
			deleteBt := Div.AddButtonConfirm(2, 0, 1, 1, "Delete", "Are you sure?")
			deleteBt.confirmed = func() {
				if st.deleted != nil {
					st.deleted()
				}
			}
		}
	}
}
