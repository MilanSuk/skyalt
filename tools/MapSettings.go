package main

type MapSettings struct {
	Enable    bool
	Tiles_url string

	Copyright     string
	Copyright_url string
}

func NewMapSettings(file string, caller *ToolCaller) (*MapSettings, error) {
	st := &MapSettings{}

	st.Enable = true
	st.Tiles_url = "https://tile.openstreetmap.org/{z}/{x}/{y}.png"
	st.Copyright = "(c)OpenStreetMap contributors"
	st.Copyright_url = "https://www.openstreetmap.org/copyright"

	return _loadInstance(file, "MapSettings", "json", st, true, caller)
}
