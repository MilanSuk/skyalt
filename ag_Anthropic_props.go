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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Anthropic_completion_props struct {
	Model string `json:"model"`

	System   string                     `json:"system"`
	Messages []Anthropic_completion_msg `json:"messages"`
	Stream   bool                       `json:"stream"`

	Tools []*Anthropic_completion_tool `json:"tools,omitempty"`

	Temperature float64 `json:"temperature"` //1.0
	Max_tokens  int     `json:"max_tokens"`
}

type Anthropic_completion_msg_content_Image struct {
	Type       string `json:"type"`                 //"base64"
	Media_type string `json:"media_type,omitempty"` //"image/jpeg"
	Data       string `json:"Data,omitempty"`
}

type Anthropic_completion_msg_Content struct {
	Type string `json:"type"` //"image", "text", "tool_use", "tool_result"
	Text string `json:"text,omitempty"`

	//tool result
	Tool_use_id string `json:"tool_use_id,omitempty"`
	Content     string `json:"content,omitempty"`
	//'Name' is also set - OpenAI Mistral wants it

	//tool call
	Id    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`  //"get_weather"
	Input json.RawMessage `json:"input,omitempty"` //{"location": "San Francisco, CA", "unit": "celsius"}

	//image
	Source *Anthropic_completion_msg_content_Image `json:"source,omitempty"`
}

type Anthropic_completion_msg struct {
	Role    string                             `json:"role"` //"user", "assistant", note: "system" is not here, it's top level: "props"
	Content []Anthropic_completion_msg_Content `json:"content"`
}

func (msg *Anthropic_completion_msg) AddText(str string) {
	msg.Content = append(msg.Content, Anthropic_completion_msg_Content{Type: "text", Text: str})
}

func (msg *Anthropic_completion_msg) AddImage(data []byte, media_type string) { //ext="image/png","image/jpeg", "image/webp", "image/gif"(non-animated)
	bs64 := base64.StdEncoding.EncodeToString(data)
	msg.Content = append(msg.Content, Anthropic_completion_msg_Content{Type: "image", Source: &Anthropic_completion_msg_content_Image{Type: "base64", Media_type: media_type, Data: bs64}})
}
func (msg *Anthropic_completion_msg) AddImageFile(path string) error {
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

func (msg *Anthropic_completion_msg) AddToolResult(tool_name string, tool_use_id string, result string) {
	msg.Content = append(msg.Content, Anthropic_completion_msg_Content{Type: "tool_result", Tool_use_id: tool_use_id, Content: result, Name: tool_name})
}

type Anthropic_completion_tool struct {
	Name         string                        `json:"name"`
	Description  string                        `json:"description"`
	Input_schema OpenAI_completion_tool_schema `json:"input_schema"`
}

func NewAnthropic_completion_tool(name, description string) *Anthropic_completion_tool {
	fn := &Anthropic_completion_tool{Name: name, Description: description}
	fn.Input_schema.Type = "object"
	fn.Input_schema.AdditionalProperties = false
	fn.Input_schema.Properties = make(map[string]*OpenAI_completion_tool_function_parameters_properties)
	return fn
}
