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
	"log"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/go-audio/audio"
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
	UID string

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

	delta      func(msg *ChatMsg)
	wip_answer string
	msg        *AppsRouterMsg
}

func NewLLMCompletion() *LLMComplete {
	comp := &LLMComplete{}
	comp.Temperature = 0.2
	comp.Max_tokens = 32768 //65536
	comp.Top_p = 0.95       //1.0
	comp.Frequency_penalty = 0
	comp.Presence_penalty = 0
	comp.Reasoning_effort = ""
	comp.Max_iteration = 1
	return comp
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
	services *Services

	running      []*LLMComplete
	running_lock sync.Mutex

	Cache      []LLMComplete
	cache_lock sync.Mutex
}

func NewLLMs(services *Services) (*LLMs, error) {
	llms := &LLMs{services: services}

	//open
	{
		fl, err := os.ReadFile("temp/llms_cache.json")
		if err == nil {
			LogsJsonUnmarshal(fl, &llms.Cache)
		}
	}
	return llms, nil
}

func (llms *LLMs) findCache(st *LLMComplete) bool {
	llms.cache_lock.Lock()
	defer llms.cache_lock.Unlock()

	for i := range llms.Cache {
		if llms.Cache[i].Cmp(st) {
			*st = llms.Cache[i]
			return true
		}
	}
	return false
}

func (llms *LLMs) addCache(st *LLMComplete) {
	llms.cache_lock.Lock()
	defer llms.cache_lock.Unlock()

	llms.Cache = append(llms.Cache, *st)
	Tools_WriteJSONFile("temp/llms_cache.json", llms.Cache)
}

func (llms *LLMs) Find(uid string, msg *AppsRouterMsg) *LLMComplete {
	llms.running_lock.Lock()
	defer llms.running_lock.Unlock()

	for _, it := range llms.running {
		if it.UID == uid {
			if it.msg.CmpLastStack(msg) {
				return it
			}
		}
	}
	return nil
}

// usecase: "tools", "code", "chat"
func (llms *LLMs) Complete(st *LLMComplete, msg *AppsRouterMsg, usecase string) error {

	st.msg = msg

	dev := &llms.services.sync.Device

	provider := dev.App_provider
	model := dev.App_model

	switch strings.ToLower(usecase) {
	case "code":
		provider = dev.Code_provider
		model = dev.Code_model
	}

	st.Out_usage.Model = model

	//Tools
	if llms.services.fnGetAppPortAndTools == nil {
		log.Fatalf("fnGetAppPortAndTools is nill")
	}

	app_port, tools, err := llms.services.fnGetAppPortAndTools(st.AppName)
	if err != nil {
		return err
	}
	if len(tools) > 0 {
		var err error
		st.Out_tools, err = LogsJsonMarshal(tools)
		if err != nil {
			return err
		}
	}

	//find in cache
	if llms.findCache(st) {
		return nil
	}

	if st.Max_iteration <= 0 {
		st.Max_iteration = 1
	}

	//add into running list
	{
		//add
		llms.running_lock.Lock()
		llms.running = append(llms.running, st)
		llms.running_lock.Unlock()

		defer func() {
			//remove
			llms.running_lock.Lock()
			defer llms.running_lock.Unlock()
			for i, it := range llms.running {
				if it == st {
					llms.running = slices.Delete(llms.running, i, i+1)
					break
				}
			}
		}()
	}

	/*
		//keep for testing - bypass findCache() above
		if st.delta != nil {
			for i := range 1000 {
				st.delta(&ChatMsg{Content: OpenAI_content{Calls: &OpenAI_completion_msgCalls{Content: fmt.Sprintf("hello world: %d", i)}}})

				for range 100 {
					time.Sleep(10 * time.Millisecond)
					if !msg.GetContinue() {
						return nil
					}
				}
			}
		}*/

	//call
	switch strings.ToLower(provider) {
	case "xai":
		err := llms.services.sync.LLM_xai.Complete(st, app_port, tools, msg)
		if err != nil {
			return err
		}
	case "mistral":
		err := llms.services.sync.LLM_mistral.Complete(st, app_port, tools, msg)
		if err != nil {
			return err
		}
	case "openai":
		err := llms.services.sync.LLM_openai.Complete(st, app_port, tools, msg)
		if err != nil {
			return err
		}
	case "groq":
		err := llms.services.sync.LLM_groq.Complete(st, app_port, tools, msg)
		if err != nil {
			return err
		}

	case "llama.cpp":
		err := llms.services.sync.LLM_llama.Complete(st, app_port, tools, msg)
		if err != nil {
			return err
		}
	default:
		return LogsErrorf("provider '%s' not found", provider)
	}

	//add & save cache
	llms.addCache(st)

	return nil
}

func (llms *LLMs) GetUsage() []LLMMsgUsage {
	var ret []LLMMsgUsage
	for _, it := range llms.Cache {
		ret = append(ret, it.Out_usage)
	}
	return ret
}

func (llms *LLMs) GenerateImage(st *LLMGenerateImage, msg *AppsRouterMsg) error {
	dev := &llms.services.sync.Device

	//call
	switch strings.ToLower(dev.Image_provider) {
	case "xai":
		err := llms.services.sync.LLM_xai.GenerateImage(st, msg)
		if err != nil {
			return err
		}
		//other providers ....
	}

	return nil
}

func (llms *LLMs) Transcribe(st *LLMTranscribe) error {
	dev := &llms.services.sync.Device

	//call
	switch strings.ToLower(dev.STT_provider) {
	case "openai":
		err := llms.services.sync.LLM_openai.Transcribe(st)
		if err != nil {
			return err
		}
	case "groq":
		err := llms.services.sync.LLM_groq.Transcribe(st)
		if err != nil {
			return err
		}

	case "whisper.cpp":
		err := llms.services.sync.LLM_wsp.Transcribe(st)
		if err != nil {
			return err
		}
		//other providers ....
	}

	return nil
}

func (llms *LLMs) TranscribeBuff(buff *audio.IntBuffer, format string, response_format string) (string, error) {
	//convert
	Out_bytes, err := FFMpeg_convertIntoFile(buff, format, 16000)
	if err != nil {
		return "", err
	}

	comp := LLMTranscribe{
		AudioBlob:       Out_bytes,
		BlobFileName:    "blob." + format,
		Temperature:     0,
		Response_format: response_format,
	}

	err = llms.Transcribe(&comp)
	if err != nil {
		return "", err
	}

	return string(comp.Out_Output), nil
}
