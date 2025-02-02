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
	"net/http"
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

func OpenAI_completion_Run(input OpenAI_completion_props, Completion_url string, api_key string) (OpenAIOut, error) {
	jsProps, err := json.Marshal(input)
	if err != nil {
		return OpenAIOut{}, err
	}
	body := bytes.NewReader(jsProps)

	req, err := http.NewRequest(http.MethodPost, Completion_url, body)
	if err != nil {
		return OpenAIOut{}, fmt.Errorf("NewRequest() failed: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+api_key)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return OpenAIOut{}, fmt.Errorf("Do() failed: %w", err)
	}
	defer res.Body.Close()

	js, err := io.ReadAll(res.Body)
	if err != nil {
		return OpenAIOut{}, err
	}

	if len(js) == 0 {
		return OpenAIOut{}, fmt.Errorf("output is empty")
	}

	var out OpenAIOut
	err = json.Unmarshal(js, &out)
	if err != nil {
		return OpenAIOut{}, fmt.Errorf("%w. %s", err, string(js))
	}
	if out.Error != nil && out.Error.Message != "" {
		return OpenAIOut{}, errors.New(out.Error.Message)
	}

	if res.StatusCode != 200 {
		return OpenAIOut{}, fmt.Errorf("statusCode %d != 200, response: %s", res.StatusCode, string(js))
	}

	return out, nil
}
