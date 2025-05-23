package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

// [ignore]
type LLMxAICompleteChat struct {
	Model             string
	Temperature       float64
	Max_tokens        int
	Top_p             float64
	Frequency_penalty float64
	Presence_penalty  float64
	Reasoning_effort  string //"low", "medium", "high"

	PreviousMessages []byte //[]*ChatMsg
	SystemMessage    string
	UserMessage      string
	UserFiles        []string

	Tools []string

	Response_format string

	Max_iteration int

	Out_StatusCode   int
	Out_messages     []byte //[]*ChatMsg
	Out_last_message string

	delta func(msg *ChatMsg)
}

func (st *LLMxAICompleteChat) run(caller *ToolCaller, ui *UI) error {

	source_llm, err := NewLLMxAI("", caller)
	if err != nil {
		return err
	}

	source_llm.Check(caller)

	/*{
		var msgs ChatMsgs
		json.Unmarshal(st.PreviousMessages, &msgs)
		m1 := msgs.AddUserMessage(st.UserMessage, st.UserFiles)
		if st.delta != nil {
			st.delta(m1)
		}

		i := 0
		N := 60
		for i < N {
			if !caller.Progress(float64(i)/float64(N), "completing") {
				fmt.Println("completion interrupted", caller.msg_id)
				return nil
			}

			//fmt.Println("added", st.UserMessage)
			{
				m2 := msgs.AddAssistentCalls("", "response-"+st.UserMessage+"_"+strconv.Itoa(i), nil, ChatMsgUsage{}, 0, "xai", "grok-3-mini")
				if st.delta != nil {
					st.delta(m2)
				}
			}

			time.Sleep(1 * time.Second)
			i++
		}

		st.Out_messages, _ = json.Marshal(msgs)
		return nil
	}*/

	//Tools
	var tools []*OpenAI_completion_tool
	{
		slices.Sort(st.Tools)
		st.Tools = slices.Compact(st.Tools) //remove duplicities

		js := callFuncGetToolsShemas(st.Tools)
		if len(js) > 0 {
			err := json.Unmarshal(js, &tools)
			if err != nil {
				return err
			}
		}
	}

	//Messages
	var msgs ChatMsgs
	err = json.Unmarshal(st.PreviousMessages, &msgs)
	if err != nil {
		return err
	}

	if st.UserMessage != "" || len(st.UserFiles) > 0 {
		m1, err := msgs.AddUserMessage(st.UserMessage, st.UserFiles)
		if err != nil {
			return err
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

	last_msg := ""

	max_iter := st.Max_iteration
	if max_iter <= 0 {
		max_iter = 20
	}
	iter := 0
	for iter < max_iter {
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
			Model: st.Model,

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

		fnStreaming := func(msg *ChatMsg) bool {

			msg.Provider = source_llm.Provider
			msg.Model = st.Model
			msg.Seed = seed
			msg.Stream = true
			msg.ShowParameters = true
			msg.ShowReasoning = true

			st.delta(msg)

			return caller.Progress(0, "completing")
		}

		//print
		{
			js, err := json.MarshalIndent(props, "", "  ")
			if err == nil {
				callFuncPrint("---\n" + string(js) + "---\n")
			}
		}

		jsProps, err := json.Marshal(props)
		if err != nil {
			return err
		}
		out, status, dt, time_to_first_token, err := OpenAI_completion_Run(jsProps, source_llm.OpenAI_url, source_llm.API_key, fnStreaming)
		st.Out_StatusCode = status
		if err != nil {
			return err
		}

		if !caller.Progress(0, "completing") {
			return nil
		}

		if len(out.Choices) > 0 {

			var usage ChatMsgUsage
			{
				usage.Prompt_tokens = out.Usage.Prompt_tokens
				usage.Input_cached_tokens = out.Usage.Input_cached_tokens
				usage.Completion_tokens = out.Usage.Completion_tokens
				usage.Reasoning_tokens = out.Usage.Completion_tokens_details.Reasoning_tokens
				mod, _ := source_llm.FindProviderModel(st.Model)
				if mod != nil {
					usage.Prompt_price, usage.Reasoning_price, usage.Input_cached_price, usage.Completion_price = mod.GetTextPrice(usage.Prompt_tokens, usage.Reasoning_tokens, usage.Input_cached_tokens, usage.Completion_tokens)
				}
			}

			calls := out.Choices[0].Message.Tool_calls
			m2 := msgs.AddAssistentCalls(out.Choices[0].Message.Reasoning_content, out.Choices[0].Message.Content, calls, usage, dt, time_to_first_token, source_llm.Provider, st.Model)
			if st.delta != nil {
				st.delta(m2)
			}

			last_msg = out.Choices[0].Message.Content

			for _, call := range calls {
				caller.source_structs = nil //reset to get new list of sources

				var result string
				tool_ui, ress, err := CallToolByName(call.Function.Name, []byte(call.Function.Arguments), caller)
				if err == nil {

					resJs, err := json.Marshal(ress)
					if err != nil {
						return err
					}
					var res map[string]interface{}
					err = json.Unmarshal(resJs, &res)
					if err != nil {
						return err
					}

					//Out_ -> result
					{
						num_outs := 0
						for nm := range res {
							if strings.HasPrefix(strings.ToLower(nm), "out") {
								num_outs++
							}
						}
						for nm, val := range res {
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
				} else {
					result = "Error: " + err.Error()
				}

				hasUI := (tool_ui != nil && tool_ui.Is())
				if hasUI {
					if result != "" {
						result += "\n"
					}
					result += "Successfully shown on screen."
				}

				res_msg := msgs.AddCallResult(call.Function.Name, call.Id, result)
				if hasUI {
					res_msg.UI_func = call.Function.Name
					res_msg.UI_params = ress
				}
				if st.delta != nil {
					st.delta(res_msg)
				}

				//add sources(icons for tabs)
				res_msg.Sources_structs = caller.source_structs
				slices.Sort(res_msg.Sources_structs)
				res_msg.Sources_structs = slices.Compact(res_msg.Sources_structs)
			}

			if len(calls) == 0 {
				break
			}
		}
		iter++
	}

	st.Out_last_message = last_msg

	st.Out_messages, err = json.Marshal(msgs)
	if err != nil {
		return err
	}

	return nil
}

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
	Choices []OpenAIOutChoice
	Usage   OpenAIOut_Usage
	Error   *OpenAIOutError
}

type OpenAI_completion_format struct {
	Type string `json:"type"` //json_object
	//Json_schema
}

type OpenAI_completion_msgSystem struct {
	Role    string `json:"role"` //"system"
	Content string `json:"content"`
}

type OpenAI_completion_tool_function_parameters_properties struct {
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
}

type OpenAI_completion_Stream_options struct {
	Include_usage bool `json:"include_usage"`
}
type OpenAI_completion_props struct {
	Seed  int    `json:"seed"`
	Model string `json:"model"`

	Tools          []*OpenAI_completion_tool        `json:"tools,omitempty"`
	Messages       []interface{}                    `json:"messages"` //OpenAI_completion_msgSystem, OpenAI_completion_msg_Content, OpenAI_completion_msgCalls, OpenAI_completion_msgCalls,
	Stream         bool                             `json:"stream"`
	Stream_options OpenAI_completion_Stream_options `json:"stream_options"`

	Temperature       float64 `json:"temperature"`                 //1.0
	Max_tokens        int     `json:"max_tokens"`                  //
	Top_p             float64 `json:"top_p"`                       //1.0
	Frequency_penalty float64 `json:"frequency_penalty,omitempty"` //0
	Presence_penalty  float64 `json:"presence_penalty,omitempty"`  //0

	Response_format *OpenAI_completion_format `json:"response_format,omitempty"`

	Reasoning_effort string `json:"reasoning_effort,omitempty"` //"low", "high"
}

var g_global_OpenAI_completion_lock sync.Mutex

func OpenAI_completion_Run(jsProps []byte, Completion_url string, api_key string, fnStreaming func(msg *ChatMsg) bool) (OpenAIOut, int, float64, float64, error) {
	g_global_OpenAI_completion_lock.Lock()
	defer g_global_OpenAI_completion_lock.Unlock()

	st := time.Now().UnixMicro()

	if !strings.HasSuffix(Completion_url, "/") {
		Completion_url += "/"
	}
	Completion_url += "chat/completions"

	body := bytes.NewReader(jsProps)

	req, err := http.NewRequest(http.MethodPost, Completion_url, body)
	if err != nil {
		return OpenAIOut{}, -1, 0, -1, fmt.Errorf("NewRequest() failed: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+api_key)
	if fnStreaming != nil {
		req.Header.Add("Accept", "text/event-stream")
		req.Header.Add("Cache-Control", "no-cache")
		req.Header.Add("Connection", "keep-alive")
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return OpenAIOut{}, -1, 0, -1, fmt.Errorf("Do() failed: %w", err)
	}
	defer res.Body.Close()

	var ret OpenAIOut

	time_to_first_token := -1.0

	if fnStreaming != nil {
		// Check response status
		if res.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(res.Body)
			return OpenAIOut{}, -1, 0, -1, fmt.Errorf("unexpected status code: %d, body: %s", res.StatusCode, string(body))
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
				return OpenAIOut{}, -1, 0, -1, fmt.Errorf("error reading stream: %w", err)
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

			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				return OpenAIOut{}, -1, 0, -1, fmt.Errorf("failed to parse stream response: %w", err)
			}

			if streamResp.Usage.Total_tokens > 0 {
				ret.Usage = streamResp.Usage
			}

			for _, choice := range streamResp.Choices {
				if len(ret.Choices) == 0 {
					ret.Choices = append(ret.Choices, OpenAIOutChoice{})
				}

				if time_to_first_token < 0 && choice.Delta.Content != "" {
					time_to_first_token = float64(time.Now().UnixMicro()-st) / 1000000
				}

				//add content
				ret.Choices[0].Message.Content += choice.Delta.Content

				//reasoning content
				ret.Choices[0].Message.Reasoning_content += choice.Delta.Reasoning_content

				//add tools
				ret.Choices[0].Message.Tool_calls = append(ret.Choices[0].Message.Tool_calls, choice.Delta.Tool_calls...)

				//callback
				var msgs ChatMsgs
				msgs.AddAssistentCalls(ret.Choices[0].Message.Reasoning_content, ret.Choices[0].Message.Content, ret.Choices[0].Message.Tool_calls, ChatMsgUsage{}, float64(time.Now().UnixMicro()-st)/1000000, time_to_first_token, "", "")
				if !fnStreaming(msgs.Messages[0]) {
					streaming_ok = false
					break //interrupted
				}
			}
		}

	} else {
		js, err := io.ReadAll(res.Body)
		if err != nil {
			return OpenAIOut{}, res.StatusCode, 0, -1, err
		}

		if res.StatusCode != http.StatusOK {
			return OpenAIOut{}, res.StatusCode, 0, -1, fmt.Errorf("statusCode %d != %d, response: %s", res.StatusCode, http.StatusOK, string(js))
		}

		if len(js) == 0 {
			return OpenAIOut{}, res.StatusCode, 0, -1, fmt.Errorf("output is empty")
		}

		err = json.Unmarshal(js, &ret)
		if err != nil {
			return OpenAIOut{}, res.StatusCode, 0, -1, fmt.Errorf("%w. %s", err, string(js))
		}

		if ret.Error != nil && ret.Error.Message != "" {
			return OpenAIOut{}, res.StatusCode, 0, -1, errors.New(ret.Error.Message)
		}
	}

	dt := float64(time.Now().UnixMicro()-st) / 1000000
	return ret, res.StatusCode, dt, time_to_first_token, nil
}
