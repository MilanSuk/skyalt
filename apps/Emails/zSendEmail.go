package main

import (
	"fmt"
	"net/smtp"
)

// Sends an email to the specified recipient with the given subject and body.
type SendEmail struct {
	// Username(in format user@email.com) for SMTP authentication.
	Username string

	// Email address of the recipient.
	To string

	// The subject of the email.
	Subject string

	// The content of the email.
	Body string
}

func (st *SendEmail) run(caller *ToolCaller, ui *UI) error {
	source_emails, err := NewEmails("")
	if err != nil {
		return err
	}

	// Find Username in logins
	login, found := source_emails.Logins[st.Username]
	if !found {
		return fmt.Errorf("username '%s' not found", st.Username)
	}

	// Set up the SMTP client
	auth := smtp.PlainAuth("", st.Username, login.Password, login.Server)

	// Message
	message := "From: " + st.Username + "\r\n" +
		"To: " + st.To + "\r\n" +
		"Subject: " + st.Subject + "\r\n" +
		"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n" +
		st.Body

	// Connect to the server, authenticate, set the sender and recipient(s), and send the email
	err = smtp.SendMail(fmt.Sprintf("%s:%d", login.Server, login.Port), auth, st.Username, []string{st.To}, []byte(message))
	if err != nil {
		return fmt.Errorf("smtp.SendMail failed: %v", err)
	}

	return nil
}
