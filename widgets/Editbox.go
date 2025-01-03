package main

import (
	"image/color"
	"strconv"
)

type Editbox struct {
	Cd color.RGBA

	Value      *string
	ValueFloat *float64
	ValueInt   *int
	FloatPrec  int //for 'ValueFloat'

	Ghost   string
	Tooltip string

	Align_h int
	Align_v int

	Icon        string
	Icon_margin float64

	Formating    bool
	Multiline    bool
	Linewrapping bool

	DrawBackWhenNonEmpty bool
	ResetButton          bool

	RefreshDelaySec float64 //? ...

	changed func()
	enter   func()
}

func (layout *Layout) AddEditbox(x, y, w, h int, value *string) *Editbox {
	props := &Editbox{Value: value, Align_v: 1, Formating: true}
	layout._createDiv(x, y, w, h, "Editbox", props.Build, props.Draw, props.Input)
	return props
}

func (layout *Layout) AddEditboxInt2(x, y, w, h int, value *int) (*Editbox, *Layout) {
	props := &Editbox{ValueInt: value, Align_v: 1, Formating: true}
	return props, layout._createDiv(x, y, w, h, "Editbox", props.Build, props.Draw, props.Input)
}
func (layout *Layout) AddEditboxInt(x, y, w, h int, value *int) *Editbox {
	props, _ := layout.AddEditboxInt2(x, y, w, h, value)
	return props
}

func (layout *Layout) AddEditboxFloat(x, y, w, h int, value *float64, prec int) *Editbox {
	props := &Editbox{ValueFloat: value, FloatPrec: prec, Align_v: 1, Formating: true}
	layout._createDiv(x, y, w, h, "Editbox", props.Build, props.Draw, props.Input)
	return props
}

func (layout *Layout) AddEditboxMultiline(x, y, w, h int, value *string) (*Editbox, *Layout) {
	props := &Editbox{Value: value, Align_v: 1, Formating: true, Multiline: true, Linewrapping: true}
	return props, layout._createDiv(x, y, w, h, "Editbox", props.Build, props.Draw, props.Input)
}

func (layout *Layout) AddEditboxSearch(x, y, w, h int, value *string, refreshDelaySec float64) *Editbox {
	if refreshDelaySec < 0 {
		refreshDelaySec = 0.5
	}
	props := &Editbox{Value: value, RefreshDelaySec: refreshDelaySec,
		Ghost: "Search", Icon: "resources/search.png", Icon_margin: 0.2,
		DrawBackWhenNonEmpty: true, ResetButton: true}

	layout._createDiv(x, y, w, h, "Editbox", props.Build, props.Draw, props.Input)
	return props
}

func (st *Editbox) Build(layout *Layout) {

	st.buildContextDialog(layout)

	if !st.Multiline {
		layout.ScrollH.Narrow = true
	}

	layout.dropFile = func(path string) {
		*st.Value += path
	}

	layout.fnSetEditbox = func(value string, enter bool) {
		diff := false
		if st.Value != nil {
			v := value
			diff = (*st.Value != v)
			*st.Value = v
		}
		if st.ValueFloat != nil {
			v, _ := strconv.ParseFloat(value, 64)
			diff = (*st.ValueFloat != v)
			*st.ValueFloat = v
		}
		if st.ValueInt != nil {
			v, _ := strconv.Atoi(value)
			diff = (*st.ValueInt != v)
			*st.ValueInt = v
		}

		if diff && st.changed != nil {
			st.changed()
		}
		if enter && st.enter != nil {
			st.enter()
		}
	}
}

func (st *Editbox) Draw(rect Rect, layout *Layout) (paint LayoutPaint) {
	rectOrig := rect
	rectOrig = rectOrig.Cut(0.03)

	rectLabel := rect

	//color
	cd := Paint_GetPalette().OnB
	if st.Cd.A > 0 {
		cd = st.Cd
	}

	//back
	if st.DrawBackWhenNonEmpty && *st.Value != "" {
		backCd := Color_Aprox(Paint_GetPalette().P, Paint_GetPalette().B, 0.5)
		paint.Rect(rectOrig, backCd, backCd, backCd, 0)
	}

	//icon
	if st.Icon != "" {
		rectIcon := rectLabel
		rectIcon.W = 1
		rectIcon = rectIcon.Cut(st.Icon_margin)

		rectLabel = rectLabel.CutLeft(1)

		paint.File(rectIcon, false, st.Icon, cd, cd, cd, 1, 1)
	}

	//reset
	if st.ResetButton {
		rectReset := rectLabel
		rectReset = rectReset.CutRight(1)
		rectReset.X += rectReset.W
		rectReset.W = 1

		paint.Text(rectReset, "⌫", "", cd, cd, cd, false, false, 1, 1, false, false, false, 0)

		paint.CursorEx(rectReset, "hand")
		paint.TooltipEx(rectReset, "Clear", false)

		rectLabel = rectLabel.CutRight(1)
	}

	paint.CursorEx(rectLabel, "ibeam")
	paint.TooltipEx(rectLabel, st.Tooltip, false)

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

		paint.File(rectIcon, false, st.Icon, cd, cd, cd, 1, 1)
	}

	//draw text
	paint.Text(rectLabel, value, st.Ghost, cd, cd, cd, true, true, uint8(st.Align_h), uint8(st.Align_v), st.Formating, st.Multiline, st.Linewrapping, 0.06)
	return
}

func (st *Editbox) Input(in LayoutInput, layout *Layout) {

	//open context menu
	active := in.IsActive
	inside := in.IsInside && (active || !in.IsUse)
	if in.IsUp && active && inside && in.AltClick {
		dia := layout.FindDialog("context")
		if dia != nil {
			dia.OpenOnTouch()
		}
	}
}

func (st *Editbox) buildContextDialog(layout *Layout) {

	dia := layout.AddDialog("context")
	dia.Layout.SetColumn(0, 1, 5)

	SelectAll := dia.Layout.AddButton(0, 0, 1, 1, "Select All")
	SelectAll.Align = 0
	SelectAll.Background = 0.25
	SelectAll.clicked = func() {
		layout.SelectAllText()
		dia.Close()
	}

	Copy := dia.Layout.AddButton(0, 1, 1, 1, "Copy")
	Copy.Align = 0
	Copy.Background = 0.25
	Copy.clicked = func() {
		layout.CopyText()
		dia.Close()
	}

	Cut := dia.Layout.AddButton(0, 2, 1, 1, "Cut")
	Cut.Align = 0
	Cut.Background = 0.25
	Cut.clicked = func() {
		layout.CutText()
		dia.Close()
	}

	Paste := dia.Layout.AddButton(0, 3, 1, 1, "Paste")
	Paste.Align = 0
	Paste.Background = 0.25
	Paste.clicked = func() {
		layout.PasteText()
		dia.Close()
	}
}
