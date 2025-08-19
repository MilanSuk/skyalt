package main

import (
	"image/color"
	"strconv"
)

type Editbox struct {
	Tooltip string

	Cd color.RGBA

	ValuePointer   interface{} //*string, *int, *float64
	ValueFloatPrec int

	Ghost string

	Align_h int //0=left, 1=center, 2=right
	Align_v int //0=top, 1=center, 2=bottom

	Formating    bool
	Multiline    bool
	Linewrapping bool

	Password bool

	changed func()
	enter   func()

	fnRefreshValueFromTemp func()
}

func (layout *Layout) AddEditbox(x, y, w, h int, valuePointer interface{}) *Editbox {
	props := &Editbox{ValuePointer: valuePointer, Align_v: 1, Formating: true, ValueFloatPrec: 2, Cd: layout.GetPalette().OnB}
	lay := layout._createDiv(x, y, w, h, "Editbox", props.Build, props.Draw, props.Input)
	lay.fnAutoResize = props.autoResize
	lay.fnGetAutoResizeMargin = props.getAutoResizeMargin

	props.fnRefreshValueFromTemp = func() {
		edit := layout.ui.edit
		if edit.Is(layout) {
			edit.Set(layout.UID, true, edit.orig_value, props.getValue(), false, false, true, true, layout.ui)
		}
	}

	return props
}

func (st *Editbox) getLLMTip(layout *Layout) string {
	return Layout_buildLLMTip("Editbox", "saying", st.getValue(), st.Tooltip)
}

func (st *Editbox) Build(layout *Layout) {
	layout.Back_rounding = true //for disable

	st.buildContextDialog(layout)

	if !st.Multiline {
		layout.scrollH.Narrow = true
	}

	layout.dropFile = func(path string) {
		if st.setValueAdd(path) {
			if st.changed != nil {
				st.changed()
			}
		}
	}

	layout.fnSetEditbox = func(value string, enter bool) {
		diff := st.setValue(value)

		//note: never call both(avoid 2x tool's change() async)
		if enter && st.enter != nil {
			st.enter()
		} else if diff && st.changed != nil {
			st.changed()
		}

	}

	layout.fnGetEditbox = func() string {
		return st.getValue()
	}

}

func (st *Editbox) Draw(rect Rect, layout *Layout) (paint LayoutPaint) {
	paint.CursorEx(rect, "ibeam")
	paint.TooltipEx(rect, st.Tooltip, false)

	//background
	cd := layout.GetPalette().B
	paint.RectRad(rect, cd, cd, cd, 0, layout.getRounding())

	//text
	tx := paint.Text(Rect{}, st.getValue(), st.Ghost, st.Cd, st.Cd, st.Cd, true, true, uint8(st.Align_h), uint8(st.Align_v))
	tx.Formating = st.Formating
	tx.Multiline = st.Multiline
	tx.Linewrapping = st.Linewrapping
	tx.Password = st.Password
	tx.Margin = st.getAutoResizeMargin()

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

func (st *Editbox) autoResize(layout *Layout) {
	value := st.getValue()
	if layout.ui.edit.Is(layout) {
		value = layout.ui.edit.temp
	}
	layout.resizeFromPaintText(value, st.Multiline, st.Linewrapping, st.getAutoResizeMargin())
}
func (st *Editbox) getAutoResizeMargin() [4]float64 {
	m := (1 - WinFontProps_GetDefaultLineH()) / 2
	return [4]float64{m, m, m, m}
}

func (st *Editbox) setValueAdd(value string) bool {
	switch v := st.ValuePointer.(type) {
	case *string:
		*v += value
		return true
	}
	return false
}

func (st *Editbox) setValue(value string) bool {
	diff := false
	switch v := st.ValuePointer.(type) {
	case *string:
		diff = (*v != value)
		*v = value

	case *int:
		val, _ := strconv.Atoi(value)
		diff = (*v != val)
		*v = val
	case *int64:
		val, _ := strconv.Atoi(value)
		diff = (*v != int64(val))
		*v = int64(val)
	case *int32: //also rune
		val, _ := strconv.Atoi(value)
		diff = (*v != int32(val))
		*v = int32(val)
	case *int16:
		val, _ := strconv.Atoi(value)
		diff = (*v != int16(val))
		*v = int16(val)
	case *int8:
		val, _ := strconv.Atoi(value)
		diff = (*v != int8(val))
		*v = int8(val)

	case *uint:
		val, _ := strconv.Atoi(value)
		diff = (*v != uint(val))
		*v = uint(val)
	case *uint64:
		val, _ := strconv.Atoi(value)
		diff = (*v != uint64(val))
		*v = uint64(val)
	case *uint32:
		val, _ := strconv.Atoi(value)
		diff = (*v != uint32(val))
		*v = uint32(val)
	case *uint16:
		val, _ := strconv.Atoi(value)
		diff = (*v != uint16(val))
		*v = uint16(val)
	case *uint8: //also byte
		val, _ := strconv.Atoi(value)
		diff = (*v != uint8(val))
		*v = uint8(val)

	case *float64:
		val, _ := strconv.ParseFloat(value, 64)
		diff = (*v != val)
		*v = val
	case *float32:
		val, _ := strconv.ParseFloat(value, 32)
		diff = (*v != float32(val))
		*v = float32(val)

	case *bool:
		val, _ := strconv.ParseBool(value)
		diff = (*v != val)
		*v = val
	}

	return diff
}

func (st *Editbox) getValue() string {
	switch v := st.ValuePointer.(type) {
	case *string:
		return *v
	case *int:
		return strconv.Itoa(*v)
	case *int64:
		return strconv.Itoa(int(*v))
	case *int32: //also rune
		return strconv.Itoa(int(*v))
	case *int16:
		return strconv.Itoa(int(*v))
	case *int8:
		return strconv.Itoa(int(*v))

	case *float64:
		return strconv.FormatFloat(*v, 'f', st.ValueFloatPrec, 64)
	case *float32:
		return strconv.FormatFloat(float64(*v), 'f', st.ValueFloatPrec, 64)

	case *bool:
		return strconv.FormatBool(*v)
	}

	return ""
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

	STT := dia.Layout.AddButton(0, 4, 1, 1, OsTrnString(layout.ui.router.services.mic.Find(layout.UID) == nil, "Start", "Stop")+" recording speech")
	STT.Tooltip = "Record microphone and convert speech-to-text"
	STT.Align = 0
	STT.Background = 0.25
	STT.clicked = func() {
		layout.RecordText()
		dia.Close()
	}
}
