package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type AssistantChat struct {
	Prompt  string
	Picks   []LayoutPick
	AppName string

	Model Models

	Stats map[string]float64 //seconds per single (input) byte
}

func (layout *Layout) AddAssistant(x, y, w, h int, props *AssistantChat) *AssistantChat {
	layout._createDiv(x, y, w, h, "Assistant", props.Build, nil, nil)
	return props
}

var g_AssistantChat *AssistantChat

func OpenFile_AssistantChat() *AssistantChat {
	if g_AssistantChat == nil {
		g_AssistantChat = &AssistantChat{Model: Models{Model: "grok-2-1212"}}
		_read_file("AssistantChat-AssistantChat", g_AssistantChat)
	}
	return g_AssistantChat
}

func (st *AssistantChat) Build(layout *Layout) {
	//...
}

func (ast *AssistantChat) executePrompt(jobName string, msgs []OpenAI_completion_msg, done func(string)) {

	st := (float64(time.Now().UnixMilli()) / 1000)

	statKey := ast.Model.GetService() + "_" + ast.Model.Model + "_" + jobName + "_" + ast.AppName
	var numInputBytes int
	{
		js, _ := json.Marshal(msgs) //inputs
		numInputBytes = len(js)
	}

	if ast.Stats == nil {
		ast.Stats = make(map[string]float64)
	}

	done2 := func(out string) {
		dt := ((float64(time.Now().UnixMilli()) / 1000) - st)
		ast.Stats[statKey] = dt / float64(numInputBytes)
		//fmt.Println("+++save", statKey, ast.Stats[statKey])
		done(out) //call
	}

	var job *Job
	switch ast.Model.GetService() {
	case "xai":
		chat := NewGlobal_Xai_completion(jobName)
		chat.Properties.Model = ast.Model.Model
		chat.Properties.Messages = msgs
		chat.Properties.Response_format = &OpenAI_completion_format{Type: "json_object"}
		chat.done = done2
		job = chat.Start()

	case "openai":
		chat := NewGlobal_OpenAI_completion(jobName)
		chat.Properties.Model = ast.Model.Model
		chat.Properties.Messages = msgs
		chat.Properties.Response_format = &OpenAI_completion_format{Type: "json_object"}
		chat.done = done2
		job = chat.Start()

	case "anthropic":
		chat := NewGlobal_Anthropic_completion(jobName)
		chat.Properties.Model = ast.Model.Model
		chat.Properties.Messages = msgs
		//chat.Properties.Response_format = &OpenAI_completion_format{Type: "json_object"}
		chat.done = done2
		job = chat.Start()

	case "groq":
		chat := NewGlobal_Groq_completion(jobName)
		chat.Properties.Model = ast.Model.Model
		chat.Properties.Messages = msgs
		chat.Properties.Response_format = &OpenAI_completion_format{Type: "json_object"}
		chat.done = done2
		job = chat.Start()
	}

	SecPerByte, found := ast.Stats[statKey]
	if found {
		//fmt.Println("+++estimate", SecPerByte*float64(numInputBytes))
		job.SetEstimate(st + SecPerByte*float64(numInputBytes))
	} else {
		job.SetEstimate(st + 20)
	}

}

