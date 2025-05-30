package main

import "strconv"

// Show list of e-mail credentials.
type ShowListOfEmails struct {
}

func (st *ShowListOfEmails) run(caller *ToolCaller, ui *UI) error {
	source_emails, err := NewEmails("")
	if err != nil {
		return err
	}

	ui.SetColumn(0, 3, 7)
	ui.SetColumn(1, 3, 8)
	ui.SetColumn(2, 2, 3)

	ui.AddTextLabel(0, 0, 2, 1, "List of e-mails")

	y := 1
	ui.AddText(0, y, 1, 1, "<i>Username")
	ui.AddText(1, y, 1, 1, "<i>Server")
	ui.AddText(2, y, 1, 1, "<i>Port")
	y++

	for username, it := range source_emails.Logins {
		ui.AddText(0, y, 1, 1, username)
		ui.AddText(1, y, 1, 1, it.Server)
		ui.AddText(2, y, 1, 1, strconv.Itoa(it.Port))
		y++
	}

	return nil
}
