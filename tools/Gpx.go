package main

import (
	"encoding/xml"
	"image/color"
	"math"
	"time"
)

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

func NewGPX(file string, caller *ToolCaller) (*Gpx, error) {
	st := &Gpx{}
	return _loadInstance(file, "Gpx", "xml", st, false, caller)
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

func (gpx *Gpx) getGPXSegments() (segments []UIOsmMapSegment) {
	for _, track := range gpx.Trk {
		for _, seg := range track.Segment {
			mapSeg := UIOsmMapSegment{
				Label:  track.Name,
				Trkpts: make([]UIOsmMapSegmentTrk, len(seg.Points)),
			}
			for i, point := range seg.Points {
				mapSeg.Trkpts[i] = UIOsmMapSegmentTrk{
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

func (gpx *Gpx) computeSpeedColors(segments []UIOsmMapSegment) {

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
