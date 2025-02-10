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
	"strings"
	"sync"
)

type AnthropicOut_Error struct {
	Message string
}
type AnthropicOut_Usage struct {
	Cache_creation_input_tokens int //Number of tokens written to the cache when creating a new entry.
	Cache_read_input_tokens     int //Number of tokens retrieved from the cache for this request.

	Input_tokens  int //Number of input tokens which were not read from or used to create a cache.
	Output_tokens int
}

type AnthropicOut struct {
	Role    string
	Content []Anthropic_completion_msg_Content
	Error   *AnthropicOut_Error
	Usage   AnthropicOut_Usage
}

var g_globa_Anthropic_completion_lock sync.Mutex

func Anthropic_completion_Run(input Anthropic_completion_props, Completion_url string, api_key string) (AnthropicOut, int, error) {
	g_globa_Anthropic_completion_lock.Lock()
	defer g_globa_Anthropic_completion_lock.Unlock()

	if !strings.HasSuffix(Completion_url, "/") {
		Completion_url += "/"
	}
	Completion_url += "messages"

	jsProps, err := json.Marshal(input)
	if err != nil {
		return AnthropicOut{}, -1, err
	}
	body := bytes.NewReader(jsProps)

	req, err := http.NewRequest(http.MethodPost, Completion_url, body)
	if err != nil {
		return AnthropicOut{}, -1, fmt.Errorf("NewRequest() failed: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-api-key", api_key)
	req.Header.Add("anthropic-version", "2023-06-01")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return AnthropicOut{}, -1, fmt.Errorf("Do() failed: %w", err)
	}
	defer res.Body.Close()

	js, err := io.ReadAll(res.Body)
	if err != nil {
		return AnthropicOut{}, res.StatusCode, err
	}

	var out AnthropicOut
	err = json.Unmarshal(js, &out)
	if err != nil {
		return AnthropicOut{}, res.StatusCode, err
	}
	if out.Error != nil && out.Error.Message != "" {
		return AnthropicOut{}, res.StatusCode, errors.New(out.Error.Message)
	}
	if res.StatusCode != 200 {
		return AnthropicOut{}, res.StatusCode, fmt.Errorf("statusCode %d != 200, response: %s", res.StatusCode, string(js))
	}
	return out, res.StatusCode, nil
}
