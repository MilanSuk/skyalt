package main

import (
	"encoding/xml"
	"fmt"
	"image/color"
	"io"
	"math"
	"os"
	"time"
)

type ActivitiesActivity struct {
	FilePath string

	Cam OsmMapCam //map view position

	Title       string
	Description string

	Activity string //"walking", "running", "swimming"

	Date           int64 //unix time
	Total_distance float64
	Total_time_sec float64
	Calories       float64 //burn calories
}

type Activities struct {
	Activities        []ActivitiesActivity
	Selected_activity int
}

func (layout *Layout) AddActivities(x, y, w, h int, props *Activities) *Activities {
	layout._createDiv(x, y, w, h, "Activities", props.Build, nil, nil)
	return props
}

var g_Activities *Activities

func OpenFile_Activities() *Activities {
	if g_Activities == nil {
		g_Activities = &Activities{}
		_read_file("Activities-Activities", g_Activities)
	}
	return g_Activities
}

func (actvs *Activities) Build(layout *Layout) {

	actvs.checkActivities()

	layout.SetColumnResizable(0, 3, 10, 5)
	layout.SetColumn(1, 1, 100)
	layout.SetColumnResizable(2, 4, 100, 7)
	layout.SetRow(0, 1, 100)

	var sel *ActivitiesActivity
	if actvs.Selected_activity >= 0 && actvs.Selected_activity < len(actvs.Activities) {
		sel = &actvs.Activities[actvs.Selected_activity]
	}

	ListDiv := layout.AddLayout(0, 0, 1, 1)
	{
		ListDiv.SetColumn(0, 1, 100)
		ListDiv.SetRow(0, 1, 100)

		ItemsDiv := ListDiv.AddLayout(0, 0, 1, 1)
		actvs.buildListOfFiles(ItemsDiv)

		AddBt := ListDiv.AddButton(0, 1, 1, 1, "+")
		AddBt.Background = 0.5
		AddBt.clicked = func() {
			//maybe dialog? ...
			actvs.Activities = append(actvs.Activities, ActivitiesActivity{}) //...
			actvs.Selected_activity = len(actvs.Activities) - 1
		}

		//UserInfo
		{
			UserDia, UserLay := ListDiv.AddDialogBorder("user_info", "User Information", 7)
			UserLay.SetColumn(0, 1, 100)
			UserLay.SetRowFromSub(0)
			UserLay.AddUserInfo(0, 0, 1, 1, OpenFile_UserInfo())

			UserBt := ListDiv.AddButton(0, 2, 1, 1, "User info")
			UserBt.Background = 0.5
			UserBt.clicked = func() {
				UserDia.OpenCentered()
			}
		}
	}

	//activity
	var segments []OsmMapSegment
	if sel != nil {

		//map
		var err error
		segments, err = Activities_loadGPXFile(sel.FilePath)
		if err == nil {

			Div := layout.AddLayout(1, 0, 1, 1)
			Div.SetColumn(0, 1, 100)
			Div.SetRow(0, 2.5, 2.5)
			Div.SetRow(1, 1, 100)

			computeSpeedColors(segments)

			MapDiv := Div.AddLayout(0, 1, 1, 1)
			MapDiv.SetColumn(0, 1, 100)
			MapDiv.SetRow(0, 1, 100)
			mp := MapDiv.AddOsmMap(0, 0, 1, 1, &sel.Cam)
			if len(segments) > 0 {
				mp.AddRoute(OsmMapRoute{Segments: segments})
			}

			InfoDiv := Div.AddLayout(0, 0, 1, 1)
			InfoDiv.SetColumn(0, 1, 100)
			InfoDiv.SetColumn(1, 1, 100)
			InfoDiv.SetColumn(2, 1, 100)
			InfoDiv.SetColumn(3, 1, 100)
			InfoDiv.SetRow(0, 1, 100)
			InfoDiv.SetRow(1, 1, 100)
			totalTime, totalDistance := Activities_computeTimeAndDistance(segments)
			cals := Activities_computeCaloriesBurned(totalTime, totalDistance, OpenFile_UserInfo())

			tx := InfoDiv.AddText(0, 0, 1, 1, "<h1>"+fmt.Sprintf("%.2f", totalDistance))
			tx.Align_h = 1
			tx.Align_v = 2
			tx = InfoDiv.AddText(0, 1, 1, 1, "Distance(km)")
			tx.Align_h = 1
			tx.Align_v = 0

			tx = InfoDiv.AddText(1, 0, 1, 1, "<h1>"+fmt.Sprintf("%d:%02d:%02d", int(totalTime.Hours()), int(totalTime.Minutes())%60, int(totalTime.Seconds())%60))
			tx.Align_h = 1
			tx.Align_v = 2
			tx = InfoDiv.AddText(1, 1, 1, 1, "Time")
			tx.Align_h = 1
			tx.Align_v = 0

			avgTime := time.Duration(float64(totalTime) / totalDistance)
			tx = InfoDiv.AddText(2, 0, 1, 1, "<h1>"+fmt.Sprintf("%02d:%02d", int(avgTime.Minutes())%60, int(avgTime.Seconds())%60))
			tx.Align_h = 1
			tx.Align_v = 2
			tx = InfoDiv.AddText(2, 1, 1, 1, "Avg. Pace(km)")
			tx.Align_h = 1
			tx.Align_v = 0

			tx = InfoDiv.AddText(3, 0, 1, 1, fmt.Sprintf("<h1>%.0f", cals))
			tx.Align_h = 1
			tx.Align_v = 2
			tx = InfoDiv.AddText(3, 1, 1, 1, "Calories(kcal)")
			tx.Align_h = 1
			tx.Align_v = 0
		}
	}

	ChartDiv := layout.AddLayout(2, 0, 1, 1)
	{
		ChartDiv.SetColumn(0, 1, 100)

		ChartDiv.SetRow(1, 1, 100)
		ChartDiv.SetRow(2, 0.1, 0.1)
		ChartDiv.SetRow(4, 1, 100)

		tx := ChartDiv.AddText(0, 0, 1, 1, "<b>Elevation")
		tx.Align_h = 1
		{
			line := ChartLine{Label: "Elevation", Cd: color.RGBA{0, 0, 0, 255}}

			last_addDistance := 0.0
			totalDistance := 0.0
			for _, seg := range segments {
				for i := range seg.Trkpts {

					if last_addDistance+0.05 < totalDistance || i == 0 || i == len(seg.Trkpts)-1 {
						line.Points = append(line.Points, ChartPoint{
							X:  totalDistance,
							Y:  math.Round(seg.Trkpts[i].Ele),
							Cd: seg.Trkpts[i].Cd,
						})
						last_addDistance = totalDistance
					}

					if i > 0 {
						totalDistance += Activities_haversine(seg.Trkpts[i-1].Lat, seg.Trkpts[i-1].Lon, seg.Trkpts[i].Lat, seg.Trkpts[i].Lon)
					}
				}
			}

			ch := ChartDiv.AddChartLines(0, 1, 1, 1, []ChartLine{line})
			ch.Point_rad = 0
			ch.Line_thick = 0.06
			ch.X_unit = "km"
			ch.Y_unit = "m"
			ch.Bound_y0 = true
		}

		ChartDiv.AddDivider(0, 2, 1, 1, true)

		tx = ChartDiv.AddText(0, 3, 1, 1, "<b>Paces")
		tx.Align_h = 1
		{
			var columns []ChartColumn
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

					distance := Activities_haversine(prev.Lat, prev.Lon, curr.Lat, curr.Lon)

					currTime, err := time.Parse(time.RFC3339, curr.Time)
					if err != nil {
						fmt.Println("Error parsing time:", err)
						continue
					}

					if totalDistance < math.Round(totalDistance) && totalDistance+distance >= math.Round(totalDistance) {
						pace := currTime.Sub(last_time).Seconds() / (totalDistance - last_dist) //diff_time / diff_distance = time_in_seconds

						//Label: fmt.Sprintf("%d:%d", int(pace)/60, int(pace)%60)
						columns = append(columns, ChartColumn{Values: []ChartColumnValue{{Value: pace, Cd: Paint_GetPalette().P}}})
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
					columns[i].Values[0].Cd = getSpeedColor(1 - t)
				}
			}

			cc := ChartDiv.AddChartColumns(0, 4, 1, 1, columns, x_labels)
			cc.ColumnMargin = 0.1
			cc.X_unit = "km"
			cc.Y_as_time = true
			cc.Bound_y0 = true
		}
	}
}

