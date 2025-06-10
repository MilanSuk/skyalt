package main

import (
	"encoding/json"
	"fmt"
	"image/color"
)

type ChatMsgUsage struct {
	Prompt_tokens       int
	Input_cached_tokens int
	Completion_tokens   int
	Reasoning_tokens    int

	Prompt_price       float64
	Input_cached_price float64
	Completion_price   float64
	Reasoning_price    float64
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
	CreatedTimeSec float64
	Provider       string //empty = user wrote it
	Model          string
	Seed           int

	Content OpenAI_content

	FinalTextSize int //without reasoning
	ShowReasoning bool

	UI_func     string
	UI_paramsJs string

	Usage ChatMsgUsage

	Time             float64
	TimeToFirstToken float64

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

func (st *ChatInput) MergePick(in LayoutPick) {
	//find & update color
	if in.LLMTip != "" {
		for i, it := range st.Picks {
			if it.LLMTip == in.LLMTip {
				in.Cd = st.Picks[i].Cd
				break
			}
		}
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
func (msg *ChatMsg) GetSpeed() float64 {
	toks := msg.Usage.Completion_tokens + msg.Usage.Reasoning_tokens
	if msg.Time == 0 {
		return 0
	}
	return float64(toks) / msg.Time
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
		dt += msgs.Messages[i].Time
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
