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
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

var g_agent_server *NetServer
var g_agent_passwords *Secrets

func Agent_initGlobal() {
	g_agent_server = NewNetServer(8090)
	g_agent_passwords = NewSecrets()
}
func Agent_destroyGlobal() {
	g_agent_server.Destroy()
	g_agent_passwords.Destroy()
}

type AgentMsg struct {
	CreatedTimeSec float64
	CreatedBy      string //name of service(AI), empty = user wrote it

	Role    string
	Content []Anthropic_completion_msg_Content

	//stats
	InputTokens       int
	InputCachedTokens int
	OutputTokens      int
	Time              float64
}

type Agent struct {
	Description string //Chat label

	Use_case string
	Folder   string

	Input ChatInput

	PresetSystemPrompt string
	Messages           []AgentMsg

	Call_id              string
	SubAgents            []*Agent
	Selected_sub_call_id string

	ShowToolParams []string //Call_id

	Sandbox_violations []string

	Stopped string

	NoTools bool
}

func NewAgent(folder string, use_case string, systemPrompt string) *Agent {
	agent := &Agent{Folder: folder, Use_case: use_case, PresetSystemPrompt: systemPrompt}
	return agent
}

func (agent *Agent) getFolder() string {
	folder := agent.Folder
	if folder == "" {
		folder = "tools"
	}
	return folder
}
func (agent *Agent) getUseCase() string {
	use_case := agent.Use_case
	if use_case == "" {
		use_case = "main"
	}
	return use_case
}

func (agent *Agent) getSystemPrompt() string {
	prompt := agent.PresetSystemPrompt
	if prompt == "" {
		prompt = `You are an AI tool calling assistant, who enjoys precision and carefully follows the user's requirements."

If you can not find the right tool, use the tool 'create_new_tool'. If there is some problem with a tool(for example, a bug) then use the tool 'update_tool'.
Don't ask to use, change, or create a tool, just do it! Call tools sequentially. Avoid tool call as parameter value.

User informations, device settings and files are store on the disk, read them with 'access_disk' tool.
If the variable was read from disk and it's changed, you should probably write it back with 'access_disk' tool.

If user want to visualize something, use the tools starting with 'ui_'.

Tool parameter values must be real, don't use placeholder(aka example.com) and don't make up numbers or strings! Use 'access_disk' tool to search for the value.
`
		//avoid script recursion ...............
		//- add attr .ScriptFile into Agent
		//   - if it not empty, add it to system prompt

	}
	return prompt
}

func (agent *Agent) addTools(antTools *[]*Anthropic_completion_tool, oaiTools *[]*OpenAI_completion_tool) error {

	toolList, scriptList, err := GetToolsList(agent.getFolder())
	if err != nil {
		return err
	}

	//tools
	for _, toolName := range toolList {
		path := filepath.Join(agent.getFolder(), toolName)
		if NeedCompileTool(path) {
			err := CompileTool(path)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}

		//add tool
		{
			openAIAPI_, anthropicAPI, err := ConvertCodeIntoTool(path)
			if err != nil {
				log.Fatal(err)
			}
			if antTools != nil {
				*antTools = append(*antTools, anthropicAPI)
			}
			if oaiTools != nil {
				*oaiTools = append(*oaiTools, openAIAPI_)
			}
		}
	}

	//scripts
	for _, scriptName := range scriptList {
		path := filepath.Join(agent.getFolder(), scriptName)

		//add tool
		{
			openAIAPI_, anthropicAPI, err := ConvertScriptIntoTool(path)
			if err != nil {
				log.Fatal(err)
			}
			if antTools != nil {
				*antTools = append(*antTools, anthropicAPI)
			}
			if oaiTools != nil {
				*oaiTools = append(*oaiTools, openAIAPI_)
			}
		}
	}

	return nil
}

