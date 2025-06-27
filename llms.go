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
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
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
	Index    int                                             `json:"index,omitempty"`
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

type LLMMsgUsage struct {
	Provider         string //empty = user wrote it
	Model            string
	CreatedTimeSec   float64
	TimeToFirstToken float64
	DTime            float64

	Prompt_tokens       int
	Input_cached_tokens int
	Completion_tokens   int
	Reasoning_tokens    int

	Prompt_price       float64
	Input_cached_price float64
	Completion_price   float64
	Reasoning_price    float64
}

func (u *LLMMsgUsage) TotalPrice() float64 {
	return u.Prompt_price + u.Input_cached_price + u.Completion_price + u.Reasoning_price
}
func (dst *LLMMsgUsage) Add(src *LLMMsgUsage) {

	dst.Model = src.Model
	dst.DTime += src.DTime

	if dst.CreatedTimeSec == 0 {
		dst.CreatedTimeSec = src.CreatedTimeSec
	}
	if dst.TimeToFirstToken == 0 {
		dst.TimeToFirstToken = src.TimeToFirstToken
	}

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

	Response_format string //"", "json_object"

	Max_iteration int

	Out_StatusCode int
	Out_messages   []byte //[]*ChatMsg
	Out_tools      []byte

	Out_answer    string
	Out_reasoning string

	Out_usage LLMMsgUsage

	delta func(msg *ChatMsg)
}

func NewLLMCompletion(systemMessage string, userMessage string) *LLMComplete {
	return &LLMComplete{Temperature: 0.2, Max_tokens: 16384, Top_p: 0.95, SystemMessage: systemMessage, UserMessage: userMessage}
}

func (a *LLMComplete) Cmp(b *LLMComplete) bool {
	return a.Out_usage.Model == b.Out_usage.Model &&
		a.Temperature == b.Temperature &&
		a.Top_p == b.Top_p &&
		a.SystemMessage == b.SystemMessage &&
		a.UserMessage == b.UserMessage &&
		bytes.Equal(a.Out_tools, b.Out_tools) &&
		bytes.Equal(a.PreviousMessages, b.PreviousMessages)
}

type LLMGenerateImage struct {
	Prompt     string //Prompt for image generation.
	Num_images int    //Number of images to be generated

	Out_StatusCode      int
	Out_images          [][]byte
	Out_revised_prompts []string
	Out_dtime_sec       float64
}

type LLMTranscribe struct {
	AudioBlob    []byte
	BlobFileName string //ext.... (blob.wav, blob.mp3)

	Temperature     float64 //0
	Response_format string

	Out_StatusCode int
	Out_Output     []byte
}

type LLMSpeech struct {
	Text string

	Out_StatusCode int
	Out_Output     []byte
}

type LLMs struct {
	router *ToolsRouter

	lock sync.Mutex

	Cache []LLMComplete
}

func NewLLMs(router *ToolsRouter) (*LLMs, error) {
	llms := &LLMs{router: router}

	//open
	{
		fl, err := os.ReadFile("temp/llms_cache.json")
		if err == nil {
			json.Unmarshal(fl, &llms.Cache)
		}
	}
	return llms, nil
}

