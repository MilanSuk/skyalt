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
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type OpenAI_completion_out struct {
	Citations  []string
	Content    string
	Tool_calls []OpenAI_completion_msg_Content_ToolCall
}

type OpenAIOutChoice struct {
	Message OpenAI_completion_out
	//Delta   STMsg
}
type OpenAIOutUsage struct {
	Prompt_tokens       int
	Input_cached_tokens int
	Completion_tokens   int
	Total_tokens        int
}
type OpenAIOutError struct {
	Message string
}
type OpenAIOut struct {
	Citations []string
	Choices   []OpenAIOutChoice
	Usage     OpenAIOutUsage
	Error     *OpenAIOutError
}

var g_globa_OpenAI_completion_lock sync.Mutex

func OpenAI_completion_Run(input OpenAI_completion_props, Completion_url string, api_key string) (OpenAIOut, int, error) {
	g_globa_OpenAI_completion_lock.Lock()
	defer g_globa_OpenAI_completion_lock.Unlock()

	if !strings.HasSuffix(Completion_url, "/") {
		Completion_url += "/"
	}
	Completion_url += "chat/completions"

	jsProps, err := json.Marshal(input)
	if err != nil {
		return OpenAIOut{}, -1, err
	}
	body := bytes.NewReader(jsProps)

	req, err := http.NewRequest(http.MethodPost, Completion_url, body)
	if err != nil {
		return OpenAIOut{}, -1, fmt.Errorf("NewRequest() failed: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+api_key)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return OpenAIOut{}, -1, fmt.Errorf("Do() failed: %w", err)
	}
	defer res.Body.Close()

	js, err := io.ReadAll(res.Body)
	if err != nil {
		return OpenAIOut{}, res.StatusCode, err
	}

	if len(js) == 0 {
		return OpenAIOut{}, res.StatusCode, fmt.Errorf("output is empty")
	}

	var out OpenAIOut
	err = json.Unmarshal(js, &out)
	if err != nil {
		return OpenAIOut{}, res.StatusCode, fmt.Errorf("%w. %s", err, string(js))
	}
	if out.Error != nil && out.Error.Message != "" {
		return OpenAIOut{}, res.StatusCode, errors.New(out.Error.Message)
	}

	if res.StatusCode != 200 {
		return OpenAIOut{}, res.StatusCode, fmt.Errorf("statusCode %d != 200, response: %s", res.StatusCode, string(js))
	}

	if len(out.Choices) > 0 {
		if out.Choices[0].Message.Content == "<|separator|>" {
			fmt.Println("<|separator|>", string(js))
		}
	}

	return out, res.StatusCode, nil
}

func OpenAI_stt_Run(model string, file_name string, blob []byte, temperature float64, response_format string, Completion_url string, api_key string) ([]byte, int, error) {
	g_globa_OpenAI_completion_lock.Lock()
	defer g_globa_OpenAI_completion_lock.Unlock()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if !strings.HasSuffix(Completion_url, "/") {
		Completion_url += "/"
	}
	Completion_url += "audio/transcriptions"

	//set parameters
	{
		part, err := writer.CreateFormFile("file", file_name)
		if err != nil {
			return nil, -1, fmt.Errorf("CreateFormFile() failed: %w", err)
		}
		part.Write(blob)

		writer.WriteField("temperature", strconv.FormatFloat(temperature, 'f', -1, 64))
		writer.WriteField("response_format", response_format)
		//props.Write(writer)

		if model != "" {
			writer.WriteField("model", model)

			if response_format == "verbose_json" && strings.Contains(Completion_url, "api.openai.com") {
				writer.WriteField("timestamp_granularities[]", "word")
				writer.WriteField("timestamp_granularities[]", "segment")
			}
		}
	}
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, Completion_url, body)
	if err != nil {
		return nil, -1, fmt.Errorf("NewRequest() failed: %w", err)
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	if api_key != "" {
		req.Header.Add("Authorization", "Bearer "+api_key)
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, -1, fmt.Errorf("Do() failed: %w", err)
	}

	resBody, err := io.ReadAll(res.Body) //job.close ...
	if err != nil {
		return nil, res.StatusCode, fmt.Errorf("ReadAll() failed: %w", err)
	}

	if res.StatusCode != 200 {
		return nil, res.StatusCode, fmt.Errorf("statusCode %d != 200, response: %s", res.StatusCode, string(resBody))
	}

	return resBody, res.StatusCode, nil
}

func OpenAI_tts_Run(text string, model string, voice string, Completion_url string, api_key string) ([]byte, int, error) {
	g_globa_OpenAI_completion_lock.Lock()
	defer g_globa_OpenAI_completion_lock.Unlock()

	if !strings.HasSuffix(Completion_url, "/") {
		Completion_url += "/"
	}
	Completion_url += "audio/speech"

	type TTS struct {
		Model string `json:"model"`
		Input string `json:"input"`
		Voice string `json:"voice"`
	}

	tts := TTS{Model: model, Voice: voice, Input: text}
	js, err := json.Marshal(tts)
	if err != nil {
		return nil, -1, fmt.Errorf("Marshal() failed: %w", err)
	}
	body := bytes.NewReader(js)

	req, err := http.NewRequest(http.MethodPost, Completion_url, body)
	if err != nil {
		return nil, -1, fmt.Errorf("NewRequest() failed: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+api_key)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, -1, fmt.Errorf("Do() failed: %w", err)
	}
	defer res.Body.Close()

	out, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, res.StatusCode, fmt.Errorf("ReadAll() failed: %w", err)
	}

	if res.StatusCode != 200 {
		return nil, res.StatusCode, fmt.Errorf("statusCode %d != 200, response: %s", res.StatusCode, string(js))
	}

	return out, res.StatusCode, nil
}