func (ast *AssistantChat) Send() {
	ast.Prompt = strings.TrimSpace(ast.Prompt)

	if len(ast.Prompt) < 5 {
		Layout_WriteError(errors.New("prompt is too short"))
		return
	}

	SystemPrompt := "You are an AI programming assistant, who enjoys precision and carefully follows the user's requirements. You write code only in Golang."
	UserPrompt, err := ast.prompt_1()
	if err != nil {
		Layout_WriteError(fmt.Errorf("prompt_1() %w", err))
		return
	}
	fmt.Println("prompt_1_input:", UserPrompt)

	var msgs []OpenAI_completion_msg
	{
		msg_sys := OpenAI_completion_msg{Role: "system", Content: SystemPrompt}
		msgs = append(msgs, msg_sys)
		msg_usr := OpenAI_completion_msg{Role: "user", Content: UserPrompt}
		msgs = append(msgs, msg_usr)
	}

	done1 := func(Out string) {
		if Out == "" {
			Layout_WriteError(fmt.Errorf("Chat1 output is empty"))
			return
		}
		Out = FixOutput(Out)
		fmt.Println("prompt_1_output:", Out)

		UserPrompt, jobNameEx, err := ast.prompt_2([]byte(Out))
		if err != nil {
			Layout_WriteError(fmt.Errorf("Prompt_2() %w", err))
			return
		}
		fmt.Println("prompt_2_input:", UserPrompt)

		var msgs []OpenAI_completion_msg
		{
			msg_sys := OpenAI_completion_msg{Role: "system", Content: SystemPrompt}
			msgs = append(msgs, msg_sys)
			msg_usr := OpenAI_completion_msg{Role: "user", Content: UserPrompt}
			msgs = append(msgs, msg_usr)
		}

		done2 := func(Out string) {
			fmt.Println("---done2 starts")
			if Out == "" {
				Layout_WriteError(fmt.Errorf("Chat2 output is empty"))
				return
			}
			Out = FixOutput(Out)
			fmt.Println("prompt_2_output:", Out)

			//project new code to file(s)
			type OutputFile struct {
				Name    string
				Content string
			}
			type OutputFiles struct {
				Files []OutputFile
			}

			var outs OutputFiles
			err := json.Unmarshal([]byte(Out), &outs)
			if err != nil {
				var out OutputFile
				err := json.Unmarshal([]byte(Out), &out)
				if err != nil {
					Layout_WriteError(err)
					return
				}
				outs.Files = append(outs.Files, out)
			}

			fmt.Println("---writing files")

			for _, f := range outs.Files {

				//remove marks
				{
					lines := strings.Split(string(f.Content), "\n")
					mark := "//{MARK"
					for i := len(lines) - 1; i >= 0; i-- {
						d := strings.Index(lines[i], mark)
						if d >= 0 {
							lines[i] = lines[i][:d] //cut
						}
					}
					f.Content = strings.Join(lines, "\n")
				}

				//write file
				fmt.Println("Writing", f.Name)
				err = os.WriteFile("widgets/"+f.Name, []byte(f.Content), 0644)
				if err != nil {
					Layout_WriteError(err)
					return
				}
			}

			ast.reset()

			Layout_Recompile()
		}
		ast.executePrompt("AssistantChat2-"+jobNameEx, msgs, done2)
	}
	ast.executePrompt("AssistantChat1", msgs, done1)
}

func (st *AssistantChat) reset() {
	st.Prompt = ""
	st.Picks = nil
	Layout_ResetBrushes()
}

func FixOutput(str string) string {
	str, _ = strings.CutPrefix(str, "```json")
	str, _ = strings.CutPrefix(str, "```")

	str, _ = strings.CutSuffix(str, "```")
	return str
}

func (st *AssistantChat) convertUserPromptToMarks(userPrompt string) string {
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

	for _, pk := range st.Picks {
		src := _AssistantChat_getPromptMarkLabel(pk)
		dst := _AssistantChat_getPromptMarkOrig(pk)
		userPrompt = strings.ReplaceAll(userPrompt, src, dst)
	}

	return userPrompt
}

const base_promp = `Here are the <description> : <json_response>:
- Open/Show widget: {"open": true, "name": <widget_name>}
- Create new widget: {"create": true, "name": <widget_name>}
- Modify widget code(change layout/grid(column, row) or feature): {"code": true}
- Change data(not code): {"data": true}
- General question about code: {"q_code": true}
- General question about data: {"q_data": true}
- General question about anything else: {"q_other": true}

Note: widget = app,layout,page,program.

This is the prompt from user:
%s

%s

Your job is to match the user prompt to <description> and output <json_response>. Please respond in the JSON format for example: {"open": true, "name": "App"}.
`

const prompt_marks_explain = `Note: A prompt can have special marks(for ex.: {MARK 1; x:3, y:2, w:1, h:1}).
Mark refers to line in code(mark is in comment for example: subDiv := layout.AddLayout(2, 3, 1, 1)	//{MARK 2})
Mark also has x,y,w,h which represent layout grid position which was selected by user. Grid position is usefull for creating/editing layout components, for example prompt "Delete {MARK 2; x:0, y:5, w:2, h:1}" must delete line with AddText in bellow code:
subDiv := layout.AddLayout(0, 1, 3, 5)	//{MARK 2}
subDiv.AddText(0, 5, 2, 1, "Hello")
`