// usecase: "tools", "code", "chat"
func (llms *LLMs) Complete(st *LLMComplete, msg *ToolsRouterMsg, usecase string) error {

	dev := &llms.router.sync.Device
	model := ""
	switch strings.ToLower(dev.Chat_provider) {

	case "xai":
		model = "grok-3"
		if !dev.Chat_smarter {
			model += "-mini"
		}
		if dev.Chat_faster {
			model += "-fast"
		}

	case "mistral":
		switch usecase {
		case "tools", "code":
			model = "devstral-small-latest"
			if dev.Chat_smarter {
				model = "codestral-latest"
			}
		case "chat":
			model = "mistral-small-latest"
			if dev.Chat_smarter {
				model = "mistral-large-latest"
			}
		}

	case "openai":
		switch usecase {
		case "tools", "code":
			model = "gpt-4.1-mini"
			if dev.Chat_smarter {
				model = "o4-mini"
			}
		case "chat":
			model = "gpt-4.1-mini"
			if dev.Chat_smarter {
				model = "o4-mini"
			}
		}

	case "llama.cpp":
		model = ""
	}

	st.Out_usage.Model = model

	//get provider & model name
	/*if llms.router.sync.LLM_xai != nil {
		chat, _ := llms.router.sync.LLM_xai.FindModel(st.Model)
		if chat != nil {
			st.Model = chat.Id
			provider = "xai"
		}
	}
	if llms.router.sync.LLM_mistral != nil {
		chat, _ := llms.router.sync.LLM_mistral.FindModel(st.Model)
		if chat != nil {
			st.Model = chat.Id
			provider = "mistral"
		}
	}
	if llms.router.sync.LLM_openai != nil {
		chat, _ := llms.router.sync.LLM_openai.FindModel(st.Model)
		if chat != nil {
			st.Model = chat.Id
			provider = "openai"
		}
	}
	if provider == "" && st.Model == "llamacpp" {
		provider = "llamacpp"
	}*/

	//Tools
	var tools []*ToolsOpenAI_completion_tool
	app_port := -1
	if st.AppName != "" {
		app := llms.router.FindApp(st.AppName)
		if app != nil {
			tools = app.GetAllSchemas()

			var err error
			st.Out_tools, err = json.Marshal(tools)
			if err != nil {
				return err
			}

			//start it
			err = app.CheckRun()
			if err != nil {
				return err
			}

			app_port = app.Process.port

		} else {
			return fmt.Errorf("app '%s' not found", st.AppName)
		}
	}

	//find in cache
	for i := range llms.Cache {
		if llms.Cache[i].Cmp(st) {
			*st = llms.Cache[i]
			return nil
		}
	}

	if st.Max_iteration <= 0 {
		st.Max_iteration = 1
	}

	llms.lock.Lock()
	defer llms.lock.Unlock()

	//call
	switch strings.ToLower(dev.Chat_provider) {
	case "xai":
		err := llms.router.sync.LLM_xai.Complete(st, app_port, tools, llms.router, msg)
		if err != nil {
			return err
		}
	case "mistral":
		err := llms.router.sync.LLM_mistral.Complete(st, app_port, tools, llms.router, msg)
		if err != nil {
			return err
		}
	case "openai":
		err := llms.router.sync.LLM_openai.Complete(st, app_port, tools, llms.router, msg)
		if err != nil {
			return err
		}

	case "llama.cpp":
		err := llms.router.sync.LLM_llama.Complete(st, app_port, tools, llms.router, msg)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("provider not found")
	}

	//add & save cache
	{
		llms.Cache = append(llms.Cache, *st)
		_, err := Tools_WriteJSONFile("temp/llms_cache.json", llms.Cache)
		llms.router.log.Error(err)
	}

	return nil
}

func (llms *LLMs) GetUsage() []LLMMsgUsage {
	var ret []LLMMsgUsage
	for _, it := range llms.Cache {
		ret = append(ret, it.Out_usage)
	}
	return ret
}

func (llms *LLMs) GenerateImage(st *LLMGenerateImage, msg *ToolsRouterMsg) error {
	llms.lock.Lock()
	defer llms.lock.Unlock()

	dev := &llms.router.sync.Device

	//call
	switch strings.ToLower(dev.Image_provider) {
	case "xai":
		err := llms.router.sync.LLM_xai.GenerateImage(st, llms.router, msg)
		if err != nil {
			return err
		}
		//other providers ....
	}

	return nil
}

func (llms *LLMs) Transcribe(st *LLMTranscribe) error {
	llms.lock.Lock()
	defer llms.lock.Unlock()

	dev := &llms.router.sync.Device

	//call
	switch strings.ToLower(dev.STT_provider) {
	case "openai":
		err := llms.router.sync.LLM_openai.Transcribe(st)
		if err != nil {
			return err
		}

	case "whisper.cpp":
		err := llms.router.sync.LLM_wsp.Transcribe(st)
		if err != nil {
			return err
		}
		//other providers ....
	}

	return nil
}
