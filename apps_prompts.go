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
	"slices"
	"strings"
	"sync"
)

type ToolsPromptMessages struct {
	Message   string
	Reasoning string
}

type ToolsPromptTYPE int

const (
	ToolsPrompt_STORAGE ToolsPromptTYPE = iota
	ToolsPrompt_FUNCTION
	ToolsPrompt_TOOL
	ToolsPrompt_START
)

type ToolsPromptCode struct {
	Code   string
	Errors []ToolsCodeError
	Usage  LLMMsgUsage
}

type ToolsPrompt struct {
	Type ToolsPromptTYPE
	Name string

	Prompt string //LLM input

	//LLM output
	Messages []ToolsPromptMessages

	CodeVersions []ToolsPromptCode
	//Code string

	//from code
	Schema *ToolsOpenAI_completion_tool
	//Errors []ToolsCodeError

	//Usage LLMMsgUsage

	previousMessages []byte
}

func NewToolsPrompt(Type ToolsPromptTYPE, Name string) *ToolsPrompt {
	return &ToolsPrompt{Type: Type, Name: Name}
}

func (prompt *ToolsPrompt) IsCodeWithoutErrors() bool {
	return (len(prompt.CodeVersions) > 0 && len(prompt.CodeVersions[len(prompt.CodeVersions)-1].Errors) == 0)
}

func (prompt *ToolsPrompt) GetLastCode() string {
	if len(prompt.CodeVersions) == 0 {
		return ""
	}
	return prompt.CodeVersions[len(prompt.CodeVersions)-1].Code
}

func (prompt *ToolsPrompt) updateSchema() error {
	if prompt.Type != ToolsPrompt_TOOL || len(prompt.CodeVersions) == 0 {
		return nil
	}

	schema, err := BuildToolsOpenAI_completion_tool(prompt.Name, prompt.Name+".go", prompt.GetLastCode())
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

	//add new code version
	{
		code := final_msg
		if goCode.Len() > 0 {
			code = strings.TrimSpace(goCode.String())
		}
		prompt.CodeVersions = append(prompt.CodeVersions, ToolsPromptCode{Code: code, Usage: *usage})
	}

	prompt.Messages = append(prompt.Messages, ToolsPromptMessages{Message: final_msg, Reasoning: reasoning_msg})

	prompt.previousMessages = previousMessages

	/*if loadName {
		prompt.Name, _ = _ToolsPrompt_getFileName(prompt.Code)
	}*/
}

type ToolsPromptGen struct {
	Name    string
	Message string
}

type ToolsPrompts struct {
	Changed bool

	Prompts  []*ToolsPrompt
	Err      string
	Err_line int

	StartPrompt string

	Generating_msg_id string
	Generating_items  []*ToolsPromptGen

	refresh bool

	lock sync.Mutex
}

func (app *ToolsPrompts) Destroy() error {
	return nil
}

func (app *ToolsPrompts) ResetGenMsgs(msg_id string) {
	app.lock.Lock()
	defer app.lock.Unlock()

	app.Generating_msg_id = msg_id
	app.Generating_items = nil
}
func (app *ToolsPrompts) AddGenMsg(name string, msg string) {
	app.lock.Lock()
	defer app.lock.Unlock()

	for _, it := range app.Generating_items {
		if it.Name == name {
			//update
			it.Message = msg
			return
		}
	}
	//add
	app.Generating_items = append(app.Generating_items, &ToolsPromptGen{Name: name, Message: msg})
}
func (app *ToolsPrompts) RemoveGenMsg(name string) {
	app.lock.Lock()
	defer app.lock.Unlock()

	for i, it := range app.Generating_items {
		if it.Name == name {
			app.Generating_items = slices.Delete(app.Generating_items, i, i+1)
			return
		}
	}
}

func (app *ToolsPrompts) SetCodeErrors(errs []ToolsCodeError) {

	//reset
	/*for _, prompt := range app.Prompts {
		prompt.Errors = nil
	}*/

	//add
	for _, er := range errs {
		file_name := strings.TrimRight(filepath.Base(er.File), filepath.Ext(er.File))

		prompt := app.FindPromptName(file_name)
		if prompt != nil {
			if len(prompt.CodeVersions) > 0 {
				prompt.CodeVersions[len(prompt.CodeVersions)-1].Errors = append(prompt.CodeVersions[len(prompt.CodeVersions)-1].Errors, er)
			}
		} else {
			fmt.Printf("Code file '%s' not found\n", er.File)
		}
	}
}

func (app *ToolsPrompts) FindStorage() *ToolsPrompt {
	return app.FindPromptName("Storage")
}

