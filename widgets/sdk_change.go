package main

// Set step size to 2 for this slider{MARK 28: x:0,y:1,w:3,h:1}.

// Assuming Layout and related methods are defined elsewhere in the codebase.

// UpdateTestSlider updates the slider step size for the Test widget.
func UpdateTestSlider() {
	test := OpenFile_Test()
	if test != nil {
		layout := &Layout{} // Assuming Layout is initialized appropriately
		test.Build(layout)

		// Modify the slider step size
		layout.AddSliderInt(0, 1, 3, 2, &test.Count, -10, 10, 2) // Updated step size to 2
	}
}