func (agent *Agent) buildAnthropic() (Anthropic_completion_props, string, string) {
	ag_props := Agent_findChatAgentProperties(agent.getUseCase())
	if ag_props == nil {
		log.Fatalf("use_case %s not found.", agent.getUseCase())
	}
	login, _ := FindLoginChatModel(ag_props.Model) //err .....
	if login == nil {
		log.Fatalf("model %s not found.", ag_props.Model)
	}
	if login.Api_key_id == "" {
		log.Fatal(fmt.Errorf("no api_key for service '%s'", ag_props.Model))
	}
	api_key, err := g_agent_passwords.Find(login.Api_key_id)
	if err != nil {
		log.Fatal(err)
	}

	var props Anthropic_completion_props
	//system message
	props.System = agent.getSystemPrompt()

	//props
	props.Stream = false
	props.Model = ag_props.Model
	props.Temperature = ag_props.Temperature
	props.Max_tokens = ag_props.Max_tokens

	//tools
	err = agent.addTools(&props.Tools, nil)
	if err != nil {
		log.Fatal(err)
	}

	//messages
	for _, msg := range agent.Messages {
		props.Messages = append(props.Messages, Anthropic_completion_msg{Role: msg.Role, Content: msg.Content})
	}

	return props, login.Anthropic_completion_url, api_key
}
func (agent *Agent) buildOpenAI() (OpenAI_completion_props, string, string) {
	ag_props := Agent_findChatAgentProperties(agent.getUseCase())
	if ag_props == nil {
		log.Fatalf("use_case %s not found.", agent.getUseCase())
	}
	login, _ := FindLoginChatModel(ag_props.Model) //err .....
	if login == nil {
		log.Fatalf("model %s not found.", ag_props.Model)
	}
	if login.Api_key_id == "" {
		log.Fatal(fmt.Errorf("no api_key for service '%s'", ag_props.Model))
	}
	api_key, err := g_agent_passwords.Find(login.Api_key_id)
	if err != nil {
		log.Fatal(err)
	}

	var props OpenAI_completion_props
	//system message
	{
		msg := OpenAI_completion_msgPlain{Role: "system", Content: agent.getSystemPrompt()}
		props.Messages = append(props.Messages, msg)
	}

	//props
	props.Stream = false
	props.Model = ag_props.Model
	props.Temperature = ag_props.Temperature
	props.Max_tokens = ag_props.Max_tokens
	props.Top_p = ag_props.Top_p
	props.Frequency_penalty = ag_props.Frequency_penalty
	props.Presence_penalty = ag_props.Presence_penalty

	//tools
	err = agent.addTools(nil, &props.Tools)
	if err != nil {
		log.Fatal(err)
	}

	//messages
	for _, msg := range agent.Messages {

		if len(msg.Content) == 1 && msg.Content[0].Type == "text" {
			//text alone
			props.Messages = append(props.Messages, OpenAI_completion_msgPlain{Role: msg.Role, Content: msg.Content[0].Text})
		} else {

			mConst := OpenAI_completion_msg{Role: msg.Role}
			mCall := OpenAI_completion_msgCalls{Role: msg.Role}
			mResult := OpenAI_completion_msgResult{Role: "tool"}

			for _, it := range msg.Content {
				switch it.Type {
				case "text":
					mConst.AddText(it.Text)

				case "image":
					mConst.AddImage([]byte(it.Source.Data), it.Source.Media_type)

				case "tool_use":
					args, err := it.Input.MarshalJSON()
					if err != nil {
						log.Fatal(err)
					}
					mCall.Tool_calls = append(mCall.Tool_calls, OpenAI_completion_msg_Content_ToolCall{Id: it.Id, Type: "function", Function: OpenAI_completion_msg_Content_ToolCall_Function{Name: it.Name, Arguments: string(args)}})

				case "tool_result":
					mResult.Tool_call_id = it.Tool_use_id
					mResult.Content = it.Content
					mResult.Name = agent.FindToolName(it.Tool_use_id)
				}
			}

			if len(mConst.Content) > 0 {
				props.Messages = append(props.Messages, mConst)
			}
			if len(mCall.Tool_calls) > 0 {
				props.Messages = append(props.Messages, mCall)
			}
			if mResult.Tool_call_id != "" {
				props.Messages = append(props.Messages, mResult)
			}
		}
	}

	return props, login.OpenAI_completion_url, api_key
}

