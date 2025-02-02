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
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type OpenAI_completion_props struct {
	Model    string        `json:"model"`
	Messages []interface{} `json:"messages"` //OpenAI_completion_msgPlain, OpenAI_completion_msg, OpenAI_completion_msgCalls, OpenAI_completion_msgResult
	Stream   bool          `json:"stream"`

	Tools []*OpenAI_completion_tool `json:"tools,omitempty"`

	Temperature       float64 `json:"temperature"`                 //1.0
	Max_tokens        int     `json:"max_tokens"`                  //
	Top_p             float64 `json:"top_p"`                       //1.0
	Frequency_penalty float64 `json:"frequency_penalty,omitempty"` //0
	Presence_penalty  float64 `json:"presence_penalty,omitempty"`  //0

	Response_format *OpenAI_completion_format `json:"response_format,omitempty"`

	Reasoning_effort float64 `json:"reasoning_effort,omitempty"` //"low", "medium", "high"

}

type OpenAI_completion_tool_function_parameters_properties struct {
	Type        string   `json:"type"` //"number", "string"
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default,omitempty"`
}
type OpenAI_completion_tool_schema struct {
	Type                 string   `json:"type"` //"object"
	Required             []string `json:"required,omitempty"`
	AdditionalProperties bool     `json:"additionalProperties"`

	Properties map[string]*OpenAI_completion_tool_function_parameters_properties `json:"properties"`
	//Properties  string                                                            `json:"properties"`
	//properties2 map[string]*OpenAI_completion_tool_function_parameters_properties //fake
}
type OpenAI_completion_tool_function struct {
	Name        string                        `json:"name"`
	Description string                        `json:"description"`
	Parameters  OpenAI_completion_tool_schema `json:"parameters"`
	Strict      bool                          `json:"strict"`
}

func (prm *OpenAI_completion_tool_schema) AddParam(name, typee, description string) *OpenAI_completion_tool_function_parameters_properties {
	if strings.Contains(strings.ToLower(typee), "float") || strings.Contains(strings.ToLower(typee), "int") {
		typee = "number"
	}

	prm.Required = append(prm.Required, name)

	p := &OpenAI_completion_tool_function_parameters_properties{Type: typee, Description: description}
	prm.Properties[name] = p

	//dirty hack - Xai(anthropic api) wants attribute .Properties as string, not map ...
	/*if prm.properties2 == nil {
		prm.properties2 = make(map[string]*OpenAI_completion_tool_function_parameters_properties)
	}
	prm.properties2[name] = p
	js, err := json.Marshal(prm.properties2)
	if err == nil {
		prm.Properties = string(js)
	}*/

	return p
}

func NewOpenAI_completion_tool(name, description string) *OpenAI_completion_tool {
	fn := &OpenAI_completion_tool{Type: "function"}
	fn.Function = OpenAI_completion_tool_function{Name: name, Description: description, Strict: false}
	fn.Function.Parameters.Type = "object"
	fn.Function.Parameters.AdditionalProperties = false
	fn.Function.Parameters.Properties = make(map[string]*OpenAI_completion_tool_function_parameters_properties)
	return fn
}

type OpenAI_completion_tool struct {
	Type     string                          `json:"type"` //"object"
	Function OpenAI_completion_tool_function `json:"function"`
}

func (props *OpenAI_completion_props) AddToolFunc(name, description string) *OpenAI_completion_tool_function {
	tool := NewOpenAI_completion_tool(name, description)
	props.Tools = append(props.Tools, tool)
	return &tool.Function
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

type OpenAI_completion_msg_Content_Image_url struct {
	Detail string `json:"detail,omitempty"` //"low", "high", "auto"
	Url    string `json:"url,omitempty"`    //"data:image/jpeg;base64,<base64_image_string>"
}
type OpenAI_completion_msg_Content struct {
	Type      string                                   `json:"type"` //"image_url", "text"
	Text      string                                   `json:"text,omitempty"`
	Image_url *OpenAI_completion_msg_Content_Image_url `json:"image_url,omitempty"`
}

type OpenAI_completion_msgPlain struct {
	Role    string `json:"role"` //"system"
	Content string `json:"content"`
}

type OpenAI_completion_msg struct {
	Role    string                          `json:"role"` //"system", "user", "assistant", "tool"
	Content []OpenAI_completion_msg_Content `json:"content"`
}

type OpenAI_completion_msgCalls struct {
	Role       string                                   `json:"role"` //"system", "user", "assistant", "tool"
	Content    string                                   `json:"content"`
	Tool_calls []OpenAI_completion_msg_Content_ToolCall `json:"tool_calls,omitempty"`
}

type OpenAI_completion_msgResult struct {
	Role         string `json:"role"` //"system", "user", "assistant", "tool"
	Content      string `json:"content"`
	Tool_call_id string `json:"tool_call_id,omitempty"`
	Name         string `json:"name,omitempty"` //Tool name: Mistral wants this
}

func (msg *OpenAI_completion_msg) AddText(str string) {
	msg.Content = append(msg.Content, OpenAI_completion_msg_Content{Type: "text", Text: str})
}
func (msg *OpenAI_completion_msg) AddImage(data []byte, media_type string) { //ext="image/png","image/jpeg", "image/webp", "image/gif"(non-animated)
	prefix := "data:" + media_type + ";base64,"
	bs64 := base64.StdEncoding.EncodeToString(data)
	msg.Content = append(msg.Content, OpenAI_completion_msg_Content{Type: "image_url", Image_url: &OpenAI_completion_msg_Content_Image_url{Detail: "high", Url: prefix + bs64}})
}
func (msg *OpenAI_completion_msg) AddImageFile(path string) error {
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

type OpenAI_completion_format struct {
	Type string `json:"type"` //json_object
	//Json_schema ...
}