const code_marks_explain = `Note: If user prompt is about adding or deleting new column or row, you have to change(rewrite code) coordinates for layout.Add...(x,y,w,h) calls!`
const code_attrs_explain = `Note: If user prompt contains creating new attribute/variable/storage you have to add new attribute into widget's struct.`

func (ast *AssistantChat) prompt_1() (string, error) {

	userPrompt := ast.convertUserPromptToMarks(ast.Prompt)

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
	str.WriteString(fmt.Sprintf(base_promp, userPrompt, prompt_marks_explain))

	return str.String(), nil
}

func (ast *AssistantChat) prompt_2(output_1 []byte) (string, string, error) {

	userPrompt := ast.convertUserPromptToMarks(ast.Prompt)

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
		var outs []Out
		err := json.Unmarshal(output_1, &outs)
		if err != nil {
			return "", "", err
		}
		if len(outs) > 0 {
			out = outs[0]
		}
	}

	info, err := Compile_get_files_info()
	if err != nil {
		return "", "", err
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
		return "", "", err
	}

	jobNameEx := ""
	var str strings.Builder
	if out.Open {
		jobNameEx = "open"

		str.WriteString("\n\n// Here are the functions to show widgets:\n")
		str.WriteString("```go\n" + addFns.String() + "\n```")

		str.WriteString("\n\n// Here are the functions to access data(configurations) from disk:\n")
		str.WriteString("```go\n" + fileFns.String() + "\n```")

		str.WriteString("\n\n// Here are the files:\n")
		edited_file := "ShowApp"
		ast.addReadFile(edited_file, &str)

		str.WriteString("Note: Build() has layout.Add...(), which represent opened(shown) widget. Rewrite this line(same coord: 0,0,1,1) to show widget from the prompt.\n\n")

		str.WriteString("This is the prompt from the user:\n")
		str.WriteString(_AssistantChat_buildUserPrompt(userPrompt) + "\n")
		str.WriteString("\n" + prompt_marks_explain + "\n")

		str.WriteString(fmt.Sprintf("\nYour job is implement the request(user prompt) by editting file %s.\n\n", edited_file))

		str.WriteString(`Don't add main() function to the code. Please respond in the JSON format {"files": [{"name": <file_name.go>, "content": <code>}]}`)

	} else if out.Create {
		jobNameEx = "create"

		if len(out.Name) > 0 {
			out.Name = strings.ToUpper(out.Name[:1]) + out.Name[1:] //1st letter must be upper
		}

		//str.WriteString("\n// Here are the widget's structures:\n")
		//str.WriteString("```go\n" + structs.String() + "\n```")

		str.WriteString("\n\n// Here are the functions to add Widgets:\n")
		str.WriteString("```go\n" + addFns.String() + "\n```")

		str.WriteString("\n\n// Here are the functions to access data(configurations) from disk:\n")
		str.WriteString("```go\n" + fileFns.String() + "\n```")

		str.WriteString("\n\n// Here are the files:\n")
		new_file := strings.TrimSuffix(out.Name, filepath.Ext(out.Name))
		edited_file := "ShowApp"
		{
			//example
			str.WriteString(fmt.Sprintf("file: %s.go\n", new_file))
			new_code := fmt.Sprintf(`package main
type %s struct {
}
func (layout *Layout) Add%s(x, y, w, h int, props *%s) *%s {
	layout._createDiv(x, y, w, h, "%s", props.Build, nil, nil)
	return props
}
var g_%s *%s
func OpenFile_%s() *%s {
	if g_%s == nil {
		g_%s = &%s{}
		_read_file("%s-%s", g_%s)
	}
	return g_%s
}
func (st *%s) Build(layout *Layout) {
}`, new_file, new_file, new_file, new_file, new_file,
				new_file, new_file, new_file, new_file, new_file, new_file, new_file, new_file, new_file, new_file, new_file,
				new_file)
			//add code
			ast.addCode(new_file, new_code, &str)
		}
		{
			ast.addReadFile(edited_file, &str)
		}

		str.WriteString("This is the prompt from the user:\n")
		str.WriteString(_AssistantChat_buildUserPrompt(userPrompt) + "\n")
		str.WriteString("\n" + prompt_marks_explain + "\n")

		str.WriteString(fmt.Sprintf("\nYour job is implement the request(user prompt) by editting code in files %s.go and %s.go(grid position[0, 0, 1, 1], replace old widget).\n\n", new_file, edited_file))

		str.WriteString(`Don't add main() function to the code. Please respond in the JSON format {"files": [{"name": <file_name.go>, "content": <code>}]}`)

	} else if out.Code {
		jobNameEx = "code"

		str.WriteString("\n// Here are the layout structures and functions:\n")
		str.WriteString("```go\n" + layoutStFns.String() + "\n```")

		str.WriteString("\n// Here are the widget's structures:\n")
		str.WriteString("```go\n" + structs.String() + "\n```")

		str.WriteString("\n\n// Here are the functions to add Widgets:\n")
		str.WriteString("```go\n" + addFns.String() + "\n```")

		str.WriteString("\n\n// Here are the functions to access data(configurations) from disk:\n")
		str.WriteString("```go\n" + fileFns.String() + "\n```")

		str.WriteString("\n\n// Here are the files:\n")
		ast.addPromptCode(&str)

		str.WriteString("This is the prompt from the user:\n")
		str.WriteString(_AssistantChat_buildUserPrompt(userPrompt) + "\n")
		str.WriteString("\n" + prompt_marks_explain + "\n")
		str.WriteString("\n" + code_marks_explain + "\n")
		str.WriteString("\n" + code_attrs_explain + "\n")

		str.WriteString("\nYour job is implement the request(user prompt) by editting code in files. Usually you edit only one file.\n\n")

		str.WriteString(`Don't add main() function to the code. Please respond in the JSON format {"files": [{"name": <file_name.go>, "content": <code>}]}`)

	} else if out.Data {
		jobNameEx = "data"

		str.WriteString("\n// Here are the widget's structures:\n")
		str.WriteString("```go\n" + structs.String() + "\n```")

		str.WriteString("\n\n// Here are the functions to access data(configurations) from disk:\n")
		str.WriteString("```go\n" + fileFns.String() + "\n```")

		str.WriteString("\n\n// Here are the files:\n")
		ast.addPromptCode(&str)

		str.WriteString("This is the prompt from the user:\n")
		str.WriteString(_AssistantChat_buildUserPrompt(userPrompt) + "\n")
		str.WriteString("\n" + prompt_marks_explain + "\n")

		str.WriteString("\nYour job is implement the request(user prompt) by editting code in file sdk_change.go. Use functions OpenFile_?() to read/write data.\n\n")

		str.WriteString(`Don't add main() function to the code. Please respond in the JSON format {"files": [{"name": <file_name.go>, "content": <code>}]}`)
	} else if out.Q_code {
		jobNameEx = "q_code"
		//...
		//system promp říká že je programátor, který píše kód! ...
	} else if out.Q_data {
		jobNameEx = "q_data"
		//...
	} else if out.Q_other {
		jobNameEx = "q_other"
		//...
	} else {
		return "", "", fmt.Errorf("unrecognized user prompt")
	}

	return str.String(), jobNameEx, nil
}

