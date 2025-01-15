package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (layout *Layout) AddOpenAI_completion_props(x, y, w, h int, props *OpenAI_completion_props) *OpenAI_completion_props {
	layout._createDiv(x, y, w, h, "OpenAI_completion_props", props.Build, nil, nil)
	return props
}

type OpenAI_completion_props struct {
	Model    string                  `json:"model"`
	Messages []OpenAI_completion_msg `json:"messages"`
	Stream   bool                    `json:"stream"`

	Tools []*OpenAI_completion_tool `json:"tools,omitempty"`

	Temperature       float64 `json:"temperature"`       //1.0
	Max_tokens        int     `json:"max_tokens"`        //
	Top_p             float64 `json:"top_p"`             //1.0
	Frequency_penalty float64 `json:"frequency_penalty"` //0
	Presence_penalty  float64 `json:"presence_penalty"`  //0

	Response_format *OpenAI_completion_format `json:"response_format"`
}

type OpenAI_completion_tool_function_parameters_properties struct {
	Type        string   `json:"type"` //"number", "string"
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default,omitempty"`
}
type OpenAI_completion_tool_function_parameters struct {
	Type                 string
	Properties           map[string]OpenAI_completion_tool_function_parameters_properties
	Required             []string
	AdditionalProperties bool
}
type OpenAI_completion_tool_function struct {
	Name        string                                     `json:"name"`
	Description string                                     `json:"description"`
	Parameters  OpenAI_completion_tool_function_parameters `json:"parameters"`
	Strict      bool                                       `json:"strict"`
}

func (prm *OpenAI_completion_tool_function) AddParam(name, typee, description string) {
	prm.Parameters.Properties[name] = OpenAI_completion_tool_function_parameters_properties{Type: typee, Description: description}
	prm.Parameters.Required = append(prm.Parameters.Required, name)
}

type OpenAI_completion_tool struct {
	Type     string                          `json:"type"` //"object"
	Function OpenAI_completion_tool_function `json:"function"`
}

func (props *OpenAI_completion_props) AddToolFunc(name, description string) *OpenAI_completion_tool_function {
	tool := &OpenAI_completion_tool{Type: "function"}

	tool.Function = OpenAI_completion_tool_function{Name: name, Description: description, Strict: true}
	tool.Function.Parameters.Type = "object"
	tool.Function.Parameters.AdditionalProperties = false

	props.Tools = append(props.Tools, tool)
	return &tool.Function
}

type OpenAI_completion_msg_Content_ToolCall_Function struct {
	Name      string                 `json:"name,omitempty"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
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

	Tool_calls []OpenAI_completion_msg_Content_ToolCall `json:"tool_calls,omitempty"`
}
type OpenAI_completion_msg struct {
	Role    string                          `json:"role"` //"system", "user", "assistant"
	Content []OpenAI_completion_msg_Content `json:"content"`
}

func (msg *OpenAI_completion_msg) AddText(str string) {
	msg.Content = append(msg.Content, OpenAI_completion_msg_Content{Type: "text", Text: str})
}
func (msg *OpenAI_completion_msg) AddImage(data []byte, ext string) { //ext="png","jpeg", "webp", "gif"(non-animated)
	prefix := "data:image/" + ext + ";base64,"
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

	msg.AddImage(data, ext)
	return nil
}

type OpenAI_completion_format struct {
	Type string `json:"type"` //json_object
	//Json_schema ...
}

func (st *OpenAI_completion_props) Build(layout *Layout) {

	layout.SetColumn(0, 2, 3.5)
	layout.SetColumn(1, 1, 100)

	y := 0

	layout.AddText(0, y, 1, 1, "Model")
	layout.AddCombo(1, y, 1, 1, &st.Model, OpenAI_GetChatModelList(), OpenAI_GetChatModelList())
	y++

	tx := layout.AddText(0, y, 1, 1, "Streaming")
	tx.Tooltip = "See result as is generated."
	layout.AddSwitch(1, y, 1, 1, "", &st.Stream)
	y++

	sl := layout.AddSliderEdit(0, y, 2, 1, &st.Temperature, 0, 2, 0.1)
	sl.ValuePointerPrec = 1
	sl.Description = "Temperature"
	sl.Tooltip = "The sampling temperature, between 0 and 1. Higher values like 0.8 will make the output more random, while lower values like 0.2 will make it more focused and deterministic. If set to 0, the model will use log probability to automatically increase the temperature until certain thresholds are hit."
	sl.Description_width = 3.5
	sl.Edit_width = 2
	sl.Legend = true
	y++

	sl = layout.AddSliderEditInt(0, y, 2, 1, &st.Max_tokens, 128, 4096, 1)
	sl.ValuePointerPrec = 0
	sl.Description = "Max Tokens"
	sl.Tooltip = "The maximum number of tokens that can be generated in the chat completion. The total length of input tokens and generated tokens is limited by the model's context length."
	sl.Description_width = 3.5
	sl.Edit_width = 2
	sl.Legend = true
	y++

	sl = layout.AddSliderEdit(0, y, 2, 1, &st.Top_p, 0, 1, 0.1)
	sl.ValuePointerPrec = 1
	sl.Description = "Top P"
	sl.Tooltip = "An alternative to sampling with temperature, called nucleus sampling, where the model considers the results of the tokens with top_p probability mass. So 0.1 means only the tokens comprising the top 10% probability mass are considered."
	sl.Description_width = 3.5
	sl.Edit_width = 2
	sl.Legend = true
	y++

	sl = layout.AddSliderEdit(0, y, 2, 1, &st.Frequency_penalty, -2, 2, 0.1)
	sl.ValuePointerPrec = 1
	sl.Description = "Frequency Penalty"
	sl.Tooltip = "Number between -2.0 and 2.0. Positive values penalize new tokens based on their existing frequency in the text so far, decreasing the model's likelihood to repeat the same line verbatim."
	sl.Description_width = 3.5
	sl.Edit_width = 2
	sl.Legend = true
	y++

	sl = layout.AddSliderEdit(0, y, 2, 1, &st.Presence_penalty, -2, 2, 0.1)
	sl.ValuePointerPrec = 1
	sl.Description = "Presence Penalty"
	sl.Tooltip = "Number between -2.0 and 2.0. Positive values penalize new tokens based on whether they appear in the text so far, increasing the model's likelihood to talk about new topics."
	sl.Description_width = 3.5
	sl.Edit_width = 2
	sl.Legend = true
	y++

	ResetBt := layout.AddButton(0, y, 1, 1, "Reset")
	ResetBt.Background = 0.5
	ResetBt.clicked = func() {
		st.Reset()
	}
	y++
}

func (props *OpenAI_completion_props) Reset() {
	if props.Model == "" {
		props.Model = OpenFile_OpenAI().ChatModel
	}
	props.Stream = true
	props.Temperature = 1.0
	props.Max_tokens = 4046
	props.Top_p = 0.7 //1.0
	props.Frequency_penalty = 0
	props.Presence_penalty = 0
	//props.Seed = -1
	//props.Cache_prompt = false
	//props.Stop = []string{"</s>", "<|im_start|>", "<|im_end|>", "Llama:", "User:"}
}
