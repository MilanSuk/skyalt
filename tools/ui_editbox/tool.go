package main

// Show a editbox(on/off) on screen.
type ui_editbox struct {
	X int //X position
	Y int //Y position
	W int //Width
	H int //Height

	Description string //Name of value and where it's come from

	Value string //Current value
}

func (st *ui_editbox) run() string {
	return "success"
}
