package main

import (
	"fmt"
	"image/color"
	"strconv"
)

type ChartColumnValue struct {
	Value float64
	Label string
	Cd    color.RGBA
}

func (v *ChartColumnValue) Get(min, max float64) float64 {
	x := (v.Value - min) / (max - min)
	return x
}

type ChartColumn struct {
	Values []ChartColumnValue
}

func (col *ChartColumn) GetSum() float64 {
	sum := 0.0
	for _, pt := range col.Values {
		sum += pt.Value
	}
	return sum
}

func ChartColumns_getBound(columns []ChartColumn, bound_y0 bool) (float64, float64) {
	var min float64
	var max float64

	for i, col := range columns {
		if i == 0 && bound_y0 {
			min = columns[0].GetSum()
			max = min
		}
		if col.GetSum() < min {
			min = col.GetSum()
		}
		if col.GetSum() > max {
			max = col.GetSum()
		}
	}

	return min, max
}

type ChartColumns struct {
	Tooltip string

	X_unit, Y_unit string
	Bound_y0       bool
	Y_as_time      bool
	Columns        []ChartColumn
	X_Labels       []string
	ColumnMargin   float64
}

func (layout *Layout) AddChartColumns(x, y, w, h int, columns []ChartColumn, x_Labels []string) *ChartColumns {
	props := &ChartColumns{Columns: columns, X_Labels: x_Labels}
	lay := layout._createDiv(x, y, w, h, "ChartColumns", props.Build, nil, nil)
	lay.fnGetLLMTip = props.getLLMTip
	return props
}

func (st *ChartColumns) getLLMTip(layout *Layout) string {
	return Layout_buildLLMTip("ChartColumns", "", "", st.Tooltip)
}

func (st *ChartColumns) Build(layout *Layout) {
	axis_x := 1.0
	axis_y := 1.5
	unit_label_space := 0.5 //space around layout
	axis_space := 0.2       //space for Paint_circleRad()
	small_axis_line := 0.1  //small lines on axis

	layout.SetColumn(0, axis_y, axis_y)
	layout.SetColumn(1, 1, Layout_MAX_SIZE)
	layout.SetRow(0, 1, Layout_MAX_SIZE)
	layout.SetRow(1, axis_x, axis_x)

	DivY := layout.AddLayout(0, 0, 1, 2)
	DivX := layout.AddLayout(0, 1, 2, 1)
	DivG := layout.AddLayout(1, 0, 1, 1)

	defCd := layout.GetPalette().OnB

	if len(st.Columns) == 0 || len(st.X_Labels) == 0 {
		return
	}

	//bound
	min, max := ChartColumns_getBound(st.Columns, st.Bound_y0)

	//draw Graph
	DivG.fnDraw = func(rect Rect, layout *Layout) (paint LayoutPaint) {
		rect = rect.Cut(axis_space) //!
		rect = rect.CutTop(unit_label_space)
		rect = rect.CutRight(unit_label_space)

		for c, col := range st.Columns {

			w := rect.W / float64(len(st.Columns))
			x := float64(c) * w

			sum := 0.0
			for _, val := range col.Values {
				sum += val.Value

				sy := (sum - min) / (max - min)
				h := val.Value / (max - min)

				rc := rect
				rc.X += x
				rc.W = w
				rc.Y += rect.H * (1 - sy) //reverse
				rc.H = h * rect.H

				rc = rc.CutLeft(st.ColumnMargin).CutRight(st.ColumnMargin)

				paint.Rect(rc, val.Cd, val.Cd, val.Cd, 0)

				str := ChartColumn_Format(val.Value, st.Y_as_time)
				separ := ""
				if val.Label != "" {
					separ = ": "
				}
				paint.TooltipEx(rc, fmt.Sprintf("%s%s%s", val.Label, separ, str), false)
			}
		}
		return
	}

	DivX.fnDraw = func(rect Rect, layout *Layout) (paint LayoutPaint) {
		rect.X += axis_y
		rect.W -= axis_y
		rect = rect.CutLeft(axis_space).CutRight(axis_space)
		rect = rect.CutRight(unit_label_space)

		paint.Line(rect, 0, 0, 1, 0, defCd, 0.03) //axis

		if min == max {
			return
		}

		//x-values
		{
			for c := range st.Columns {
				if c >= len(st.X_Labels) {
					continue
				}

				w := rect.W / float64(len(st.Columns))
				x := float64(c) * w

				rc := rect
				rc.X += x
				rc.W = w
				rc = rc.CutLeft(-unit_label_space).CutRight(-unit_label_space)

				str := st.X_Labels[c]
				if c == len(st.Columns)-1 {
					str += st.X_unit
				}

				rc = rc.CutLeft(st.ColumnMargin).CutRight(st.ColumnMargin)
				paint.Text(rc, str, "", defCd, defCd, defCd, false, false, 1, 1)

				rc.H = small_axis_line
				paint.Line(rc, 0.5, 0, 0.5, 1, defCd, 0.03)
			}
		}
		return
	}

	DivY.fnDraw = func(rect Rect, layout *Layout) (paint LayoutPaint) {
		rect.H -= axis_x
		rect = rect.CutTop(axis_space).CutBottom(axis_space)
		rect = rect.CutTop(unit_label_space)
		rect = rect.CutRight(0.03) //paint space for axis line

		paint.Line(rect, 1, 0, 1, 1, defCd, 0.03) //axis

		if min == max {
			return
		}

		//y-values
		{
			num_values := rect.H
			jump := 0.125
			for (max-min)/jump > num_values {
				jump *= 2
			}

			for vy := min; vy <= max; vy += jump {
				y := (vy - min) / (max - min)
				y = 1 - y //reverse

				rc := rect
				rc.Y += rect.H*y - 0.5
				rc.H = 1

				str := ChartColumn_Format(vy, st.Y_as_time)
				paint.Text(rc, str, "", defCd, defCd, defCd, false, false, 1, 1)

				rc.X = rc.W - small_axis_line
				rc.W = small_axis_line
				rc.Y += 0.5
				paint.Line(rc, 0, 0, 1, 0, defCd, 0.03)
			}
		}
		return
	}
}

func ChartColumn_Format(value float64, as_time bool) string {
	var str string
	if as_time {
		seconds := int(value)
		days := seconds / (24 * 3600)
		seconds %= 24 * 3600
		hours := seconds / 3600
		seconds %= 3600
		minutes := seconds / 60
		seconds %= 60

		if days > 0 {
			str = fmt.Sprintf("%d:%02d:%02d:%02d", days, hours, minutes, seconds)
		} else if hours > 0 {
			str = fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
		} else {
			str = fmt.Sprintf("%02d:%02d", minutes, seconds)
		}
	} else {
		str = strconv.FormatFloat(value, 'f', -1, 64)
	}
	return str
}
