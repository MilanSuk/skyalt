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
	"go/ast"
	"go/parser"
	"go/token"
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

	Code string

	//from code
	Name   string
	Schema *ToolsOpenAI_completion_tool
	Errors []ToolsCodeError

	Usage LLMMsgUsage

	header_line int
}

func (prompt *ToolsPrompt) updateSchema() error {
	schema, err := BuildToolsOpenAI_completion_tool(prompt.Name, prompt.Name+".go", prompt.Code)
	if err != nil {
		return err
	}

	prompt.Schema = schema
	return nil
}

func (prompt *ToolsPrompt) GetErrors(file string) (errs []ToolsCodeError) {
	for _, er := range prompt.Errors {
		if er.File == file {
			errs = append(errs, er)
		}
	}
	return
}

func (prompt *ToolsPrompt) setMessage(final_msg string, reasoning_msg string, usage *LLMMsgUsage, loadName bool) {

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

	if loadName {
		prompt.Name, _ = _ToolsPrompt_getFileName(prompt.Code)
	}

}

type ToolsPrompts struct {
	PromptsFileTime int64

	Prompts  []*ToolsPrompt
	Err      string
	Err_line int
}

func NewToolsPrompts() *ToolsPrompts {
	app := &ToolsPrompts{}
	return app
}

func (app *ToolsPrompts) Destroy() error {
	return nil
}

