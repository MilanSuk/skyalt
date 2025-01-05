package main

import "strconv"

// Test struct holds the Count attribute.
type Test struct {
	Count int
}

// AddTest adds a Test widget to the layout.
func (layout *Layout) AddTest(x, y, w, h int, props *Test) *Test {
	layout._createDiv(x, y, w, h, "Test", props.Build, nil, nil)
	return props
}

var g_Test *Test

// OpenFile_Test opens the Test configuration file.
func OpenFile_Test() *Test {
	if g_Test == nil {
		g_Test = &Test{}
		_read_file("Test-Test", g_Test)
	}
	return g_Test
}

// Build constructs the Test layout with buttons and an editbox for Count.
func (st *Test) Build(layout *Layout) {
	// Set column max to 5 for these three columns
	layout.SetColumn(0, 0, 5)
	layout.SetColumn(1, 0, 5)
	layout.SetColumn(2, 0, 5)

	// Set column max to 5 for this row
	layout.SetRow(0, 0, 5)

	// Add button to increment the count
	incrementButton := layout.AddButton(0, 0, 1, 1, "+")
	incrementButton.clicked = func() {
		st.Count++
		layout.Redraw()
	}

	// Add button to decrement the count
	decrementButton := layout.AddButton(2, 0, 1, 1, "-")
	decrementButton.clicked = func() {
		st.Count--
		layout.Redraw()
	}

	// Add an editbox for the Count attribute at the specified position.
	countStr := strconv.Itoa(st.Count)
	editbox := layout.AddEditbox(1, 0, 1, 1, &countStr)
	editbox.Align_h = 1 // Center the text horizontally

	// Add a slider for the Count attribute at the specified position.
	layout.AddSliderInt(0, 1, 3, 2, &st.Count, -10, 10, 1)
}
