package main

import (
	"log"
)

// Write file into disk.
type write_file struct {
	Path string // Full path to the file
	Data string // File data
}

func (st *write_file) run() string {
	err := _os_WriteFile(st.Path, []byte(st.Data), 0644)
	if err != nil {
		log.Fatal(err)
	}
	return "success"
}
