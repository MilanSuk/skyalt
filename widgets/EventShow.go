package main

type EventShow struct {
	Event   *EventsEvent
	changed func()
	deleted func()
}

func (layout *Layout) AddEventShow(x, y, w, h int) *EventShow {
	props := &EventShow{}
	layout._createDiv(x, y, w, h, "EventShow", props.Build, nil, nil)
	return props
}

func (st *EventShow) Build(layout *Layout) {

	layout.SetColumn(0, 1, 100)

	//Start date
	{
		BeginDiv := layout.AddLayout(0, 0, 1, 1)
		BeginDiv.SetColumn(0, 3, 3)
		BeginDiv.SetColumn(1, 1, 100)
		BeginDiv.AddText(0, 0, 1, 1, "Begin")
		BeginDiv.AddText(1, 0, 1, 1, Layout_ConvertTextDateTime(st.Event.Start))
	}

	//End date
	{
		FinishDiv := layout.AddLayout(0, 1, 1, 1)
		FinishDiv.SetColumn(0, 3, 3)
		FinishDiv.SetColumn(1, 1, 100)
		FinishDiv.AddText(0, 0, 1, 1, "Finish")
		FinishDiv.AddText(1, 0, 1, 1, Layout_ConvertTextDateTime(st.Event.Start+st.Event.Duration))
	}

	//Title
	{
		FinishDiv := layout.AddLayout(0, 2, 1, 1)
		FinishDiv.SetColumn(0, 3, 3)
		FinishDiv.SetColumn(1, 1, 100)
		FinishDiv.AddText(0, 0, 1, 1, "Title")
		FinishDiv.AddText(1, 0, 1, 1, st.Event.Title)
	}

	//Description
	{
		FinishDiv := layout.AddLayout(0, 3, 1, 1)
		FinishDiv.SetColumn(0, 3, 3)
		FinishDiv.SetColumn(1, 1, 100)
		FinishDiv.AddText(0, 0, 1, 1, "Description")
		FinishDiv.AddText(1, 0, 1, 1, st.Event.Description)
	}

	//File
	{
		FileDiv := layout.AddLayout(0, 4, 1, 1)
		FileDiv.SetColumn(0, 3, 3)
		FileDiv.SetColumn(1, 1, 100)
		FileDiv.AddText(0, 0, 1, 1, "File")
		FileDiv.AddFilePickerButton(1, 0, 1, 1, &st.Event.File, true)
	}

	//Color
	{
		FinishDiv := layout.AddLayout(0, 5, 1, 1)
		FinishDiv.SetColumn(0, 3, 3)
		FinishDiv.SetColumn(1, 1, 100)
		FinishDiv.AddText(0, 0, 1, 1, "Color")
		FinishDiv.AddColorPickerButton(1, 0, 1, 1, &st.Event.Color)
	}

	Div := layout.AddLayout(0, 7, 1, 1)
	{
		Div.SetColumn(0, 1, 100)
		Div.SetColumn(1, 1, 100)

		EditDia, EditLay := layout.AddDialogBorder("edit_event", "Event", 14)
		EditLay.SetColumn(0, 1, 100)
		EditLay.SetRowFromSub(0, 1, 100)
		ed := EditLay.AddEventEdit(0, 0, 1, 1)
		ed.Event = st.Event
		ed.ChangedButtonTitle = "Edit"
		ed.changed = func() {
			if st.changed != nil {
				st.changed()
			}
			EditDia.Close()
		}
		ed.deleted = func() {
			if st.deleted != nil {
				st.deleted()
			}
			EditDia.Close()
		}

		editBt := Div.AddButton(0, 0, 1, 1, "Edit")
		editBt.clicked = func() {
			EditDia.OpenCentered()
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
