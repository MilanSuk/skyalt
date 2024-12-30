package main

import "fmt"

type Counter struct {
	Count      int
	NumOfLines int

	MapCam OsmMapCam

	Items []string
}

func (layout *Layout) AddCounter(x, y, w, h int, props *Counter) *Counter {
	layout._createDiv(x, y, w, h, "Counter", props.Build, nil, nil)
	return props
}

var g_Counter *Counter

func NewFile_Counter() *Counter {
	if g_Counter == nil {
		g_Counter = &Counter{Count: 7, NumOfLines: 3}
		_read_file("Counter-Counter", g_Counter)
	}
	return g_Counter
}

func (st *Counter) Build(layout *Layout) {
	layout.SetColumn(0, 1, 9)
	layout.SetColumn(1, 1, 5)   // New empty column
	layout.SetColumn(2, 5, 5)   // Original column now at index 2
	layout.SetColumn(3, 1, 100) // New column for the map component
	layout.SetRow(0, 1, 7)      // Set row max to 7

	div := layout.AddLayout(0, 0, 1, 1)
	div.SetColumnResizable(0, 1, 15, 2)
	div.SetColumn(1, 1, 100)
	div.SetColumn(2, 1, 100)
	div.SetColumn(3, 1, 100) // New column for the text component
	div.SetRow(0, 1, 100)

	inc, incL := div.AddButton2(0, 0, 1, 1, NewButton("+"))
	inc.Tooltip = "Increment counter"
	incL.LLMTip = "Add 1 into '.Count'"
	inc.clicked = func() {
		st.Count++
	}

	val, valL := div.AddEditboxInt2(1, 0, 1, 1, &st.Count)
	valL.LLMTip = "'.Count'"
	val.Align_h = 1

	dec, decL := div.AddButton2(2, 0, 1, 1, NewButton("-"))
	dec.Tooltip = "Decrement counter"
	decL.LLMTip = "Subtract 1 from '.Count'"
	dec.clicked = func() {
		st.Count--
	}

	// Use value from the editbox for the text component
	textComp := div.AddText(3, 0, 1, 1, fmt.Sprintf("%d", st.Count))
	textComp.Tooltip = "This is a new text component"

	sliVal := float64(st.Count)
	sli := layout.AddSlider(0, 1, 1, 1, &sliVal, -10, 10, 1)
	sli.changed = func() {
		st.Count = int(sliVal)
	}

	// Add slider at the specified position
	sliderVal := float64(st.NumOfLines)
	slider := layout.AddSlider(0, 4, 1, 1, &sliderVal, 1, 10, 1)
	slider.changed = func() {
		st.NumOfLines = int(sliderVal)
	}

	// Add text to show NumOfLines
	layout.AddText(0, 5, 1, 1, fmt.Sprintf("Number of Lines: %d", st.NumOfLines))

	// Add new row below
	layout.SetRow(2, 1, 1) // Changed from 3 to 1

	// Add a new button in the new row
	newBtn := layout.AddButton(0, 6, 1, 1, NewButton("New Button"))
	newBtn.Tooltip = "This is a new button added in the new row"
	newBtn.clicked = func() {
		st.NumOfLines++
	}

	// Add map here
	layout.AddOsmMap(1, 0, 3, 1, &st.MapCam)

	// Add new button here
	layout.SetRow(1, 1, 8) // Adjust row to accommodate new button
	decNumOfLines := layout.AddButton(0, 7, 1, 1, NewButton("-"))
	decNumOfLines.Tooltip = "Decrement Number of Lines"
	decNumOfLines.clicked = func() {
		if st.NumOfLines > 0 {
			st.NumOfLines--
		}
	}
}
