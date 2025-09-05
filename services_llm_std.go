package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type OpenAIOutChoice_Message struct {
	Content           string //Final answer
	Reasoning_content string

	Tool_calls []OpenAI_completion_msg_Content_ToolCall
}

type OpenAIOutChoice struct {
	Message OpenAIOutChoice_Message
}
type OpenAIOut_UsageDetails struct {
	Reasoning_tokens int
}
type OpenAIOut_Usage struct {
	Prompt_tokens       int
	Input_cached_tokens int
	Completion_tokens   int
	Total_tokens        int

	Completion_tokens_details OpenAIOut_UsageDetails
}

type OpenAIOutError struct {
	Message string
}
type OpenAIOut struct {
	Choices   []OpenAIOutChoice
	Usage     OpenAIOut_Usage
	Citations []string
	Error     *OpenAIOutError
}

type OpenAI_completion_format struct {
	Type string `json:"type"` //json_object
	//Json_schema
}

type OpenAI_completion_msgSystem struct {
	Role    string `json:"role"` //"system"
	Content string `json:"content"`
}

/*type OpenAI_completion_tool_function_parameters_properties struct {
	Type        string   `json:"type"` //"number", "string"
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default,omitempty"`

	Items *OpenAI_completion_tool_function_parameters_properties `json:"items,omitempty"` //for arrays
}
type OpenAI_completion_tool_schema struct {
	Type                 string   `json:"type"` //"object"
	Required             []string `json:"required,omitempty"`
	AdditionalProperties bool     `json:"additionalProperties"`

	Properties map[string]*OpenAI_completion_tool_function_parameters_properties `json:"properties"`
}
type OpenAI_completion_tool_function struct {
	Name        string                        `json:"name"`
	Description string                        `json:"description"`
	Parameters  OpenAI_completion_tool_schema `json:"parameters"`
	Strict      bool                          `json:"strict"`
}

type OpenAI_completion_tool struct {
	Type     string                          `json:"type"` //"object"
	Function OpenAI_completion_tool_function `json:"function"`
}*/

type OpenAI_completion_Search_parameters struct {
	Mode string `json:"mode,omitempty"` //"auto", "on", "off"

	Return_citations   bool `json:"return_citations,omitempty"`
	Max_search_results int  `json:"max_search_results,omitempty"` //default is 20
}

type OpenAI_completion_Stream_options struct {
	Include_usage bool `json:"include_usage,omitempty"`
}
type OpenAI_completion_props struct {
	Seed  int    `json:"seed,omitempty"`
	Model string `json:"model"`

	Tools          []*ToolsOpenAI_completion_tool   `json:"tools,omitempty"`
	Messages       []interface{}                    `json:"messages"` //OpenAI_completion_msgSystem, OpenAI_completion_msg_Content, OpenAI_completion_msgCalls, OpenAI_completion_msgCalls,
	Stream         bool                             `json:"stream"`
	Stream_options OpenAI_completion_Stream_options `json:"stream_options,omitempty"`

	Temperature           float64 `json:"temperature"`                     //1.0
	Max_tokens            int     `json:"max_tokens,omitempty"`            //
	Max_completion_tokens int     `json:"max_completion_tokens,omitempty"` //for o1+ models
	Top_p                 float64 `json:"top_p,omitempty"`                 //1.0
	Frequency_penalty     float64 `json:"frequency_penalty,omitempty"`     //0
	Presence_penalty      float64 `json:"presence_penalty,omitempty"`      //0

	Response_format *OpenAI_completion_format `json:"response_format,omitempty"`

	Reasoning_effort string `json:"reasoning_effort,omitempty"` //"low", "high"

	Search_parameters *OpenAI_completion_Search_parameters `json:"search_parameters,omitempty"`
}

//var g_global_OpenAI_completion_lock sync.Mutex