func (agent *Agent) AddUserPromptText(userPrompt string, createdBy string) {
	msg := Anthropic_completion_msg{}
	msg.AddText(userPrompt)

	agent.Messages = append(agent.Messages, AgentMsg{CreatedBy: createdBy, CreatedTimeSec: OsTime(), Role: "user", Content: msg.Content})
}
func (agent *Agent) AddUserPromptTextAndImages(text string, ImageFiles []string) {
	msg := Anthropic_completion_msg{}
	if text != "" {
		msg.AddText(text)
	}
	for _, file := range ImageFiles {
		err := msg.AddImageFile(file)
		if err != nil {
			Layout_WriteError(err)
		}
	}
	agent.Messages = append(agent.Messages, AgentMsg{CreatedBy: "", CreatedTimeSec: OsTime(), Role: "user", Content: msg.Content})
}

func (agent *Agent) AddCallResult(tool_name string, tool_use_id string, result string, createdBy string) {
	msg := Anthropic_completion_msg{}
	msg.AddToolResult(tool_name, tool_use_id, result)

	agent.Messages = append(agent.Messages, AgentMsg{CreatedBy: createdBy, CreatedTimeSec: OsTime(), Role: "user", Content: msg.Content})
}

func Agent_findChatAgentProperties(use_case string) *ChatAgent {
	var main_agent *ChatAgent
	agentPathes, _ := OpenDir_ChatAgents() //err ....
	for _, ag := range agentPathes {

		agent := OpenFile_AgentProperties(ag)

		nm := strings.ToLower(filepath.Base(ag))
		if nm == use_case {
			return agent
		}

		if nm == "main" {
			main_agent = agent
		}
	}

	return main_agent
}

func Agent_findSTTAgentProperties(use_case string) *STTAgent {
	var main_agent *STTAgent
	agentPathes, _ := OpenDir_STTAgents() //err ....
	for _, ag := range agentPathes {

		agent := OpenFile_STTAgent(ag)

		nm := strings.ToLower(filepath.Base(ag))
		if nm == use_case {
			return agent
		}

		if nm == "main" {
			main_agent = agent
		}
	}

	return main_agent
}

func (agent *Agent) FindSubAgent(sub_call_id string) *Agent {
	for _, it := range agent.SubAgents {
		if it.Call_id == sub_call_id {
			return it
		}
	}
	return nil
}
func (agent *Agent) FindSubCallUseContent(call_id string) (int, int) {
	for i, m := range agent.Messages {
		for j, t := range m.Content {
			if t.Type == "tool_use" && t.Id == call_id {
				return i, j
			}
		}
	}
	return -1, -1
}
func (agent *Agent) FindSubCallResultContent(call_id string) (*Anthropic_completion_msg_Content, int) {
	for i, m := range agent.Messages {
		for j, t := range m.Content {
			if t.Type == "tool_result" && t.Tool_use_id == call_id {
				return &agent.Messages[i].Content[j], i
			}
		}
	}
	return nil, -1
}

func (agent *Agent) FindToolName(call_id string) string {
	for _, m := range agent.Messages {
		for _, t := range m.Content {
			if t.Type == "tool_use" && t.Id == call_id {
				return t.Name
			}
		}
	}
	return ""
}

func (agent *Agent) IsShowToolParameters(call_id string) bool {
	for _, it := range agent.ShowToolParams {
		if it == call_id {
			return true
		}
	}
	return false
}
func (agent *Agent) SetShowToolParameters(call_id string, show bool) {
	exist := agent.IsShowToolParameters(call_id)
	if show {
		if !exist {
			agent.ShowToolParams = append(agent.ShowToolParams, call_id)
		}
	} else {
		if exist {
			for i, it := range agent.ShowToolParams {
				if it == call_id {
					agent.ShowToolParams = slices.Delete(agent.ShowToolParams, i, i+1)
					break
				}
			}
		}
	}
}

func (agent *Agent) RemoveUnfinishedMsg() {

	if agent.Stopped != "" {
		//if last agent is un-finished than delete message which call it and all later messages
		last_agent := len(agent.SubAgents) - 1
		if last_agent >= 0 && agent.SubAgents[last_agent].Stopped != "" {
			msg_i, _ := agent.FindSubCallUseContent(agent.SubAgents[last_agent].Call_id)
			if msg_i >= 0 {
				agent.Messages = slices.Delete(agent.Messages, msg_i, len(agent.Messages))
			}
		}
		agent.Stopped = "" //reset
	}
}

