package main

func (st *Counter) Build() {
	st.lock.Lock()
	defer st.lock.Unlock()

	// Set grid
	st.layout.SetColumn(0, 1, 9)
	st.layout.SetRow(0, 1, 3)

	div := st.layout.AddLayout(0, 0, 1, 1)
	//div.SetColumn(0, 1, 100)
	div.SetColumnResizable(0, 1, 3, 2)
	div.SetColumn(1, 1, 100)
	div.SetColumn(2, 1, 100)
	div.SetRow(0, 1, 100)

	dia := div.AddDialog("test")
	dia.AddText(0, 0, 1, 1, "A")

	inc := div.AddButton(0, 0, 1, 1, NewButton("+"))
	inc.Tooltip = "Increment counter"
	inc.layout.LLMTip = "Add 1 into '.Count'"
	inc.clicked = func() {
		dia.OpenDialogRelative(inc.layout)
		//dia.OpenDialogCentered()
		st.Count++
	}

	//val := div.AddText(1, 0, 1, 1, strconv.Itoa(st.Count))
	val := div.AddEditboxInt(1, 0, 1, 1, &st.Count)
	val.layout.LLMTip = "'.Count'"
	val.Align_h = 1 // center text

	dec := div.AddButton(2, 0, 1, 1, NewButton("-"))
	dec.Tooltip = "Decrement counter"
	dec.layout.LLMTip = "Subtract 1 from '.Count'"
	dec.clicked = func() {
		st.Count--
	}

	sliVal := float64(st.Count)
	sli := st.layout.AddSlider(0, 1, 1, 1, &sliVal, -10, 10, 1)
	sli.changed = func() {
		st.Count = int(sliVal)
	}
}
