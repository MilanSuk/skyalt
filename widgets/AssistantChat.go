package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"image/color"
	"os"
	"path/filepath"
	"strings"
)

type AssistantChat struct {
	Prompt string
	Picks  []LayoutPick
}

func (layout *Layout) AddAssistant(x, y, w, h int, props *AssistantChat) *AssistantChat {
	layout._createDiv(x, y, w, h, "Assistant", props.Build, nil, nil)
	return props
}

var g_AssistantChat *AssistantChat

func NewFile_AssistantChat() *AssistantChat {
	if g_AssistantChat == nil {
		g_AssistantChat = &AssistantChat{}
		_read_file("AssistantChat-AssistantChat", g_AssistantChat)
	}
	return g_AssistantChat
}

func (st *AssistantChat) Build(layout *Layout) {
	//...
}

func (ast *AssistantChat) Send() {
	ast.Prompt = strings.TrimSpace(ast.Prompt)

	if len(ast.Prompt) < 5 {
		Layout_WriteError(errors.New("prompt is too short"))
		return
	}

	SystemPrompt := "You are an AI programming assistant, who enjoys precision and carefully follows the user's requirements. You write code only in Golang."
	UserPrompt, err := Prompt_1(FixUserPrompt(ast.Prompt), ast.Picks)
	if err != nil {
		Layout_WriteError(fmt.Errorf("Prompt_1() %w", err))
		return
	}
	fmt.Println("Prompt_1_input:", UserPrompt)

	var msgs []OpenAI_chat_msg
	msgs = append(msgs, OpenAI_chat_msg{Role: "system", Content: SystemPrompt})
	msgs = append(msgs, OpenAI_chat_msg{Role: "user", Content: UserPrompt})
	chat1 := NewGlobal_OpenAI_chat("AssistantChat_1")
	chat1.Properties.Messages = msgs
	chat1.Properties.Response_format.Type = "json_object"
	chat1.done = func() {
		if chat1.Out == "" {
			Layout_WriteError(fmt.Errorf("Chat1 output is empty"))
			return
		}
		fmt.Println("Prompt_1_output:", chat1.Out)

		UserPrompt, err := Prompt_2([]byte(chat1.Out), FixUserPrompt(ast.Prompt), ast.Picks)
		if err != nil {
			Layout_WriteError(fmt.Errorf("Prompt_2() %w", err))
			return
		}
		fmt.Println("Prompt_2_input:", UserPrompt)

		var msgs []OpenAI_chat_msg
		msgs = append(msgs, OpenAI_chat_msg{Role: "system", Content: SystemPrompt})
		msgs = append(msgs, OpenAI_chat_msg{Role: "user", Content: UserPrompt})
		chat2 := NewGlobal_OpenAI_chat("AssistantChat_2")
		chat2.Properties.Messages = msgs
		chat2.Properties.Response_format.Type = "json_object"
		chat2.done = func() {
			if chat2.Out == "" {
				Layout_WriteError(fmt.Errorf("Chat2 output is empty"))
				return
			}
			fmt.Println("Prompt_2_output:", chat2.Out)

			//project new code to file(s)
			type OutputFile struct {
				Name    string
				Content string
			}
			type Output struct {
				Files []OutputFile
			}
			var out Output
			err := json.Unmarshal([]byte(chat2.Out), &out)
			if err != nil {
				Layout_WriteError(err)
				return
			}
			for _, f := range out.Files {
				//write file
				err = os.WriteFile("widgets/"+f.Name, []byte(f.Content), 0644)
				if err != nil {
					Layout_WriteError(err)
					return
				}
			}

			ast.Prompt = ""

			Layout_Recompile()
		}
		chat2.Start()
	}
	chat1.Start()
}

func FixUserPrompt(userPrompt string) string {
	userPrompt = _Assistant_RemoveFormatingRGBA(userPrompt)

	if userPrompt == "" {
		return ""
	}

	//1st letter must be Upper
	userPrompt = strings.ToUpper(userPrompt[:1]) + userPrompt[1:]

	//ends with dot
	if !strings.ContainsAny(userPrompt[len(userPrompt)-1:], ".!?") {
		userPrompt += "."
	}

	return userPrompt
}

