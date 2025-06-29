package main

// Adds new SMTP email credentials into the database.
type ShowAddEmailLogin struct {
	Username   string // Username(in format user@email.com) for SMTP authentication. Optional, default is "". [optional]
	SMTPServer string // Address of the SMTP server. Optional, default is "". [optional]
	SMTPPort   int    // Tort number of the SMTP server. Optional, default is 0. [optional]
}

func (st *ShowAddEmailLogin) run(caller *ToolCaller, ui *UI) error {
	source_emails, err := NewEmails("")
	if err != nil {
		return err
	}

	ui.SetColumn(0, 2, 4)
	ui.SetColumn(1, 3, 16)

	ui.AddTextLabel(0, 0, 2, 1, "Add new e-mail credentials")

	ui.AddText(0, 1, 1, 1, "E-mail Address")
	usr := ui.AddEditboxString(1, 1, 1, 1, &st.Username)
	usr.Ghost = "example@email.com"

	var password string
	ui.AddText(0, 2, 1, 1, "Password")
	ui.AddEditboxString(1, 2, 1, 1, &password)

	ui.AddText(0, 3, 1, 1, "Server")
	srv := ui.AddEditboxString(1, 3, 1, 1, &st.SMTPServer)
	srv.Ghost = "smtp.example.com"

	ui.AddText(0, 4, 1, 1, "Port")
	prt := ui.AddEditboxInt(1, 3, 4, 1, &st.SMTPPort)
	prt.Ghost = "587"

	bt := ui.AddButton(0, 6, 2, 1, "Add new e-mail")
	bt.clicked = func() error {
		//update
		source_emails.Logins[st.Username] = &EmailsLogin{Password: password, Server: st.SMTPServer, Port: st.SMTPPort}
		return nil
	}

	return nil
}
