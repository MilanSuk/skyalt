package main

// Show a slider on screen.
type ui_slider struct {
	X int //X position
	Y int //Y position
	W int //Width
	H int //Height

	Description string //Name of value and where it's come from

	Min   float64 //Minimum range
	Max   float64 //Maximum range
	Step  float64 //Step size
	Value float64 //Current value
}

func (st *ui_slider) run() string {
	return "success"
}
