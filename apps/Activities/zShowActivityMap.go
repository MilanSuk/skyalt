package main

import (
	"fmt"
)

// Shows map ui with the activity file
type ShowActivityMap struct {
	ActivityID string //ID of the activity

	Out_CamLon  float64
	Out_CamLat  float64
	Out_CamZoom float64
}

func (st *ShowActivityMap) run(caller *ToolCaller, ui *UI) error {
	source_activities, err := NewActivities("")
	if err != nil {
		return err
	}

	activity, found := source_activities.Activities[st.ActivityID]
	if !found {
		return fmt.Errorf("activity '%s' not found", st.ActivityID)
	}

	// load
	gpx, err := source_activities.GetGpx(st.ActivityID)
	if err != nil {
		return err
	}
	segments := gpx.getGPXSegments()

	ui.SetColumn(0, 1, 100)
	ui.SetRow(1, 10, 10)

	ui.AddTextLabel(0, 0, 1, 1, fmt.Sprintf("Map - %s", caller.ConvertTextDate(int64(activity.Date))))

	mp := ui.AddOsmMap(0, 1, 1, 1, &st.Out_CamLon, &st.Out_CamLat, &st.Out_CamZoom)
	if len(segments) > 0 {
		mp.AddRoute(UIOsmMapRoute{Segments: segments})
	}

	return nil
}
