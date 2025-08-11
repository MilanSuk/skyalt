package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"time"
)

const g_ShowApp_prompt_height = 7

type RootChat struct {
	FileName string
	Label    string
	Pinned   bool
}

// Root
type Root struct {
	Chats           []RootChat
	Selected_chat_i int
}

func NewChats(file string) (*Root, error) {
	st := &Root{}

	return LoadFile(file, "Chats", "json", st, true)
}

func (app *Root) refreshChats() (*Chat, string, error) {

	//check range
	if app.Selected_chat_i >= 0 && app.Selected_chat_i >= len(app.Chats) {
		app.Selected_chat_i = len(app.Chats) - 1
	}

	chats_folder := filepath.Join("Chats")
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
		sourceChat, err := NewChat(filepath.Join("Chats", fileName))
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

type ChatInput struct {
	Text string

	Files          []string
	FilePickerPath string

	Multilined bool
}

// Chat has label, messages.
type Chat struct {
	file        string
	scroll_down bool

	Input ChatInput

	PresetSystemPrompt string
	Messages           ChatMsgs

	Selected_user_msg int

	Error string

	TempMessages ChatMsgs

	Sources []string
}

func NewChat(file string) (*Chat, error) {
	st := &Chat{file: file}
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

func (st *ChatInput) Reset() {
	st.Text = ""
	st.Files = nil
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

func (chat *Chat) _sendIt(caller *ToolCaller, root *Root, continuee bool) error {
	//MsgsDiv.VScrollToTheBottom(false, caller)
	caller.SendFlushCmd()

	if !continuee && chat.Input.Text == "" {
		return nil //empty text
	}

	caller.SetMsgName(chat.GetChatID())

	chat.CutMessages(chat.Selected_user_msg)

	chat.scroll_down = true

	err := chat.complete(caller, root, continuee)
	if err != nil {
		return fmt.Errorf("complete() failed: %v", err)
	}
	if !continuee {
		chat.Input.Reset()
	}

	//MsgsDiv.VScrollToTheBottom(true, caller)
	return nil
}

func (chat *Chat) complete(caller *ToolCaller, root *Root, continuee bool) error {

	//needSummary := (len(chat.Messages.Messages) == 0 /*|| len(chat.Label) < 8*/)

	user_msg := chat.Input.Text
	files := chat.Input.Files

	var comp LLMCompletion
	comp.Temperature = 0.2
	comp.Max_tokens = 4096
	comp.Top_p = 0.95 //1.0
	comp.Frequency_penalty = 0
	comp.Presence_penalty = 0
	comp.Reasoning_effort = "" //low, high

	//comp.AppName = appName

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
		system_prompt := `You are an AI assistant, who enjoys precision and carefully follows the user's requirements."`

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
		return fmt.Errorf("comp.run() failed: %v", err)
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

func (app *Root) NumPins() int {
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

func (app *Root) RemoveChat(chat RootChat) error {

	//create "trash" folder
	os.MkdirAll(filepath.Join("Chats", "trash"), os.ModePerm)

	//copy file
	err := OsCopyFile(filepath.Join("Chats", "trash", chat.FileName),
		filepath.Join("Chats", chat.FileName))
	if err != nil {
		return err
	}

	//remove file
	os.Remove(filepath.Join("Chats", chat.FileName))

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