func (agent *Agent) HashUiCalls() bool {
	for _, msg := range agent.Messages {
		for _, it := range msg.Content {
			if it.Type == "tool_use" {
				if strings.HasPrefix(it.Name, "ui_") {
					return true
				}
			}
		}
	}
	return false
}

func (agent *Agent) IsModelAnthropic() bool {

	props := Agent_findChatAgentProperties(agent.getUseCase())
	if props == nil {
		log.Fatalf("use_case %s not found.", agent.getUseCase())
	}
	login, _ := FindLoginChatModel(props.Model) //err .....
	if login == nil {
		log.Fatalf("model %s not found.", props.Model)
	}

	return login.Anthropic_completion_url != ""
}

func (agent *Agent) GetFirstMessage() string {
	for _, msg := range agent.Messages {
		if msg.Role == "user" {
			for _, ct := range msg.Content {
				if ct.Type == "text" {
					return ct.Text
				}
			}
		}
	}
	return ""
}
func (agent *Agent) GetFinalMessage() string {
	if len(agent.Messages) == 0 {
		return ""
	}

	content := agent.Messages[len(agent.Messages)-1].Content
	if len(content) > 0 {
		//content[0].Type	//"image", "text", "tool_use", "tool_result" .....
		return content[0].Text
	}

	return ""
}

func (msg *AgentMsg) _getPrice(ag_props *ChatAgent, login *LLMLogin) (float64, float64, float64) {
	var input_price, cached_price, output_price float64

	for _, m := range login.ChatModels {
		if m.Name == ag_props.Model {
			input_price = m.Input_price / 1000000
			cached_price = m.Cached_price / 1000000
			output_price = m.Output_price / 1000000
		}
	}

	for _, m := range login.STTModels {
		if m.Name == ag_props.Model {
			output_price = m.Price / 1000000
		}
	}

	for _, m := range login.TTSModels {
		if m.Name == ag_props.Model {
			output_price = m.Price / 1000000
		}
	}

	return float64(msg.InputTokens) * input_price, float64(msg.InputCachedTokens) * cached_price, float64(msg.OutputTokens) * output_price
}

func (msg *AgentMsg) GetChatPrice(agent *Agent) (float64, float64, float64) {
	ag_props := Agent_findChatAgentProperties(agent.getUseCase())
	if ag_props == nil {
		log.Fatalf("use_case %s not found.", agent.getUseCase())
	}
	login, _ := FindLoginChatModel(ag_props.Model) //err .....
	if login == nil {
		log.Fatalf("model %s not found.", ag_props.Model)
	}
	return msg._getPrice(ag_props, login)
}

func (msg *AgentMsg) GetSpeed() float64 {
	toks := msg.OutputTokens
	if msg.Time == 0 {
		return 0
	}
	return float64(toks) / msg.Time
}

func (agent *Agent) _getTotalPrice(ag_props *ChatAgent, login *LLMLogin) (in, inCached, out float64) {
	//messages
	for _, m := range agent.Messages {
		i, ic, o := m._getPrice(ag_props, login)

		in += i
		inCached += ic
		out += o
	}

	//sub-agents
	for _, a := range agent.SubAgents {
		i, ic, o := a._getTotalPrice(ag_props, login)

		in += i
		inCached += ic
		out += o
	}

	return
}

func (agent *Agent) GetTotalChatPrice() (in, inCached, out float64) {
	ag_props := Agent_findChatAgentProperties(agent.getUseCase())
	if ag_props == nil {
		log.Fatalf("use_case %s not found.", agent.getUseCase())
	}
	login, _ := FindLoginChatModel(ag_props.Model) //err .....
	if login == nil {
		log.Fatalf("model %s not found.", ag_props.Model)
	}

	return agent._getTotalPrice(ag_props, login)
}

func (agent *Agent) GetTotalSpeed() float64 {
	toks := agent.GetTotalOutputTokens()
	dt := agent.GetTotalTime()
	if dt == 0 {
		return 0
	}
	return float64(toks) / dt

}

func (agent *Agent) GetTotalTime() float64 {
	dt := 0.0

	//messages
	for _, m := range agent.Messages {
		dt += m.Time
	}

	//sub-agents
	for _, a := range agent.SubAgents {
		dt += a.GetTotalTime()
	}

	return dt
}

