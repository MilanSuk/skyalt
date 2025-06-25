package main

import "fmt"

// Deletes SMTP e-mail credentials from the database.
type DeleteEmailLogin struct {
	Username string // Username(in format user@email.com) for SMTP authentication.
}

func (st *DeleteEmailLogin) run(caller *ToolCaller, ui *UI) error {
	source_emails, err := NewEmails("")
	if err != nil {
		return err
	}

	// Check if it exists
	_, found := source_emails.Logins[st.Username]
	if !found {
		return fmt.Errorf("username '%s' not found", st.Username)
	}

	// Remove items
	delete(source_emails.Logins, st.Username)

	return nil
}
