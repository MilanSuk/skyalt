package main

// Set layout's grid row dimension.
type ui_set_row struct {
	Index int     //Row's index. Starts with 0.
	Min   float64 //Minimum size of row.
	Max   float64 //Maximum size of row.
}

func (st *ui_set_row) run() string {
	return "success"
}