func (agent *Agent) GetTotalOutputTokens() int {
	tokens := 0

	//messages
	for _, m := range agent.Messages {
		tokens += m.OutputTokens
	}

	//sub-agents
	for _, a := range agent.SubAgents {
		tokens += a.GetTotalOutputTokens()
	}

	return tokens
}

func (agent *Agent) Exe(job *Job) bool {
	if agent.IsModelAnthropic() {
		props, completion_url, api_key := agent.buildAnthropic()
		if agent.NoTools {
			props.Tools = nil
		}

		startTime := float64(time.Now().UnixMilli()) / 1000

		out, _, err := Anthropic_completion_Run(props, completion_url, api_key)
		if err != nil {
			log.Fatal(err)
		}

		dt := (float64(time.Now().UnixMilli()) / 1000) - startTime

		fmt.Printf("+LLM(%s) generated %dtoks which took %.1fsec = %.1f toks/sec\n", agent.getFolder(), out.Usage.Output_tokens, dt, float64(out.Usage.Output_tokens)/dt)
		//fmt.Printf("+LLM(%s) returns content: %s\n", agent.Folder, content)
		//fmt.Printf("+LLM(%s) returns tool_calls: %v\n", agent.Folder, tool_calls)

		msg := AgentMsg{Role: "assistant", Content: out.Content, CreatedTimeSec: OsTime(), CreatedBy: props.Model}

		msg.InputTokens = out.Usage.Input_tokens + out.Usage.Cache_creation_input_tokens
		msg.InputCachedTokens = out.Usage.Cache_read_input_tokens
		msg.OutputTokens = out.Usage.Output_tokens
		msg.Time = dt

		agent.Messages = append(agent.Messages, msg)

		return agent.callTools(out.Content, &props, nil, job)
	} else {
		props, completion_url, api_key := agent.buildOpenAI()
		if agent.NoTools {
			props.Tools = nil
		}

		startTime := float64(time.Now().UnixMilli()) / 1000

		var out OpenAIOut
		for {
			var statuCode int
			var err error
			out, statuCode, err = OpenAI_completion_Run(props, completion_url, api_key)
			if err == nil {
				break //done
			}
			if statuCode == 429 {
				time.Sleep(2000 * time.Millisecond)
				continue //again
			}
			if err != nil {
				log.Fatal(err)
			}
		}

		dt := (float64(time.Now().UnixMilli()) / 1000) - startTime

		var content string
		var tool_calls []OpenAI_completion_msg_Content_ToolCall
		if len(out.Choices) > 0 {
			content = out.Choices[0].Message.Content
			tool_calls = out.Choices[0].Message.Tool_calls
		}

		fmt.Printf("+LLM(%s) generated %dtoks which took %.1fsec = %.1f toks/sec\n", agent.getFolder(), out.Usage.Completion_tokens, dt, float64(out.Usage.Completion_tokens)/dt)
		fmt.Printf("+LLM(%s) returns content: %s\n", agent.getFolder(), content)
		fmt.Printf("+LLM(%s) returns tool_calls: %v\n", agent.getFolder(), tool_calls)

		content = strings.ReplaceAll(content, "<|separator|>", "")
		//if content == "<|separator|>" {
		//	return false
		//}

		var outContent []Anthropic_completion_msg_Content
		{
			contentWithCitations := content
			if len(out.Citations) > 0 {
				contentWithCitations += "\nCitations:\n"
				for _, ct := range out.Citations {
					contentWithCitations += ct + "\n"
				}
			}
			if contentWithCitations != "" {
				outContent = append(outContent, Anthropic_completion_msg_Content{Type: "text", Text: contentWithCitations})
			}
		}
		for _, tool := range tool_calls {
			var args json.RawMessage
			args.UnmarshalJSON([]byte(tool.Function.Arguments))
			outContent = append(outContent, Anthropic_completion_msg_Content{Type: "tool_use", Id: tool.Id, Name: tool.Function.Name, Input: args})
		}

		msg := AgentMsg{Role: "assistant", Content: outContent, CreatedTimeSec: OsTime(), CreatedBy: props.Model}

		msg.InputTokens = out.Usage.Prompt_tokens
		msg.InputCachedTokens = out.Usage.Input_cached_tokens
		msg.OutputTokens = out.Usage.Completion_tokens
		msg.Time = dt

		agent.Messages = append(agent.Messages, msg)
		return agent.callTools(outContent, nil, &props, job)
	}
}

