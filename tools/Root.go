package main

type RootTab struct {
	ChatID int64
	Label  string

	Use_sources []string
}

type Root struct {
	Show   string
	Memory string

	Autosend float64

	Tabs           []RootTab
	Selected_tab_i int
}

func NewRoot(file string, caller *ToolCaller) (*Root, error) {
	st := &Root{}

	return _loadInstance(file, "Root", "json", st, true, caller)
}
