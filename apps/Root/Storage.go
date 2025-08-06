package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type RootChat struct {
	FileName string
	Label    string
	Pinned   bool
}

func (app *RootApp) GetFolderPath() string {
	return filepath.Join("..", app.Name)
}

type RootDev struct {
	Enable bool

	PromptsHistory []string

	ShowSide bool
	SideFile string //Name.go
	MainMode string //"prompts", "secrets"
	SideMode string //"schema", "code", "msg"

	SideFile_version int
}

type RootApp struct {
	Name            string
	Chats           []RootChat
	Selected_chat_i int
	ShowSide        bool

	Dev RootDev
}

// Root
type Root struct {
	ShowSettings bool
	Memory       string

	Autosend float64

	Apps           []*RootApp
	Selected_app_i int

	Last_log_time float64
}

func NewRoot(file string) (*Root, error) {
	st := &Root{ShowSettings: true}

	return LoadFile(file, "Root", "json", st, true)
}

func (root *Root) IsAppExist(name string) bool {
	_, err := os.Stat(filepath.Join("..", name))
	return err == nil && !os.IsNotExist(err)
}

func (root *Root) refreshApps() (*RootApp, error) {
	fls, err := os.ReadDir("..")
	if err != nil {
		return nil, err
	}
	//add new apps
	for _, fl := range fls {
		if !fl.IsDir() || fl.Name() == "Root" {
			continue
		}

		found := false
		for _, app := range root.Apps {
			if app.Name == fl.Name() {
				found = true
				break
			}
		}
		if !found {
			root.Apps = append(root.Apps, &RootApp{Name: fl.Name()})
		}
	}
	//remove deleted app
	for i := len(root.Apps) - 1; i >= 0; i-- {
		found := false
		for _, fl := range fls {
			if !fl.IsDir() || fl.Name() == "Root" {
				continue
			}

			if fl.Name() == root.Apps[i].Name {
				found = true
				break
			}
		}
		if !found {
			root.Apps = slices.Delete(root.Apps, i, i+1)
		}
	}

	//check icons
	for _, app := range root.Apps {
		icon := filepath.Join("..", app.Name, "icon")
		if !_isFileExists(icon) {
			_copyFile(icon, filepath.Join("..", "..", "resources", "think.png")) //copy default icon
		}
	}

	//check select in range
	if root.Selected_app_i >= 0 {
		if root.Selected_app_i >= len(root.Apps) {
			root.Selected_app_i = len(root.Apps) - 1
		}
	}
	if root.Selected_app_i >= 0 {
		return root.Apps[root.Selected_app_i], nil
	}

	return nil, nil
}

func (app *RootApp) refreshChats() (*Chat, string, error) {

	//check range
	if app.Selected_chat_i >= 0 && app.Selected_chat_i >= len(app.Chats) {
		app.Selected_chat_i = len(app.Chats) - 1
	}

	chats_folder := filepath.Join("..", app.Name, "Chats")
	if _, err := os.Stat(chats_folder); os.IsNotExist(err) {
		//no chat folder
		app.Chats = nil
		return nil, "", nil //ok
	}

	fls, err := os.ReadDir(chats_folder)
	if err != nil {
		return nil, "", nil //maybe no chat
	}
	//add new chats
	for _, fl := range fls {
		if fl.IsDir() {
			continue
		}

		found := false
		for _, chat := range app.Chats {
			if chat.FileName == fl.Name() {
				found = true
				break
			}
		}
		if !found {
			app.Chats = append(app.Chats, RootChat{FileName: fl.Name()})
		}
	}
	//remove deleted chats
	for i := len(app.Chats) - 1; i >= 0; i-- {
		found := false
		for _, fl := range fls {
			if fl.IsDir() {
				continue
			}

			if fl.Name() == app.Chats[i].FileName {
				found = true
				break
			}
		}
		if !found {
			app.Chats = slices.Delete(app.Chats, i, i+1)
		}
	}

	//all pinned chats must be at the beginning
	{
		num_pinned := app.NumPins()
		for i := num_pinned; i < len(app.Chats); i++ {
			chat := app.Chats[i]
			if chat.Pinned {
				//move
				app.Chats = slices.Delete(app.Chats, i, i+1)
				app.Chats = slices.Insert(app.Chats, num_pinned, chat)
				num_pinned++
			}
		}
	}

	//check range again
	if app.Selected_chat_i >= 0 && app.Selected_chat_i >= len(app.Chats) {
		app.Selected_chat_i = len(app.Chats) - 1
	}

	//update and return
	if app.Selected_chat_i >= 0 {
		fileName := app.Chats[app.Selected_chat_i].FileName
		sourceChat, err := NewChat(filepath.Join("..", app.Name, "Chats", fileName))
		if err != nil {
			return nil, "", err
		}

		/*if sourceChat != nil {
			//reload
			app.Chats[app.Selected_chat_i].Label = sourceChat.Label
		}*/

		return sourceChat, fileName, nil
	}

	return nil, "", nil
}

