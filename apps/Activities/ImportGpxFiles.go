package main

import (
	"os"
	"path/filepath"
	"strings"
)

type ImportGpxFiles struct {
	Paths           []string // list of file or folder paths to import from
	Out_ImportedIDs []string // list of IDs of successfully imported activities
}

func (st *ImportGpxFiles) run(caller *ToolCaller, ui *UI) error {
	var imported []string
	for _, path := range st.Paths {
		info, err := os.Stat(path)
		if err != nil {
			continue // skip invalid paths
		}
		if info.IsDir() {
			files, err := os.ReadDir(path)
			if err != nil {
				continue
			}
			for _, f := range files {
				if !f.IsDir() && strings.HasSuffix(strings.ToLower(f.Name()), ".gpx") {
					fullPath := filepath.Join(path, f.Name())
					id, err := ImportGpx(fullPath)
					if err == nil {
						imported = append(imported, id)
					}
				}
			}
		} else {
			if strings.HasSuffix(strings.ToLower(path), ".gpx") {
				id, err := ImportGpx(path)
				if err == nil {
					imported = append(imported, id)
				}
			}
		}
	}
	st.Out_ImportedIDs = imported
	return nil
}
