package main

import (
	"fmt"
	"image/color"
)

type Text struct {
	Tooltip string
	Cd      color.RGBA

	Value string

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

	lay.fnAutoResize = props.autoResize
	lay.fnGetAutoResizeMargin = props.getAutoResizeMargin
	lay.fnGetLLMTip = props.getLLMTip

	return props, lay
}

func (layout *Layout) AddText(x, y, w, h int, value string) *Text {
	props, _ := layout.AddText2(x, y, w, h, value)
	return props
}

func (st *Text) getLLMTip(layout *Layout) string {
	return fmt.Sprintf("Type: Text. Value: %s. Tooltip: %s", st.Value, st.Tooltip)
}

func (st *Text) Build(layout *Layout) {
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
		tx := paint.Text(Rect{}, st.Value, "", st.Cd, st.Cd, st.Cd, st.Selection, false, uint8(st.Align_h), uint8(st.Align_v))
		tx.Formating = st.Formating
		tx.Multiline = st.Multiline
		tx.Linewrapping = st.Linewrapping
		tx.Margin = st.getAutoResizeMargin()
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

func (st *Text) autoResize(layout *Layout) bool {
	return layout.resizeFromPaintText(st.Value, st.Multiline, st.Linewrapping, st.getAutoResizeMargin())
}
func (st *Text) getAutoResizeMargin() [4]float64 {
	m := (1 - WinFontProps_GetDefaultLineH()) / 2
	return [4]float64{m, m, m, m}
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