func (usage *LLMMsgUsage) GetSpeed() float64 {
	toks := usage.Completion_tokens + usage.Reasoning_tokens
	if usage.DTime == 0 {
		return 0
	}
	return float64(toks) / usage.DTime
}

type OpenAI_completion_msg_Content_Image_url struct {
	Detail string `json:"detail,omitempty"` //"low", "high", "auto"
	Url    string `json:"url,omitempty"`    //"data:image/jpeg;base64,<base64_image_string>"
}
type OpenAI_completion_msg_Content struct {
	Type      string                                   `json:"type"` //"image_url", "text"
	Text      string                                   `json:"text,omitempty"`
	Image_url *OpenAI_completion_msg_Content_Image_url `json:"image_url,omitempty"`
}

type OpenAI_completion_msgContent struct {
	Role    string                          `json:"role"` //"system", "user", "assistant", "tool"
	Content []OpenAI_completion_msg_Content `json:"content"`
}

type OpenAI_completion_msgCalls struct {
	Role       string                                   `json:"role"` //"system", "user", "assistant", "tool"
	Content    string                                   `json:"content"`
	Tool_calls []OpenAI_completion_msg_Content_ToolCall `json:"tool_calls,omitempty"`
}

type OpenAI_completion_msg_Content_ToolCall_Function struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}
type OpenAI_completion_msg_Content_ToolCall struct {
	Id       string                                          `json:"id,omitempty"`
	Type     string                                          `json:"type,omitempty"`
	Function OpenAI_completion_msg_Content_ToolCall_Function `json:"function,omitempty"`
}

type OpenAI_completion_msgResult struct {
	Role         string `json:"role"` //"system", "user", "assistant", "tool"
	Content      string `json:"content"`
	Tool_call_id string `json:"tool_call_id,omitempty"`
	Name         string `json:"name,omitempty"` //Tool name: Mistral wants this
}

type OpenAI_content struct {
	Msg    *OpenAI_completion_msgContent `json:",omitempty"`
	Calls  *OpenAI_completion_msgCalls   `json:",omitempty"`
	Result *OpenAI_completion_msgResult  `json:",omitempty"`
}

type ChatMsg struct {
	Seed int

	Content OpenAI_content

	ReasoningSize int //Final text is after
	ShowReasoning bool

	UI_func     string
	UI_paramsJs string

	Usage LLMMsgUsage

	ShowParameters bool

	Stream bool
}
type ChatMsgs struct {
	Messages []*ChatMsg
}

type LayoutPromptColor struct {
	Label string
	Cd    color.RGBA
}

func (cd *LayoutPromptColor) GetLabel() string {
	return fmt.Sprintf("<rgba%d,%d,%d,%d>{%s}</rgba>", cd.Cd.R, cd.Cd.G, cd.Cd.B, cd.Cd.A, cd.Label)
}