func (app *ToolsPrompts) FindPromptName(name string) *ToolsPrompt {
	for _, prompt := range app.Prompts {
		if prompt.Name == name {
			return prompt
		}
	}
	return nil
}

func (app *ToolsPrompts) _reloadFromCodeFiles(folderPath string) error {

	app.Prompts = nil
	app.Err = ""
	app.Err_line = 0

	files, err := os.ReadDir(folderPath)
	if err != nil {
		return err
	}

	//add new tools
	for _, info := range files {
		if info.IsDir() || filepath.Ext(info.Name()) != ".go" || info.Name() == "main.go" {
			continue
		}

		toolName, _ := strings.CutSuffix(info.Name(), ".go") //remove 'z' and '.go'

		code, err := os.ReadFile(filepath.Join(folderPath, info.Name()))
		if err != nil {
			return err
		}

		tp := ToolsPrompt_TOOL
		if info.Name() == "Storage.go" {
			tp = ToolsPrompt_STORAGE
		}

		item := NewToolsPrompt(tp, toolName)
		item.CodeVersions = append(item.CodeVersions, ToolsPromptCode{Code: string(code)})
		err = item.updateSchema()
		if err != nil {
			return err
		}

		app.Prompts = append(app.Prompts, item)

	}

	//remove deleted tools
	for i := len(app.Prompts) - 1; i >= 0; i-- {
		found := false
		for _, file := range files {
			if !file.IsDir() && file.Name() == app.Prompts[i].Name+".go" {
				found = true
				break
			}
		}
		if !found {
			app.Prompts = slices.Delete(app.Prompts, i, i+1)
		}
	}

	return nil
}