func _AssistantChat_buildUserPrompt(userPrompt string) string {
	prompt := _Assistant_RemoveFormatingRGBA(userPrompt)

	prompt = strings.ReplaceAll(prompt, "\n", " ")
	prompt = strings.TrimSpace(prompt)

	return prompt
}

func (ast *AssistantChat) addCode(appName string, code string, out *strings.Builder) {
	out.WriteString(fmt.Sprintf("file: %s.go\n", appName))
	out.WriteString("```go\n" + code + "```\n\n")
}
func (ast *AssistantChat) addReadFile(appName string, out *strings.Builder) error {
	fileName := appName + ".go"
	path := filepath.Join("widgets", fileName)
	fl, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	ast.addCode(appName, string(fl), out)
	return nil
}
func (ast *AssistantChat) addPromptCode(out *strings.Builder) error {
	code, err := ast.GetCodeWithMarks()
	if err != nil {
		return err
	}

	ast.addCode(ast.AppName, code, out)
	return nil
}

func (ast *AssistantChat) GetCodeWithMarks() (string, error) {

	fileName := ast.AppName + ".go"
	path := filepath.Join("widgets", fileName)
	fl, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(fl), "\n")

	//remove old marks
	mark := "{MARK"
	for i := len(lines) - 1; i >= 0; i-- {
		d := strings.Index(lines[i], mark)
		if d >= 0 {
			lines[i] = lines[i][:d] //cut
		}
	}

	//add marks as comments
	for i, pk := range ast.Picks {

		//every mark can be only once
		found_line := false
		for ii := 0; ii < i; ii++ {
			if ast.Picks[ii].Line == pk.Line {
				found_line = true
				break
			}
		}

		if !found_line {
			lines[pk.Line-1] += "\t//" + _AssistantChat_GetCodeMark(pk)
		}
	}

	return strings.Join(lines, "\n"), nil
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

	//clean(remove for example [BLANK_AUDIO])
	for i := len(verb.Segments) - 1; i >= 0; i-- {
		word := strings.Trim(verb.Segments[i].Text, " \t")
		if len(word) > 0 && word[0] == '[' && word[len(word)-1] == ']' {
			fmt.Println("removed word", word)
			verb.Segments = append(verb.Segments[:i], verb.Segments[i+1:]...) //remove
		}
	}

	// jump over older picks
	pick_i := 0
	for _, pk := range st.Picks {
		if pk.time_sec < voiceStart_sec {
			pick_i++
		}
	}
	// build prompt
	prompt := ""
	for _, seg := range verb.Segments {
		for _, w := range seg.Words {
			for pick_i < len(st.Picks) && st.Picks[pick_i].time_sec < (voiceStart_sec+w.Start) { //for(!)
				prompt += _AssistantChat_getPromptMarkLabel(st.Picks[pick_i])
				pick_i++
			}
			prompt += w.Word
		}
	}
	//add rest
	for pick_i < len(st.Picks) {
		prompt += _AssistantChat_getPromptMarkLabel(st.Picks[pick_i])
		pick_i++
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
	//find
	for _, it := range ast.Picks {
		if it.Cmp(&item) {
			ast.Prompt += _AssistantChat_getPromptMarkLabel(it)
			return true //found
		}
	}

	//add
	ast.Picks = append(ast.Picks, item)
	ast.Prompt += _AssistantChat_getPromptMarkLabel(item)

	return false //new
}

