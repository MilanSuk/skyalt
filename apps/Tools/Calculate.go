package main

import (
	"github.com/mnogu/go-calculator"
)

// Calculator which accepts expression(string) and return result as float-point number.
type Calculate struct {
	Expression string //Expression to calculate. For example (2*3)+5.
	Out_result float64
}

func (st *Calculate) run(caller *ToolCaller, ui *UI) error {

	var err error
	st.Out_result, err = calculator.Calculate(st.Expression)

	return err
}
