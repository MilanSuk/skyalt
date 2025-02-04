package main

import "strconv"

// Show a slider on screen. It returns new Value(changed by user).
type ui_slider struct {
	Min   float64 //Minimum range
	Max   float64 //Maximum range
	Step  float64 //Step size
	Value float64 //Current value
}

func (st *ui_slider) run() string {
	return strconv.FormatFloat(st.Value, 'f', -1, 64)
}
