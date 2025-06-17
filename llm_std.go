package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
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

				if time_to_first_token < 0 && (choice.Delta.Content != "" || choice.Delta.Reasoning_content != "") {
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
				msgs.AddAssistentCalls(ret.Choices[0].Message.Reasoning_content, ret.Choices[0].Message.Content, ret.Choices[0].Message.Tool_calls, LLMMsgUsage{}, float64(time.Now().UnixMicro()-st)/1000000, time_to_first_token, "", "")
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

var g_global_OpenAI_genImage_lock sync.Mutex

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
	g_global_OpenAI_genImage_lock.Lock()
	defer g_global_OpenAI_genImage_lock.Unlock()

	st := time.Now().UnixMicro()

	if !strings.HasSuffix(Completion_url, "/") {
		Completion_url += "/"
	}
	Completion_url += "images/generations"

	body := bytes.NewReader(jsProps)

	req, err := http.NewRequest(http.MethodPost, Completion_url, body)
	if err != nil {
		return OpenAIGenImageOut{}, -1, 0, fmt.Errorf("NewRequest() failed: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+api_key)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return OpenAIGenImageOut{}, -1, 0, fmt.Errorf("Do() failed: %w", err)
	}
	defer res.Body.Close()

	js, err := io.ReadAll(res.Body)
	if err != nil {
		return OpenAIGenImageOut{}, res.StatusCode, 0, err
	}

	if len(js) == 0 {
		return OpenAIGenImageOut{}, res.StatusCode, 0, fmt.Errorf("output is empty")
	}

	var out OpenAIGenImageOut
	err = json.Unmarshal(js, &out)
	if err != nil {
		return OpenAIGenImageOut{}, res.StatusCode, 0, fmt.Errorf("%w. %s", err, string(js))
	}
	if out.Error != nil && out.Error.Message != "" {
		return OpenAIGenImageOut{}, res.StatusCode, 0, errors.New(out.Error.Message)
	}

	if res.StatusCode != 200 {
		return OpenAIGenImageOut{}, res.StatusCode, 0, fmt.Errorf("statusCode %d != 200, response: %s", res.StatusCode, string(js))
	}

	tm := float64(time.Now().UnixMicro()-st) / 1000000

	return out, res.StatusCode, tm, nil
}