func _AssistantChat_GetCodeMark(item LayoutPick) string {
	return fmt.Sprintf("{MARK %d}", item.Line)
}

func _AssistantChat_getPromptMarkOrig(item LayoutPick) string {
	return fmt.Sprintf("{MARK %d: x:%d,y:%d,w:%d,h:%d}", item.Line, item.X, item.Y, item.W, item.H)
}
func _AssistantChat_getPromptMarkLabel(item LayoutPick) string {
	return fmt.Sprintf("{%s}", item.Label)
}

func (ast *AssistantChat) Assistant_recomputePromptColors() {
	prompt := _Assistant_RemoveFormatingRGBA(ast.Prompt)

	for _, pk := range ast.Picks {
		mark := _AssistantChat_getPromptMarkLabel(pk)
		start := len(prompt)
		for start >= 0 {
			start = strings.LastIndex(prompt[:start], mark)
			if start >= 0 {
				end := start + len(mark)

				newStr := fmt.Sprintf("<rgba%d,%d,%d,%d>", pk.Cd.R, pk.Cd.G, pk.Cd.B, pk.Cd.A) + mark + "</rgba>"
				prompt = prompt[:start] + newStr + prompt[end:]
			}
		}
	}

	ast.Prompt = prompt
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
			if tp == "" && strings.HasPrefix(x.Name.Name, "OpenFile_") {
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
