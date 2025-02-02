package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// Update the tool's code by Prompt.
type update_tool struct {
	Name   string //Tool name.
	Prompt string //How do you wanna change the code. It can be about fixing bug, adding new functionality, change input parameters.
}

func (st *update_tool) run() string {

	toolCode, err := os.ReadFile(fmt.Sprintf("tools/%s/tool.go", st.Name))
	if err != nil {
		log.Fatal(err)
	}

	SystemPrompt := "You are an AI programming assistant, who enjoys precision and carefully follows the user's requirements. You write code in Go-lang."

	UserPrompt := "This is prompt from user:\n"
	UserPrompt += st.Prompt
	UserPrompt += "\n"

	UserPrompt += "Based on this prompt modify this code:\n"
	UserPrompt += fmt.Sprintf("```go\n%s\n```", toolCode)

	UserPrompt += "\n"
	UserPrompt += "If an error occurs, use log.Fatalf. Look at the tool and attributes descriptions and maybe update them.\n"
	UserPrompt += "Don't change header of tool's run() method. Output only code, no explanation. Implement everything, no placeholders. Don't add main() function to the code."

	fmt.Println("update_tool UserPrompt:", UserPrompt)

	code_answer := SDK_RunAgent("programmer", 20, 20000, SystemPrompt, UserPrompt)

	var ok bool
	code_answer, ok = strings.CutPrefix(code_answer, "```go")
	if ok {
		code_answer, ok = strings.CutSuffix(code_answer, "```")
		if ok {
			compile_answer := SDK_SetToolCode(st.Name, code_answer)
			if compile_answer == "" {
				return "success"
			} else {
				//pass error back
			}
		}
	}

	return "failed"
}