func (agent *Agent) ExeLoop(max_iters int, max_tokens int, job *Job) {
	if max_iters <= 0 {
		max_iters = 20
	}
	if max_tokens <= 0 {
		max_tokens = 20000
	}

	orig_max_iters := max_iters
	orig_max_tokens := max_tokens

	if max_iters <= 0 {
		max_iters = 1000000000 //1B
	}
	if max_tokens <= 0 {
		max_tokens = 1000000000 //1B
	}

	agent.Stopped = "" //reset

	for max_iters > 0 {
		if !agent.Exe(job) {
			return //0 tool calls = natural end
		}

		if !job.IsRunning() {
			fmt.Printf("Warning: Agent was stopped by user\n")
			agent.Stopped = "Agent was stopped by user"
			return
		}

		if agent.GetTotalOutputTokens() >= max_tokens {
			fmt.Printf("Warning: Agent reached max tokens(%d)\n", orig_max_tokens)
			agent.Stopped = fmt.Sprintf("Agent reached max tokens(%d)\n", orig_max_tokens)
			return
		}

		max_iters--
	}

	fmt.Printf("Warning: Agent reached max iters(%d)\n", orig_max_iters)
	agent.Stopped = fmt.Sprintf("Agent reached max iters(%d)\n", orig_max_tokens)
}

func (agent *Agent) callTools(tool_calls []Anthropic_completion_msg_Content, ant *Anthropic_completion_props, oai *OpenAI_completion_props, job *Job) bool {
	num_calles := 0

	for _, it := range tool_calls {
		if !job.IsRunning() {
			break
		}

		if it.Type != "tool_use" {
			continue
		}

		args, err := it.Input.MarshalJSON()
		if err != nil {
			log.Fatal(err)
		}

		if ant != nil {
			for _, tool := range ant.Tools {
				if !job.IsRunning() {
					break
				}

				if tool.Name == it.Name {
					var answerJs string
					//call
					if agent.isScript(it.Name) {
						answerJs = agent.callScript(it.Id, it.Name, string(args), job)
					} else {
						answerJs = agent.callTool(it.Id, it.Name, string(args), job)
					}
					//save answer
					agent.AddCallResult(it.Name, it.Id, answerJs, "")
				}
			}
		}
		if oai != nil {
			for _, tool := range oai.Tools {
				if !job.IsRunning() {
					break
				}

				if tool.Function.Name == it.Name {
					var answerJs string
					//call
					if agent.isScript(it.Name) {
						answerJs = agent.callScript(it.Id, it.Name, string(args), job)
					} else {
						answerJs = agent.callTool(it.Id, it.Name, string(args), job)
					}
					//save answer
					agent.AddCallResult(it.Name, it.Id, answerJs, "")
				}
			}
		}

		if strings.HasPrefix(it.Name, "ui") {
			return false //stop
		}

		num_calles++
	}

	return num_calles > 0
}

func (agent *Agent) isScript(toolName string) bool {
	tool := filepath.Join(agent.getFolder(), toolName)
	return OsFileExists(tool + ".sky")
}

func (agent *Agent) callScript(call_id string, toolName string, arguments string, job *Job) string {
	toolPath := filepath.Join(agent.getFolder(), toolName+".sky")

	_, _, msgs, variables, err := GetScriptFileInfo(toolPath)
	if err != nil {
		return "" //...
	}

	//replace arguments in 'msgs'
	var args map[string]interface{}
	err = json.Unmarshal([]byte(arguments), &args)
	if err != nil {
		return "" //...
	}
	for i := range msgs {
		for name, vr := range variables {
			valInterface, found := args[name]
			if found {
				for _, exp := range vr.Exps {
					js, err := json.Marshal(valInterface)
					if err == nil {
						msgs[i] = strings.ReplaceAll(msgs[i], exp, string(js))
					}
				}
			}
		}
	}

	//new sub-agent
	sub_agent := NewAgent("", "", "")
	sub_agent.Call_id = call_id
	agent.SubAgents = append(agent.SubAgents, sub_agent)
	agent.Selected_sub_call_id = call_id //show

	//add description of the tool as the first message .................

	//call every message
	for _, msg := range msgs {
		if !job.IsRunning() {
			break
		}

		sub_agent.AddUserPromptText(msg, "guide")
		sub_agent.ExeLoop(0, 0, job)

		msg := AgentMsg{Role: "assistant", Content: []Anthropic_completion_msg_Content{{Type: "text", Text: "got it"}}, CreatedTimeSec: OsTime(), CreatedBy: "assistant"}
		sub_agent.Messages = append(sub_agent.Messages, msg)
	}

	return sub_agent.GetFinalMessage()
}