//maybe every widget can have agent? .............

const base_promp = `Here are the <description> : <JSON> responses:
- Open/Show widget: {"open": true, "name": <widget_name>}
- Create new widget: {"create": true, "name": <widget_name>}
- Modify widget: {"code": true}
- Change data(not code): {"data": true}
- General question about code: {"q_code": true}
- General question about data: {"q_data": true}
- General question about anything else: {"q_other": true}

Note: widget = app,layout,page,program.

This is the prompt from user:
%s

Note: A prompt can have marks(for ex.: {"mark":5}). Every mark refers to some line in code(for ex.: layout.AddText() //{"mark":5}).")

Your job is to match the user prompt to <description> and output JSON <response>. Please respond in the JSON format only!
`

func Prompt_1(userPrompt string, picks []LayoutPick) (string, error) {

	var str strings.Builder

	//get list
	widgets, err := Compile_get_files_info()
	if err != nil {
		return "", err
	}

	//build list
	var apps strings.Builder
	for _, w := range widgets {
		if w.IsFile {
			apps.WriteString(w.Name + "\n")
		}
	}

	str.WriteString("\n\n// Here is the list of Widgets:\n")
	str.WriteString(apps.String() + "\n\n")
	str.WriteString(fmt.Sprintf(base_promp, userPrompt))

	return str.String(), nil
}