func OpenAI_completion_Run(jsProps []byte, Completion_url string, api_key string, fnStreaming func(msg *ChatMsg) bool, msg *AppsRouterMsg) (OpenAIOut, int, float64, float64, error) {

	//maybe save start_time and sleep() here until some delay ....

	//g_global_OpenAI_completion_lock.Lock()
	//defer g_global_OpenAI_completion_lock.Unlock()

	st := time.Now().UnixMicro()

	if !strings.HasSuffix(Completion_url, "/") {
		Completion_url += "/"
	}
	Completion_url += "chat/completions"

	body := bytes.NewReader(jsProps)

	req, err := http.NewRequest(http.MethodPost, Completion_url, body)
	if LogsError(err) != nil {
		return OpenAIOut{}, -1, 0, -1, err
	}
	req.Header.Add("Content-Type", "application/json")
	if api_key == "" {
		req.Header.Add("Authorization", "Bearer no-key")
	} else {
		req.Header.Add("Authorization", "Bearer "+api_key)
	}

	if fnStreaming != nil {
		req.Header.Add("Accept", "text/event-stream")
		req.Header.Add("Cache-Control", "no-cache")
		req.Header.Add("Connection", "keep-alive")
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if LogsError(err) != nil {
		return OpenAIOut{}, -1, 0, -1, err
	}
	defer res.Body.Close()

	var ret OpenAIOut

	time_to_first_token := -1.0

	if fnStreaming != nil {
		// Check response status
		if res.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(res.Body)
			return OpenAIOut{}, -1, 0, -1, LogsErrorf("unexpected status code: %d, body: %s", res.StatusCode, string(body))
		}

		// Read streaming response
		reader := bufio.NewReader(res.Body)
		streaming_ok := true
		for streaming_ok {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				return OpenAIOut{}, -1, 0, -1, LogsErrorf("error reading stream: %w", err)
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Check for SSE data prefix
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			//callFuncPrint("data: " + data)

			// Parse JSON response
			type Delta struct {
				Content           string                                   `json:"content"`
				Reasoning_content string                                   `json:"reasoning_content"`
				Tool_calls        []OpenAI_completion_msg_Content_ToolCall `json:"tool_calls"`
			}
			type StreamChoice struct {
				Delta Delta `json:"delta"`
				//FinishReason string `json:"finish_reason"`
			}
			type StreamResponse struct {
				Choices []StreamChoice  `json:"choices"`
				Usage   OpenAIOut_Usage `json:"usage"`
			}
			var streamResp StreamResponse

			err = LogsJsonUnmarshal([]byte(data), &streamResp)
			if err != nil {
				return OpenAIOut{}, -1, 0, -1, err
			}

			if streamResp.Usage.Total_tokens > 0 {
				ret.Usage = streamResp.Usage
			}

			for _, choice := range streamResp.Choices {
				if len(ret.Choices) == 0 {
					ret.Choices = append(ret.Choices, OpenAIOutChoice{})
				}

				if time_to_first_token < 0 && (choice.Delta.Content != "" || choice.Delta.Reasoning_content != "") {
					time_to_first_token = float64(time.Now().UnixMicro()-st) / 1000000
				}

				//add content
				ret.Choices[0].Message.Content += choice.Delta.Content

				//reasoning content
				ret.Choices[0].Message.Reasoning_content += choice.Delta.Reasoning_content

				//add tools
				for _, tool_call := range choice.Delta.Tool_calls {
					index := tool_call.Index

					if index >= len(ret.Choices[0].Message.Tool_calls) {
						ret.Choices[0].Message.Tool_calls = append(ret.Choices[0].Message.Tool_calls, tool_call)
					} else {
						ret.Choices[0].Message.Tool_calls[index].Function.Arguments += tool_call.Function.Arguments
					}
				}
				//ret.Choices[0].Message.Tool_calls = append(ret.Choices[0].Message.Tool_calls, choice.Delta.Tool_calls...)

				//callback
				var msgs ChatMsgs
				msgs.AddAssistentCalls(ret.Choices[0].Message.Reasoning_content, ret.Choices[0].Message.Content, ret.Choices[0].Message.Tool_calls, LLMMsgUsage{})
				if !fnStreaming(msgs.Messages[0]) {
					streaming_ok = false
					break //interrupted
				}
			}

			if !msg.GetContinue() {
				return OpenAIOut{}, res.StatusCode, 0, -1, LogsErrorf("interrupted")
			}
		}

	} else {
		js, err := io.ReadAll(res.Body)
		if err != nil {
			return OpenAIOut{}, res.StatusCode, 0, -1, err
		}

		if res.StatusCode != http.StatusOK {
			return OpenAIOut{}, res.StatusCode, 0, -1, LogsErrorf("statusCode %d != %d, response: %s", res.StatusCode, http.StatusOK, string(js))
		}

		if len(js) == 0 {
			return OpenAIOut{}, res.StatusCode, 0, -1, LogsErrorf("output is empty")
		}

		err = LogsJsonUnmarshal(js, &ret)
		if err != nil {
			return OpenAIOut{}, res.StatusCode, 0, -1, err
		}

		if ret.Error != nil && ret.Error.Message != "" {
			return OpenAIOut{}, res.StatusCode, 0, -1, errors.New(ret.Error.Message)
		}
	}

	dt := float64(time.Now().UnixMicro()-st) / 1000000
	return ret, res.StatusCode, dt, time_to_first_token, nil
}

