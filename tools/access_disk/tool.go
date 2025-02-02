package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Search for or change user's and device files.
type access_disk struct {
	Description string //Describe action(read, write, delete). Place or hints where the data could be stored. If writing, mention the value(s).
}

func (st *access_disk) run() string {

	var files []string
	getStructure("disk", &files)
	filesStr := strings.Join(files, "\n")

	SystemPrompt := `You are an AI assistant, who enjoys precision and carefully follows the user's requirements.
	You read, write, delete file or update part(s) of file. You use tools all the time.
	Don't ask to use a tool, just do it! Call tools sequentially. Avoid tool call as parameter value.`

	UserPrompt := ""
	UserPrompt += "This is the list of disk's files(full pathes):\n"
	UserPrompt += filesStr
	UserPrompt += "\n\n"

	UserPrompt += "This is the prompt from user:\n"
	UserPrompt += st.Description
	UserPrompt += "\n\n"
	UserPrompt += "Locate file which you want to work with. If you need, you can look into file with 'read_file' tool.\n"
	UserPrompt += "After you finished, create answer which reflects success or fail. If value was read, return the value and describe place from which value(s) was selected."

	fmt.Println("UserPrompt:", UserPrompt)

	answer := SDK_RunAgent("main", 20, 20000, SystemPrompt, UserPrompt)

	return answer
}

func getStructure(path string, items *[]string) {
	disk, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, it := range disk {
		if it.Name() == "_sky_ini" {
			continue
		}

		if it.IsDir() {
			getStructure(filepath.Join(path, it.Name()), items)
		} else {
			//add
			*items = append(*items, filepath.Join(path, it.Name()))
		}
	}
}