func Prompt_2(output_1 []byte, userPrompt string, picks []LayoutPick) (string, error) {
	type Out struct {
		Open    bool
		Create  bool
		Code    bool
		Data    bool
		Q_code  bool
		Q_data  bool
		Q_other bool

		Name string
	}
	var out Out

	err := json.Unmarshal(output_1, &out)
	if err != nil {
		return "", err
	}

	info, err := Compile_get_files_info()
	if err != nil {
		return "", err
	}

	var structs strings.Builder
	var addFns strings.Builder
	var fileFns strings.Builder

	for _, f := range info {
		for _, fn := range f.Structs {
			structs.WriteString(fn)
			structs.WriteByte('\n')
		}
		for _, fn := range f.AddFuncs {
			addFns.WriteString(fn)
			addFns.WriteByte('\n')
		}
		for _, fn := range f.FileFuncs {
			fileFns.WriteString(fn)
			fileFns.WriteByte('\n')
		}
	}
	layoutStFns, err := _build_sdk_layout()
	if err != nil {
		return "", err
	}

	var str strings.Builder
	if out.Open {
		str.WriteString("\n\n// Here are the functions to show Widgets:\n")
		str.WriteString("```go\n" + addFns.String() + "\n```")

		str.WriteString("\n\n// Here are the functions to access data(configurations) from disk:\n")
		str.WriteString("```go\n" + fileFns.String() + "\n```")

		str.WriteString("\n\n// Here are the files:\n")
		edit_file := "ShowApp.go"
		_AssistantChat_buildPromptFile(edit_file, userPrompt, picks, &str)

		str.WriteString("Note: Build() has layout.Add...(), which represent opened(shown) widget. Rewrite this line(same coord: 0,0,1,1) to show widget from the prompt.\n\n")

		str.WriteString("This is the prompt from the user:\n")
		str.WriteString(_AssistantChat_buildUserPrompt(userPrompt, picks) + "\n")
		str.WriteString("\n" + `Note: A prompt can have marks(for ex.: {"mark":5}). Every mark refers to some line in code(for ex.: layout.AddText() //{"mark":5}) inside files above.` + "\n")

		str.WriteString(fmt.Sprintf("\nYour job is implement the request(user prompt) by editting file %s.\n\n", edit_file))

		str.WriteString("Don't add main() function to the code. Please respond in the JSON format {files: [{name: <name.go>, content: <code>}]}")

	} else if out.Create {
		str.WriteString("\n// Here are the widget's structures:\n")
		str.WriteString("```go\n" + structs.String() + "\n```")

		str.WriteString("\n\n// Here are the functions to add Widgets:\n")
		str.WriteString("```go\n" + addFns.String() + "\n```")

		str.WriteString("\n\n// Here are the functions to access data(configurations) from disk:\n")
		str.WriteString("```go\n" + fileFns.String() + "\n```")

		str.WriteString("\n\n// Here are the files:\n")
		edit_struct := strings.TrimSuffix(out.Name, filepath.Ext(out.Name))

		{
			//define new file as example

			str.WriteString(fmt.Sprintf("file: %s.go\n", edit_struct))
			edit_code := fmt.Sprintf(`package main
				type %s struct {
					layout *Layout
				}
				func (st *%s) Build(layout *Layout) {
				}`, edit_struct, edit_struct)
			str.WriteString("```go\n" + edit_code + "```\n\n")

			//add files from picks
			go_files := _AssistantChat_createCodeFileList(picks)
			for _, file := range go_files {
				_AssistantChat_buildPromptFile(file, userPrompt, picks, &str)
			}
		}

		str.WriteString("This is the prompt from the user:\n")
		str.WriteString(_AssistantChat_buildUserPrompt(userPrompt, picks) + "\n")
		str.WriteString("\n" + `Note: A prompt can have marks(for ex.: {"mark":5}). Every mark refers to some line in code(for ex.: layout.AddText() //{"mark":5}) inside files above.` + "\n")

		str.WriteString(fmt.Sprintf("\nYour job is implement the request(user prompt) by editting code in file %s.go.\n\n", edit_struct))

		str.WriteString("Don't add main() function to the code. Please respond in the JSON format {files: [{name: <name.go>, content: <code>}]}")

	} else if out.Code {
		str.WriteString("\n// Here are the layout structures and functions:\n")
		str.WriteString("```go\n" + layoutStFns.String() + "\n```")

		str.WriteString("\n// Here are the widget's structures:\n")
		str.WriteString("```go\n" + structs.String() + "\n```")

		str.WriteString("\n\n// Here are the functions to add Widgets:\n")
		str.WriteString("```go\n" + addFns.String() + "\n```")

		str.WriteString("\n\n// Here are the functions to access data(configurations) from disk:\n")
		str.WriteString("```go\n" + fileFns.String() + "\n```")

		if len(picks) > 0 {
			str.WriteString("\n\n// Here are the files:\n")
			//add currently open file into 'picks' ...

			//add files from picks
			go_files := _AssistantChat_createCodeFileList(picks)
			for _, file := range go_files {
				_AssistantChat_buildPromptFile(file, userPrompt, picks, &str)
			}
		}

		str.WriteString("This is the prompt from the user:\n")
		str.WriteString(_AssistantChat_buildUserPrompt(userPrompt, picks) + "\n")
		str.WriteString("\n" + `Note: A prompt can have marks(for ex.: {"mark":5}). Every mark refers to some line in code(for ex.: layout.AddText() //{"mark":5}) inside files above.` + "\n")

		str.WriteString("\nYour job is implement the request(user prompt) by editting code in files. Usually you edit only one file.\n\n")

		str.WriteString("Don't add main() function to the code. Please respond in the JSON format {files: [{name: <name.go>, content: <code>}]}")

	} else if out.Data {
		str.WriteString("\n// Here are the widget's structures:\n")
		str.WriteString("```go\n" + structs.String() + "\n```")

		str.WriteString("\n\n// Here are the functions to access data(configurations) from disk:\n")
		str.WriteString("```go\n" + fileFns.String() + "\n```")

		str.WriteString("\n\n// Here are the files:\n")
		//add files from picks + 'sdk_change.go'
		go_files := _AssistantChat_createCodeFileList(picks)
		go_files = append(go_files, "sdk_change.go")
		for _, file := range go_files {
			_AssistantChat_buildPromptFile(file, userPrompt, picks, &str)
		}

		str.WriteString("This is the prompt from the user:\n")
		str.WriteString(_AssistantChat_buildUserPrompt(userPrompt, picks) + "\n")
		str.WriteString("\n" + `Note: A prompt can have marks(for ex.: {"mark":5}). Every mark refers to some line in code(for ex.: layout.AddText() //{"mark":5}) inside files above.` + "\n")

		str.WriteString("\nYour job is implement the request(user prompt) by editting code in file sdk_change.go. Use functions NewFile_?() to read/write data.\n\n")

		str.WriteString("Don't add main() function to the code. Please respond in the JSON format {files: [{name: <name.go>, content: <code>}]}")
	} else if out.Q_code {
		//...
		//system promp říká že je programátor, který píše kód! ...
	} else if out.Q_data {
		//...
	} else if out.Q_other {
		//...
	} else {
		return "", fmt.Errorf("unrecognized user prompt")
	}

	return str.String(), nil
}

