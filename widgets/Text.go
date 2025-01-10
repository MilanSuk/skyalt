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

	ScrollToStart bool
	ScrollToEnd   bool
}

func (layout *Layout) AddText(x, y, w, h int, value string) *Text {
	props := &Text{Value: value, Align_v: 1, Selection: true, Formating: true, Cd: Paint_GetPalette().OnB}
	layout._createDiv(x, y, w, h, "Text", props.Build, props.Draw, props.Input)
	return props
}

func (layout *Layout) AddTextMultiline(x, y, w, h int, value string) *Text {
	props := &Text{Value: value, Selection: true, Formating: true, Multiline: true, Linewrapping: true, Cd: Paint_GetPalette().OnB}
	layout._createDiv(x, y, w, h, "Text", props.Build, props.Draw, props.Input)
	return props
}

func (st *Text) Build(layout *Layout) {

	{
		var paint LayoutPaint
		layout.UserCRFromText = st.addPaintText(Rect{}, &paint)
	}

	st.buildContextDialog(layout)
	if !st.Multiline {
		layout.ScrollH.Narrow = true
	}

	if st.ScrollToStart {
		st.ScrollToStart = false
		layout.HScrollToTheLeft()
		layout.VScrollToTheTop()
	}
	if st.ScrollToEnd {
		st.ScrollToEnd = false
		layout.HScrollToTheRight()
		layout.VScrollToTheBottom()
	}
}

func (st *Text) Draw(rect Rect, layout *Layout) (paint LayoutPaint) {

	if st.Selection {
		paint.Cursor("ibeam", rect)
		paint.TooltipEx(rect, st.Tooltip, false)
	}

	//draw text
	if st.Value != "" {
		st.addPaintText(rect, &paint)
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

func (st *Text) addPaintText(rect Rect, paint *LayoutPaint) *LayoutDrawText {
	tx := paint.Text(rect, st.Value, "", st.Cd, st.Cd, st.Cd, st.Selection, false, uint8(st.Align_h), uint8(st.Align_v))
	tx.Formating = st.Formating
	tx.Multiline = st.Multiline
	tx.Linewrapping = st.Linewrapping
	tx.Margin = 0.06
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
