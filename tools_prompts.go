/*
Copyright 2025 Milan Suk

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this db except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type ToolsPromptMessages struct {
	Message   string
	Reasoning string
}
type ToolsPrompt struct {
	Prompt string //LLM input

	//LLM output
	Messages []ToolsPromptMessages

	Code string

	//from code
	Name   string
	Schema *ToolsOpenAI_completion_tool
	Errors []ToolsCodeError

	Usage LLMMsgUsage

	header_line int

	previousMessages []byte
}

func (prompt *ToolsPrompt) updateSchema() error {
	if prompt.Name == "Storage" || prompt.Code == "" {
		return nil
	}

	schema, err := BuildToolsOpenAI_completion_tool(prompt.Name, prompt.Name+".go", prompt.Code)
	if err != nil {
		return err
	}

	prompt.Schema = schema
	return nil
}

func (prompt *ToolsPrompt) setMessage(final_msg string, reasoning_msg string, usage *LLMMsgUsage, previousMessages []byte) {

	re := regexp.MustCompile("(?s)```(?:go|golang)(.*?)```")
	matches := re.FindAllStringSubmatch(final_msg, -1)

	var goCode strings.Builder
	for _, match := range matches {
		if len(match) > 1 {
			goCode.WriteString(match[1])
			goCode.WriteString("\n")
		}
	}

	if goCode.Len() > 0 {
		prompt.Code = strings.TrimSpace(goCode.String())
	} else {
		prompt.Code = final_msg
	}
	prompt.Messages = append(prompt.Messages, ToolsPromptMessages{Message: final_msg, Reasoning: reasoning_msg})
	prompt.Usage = *usage

	prompt.previousMessages = previousMessages

	/*if loadName {
		prompt.Name, _ = _ToolsPrompt_getFileName(prompt.Code)
	}*/
}

type ToolsPrompts struct {
	PromptsFileTime int64
	SecretsFileTime int64

	Prompts  []*ToolsPrompt
	Err      string
	Err_line int

	Generating_name string
	Generating_msg  string
}

func NewToolsPrompts(folderPath string) (*ToolsPrompts, error) {
	app := &ToolsPrompts{}
	return app, nil
}

func (app *ToolsPrompts) Destroy() error {
	return nil
}

