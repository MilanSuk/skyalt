package main

import (
	"fmt"
	"sort"
)

type ShowListOfActivities struct {
	DateStart        string // optional, format: YYYY-MM-DD HH:MM [optional]
	DateEnd          string // optional, format: YYYY-MM-DD HH:MM [optional]
	SortBy           string // optional [optional][options: "date", "distance", "duration"]
	SortAscending    bool   // optional [optional]
	MaxNumberOfItems int    // optional, zero or negative to show all [optional]
}

func (st *ShowListOfActivities) run(caller *ToolCaller, ui *UI) error {
	sortBy := st.SortBy
	if sortBy == "" {
		sortBy = "date"
	}

	activities, err := LoadActivities()
	if err != nil {
		return err
	}

	filtered, err := FilterActivities(st.DateStart, st.DateEnd, sortBy, st.SortAscending, st.MaxNumberOfItems)
	if err != nil {
		return err
	}

	typeSet := make(map[string]struct{})
	for _, a := range activities.Activities {
		if a.Type != "" {
			typeSet[a.Type] = struct{}{}
		}
	}
	var types []string
	for t := range typeSet {
		types = append(types, t)
	}
	sort.Strings(types)
	if len(types) == 0 {
		types = []string{"running", "cycling", "walking", "hiking", "swimming"}
	}

	ui.addTextH1("List of activities")

	table := ui.addTable("")
	hdr := table.addLine("")
	hdr.addText("Date", "")
	hdr.addText("Type", "")
	hdr.addText("Description", "")
	hdr.addText("Distance(km)", "")
	hdr.addText("Duration(h:m:s)", "")
	hdr.addText("", "")
	table.addDivider()

	for i := range filtered {
		act := &filtered[i]
		line := table.addLine(fmt.Sprintf("ActivityID=%s", act.ID))

		line.addText(SdkGetDateTime(act.StartDate), "")

		line.addDropDown(&act.Type, types, types, "")

		line.addEditboxString(&act.Description, "")

		distStr := "-"
		if act.Distance > 0 {
			distStr = fmt.Sprintf("%.1f", act.Distance)
		}
		line.addText(distStr, "")

		dur := act.Duration
		h := dur / 3600
		m := (dur % 3600) / 60
		s := dur % 60
		durStr := "-"
		if dur > 0 {
			durStr = fmt.Sprintf("%d:%02d:%02d", h, m, s)
		}
		line.addText(durStr, "")

		line.addPromptMenu([]string{"Show elevation", "Delete activity"}, fmt.Sprintf("ActivityID=%s", act.ID))
	}

	return nil
}
