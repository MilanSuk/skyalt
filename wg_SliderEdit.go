package main

type SliderEdit struct {
	Description string
	Tooltip     string

	ValuePointer   interface{}
	Min, Max, Step float64
	Legend         bool

	ValuePointerPrec int

	Description_width, Slider_width, Edit_width float64

	DrawSteps bool

	changed func()
}

func (layout *Layout) AddSliderEdit(x, y, w, h int, valuePointer interface{}, min, max, step float64) *SliderEdit {
	props := &SliderEdit{Description_width: 100, Slider_width: 100, Edit_width: 100, ValuePointer: valuePointer, ValuePointerPrec: -1, Min: min, Max: max, Step: step}
	layout._createDiv(x, y, w, h, "SliderEdit", props.Build, nil, nil)
	return props
}
func (layout *Layout) AddSliderEditInt(x, y, w, h int, value *int, min, max, step float64) *SliderEdit {
	v := float64(*value)
	props := layout.AddSliderEdit(x, y, w, h, &v, min, max, step)
	props.changed = func() {
		*value = int(v)
	}
	return props
}

func (st *SliderEdit) Build(layout *Layout) {
	layout.SetColumn(0, st.Description_width, st.Description_width)
	layout.SetColumn(1, 1, st.Slider_width)
	layout.SetColumn(2, 1, st.Edit_width)

	tx := layout.AddText(0, 0, 1, 1, st.Description) //description
	tx.Tooltip = st.Tooltip

	sli := layout.AddSlider(1, 0, 1, 1, st.ValuePointer, st.Min, st.Max, st.Step)
	sli.ShowDrawSteps = st.DrawSteps
	sli.ShowLegend = st.Legend
	sli.changed = func() {
		if st.changed != nil {
			st.changed()
		}
	}

	ed := layout.AddEditbox(2, 0, 1, 1, st.ValuePointer)
	ed.ValueFloatPrec = st.ValuePointerPrec
	ed.changed = sli.changed
}
