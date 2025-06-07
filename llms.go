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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

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
	CreatedTimeSec float64
	Provider       string //empty = user wrote it
	Model          string
	Seed           int

	Content OpenAI_content

	FinalTextSize int //without reasoning
	ShowReasoning bool

	UI_func     string
	UI_paramsJs string

	Usage LLMMsgUsage

	Time             float64
	TimeToFirstToken float64

	ShowParameters bool

	Stream bool
}
type ChatMsgs struct {
	Messages []*ChatMsg
}

type LLMMsgUsage struct {
	Prompt_tokens       int
	Input_cached_tokens int
	Completion_tokens   int
	Reasoning_tokens    int

	Prompt_price       float64
	Input_cached_price float64
	Completion_price   float64
	Reasoning_price    float64
}

func (dst *LLMMsgUsage) Add(src *LLMMsgUsage) {
	dst.Prompt_tokens += src.Prompt_tokens
	dst.Input_cached_tokens += src.Input_cached_tokens
	dst.Completion_tokens += src.Completion_tokens
	dst.Reasoning_tokens += src.Reasoning_tokens

	dst.Prompt_price += src.Prompt_price
	dst.Input_cached_price += src.Input_cached_price
	dst.Completion_price += src.Completion_price
	dst.Reasoning_price += src.Reasoning_price
}

type LLMComplete struct {
	Provider          string //"xai"
	Temperature       float64
	Top_p             float64
	Max_tokens        int
	Frequency_penalty float64
	Presence_penalty  float64
	Reasoning_effort  string //"low", "medium", "high"

	AppName string //load tools from

	PreviousMessages []byte //[]*ChatMsg
	SystemMessage    string
	UserMessage      string
	UserFiles        []string

	Response_format string

	Max_iteration int

	Out_StatusCode             int
	Out_messages               []byte //[]*ChatMsg
	Out_last_final_message     string
	Out_last_reasoning_message string

	Out_usage LLMMsgUsage

	delta func(msg *ChatMsg)
}

func (a *LLMComplete) Cmp(b *LLMComplete) bool {
	return a.Temperature == b.Temperature &&
		a.Top_p == b.Top_p &&
		bytes.Equal(a.PreviousMessages, b.PreviousMessages) &&
		a.SystemMessage == b.SystemMessage &&
		a.UserMessage == b.UserMessage
}

type LLMs struct {
	router *ToolsRouter

	Requests []LLMComplete
}

func NewLLMs(router *ToolsRouter) (*LLMs, error) {
	llms := &LLMs{router: router}

	//open
	{
		fl, err := os.ReadFile("apps/requests.json")
		if err == nil {
			json.Unmarshal(fl, &llms.Requests)
		}
	}
	return llms, nil
}

func (llms *LLMs) Complete(st *LLMComplete, msg *ToolsRouterMsg) error {

	//find
	for i := range llms.Requests {
		if llms.Requests[i].Cmp(st) {
			*st = llms.Requests[i]
			return nil
		}
	}

	//call
	switch st.Provider {
	case "xai":
		err := llms.router.sync.LLM_xai.run(st, llms.router, msg)
		if llms.router.log.Error(err) != nil {
			return err
		}
		//other providers
	}

	//add & save
	{
		llms.Requests = append(llms.Requests, *st)
		_, err := Tools_WriteJSONFile("apps/requests.json", llms.Requests)
		llms.router.log.Error(err)
	}

	return nil
}

func (msgs *ChatMsgs) AddUserMessage(text string, files []string) (*ChatMsg, error) {
	content := OpenAI_content{}
	content.Msg = &OpenAI_completion_msgContent{Role: "user"}
	if text != "" {
		content.Msg.AddText(text)
	}
	for _, file := range files {
		err := content.Msg.AddImageFile(file)
		if err != nil {
			return nil, err
		}
	}

	msg := &ChatMsg{CreatedTimeSec: float64(time.Now().UnixMilli()) / 1000, Content: content}
	msgs.Messages = append(msgs.Messages, msg)
	return msg, nil
}

func ChatMsg_GetReasoningTextIntro() string {
	return "\n\nReasoning: "
}

func (msgs *ChatMsgs) AddAssistentCalls(reasoning_text, final_text string, tool_calls []OpenAI_completion_msg_Content_ToolCall, usage LLMMsgUsage, dtime float64, timeToFirstToken float64, providerName string, modelName string) *ChatMsg {
	text := final_text
	if reasoning_text != "" {
		text = final_text + ChatMsg_GetReasoningTextIntro() + reasoning_text
	}

	content := OpenAI_content{}
	content.Calls = &OpenAI_completion_msgCalls{Role: "assistant", Content: text, Tool_calls: tool_calls}

	msg := &ChatMsg{Provider: providerName, Model: modelName, CreatedTimeSec: float64(time.Now().UnixMilli()) / 1000, Content: content, Usage: usage, Time: dtime, TimeToFirstToken: timeToFirstToken, FinalTextSize: len(final_text)}
	msgs.Messages = append(msgs.Messages, msg)
	return msg
}
func (msgs *ChatMsgs) AddCallResult(tool_name string, tool_use_id string, result string) *ChatMsg {
	content := OpenAI_content{}
	content.Result = &OpenAI_completion_msgResult{Role: "tool", Tool_call_id: tool_use_id, Name: tool_name, Content: result}

	msg := &ChatMsg{CreatedTimeSec: float64(time.Now().UnixMilli()) / 1000, Content: content}
	msgs.Messages = append(msgs.Messages, msg)
	return msg
}

func (msg *OpenAI_completion_msgContent) AddText(str string) {
	msg.Content = append(msg.Content, OpenAI_completion_msg_Content{Type: "text", Text: str})
}
func (msg *OpenAI_completion_msgContent) AddImage(data []byte, media_type string) { //ext="image/png","image/jpeg", "image/webp", "image/gif"(non-animated)
	prefix := "data:" + media_type + ";base64,"
	bs64 := base64.StdEncoding.EncodeToString(data)
	msg.Content = append(msg.Content, OpenAI_completion_msg_Content{Type: "image_url", Image_url: &OpenAI_completion_msg_Content_Image_url{Detail: "high", Url: prefix + bs64}})
}
func (msg *OpenAI_completion_msgContent) AddImageFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	ext := filepath.Ext(path)
	ext, _ = strings.CutPrefix(ext, ".")
	if ext == "" {
		return fmt.Errorf("missing file type(.ext)")
	}

	msg.AddImage(data, "image/"+ext)
	return nil
}
