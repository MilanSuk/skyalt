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
	"image/color"
)

var g_syntax_yellow = color.RGBA{200, 200, 50, 255}
var g_syntax_red = color.RGBA{200, 50, 50, 255}
var g_syntax_green = color.RGBA{50, 200, 50, 255}
var g_syntax_blue = color.RGBA{50, 50, 200, 255}
var g_syntax_purple = color.RGBA{200, 50, 200, 255}

// Item represents a highlighted section of code with its position and color
type Item struct {
	Start int
	End   int
	Color color.RGBA
}

// visitor implements ast.Visitor to traverse the AST and highlight items
type visitor struct {
	fset        *token.FileSet
	items       []Item
	structNames map[string]bool       // Set of struct type names
	scopes      []map[string]bool     // Stack of variable scopes
	keywords    map[string]color.RGBA // Keywords to highlight
}

// NewVisitor initializes a new visitor with default settings
func NewVisitor(fset *token.FileSet) *visitor {
	return &visitor{
		fset:        fset,
		items:       []Item{},
		structNames: make(map[string]bool),
		scopes:      []map[string]bool{make(map[string]bool)}, // Global scope
		keywords: map[string]color.RGBA{
			"func":   g_syntax_yellow, // Yellow for keywords
			"var":    g_syntax_yellow,
			"type":   g_syntax_yellow,
			"return": g_syntax_yellow,
		},
	}
}

// Visit implements the ast.Visitor interface
func (v *visitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		// Exiting a scope
		if len(v.scopes) > 1 {
			v.scopes = v.scopes[:len(v.scopes)-1]
		}
		return nil
	}

	switch n := node.(type) {
	case *ast.File:
		// Collect struct type names at package level
		for _, decl := range n.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
				for _, spec := range genDecl.Specs {
					if tspec, ok := spec.(*ast.TypeSpec); ok {
						if _, ok := tspec.Type.(*ast.StructType); ok {
							v.structNames[tspec.Name.Name] = true
						}
					}
				}
			}
		}
		return v

	case *ast.BlockStmt:
		// Entering a new block scope
		v.scopes = append(v.scopes, make(map[string]bool))
		return v

	case *ast.FuncDecl:
		// Highlight "func" keyword
		start := v.fset.Position(n.Pos()).Offset
		v.items = append(v.items, Item{Start: start, End: start + len("func"), Color: v.keywords["func"]})
		// Highlight function name (declaration)
		nameStart := v.fset.Position(n.Name.Pos()).Offset
		nameEnd := v.fset.Position(n.Name.End()).Offset
		v.items = append(v.items, Item{Start: nameStart, End: nameEnd, Color: g_syntax_blue}) // Blue for functions
		// New scope for function body
		v.scopes = append(v.scopes, make(map[string]bool))
		// Add parameters to scope
		if n.Type.Params != nil {
			for _, param := range n.Type.Params.List {
				for _, name := range param.Names {
					v.scopes[len(v.scopes)-1][name.Name] = true
				}
			}
		}
		return v

	case *ast.GenDecl:
		switch n.Tok {
		case token.VAR:
			// Highlight "var" keyword
			start := v.fset.Position(n.TokPos).Offset
			v.items = append(v.items, Item{Start: start, End: start + len("var"), Color: v.keywords["var"]})
			// Highlight variable names and add to scope
			for _, spec := range n.Specs {
				vspec := spec.(*ast.ValueSpec)
				for _, name := range vspec.Names {
					nameStart := v.fset.Position(name.Pos()).Offset
					nameEnd := v.fset.Position(name.End()).Offset
					v.items = append(v.items, Item{Start: nameStart, End: nameEnd, Color: g_syntax_green}) // Green for variables
					v.scopes[len(v.scopes)-1][name.Name] = true
				}
			}
		case token.TYPE:
			// Highlight "type" keyword
			start := v.fset.Position(n.TokPos).Offset
			v.items = append(v.items, Item{Start: start, End: start + len("type"), Color: v.keywords["type"]})
			// Highlight struct names and fields
			for _, spec := range n.Specs {
				tspec := spec.(*ast.TypeSpec)
				if structType, ok := tspec.Type.(*ast.StructType); ok {
					// Highlight struct name
					nameStart := v.fset.Position(tspec.Name.Pos()).Offset
					nameEnd := v.fset.Position(tspec.Name.End()).Offset
					v.items = append(v.items, Item{Start: nameStart, End: nameEnd, Color: g_syntax_red}) // Red for structs
					// Highlight struct field names
					for _, field := range structType.Fields.List {
						for _, fieldName := range field.Names {
							fieldStart := v.fset.Position(fieldName.Pos()).Offset
							fieldEnd := v.fset.Position(fieldName.End()).Offset
							v.items = append(v.items, Item{Start: fieldStart, End: fieldEnd, Color: g_syntax_purple}) // Purple for field names
						}
					}
				}
			}
		}
		return v

	case *ast.AssignStmt:
		// Handle short variable declarations (:=)
		if n.Tok == token.DEFINE {
			for _, lhs := range n.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok {
					nameStart := v.fset.Position(ident.Pos()).Offset
					nameEnd := v.fset.Position(ident.End()).Offset
					v.items = append(v.items, Item{Start: nameStart, End: nameEnd, Color: g_syntax_green})
					v.scopes[len(v.scopes)-1][ident.Name] = true
				}
			}
		}
		return v

	case *ast.Ident:
		// Highlight variable usage
		if v.isVariable(n.Name) {
			start := v.fset.Position(n.Pos()).Offset
			end := v.fset.Position(n.End()).Offset
			v.items = append(v.items, Item{Start: start, End: end, Color: g_syntax_green}) // Green for variable usage
		} else if v.isStructName(n.Name) {
			// Highlight struct name usage
			start := v.fset.Position(n.Pos()).Offset
			end := v.fset.Position(n.End()).Offset
			v.items = append(v.items, Item{Start: start, End: end, Color: g_syntax_red}) // Red for struct usage
		} else {
			fmt.Println("unknown")
		}
		return v

	case *ast.SelectorExpr:
		// Highlight struct field names in usage (e.g., instance.field)
		//if ident, ok := n.Sel.(*ast.Ident); ok {
		fieldStart := v.fset.Position(n.Sel.Pos()).Offset
		fieldEnd := v.fset.Position(n.Sel.End()).Offset
		v.items = append(v.items, Item{Start: fieldStart, End: fieldEnd, Color: g_syntax_purple}) // Purple for field names
		//}
		return v

	case *ast.CallExpr:
		// Highlight function name when called
		if ident, ok := n.Fun.(*ast.Ident); ok {
			funcStart := v.fset.Position(ident.Pos()).Offset
			funcEnd := v.fset.Position(ident.End()).Offset
			v.items = append(v.items, Item{Start: funcStart, End: funcEnd, Color: g_syntax_blue}) // Blue for function calls
		}
		return v
	}
	return v
}

// isVariable checks if a name is a variable in the current scope
func (v *visitor) isVariable(name string) bool {
	for i := len(v.scopes) - 1; i >= 0; i-- {
		if v.scopes[i][name] {
			return true
		}
	}
	return false
}

// isStructName checks if a name is a known struct type
func (v *visitor) isStructName(name string) bool {
	return v.structNames[name]
}

// getItems parses the code and returns highlighted items
func getItems(code string) ([]Item, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "source.go", code, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	v := NewVisitor(fset)
	ast.Walk(v, file)
	return v.items, nil
}