type LLMMsgStats struct {
	Function string
	Usage    LLMMsgUsage
}

func OpenAI_Complete(Provider string, OpenAI_url string, API_key string, st *LLMComplete, app_port int, tools []*ToolsOpenAI_completion_tool, msg *AppsRouterMsg, fnGetTextPrice func(in, reason, cached, out int) (float64, float64, float64, float64)) ([]LLMMsgStats, error) {

	//Messages
	var msgs ChatMsgs
	if len(st.PreviousMessages) > 0 {
		err := LogsJsonUnmarshal(st.PreviousMessages, &msgs)
		if err != nil {
			return nil, err
		}
	}

	if st.UserMessage != "" || len(st.UserFiles) > 0 {
		m1, err := msgs.AddUserMessage(st.UserMessage, st.UserFiles)
		if err != nil {
			return nil, err
		}
		if st.delta != nil {
			st.delta(m1)
		}
	}

	seed := 1
	if len(msgs.Messages) > 0 {
		seed = msgs.Messages[len(msgs.Messages)-1].Seed
		if seed <= 0 {
			seed = 1
		}
	}

	var ret_stats []LLMMsgStats

	last_final_msg := ""
	last_reasoning_msg := ""

	iter := 0
	for iter < st.Max_iteration {
		//convert msgs to OpenAI
		var messages []interface{}
		messages = append(messages, OpenAI_completion_msgSystem{Role: "system", Content: st.SystemMessage})
		for _, msg := range msgs.Messages {
			if msg.Content.Msg != nil {
				messages = append(messages, msg.Content.Msg)
			}
			if msg.Content.Calls != nil {
				messages = append(messages, msg.Content.Calls)
			}
			if msg.Content.Result != nil {
				messages = append(messages, msg.Content.Result)
			}
		}

		props := OpenAI_completion_props{
			Stream:         true,
			Stream_options: OpenAI_completion_Stream_options{Include_usage: true},

			Seed:  seed,
			Model: st.Out_usage.Model,

			Tools:    tools,
			Messages: messages,

			Temperature:       st.Temperature,
			Max_tokens:        st.Max_tokens,
			Top_p:             st.Top_p,
			Frequency_penalty: st.Frequency_penalty,
			Presence_penalty:  st.Presence_penalty,
			Reasoning_effort:  st.Reasoning_effort,
		}
		if st.Response_format != "" {
			props.Response_format = &OpenAI_completion_format{Type: st.Response_format}
		}

		fnStreaming := func(chatMsg *ChatMsg) bool {
			chatMsg.Seed = seed
			chatMsg.Stream = true
			chatMsg.ShowParameters = true
			chatMsg.ShowReasoning = true

			if st.delta != nil {
				st.delta(chatMsg)
			}

			return msg.GetContinue()
		}

		jsProps, err := LogsJsonMarshal(props)
		if err != nil {
			return nil, err
		}
		out, status, dt, time_to_first_token, err := OpenAI_completion_Run(jsProps, OpenAI_url, API_key, fnStreaming, msg)
		st.Out_StatusCode = status
		if err != nil {
			return nil, err
		}

		if !msg.GetContinue() {
			return nil, nil
		}

		if len(out.Choices) > 0 {

			var usage LLMMsgUsage
			{
				usage.Prompt_tokens = out.Usage.Prompt_tokens
				usage.Input_cached_tokens = out.Usage.Input_cached_tokens
				usage.Completion_tokens = out.Usage.Completion_tokens
				usage.Reasoning_tokens = out.Usage.Completion_tokens_details.Reasoning_tokens

				usage.Provider = Provider
				usage.Model = st.Out_usage.Model
				usage.CreatedTimeSec = float64(time.Now().UnixMicro()) / 1000000
				usage.TimeToFirstToken = time_to_first_token
				usage.DTime = dt

				if fnGetTextPrice != nil {
					usage.Prompt_price, usage.Reasoning_price, usage.Input_cached_price, usage.Completion_price = fnGetTextPrice(usage.Prompt_tokens, usage.Reasoning_tokens, usage.Input_cached_tokens, usage.Completion_tokens)
				}

				//add
				{
					st.Out_usage.Add(&usage)
				}
			}

			calls := out.Choices[0].Message.Tool_calls
			m2 := msgs.AddAssistentCalls(out.Choices[0].Message.Reasoning_content, out.Choices[0].Message.Content, calls, usage)
			if st.delta != nil {
				st.delta(m2)
			}

			last_final_msg = out.Choices[0].Message.Content
			last_reasoning_msg = out.Choices[0].Message.Reasoning_content

			for _, call := range calls {
				var result string

				//call it
				resJs, uiGob, cmdsGob, err := _ToolsCaller_CallBuild(app_port, msg.msg_id, 0, call.Function.Name, []byte(call.Function.Arguments))
				if err != nil {
					return nil, err
				}
				//resJs, tool_ui, err := CallToolApp(st.AppName, call.Function.Name, []byte(call.Function.Arguments), caller)

				//add cmds
				msg.out_flushed_cmdsGob = append(msg.out_flushed_cmdsGob, cmdsGob)

				resMap := make(map[string]interface{})
				err = LogsJsonUnmarshal(resJs, &resMap)
				if err != nil {
					return nil, err
				}

				//Out_ -> result
				{
					num_outs := 0
					for nm := range resMap {
						if strings.HasPrefix(strings.ToLower(nm), "out") {
							num_outs++
						}
					}
					for nm, val := range resMap {
						if strings.HasPrefix(strings.ToLower(nm), "out") {
							var vv string
							var tp string
							switch v := val.(type) {
							case string:
								tp = "string"
								vv = v
							case float64:
								tp = "float64"
								vv = strconv.FormatFloat(v, 'f', -1, 64)
							case int:
								tp = "int"
								vv = strconv.FormatInt(int64(v), 10)
							case int64:
								tp = "int64"
								vv = strconv.FormatInt(int64(v), 10)
							default:
								tp = "unknown"
								vv = fmt.Sprintf("%v", v)
							}

							if num_outs == 1 {
								result = vv
								break
							} else {
								result += fmt.Sprintf("%s(%s): %s\n", nm, tp, vv)
							}
						}
					}
				}

				var tool_ui UI
				LogsGobUnmarshal(uiGob, &tool_ui)

				hasUI := tool_ui.Is()
				if hasUI {
					if result != "" {
						result += "\n"
					}
					result += "Successfully shown on screen."
				}

				res_msg := msgs.AddCallResult(call.Function.Name, call.Id, result)
				if hasUI {
					res_msg.UI_func = call.Function.Name
					res_msg.UI_paramsJs = string(resJs)
				}
				if st.delta != nil {
					st.delta(res_msg)
				}
			}

			//log stats
			ret_stats = append(ret_stats, LLMMsgStats{
				Function: "completion",
				Usage:    usage,
			})

			if len(calls) == 0 {
				break
			}
		}
		iter++
	}

	st.Out_answer = last_final_msg
	st.Out_reasoning = last_reasoning_msg

	var err error
	st.Out_messages, err = LogsJsonMarshal(msgs)
	if err != nil {
		return nil, err
	}

	return ret_stats, nil
}

