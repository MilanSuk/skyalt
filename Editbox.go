package main

import "strconv"

func (st *Editbox) Build() {
	st.lock.Lock()
	defer st.lock.Unlock()

	st.buildContextDialog()

	if !st.Multiline {
		st.layout.ScrollH.Narrow = true
	}

	st.layout.fnSetEditbox = func(value string) {
		if st.Value != nil {
			*st.Value = value
		}
		if st.ValueFloat != nil {
			*st.ValueFloat, _ = strconv.ParseFloat(value, 64)
		}
		if st.ValueInt != nil {
			*st.ValueInt, _ = strconv.Atoi(value)
		}
	}
}

func (st *Editbox) Draw(rect Rect) {
	st.lock.Lock()
	defer st.lock.Unlock()

	layout := st.layout

	rectLabel := rect

	layout.Paint_cursor("ibeam", rect)
	layout.Paint_tooltipEx(rectLabel, st.Tooltip, false)

	//color
	cd := layout.GetPalette().OnB
	if st.Cd.A > 0 {
		cd = st.Cd
	}

	var value string
	if st.Value != nil {
		value = *st.Value
	}
	if st.ValueFloat != nil {
		value = strconv.FormatFloat(*st.ValueFloat, 'f', st.FloatPrec, 64) //update string from float. Map use it for lon/lat/zoom editboxes
	}
	if st.ValueInt != nil {
		value = strconv.Itoa(*st.ValueInt)
	}

	//draw icon
	if st.Icon != "" {
		rectIcon := rectLabel
		if value != "" {
			rectIcon.W = 1
		}
		rectIcon = rectIcon.Cut(st.Icon_margin)

		rectLabel = rectLabel.CutLeft(1)

		layout.Paint_file(rectIcon, false, st.Icon, cd, cd, cd, 1, 1)
	}

	//draw text
	layout.Paint_text(rectLabel, value, st.Ghost, cd, cd, cd, true, true, uint8(st.Align_h), uint8(st.Align_v), st.Formating, st.Multiline, st.Linewrapping, 0.06)
}

func (st *Editbox) Input(in LayoutInput) {
	st.lock.Lock()
	defer st.lock.Unlock()

	//open context menu
	active := in.IsActive
	inside := in.IsInside && (active || !in.IsUse)
	if in.IsUp && active && inside && in.AltClick {
		dia := st.layout.FindDialog("context")
		if dia != nil {
			dia.OpenDialogOnTouch()
		}
	}
}

func (st *Editbox) buildContextDialog() {

	dia := st.layout.AddDialog("context")
	dia.SetColumn(0, 1, 5)

	SelectAll := dia.AddButton(0, 0, 1, 1, NewButton("Select All"))
	SelectAll.Align = 0
	SelectAll.Background = 0.25
	SelectAll.clicked = func() {
		st.layout.SelectAllText()
		dia.CloseDialog()
	}

	Copy := dia.AddButton(0, 1, 1, 1, NewButton("Copy"))
	Copy.Align = 0
	Copy.Background = 0.25
	Copy.clicked = func() {
		st.layout.CopyText()
		dia.CloseDialog()
	}

	Cut := dia.AddButton(0, 2, 1, 1, NewButton("Cut"))
	Cut.Align = 0
	Cut.Background = 0.25
	Cut.clicked = func() {
		st.layout.CutText()
		dia.CloseDialog()
	}

	Paste := dia.AddButton(0, 3, 1, 1, NewButton("Paste"))
	Paste.Align = 0
	Paste.Background = 0.25
	Paste.clicked = func() {
		st.layout.PasteText()
		dia.CloseDialog()
	}
}