var g_prompt_colors = []LayoutPromptColor{
	{Label: "red", Cd: color.RGBA{255, 0, 0, 255}},
	{Label: "green", Cd: color.RGBA{0, 255, 0, 255}},
	{Label: "blue", Cd: color.RGBA{0, 0, 255, 255}},

	{Label: "orange", Cd: color.RGBA{255, 165, 0, 255}},
	{Label: "pink", Cd: color.RGBA{255, 192, 203, 255}},
	{Label: "yellow", Cd: color.RGBA{200, 200, 0, 255}},

	{Label: "aqua", Cd: color.RGBA{0, 255, 255, 255}},
	{Label: "fuchsia", Cd: color.RGBA{255, 0, 255, 255}},
	{Label: "olive", Cd: color.RGBA{128, 128, 0, 255}},
	{Label: "teal", Cd: color.RGBA{0, 128, 128, 255}},
	{Label: "purple", Cd: color.RGBA{128, 0, 128, 255}},
	{Label: "navy", Cd: color.RGBA{0, 0, 128, 255}},
	{Label: "marron", Cd: color.RGBA{128, 0, 0, 255}},
	{Label: "lime", Cd: color.RGBA{0, 255, 0, 255}},
	{Label: "brown", Cd: color.RGBA{165, 42, 42, 255}},
	{Label: "grey", Cd: color.RGBA{128, 128, 128, 255}},
}

type LayoutPick struct {
	Cd     LayoutPromptColor
	LLMTip string
	Points []UIPaintBrushPoint
	Dash_i int
}

type ChatInput struct {
	Text string

	Files          []string
	FilePickerPath string

	Multilined bool

	Picks    []LayoutPick
	Text_mic string //copy of .Text before mic starts recording
}

// Chat has label, messages.
type Chat struct {
	file string

	Input ChatInput

	PresetSystemPrompt string
	Messages           ChatMsgs

	Selected_user_msg int

	Error string

	TempMessages ChatMsgs

	Sources []string
}

func NewChat(file string) (*Chat, error) {
	st := &Chat{ /*Label: "chat"*/ }
	return LoadFile(file, "Chat", "json", st, true)
}

func (st *Chat) GetChatID() string {
	return "chat_" + st.file
}

func (st *Chat) GetNumUserMessages() int {
	n := 0
	for _, msg := range st.Messages.Messages {
		if msg.Content.Msg != nil {
			n++
		}
	}
	for _, msg := range st.TempMessages.Messages {
		if msg.Content.Msg != nil {
			n++
		}
	}
	return n
}
func (st *Chat) CutMessages(user_i int) {
	n := 0
	for i, msg := range st.Messages.Messages {
		if msg.Content.Msg != nil { //user message
			if n > user_i {
				st.Messages.Messages = st.Messages.Messages[:i] //cut
				st.TempMessages.Messages = nil                  //cut
				return
			}
			n++
		}
	}
	for i, msg := range st.TempMessages.Messages {
		if msg.Content.Msg != nil { //user message
			if n > user_i {
				st.TempMessages.Messages = st.TempMessages.Messages[:i] //cut
				return
			}
			n++
		}
	}
}

func (st *Chat) GetResponse(user_i int) (ret []*ChatMsg) {
	n := 0
	for _, msg := range st.Messages.Messages {
		if n > user_i {
			ret = append(ret, msg)
		}
		if msg.Content.Msg != nil { //user message
			if n > user_i {
				return
			}
			n++
		}
	}
	for _, msg := range st.TempMessages.Messages {
		if n > user_i {
			ret = append(ret, msg)
		}
		if msg.Content.Msg != nil { //user message
			if n > user_i {
				return
			}
			n++
		}
	}
	return
}

func (st *Chat) FindUserMessage(pos int) string {
	n := 0
	for _, msg := range st.Messages.Messages {
		if msg.Content.Msg != nil {
			if pos == n {
				for _, it := range msg.Content.Msg.Content {
					if it.Type == "text" && it.Text != "" {
						return it.Text
					}
				}
			}
			n++
		}
	}
	for _, msg := range st.TempMessages.Messages {
		if msg.Content.Msg != nil {
			if pos == n {
				for _, it := range msg.Content.Msg.Content {
					if it.Type == "text" && it.Text != "" {
						return it.Text
					}
				}
			}
			n++
		}

	}
	return ""
}

func (st *Chat) FindToolCallUserMessage(tool_call_id string) int {
	n := 0
	for _, msg := range st.Messages.Messages {
		if msg.Content.Msg != nil {
			n++
		}
		if msg.Content.Result != nil && msg.Content.Result.Tool_call_id == tool_call_id {
			return n - 1
		}
	}
	for _, msg := range st.TempMessages.Messages {
		if msg.Content.Msg != nil {
			n++
		}
		if msg.Content.Result != nil && msg.Content.Result.Tool_call_id == tool_call_id {
			return n - 1
		}
	}
	return -1
}

