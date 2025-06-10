package main

// User gender, born year, height, weight
type UserBodyMeasurements struct {
	Female   bool
	BornYear int
	Height   float64 //meters
	Weight   float64 //kilograms
}

func NewUserBodyMeasurements(file string) (*UserBodyMeasurements, error) {
	st := &UserBodyMeasurements{}
	st.BornYear = 2000
	st.Female = true
	st.Height = 170
	st.Weight = 60

	return LoadFile(file, "UserBodyMeasurements", "json", st, true)
}
