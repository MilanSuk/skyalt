package main

import (
	"fmt"
	"math"
	"time"
)

// Shows elevation chart from the activity file.
type ShowActivityPacesChart struct {
	ActivityID string //ID of the activity
}

func (st *ShowActivityPacesChart) run(caller *ToolCaller, ui *UI) error {
	source_activities, err := NewActivities("")
	if err != nil {
		return err
	}

	activity, found := source_activities.Activities[st.ActivityID]
	if !found {
		return fmt.Errorf("activity '%s' not found", st.ActivityID)
	}

	// load
	gpx, err := source_activities.GetGpx(st.ActivityID)
	if err != nil {
		return err
	}
	segments := gpx.getGPXSegments()

	ui.SetColumn(0, 1, 100)
	ui.SetRow(1, 10, 10)

	ui.AddTextLabel(0, 0, 1, 1, fmt.Sprintf("Paces - %s", caller.ConvertTextDate(int64(activity.Date))))

	var columns []UIChartColumn
	var x_labels []string

	var last_time time.Time
	var last_dist float64
	totalDistance := 0.0
	for _, seg := range segments {
		for i := range seg.Trkpts {

			if i == 0 {
				last_time, _ = time.Parse(time.RFC3339, seg.Trkpts[i].Time)
				continue
			}

			prev := seg.Trkpts[i-1]
			curr := seg.Trkpts[i]

			distance := gpx.haversine(prev.Lat, prev.Lon, curr.Lat, curr.Lon)

			currTime, err := time.Parse(time.RFC3339, curr.Time)
			if err != nil {
				fmt.Println("Error parsing time:", err)
				continue
			}

			if totalDistance < math.Round(totalDistance) && totalDistance+distance >= math.Round(totalDistance) {
				pace := currTime.Sub(last_time).Seconds() / (totalDistance - last_dist) //diff_time / diff_distance = time_in_seconds

				//Label: fmt.Sprintf("%d:%d", int(pace)/60, int(pace)%60)
				columns = append(columns, UIChartColumn{Values: []UIChartColumnValue{{Value: pace, Cd: UI_GetPalette().P}}})
				x_labels = append(x_labels, fmt.Sprintf("%d", int(totalDistance)+1))

				last_time = currTime
				last_dist = totalDistance
			}

			totalDistance += distance
		}
	}

	if len(columns) > 0 {
		min := columns[0].Values[0].Value
		max := min
		for i := range columns {
			v := columns[i].Values[0].Value
			if v < min {
				min = v
			}
			if v > max {
				max = v
			}
		}
		for i := range columns {
			t := (columns[i].Values[0].Value - min) / (max - min)
			columns[i].Values[0].Cd = gpx.getSpeedColor(1 - t)
		}
	}

	cc := ui.AddChartColumns(0, 1, 1, 1, columns, x_labels)
	cc.ColumnMargin = 0.1
	cc.X_unit = "km"
	cc.Y_as_time = true
	cc.Bound_y0 = true

	return nil
}
