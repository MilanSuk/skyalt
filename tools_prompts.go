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

type ToolsPrompt struct {
	Prompt string //LLM input

	//LLM output
	Message   string
	Reasoning string

	Code  string
	Model string

	//from code
	Name   string
	Schema *ToolsOpenAI_completion_tool
	Errors []ToolsCodeError

	Usage LLMMsgUsage

	header_line int
}

func (prompt *ToolsPrompt) updateSchema() error {
	if prompt.Name == "Storage" {
		return nil
	}

	schema, err := BuildToolsOpenAI_completion_tool(prompt.Name, prompt.Name+".go", prompt.Code)
	if err != nil {
		return err
	}

	prompt.Schema = schema
	return nil
}

func (prompt *ToolsPrompt) setMessage(final_msg string, reasoning_msg string, usage *LLMMsgUsage, model string) {

	re := regexp.MustCompile("(?s)```(?:go|golang)\n(.*?)\n```")
	matches := re.FindAllStringSubmatch(final_msg, -1)

	var goCode strings.Builder
	for _, match := range matches {
		if len(match) > 1 {
			goCode.WriteString(match[1])
			goCode.WriteString("\n")
		}
	}

	prompt.Code = strings.TrimSpace(goCode.String())
	prompt.Message = final_msg
	prompt.Reasoning = reasoning_msg
	prompt.Usage = *usage
	prompt.Model = model

	/*if loadName {
		prompt.Name, _ = _ToolsPrompt_getFileName(prompt.Code)
	}*/
}

type ToolsPrompts struct {
	PromptsFileTime int64

	Prompts  []*ToolsPrompt
	Err      string
	Err_line int

	Generating_name string
	Generating_msg  string
}

