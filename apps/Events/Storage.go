package main

type Color struct {
	R uint8
	G uint8
	B uint8
	A uint8
}

type Event struct {
	EventID     string
	Title       string
	Description string
	Files       []string
	Start       int64
	Duration    int64
	GroupID     string
}

type Group struct {
	GroupID string
	Label   string
	Color   Color
}

type Events struct {
	Items map[string]Event
}

type Groups struct {
	Items map[string]Group
}

func LoadEvents() (*Events, error) {
	ev := &Events{
		Items: make(map[string]Event),
	}
	return ReadJSONFile("Events.json", ev)
}

func LoadGroups() (*Groups, error) {
	gs := &Groups{
		Items: make(map[string]Group),
	}
	return ReadJSONFile("Groups.json", gs)
}

//<other storage functions here>
