package main

type Test struct {
	Count int
}

func (layout *Layout) AddTest(x, y, w, h int, props *Test) *Test {
	layout._createDiv(x, y, w, h, "Test", props.Build, nil, nil)
	return props
}

var g_Test *Test

func OpenFile_Test() *Test {
	if g_Test == nil {
		g_Test = &Test{}
		_read_file("Test-Test", g_Test)
	}
	return g_Test
}

func (st *Test) Build(layout *Layout) {
	layout.SetColumn(0, 0, 100)
	layout.SetColumn(1, 0, 5)
	layout.SetColumn(2, 0, 5)
	layout.SetColumn(3, 0, 5)
	layout.SetColumn(4, 0, 100)

	layout.SetRow(0, 0, 5)

	incrementButton := layout.AddButton(1, 0, 1, 1, "+")
	incrementButton.clicked = func() {
		st.Count++
	}

	editbox := layout.AddEditbox(2, 0, 1, 1, &st.Count)
	editbox.Align_h = 1 // Center the text horizontally

	decrementButton := layout.AddButton(3, 0, 1, 1, "-")
	decrementButton.clicked = func() {
		st.Count--
	}

	layout.AddSlider(1, 1, 3, 1, &st.Count, -10, 10, 1)
}
