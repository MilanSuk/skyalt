package main

import (
	"image/color"
)

type Text struct {
	Cd color.RGBA

	Value   string
	Tooltip string

	Align_h int
	Align_v int

	Selection    bool
	Formating    bool
	Multiline    bool
	Linewrapping bool
}

func (layout *Layout) AddText2(x, y, w, h int, value string) (*Text, *Layout) {
	props := &Text{Value: value, Align_v: 1, Selection: true, Formating: true, Cd: layout.GetPalette().OnB}
	lay := layout._createDiv(x, y, w, h, "Text", props.Build, props.Draw, props.Input)
	return props, lay
}

func (layout *Layout) AddText(x, y, w, h int, value string) *Text {
	props, _ := layout.AddText2(x, y, w, h, value)
	return props
}

func (layout *Layout) AddTextMultiline(x, y, w, h int, value string) *Text {
	props := &Text{Value: value, Selection: true, Formating: true, Multiline: true, Linewrapping: true, Cd: layout.GetPalette().OnB}
	layout._createDiv(x, y, w, h, "Text", props.Build, props.Draw, props.Input)
	return props
}

func (st *Text) Build(layout *Layout) {

	layout.UserCRFromText = st.addPaintText(&LayoutPaint{})

	st.buildContextDialog(layout)
	if !st.Multiline {
		layout.scrollH.Narrow = true
	}
}

func (st *Text) Draw(rect Rect, layout *Layout) (paint LayoutPaint) {

	if st.Selection {
		paint.Cursor("ibeam", rect)
		paint.TooltipEx(rect, st.Tooltip, false)
	}

	//draw text
	if st.Value != "" {
		st.addPaintText(&paint)
	}

	return
}

func (st *Text) Input(in LayoutInput, layout *Layout) {
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

func (st *Text) addPaintText(paint *LayoutPaint) *LayoutDrawText {
	tx := paint.Text(Rect{}, st.Value, "", st.Cd, st.Cd, st.Cd, st.Selection, false, uint8(st.Align_h), uint8(st.Align_v))
	tx.Formating = st.Formating
	tx.Multiline = st.Multiline
	tx.Linewrapping = st.Linewrapping

	m := (1 - WinFontProps_GetDefaultLineH()) / 2
	tx.Margin[0] = m
	tx.Margin[1] = m
	tx.Margin[2] = m
	tx.Margin[3] = m
	return tx
}

func (st *Text) buildContextDialog(layout *Layout) {
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
}
