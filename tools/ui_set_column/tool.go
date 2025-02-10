package main

// Set layout's grid column dimension.
type ui_set_column struct {
	Index int     //Column's index. Starts with 0.
	Min   float64 //Minimum size of column.
	Max   float64 //Maximum size of column.
}

func (st *ui_set_column) run() string {
	return "success"
}