type OpenAI_getImage_props struct {
	Prompt string `json:"prompt"`
	Model  string `json:"model"`
	N      int    `json:"n"`

	Response_format string `json:"response_format"` //Response format to return the image in. Can be url or b64_json.

	//Quality string `json:"quality"` //Quality of the image.
	//Size    string `json:"size"`    //Size of the image.
	//Style   string `json:"style"`   //Style of the image.
	//User    string `json:"user"`    //A unique identifier representing your end-user, which can help xAI to monitor and detect abuse.
}

//var g_global_OpenAI_genImage_lock sync.Mutex

type OpenAIGenImageOutData struct {
	B64_json       string //"data:image/png;base64,..."
	Revised_prompt string
}

type OpenAIGenImage_UsageDetails struct {
	Reasoning_tokens int
}
type OpenAIGenImage_Usage struct {
	Prompt_tokens       int
	Input_cached_tokens int
	Completion_tokens   int
	Total_tokens        int

	Completion_tokens_details OpenAIGenImage_UsageDetails
}

type OpenAIGenImage_Error struct {
	Message string
}

type OpenAIGenImageOut struct {
	Data  []OpenAIGenImageOutData
	Usage OpenAIGenImage_Usage
	Error *OpenAIGenImage_Error
}

func OpenAI_genImage_Run(jsProps []byte, Completion_url string, api_key string) (OpenAIGenImageOut, int, float64, error) {
	//g_global_OpenAI_genImage_lock.Lock()
	//defer g_global_OpenAI_genImage_lock.Unlock()

	st := time.Now().UnixMicro()

	if !strings.HasSuffix(Completion_url, "/") {
		Completion_url += "/"
	}
	Completion_url += "images/generations"

	body := bytes.NewReader(jsProps)

	req, err := http.NewRequest(http.MethodPost, Completion_url, body)
	if LogsError(err) != nil {
		return OpenAIGenImageOut{}, -1, 0, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+api_key)

	client := &http.Client{}
	res, err := client.Do(req)
	if LogsError(err) != nil {
		return OpenAIGenImageOut{}, -1, 0, err
	}
	defer res.Body.Close()

	js, err := io.ReadAll(res.Body)
	if err != nil {
		return OpenAIGenImageOut{}, res.StatusCode, 0, err
	}

	if len(js) == 0 {
		return OpenAIGenImageOut{}, res.StatusCode, 0, LogsErrorf("output is empty")
	}

	var out OpenAIGenImageOut
	err = LogsJsonUnmarshal(js, &out)
	if err != nil {
		return OpenAIGenImageOut{}, res.StatusCode, 0, err
	}
	if out.Error != nil && out.Error.Message != "" {
		return OpenAIGenImageOut{}, res.StatusCode, 0, errors.New(out.Error.Message)
	}

	if res.StatusCode != http.StatusOK {
		return OpenAIGenImageOut{}, res.StatusCode, 0, LogsErrorf("statusCode %d != 200, response: %s", res.StatusCode, string(js))
	}

	tm := float64(time.Now().UnixMicro()-st) / 1000000

	return out, res.StatusCode, tm, nil
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

	msg := &ChatMsg{Content: content, Usage: LLMMsgUsage{CreatedTimeSec: float64(time.Now().UnixMicro()) / 1000000}}
	msgs.Messages = append(msgs.Messages, msg)
	return msg, nil
}

