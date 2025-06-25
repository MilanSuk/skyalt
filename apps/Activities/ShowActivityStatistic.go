package main

import (
	"fmt"
	"time"
)

// Shows ui with statistic from the activity file
type ShowActivityStatistic struct {
	ActivityID string //activity ID
}

func (st *ShowActivityStatistic) run(caller *ToolCaller, ui *UI) error {
	source_activities, err := NewActivities("")
	if err != nil {
		return err
	}

	activity, found := source_activities.Activities[st.ActivityID]
	if !found {
		return fmt.Errorf("activity '%s' not found", st.ActivityID)
	}

	ui.SetColumn(0, 1, 100)
	ui.SetRow(1, 2.5, 2.5)

	ui.AddTextLabel(0, 0, 1, 1, fmt.Sprintf("Statistic - %s", SdkGetDate(int64(activity.Date))))

	ui.Back_cd = UI_GetPalette().GetGrey(0.1)

	InfoDiv := ui.AddLayout(0, 1, 1, 1)
	InfoDiv.Back_cd = UI_GetPalette().GetGrey(0.1)
	InfoDiv.SetColumn(0, 1, 100)
	InfoDiv.SetColumn(1, 1, 100)
	InfoDiv.SetColumn(2, 1, 100)
	InfoDiv.SetColumn(3, 1, 100)
	InfoDiv.SetRow(0, 1, 100)
	InfoDiv.SetRow(1, 1, 100)

	//load
	gpx, err := source_activities.GetGpx(st.ActivityID)
	if err != nil {
		return err
	}

	//get info
	totalTime_sec, totalDistance_m, startDate := gpx.GetInfo()
	cals, err := st.computeCaloriesBurned(totalTime_sec, totalDistance_m, startDate, caller)
	if err != nil {
		return err
	}

	tx := InfoDiv.AddText(0, 0, 1, 1, "<h1>"+fmt.Sprintf("%.2f", totalDistance_m/1000))
	tx.Align_h = 1
	tx.Align_v = 2
	tx = InfoDiv.AddText(0, 1, 1, 1, "Distance(km)")
	tx.Align_h = 1
	tx.Align_v = 0

	tm := time.Duration(totalTime_sec * float64(time.Second))
	tx = InfoDiv.AddText(1, 0, 1, 1, "<h1>"+fmt.Sprintf("%d:%02d:%02d", int(tm.Hours()), int(tm.Minutes())%60, int(tm.Seconds())%60))
	tx.Align_h = 1
	tx.Align_v = 2
	tx = InfoDiv.AddText(1, 1, 1, 1, "Time")
	tx.Align_h = 1
	tx.Align_v = 0

	avgTm := time.Duration(totalTime_sec * float64(time.Second) / (totalDistance_m / 1000))
	tx = InfoDiv.AddText(2, 0, 1, 1, "<h1>"+fmt.Sprintf("%02d:%02d", int(avgTm.Minutes())%60, int(avgTm.Seconds())%60))
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

	return nil
}

func (st *ShowActivityStatistic) computeCaloriesBurned(totalTime_sec float64, totalDistance_m float64, startDate float64, caller *ToolCaller) (float64, error) {
	source_body, err := NewUserBodyMeasurements("")
	if err != nil {
		return 0, err
	}

	// Calculate BMR (Basal Metabolic Rate) using Harris-Benedict equation
	age := time.Unix(int64(startDate), 0).Year() - source_body.BornYear + 1
	var bmr float64
	if !source_body.Female {
		bmr = 88.362 + (13.397 * source_body.Weight) + (4.799 * source_body.Height * 100) - (5.677 * float64(age))
	} else {
		bmr = 447.593 + (9.247 * source_body.Weight) + (3.098 * source_body.Height * 100) - (4.330 * float64(age))
	}

	// Calculate average speed in km/h
	averageSpeed := (totalDistance_m / 1000) / (totalTime_sec / 3600) //km / hour

	//https://en.wikipedia.org/wiki/Metabolic_equivalent_of_task

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
	activityCaloriesPerMinute := ((met - 1) * 3.5 * source_body.Weight) / 200
	totalCalories := (bmrPerMinute + activityCaloriesPerMinute) * (totalTime_sec / 60) //per minutes

	return totalCalories, nil
}
