package main

import (
	"encoding/xml"
	"image/color"
	"math"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type ActivitiesItem struct {
	Type        string
	Description string

	Date     float64 //unix
	Duration float64 //seconds
	Distance float64 //meters
}

// All users activities. Every activity has type, description, date, duration and distance.
type Activities struct {
	Activities map[string]*ActivitiesItem
}

func NewActivities(file string) (*Activities, error) {
	st := &Activities{}
	st.Activities = make(map[string]*ActivitiesItem)

	return LoadFile(file, "Activities", "json", st, true)
}

func (acts *Activities) GetTypeLabels() []string {
	return []string{"", "Run", "Ride", "Swim", "Walk", "Hike", "Inline", "Workout", "Ski", "Snowboard", "Pilates", "Tenis", "Yoga"}
}

func (acts *Activities) GetTypeValues() []string {
	labels := acts.GetTypeLabels()
	values := []string{}
	for _, it := range labels {
		values = append(values, strings.ToLower(it))
	}
	return values
}

func (acts *Activities) GetFilePath(ActivityID string) string {
	return filepath.Join("Activities", "Gpx-"+ActivityID+".xml")
}

func (acts *Activities) GetGpx(ActivityID string) (*Gpx, error) {
	gpx, err := NewGPX(acts.GetFilePath(ActivityID))
	if err != nil {
		return nil, err
	}
	return gpx, nil
}

func (acts *Activities) Filter(DateStart string, DateEnd string, SortBy string, SortAscending bool, MaxNumberOfItems int) ([]string, error) {
	var stTime, enTime float64

	//check
	if DateStart != "" {
		tm, err := time.ParseInLocation("2006-01-02 15:04", DateStart, time.Local)
		if err != nil {
			return nil, err
		}
		stTime = float64(tm.Unix())
	}
	if DateEnd != "" {
		tm, err := time.ParseInLocation("2006-01-02 15:04", DateEnd, time.Local)
		if err != nil {
			return nil, err
		}
		enTime = float64(tm.Unix())
	}

	var sorted []string
	for id, it := range acts.Activities {
		if stTime > 0 && it.Date < stTime {
			continue
		}
		if enTime > 0 && it.Date > enTime {
			continue
		}
		sorted = append(sorted, id)
	}

	// Sort
	switch SortBy {
	case "date":
		sort.Slice(sorted, func(i, j int) bool {
			a := acts.Activities[sorted[i]]
			b := acts.Activities[sorted[j]]
			return a.Date > b.Date //latest
		})

	case "distance":
		sort.Slice(sorted, func(i, j int) bool {
			a := acts.Activities[sorted[i]]
			b := acts.Activities[sorted[j]]
			return a.Distance > b.Distance //longest
		})
	case "duration":
		sort.Slice(sorted, func(i, j int) bool {
			a := acts.Activities[sorted[i]]
			b := acts.Activities[sorted[j]]
			return a.Duration > b.Duration //longest
		})
	}

	// Order
	if !SortAscending {
		for i := 0; i < len(sorted)/2; i++ {
			sorted[i], sorted[len(sorted)-i-1] = sorted[len(sorted)-i-1], sorted[i]
		}
	}

	// Cut by maximum number of items
	if MaxNumberOfItems > 0 {
		n := len(sorted)
		if n > MaxNumberOfItems {
			n = MaxNumberOfItems
		}
		sorted = sorted[:n]
	}

	return sorted, nil
}

func (acts *Activities) _importGPXFile(src_path string, tp string, description string) (string, error) {
	//copy file
	file := strconv.FormatInt(time.Now().UnixNano(), 10)
	err := OsCopyFile(acts.GetFilePath(file), src_path)
	if err != nil {
		return "", err
	}

	//load
	gpx, err := NewGPX(src_path)
	if err != nil {
		return "", err
	}
	//get stats
	totalTime, totalDistance_m, startTime := gpx.GetInfo()

	//add into array
	acts.Activities[file] = &ActivitiesItem{Type: tp, Description: description, Date: startTime, Duration: totalTime, Distance: totalDistance_m}

	return file, nil
}

type Point struct {
	Lat  float64 `xml:"lat,attr"`
	Lon  float64 `xml:"lon,attr"`
	Ele  float64 `xml:"ele"`
	Time string  `xml:"time"`
}
type Segment struct {
	Points []Point `xml:"trkpt"`
}
type Track struct {
	Name    string    `xml:"name"`
	Segment []Segment `xml:"trkseg"`
}

// GPX with recorded activity.
type Gpx struct {
	XMLName xml.Name `xml:"gpx"`
	Trk     []Track  `xml:"trk"`
}

func NewGPX(file string) (*Gpx, error) {
	st := &Gpx{}
	return LoadFile(file, "Gpx", "xml", st, false)
}

func (gpx *Gpx) GetInfo() (float64, float64, float64) {

	var totalTime time.Duration
	var totalDistance_m float64
	var startTime time.Time

	startTimeSet := false

	for _, track := range gpx.Trk {
		for _, seg := range track.Segment {
			for i := 1; i < len(seg.Points); i++ {
				prevPoint := seg.Points[i-1]
				currentPoint := seg.Points[i]

				// Compute time difference
				prevTime, _ := time.Parse(time.RFC3339, prevPoint.Time)
				currentTime, _ := time.Parse(time.RFC3339, currentPoint.Time)
				totalTime += currentTime.Sub(prevTime)

				// Compute distance using Haversine formula
				distance := gpx.haversine(prevPoint.Lat, prevPoint.Lon, currentPoint.Lat, currentPoint.Lon)
				totalDistance_m += distance * 1000

				if !startTimeSet {
					startTime = prevTime
					startTimeSet = true
				}
			}
		}
	}

	return totalTime.Seconds(), totalDistance_m, float64(float64(startTime.UnixMilli()) / 1000)
}

func (gpx *Gpx) toRadians(deg float64) float64 {
	return deg * (math.Pi / 180)
}
func (gpx *Gpx) haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371 // km

	dLat := gpx.toRadians(lat2 - lat1)
	dLon := gpx.toRadians(lon2 - lon1)
	lat1 = gpx.toRadians(lat1)
	lat2 = gpx.toRadians(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1)*math.Cos(lat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}

func (gpx *Gpx) getGPXSegments() (segments []UIMapSegment) {
	for _, track := range gpx.Trk {
		for _, seg := range track.Segment {
			mapSeg := UIMapSegment{
				Label:  track.Name,
				Trkpts: make([]UIMapSegmentTrk, len(seg.Points)),
			}
			for i, point := range seg.Points {
				mapSeg.Trkpts[i] = UIMapSegmentTrk{
					Lon:  point.Lon,
					Lat:  point.Lat,
					Ele:  point.Ele,
					Time: point.Time,
				}
			}

			segments = append(segments, mapSeg)
		}
	}

	gpx.computeSpeedColors(segments)
	return
}

func (gpx *Gpx) getSpeedColor(normalizeValue float64) color.RGBA {

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

func (gpx *Gpx) speedToColor(speed, minSpeed, maxSpeed, avgSpeed float64) color.RGBA {

	var normalizedSpeed float64
	if speed < avgSpeed {
		normalizedSpeed = (avgSpeed - speed) / (avgSpeed - minSpeed) / 2
	} else {
		normalizedSpeed = 0.5 + (speed-avgSpeed)/(maxSpeed-avgSpeed)/2
	}

	return gpx.getSpeedColor(normalizedSpeed)
}

func (gpx *Gpx) computeSpeedColors(segments []UIMapSegment) {

	var minSpeed, maxSpeed float64
	var speeds []float64

	for _, segment := range segments {

		if len(segment.Trkpts) < 2 {
			continue
		}

		for i := 1; i < len(segment.Trkpts); i++ {
			prev := segment.Trkpts[i-1]
			curr := segment.Trkpts[i]

			distance := gpx.haversine(prev.Lat, prev.Lon, curr.Lat, curr.Lon)
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
			segment.Trkpts[i].Cd = gpx.speedToColor(speeds[speeds_pos], minSpeed, maxSpeed, avgSpeed)
			speeds_pos++
		}
		segment.Trkpts[0].Cd = segment.Trkpts[1].Cd // Set first point color same as second
	}
}

// User gender, born year, height, weight
type UserBodyMeasurements struct {
	Female   bool
	BornYear int
	Height   float64 //meters
	Weight   float64 //kilograms
}

func NewUserBodyMeasurements(file string) (*UserBodyMeasurements, error) {
	st := &UserBodyMeasurements{}
	st.BornYear = 2000
	st.Female = true
	st.Height = 170
	st.Weight = 60

	return LoadFile(file, "UserBodyMeasurements", "json", st, true)
}
