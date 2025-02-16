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
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func GetToolsList(tool string) ([]string, []string, error) {
	files, err := os.ReadDir(tool)
	if err != nil {
		return nil, nil, err
	}
	var tools []string
	var scripts []string
	for _, file := range files {
		if file.IsDir() {
			tools = append(tools, file.Name())
		} else if strings.HasSuffix(file.Name(), ".sky") {
			scripts = append(scripts, file.Name())
		}
	}
	return tools, scripts, nil
}

func NeedCompileTool(tool string) bool {

	dst_tool := filepath.Join("temp", tool)

	//check time stamp
	{
		js := GetToolTimeStamp(tool)
		js2, _ := os.ReadFile(filepath.Join(dst_tool, "ini"))
		if !bytes.Equal(js, js2) {
			return true
		}
	}

	//check if tool bin exist
	_, err := os.Stat(filepath.Join(dst_tool, "bin"))
	return os.IsNotExist(err)
}

func GetToolTimeStamp(tool string) []byte {
	infoSdk, _ := os.Stat("tools/sdk.go")
	infoTool, _ := os.Stat(filepath.Join(tool, "tool.go"))
	js, _ := json.Marshal(infoSdk.ModTime().UnixNano() + infoTool.ModTime().UnixNano())

	return js
}

func CompileTool(src_tool string) error {
	stName := filepath.Base(src_tool)
	dst_tool := filepath.Join("temp", src_tool)
	err := os.MkdirAll(dst_tool, os.ModePerm)
	if err != nil {
		return err
	}

	src_toolPath := filepath.Join(src_tool, "tool.go")
	dst_toolPath := filepath.Join(dst_tool, "tool.go")
	dst_mainPath := filepath.Join(dst_tool, "main.go")
	dst_iniPath := filepath.Join(dst_tool, "ini")
	dst_binPath := filepath.Join(dst_tool, "bin")

	//add tool file
	{
		codeOrig, err := os.ReadFile(src_toolPath)
		if err != nil {
			return err
		}
		code := ApplySandbox(string(codeOrig), false)

		err = os.WriteFile(dst_toolPath, []byte(code), 0644)
		if err != nil {
			return err
		}
	}

	//add sdk.go
	{
		sdk, err := os.ReadFile("tools/sdk.go")
		if err != nil {
			return err
		}

		err = os.WriteFile(dst_mainPath, []byte(strings.Replace(string(sdk), "_replace_with_tool_structure_", stName, 1)), 0644)
		if err != nil {
			return err
		}
	}

	//remove old bin
	{
		os.Remove(dst_binPath)
	}

	//fix files
	{
		fmt.Printf("Fixing %s ... ", dst_tool)
		st := float64(time.Now().UnixMilli()) / 1000
		cmd := exec.Command("goimports", "-l", "-w", ".")
		cmd.Dir = dst_tool
		var stderr bytes.Buffer
		cmd.Stderr = &stderr //os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("goimports failed: %s", stderr.String())
		}
		fmt.Printf("done in %.3fsec\n", (float64(time.Now().UnixMilli())/1000)-st)
	}

	//update packages
	{
		fmt.Printf("Fixing %s ... ", dst_tool)
		st := float64(time.Now().UnixMilli()) / 1000
		cmd := exec.Command("go", "mod", "tidy") //"go mod init <name>" If it not exists? ....
		cmd.Dir = dst_tool
		var stderr bytes.Buffer
		cmd.Stderr = &stderr //os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("goimports failed: %s", stderr.String())
		}
		fmt.Printf("done in %.3fsec\n", (float64(time.Now().UnixMilli())/1000)-st)
	}

	//compile
	{
		fmt.Printf("Compiling %s ... ", dst_tool)
		st := float64(time.Now().UnixMilli()) / 1000
		cmd := exec.Command("go", "build", "-o", "bin")
		cmd.Dir = dst_tool
		var stderr bytes.Buffer
		cmd.Stderr = &stderr //os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("compiler failed: %s", stderr.String())
		}
		fmt.Printf("done in %.3fsec\n", (float64(time.Now().UnixMilli())/1000)-st)
	}

	//write time stamp
	{
		err = os.WriteFile(dst_iniPath, GetToolTimeStamp(src_tool), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func ApplySandbox(code string, reverse bool) string {
	fl, err := os.ReadFile("tools/sdk_sandbox_fns.txt")
	if err != nil {
		log.Fatal(err)
	}

	lines := strings.Split(string(fl), "\n")
	for _, ln := range lines {
		if ln == "" {
			continue //skip
		}
		var src, dst string
		n, err := fmt.Sscanf(ln, "%s %s", &src, &dst)
		if n != 2 || err != nil {
			log.Fatal(err)
		}

		if reverse {
			src, dst = dst, src
		}

		code = strings.ReplaceAll(code, src, dst)
	}

	return code
}

func ConvertCodeIntoTool(folder string) (*OpenAI_completion_tool, *Anthropic_completion_tool, error) {
	stName := OsFileGetNameWithoutExt(folder)

	toolPath := filepath.Join(folder, "tool.go")

	node, err := parser.ParseFile(token.NewFileSet(), toolPath, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing file: %v", err)
	}

	var oai *OpenAI_completion_tool
	var ant *Anthropic_completion_tool

	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			structDoc := ""
			if genDecl.Doc != nil {
				structDoc = strings.TrimSpace(genDecl.Doc.Text())
			}

			if stName != typeSpec.Name.Name {
				continue
			}

			oai = NewOpenAI_completion_tool(typeSpec.Name.Name, structDoc)
			ant = NewAnthropic_completion_tool(typeSpec.Name.Name, structDoc)

			for _, field := range structType.Fields.List {
				fieldNames := make([]string, len(field.Names))
				for i, name := range field.Names {
					fieldNames[i] = name.Name
				}

				fieldDoc := ""
				if field.Doc != nil {
					fieldDoc = strings.TrimSpace(field.Doc.Text())
				}
				if field.Comment != nil {
					fieldDoc = strings.TrimSpace(field.Comment.Text())
				}

				if len(fieldNames) > 0 {
					oai.Function.Parameters.AddParam(strings.Join(fieldNames, ", "), _exprToString(field.Type), fieldDoc)
					ant.Input_schema.AddParam(strings.Join(fieldNames, ", "), _exprToString(field.Type), fieldDoc)
				}
			}
		}
	}

	if oai == nil {
		return nil, nil, fmt.Errorf("struct %s not found", stName)
	}

	return oai, ant, nil
}