func (st *ChatInput) FindPick(llmTip string) int {
	if llmTip == "" {
		return -1
	}

	for i, it := range st.Picks {
		if it.LLMTip == llmTip {
			return i
		}
	}
	return -1
}

func (st *ChatInput) MergePick(in LayoutPick) {
	//find & update color
	i := st.FindPick(in.LLMTip)
	if i >= 0 {
		in.Cd = st.Picks[i].Cd
	} else {
		in.Cd = g_prompt_colors[len(st.Picks)%len(g_prompt_colors)]
	}

	//add
	st.Picks = append(st.Picks, in)

	if st.Text != "" && st.Text[len(st.Text)-1] != ' ' {
		st.Text += " " //space before new mark
	}
	st.Text += in.Cd.GetLabel() + " "
}

func (st *ChatInput) Reset() {
	st.Text = ""
	st.Files = nil
	st.Picks = nil
}

func (msg *ChatMsg) HasUI() bool {
	return msg.Content.Result != nil && msg.UI_func != ""
}

func (msgs *ChatMsgs) GetTotalPrice(st_i, en_i int) (input, inCached, output float64) {
	if en_i < 0 {
		en_i = len(msgs.Messages)
	}
	for i := st_i; i < en_i; i++ {
		input += msgs.Messages[i].Usage.Prompt_price
		inCached += msgs.Messages[i].Usage.Input_cached_price
		output += msgs.Messages[i].Usage.Completion_price + msgs.Messages[i].Usage.Reasoning_price
	}

	return
}

func (msgs *ChatMsgs) GetTotalSpeed(st_i, en_i int) float64 {
	toks := msgs.GetTotalOutputTokens(st_i, en_i)
	dt := msgs.GetTotalTime(st_i, en_i)
	if dt == 0 {
		return 0
	}
	return float64(toks) / dt

}

func (msgs *ChatMsgs) GetTotalTime(st_i, en_i int) float64 {
	dt := 0.0

	if en_i < 0 {
		en_i = len(msgs.Messages)
	}
	for i := st_i; i < en_i; i++ {
		dt += msgs.Messages[i].Usage.DTime
	}

	return dt
}

func (msgs *ChatMsgs) GetTotalInputTokens(st_i, en_i int) int {
	tokens := 0

	if en_i < 0 {
		en_i = len(msgs.Messages)
	}
	for i := st_i; i < en_i; i++ {
		tokens += msgs.Messages[i].Usage.Prompt_tokens
	}

	return tokens
}
func (msgs *ChatMsgs) GetTotalOutputTokens(st_i, en_i int) int {
	tokens := 0

	if en_i < 0 {
		en_i = len(msgs.Messages)
	}
	for i := st_i; i < en_i; i++ {
		tokens += msgs.Messages[i].Usage.Completion_tokens + msgs.Messages[i].Usage.Reasoning_tokens
	}

	return tokens
}

func (msgs *ChatMsgs) FindResultContent(call_id string) (*ChatMsg, int) {
	for i, m := range msgs.Messages {
		if m.Content.Result != nil && m.Content.Result.Tool_call_id == call_id {
			return m, i
		}
	}
	return nil, -1
}

func (call *OpenAI_completion_msg_Content_ToolCall_Function) GetArgsAsStrings() (map[string]json.RawMessage, error) {
	var attrs map[string]json.RawMessage
	err := json.Unmarshal([]byte(call.Arguments), &attrs)
	if err != nil {
		return nil, err
	}

	return attrs, nil
}

