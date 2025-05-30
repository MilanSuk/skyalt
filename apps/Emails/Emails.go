package main

type EmailsLogin struct {
	Password string
	Server   string
	Port     int
}

// List of e-mails login credentials.
type Emails struct {
	Logins map[string]*EmailsLogin
}

func NewEmails(file string) (*Emails, error) {
	st := &Emails{}
	return _loadInstance(file, "Emails", "json", st, true)
}