func (agent *Agent) callTool(call_id string, toolName string, arguments string, job *Job) string {
	toolPath := filepath.Join(agent.getFolder(), toolName)

	//call
	binPath := filepath.Join(filepath.Join("temp", toolPath), "bin")
	cmd := exec.Command("./"+binPath, strconv.Itoa(g_agent_server.port))
	cmd.Dir = ""
	OutStr := new(strings.Builder)
	ErrStr := new(strings.Builder)
	cmd.Stdout = OutStr
	cmd.Stderr = ErrStr
	err := cmd.Start()
	if err != nil {
		fmt.Println("Error:", err)
	}

	cl, err := g_agent_server.Accept()
	if err != nil {
		fmt.Println("Error:", err)
	}
	err = cl.WriteArray([]byte(arguments))
	if err != nil {
		fmt.Println("Error:", err)
	}

	var js []byte
	var tp uint64
	for tp != 1 {
		tp, err = cl.ReadInt()
		if err != nil {
			break
		}

		switch tp {
		case 1: //result
			js, _ = cl.ReadArray()

		case 2: //SDK_RunAgent
			max_iters, _ := cl.ReadInt()
			max_tokens, _ := cl.ReadInt()
			use_cases, _ := cl.ReadArray()
			systemPrompt, _ := cl.ReadArray()
			userPrompt, _ := cl.ReadArray()

			//init
			sub_agent := NewAgent(toolPath, string(use_cases), string(systemPrompt))
			sub_agent.AddUserPromptText(string(userPrompt), "")
			sub_agent.Call_id = call_id
			agent.SubAgents = append(agent.SubAgents, sub_agent)
			agent.Selected_sub_call_id = call_id //show

			//run
			sub_agent.ExeLoop(int(max_iters), int(max_tokens), job)

			//send result back
			cl.WriteArray([]byte(sub_agent.GetFinalMessage()))

		case 3: //SDK_SetToolCode
			toolName, _ := cl.ReadArray()
			toolCode, _ := cl.ReadArray()

			path := filepath.Join("tools", string(toolName)) //main folder
			os.MkdirAll(path, os.ModePerm)
			err := os.WriteFile(filepath.Join(path, "tool.go"), toolCode, 0644)
			if err != nil {
				fmt.Println(err)
			}

			err = CompileTool(path)
			if err == nil {
				//ok
				cl.WriteArray(nil)
			} else {
				//error
				cl.WriteArray([]byte(fmt.Sprintf("Tool '%s' was created, but compiler reported error: %v", path, err)))
			}

		case 4: //SDK_Sandbox_violation
			info, _ := cl.ReadArray()
			if agent != nil {
				agent.Sandbox_violations = append(agent.Sandbox_violations, string(info))
				fmt.Println("Sandbox violation:", string(info))
			}
			cl.WriteInt(1) //block it

		case 5: //SDK_GetPassword
			id, _ := cl.ReadArray()
			if agent != nil {
				password, err := g_agent_passwords.Find(string(id))
				if err == nil {
					cl.WriteInt(1) //ok
					cl.WriteArray([]byte(password))
				} else {
					cl.WriteInt(0) //error
					cl.WriteArray([]byte(err.Error()))
				}
				fmt.Println("Search for password:", string(id))
			}
			cl.WriteInt(1) //block it

		}
	}

	err = cmd.Wait()
	if err != nil {
		//tool crashed

		fmt.Println("OutStr", OutStr.String())
		fmt.Println("ErrStr", ErrStr.String())
		fmt.Println("Err", err.Error())

		wd, _ := os.Getwd()
		js = []byte(fmt.Sprintf("Tool '%s' crashed: %s", toolPath, strings.ReplaceAll(ErrStr.String(), wd, "")))
	}

	return string(js)
}
