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

	Icon        string
	Icon_margin float64

	Selection    bool
	Formating    bool
	Multiline    bool
	Linewrapping bool

	ScrollToStart bool
	ScrollToEnd   bool
}

func (layout *Layout) AddText(x, y, w, h int, value string) *Text {
	props := &Text{Value: value, Align_v: 1, Selection: true, Formating: true}
	layout._createDiv(x, y, w, h, "Text", props.Build, props.Draw, props.Input)
	return props
}

func (layout *Layout) AddTextMultiline(x, y, w, h int, value string) *Text {
	props := &Text{Value: value, Selection: true, Formating: true, Multiline: true, Linewrapping: true}
	layout._createDiv(x, y, w, h, "Text", props.Build, props.Draw, props.Input)
	return props
}

func (st *Text) Build(layout *Layout) {
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
	rectLabel := rect

	if st.Selection {
		paint.Cursor("ibeam", rect)
		paint.TooltipEx(rectLabel, st.Tooltip, false)
	}

	//color
	cd := Paint_GetPalette().OnB
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

		paint.File(rectIcon, false, st.Icon, cd, cd, cd, 1, 1)
	}

	//draw text
	if st.Value != "" {
		paint.Text(rectLabel, st.Value, "", cd, cd, cd, st.Selection, false, uint8(st.Align_h), uint8(st.Align_v), st.Formating, st.Multiline, st.Linewrapping, 0.06)
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
