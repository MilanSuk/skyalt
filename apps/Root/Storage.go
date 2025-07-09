package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"slices"
)

type RootChat struct {
	FileName string
	Label    string
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
}

type RootApp struct {
	Name            string
	Chats           []RootChat
	Selected_chat_i int
	Dev             RootDev
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

	//check select in range
	if root.Selected_app_i >= 0 {
		if root.Selected_app_i >= len(root.Apps) {
			root.Selected_app_i = len(root.Apps) - 1
		}
	}
	//return
	if root.Selected_app_i >= 0 {
		return root.Apps[root.Selected_app_i], nil
	}

	return nil, nil
}

func (app *RootApp) refreshChats() (*Chat, string, error) {

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

	//check selecte in range
	if app.Selected_chat_i >= 0 {
		if app.Selected_chat_i >= len(app.Chats) {
			app.Selected_chat_i = len(app.Chats) - 1
		}
	}

	//update and return
	if app.Selected_chat_i >= 0 {
		fileName := app.Chats[app.Selected_chat_i].FileName
		sourceChat, err := NewChat(filepath.Join("..", app.Name, "Chats", fileName))
		if err != nil {
			return nil, "", err
		}

		if sourceChat != nil {
			//reload
			app.Chats[app.Selected_chat_i].Label = sourceChat.Label
		}

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

type ChatMsgUI struct {
	FuncName string
	Params   map[string]interface{}
	Pos      int
}

type ChatMsg struct {
	Seed int

	Content OpenAI_content

	FinalTextSize int //without reasoning
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

type LayoutPick struct {
	Cd     LayoutPromptColor
	LLMTip string
	Points []UIPaintBrushPoint
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

	Label string //summary

	Input ChatInput

	PresetSystemPrompt string
	Messages           ChatMsgs

	Dash_call_id string

	Error string

	TempMessages ChatMsgs

	Sources []string
}

func NewChat(file string) (*Chat, error) {
	st := &Chat{Label: "chat"}
	return LoadFile(file, "Chat", "json", st, true)
}

func (st *Chat) GetChatID() string {
	return "chat_" + st.file
}

func (st *Chat) FindUI(tool_call_id string) *ChatMsg {
	for _, msg := range st.Messages.Messages {
		if msg.HasUI() {
			if msg.Content.Result.Tool_call_id == tool_call_id {
				return msg
			}
		}
	}
	for _, msg := range st.TempMessages.Messages {
		if msg.HasUI() {
			if msg.Content.Result.Tool_call_id == tool_call_id {
				return msg
			}
		}
	}
	return nil
}
func (st *Chat) FindPreviousUI(tool_call_id string) *ChatMsg {
	var preMsg *ChatMsg
	for _, msg := range st.Messages.Messages {
		if msg.HasUI() {
			if msg.Content.Result.Tool_call_id == tool_call_id {
				return preMsg
			}
			preMsg = msg
		}
	}
	for _, msg := range st.TempMessages.Messages {
		if msg.HasUI() {
			if msg.Content.Result.Tool_call_id == tool_call_id {
				return preMsg
			}
			preMsg = msg
		}
	}
	return nil
}
func (st *Chat) FindNextUI(tool_call_id string) *ChatMsg {
	next := false
	for _, msg := range st.Messages.Messages {
		if msg.HasUI() {
			if next {
				return msg
			}
			if msg.Content.Result.Tool_call_id == tool_call_id {
				next = true
			}
		}
	}
	for _, msg := range st.TempMessages.Messages {
		if msg.HasUI() {
			if next {
				return msg
			}
			if msg.Content.Result.Tool_call_id == tool_call_id {
				next = true
			}
		}
	}
	return nil
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
