package main

// Show a switch(on/off) on screen.
type ui_switch struct {
	X int //X position
	Y int //Y position
	W int //Width
	H int //Height

	Description string //Name of value and where it's come from

	Label string //Description
	Value bool   //Current value
}

func (st *ui_switch) run() string {
	return "success"
}