func _AssistantChat_createCodeFileList(picks []LayoutPick) []string {
	var list []string
	for pk := range picks {
		found := false
		for f := range list {
			if list[f] == picks[pk].File {
				found = true
				break
			}
		}
		if !found {
			list = append(list, picks[pk].File)
		}
	}
	return list
}

func _AssistantChat_buildUserPrompt(userPrompt string, picks []LayoutPick) string {
	prompt := _Assistant_RemoveFormatingRGBA(userPrompt)

	//add marks
	for pk := range picks {
		mark := LayoutPick_getMark(pk)

		if strings.Contains(prompt, mark) {
			line := picks[pk].Line
			tip := picks[pk].Tip
			st := strings.TrimSuffix(picks[pk].File, filepath.Ext(picks[pk].File))
			var dst string
			if picks[pk].Grid_w == 0 && picks[pk].Grid_h == 0 {
				dst = fmt.Sprintf("{\"mark\":\"%d\", \"source\":\"(from NewFile_%s()): %s\"}", line, st, tip)
			} else {
				dst = fmt.Sprintf("{\"mark\":\"%d\", \"grid\":\"%d,%d,%d,%d\"}", line, picks[pk].Grid_x, picks[pk].Grid_y, picks[pk].Grid_w, picks[pk].Grid_h)
			}
			prompt = strings.ReplaceAll(prompt, mark, dst)
		}
	}

	prompt = strings.ReplaceAll(prompt, "\n", " ")
	prompt = strings.TrimSpace(prompt)

	return prompt
}

func _AssistantChat_addMarksToCode(file string, userPrompt string, picks []LayoutPick) ([]string, error) {

	path := "widgets/" + file //+ ".go"
	fl, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(fl), "\n")

	//remove old marks
	for i := len(lines) - 1; i >= 0; i-- {
		d := strings.Index(lines[i], "//{\"mark\"")
		if d >= 0 {
			lines[i] = lines[i][:d] //cut
		}
	}

	//add marks as comments into code
	for pk := range picks {
		if picks[pk].File == file {
			mark := LayoutPick_getMark(pk)

			if strings.Contains(userPrompt, mark) {
				line := picks[pk].Line
				if line > 0 { //base layout already above(func Build)
					lines[line-1] += fmt.Sprintf("//{\"mark\":\"%d\"}", line)
				}
			}
		}
	}

	return lines, nil
}

func _AssistantChat_buildPromptFile(fileName string, userPrompt string, picks []LayoutPick, out *strings.Builder) error {

	lines, err := _AssistantChat_addMarksToCode(fileName, userPrompt, picks)
	if err != nil {
		return err
	}

	out.WriteString(fmt.Sprintf("file: %s\n", fileName))
	out.WriteString("```go\n" + strings.Join(lines, "\n") + "```\n\n")
	return nil
}

