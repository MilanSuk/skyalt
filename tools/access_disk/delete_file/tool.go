package main

import (
	"log"
)

// Delete file/directory from disk.
type delete_file struct {
	Path string // Full path to the file
}

func (st *delete_file) run() string {
	err := _os_Remove(st.Path)
	if err != nil {
		log.Fatal(err)
	}
	return "success"
}
