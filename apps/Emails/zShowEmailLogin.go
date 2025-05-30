package main

import (
	"fmt"
	"net/url"
)

// Edit SMTP e-mail credentials. It shows form with attributes(password, server, port) for Username, which user can change.
type ShowEmailLogin struct {
	Username string // Username(in format user@email.com) for SMTP authentication.
}

func (st *ShowEmailLogin) run(caller *ToolCaller, ui *UI) error {
	source_emails, err := NewEmails("")
	if err != nil {
		return err
	}

	ui.SetColumn(0, 2, 4)
	ui.SetColumn(1, 3, 16)

	ui.AddTextLabel(0, 0, 2, 1, "Edit e-mail credentials")

	// Check if it exists
	login, found := source_emails.Logins[st.Username]
	if !found {
		return fmt.Errorf("username '%s' not found", st.Username)
	}

	ui.AddText(0, 1, 1, 1, "Username")
	ui.AddText(1, 1, 1, 1, st.Username)

	ui.AddText(0, 2, 1, 1, "Password")
	pass := ui.AddEditboxString(1, 2, 1, 1, &login.Password)
	pass.changed = func() error {
		if login.Password == "" {
			pass.Error = "Empty field"
		}
		return nil
	}

	ui.AddText(0, 3, 1, 1, "Server")
	srv := ui.AddEditboxString(1, 3, 1, 1, &login.Server)
	srv.Ghost = "smtp.example.com"
	srv.changed = func() error {
		if login.Server == "" {
			srv.Error = "Empty field"
		} else {
			_, err := url.ParseRequestURI(login.Server)
			if err != nil {
				srv.Error = "Invalid format: " + err.Error()
			}
		}
		return nil
	}

	ui.AddText(0, 4, 1, 1, "Port")
	prt := ui.AddEditboxInt(1, 4, 1, 1, &login.Port)
	prt.Ghost = "587"
	prt.changed = func() error {
		if login.Port == 0 {
			prt.Error = "Invalid port number"
		}
		return nil
	}

	return nil
}
