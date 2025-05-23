package main

import (
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

func NewActivities(file string, caller *ToolCaller) (*Activities, error) {
	st := &Activities{}
	st.Activities = make(map[string]*ActivitiesItem)

	return _loadInstance(file, "Activities", "json", st, true, caller)
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

func (acts *Activities) GetGpx(ActivityID string, caller *ToolCaller) (*Gpx, error) {
	gpx, err := NewGPX(acts.GetFilePath(ActivityID), caller)
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

func (acts *Activities) _importGPXFile(src_path string, tp string, description string, caller *ToolCaller) (string, error) {
	//copy file
	file := strconv.FormatInt(time.Now().UnixNano(), 10)
	err := OsCopyFile(acts.GetFilePath(file), src_path)
	if err != nil {
		return "", err
	}

	//load
	gpx, err := NewGPX(src_path, caller)
	if err != nil {
		return "", err
	}
	//get stats
	totalTime, totalDistance_m, startTime := gpx.GetInfo()

	//add into array
	acts.Activities[file] = &ActivitiesItem{Type: tp, Description: description, Date: startTime, Duration: totalTime, Distance: totalDistance_m}

	return file, nil
}
