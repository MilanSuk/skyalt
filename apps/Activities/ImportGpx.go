package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"time"
)

type gpx struct {
	XMLName xml.Name `xml:"gpx"`
	Tracks  []track  `xml:"trk"`
}

type track struct {
	Name     string    `xml:"name"`
	Segments []segment `xml:"trkseg"`
}

type segment struct {
	Points []point `xml:"trkpt"`
}

type point struct {
	Lat     float64 `xml:"lat,attr"`
	Lon     float64 `xml:"lon,attr"`
	Elev    float64 `xml:"ele"`
	TimeStr string  `xml:"time"`
}

func distance(lat1, lon1, lat2, lon2 float64) float64 {
	var r float64 = 6371000
	dLat := (lat2 - lat1) * (math.Pi / 180)
	dLon := (lon2 - lon1) * (math.Pi / 180)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return r * c
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return err
	}

	return destination.Sync()
}

func ImportGpx(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var g gpx
	err = xml.Unmarshal(data, &g)
	if err != nil {
		return "", err
	}

	st := &Gpx{
		Tracks: make([]Track, len(g.Tracks)),
	}
	for ti, t := range g.Tracks {
		st.Tracks[ti] = Track{
			Name:     t.Name,
			Segments: make([]Segment, len(t.Segments)),
		}
		for si, s := range t.Segments {
			st.Tracks[ti].Segments[si] = Segment{
				Points: make([]Point, len(s.Points)),
			}
			for pi, p := range s.Points {
				tm, err := time.Parse(time.RFC3339, p.TimeStr)
				if err != nil {
					return "", err
				}
				st.Tracks[ti].Segments[si].Points[pi] = Point{
					Latitude:  p.Lat,
					Longitude: p.Lon,
					Elevation: p.Elev,
					Time:      tm.Unix(),
				}
			}
		}
	}

	var minTime, maxTime int64 = math.MaxInt64, math.MinInt64
	var totalDistance float64
	hasPoints := false
	for _, trk := range st.Tracks {
		for _, seg := range trk.Segments {
			if len(seg.Points) < 1 {
				continue
			}
			hasPoints = true
			for _, p := range seg.Points {
				if p.Time < minTime {
					minTime = p.Time
				}
				if p.Time > maxTime {
					maxTime = p.Time
				}
			}
			for i := 0; i < len(seg.Points)-1; i++ {
				p1 := seg.Points[i]
				p2 := seg.Points[i+1]
				totalDistance += distance(p1.Latitude, p1.Longitude, p2.Latitude, p2.Longitude)
			}
		}
	}
	if !hasPoints {
		return "", errors.New("no points in GPX")
	}

	duration := int(maxTime - minTime)

	var description string
	if len(st.Tracks) > 0 {
		description = st.Tracks[0].Name
	}

	activities, err := LoadActivities()
	if err != nil {
		return "", err
	}

	var ID string
	for i := 1; ; i++ {
		ID = fmt.Sprintf("activity_%d", i)
		if _, ok := activities.Activities[ID]; !ok {
			break
		}
	}

	act := Activity{
		ID:          ID,
		Type:        "running",
		Description: description,
		StartDate:   minTime,
		Duration:    duration,
		Distance:    totalDistance,
	}
	activities.Activities[ID] = act

	jsonData, err := json.Marshal(st)
	if err != nil {
		return "", err
	}
	jsonFile, err := os.Create(ID + ".json")
	if err != nil {
		return "", err
	}
	defer jsonFile.Close()
	_, err = jsonFile.Write(jsonData)
	if err != nil {
		return "", err
	}
	err = jsonFile.Sync()
	if err != nil {
		return "", err
	}

	err = os.MkdirAll("Activities", 0755)
	if err != nil {
		return "", err
	}
	gpxDst := filepath.Join("Activities", ID+".gpx")
	err = copyFile(path, gpxDst)
	if err != nil {
		return "", err
	}

	return ID, nil
}