func ChatMsg_GetDivAfterReasoning() string {
	return "\n\nFinal message: "
}

func (msgs *ChatMsgs) AddAssistentCalls(reasoning_text, final_text string, tool_calls []OpenAI_completion_msg_Content_ToolCall, usage LLMMsgUsage) *ChatMsg {
	text := final_text
	if reasoning_text != "" {
		text = reasoning_text + ChatMsg_GetDivAfterReasoning() + final_text
	}

	content := OpenAI_content{}
	content.Calls = &OpenAI_completion_msgCalls{Role: "assistant", Content: text, Tool_calls: tool_calls}

	msg := &ChatMsg{Content: content, Usage: usage, ReasoningSize: OsTrn(len(reasoning_text) == 0, 0, len(reasoning_text)+len(ChatMsg_GetDivAfterReasoning()))}
	msgs.Messages = append(msgs.Messages, msg)
	return msg
}
func (msgs *ChatMsgs) AddCallResult(tool_name string, tool_use_id string, result string) *ChatMsg {
	content := OpenAI_content{}
	content.Result = &OpenAI_completion_msgResult{Role: "tool", Tool_call_id: tool_use_id, Name: tool_name, Content: result}

	msg := &ChatMsg{Content: content, Usage: LLMMsgUsage{CreatedTimeSec: float64(time.Now().UnixMicro()) / 1000000}}
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
	if LogsError(err) != nil {
		return err
	}

	ext := filepath.Ext(path)
	ext, _ = strings.CutPrefix(ext, ".")
	if ext == "" {
		return LogsErrorf("missing file type(.ext)")
	}

	msg.AddImage(data, "image/"+ext)
	return nil
}
