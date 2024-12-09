package main

func (st *Text) Build() {
	st.lock.Lock()
	defer st.lock.Unlock()

	st.buildContextDialog()

	if !st.Multiline {
		st.layout.ScrollH.Narrow = true
	}
}

func (st *Text) Draw(rect Rect) {
	st.lock.Lock()
	defer st.lock.Unlock()

	layout := st.layout

	rectLabel := rect

	if st.Selection {
		layout.Paint_cursor("ibeam", rect)
		layout.Paint_tooltipEx(rectLabel, st.Tooltip, false)
	}

	//color
	cd := layout.GetPalette().OnB
	if st.Cd.A > 0 {
		cd = st.Cd
	}

	//draw icon
	if st.Icon != "" {
		rectIcon := rectLabel
		if st.Value != "" {
			rectIcon.W = 1
		}
		rectIcon = rectIcon.Cut(st.Icon_margin)

		rectLabel = rectLabel.CutLeft(1)

		layout.Paint_file(rectIcon, false, st.Icon, cd, cd, cd, 1, 1)
	}

	//draw text
	if st.Value != "" {
		layout.Paint_text(rectLabel, st.Value, "", cd, cd, cd, st.Selection, false, uint8(st.Align_h), uint8(st.Align_v), st.Formating, st.Multiline, st.Linewrapping, 0.06)
	}
}

func (st *Text) Input(in LayoutInput) {
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

func (st *Text) buildContextDialog() {

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
}
