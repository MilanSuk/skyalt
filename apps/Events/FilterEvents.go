package main

// Assuming necessary imports are already handled in the project,
// such as for error handling and the storage package.

func FilterEvents(startTime int64, endTime int64, groupIDs []string) ([]string, error) {
	events, err := LoadEvents()
	if err != nil {
		return nil, err
	}

	var filteredIDs []string

	for _, event := range events.Items {
		eventEnd := event.Start + event.Duration
		if event.Start >= startTime && eventEnd <= endTime {
			if len(groupIDs) == 0 || contains(groupIDs, event.GroupID) {
				filteredIDs = append(filteredIDs, event.EventID)
			}
		}
	}

	return filteredIDs, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