func (actvs *Activities) checkActivities() {
	for i := range actvs.Activities {
		if actvs.Activities[i].Title == "" {
			actvs.Activities[i].Title = "no name"
		}
	}
}

func (actvs *Activities) buildListOfFiles(layout *Layout) {
	layout.SetColumn(0, 1, 100)

	for i := range actvs.Activities {
		ItemDiv := layout.AddLayout(0, i, 1, 1)
		ItemDiv.SetColumn(0, 1, 100)
		ItemDiv.Drag_group = "chat"
		ItemDiv.Drop_group = "chat"
		ItemDiv.Drag_index = i
		ItemDiv.Drop_v = true
		ItemDiv.dropMove = func(src int, dst int) {
			Layout_MoveElement(&actvs.Activities, &actvs.Activities, src, dst)
			if actvs.Selected_activity == src {
				actvs.Selected_activity = dst
			}
		}

		bt, btLay := ItemDiv.AddButtonMenu2(0, 0, 1, 1, actvs.Activities[i].Title, "", 0)
		bt.Tooltip = actvs.Activities[i].Description
		if i == actvs.Selected_activity {
			bt.Background = 1
		}
		bt.clicked = func() {
			actvs.Selected_activity = i
		}

		ctx := ItemDiv.AddButtonIcon(1, 0, 1, 1, "resources/settings.png", 0.2, "")
		ctx.Background = 0.1
		ctxDia := layout.AddDialog(fmt.Sprintf("activity_%d", i))
		{
			ctxDia.Layout.SetColumn(0, 1, 3)
			ctxDia.Layout.SetColumn(1, 1, 10)
			y := 0
			ctxDia.Layout.AddText(0, y, 1, 1, "Title")
			ctxDia.Layout.AddEditbox(1, y, 1, 1, &actvs.Activities[i].Title)
			y++

			ctxDia.Layout.AddText(0, y, 1, 1, "Description")
			ctxDia.Layout.AddEditbox(1, y, 1, 1, &actvs.Activities[i].Description)
			y++

			ctxDia.Layout.AddText(0, y, 1, 1, "File")
			ctxDia.Layout.AddFilePickerButton(1, y, 1, 1, &actvs.Activities[i].FilePath, true)
			y++

			//type: walk, run, swimming, cycling
			//...
			//+update calories ...

			y++ //space

			Delete := ctxDia.Layout.AddButtonConfirm(0, y, 2, 1, "Delete", "Delete '"+actvs.Activities[i].Title+"'?")
			Delete.confirmed = func() {
				actvs.Activities = append(actvs.Activities[:i], actvs.Activities[i+1:]...) //remove
				ctxDia.Close()
			}
			y++
		}
		ctx.clicked = func() {
			ctxDia.OpenRelative(btLay)
		}
	}
}

