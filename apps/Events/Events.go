package main

import (
	"bufio"
	"fmt"
	"image/color"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type EventsItem struct {
	Title       string
	Description string
	Files       []string //attachments

	Start    int64 //unix time
	Duration int64 //seconds

	GroupID int64
}
type EventsGroup struct {
	Label string
	Color color.RGBA
}

// List of events and groups. Every event has title, description, files, start, duration and GroupID. Every Group has label and color.
type Events struct {
	Events map[int64]*EventsItem
	Groups map[int64]*EventsGroup
}

func NewEvents(file string, caller *ToolCaller) (*Events, error) {
	st := &Events{}
	st.Events = make(map[int64]*EventsItem)
	st.Groups = make(map[int64]*EventsGroup)
	return _loadInstance(file, "Events", "json", st, true, caller)
}

func (ev *Events) getSortedGroupIDs() (sorted []int64) {
	for k := range ev.Groups {
		sorted = append(sorted, k)
	}
	sort.Slice(sorted, func(i, j int) bool {
		a := sorted[i]
		b := sorted[j]
		return a < b
	})
	return
}

func (ev *Events) getGroupsLabels() (labels []string) {
	sorted := ev.getSortedGroupIDs()
	for _, id := range sorted {
		group := ev.Groups[id]
		labels = append(labels, group.Label)
	}
	return
}
func (ev *Events) getGroupsValues() (values []string) {
	sorted := ev.getSortedGroupIDs()
	for _, id := range sorted {
		values = append(values, strconv.FormatInt(id, 10))
	}
	return
}

func (ev *Events) filterEvents(ST, EN int64, groupsIDs []int) []int64 {
	var ret []int64

	//filter
	for id, it := range ev.Events {
		//groups
		if len(groupsIDs) > 0 {
			found := false
			for _, id := range groupsIDs {
				if it.GroupID == int64(id) {
					found = true
					break
				}
			}
			if !found {
				continue //skip
			}
		}

		//time
		itEnd := (it.Start + it.Duration)
		if (it.Start >= ST && it.Start < EN) ||
			(itEnd >= ST && itEnd < EN) ||
			(it.Start < ST && itEnd > EN) {
			ret = append(ret, id)
		}
	}

	//sort
	sort.Slice(ret, func(i, j int) bool {
		a := ev.Events[ret[i]]
		b := ev.Events[ret[j]]
		return a.Start < b.Start
	})

	return ret
}

func (ev *Events) Filter(DateStart string, DateEnd string) ([]int64, error) {
	var stTime, enTime int64

	//check
	if DateStart != "" {
		tm, err := time.ParseInLocation("2006-01-02 15:04", DateStart, time.Local)
		if err != nil {
			return nil, err
		}
		stTime = tm.Unix()
	}
	if DateEnd != "" {
		tm, err := time.ParseInLocation("2006-01-02 15:04", DateEnd, time.Local)
		if err != nil {
			return nil, err
		}
		enTime = tm.Unix()
	}

	var filtered []int64
	for id, it := range ev.Events {
		if stTime > 0 && it.Start < stTime {
			continue
		}
		if enTime > 0 && (it.Start+it.Duration) > enTime {
			continue
		}
		filtered = append(filtered, id)
	}

	return filtered, nil
}

func (ev *Events) ExportICS(events []int64, filename string) error {
	// Start building iCalendar content
	var builder strings.Builder

	// Write header
	builder.WriteString("BEGIN:VCALENDAR\r\n")
	builder.WriteString("VERSION:2.0\r\n")
	builder.WriteString("PRODID:-//Your Company//Your Product//EN\r\n")

	// Process each event
	for _, event_id := range events {
		event := ev.Events[event_id]

		builder.WriteString("BEGIN:VEVENT\r\n")

		// Generate unique identifier
		builder.WriteString(fmt.Sprintf("UID:%d-%s@yourdomain.com\r\n", event.Start, strings.ReplaceAll(event.Title, " ", "-")))

		// Convert Unix timestamp to UTC time
		startTime := time.Unix(event.Start, 0).UTC()
		endTime := startTime.Add(time.Duration(event.Duration) * time.Second)

		// Format dates in iCalendar format
		builder.WriteString(fmt.Sprintf("DTSTART:%s\r\n", startTime.Format("20060102T150405Z")))
		builder.WriteString(fmt.Sprintf("DTEND:%s\r\n", endTime.Format("20060102T150405Z")))

		// Add creation timestamp
		builder.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", time.Now().UTC().Format("20060102T150405Z")))

		// Add title and description
		builder.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", ev.escapeICSText(event.Title)))
		if event.Description != "" {
			builder.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", ev.escapeICSText(event.Description)))
		}

		// Add color if specified
		cd := ev.findGroupColor(event.GroupID)
		if cd != (color.RGBA{}) {
			builder.WriteString(fmt.Sprintf("COLOR:#%02X%02X%02X\r\n", cd.R, cd.G, cd.B))
		}

		// Add attachments if any
		for _, file := range event.Files {
			builder.WriteString(fmt.Sprintf("ATTACH:%s\r\n", file))
		}

		builder.WriteString("END:VEVENT\r\n")
	}

	// Write footer
	builder.WriteString("END:VCALENDAR\r\n")

	// Write to file
	return os.WriteFile(filename, []byte(builder.String()), 0644)
}

func (ev *Events) findGroupColor(groupID int64) color.RGBA {
	group, found := ev.Groups[groupID]
	if found {
		return group.Color
	}
	return color.RGBA{}
}
func (ev *Events) findGroupColorOrDefault(groupID int64, caller *ToolCaller) color.RGBA {
	cd := ev.findGroupColor(groupID)
	if cd != (color.RGBA{}) {
		return cd
	}
	return UI_GetPalette().P
}

func (ev *Events) escapeICSText(text string) string {
	text = strings.ReplaceAll(text, "\\", "\\\\")
	text = strings.ReplaceAll(text, ";", "\\;")
	text = strings.ReplaceAll(text, ",", "\\,")
	text = strings.ReplaceAll(text, "\n", "\\n")
	return text
}

func (ev *Events) ImportICS(filePath string, groupID int64) ([]*EventsItem, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var events []*EventsItem
	var currentEvent *EventsItem

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check for event start/end
		if line == "BEGIN:VEVENT" {
			currentEvent = &EventsItem{GroupID: groupID}
			continue
		}
		if line == "END:VEVENT" {
			if currentEvent != nil {
				events = append(events, currentEvent)
				currentEvent = nil
			}
			continue
		}

		// Skip if we're not inside an event
		if currentEvent == nil {
			continue
		}

		// Parse event properties
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.ToUpper(parts[0])
		value := parts[1]

		switch key {
		case "SUMMARY":
			currentEvent.Title = value
		case "DESCRIPTION":
			currentEvent.Description = value
		case "ATTACH":
			currentEvent.Files = append(currentEvent.Files, value)
		case "DTSTART":
			timestamp, err := ev._parseICSTime(value)
			if err == nil {
				currentEvent.Start = timestamp
			}
		case "DURATION":
			duration, err := ev._parseICSDuration(value)
			if err == nil {
				currentEvent.Duration = duration
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (ev *Events) _parseICSTime(timeStr string) (int64, error) {
	// Assuming format: YYYYMMDDTHHMMSSZ
	timeStr = strings.TrimSuffix(timeStr, "Z")
	t, err := time.ParseInLocation("20060102T150405", timeStr, time.Local)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

func (ev *Events) _parseICSDuration(durationStr string) (int64, error) {
	// Assuming format: PnDTnHnMnS
	durationStr = strings.TrimPrefix(durationStr, "P")
	var seconds int64 = 0

	// Split into days and time parts
	parts := strings.Split(durationStr, "T")

	if len(parts) > 0 && parts[0] != "" {
		// Parse days
		days := strings.TrimSuffix(parts[0], "D")
		if d, err := strconv.ParseInt(days, 10, 64); err == nil {
			seconds += d * 24 * 60 * 60
		}
	}

	if len(parts) > 1 {
		timePart := parts[1]

		// Parse hours
		if idx := strings.Index(timePart, "H"); idx != -1 {
			if h, err := strconv.ParseInt(timePart[:idx], 10, 64); err == nil {
				seconds += h * 60 * 60
			}
			timePart = timePart[idx+1:]
		}

		// Parse minutes
		if idx := strings.Index(timePart, "M"); idx != -1 {
			if m, err := strconv.ParseInt(timePart[:idx], 10, 64); err == nil {
				seconds += m * 60
			}
			timePart = timePart[idx+1:]
		}

		// Parse seconds
		if idx := strings.Index(timePart, "S"); idx != -1 {
			if s, err := strconv.ParseInt(timePart[:idx], 10, 64); err == nil {
				seconds += s
			}
		}
	}

	return seconds, nil
}
