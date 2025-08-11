package main

import (
	"errors"
	"sort"
	"time"
)

func FilterActivities(dateStart, dateEnd, sortBy string, sortAscending bool, maxNumberOfItems int) ([]Activity, error) {
	acts, err := LoadActivities()
	if err != nil {
		return nil, err
	}

	var startUnix *int64
	if dateStart != "" {
		t, err := time.Parse(time.RFC3339, dateStart)
		if err != nil {
			return nil, err
		}
		u := t.Unix()
		startUnix = &u
	}

	var endUnix *int64
	if dateEnd != "" {
		t, err := time.Parse(time.RFC3339, dateEnd)
		if err != nil {
			return nil, err
		}
		u := t.Unix()
		endUnix = &u
	}

	var filtered []Activity
	for _, act := range acts.Activities {
		match := true
		if startUnix != nil && act.StartDate < *startUnix {
			match = false
		}
		if endUnix != nil && act.StartDate > *endUnix {
			match = false
		}
		if match {
			filtered = append(filtered, act)
		}
	}

	validSortBy := map[string]bool{"date": true, "distance": true, "duration": true}
	if _, ok := validSortBy[sortBy]; !ok {
		return nil, errors.New("invalid sortBy")
	}

	sort.Slice(filtered, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "date":
			less = filtered[i].StartDate < filtered[j].StartDate
		case "distance":
			less = filtered[i].Distance < filtered[j].Distance
		case "duration":
			less = filtered[i].Duration < filtered[j].Duration
		}
		if !sortAscending {
			return !less
		}
		return less
	})

	if maxNumberOfItems > 0 && len(filtered) > maxNumberOfItems {
		filtered = filtered[:maxNumberOfItems]
	}

	return filtered, nil
}