type Segment struct {
	Points []Point `xml:"trkpt"`
}
type Point struct {
	Lat  float64 `xml:"lat,attr"`
	Lon  float64 `xml:"lon,attr"`
	Ele  float64 `xml:"ele"`
	Time string  `xml:"time"`
}
type Track struct {
	Name    string    `xml:"name"`
	Segment []Segment `xml:"trkseg"`
}
type GPX struct {
	XMLName xml.Name `xml:"gpx"`
	Trk     []Track  `xml:"trk"`
}

func Activities_loadGPXFile(filename string) ([]OsmMapSegment, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var gpx GPX
	err = xml.Unmarshal(data, &gpx)
	if err != nil {
		return nil, err
	}

	var segments []OsmMapSegment
	for _, track := range gpx.Trk {
		for _, seg := range track.Segment {
			mapSeg := OsmMapSegment{
				Label:  track.Name,
				Trkpts: make([]OsmMapSegmentTrk, len(seg.Points)),
			}
			for i, point := range seg.Points {
				mapSeg.Trkpts[i] = OsmMapSegmentTrk{
					Lon:  point.Lon,
					Lat:  point.Lat,
					Ele:  point.Ele,
					Time: point.Time,
				}
			}

			segments = append(segments, mapSeg)
		}
	}

	return segments, nil
}