func (st *ChatInput) SetVoice(js []byte) error {
	//type VerboseJsonWord struct {
	//	Start, End float64
	//	Word       string
	//}
	type VerboseJsonSegment struct {
		Start, End float64
		Text       string
		//Words      []*VerboseJsonWord
	}
	type VerboseJson struct {
		Segments []*VerboseJsonSegment //later are projected to .Words
	}

	var verb VerboseJson
	err := json.Unmarshal(js, &verb)
	if err != nil {
		return fmt.Errorf("verb failed: %v", err)
	}

	//build prompt
	prompt := ""
	for _, seg := range verb.Segments {
		prompt += seg.Text
	}
	prompt = _ChatInput_replaceBackground(prompt)
	prompt = strings.TrimSpace(prompt)

	pick_i := 0
	words := []string{"this", "these", "here"}
	for pick_i < len(st.Picks) {
		min_p := len(prompt)
		word_n := 0
		for _, wd := range words {
			p := strings.Index(prompt, wd)
			if p >= 0 && p < min_p {
				min_p = p
				word_n = len(wd)
			}
		}
		if word_n > 0 {
			prompt = prompt[:min_p] + st.Picks[pick_i].Cd.GetLabel() + prompt[min_p+word_n:] //replace
			pick_i++

		} else {
			break
		}
	}

	st.Text = st.Text_mic + prompt

	return nil
}

func (st *ChatInput) GetFullPrompt() (string, []string) {
	prompt := _ChatInput_RemoveFormatingRGBA(st.Text)

	legend := ""
	sign := 'A'
	for _, br := range st.Picks {
		label := fmt.Sprintf("{%s}", br.Cd.Label)

		if strings.Contains(prompt, label) {
			new_label := fmt.Sprintf("{%c}", sign)
			prompt = strings.ReplaceAll(prompt, label, new_label)
			sign++

			legend += "\n" + new_label + ": " + br.LLMTip
		}
	}
	if legend != "" {
		prompt += "\n" + legend
	}

	return prompt, st.Files
}

func _ChatInput_replaceBackground(str string) string {
	str = strings.ReplaceAll(str, "[BLANK_AUDIO]", "")
	str = strings.ReplaceAll(str, "[NO_SPEECH]", "")
	str = strings.ReplaceAll(str, "[MUSIC]", "")
	str = strings.ReplaceAll(str, "[NOISE]", "")
	str = strings.ReplaceAll(str, "[LAUGHTER]", "")
	str = strings.ReplaceAll(str, "[APPLAUSE]", "")
	str = strings.ReplaceAll(str, "[UNKNOWN]", "")
	str = strings.ReplaceAll(str, "[INAUDIBLE]", "")
	return str
}

