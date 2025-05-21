package main

import (
	"encoding/csv"
	"strconv"
	"strings"
)

// Returns list of SMTP usernames(with server and port) in CSV format. First line is columns description.
type GetEmailLogins struct {
	Out_emails string
}

func (st *GetEmailLogins) run(caller *ToolCaller, ui *UI) error {
	source_emails, err := NewEmails("", caller)
	if err != nil {
		return err
	}

	// Build CSV output
	str := new(strings.Builder)
	w := csv.NewWriter(str)
	//Description
	err = w.Write([]string{"ID", "Date", "Distance", "Duration", "Description"})
	if err != nil {
		return err
	}

	//Items
	for name, it := range source_emails.Logins {
		ln := []string{name, it.Server, strconv.Itoa(it.Port)}
		err := w.Write(ln)
		if err != nil {
			return err
		}
	}

	w.Flush()
	err = w.Error()
	if err != nil {
		return err
	}

	st.Out_emails = str.String()
	return nil
}