func (app *ToolsPrompts) SetCodeErrors(errs []ToolsCodeError) {

	//reset
	for _, prompt := range app.Prompts {
		prompt.Errors = nil
	}

	//add
	for _, er := range errs {
		file_name := strings.TrimRight(filepath.Base(er.File), filepath.Ext(er.File))

		found := false
		for _, prompt := range app.Prompts {
			if prompt.Name == file_name {
				prompt.Errors = append(prompt.Errors, er)
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("Code file '%s' not found\n", er.File)
		}
	}
}

func (app *ToolsPrompts) FindPromptName(name string) *ToolsPrompt {
	for _, prompt := range app.Prompts {
		if prompt.Name == name {
			return prompt
		}
	}
	return nil
}

func (app *ToolsPrompts) Reload(folderPath string) (bool, error) {

	app.Prompts = nil
	app.Err = ""
	app.Err_line = 0

	promptsFilePath := filepath.Join(folderPath, "skyalt")
	fl, err := os.ReadFile(promptsFilePath)
	if err != nil {
		return false, err
	}

	saveFile := false
	structFound := false
	var last_prompt *ToolsPrompt
	lines := strings.Split(string(fl), "\n")
	for i, ln := range lines {
		ln = strings.TrimSpace(ln)

		isHash := strings.HasPrefix(strings.ToLower(ln), "#")
		isStorage := strings.HasPrefix(strings.ToLower(ln), "#storage")

		if isStorage && structFound {
			app.Err = "second '#storage' is not allowed"
			app.Err_line = i + 1
			return false, fmt.Errorf(app.Err)
		}

		if isHash {

			//extract or edit name
			toolName := "Storage"
			if !isStorage {
				toolName = ln[1:] //skip '#'
				newToolName := _ToolsPrompt_getValidFileName(toolName)

				if toolName != newToolName {
					toolName = newToolName
					ln = "#" + newToolName
					lines[i] = ln

					saveFile = true
				}
			}

			if toolName == "" {
				app.Err = "nothing after '#'"
				app.Err_line = i + 1
				return false, fmt.Errorf(app.Err)
			}

			//save
			if last_prompt != nil {
				app.Prompts = append(app.Prompts, last_prompt)
			}
			last_prompt = &ToolsPrompt{Name: toolName, header_line: i}
			if isStorage {
				structFound = true
			}
		} else {
			if last_prompt == nil && ln != "" {
				app.Err = "missing '#storage' or '#tool' header"
				app.Err_line = i + 1
				return false, fmt.Errorf(app.Err)
			}

			if last_prompt != nil {
				last_prompt.Prompt += ln + "\n" //add line
			}
		}
	}

	//add last
	if last_prompt != nil {
		app.Prompts = append(app.Prompts, last_prompt)
	}

	//clear
	for _, prompt := range app.Prompts {
		prompt.Prompt = strings.Trim(prompt.Prompt, "\n ")
	}

	if saveFile {
		str := strings.Join(lines, "\n")
		err := os.WriteFile(promptsFilePath, []byte(str), 0644)
		if err != nil {
			return false, err
		}
	}

	return saveFile, nil
}

func (app *ToolsPrompts) generatePromptCode(prompt *ToolsPrompt, storagePrompt *ToolsPrompt, comp *LLMComplete, msg *ToolsRouterMsg, llms *LLMs) error {

	var err error
	if prompt == storagePrompt {
		comp.SystemMessage, comp.UserMessage, err = app._getStorageMsg(storagePrompt)
		if err != nil {
			return err
		}
	} else {
		comp.SystemMessage, comp.UserMessage, err = app._getToolMsg(prompt, storagePrompt)
		if err != nil {
			return err
		}
	}

	if len(prompt.Errors) > 0 {
		comp.PreviousMessages = prompt.previousMessages

		//add list of errors
		lines := strings.Split(prompt.Code, "\n")
		for _, er := range prompt.Errors {
			lines[er.Line] = fmt.Sprintf("%s\t//Error(Col %d): %s", lines[er.Line], er.Col, er.Msg)
		}
		code := strings.Join(lines, "\n")

		comp.UserMessage = "```go" + code + "```\n"
		comp.UserMessage += "Above code has compiler error(s), marked in line comments(//Error). Please fix them by rewriting above code. Also remove comments with errors."
	}

	comp.delta = func(msg *ChatMsg) {
		app.Generating_name = prompt.Name
		if msg.Content.Calls != nil {
			app.Generating_msg = msg.Content.Calls.Content
		}
	}

	err = llms.Complete(comp, msg, "code")
	if err != nil {
		return err
	}

	prompt.setMessage(comp.Out_answer, comp.Out_reasoning, &comp.Out_usage, comp.Out_messages)

	return nil
}

func _ToolsPrompts_prepareLLMCompleteStruct() *LLMComplete {
	comp := &LLMComplete{}
	comp.Temperature = 0.2
	comp.Max_tokens = 32768 //65536
	comp.Top_p = 0.95       //1.0
	comp.Frequency_penalty = 0
	comp.Presence_penalty = 0
	comp.Reasoning_effort = ""
	comp.Max_iteration = 1
	//comp.Model = "gpt-4.1-mini"

	return comp
}

func (app *ToolsPrompts) GenerateStructureCode(msg *ToolsRouterMsg, llms *LLMs) error {

	defer func() {
		//reset
		app.Generating_name = ""
		app.Generating_msg = ""
	}()

	//find Storage
	storagePrompt := app.FindPromptName("Storage")
	if storagePrompt == nil {
		return fmt.Errorf("'Storage' prompt not found")
	}

	return app.generatePromptCode(storagePrompt, storagePrompt, _ToolsPrompts_prepareLLMCompleteStruct(), msg, llms)
}

func (app *ToolsPrompts) GenerateToolsCode(msg *ToolsRouterMsg, llms *LLMs) error {

	defer func() {
		//reset
		app.Generating_name = ""
		app.Generating_msg = ""
	}()

	//find Storage
	storagePrompt := app.FindPromptName("Storage")
	if storagePrompt == nil {
		return fmt.Errorf("'Storage' prompt not found")
	}

	//then generate tools code
	for _, prompt := range app.Prompts {
		if prompt == storagePrompt || prompt.Prompt == "" {
			continue
		}

		err := app.generatePromptCode(prompt, storagePrompt, _ToolsPrompts_prepareLLMCompleteStruct(), msg, llms)
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *ToolsPrompts) WriteFiles(folderPath string, secrets *ToolsSecrets) error {

	//remove all .go files
	{
		files, err := os.ReadDir(folderPath)
		if err != nil {
			return err
		}
		for _, info := range files {
			//name := strings.TrimRight(info.Name(), filepath.Ext(info.Name()))

			if info.IsDir() || filepath.Ext(info.Name()) != ".go" || info.Name() == "main.go" {
				continue
			}
			os.Remove(filepath.Join(folderPath, info.Name()))
		}
	}

	//write code into files
	for _, prompt := range app.Prompts {
		if prompt.Name == "" || prompt.Code == "" {
			continue
		}

		new_code := secrets.ReplaceAliases(prompt.Code)

		path := filepath.Join(folderPath, prompt.Name+".go")
		old_code, _ := os.ReadFile(path)
		if string(old_code) != new_code { //note: command goimports may edited the code :(
			err := os.WriteFile(path, []byte(new_code), 0644)
			if err != nil {
				return err
			}
		}
	}

	for _, prompt := range app.Prompts {
		err := prompt.updateSchema()
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *ToolsPrompts) _getStorageMsg(structPrompt *ToolsPrompt) (string, string, error) {

	apisFile, err := os.ReadFile("sdk/_api_storage.go")
	if err != nil {
		return "", "", err
	}

	exampleFile, err := os.ReadFile("sdk/_example_storage.go")
	if err != nil {
		return "", "", err
	}

	systemMessage := "You are a programmer. You write code in Go language. Here is the list of files in project folder.\n"

	systemMessage += "file - apis.go:\n```go" + string(apisFile) + "```\n"
	systemMessage += "file - storage.go:\n```go" + string(exampleFile) + "```\n"

	systemMessage += "Based on user message, rewrite storage.go file. Your job is to design structures. Write additional functions only if user ask for them. You may write multiple structures.\n"

	systemMessage += "Structure attributes can not have pointers, because they will be saved as JSON, so instead of pointer(s) use ID which is saved in map[interger or string ID].\n"

	systemMessage += "Load<name_of_struct>() functions always returns pointer, not array."

	systemMessage += "Do not call os.ReadFile() + json.Unmarshal(), instead call ReadJSONFile(). Do not call os.WriteFile(), saving data in structures into disk is automatic."

	systemMessage += "Never define constants('const') use variables('var') for everything.\n"

	//maybe add old file structures, because it's needed that struct and attributes names are same ....

	userMessage := structPrompt.Prompt

	return systemMessage, userMessage, nil
}

func (app *ToolsPrompts) _getToolMsg(prompt *ToolsPrompt, structPrompt *ToolsPrompt) (string, string, error) {

	apisFile, err := os.ReadFile("sdk/_api_tool.go")
	if err != nil {
		return "", "", err
	}

	exampleFile, err := os.ReadFile("sdk/_example_tool.go")
	if err != nil {
		return "", "", err
	}

	storageFile := structPrompt.Code

	toolFile := fmt.Sprintf(`
package main

type %s struct {
	//<tool's arguments>
}

func (st *%s) run(caller *ToolCaller, ui *UI) error {

	//<code based on prompt>

	return nil
}
`, prompt.Name, prompt.Name)

	systemMessage := "You are a programmer. You write code in Go language. Here is the list of files in project folder.\n"

	systemMessage += "file - apis.go:\n```go" + string(apisFile) + "```\n"
	systemMessage += "file - storage.go:\n```go" + storageFile + "```\n"
	systemMessage += "file - example.go:\n```go" + string(exampleFile) + "```\n"
	systemMessage += "file - tool.go:\n```go" + toolFile + "```\n"

	systemMessage += "Based on user message, rewrite tool.go file. Your job is to design function(tool). Look into example.go to understand how APIs and storage functions work.\n"

	systemMessage += "Figure out <tool's arguments> based on user prompt. They are two types of arguments - inputs and outputs. Output arguments must start with 'Out_', Input arguments don't have any prefix. All arguments must start with upper letter. Every argument must have description as comment. You can add extra marks(with brackets) at the end of comment:\n"
	systemMessage += "[optional] - caller can ignore attribute"
	systemMessage += `[options: <list of options] - caller must pick up from list of values. Example 1: [options: "first", "second", "third"]. Example 2: [options: 2, 3, 5, 7, 11]`

	systemMessage += "If you need to access the storage, call function Load...() from storage.go which return data. Don't call save/write on that data, it's automaticaly called after function ends.\n"

	systemMessage += "Never define constants('const') use variables('var') for everything.\n"

	//systemMessage += fmt.Sprintf("You may add help functions into tool.go. They should start with ```func (st *%s)NameOfHelpFunction```\n", prompt.Name)

	userMessage := prompt.Prompt

	return systemMessage, userMessage, nil
}

func _ToolsPrompt_getValidFileName(s string) string {
	s = strings.TrimSpace(s)

	re := regexp.MustCompile(`[^a-zA-Z0-9_]+`)
	words := re.Split(s, -1)

	// Use a strings.Builder for efficient concatenation
	var builder strings.Builder

	// Process each word
	for i, word := range words {
		if word != "" {
			if len(word) > 0 && word[0] >= 'a' && word[0] <= 'z' {
				// Capitalize the first letter if it's lowercase
				builder.WriteString(strings.ToUpper(word[0:1]) + word[1:])
			} else if i == 0 && len(word) > 0 && word[0] >= '0' && word[0] <= '9' {
				//ignore 1st word starting with digits
			} else {
				// Keep the word as is if it starts with uppercase, or underscore
				builder.WriteString(word)
			}
		}
	}

	return builder.String()
}

/*func _ToolsPrompt_getFileName(code string) (string, error) {
	file, err := parser.ParseFile(token.NewFileSet(), "", code, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("failed to parse source: %v", err)
	}

	var structName string
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Name.Name == "run" {
				if fn.Recv != nil && len(fn.Recv.List) > 0 {
					switch t := fn.Recv.List[0].Type.(type) {
					case *ast.StarExpr: //(*Type)
						if ident, ok := t.X.(*ast.Ident); ok {
							structName = ident.Name
						}
					case *ast.Ident: //(Type)
						structName = t.Name
					}
				}
			}
			return false
		}
		return true
	})

	return structName, nil
}*/