func Activities_computeTimeAndDistance(segments []OsmMapSegment) (totalTime time.Duration, totalDistance float64) {
	for _, segment := range segments {
		for i := 1; i < len(segment.Trkpts); i++ {
			prevPoint := segment.Trkpts[i-1]
			currentPoint := segment.Trkpts[i]

			// Compute time difference
			prevTime, _ := time.Parse(time.RFC3339, prevPoint.Time)
			currentTime, _ := time.Parse(time.RFC3339, currentPoint.Time)
			totalTime += currentTime.Sub(prevTime)

			// Compute distance using Haversine formula
			distance := Activities_haversine(prevPoint.Lat, prevPoint.Lon, currentPoint.Lat, currentPoint.Lon)
			totalDistance += distance
		}
	}
	return totalTime, totalDistance
}

func Activities_haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371 // km

	dLat := Activities_toRadians(lat2 - lat1)
	dLon := Activities_toRadians(lon2 - lon1)
	lat1 = Activities_toRadians(lat1)
	lat2 = Activities_toRadians(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1)*math.Cos(lat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}

func Activities_toRadians(deg float64) float64 {
	return deg * (math.Pi / 180)
}

func Activities_computeCaloriesBurned(totalTime time.Duration, totalDistance float64, user *UserInfo) float64 {
	// Calculate BMR (Basal Metabolic Rate) using Harris-Benedict equation
	age := int((time.Now().Unix() - user.Born) / (365 * 24 * 60 * 60))
	var bmr float64
	if user.Gender == "male" {
		bmr = 88.362 + (13.397 * user.Weight) + (4.799 * user.Height * 100) - (5.677 * float64(age))
	} else {
		bmr = 447.593 + (9.247 * user.Weight) + (3.098 * user.Height * 100) - (4.330 * float64(age))
	}

	// Calculate total time and distance
	//totalTime, totalDistance := computeTimeAndDistance(segments)

	// Calculate average speed in km/h
	averageSpeed := totalDistance / totalTime.Hours()

	//https://en.wikipedia.org/wiki/Metabolic_equivalent_of_task ...

	// Estimate MET (Metabolic Equivalent of Task) based on average speed
	var met float64
	switch {
	case averageSpeed < 4:
		met = 2.3 // walking, 2 mph (3.2 km/h), level surface
	case averageSpeed < 5:
		met = 3.6 // walking, 3 mph (4.8 km/h), level surface
	case averageSpeed < 7:
		met = 5.0 // walking, 4 mph (6.4 km/h), level surface
	case averageSpeed < 8:
		met = 7.0 // jogging, general
	case averageSpeed < 9:
		met = 8.0 // running, 5 mph (8 km/h)
	case averageSpeed < 10:
		met = 9.8 // running, 6 mph (9.7 km/h)
	default:
		met = 11.8 // running, 7 mph (11.3 km/h)
	}

	// Calculate calories burned
	//caloriesPerMinute := (met * 3.5 * user.Weight) / 200
	//totalCalories := caloriesPerMinute * totalTime.Minutes()

	// Calculate calories burned
	bmrPerMinute := bmr / 1440 // BMR per minute (1440 minutes in a day)
	activityCaloriesPerMinute := ((met - 1) * 3.5 * user.Weight) / 200
	totalCalories := (bmrPerMinute + activityCaloriesPerMinute) * totalTime.Minutes()

	return totalCalories
}