func (app *ToolsPrompts) SetCodeErrors(errs []ToolsCodeError) {
	for _, er := range errs {
		file_name := strings.TrimRight(er.File, filepath.Ext(er.File))

		for _, prompt := range app.Prompts {
			if prompt.Name == file_name {
				prompt.Errors = append(prompt.Errors, er)
				break
			}
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

func (app *ToolsPrompts) Reload(folderPath string) error {

	app.Prompts = nil
	app.Err = ""
	app.Err_line = 0

	promptsFilePath := filepath.Join(folderPath, "skyalt")
	fl, err := os.ReadFile(promptsFilePath)
	if err != nil {
		return err
	}

	structFound := false
	var last_prompt *ToolsPrompt
	lines := strings.Split(string(fl), "\n")
	for i, ln := range lines {
		ln = strings.TrimSpace(ln)

		isStruct := strings.HasPrefix(strings.ToLower(ln), "#storage")
		isTool := strings.HasPrefix(strings.ToLower(ln), "#tool")

		if isStruct && structFound {
			app.Err = "second '#storage' is not allowed"
			app.Err_line = i + 1
			break
		}

		if isStruct || isTool {
			//save
			if last_prompt != nil {
				app.Prompts = append(app.Prompts, last_prompt)
			}
			if isStruct {
				last_prompt = &ToolsPrompt{Name: "Structures", header_line: i}
				structFound = true
			} else {
				last_prompt = &ToolsPrompt{header_line: i}
			}
		} else {

			if last_prompt == nil && ln != "" {
				app.Err = "missing '#storage' or '#tool' header"
				app.Err_line = i + 1
				break
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

	return nil
}

func (app *ToolsPrompts) Generate(folderPath string, router *ToolsRouter) error {

	//find structure
	structPrompt := app.FindPromptName("Structures")

	var comp LLMComplete
	comp.Model = "grok-3-mini" //grok-3-mini-fast
	comp.Temperature = 0.2
	comp.Max_tokens = 65536
	comp.Top_p = 0.7 //1.0
	comp.Frequency_penalty = 0
	comp.Presence_penalty = 0
	comp.Reasoning_effort = ""
	comp.Max_iteration = 1

	msg := router.AddRecompileMsg(folderPath)
	defer msg.Done()

	//generate Structures.go
	if structPrompt != nil {
		var err error
		comp.SystemMessage, comp.UserMessage, err = app._getStructureMsg(structPrompt)
		if err != nil {
			return err
		}

		err = router.llms.Complete(&comp, msg)
		if err != nil {
			return err
		}

		structPrompt.setMessage(comp.Out_last_final_message, comp.Out_last_reasoning_message, &comp.Out_usage, false)
	}

	//generate tools code
	for _, prompt := range app.Prompts {
		if prompt == structPrompt {
			continue
		}
		var err error
		comp.SystemMessage, comp.UserMessage, err = app._getToolMsg(prompt, structPrompt)
		if err != nil {
			return err
		}

		err = router.llms.Complete(&comp, msg)
		if err != nil {
			return err
		}

		prompt.setMessage(comp.Out_last_final_message, comp.Out_last_reasoning_message, &comp.Out_usage, true)
	}

	promptsFilePath := filepath.Join(folderPath, "skyalt")
	fl, err := os.ReadFile(promptsFilePath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(fl), "\n")

	for _, prompt := range app.Prompts {
		if prompt.Code == "" || prompt.Name == "" || prompt.Name == "Structures" {
			continue
		}

		lines[prompt.header_line] = "#Tool - " + prompt.Name //update tool name

		err := prompt.updateSchema()
		if err != nil {
			return err
		}
	}

	newFl := strings.Join(lines, "\n")
	err = os.WriteFile(promptsFilePath, []byte(newFl), 0644)
	if err != nil {
		return err
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

	return nil
}

func (app *ToolsPrompts) _getStructureMsg(structPrompt *ToolsPrompt) (string, string, error) {
	exampleFile := `
package main

type ExampleStruct struct {
	//<attributes>
}

func NewExampleStruct(filePath string) (*ExampleStruct, error) {
	st := &ExampleStruct{}

	//<set 'st' default values here>

	return _loadInstance(filePath, "ExampleStruct", "json", st, true)
}

//<structures functions here>
`
	systemMessage := "You are a programmer. You write code in Go language.\n"
	systemMessage += "Here is the example file with code:\n```go" + exampleFile + "```\n"

	systemMessage += "Based on user message, rewrite above code. Your job is to design structures. Write functions only if user ask for them. You may write multiple structures, but output everything in one code block.\n"

	systemMessage += "Structures can't have pointers, because they will be saved as JSON, so instead of pointer(s) use ID which is saved in map[interger or string ID].\n"

	//maybe add old file structures, because it's needed that struct and attributes names are same ...............

	userMessage := structPrompt.Prompt

	return systemMessage, userMessage, nil
}

func (app *ToolsPrompts) _getAPIFile() (string, error) {
	fl, err := os.ReadFile("apps/main.go")
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

	exampleFile := `
package main

type ExampleTool struct {
	//<tool's arguments>
}

func (st *ExampleTool) run(caller *ToolCaller, ui *UI) error {

	//<code based on prompt>

	return nil
}
`

	toolsNames := ""
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
	}

	systemMessage := "You are a programmer. You write code in Go language.\n"

	systemMessage += "Here is the file with the API functions:\n```go" + apisFile + "```\n"
	systemMessage += "Here is the file with the structures(and help functions) which represent the storage:\n```go" + storageFile + "```\n"
	systemMessage += "Here is the example file with code:\n```go" + exampleFile + "```\n"

	systemMessage += "Based on user message, rewrite above example file code. Your job is to design function(tool).\n"

	systemMessage += "Rename 'ExampleTool' with tools name based on user prompt. Don't use the word 'tool' in the name."
	if toolsNames != "" {
		systemMessage += "Here are the names of tools which already exists, don't use them:" + toolsNames + "\n"
	}

	systemMessage += "Figure out <tool's arguments> based on user prompt. They are two types of arguments - inputs and outputs. Output arguments must start with 'Out_', Input arguments don't have any prefix. All arguments must start with upper letter. Every argument must have description as comment.\n"

	systemMessage += "If you need to access the storage, use \"\" as 'filePath'. This will access default file for specific file structure. Don't redeclare storage structures(and functions) in tool's code.\n"

	systemMessage += "You may write help functions. They should start with ```func (st *ExampleTool)NameOfHelpFunction``` and output everything in one code block.\n"

	userMessage := prompt.Prompt

	return systemMessage, userMessage, nil
}

func _ToolsPrompt_getFileName(code string) (string, error) {
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
}
