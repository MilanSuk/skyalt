package main

type UserBodyMeasurements struct {
	Gender    string  `json:"gender"`
	BirthYear int     `json:"birth_year"`
	Height    float64 `json:"height"`
	Weight    float64 `json:"weight"`
}

func LoadUserBodyMeasurements() (*UserBodyMeasurements, error) {
	st := &UserBodyMeasurements{
		Gender:    "man",
		BirthYear: 2000,
		Height:    1.8,
		Weight:    75,
	}

	return ReadJSONFile("UserBodyMeasurements.json", st)
}

type Activity struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	StartDate   int64   `json:"start_date"`
	Duration    int     `json:"duration"`
	Distance    float64 `json:"distance"`
}

type Activities struct {
	Activities map[string]Activity `json:"activities"`
}

func LoadActivities() (*Activities, error) {
	st := &Activities{Activities: make(map[string]Activity)}

	return ReadJSONFile("Activities.json", st)
}

type Point struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Elevation float64 `json:"elevation"`
	Time      int64   `json:"time"`
}

type Segment struct {
	Points []Point `json:"points"`
}

type Track struct {
	Name     string    `json:"name"`
	Segments []Segment `json:"segments"`
}

type Gpx struct {
	Tracks []Track `json:"tracks"`
}

func LoadGpx(id string) (*Gpx, error) {
	st := &Gpx{}

	return ReadJSONFile(id+".json", st)
}
