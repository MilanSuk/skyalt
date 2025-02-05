package main

import (
	"fmt"
	"strings"
)

// Create new tool from description.
type create_new_tool struct {
	Name        string //Tool name. No spaces(use '_' instead) or special characters.
	Description string //Prompt with the name of tool, parameters(name, type, description) and detail description of functionality.
}

func (st *create_new_tool) run() string {
	SystemPrompt := "You are an AI programming assistant, who enjoys precision and carefully follows the user's requirements. You write code in Go-lang."

	UserPrompt := ""

	UserPrompt += "These are the APIs:\n"
	UserPrompt += "//When you login to any service this function converts password_id into password.\n"
	UserPrompt += "func SDK_GetPassword(id string) (string, error)	//accepts password_id or api_key_id and returns plain password as string.\n\n"
	UserPrompt += "\n"

	UserPrompt += "This is the file(code) template:"
	UserPrompt += "```go\n"
	UserPrompt += fmt.Sprintf(`package main
//<tool_description>
type %s struct {
	<tool_input_parameters_with_descriptions_as_comments>
}
func (st *%s) run() <tool_return_type> {
	<tool_implementation>	//If there is error, use log.Fatalf
}`, st.Name, st.Name)
	UserPrompt += "```"
	UserPrompt += "\n\n"

	UserPrompt += "This is the prompt from user:\n"
	UserPrompt += st.Description
	UserPrompt += "\n\n"

	UserPrompt += "Based on the user's prompt modify the file template."
	UserPrompt += "\n"
	UserPrompt += "<tool_return_type> must be only one. If you are not sure what it should be, use string and return \"success\"."
	UserPrompt += "\n"
	UserPrompt += "You can add more input attributes(more than what is mention in the user prompt). It's very important that the code don't have any placeholders or constants which should programmer changed later(example.com, etc.). Write production ready code only!"
	UserPrompt += "\n"
	UserPrompt += "Never use plain password and api_key as tool_input_parameter, always user password_id or api_key_id and convert them with SDK_GetPassword."
	UserPrompt += "\n"
	UserPrompt += "If an error occurs, use log.Fatalf. Output only modified template above. Implement everything, no placeholders! Don't add main() function to the code."

	fmt.Println("create_new_tool UserPrompt:", UserPrompt)

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
				return compile_answer //error
			}
		}
	}

	return "failed"
}
