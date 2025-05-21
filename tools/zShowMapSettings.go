package main

// Shows Map settings(Enable, Tiles url, copyright, etc.)
type ShowMapSettings struct {
}

func (st *ShowMapSettings) run(caller *ToolCaller, ui *UI) error {
	source_map, err := NewMapSettings("", caller)
	if err != nil {
		return err
	}

	ui.SetColumn(0, 2, 4)
	ui.SetColumn(1, 3, 16)

	ui.AddTextLabel(0, 0, 2, 1, "Map settings")

	y := 1

	sw := ui.AddSwitch(1, y, 1, 1, "Enable", &source_map.Enable)
	sw.changed = func() error {
		return nil
	}
	y++

	ui.AddText(0, y, 1, 1, "Tiles URL")
	ed := ui.AddEditboxString(1, y, 1, 1, &source_map.Tiles_url)
	ed.changed = func() error {
		return nil
	}
	y++

	ui.AddText(0, y, 1, 1, "Copyright")
	ed = ui.AddEditboxString(1, y, 1, 1, &source_map.Copyright)
	ed.changed = func() error {
		return nil
	}
	y++

	ui.AddText(0, y, 1, 1, "Copyright_url")
	ed = ui.AddEditboxString(1, y, 1, 1, &source_map.Copyright_url)
	ed.changed = func() error {
		return nil
	}
	y++

	return nil
}
