package main

import (
	"encoding/csv"
	"strconv"
	"strings"
	"time"
)

// Returns list of runs in CSV format. First line is columns description.
type GetListOfActivities struct {
	DateStart string //Select only runs after this data. Date format: YYYY-MM-DD HH:MM [optional]
	DateEnd   string //Select only runs before this data. Date format: YYYY-MM-DD HH:MM [optional]

	SortBy        string //Sort activities by. [options: date, distance, duration] [optional]
	SortAscending bool   //Ascending order for SortBy parameter. [optional]

	MaxNumberOfItems int //Maximum number of items returned. Zero if ignored. [optional]

	Out_activities string
}

func (st *GetListOfActivities) run(caller *ToolCaller, ui *UI) error {
	source_activities, err := NewActivities("")
	if err != nil {
		return err
	}

	sorted, err := source_activities.Filter(st.DateStart, st.DateEnd, st.SortBy, st.SortAscending, st.MaxNumberOfItems)
	if err != nil {
		return err
	}

	// Build CSV output
	str := new(strings.Builder)
	w := csv.NewWriter(str)

	//Description
	err = w.Write([]string{"ID", "Date", "Distance", "Duration", "Description"})
	if err != nil {
		return err
	}
	//Items
	for _, id := range sorted {
		it := source_activities.Activities[id]
		ln := []string{id, time.Unix(int64(it.Date), 0).Format("Mon, 02 Jan 2006 15:04"), strconv.FormatFloat(it.Distance, 'f', -1, 64), strconv.FormatFloat(it.Duration, 'f', -1, 64), it.Description}

		err := w.Write(ln)
		if err != nil {
			return err
		}
	}

	w.Flush()
	err = w.Error()
	if err != nil {
		return err
	}

	st.Out_activities = str.String()
	return nil
}
