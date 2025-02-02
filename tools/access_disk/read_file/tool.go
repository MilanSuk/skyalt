package main

import (
	"log"
	"os"
)

// Read file from disk. Return string with file content.
type read_file struct {
	Path string // Full path to the file
}

func (st *read_file) run() string {
	data, err := os.ReadFile(st.Path)
	if err != nil {
		log.Fatal(err)
	}
	return string(data)
}
