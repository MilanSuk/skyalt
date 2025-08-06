package main

import (
	"fmt"
	"image/color"
	"strconv"
)

type ChartPoint struct {
	X  float64
	Y  float64
	Cd color.RGBA
}

func (v *ChartPoint) Get(min, max ChartPoint) (float64, float64) {
	x := (v.X - min.X) / (max.X - min.X)
	y := (v.Y - min.Y) / (max.Y - min.Y)
	y = 1 - y //reverse
	return x, y
}

type ChartLine struct {
	Points []ChartPoint
	Label  string
	Cd     color.RGBA
}

type ChartLines struct {
	Tooltip string
	Lines   []ChartLine

	X_unit, Y_unit        string
	Bound_x0, Bound_y0    bool
	Point_rad, Line_thick float64
	Draw_XHelpLines       bool
	Draw_YHelpLines       bool
}

func (layout *Layout) AddChartLines(x, y, w, h int, lines []ChartLine) *ChartLines {
	props := &ChartLines{Lines: lines}
	lay := layout._createDiv(x, y, w, h, "ChartLines", props.Build, nil, nil)
	lay.fnGetLLMTip = props.getLLMTip
	return props
}

func (st *ChartLines) getLLMTip(layout *Layout) string {
	return Layout_buildLLMTip("ChartLines", "", "", st.Tooltip)
}

func (st *ChartLines) Build(layout *Layout) {
	axis_x := 1.0
	axis_y := 1.0
	unit_label_space := 0.5 //space around layout
	axis_space := 0.2       //space for Paint_circleRad()
	small_axis_line := 0.1  //small lines on axis

	layout.SetColumn(0, axis_y, axis_y)
	layout.SetColumn(1, 1, 100)
	layout.SetRow(0, 1, 100)
	layout.SetRow(1, axis_x, axis_x)

	DivY := layout.AddLayout(0, 0, 1, 2)
	DivX := layout.AddLayout(0, 1, 2, 1)
	DivG := layout.AddLayout(1, 0, 1, 1)

	if len(st.Lines) == 0 {
		return
	}

	//bound
	min, max := ChartLines_getBound(st.Lines, st.Bound_x0, st.Bound_y0)

	defCd := layout.GetPalette().OnB

	//draw Graph
	DivG.fnDraw = func(rect Rect, layout *Layout) (paint LayoutPaint) {
		rect = rect.Cut(axis_space) //!
		rect = rect.CutTop(unit_label_space)
		rect = rect.CutRight(unit_label_space)

		for ln := range st.Lines {

			lineCd := st.Lines[ln].Cd
			if lineCd.A == 0 {
				lineCd = defCd
			}

			label := st.Lines[ln].Label

			var last_x, last_y float64
			for pt := range st.Lines[ln].Points {

				x, y := st.Lines[ln].Points[pt].Get(min, max)

				cd := st.Lines[ln].Points[pt].Cd
				if cd.A == 0 {
					cd = lineCd
				}

				paint.CircleRad(rect, x, y, st.Point_rad, cd, cd, cd, 0)
				if pt > 0 {
					paint.Line(rect, last_x, last_y, x, y, cd, st.Line_thick)
				}

				cdLine := lineCd
				cdLine.A = 20

				if st.Draw_XHelpLines {
					rcl := rect
					rcl.X += rect.W * x

					paint.Line(rcl, 0, 0, 0, 1, cdLine, 0.03)
				}

				if st.Draw_YHelpLines {
					rcl := rect
					rcl.Y += rect.H * y
					paint.Line(rcl, 0, 0, 1, 0, cdLine, 0.03)
				}

				{
					separ := ""
					if label != "" {
						separ = ": "
					}

					rad := st.Point_rad
					if rad < 0.25 {
						rad = 0.25
					}
					rc := rect
					rc.X += (rect.W * x) - rad
					rc.W = 2 * rad

					strX := strconv.FormatFloat(st.Lines[ln].Points[pt].X, 'f', -1, 64)
					strY := strconv.FormatFloat(st.Lines[ln].Points[pt].Y, 'f', -1, 64)
					paint.TooltipEx(rc, fmt.Sprintf("%s%s%s, %s", label, separ, strX, strY), false)
				}

				last_x = x
				last_y = y
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

		if min.X == max.X {
			return
		}

		//x-values
		{
			num_values := rect.W
			diff := (max.X - min.X)
			jump := 0.125
			for diff/jump > num_values {
				jump *= 2
			}

			for vx := min.X; vx <= max.X; vx += jump {
				x := (vx - min.X) / (max.X - min.X)

				rc := rect
				rc.X += rect.W*x - 0.5
				rc.W = 1
				rc = rc.CutLeft(-unit_label_space).CutRight(-unit_label_space)

				str := strconv.FormatFloat(vx, 'f', -1, 64)
				if vx+jump > max.X {
					str += st.X_unit
				}

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

		if min.X == max.X {
			return
		}

		//y-values
		{
			num_values := rect.H
			diff := (max.Y - min.Y)
			jump := 0.125
			for diff/jump > num_values {
				jump *= 2
			}

			if st.Y_unit != "" {
				rc := rect
				rc.Y -= unit_label_space
				rc.H = 0.5
				paint.Text(rc, st.Y_unit, "", defCd, defCd, defCd, false, false, 1, 0)
			}

			for vy := min.Y; vy <= max.Y; vy += jump {
				y := (vy - min.Y) / (max.Y - min.Y)
				y = 1 - y //reverse

				rc := rect
				rc.Y += rect.H*y - 0.5
				rc.H = 1

				str := strconv.FormatFloat(vy, 'f', -1, 64)

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

func ChartLines_getBound(lines []ChartLine, bound_x0, bound_y0 bool) (ChartPoint, ChartPoint) {
	var min, max ChartPoint

	if len(lines) > 0 && len(lines[0].Points) > 0 {
		//bound==false => can be in range 5-10(no 0)
		if bound_x0 {
			min.X = lines[0].Points[0].X
		}
		if bound_y0 {
			min.Y = lines[0].Points[0].Y
		}
	}
	max = min

	for _, line := range lines {
		for _, pt := range line.Points {
			//min
			if pt.X < min.X {
				min.X = pt.X
			}
			if pt.Y < min.Y {
				min.Y = pt.Y
			}
			//max
			if pt.X > max.X {
				max.X = pt.X
			}
			if pt.Y > max.Y {
				max.Y = pt.Y
			}
		}
	}
	return min, max
}
