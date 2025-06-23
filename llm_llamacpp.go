package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type LLMLlamacppMsgStats struct {
	Function string
	Usage    LLMMsgUsage
}

// Llamacpp LLM settings.
type LLMLlamacpp struct {
	Address string
	Port    int

	Stats []LLMLlamacppMsgStats
}

func (llama *LLMLlamacpp) Check() error {
	if llama.Address == "" {
		return fmt.Errorf("llama.cpp address is empty")
	}

	return nil
}

func (llama *LLMLlamacpp) Complete(st *LLMComplete, router *ToolsRouter, msg *ToolsRouterMsg) error {
	err := llama.Check()
	if err != nil {
		return err
	}

	//Tools
	var tools []*ToolsOpenAI_completion_tool
	var app *ToolsApp
	if st.AppName != "" {
		app = router.FindApp(st.AppName)
		if app != nil {
			tools = app.GetAllSchemas()
		} else {
			return fmt.Errorf("app '%s' not found", st.AppName)
		}
	}

	//Messages
	var msgs ChatMsgs
	if len(st.PreviousMessages) > 0 {
		err := json.Unmarshal(st.PreviousMessages, &msgs)
		if err != nil {
			return fmt.Errorf("PreviousMessages failed: %v", err)
		}
	}

	if st.UserMessage != "" || len(st.UserFiles) > 0 {
		m1, err := msgs.AddUserMessage(st.UserMessage, st.UserFiles)
		if err != nil {
			return fmt.Errorf("AddUserMessage() failed: %v", err)
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
			Seed:           seed,

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

		Provider := "llamacpp"

		fnStreaming := func(chatMsg *ChatMsg) bool {
			chatMsg.Seed = seed
			chatMsg.Stream = true
			chatMsg.ShowParameters = true
			chatMsg.ShowReasoning = true

			if st.delta != nil {
				st.delta(chatMsg)
			}

			return msg.Progress(0, "completing")
		}

		//print
		{
			js, err := json.MarshalIndent(props, "", "  ")
			if err == nil {
				fmt.Printf("---\n" + string(js) + "\n---\n")
			}
		}

		jsProps, err := json.Marshal(props)
		if err != nil {
			return fmt.Errorf("props failed: %v", err)
		}
		out, status, dt, time_to_first_token, err := OpenAI_completion_Run(jsProps, fmt.Sprintf("%s:%d/v1", llama.Address, llama.Port), "", fnStreaming)
		st.Out_StatusCode = status
		if err != nil {
			return fmt.Errorf("OpenAI_completion_Run() failed: %v", err)
		}

		if !msg.Progress(0, "completing") {
			return nil
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

				//mod, _ := llama.FindModel(st.Model)
				//if mod != nil {
				//	usage.Prompt_price, usage.Reasoning_price, usage.Input_cached_price, usage.Completion_price = mod.GetTextPrice(usage.Prompt_tokens, usage.Reasoning_tokens, usage.Input_cached_tokens, usage.Completion_tokens)
				//}

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

				//start it
				err := app.CheckRun()
				if router.log.Error(err) != nil {
					return err
				}

				//call it
				resJs, uiJs, cmdsJs, err := _ToolsCaller_CallTool(app.Process.port, msg.msg_id, 0, call.Function.Name, []byte(call.Function.Arguments), router.log.Error)
				if router.log.Error(err) != nil {
					return err
				}
				//resJs, tool_ui, err := CallToolApp(st.AppName, call.Function.Name, []byte(call.Function.Arguments), caller)

				//add cmds
				msg.out_flushed_cmds = append(msg.out_flushed_cmds, cmdsJs)

				resMap := make(map[string]interface{})
				if err == nil {
					err = json.Unmarshal(resJs, &resMap)
					if err != nil {
						return fmt.Errorf("resJs failed: %v", err)
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
				} else {
					result = "Error: " + err.Error()
				}

				var tool_ui UI
				if err == nil {
					err = json.Unmarshal(uiJs, &tool_ui)
					router.log.Error(err)
				}
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
			llama.Stats = append(llama.Stats, LLMLlamacppMsgStats{
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

	//print
	{
		js, err := json.MarshalIndent(msgs, "", "  ")
		if err == nil {
			fmt.Printf("+++\n" + string(js) + "\n+++\n")
		}
	}

	st.Out_messages, err = json.Marshal(msgs)
	if err != nil {
		return fmt.Errorf("out_messages failed: %v", err)
	}

	return nil
}
