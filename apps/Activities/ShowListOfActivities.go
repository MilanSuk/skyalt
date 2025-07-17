package main

import (
	"fmt"
	"strconv"
	"time"
)

// Shows list of Activities filtered by parameters.
type ShowListOfActivities struct {
	DateStart string //Select only runs after this data. Date format: YYYY-MM-DD HH:MM [optional]
	DateEnd   string //Select only runs before this data. Date format: YYYY-MM-DD HH:MM [optional]

	SortBy        string //Sort activities by. [options: date, distance, duration] [optional]
	SortAscending bool   //Ascending order for SortBy parameter. [optional]

	MaxNumberOfItems int //Maximum number of items returned. Zero if ignored. [optional]
}

func (st *ShowListOfActivities) run(caller *ToolCaller, ui *UI) error {
	source_activities, err := NewActivities("")
	if err != nil {
		return err
	}

	if st.SortBy == "" {
		st.SortBy = "date"
	}

	sorted, err := source_activities.Filter(st.DateStart, st.DateEnd, st.SortBy, st.SortAscending, st.MaxNumberOfItems)
	if err != nil {
		return err
	}

	ui.SetColumn(0, 10, 20)

	ui.AddTextLabel(0, 0, 1, 1, "List of activities")

	y := 1
	//header
	{
		HeadDiv := _GetListOfActivities_SetRow(ui.AddLayout(0, y, 1, 1), "")
		HeadDiv.AddText(0, 0, 1, 1, "<i>Date")
		HeadDiv.AddText(1, 0, 1, 1, "<i>Type")
		HeadDiv.AddText(2, 0, 1, 1, "<i>Description")
		HeadDiv.AddText(3, 0, 1, 1, "<i>Distance(km)")
		HeadDiv.AddText(4, 0, 1, 1, "<i>Duration(h:m:s)")
		y++
	}

	//activities
	sum_time := 0.0
	sum_distance := 0.0
	for _, id := range sorted {
		it := source_activities.Activities[id]

		tm := time.Duration(it.Duration * float64(time.Second))

		LineDiv := _GetListOfActivities_SetRow(ui.AddLayout(0, y, 1, 1), id)
		LineDiv.AddText(0, 0, 1, 1, SdkGetDateTime(int64(it.Date)))
		cb := LineDiv.AddCombo(1, 0, 1, 1, &it.Type, source_activities.GetTypeLabels(), source_activities.GetTypeValues())
		cb.changed = func() error {
			return nil
		}
		//cb.DialogWidth = 3
		ed := LineDiv.AddEditboxString(2, 0, 1, 1, &it.Description)
		ed.changed = func() error {
			return nil
		}
		LineDiv.AddText(3, 0, 1, 1, strconv.FormatFloat(it.Distance/1000, 'f', 3, 64))
		LineDiv.AddText(4, 0, 1, 1, fmt.Sprintf("%d:%02d:%02d", int(tm.Hours()), int(tm.Minutes())%60, int(tm.Seconds())%60))

		y++

		sum_distance += it.Distance
		sum_time += tm.Seconds()
	}

	ui.SetRow(y, 0.5, 0.5)
	ui.AddDivider(0, y, 1, 1, true)
	y++

	//sum
	{
		tm := time.Duration(sum_time * float64(time.Second))
		SumDiv := _GetListOfActivities_SetRow(ui.AddLayout(0, y, 1, 1), "")
		SumDiv.AddText(0, 0, 1, 1, fmt.Sprintf("<i>%d activities", len(sorted)))
		SumDiv.AddText(1, 0, 1, 1, "")
		SumDiv.AddText(2, 0, 1, 1, "")
		SumDiv.AddText(3, 0, 1, 1, "<i>"+strconv.FormatFloat(sum_distance/1000, 'f', 3, 64))
		SumDiv.AddText(4, 0, 1, 1, fmt.Sprintf("<i>%d:%02d:%02d", int(tm.Hours()), int(tm.Minutes())%60, int(tm.Seconds())%60))
	}

	return nil
}

func _GetListOfActivities_SetRow(ui *UI, activity_ID string) *UI {
	if activity_ID != "" {
		ui.Tooltip = fmt.Sprintf("ActivityID: %s", activity_ID)
	}

	ui.SetColumn(0, 1, 4)
	ui.SetColumn(1, 1, 4)
	ui.SetColumn(2, 1, 10)
	ui.SetColumn(3, 1, 3)
	ui.SetColumn(4, 1, 3)

	return ui
}