func (app *ToolsPrompts) _reloadFromPromptFile(folderPath string) (bool, error) {

	//reset
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
	startFound := false
	var last_prompt *ToolsPrompt
	lines := strings.Split(string(fl), "\n")
	for i, ln := range lines {
		ln = strings.TrimSpace(ln)

		isHash := strings.HasPrefix(strings.ToLower(ln), "#")
		isStorage := strings.HasPrefix(strings.ToLower(ln), "#storage")
		isFunction := strings.HasPrefix(strings.ToLower(ln), "#function")
		isTool := strings.HasPrefix(strings.ToLower(ln), "#tool")
		isStart := strings.HasPrefix(strings.ToLower(ln), "#start")

		if isStorage && structFound {
			app.Err = "second '#storage' is not allowed"
			app.Err_line = i + 1
			return false, LogsErrorf(app.Err)
		}

		if isStart && startFound {
			app.Err = "second '#start' is not allowed"
			app.Err_line = i + 1
			return false, LogsErrorf(app.Err)
		}

		if isHash {
			//extract or edit name
			var Type ToolsPromptTYPE
			var toolName string
			if isStorage {
				Type = ToolsPrompt_STORAGE
				toolName = "Storage"
			} else if isFunction {
				Type = ToolsPrompt_FUNCTION
			} else if isStart {
				Type = ToolsPrompt_START
				toolName = "Start"
			} else if isTool {
				Type = ToolsPrompt_TOOL
			} else {
				app.Err = "'#' must follow with 'storage', 'function', 'tool' or 'start'"
				app.Err_line = i + 1
				return false, LogsErrorf(app.Err)
			}

			if isFunction || isTool {

				if isFunction {
					toolName = ln[len("#function"):] //skip
				} else {
					toolName = ln[len("#tool"):] //skip
				}

				newToolName := _ToolsPrompt_getValidFileName(toolName)

				if newToolName == "" {
					app.Err = "missing name"
					app.Err_line = i + 1
					return false, LogsErrorf(app.Err)
				}

				if toolName != newToolName {
					toolName = newToolName

					if isFunction {
						ln = "#function " + newToolName
					} else {
						ln = "#tool " + newToolName
					}
					lines[i] = ln

					saveFile = true
				}
			}

			//save
			if last_prompt != nil {
				app.Prompts = append(app.Prompts, last_prompt)
			}
			last_prompt = NewToolsPrompt(Type, toolName)

			if isStorage {
				structFound = true
			}
			if isStart {
				startFound = true
			}
		} else {
			if last_prompt == nil && ln != "" {
				app.Err = "missing '#storage' or '#tool' header"
				app.Err_line = i + 1
				return false, LogsErrorf(app.Err)
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

	//extract start prompt
	for i, prompt := range app.Prompts {
		if prompt.Type == ToolsPrompt_START {
			app.StartPrompt = prompt.Prompt

			app.Prompts = slices.Delete(app.Prompts, i, i+1) //remove
			break
		}
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

func (app *ToolsPrompts) generatePromptCode(prompt *ToolsPrompt, msg *AppsRouterMsg, llms *LLMs) error {
	comp := NewLLMCompletion()

	var err error
	switch prompt.Type {
	case ToolsPrompt_STORAGE:
		comp.SystemMessage, comp.UserMessage, err = app._getStorageMsg(prompt)
		if err != nil {
			return err
		}
	case ToolsPrompt_FUNCTION:
		comp.SystemMessage, comp.UserMessage, err = app._getFunctionMsg(prompt)
		if err != nil {
			return err
		}
	case ToolsPrompt_TOOL:
		comp.SystemMessage, comp.UserMessage, err = app._getToolMsg(prompt)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("prompt '%s:%s' is unknown type", prompt.Type, prompt.Name)

	}

	//error(s)
	if len(prompt.CodeVersions) > 0 {
		last_code := prompt.CodeVersions[len(prompt.CodeVersions)-1]
		if len(last_code.Errors) > 0 {

			comp.PreviousMessages = prompt.previousMessages

			//add list of errors
			lines := strings.Split(last_code.Code, "\n")
			for _, er := range last_code.Errors {
				ln := er.Line - 1
				if ln >= 0 && ln < len(lines) {
					lines[ln] = fmt.Sprintf("%s\t//Error(Col %d): %s", lines[ln], er.Col, er.Msg)
				}
			}
			code := strings.Join(lines, "\n")
			comp.UserMessage = "```go" + code + "```\n"
			comp.UserMessage += "Above code has compiler error(s), marked in line comments(//Error). Please fix them by rewriting above code. Also remove comments with errors."
		}
	}

	defer app.RemoveGenMsg(prompt.Name)
	comp.delta = func(msg *ChatMsg) {
		msgStr := ""
		if msg.Content.Calls != nil {
			msgStr = msg.Content.Calls.Content
		}
		app.AddGenMsg(prompt.Name, msgStr)
	}

	err = llms.Complete(comp, msg, "code")
	if err != nil {
		return err
	}

	prompt.setMessage(comp.Out_answer, comp.Out_reasoning, &comp.Out_usage, comp.Out_messages)

	return nil
}

func (app *ToolsPrompts) RemoveOldCodeFiles(folderPath string) error {

	files, err := os.ReadDir(folderPath)
	if err != nil {
		return err
	}

	//remove all .go files
	for _, info := range files {
		//name := strings.TrimRight(info.Name(), filepath.Ext(info.Name()))

		if info.IsDir() || filepath.Ext(info.Name()) != ".go" || info.Name() == "main.go" {
			continue
		}
		os.Remove(filepath.Join(folderPath, info.Name()))
	}
	return nil
}

func (app *ToolsPrompts) WriteFiles(folderPath string, secrets *ToolsSecrets) error {

	//write code into files
	for _, prompt := range app.Prompts {
		if prompt.Name == "" || len(prompt.CodeVersions) == 0 {
			continue
		}
		new_code := secrets.ReplaceAliases(prompt.GetLastCode())

		path := filepath.Join(folderPath, prompt.Name+".go")
		old_code, _ := os.ReadFile(path)
		if string(old_code) != new_code { //note: command goimports may edited the code :(
			err := os.WriteFile(path, []byte(new_code), 0644)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (app *ToolsPrompts) UpdateSchemas() error {
	for _, prompt := range app.Prompts {
		err := prompt.updateSchema()
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *ToolsPrompts) getFunctionsHeadersCode() string {
	code := ""
	for _, prompt := range app.Prompts {
		if prompt.Type != ToolsPrompt_FUNCTION {
			continue
		}

		last_code := prompt.GetLastCode()
		d := strings.Index(last_code, "func "+prompt.Name)
		if d >= 0 {
			header := last_code[d:]
			dd := strings.IndexByte(header, '{')
			if dd >= 0 {
				header = header[:dd]

				if strings.IndexByte(prompt.Prompt, '\n') >= 0 {
					code += "/*" + prompt.Prompt + "*/" //multi-lined
				} else {
					code += "//" + prompt.Prompt //single line
				}

				code += "\n"
				code += strings.TrimSpace(header)
				code += "\n\n"
			}

		}
	}

	return code
}

func (app *ToolsPrompts) _getStorageMsg(storagePrompt *ToolsPrompt) (string, string, error) {

	apisFile, err := os.ReadFile("sdk/_api_storage.go")
	if err != nil {
		return "", "", err
	}

	exampleFile, err := os.ReadFile("sdk/_example_storage.go")
	if err != nil {
		return "", "", err
	}

	systemMessage := "You are a programmer. You write code in the Go language. You write production code - avoid placeholders or implement later type of comments. Here is the list of files in the project folder.\n"

	systemMessage += "file - apis.go:\n```go" + string(apisFile) + "```\n"
	systemMessage += "file - storage.go:\n```go" + string(exampleFile) + "```\n"

	systemMessage += "Based on the user message, rewrite the storage.go file. Your job is to design structures. Write additional functions only if the user asks for them. You may write multiple structures.\n"

	systemMessage += "Structure attributes can not have pointers, because they will be saved as JSON, so instead of pointers, use ID, which is saved in a map[integer or string ID].\n"

	systemMessage += "Load<name_of_struct>() functions always returns pointer, not array."

	systemMessage += "Do not call os.ReadFile() + json.Unmarshal(), instead call ReadJSONFile(). Do not call os.WriteFile(), saving data in structures into disk is automatic."

	systemMessage += "Never define constants('const'), use variables('var') for everything.\n"

	//maybe add old file structures, because it's needed that struct and attributes names are same ....

	userMessage := storagePrompt.Prompt

	return systemMessage, userMessage, nil
}

func (app *ToolsPrompts) _getFunctionMsg(functionPrompt *ToolsPrompt) (string, string, error) {
	storagePrompt := app.FindStorage()
	if storagePrompt == nil {
		return "", "", LogsErrorf("'Storage' prompt not found")
	}

	funcFile := fmt.Sprintf(`
package main

func %s(/*arguments*/) /*return types*/ {
	//<code based on prompt>
}
`, functionPrompt.Name)

	systemMessage := "You are a programmer. You write code in the Go language. You write production code - avoid placeholders or implement later type of comments. Here is the list of files in the project folder.\n"

	systemMessage += "file - storage.go:\n```go" + storagePrompt.GetLastCode() + "```\n"
	systemMessage += "file - " + functionPrompt.Name + ".go:\n```go" + string(funcFile) + "```\n"

	systemMessage += "Based on the user message, rewrite the " + functionPrompt.Name + "() function inside " + functionPrompt.Name + ".go file.\n"

	systemMessage += "Figure it out function argument(s), return types(s) and function body based on user message.\n"

	systemMessage += "Load<name_of_struct>() functions always returns pointer, not array."

	systemMessage += "Do not call os.ReadFile() + json.Unmarshal(), instead call ReadJSONFile(). Do not call os.WriteFile(), saving data in structures into disk is automatic."

	systemMessage += "Never define constants('const'), use variables('var') for everything.\n"

	userMessage := functionPrompt.Prompt

	return systemMessage, userMessage, nil
}

func (app *ToolsPrompts) _getToolMsg(prompt *ToolsPrompt) (string, string, error) {
	storagePrompt := app.FindStorage()
	if storagePrompt == nil {
		return "", "", LogsErrorf("'Storage' prompt not found")
	}

	apisFile, err := os.ReadFile("sdk/_api_tool.go")
	if err != nil {
		return "", "", err
	}

	exampleFile, err := os.ReadFile("sdk/_example_tool.go")
	if err != nil {
		return "", "", err
	}

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

	systemMessage := "You are a programmer. You write code in the Go language. You write production code - avoid placeholders or implement later type of comments. Here is the list of files in the project folder.\n"

	systemMessage += "file - apis.go:\n```go" + string(apisFile) + "```\n"
	systemMessage += "file - storage.go:\n```go" + storagePrompt.GetLastCode() + "\n" + app.getFunctionsHeadersCode() + "```\n"
	systemMessage += "file - example.go:\n```go" + string(exampleFile) + "```\n"
	systemMessage += "file - tool.go:\n```go" + toolFile + "```\n"

	systemMessage += "Based on the user message, rewrite the tool.go file. Your job is to design a function(tool). Look into an example.go to understand how APIs and storage functions work.\n"

	systemMessage += "Figure out <tool's arguments> based on the user prompt. There are two types of arguments - inputs and outputs. Output arguments must start with 'Out_', Input arguments don't have any prefix. All arguments must start with an upper-case letter. Every argument must have a description as a comment. You can add extra marks(with brackets []) at the end of a comment. You may add multiple marks with your pair of brackets. Here are the marks:\n"
	systemMessage += "[optional] - caller can ignore the attribute\n"
	systemMessage += `[options: <list of options>] - caller must pick up from the list of values. Example 1: [options: "first", "second", "third"]. Example 2: [options: 2, 3, 5, 7, 11]\n`
	systemMessage += "\n"
	systemMessage += "Storage.go has list of functions, use them.\n"
	systemMessage += "To access the storage, call the Load...() function in storage.go, which returns the data. Don't call save/write on that data, it's automatically called after the function ends.\n"
	systemMessage += "Never define constants('const'), use variables('var') for everything.\n"

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
	if LogsError(err) != nil {
		return "", err
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
