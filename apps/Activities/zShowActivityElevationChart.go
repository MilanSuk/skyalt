package main

import (
	"fmt"
	"image/color"
	"math"
)

// Shows elevation chart from the activity file.
type ShowActivityElevationChart struct {
	ActivityID string //ID of the activity
}

func (st *ShowActivityElevationChart) run(caller *ToolCaller, ui *UI) error {
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

	ui.AddTextLabel(0, 0, 1, 1, fmt.Sprintf("Elevation - %s", ConvertTextDate(int64(activity.Date))))

	line := UIChartLine{Label: "Elevation", Cd: color.RGBA{0, 0, 0, 255}}

	last_addDistance := 0.0
	totalDistance := 0.0
	for _, seg := range segments {
		for i := range seg.Trkpts {

			if last_addDistance+0.05 < totalDistance || i == 0 || i == len(seg.Trkpts)-1 {
				line.Points = append(line.Points, UIChartPoint{
					X:  totalDistance,
					Y:  math.Round(seg.Trkpts[i].Ele),
					Cd: seg.Trkpts[i].Cd,
				})
				last_addDistance = totalDistance
			}

			if i > 0 {
				totalDistance += gpx.haversine(seg.Trkpts[i-1].Lat, seg.Trkpts[i-1].Lon, seg.Trkpts[i].Lat, seg.Trkpts[i].Lon)
			}
		}
	}

	ch := ui.AddChartLines(0, 1, 1, 1, []UIChartLine{line})
	ch.Point_rad = 0
	ch.Line_thick = 0.06
	ch.X_unit = "km"
	ch.Y_unit = "m"
	ch.Bound_y0 = true

	return nil
}
