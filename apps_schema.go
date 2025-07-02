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
	"strings"
	"unicode"
)

type ToolsOpenAI_completion_tool_function_parameters_properties struct {
	Type        string   `json:"type"` //"number", "string"
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default,omitempty"`

	Items *ToolsOpenAI_completion_tool_function_parameters_properties `json:"items,omitempty"` //for arrays
}
type ToolsOpenAI_completion_tool_schema struct {
	Type                 string   `json:"type"` //"object"
	Required             []string `json:"required,omitempty"`
	AdditionalProperties bool     `json:"additionalProperties"`

	Properties map[string]*ToolsOpenAI_completion_tool_function_parameters_properties `json:"properties"`
}
type ToolsOpenAI_completion_tool_function struct {
	Name        string                             `json:"name"`
	Description string                             `json:"description"`
	Parameters  ToolsOpenAI_completion_tool_schema `json:"parameters"`
	Strict      bool                               `json:"strict"`
}

type ToolsOpenAI_completion_tool struct {
	Type     string                               `json:"type"` //"object"
	Function ToolsOpenAI_completion_tool_function `json:"function"`
}

func NewToolsOpenAI_completion_tool(name, description string) *ToolsOpenAI_completion_tool {
	fn := &ToolsOpenAI_completion_tool{Type: "function"}
	fn.Function = ToolsOpenAI_completion_tool_function{Name: name, Description: description, Strict: false}
	fn.Function.Parameters.Type = "object"
	fn.Function.Parameters.AdditionalProperties = false
	fn.Function.Parameters.Properties = make(map[string]*ToolsOpenAI_completion_tool_function_parameters_properties)
	return fn
}

func (prm *ToolsOpenAI_completion_tool_schema) AddParam(name, typee, description string) *ToolsOpenAI_completion_tool_function_parameters_properties {
	//array
	var items *ToolsOpenAI_completion_tool_function_parameters_properties
	if strings.Contains(strings.ToLower(typee), "[]") {
		subType := strings.ReplaceAll(typee, "[]", "")

		typee = "array"
		items = &ToolsOpenAI_completion_tool_function_parameters_properties{Type: ToolsOpenAI_convertTypeToSchemaType(subType)}
	}

	isOptional := strings.Contains(description, "[optional]")
	if isOptional {
		description = strings.ReplaceAll(description, "[optional]", "")
	} else {
		prm.Required = append(prm.Required, name)
	}

	var options []string
	options_start := "[options:"
	options_pos := strings.Index(description, options_start)
	if options_pos > 0 {
		opt := description[options_pos+len(options_start):] //cut
		d := strings.Index(opt, "]")
		if d >= 0 {
			opt = opt[:d] //cut
			options = strings.Split(opt, ",")

			description = description[:options_pos] + description[options_pos+len(options_start)+d+1:]
		}

	}

	//clean
	{
		description = strings.TrimSpace(description)
		for o := range options {
			options[o] = strings.TrimSpace(options[o])
		}
	}

	p := &ToolsOpenAI_completion_tool_function_parameters_properties{Type: ToolsOpenAI_convertTypeToSchemaType(typee), Description: description, Items: items, Enum: options}
	prm.Properties[name] = p

	return p
}

func ToolsOpenAI_convertTypeToSchemaType(tp string) string {
	if strings.Contains(strings.ToLower(tp), "float") ||
		strings.Contains(strings.ToLower(tp), "int") ||
		strings.Contains(strings.ToLower(tp), "number") {
		tp = "number"
	}

	if strings.Contains(strings.ToLower(tp), "bool") {
		tp = "boolean"
	}

	return tp
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

func BuildToolsOpenAI_completion_tool(toolName string, fileName string, code any) (*ToolsOpenAI_completion_tool, error) {
	node, err := parser.ParseFile(token.NewFileSet(), fileName, code, parser.ParseComments)
	if LogsError(err) != nil {
		return nil, err
	}

	var oai *ToolsOpenAI_completion_tool
	isIgnored := false

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

			if toolName != typeSpec.Name.Name {
				continue
			}

			if strings.Contains(structDoc, "[ignore]") {
				isIgnored = true
			}

			if !isIgnored {
				oai = NewToolsOpenAI_completion_tool(typeSpec.Name.Name, structDoc)
			}

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
					nm := strings.Join(fieldNames, ", ")
					tp := _exprToString(field.Type)

					//if tp == "UI" {
					//	tool.UiAttrs = append(tool.UiAttrs, nm)
					//skip
					if strings.HasPrefix(nm, "Out") || (len(nm) > 0 && unicode.IsLower(rune(nm[0]))) {
						//skip
					} else {
						if oai != nil {
							oai.Function.Parameters.AddParam(nm, tp, fieldDoc)
						}
					}
				}
			}
		}
	}

	if !isIgnored {
		if oai == nil {
			return nil, LogsErrorf("struct '%s' not found", toolName)
		}
	}

	return oai, nil
}
