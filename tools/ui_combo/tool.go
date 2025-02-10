package main

// Show a combo(selector from options) on screen.
type ui_combo struct {
	X int //X position
	Y int //Y position
	W int //Width
	H int //Height

	Description string //Name of value and where it's come from

	Value string //Current value

	Values []string //Options to pick value from
	Labels []string //Labels for Values
}

func (st *ui_combo) run() string {
	return "success"
}
