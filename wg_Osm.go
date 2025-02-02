package main

type Osm struct {
	Enable     bool
	Tiles_url  string
	Cache_path string

	Copyright     string
	Copyright_url string
}

func (layout *Layout) AddOsm(x, y, w, h int, props *Osm) *Osm {
	layout._createDiv(x, y, w, h, "Osm", props.Build, nil, nil)
	return props
}

func (st *Osm) Build(layout *Layout) {
	if st.Tiles_url == "" {
		st.Tiles_url = "https://tile.openstreetmap.org/{z}/{x}/{y}.png"
	}
	if st.Cache_path == "" {
		st.Cache_path = "disk/osm_tiles_cache.sqlite"
	}
	if st.Copyright == "" {
		st.Copyright = "(c)OpenStreetMap contributors"
	}
	if st.Copyright_url == "" {
		st.Copyright_url = "https://www.openstreetmap.org/copyright"
	}

	layout.SetColumn(0, 1, 4)
	layout.SetColumn(1, 1, 20)

	y := 0

	layout.AddSwitch(1, y, 1, 1, "Enable", &st.Enable)
	y++

	layout.AddText(0, y, 1, 1, "Tiles URL")
	layout.AddEditbox(1, y, 1, 1, &st.Tiles_url)
	y++

	layout.AddText(0, y, 1, 1, "Cache path")
	layout.AddFilePickerButton(1, y, 1, 1, &st.Cache_path, true)
	y++

	layout.AddText(0, y, 1, 1, "Copyright")
	layout.AddEditbox(1, y, 1, 1, &st.Copyright)
	y++

	layout.AddText(0, y, 1, 1, "Copyright_url")
	layout.AddEditbox(1, y, 1, 1, &st.Copyright_url)
	y++

}