func NewToolsPrompts() *ToolsPrompts {
	app := &ToolsPrompts{}
	return app
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

func (app *ToolsPrompts) Generate(appName string, router *ToolsRouter) error {

	defer func() {
		//reset
		app.Generating_name = ""
		app.Generating_msg = ""
	}()

	//find structure
	storagePrompt := app.FindPromptName("Storage")

	var comp LLMComplete
	comp.Temperature = 0.2
	comp.Max_tokens = 32768 //65536
	comp.Top_p = 0.95       //1.0
	comp.Frequency_penalty = 0
	comp.Presence_penalty = 0
	comp.Reasoning_effort = ""
	comp.Max_iteration = 1
	comp.Model = "gpt-4.1-nano"

	msg := router.AddRecompileMsg(appName)
	defer msg.Done()

	//generate code
	for _, prompt := range app.Prompts {

		var err error
		if prompt == storagePrompt {
			comp.SystemMessage, comp.UserMessage, err = app._getStructureMsg(storagePrompt)
			if err != nil {
				return err
			}
		} else {
			comp.SystemMessage, comp.UserMessage, err = app._getToolMsg(prompt, storagePrompt)
			if err != nil {
				return err
			}
		}

		comp.delta = func(msg *ChatMsg) {
			app.Generating_name = prompt.Name
			if msg.Content.Calls != nil {
				app.Generating_msg = msg.Content.Calls.Content
			}
		}

		err = router.llms.Complete(&comp, msg)
		if err != nil {
			return err
		}

		prompt.setMessage(comp.Out_last_final_message, comp.Out_last_reasoning_message, &comp.Out_usage, comp.Model)
	}

	return nil
}

func (app *ToolsPrompts) WriteFiles(folderPath string) error {

	//remove all .go files
	{
		files, err := os.ReadDir(folderPath)
		if err != nil {
			return err
		}
		for _, info := range files {
			name := strings.TrimRight(info.Name(), filepath.Ext(info.Name()))

			if info.IsDir() || filepath.Ext(info.Name()) != ".go" || info.Name() == "main.go" || app.FindPromptName(name) != nil {
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

		path := filepath.Join(folderPath, prompt.Name+".go")
		oldFl, _ := os.ReadFile(path)
		if string(oldFl) != prompt.Code { //note: command goimports may edited the code :(
			err := os.WriteFile(path, []byte(prompt.Code), 0644)
			if err != nil {
				return err
			}
		}
	}

	for _, prompt := range app.Prompts {
		prompt.updateSchema()
	}

	return nil
}

func (app *ToolsPrompts) _getStructureMsg(structPrompt *ToolsPrompt) (string, string, error) {

	apisFile := `func ReadJSONFile[T any](path string, defaultValues *T) (*T, error)
`
	storageFile := `
package main

type ExampleStruct struct {
	//<attributes>
}

func LoadExampleStruct() (*ExampleStruct, error) {
	st := &ExampleStruct{}	//set default values here

	return ReadJSONFile("ExampleStruct.json", st)
}

//<structures functions here>
`

	systemMessage := "You are a programmer. You write code in Go language. Here is the list of files in project folder.\n"

	systemMessage += "file - apis.go:\n```go" + apisFile + "```\n"
	systemMessage += "file - storage.go:\n```go" + storageFile + "```\n"

	systemMessage += "Based on user message, rewrite storage.go file. Your job is to design structures. Write additional functions only if user ask for them. You may write multiple structures.\n"

	systemMessage += "Structures can not have pointers, because they will be saved as JSON, so instead of pointer(s) use ID which is saved in map[interger or string ID].\n"

	systemMessage += "Do not call os.ReadFile() + json.Unmarshal(), instead call ReadJSONFile(). Do not call os.WriteFile(), saving structures into disk is automatic."

	//maybe add old file structures, because it's needed that struct and attributes names are same ...............

	userMessage := structPrompt.Prompt

	return systemMessage, userMessage, nil
}

func (app *ToolsPrompts) _getAPIFile() (string, error) {
	fl, err := os.ReadFile("sdk/main.go")
	if err != nil {
		return "", err
	}
	mainGo := string(fl)

	structsStarts := strings.Index(mainGo, "//--- Ui ---")
	structsEnd := structsStarts + strings.Index(mainGo[structsStarts:], "\nfunc")

	code := string(mainGo[structsStarts+len("//--- Ui ---") : structsEnd]) //add structs

	lines := strings.Split(mainGo[structsEnd:], "\n")
	for _, ln := range lines {
		if strings.HasPrefix(ln, "func (") {
			code += ln + "\n"
		}
	}

	code = strings.ReplaceAll(code, "`json:\",omitempty\"`", "")
	code = strings.Trim(code, "\n ")

	//fmt.Println(code)

	return code, nil
}

func (app *ToolsPrompts) _getToolMsg(prompt *ToolsPrompt, structPrompt *ToolsPrompt) (string, string, error) {
	apisFile, err := app._getAPIFile()
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

	/*toolsNames := ""
	for i, it := range app.Prompts {
		if it.Name == "" || it.Name == prompt.Name {
			continue
		}
		toolsNames += it.Name
		if i+1 < len(app.Prompts) {
			toolsNames += ", "
		} else {
			toolsNames += "."
		}
	}*/

	systemMessage := "You are a programmer. You write code in Go language. Here is the list of files in project folder.\n"

	systemMessage += "file - apis.go:\n```go" + apisFile + "```\n"
	systemMessage += "file - storage.go:\n```go" + storageFile + "```\n"
	systemMessage += "file - tool.go:\n```go" + toolFile + "```\n"

	systemMessage += "Based on user message, rewrite tool.go file. Your job is to design function(tool).\n"

	/*systemMessage += "Rename 'ExampleTool' with tools name based on user prompt. Don't use the word 'tool' in the name."
	if toolsNames != "" {
		systemMessage += "Here are the names of tools which already exists, don't use them:" + toolsNames + "\n"
	}*/

	systemMessage += "Figure out <tool's arguments> based on user prompt. They are two types of arguments - inputs and outputs. Output arguments must start with 'Out_', Input arguments don't have any prefix. All arguments must start with upper letter. Every argument must have description as comment.\n"

	systemMessage += "If you need to access the storage, call function Load...() from storage.go.\n"

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