func _exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	//case *ast.InterfaceType:
	//	return "any" // "oneOf": [{ "type": "string" }, { "type": "number" }]
	case *ast.StarExpr:
		return "*" + _exprToString(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + _exprToString(t.Elt)
		}
		return "[" + _exprToString(t.Len) + "]" + _exprToString(t.Elt)
	case *ast.SelectorExpr:
		return _exprToString(t.X) + "." + t.Sel.Name
	default:
		return fmt.Sprintf("%T", expr)
	}
}

type ScriptItem struct {
	Name        string
	Type        string
	Description string

	Exps []string
}

func GetScriptFileInfo(toolPath string) (string, string, []string, map[string]ScriptItem, error) {
	stName := OsFileGetNameWithoutExt(toolPath)

	fl, err := os.ReadFile(toolPath)
	if err != nil {
		return "", "", nil, nil, err
	}

	parts := strings.Split(string(fl), "[user]")

	if len(parts) < 2 {
		return "", "", nil, nil, fmt.Errorf("No [user] tag found")
	}

	doc := strings.TrimSpace(parts[0])
	if doc == "" {
		return "", "", nil, nil, fmt.Errorf("Documentation not found")

	}

	var msgs []string
	for i := 1; i < len(parts); i++ {
		msgs = append(msgs, strings.TrimSpace(parts[i]))
	}

	variables := make(map[string]ScriptItem)

	for _, msg := range msgs {
		pattern := regexp.MustCompile(`\{\{[^{}]*\}\}`)
		matches := pattern.FindAllString(msg, -1)

		for _, match := range matches {
			var dst ScriptItem
			err := json.Unmarshal([]byte(match[1:len(match)-1]), &dst)
			if err == nil {
				src := variables[dst.Name]
				if dst.Type != "" {
					if src.Type != "" && src.Type != dst.Type {
						fmt.Println("Warning: Type already set", match)
					}
					src.Type = dst.Type
				}
				if dst.Description != "" {
					if src.Description != "" && src.Description != dst.Description {
						fmt.Println("Warning: Description already set", match)
					}
					src.Description = dst.Description
				}
				src.Name = dst.Name
				src.Exps = append(src.Exps, match)
				variables[dst.Name] = src
			} else {
				fmt.Println("Warning:", err, match)
			}
		}
	}

	return stName, doc, msgs, variables, nil
}

func ConvertScriptIntoTool(toolPath string) (*OpenAI_completion_tool, *Anthropic_completion_tool, error) {

	stName, doc, _, variables, err := GetScriptFileInfo(toolPath)
	if err != nil {
		return nil, nil, err
	}

	oai := NewOpenAI_completion_tool(stName, doc)
	ant := NewAnthropic_completion_tool(stName, doc)

	for name, it := range variables {
		oai.Function.Parameters.AddParam(name, it.Type, it.Description)
		ant.Input_schema.AddParam(name, it.Type, it.Description)
	}

	return oai, ant, nil
}