func getSpeedColor(normalizeValue float64) color.RGBA {

	var r, g uint8
	if normalizeValue < 0.5 {
		r = 255
		g = uint8(normalizeValue * 255)
	} else {
		r = uint8((1 - normalizeValue) * 255)
		g = 255
	}

	g /= 2 //darker green
	return color.RGBA{r, g, 0, 255}

}

func speedToColor(speed, minSpeed, maxSpeed, avgSpeed float64) color.RGBA {

	var normalizedSpeed float64
	if speed < avgSpeed {
		normalizedSpeed = (avgSpeed - speed) / (avgSpeed - minSpeed) / 2
	} else {
		normalizedSpeed = 0.5 + (speed-avgSpeed)/(maxSpeed-avgSpeed)/2
	}

	return getSpeedColor(normalizedSpeed)
}

func computeSpeedColors(segments []OsmMapSegment) {

	var minSpeed, maxSpeed float64
	var speeds []float64

	for _, segment := range segments {

		if len(segment.Trkpts) < 2 {
			continue
		}

		for i := 1; i < len(segment.Trkpts); i++ {
			prev := segment.Trkpts[i-1]
			curr := segment.Trkpts[i]

			distance := Activities_haversine(prev.Lat, prev.Lon, curr.Lat, curr.Lon)
			prevTime, _ := time.Parse(time.RFC3339, prev.Time)
			currTime, _ := time.Parse(time.RFC3339, curr.Time)
			duration := currTime.Sub(prevTime).Hours()

			speed := distance / duration
			speeds = append(speeds, speed)

			if i == 1 || speed < minSpeed {
				minSpeed = speed
			}
			if i == 1 || speed > maxSpeed {
				maxSpeed = speed
			}
		}
	}

	var avgSpeed float64
	for i := range speeds {
		avgSpeed += speeds[i]
	}
	avgSpeed /= float64(len(speeds))

	speeds_pos := 0
	for _, segment := range segments {
		for i := 1; i < len(segment.Trkpts); i++ {
			segment.Trkpts[i].Cd = speedToColor(speeds[speeds_pos], minSpeed, maxSpeed, avgSpeed)
			speeds_pos++
		}
		segment.Trkpts[0].Cd = segment.Trkpts[1].Cd // Set first point color same as second
	}
}

/*func smoothSegments(segments []MapSegment, windowSize int) []MapSegment {
	smoothedSegments := make([]MapSegment, len(segments))

	for i, segment := range segments {
		smoothedSegments[i] = smoothSegment(segment, windowSize)
	}

	return smoothedSegments
}

func smoothSegment(segment MapSegment, windowSize int) MapSegment {
	smoothed := MapSegment{
		Label:  segment.Label,
		Trkpts: make([]MapSegmentTrk, len(segment.Trkpts)),
	}

	sigma := float64(windowSize) / 6.0 // Adjust this value to control smoothing strength

	for i := range segment.Trkpts {
		var sumLat, sumLon, sumEle, weightSum float64

		for j := -windowSize; j <= windowSize; j++ {
			idx := i + j
			if idx < 0 || idx >= len(segment.Trkpts) {
				continue
			}

			weight := gaussianWeight(float64(j), sigma)
			sumLat += segment.Trkpts[idx].Lat * weight
			sumLon += segment.Trkpts[idx].Lon * weight
			sumEle += segment.Trkpts[idx].Ele * weight
			weightSum += weight
		}

		smoothed.Trkpts[i] = MapSegmentTrk{
			Lat:  sumLat / weightSum,
			Lon:  sumLon / weightSum,
			Ele:  sumEle / weightSum,
			Time: segment.Trkpts[i].Time, // Keep original time
			Cd:   segment.Trkpts[i].Cd,   // Keep original color
		}
	}

	return smoothed
}

func gaussianWeight(x, sigma float64) float64 {
	return math.Exp(-(x * x) / (2 * sigma * sigma))
}
*/
