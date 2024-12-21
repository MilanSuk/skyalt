package main

import (
	"time"
)

type Counter struct {
	Count int

	Items []string
}

func (layout *Layout) AddCounter(x, y, w, h int, props *Counter) *Counter {
	layout._createDiv(x, y, w, h, "Counter", props.Build, nil, nil)
	return props
}

var g_Counter *Counter

func NewFile_Counter() *Counter {
	if g_Counter == nil {
		g_Counter = &Counter{Count: 7}
		_read_file("Counter-Counter", g_Counter)
	}
	return g_Counter
}

func (st *Counter) Build(layout *Layout) {

	// Set grid
	layout.SetColumn(0, 1, 9)
	layout.SetRow(0, 1, 3)

	div := layout.AddLayout(0, 0, 1, 1)
	//div.SetColumn(0, 1, 100)
	div.SetColumnResizable(0, 1, 15, 2)
	div.SetColumn(1, 1, 100)
	div.SetColumn(2, 1, 100)
	div.SetRow(0, 1, 100)

	dia := div.AddDialog("test")
	dia.Layout.AddText(0, 0, 1, 1, "A")

	inc, incL := div.AddButton2(0, 0, 1, 1, NewButton("+"))
	inc.Tooltip = "Increment counter"
	incL.LLMTip = "Add 1 into '.Count'"
	inc.clicked = func() {
		dia.OpenRelative(incL)
		//dia.OpenDialogCentered()
		st.Count++
	}

	//val := div.AddText(1, 0, 1, 1, strconv.Itoa(st.Count))
	val, valL := div.AddEditboxInt2(1, 0, 1, 1, &st.Count)
	valL.LLMTip = "'.Count'"
	val.Align_h = 1 // center text

	dec, decL := div.AddButton2(2, 0, 1, 1, NewButton("-"))
	dec.Tooltip = "Decrement counter"
	decL.LLMTip = "Subtract 1 from '.Count'"
	dec.clicked = func() {
		st.Count--
	}

	sliVal := float64(st.Count)
	sli := layout.AddSlider(0, 1, 1, 1, &sliVal, -10, 10, 1)
	sli.changed = func() {
		st.Count = int(sliVal)
	}

	if len(st.Items) == 0 {
		st.Items = []string{"a", "b", "c", "d"}
	}
	for i, it := range st.Items {
		_, layBt := layout.AddButton2(0, 2+i, 1, 1, NewButtonMenu(it, "", 0))

		layBt.Drag_group = "item"
		layBt.Drop_group = "item"
		layBt.Drag_index = i
		layBt.Drop_v = true
		layBt.dropMove = func(src int, dst int) {
			Layout_MoveElement(&st.Items, &st.Items, src, dst)
		}
	}
}

func (st *Counter) Run(job *Job) {
	for job.IsRunning() {
		st.Count++
		time.Sleep(1 * time.Second)
	}
}
