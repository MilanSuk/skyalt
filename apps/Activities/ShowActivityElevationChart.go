package main

import (
	"fmt"
	"image/color"
	"math"
)

type ShowActivityElevationChart struct {
	ActivityID string // ID of the activity
}

func (st *ShowActivityElevationChart) run(caller *ToolCaller, ui *UI) error {
	activities, err := LoadActivities()
	if err != nil {
		return err
	}

	activity, found := activities.Activities[st.ActivityID]
	if !found {
		return fmt.Errorf("activity ID %s not found", st.ActivityID)
	}

	dateStr := SdkGetDate(activity.StartDate)

	ui.addTextH1("Elevation for " + dateStr)

	gpx, err := LoadGpx(st.ActivityID)
	if err != nil {
		return err
	}

	var allPoints []Point
	for _, track := range gpx.Tracks {
		for _, seg := range track.Segments {
			allPoints = append(allPoints, seg.Points...)
		}
	}

	var chartPoints []UIChartPoint
	cumDist := 0.0
	if len(allPoints) > 0 {
		chartPoints = append(chartPoints, UIChartPoint{X: cumDist, Y: allPoints[0].Elevation})
		for i := 1; i < len(allPoints); i++ {
			d := haversine(allPoints[i-1].Latitude, allPoints[i-1].Longitude, allPoints[i].Latitude, allPoints[i].Longitude)
			cumDist += d
			chartPoints = append(chartPoints, UIChartPoint{X: cumDist, Y: allPoints[i].Elevation})
		}
	}

	line := UIChartLine{
		Points: chartPoints,
		Label:  "Elevation",
		Cd:     color.RGBA{0, 0, 0, 255},
	}
	lines := []UIChartLine{line}

	ui.setRowHeight(10, 20)
	chart := ui.addChartLines(lines, "")
	chart.X_unit = "km"
	chart.Y_unit = "m"
	chart.Bound_y0 = true
	chart.Draw_YHelpLines = true
	chart.Draw_XHelpLines = true
	chart.Line_thick = 2

	return nil
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * (math.Pi / 180)
	dLon := (lon2 - lon1) * (math.Pi / 180)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1*(math.Pi/180))*math.Cos(lat2*(math.Pi/180))*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
