/*
Copyright 2024 Milan Suk

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
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type Compile struct {
	parent *UiClients

	recompile bool

	cmd     *exec.Cmd
	running atomic.Bool

	compiling    atomic.Bool
	compiled_sec float64 //how long took to compile

	lastRecompileTicks int64 //maintenance
}

func NewCompile(parent *UiClients) *Compile {
	comp := &Compile{parent: parent}

	return comp
}

func (comp *Compile) Destroy() {
	//...
}

func Compile_GetIniPath() string {
	return "widgets/ini"
}

func (comp *Compile) Tick() error {

	//recompile
	if !OsIsTicksIn(comp.lastRecompileTicks, 3000) {

		// recompile
		if !comp.compiling.Load() && comp.NeedRecompile() {
			comp.compiling.Store(true)

			//stop process
			if comp.running.Load() {
				comp.parent.ExitWidgetsProcess() //send exit msg

				//wait until exit
				for comp.running.Load() {
					time.Sleep(50 * time.Millisecond)
				}
			}

			go func() {
				compile_st := OsTime()

				//recompile
				err := Compile_widgets()
				if err != nil {
					fmt.Println(err)
				}

				//done
				comp.compiling.Store(false)
				comp.compiled_sec = OsTime() - compile_st
			}()
		}

		comp.lastRecompileTicks = OsTicks()
	}

	//re-run
	if !comp.running.Load() && !comp.compiling.Load() {
		err := comp.run()
		if err != nil {
			return err
		}
		comp.parent.ui.SetRefresh()
	}

	return nil
}

func (comp *Compile) run() error {

	fmt.Println("Starting process")

	comp.running.Store(true)

	comp.cmd = exec.Command("./widgets/app", strconv.Itoa(comp.parent.server.port))
	//comp.cmd.Dir = ""
	comp.cmd.Stdout = os.Stdout
	comp.cmd.Stderr = os.Stderr

	err := comp.cmd.Start()
	if err != nil {
		comp.running.Store(false)
		return err
	}

	go func() {
		comp.cmd.Wait()
		comp.running.Store(false)
	}()

	//accept connection?
	{
		var err error
		comp.parent.client, err = comp.parent.server.Accept()
		if err != nil {
			log.Fatal(err)
		}
		if comp.parent.client == nil {
			log.Fatal(fmt.Errorf("client == nil"))
		}
	}

	return nil
}

func (comp *Compile) NeedRecompile() bool {

	if comp.recompile {
		comp.recompile = false
		return true
	}

	filesHash, err := OsFolderHash("widgets", []string{"app", "ini"})
	if err != nil {
		return true
	}

	fileHashStr := strconv.Itoa(int(filesHash))

	iniBytes, err := os.ReadFile(Compile_GetIniPath())
	if err == nil {
		if string(iniBytes) == fileHashStr {
			return false //same
		}
	}

	return true //different
}

func Compile_WriteHash() error {
	filesHash, err := OsFolderHash("widgets", []string{"app", "ini"})
	if err != nil {
		return err
	}
	err = os.WriteFile(Compile_GetIniPath(), []byte(strconv.Itoa(int(filesHash))), 0644)
	if err != nil {
		return err
	}
	return nil
}

func Compile_widgets() error {

	OsFileRemove("widgets/app") //bin

	files, err := Compile_get_files_info()
	if err != nil {
		return fmt.Errorf("Compile_get_files_info() failed: %w", err)
	}

	err = Compile_generate_save(files)
	if err != nil {
		return fmt.Errorf("Compile_generate_save() failed: %w", err)
	}

	//fix files
	fmt.Println("Fixing /widgets ... ")
	{
		cmd := exec.Command("goimports", "-l", "-w")
		cmd.Dir = "widgets"
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("goimports failed: %w", err)
		}
	}

	//compile
	fmt.Println("Compiling /widgets ... ")
	{
		cmd := exec.Command("go", "build", "-o", "app")
		cmd.Dir = "widgets"
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("go build failed: %w", err)
		}
	}
	fmt.Println("done")

	//compare & copy old app.go into history(mkdir) ...

	err = Compile_WriteHash()
	if err != nil {
		return err
	}

	return nil
}

func Compile_generate_save(files []CompileWidgetFile) error {

	var code string
	code += "package main\n\n"

	code += "func _skyalt_save() {\n"
	for _, f := range files {
		if f.IsFile {
			code += fmt.Sprintf("\tif g_%s != nil {\n\t\t_write_file(\"%s-%s\", g_%s)\n\t\tg_%s = nil\n\t}\n", f.Name, f.Name, f.Name, f.Name, f.Name)
		}
	}
	code += "}\n"

	//write the code into the file
	{
		//fmt.Println(code)
		err := os.WriteFile("widgets/sdk_save.go", []byte(code), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

type CompileWidgetFile struct {
	Name   string
	IsFile bool

	Build bool
	Input bool
	Draw  bool
}

func Compile_get_files_info() ([]CompileWidgetFile, error) {
	var widgets []CompileWidgetFile

	sdkDir, err := os.ReadDir("widgets")
	if err != nil {
		return nil, err
	}
	for _, file := range sdkDir {
		stName, found := strings.CutSuffix(file.Name(), ".go")
		if !file.IsDir() && found && !strings.HasPrefix(file.Name(), "sdk_") {

			fileCode, err := os.ReadFile(filepath.Join("widgets", file.Name()))
			if err != nil {
				return nil, err
			}

			build_pos, input_pos, draw_pos, isFile, err := Compile_findFileProperties(file.Name(), string(fileCode), stName)
			if err != nil {
				return nil, err
			}
			if build_pos >= 0 || input_pos >= 0 || draw_pos >= 0 { //is widget
				widgets = append(widgets, CompileWidgetFile{Name: stName, IsFile: isFile, Build: build_pos >= 0, Input: input_pos >= 0, Draw: draw_pos >= 0})
			}
		}
	}

	return widgets, nil
}

func Compile_findFileProperties(ghostPath string, code string, stName string) (int, int, int, bool, error) {

	node, err := parser.ParseFile(token.NewFileSet(), ghostPath, code, parser.ParseComments)
	if err != nil {
		return -1, -1, -1, false, err
	}

	build_pos := -1
	input_pos := -1
	draw_pos := -1
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:

			tp := ""
			if x.Recv != nil && len(x.Recv.List) > 0 {
				tp = string(code[x.Recv.List[0].Type.Pos()-1 : x.Recv.List[0].Type.End()-1])
			}

			//function
			if tp == "*"+stName {
				if x.Name.Name == "Build" {
					build_pos = int(x.Pos())
				}
				if x.Name.Name == "Input" {
					input_pos = int(x.Pos())
				}
				if x.Name.Name == "Draw" {
					draw_pos = int(x.Pos())
				}
			}
		}
		return true
	})

	//global vars
	isFile := false
	for _, decl := range node.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			if genDecl.Tok == token.VAR {
				for _, spec := range genDecl.Specs {
					if valueSpec, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range valueSpec.Names {
							if name.Name == "g_"+stName {
								isFile = true
							}
						}
					}
				}
			}
		}
	}

	return build_pos, input_pos, draw_pos, isFile, nil
}