func (st *AssistantChat) SetVoice(js []byte, voiceStart_sec float64) {
	type VerboseJsonWord struct {
		Start, End float64
		Word       string
	}
	type VerboseJsonSegment struct {
		Start, End float64
		Text       string
		Words      []VerboseJsonWord
	}
	type VerboseJson struct {
		Segments []VerboseJsonSegment
	}

	var verb VerboseJson
	err := json.Unmarshal(js, &verb)
	if err != nil {
		return
	}

	// jump over older picks
	pick_i := 0
	for _, pk := range st.Picks {
		if pk.Time_sec < voiceStart_sec {
			pick_i++
		}
	}
	// build prompt
	prompt := ""
	for _, seg := range verb.Segments {
		for _, w := range seg.Words {
			for pick_i < len(st.Picks) && st.Picks[pick_i].Time_sec < (voiceStart_sec+w.Start) { //for(!)
				prompt += LayoutPick_getMark(pick_i)
				pick_i++
			}
			prompt += w.Word
		}
	}

	st.Prompt = prompt
}

func _Assistant_RemoveFormatingRGBA(str string) string {
	str = strings.ReplaceAll(str, "</rgba>", "")

	for {
		st := strings.Index(str, "<rgba")
		if st < 0 {
			break
		}
		en := strings.IndexByte(str[st:], '>')
		if en >= 0 {
			str = str[:st] + str[st+en+1:]
		}
	}

	return str
}

func (ast *AssistantChat) findPickOrAdd(item LayoutPick) bool {
	for _, it := range ast.Picks {
		if it.Cmp(&item) {
			return true //found
		}
	}

	//add
	ast.Picks = append(ast.Picks, item)
	ast.Prompt += LayoutPick_getMark(len(ast.Picks) - 1)

	return false //new
}

func Assistant_recomputeColors(prompt string, picks []LayoutPick) string {
	var hsl HSL
	hsl.S = 0.8
	hsl.L = 0.5

	n := len(picks)
	hAdd := 360 / (n + 1)
	for pki := range picks {

		if strings.Contains(prompt, LayoutPick_getMark(pki)) {
			picks[pki].Cd = hsl.ToRGB()
			hsl.H += hAdd
		} else {
			picks[pki].Cd.A = 0 //invisible
		}

	}

	prompt = _Assistant_RemoveFormatingRGBA(prompt)
	for pki := range picks {
		mark := LayoutPick_getMark(pki)

		start := len(prompt)
		for start >= 0 {
			start = strings.LastIndex(prompt[:start], mark)
			if start >= 0 {
				end := start + len(mark)

				cd := picks[pki].Cd
				midStr := fmt.Sprintf("<rgba%d,%d,%d,%d>", cd.R, cd.G, cd.B, cd.A) + prompt[start:end] + "</rgba>"
				prompt = prompt[:start] + midStr + prompt[end:]
				//pos = start + len(midStr)
			}
		}
	}

	return prompt
}

type HSL struct {
	H int     //0-360
	S float64 //0-1
	L float64 //0-1
}

func _hueToRGB(v1, v2, vH float64) float64 {
	if vH < 0 {
		vH++
	}
	if vH > 1 {
		vH--
	}

	if (6 * vH) < 1 {
		return v1 + (v2-v1)*6*vH
	} else if (2 * vH) < 1 {
		return v2
	} else if (3 * vH) < 2 {
		return v1 + (v2-v1)*((2.0/3)-vH)*6
	}

	return v1
}

func (hsl HSL) ToRGB() color.RGBA {
	cd := color.RGBA{A: 255}

	if hsl.S == 0 {
		ll := hsl.L * 255
		if ll < 0 {
			ll = 0
		}
		if ll > 255 {
			ll = 255
		}
		cd.R = uint8(ll)
		cd.G = uint8(ll)
		cd.B = uint8(ll)
	} else {
		var v2 float64
		if hsl.L < 0.5 {
			v2 = (hsl.L * (1 + hsl.S))
		} else {
			v2 = ((hsl.L + hsl.S) - (hsl.L * hsl.S))
		}
		v1 := 2*hsl.L - v2

		hue := float64(hsl.H) / 360
		cd.R = uint8(255 * _hueToRGB(v1, v2, hue+(1.0/3)))
		cd.G = uint8(255 * _hueToRGB(v1, v2, hue))
		cd.B = uint8(255 * _hueToRGB(v1, v2, hue-(1.0/3)))
	}

	return cd
}

