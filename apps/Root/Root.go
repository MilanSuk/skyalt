package main

type RootChat struct {
	FileName string
	Label    string
}

type RootApp struct {
	Name            string
	Chats           []RootChat
	Selected_chat_i int
}

// Root
type Root struct {
	Mode   string
	Memory string

	Autosend float64

	Apps           []*RootApp
	Selected_app_i int
}

func NewRoot(file string, caller *ToolCaller) (*Root, error) {
	st := &Root{}

	return _loadInstance(file, "Root", "json", st, true, caller)
}