func _ChatInput_RemoveFormatingRGBA(str string) string {
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

func (chat *Chat) _sendIt(appName string, caller *ToolCaller, root *Root, continuee bool) error {
	//MsgsDiv.VScrollToTheBottom(false, caller)
	caller.SendFlushCmd()

	if !continuee && chat.Input.Text == "" {
		return nil //empty text
	}

	caller.SetMsgName(chat.GetChatID())

	chat.CutMessages(chat.Selected_user_msg)

	err := chat.complete(appName, caller, root, continuee)
	if err != nil {
		return fmt.Errorf("complete() failed: %v", err)
	}
	if !continuee {
		chat.Input.Reset()
	}

	//MsgsDiv.VScrollToTheBottom(true, caller)
	return nil
}

func (chat *Chat) complete(appName string, caller *ToolCaller, root *Root, continuee bool) error {

	//needSummary := (len(chat.Messages.Messages) == 0 /*|| len(chat.Label) < 8*/)

	user_msg, files := chat.Input.GetFullPrompt()

	var comp LLMCompletion
	comp.Temperature = 0.2
	comp.Max_tokens = 4096
	comp.Top_p = 0.95 //1.0
	comp.Frequency_penalty = 0
	comp.Presence_penalty = 0
	comp.Reasoning_effort = "" //low, high

	comp.AppName = appName

	//add default(without source) tools
	/*{
		defSource := sources["_default_"]
		if defSource != nil {
			comp.Tools = append(comp.Tools, defSource.Tools...)
		}
		comp.Tools = append(comp.Tools, "InstallDataSource")

		//add 'installed' tools
		for _, sourceName := range chat.Sources {
			source := sources[sourceName]
			if source != nil {
				comp.Tools = append(comp.Tools, source.Tools...)
			}
		}
	}*/

	old_num_msgs := len(chat.Messages.Messages)
	var err error
	comp.PreviousMessages, err = json.Marshal(chat.Messages)
	if err != nil {
		return fmt.Errorf("comp.PreviousMessages failed: %v", err)
	}

	//system
	{
		system_prompt := `You are an AI assistant, who enjoys precision and carefully follows the user's requirements.
If you need, you can use tool calling. Tools gives you precision and output data which you wouldn't know otherwise. Don't ask to call a tool, just do it! Call tools sequentially. Avoid tool call as parameter value.
If user wants to show/render/visualize some data, search for tools which 'shows'."`

		//sources
		/*system_prompt += "Here is the list of data sources in form of '<name> //<description>':\n"
		for sourceName, source := range sources {
			system_prompt += fmt.Sprintf("%s //%s", sourceName, source.Description)
			if slices.Index(chat.Sources, sourceName) >= 0 {
				system_prompt += " [installed]"
			}
			system_prompt += "\n"
		}
		system_prompt += fmt.Sprintf("Based on what you need you can install new tools by calling InstallDataSource with ChatID=%d. This will give you access to new tools.\n", st.ChatID)
		*/

		//Date time
		system_prompt += fmt.Sprintf("\nCurrent time is %s\n", time.Now().Format("Mon, 02 Jan 2006 15:04"))

		//Memory
		if root.Memory != "" {
			system_prompt += "\n" + root.Memory
		}

		comp.SystemMessage = system_prompt
	}

	if !continuee {
		comp.UserMessage = user_msg
		comp.UserFiles = files
	}
	comp.Max_iteration = 10
	comp.deltaMsg = func(msgJs []byte) {
		var msg ChatMsg
		err := json.Unmarshal(msgJs, &msg)
		if err != nil {
			return //err ....
		}

		last_i := len(chat.TempMessages.Messages) - 1
		if last_i >= 0 && chat.TempMessages.Messages[last_i].Stream {
			if msg.Stream {
				//copy
				msg.ShowParameters = chat.TempMessages.Messages[last_i].ShowParameters
				msg.ShowReasoning = chat.TempMessages.Messages[last_i].ShowReasoning
			}

			chat.TempMessages.Messages[last_i] = &msg //rewrite last
		} else {
			chat.TempMessages.Messages = append(chat.TempMessages.Messages, &msg)
		}

		//activate dash
		if msg.HasUI() {
			chat.Selected_user_msg = chat.GetNumUserMessages() - 1
		}
	}
	chat.TempMessages = ChatMsgs{} //reset

	err = comp.Run(caller)
	if err != nil {
		return fmt.Errorf("LLMxAICompleteChat.run() failed: %v", err)
	}

	chat.TempMessages = ChatMsgs{} //reset
	err = json.Unmarshal(comp.Out_messages, &chat.Messages)
	if err != nil {
		return fmt.Errorf("comp.Out_messages failed: %v", err)
	}

	//activate new dash
	for i := old_num_msgs; i < len(chat.Messages.Messages); i++ {
		msg := chat.Messages.Messages[i]
		if msg.HasUI() {
			chat.Selected_user_msg = chat.GetNumUserMessages() - 1
		}
	}

	return nil
}

func (app *RootApp) NumPins() int {
	n := 0
	for _, it := range app.Chats {
		if it.Pinned {
			n++
		} else {
			break
		}
	}
	return n
}

func (app *RootApp) RemoveChat(chat RootChat) error {

	//create "trash" folder
	os.MkdirAll(filepath.Join("..", app.Name, "Chats", "trash"), os.ModePerm)

	//copy file
	err := OsCopyFile(filepath.Join("..", app.Name, "Chats", "trash", chat.FileName),
		filepath.Join("..", app.Name, "Chats", chat.FileName))
	if err != nil {
		return err
	}

	//remove file
	os.Remove(filepath.Join("..", app.Name, "Chats", chat.FileName))

	for i, it := range app.Chats {
		if it.FileName == chat.FileName {

			app.Chats = slices.Delete(app.Chats, i, i+1)
			if i < app.Selected_chat_i {
				app.Selected_chat_i--
			}
			break
		}
	}

	return nil
}

func _isFileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func _copyFile(dst, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

func _getFileTime(path string) int64 {
	inf, err := os.Stat(path)
	if err == nil && inf != nil {
		return inf.ModTime().UnixNano()
	}
	return 0
}