func _Assistant_GetGridBuildFuncPosition(layout *Layout) (string, int) {
	Caller_file := layout.Caller_file
	Caller_line := layout.Caller_line

	//'best_layout' is Caller position of Add<name>(). But we need position of Build()!
	if layout.Name != "_layout" {
		path := filepath.Join("widgets", layout.Name+".go")
		fileCode, err := os.ReadFile(path)
		if err == nil {
			build_pos, _, _, _, _, _, _, err := _Assistant_findBuildFunction(path, string(fileCode), layout.Name)
			if err == nil && build_pos >= 0 {
				//rewrite Caller
				Caller_file = layout.Name + ".go"
				Caller_line = strings.Count(string(fileCode[:build_pos]), "\n") + 1
			}
		} else {
			fmt.Println("File not found", err)
		}
	}

	return Caller_file, Caller_line
}

type CompileWidgetFile struct {
	Name   string
	IsFile bool

	Build bool
	Input bool
	Draw  bool

	Structs   []string
	AddFuncs  []string
	FileFuncs []string
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

			build_pos, input_pos, draw_pos, isFile, structs, addFuncs, fileFuncs, err := _Assistant_findBuildFunction(file.Name(), string(fileCode), stName)
			if err != nil {
				return nil, err
			}
			if build_pos >= 0 || input_pos >= 0 || draw_pos >= 0 { //is widget
				widgets = append(widgets, CompileWidgetFile{Name: stName, IsFile: isFile, Build: build_pos >= 0, Input: input_pos >= 0, Draw: draw_pos >= 0, Structs: structs, AddFuncs: addFuncs, FileFuncs: fileFuncs})
			}
		}
	}

	return widgets, nil
}

func _Assistant_findBuildFunction(ghostPath string, code string, stName string) (int, int, int, bool, []string, []string, []string, error) {

	node, err := parser.ParseFile(token.NewFileSet(), ghostPath, code, parser.ParseComments)
	if err != nil {
		return -1, -1, -1, false, nil, nil, nil, err
	}

	build_pos := -1
	input_pos := -1
	draw_pos := -1
	var structs []string
	var addFuncs []string
	var fileFuncs []string
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {

		case *ast.GenDecl:
			if x.Tok == token.TYPE {
				//Structs
				structs = append(structs, string(code[x.Pos()-1:x.End()-1]))
			}

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

			//AddFuncs
			if tp == "*Layout" && strings.HasPrefix(x.Name.Name, "Add") {
				addFuncs = append(addFuncs, string(code[x.Pos()-1:x.Body.Lbrace-1]))
			}
			//FileFuncs
			if tp == "" && strings.HasPrefix(x.Name.Name, "NewFile_") {
				fileFuncs = append(fileFuncs, string(code[x.Pos()-1:x.Body.Lbrace-1]))
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

	return build_pos, input_pos, draw_pos, isFile, structs, addFuncs, fileFuncs, nil
}

func _build_sdk_layout() (strings.Builder, error) {

	var str strings.Builder

	code, err := os.ReadFile("widgets/sdk_layout.go")
	if err != nil {
		return str, err
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", code, parser.ParseComments)
	if err != nil {
		return str, err
	}

	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			if x.Tok == token.TYPE {
				//struct
				str.WriteString("\n")
				str.WriteString(string(code[x.Pos()-1 : x.End()-1]))
				str.WriteString("\n")
			}
		case *ast.FuncDecl:
			if len(x.Name.Name) > 0 && x.Name.Name[:1] != strings.ToLower(x.Name.Name[:1]) { //must start with upper letter
				str.WriteString(string(code[x.Pos()-1 : x.Body.Lbrace-1])) //only header
				str.WriteString("\n")
			}
		}
		return true
	})

	return str, nil
}
