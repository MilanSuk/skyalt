package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type LLMOpenaiLanguageModel struct {
	Id               string
	Created          int64
	Version          string
	Input_modalities []string //"text", "image"

	Prompt_text_token_price        int //USD cents per million token
	Prompt_image_token_price       int
	Cached_prompt_text_token_price int
	Completion_text_token_price    int

	Aliases []string
}

type LLMOpenaiImageModel struct {
	Id                string
	Created           int64
	Version           string
	Max_prompt_length int

	Image_price int //USD cents per image

	Aliases []string
}

type LLMOpenaiMsgStats struct {
	Function       string
	CreatedTimeSec float64
	Model          string

	Time             float64
	TimeToFirstToken float64

	Usage LLMMsgUsage
}

// OpenAI LLM settings.
type LLMOpenai struct {
	Provider   string
	OpenAI_url string
	DevUrl     string
	API_key    string

	LanguageModels []*LLMOpenaiLanguageModel
	ImageModels    []*LLMOpenaiImageModel

	Stats []LLMOpenaiMsgStats
}

func (oai *LLMOpenai) Check() error {
	if oai.API_key == "" {
		return fmt.Errorf("%s API key is empty", oai.Provider)
	}

	return nil
}

func (oai *LLMOpenai) FindModel(name string) (*LLMOpenaiLanguageModel, *LLMOpenaiImageModel) {
	name = strings.ToLower(name)

	for _, model := range oai.LanguageModels {
		if strings.ToLower(model.Id) == name {
			return model, nil
		}
		for _, alias := range model.Aliases {
			if strings.ToLower(alias) == name {
				return model, nil
			}
		}
	}
	for _, model := range oai.ImageModels {
		if strings.ToLower(model.Id) == name {
			return nil, model
		}
		for _, alias := range model.Aliases {
			if strings.ToLower(alias) == name {
				return nil, model
			}
		}
	}

	return nil, nil
}

/*func (oai *LLMOpenai) ReloadModels() error {

	//reset
	oai.LanguageModels = nil
	oai.ImageModels = nil

	oai.LanguageModels = append(oai.LanguageModels, &LLMOpenaiLanguageModel{
		Id:                             "gpt-4.1-nano",
		Input_modalities:               []string{"text", "image"},
		Prompt_text_token_price:        1000,
		Cached_prompt_text_token_price: 250,
		Completion_text_token_price:    4000,
	})
	oai.LanguageModels = append(oai.LanguageModels, &LLMOpenaiLanguageModel{
		Id:                             "gpt-4.1-mini",
		Input_modalities:               []string{"text", "image"},
		Prompt_text_token_price:        4000,
		Cached_prompt_text_token_price: 1000,
		Completion_text_token_price:    16000,
	})

	oai.LanguageModels = append(oai.LanguageModels, &LLMOpenaiLanguageModel{
		Id:                             "gpt-4o-mini",
		Input_modalities:               []string{"text", "image"},
		Prompt_text_token_price:        1500,
		Cached_prompt_text_token_price: 750,
		Completion_text_token_price:    6000,
	})

	oai.LanguageModels = append(oai.LanguageModels, &LLMOpenaiLanguageModel{
		Id:                             "o4-mini",
		Input_modalities:               []string{"text", "image"},
		Prompt_text_token_price:        11000,
		Cached_prompt_text_token_price: 2750,
		Completion_text_token_price:    44000,
	})

	return nil
}*/

func (oai *LLMOpenai) GetPricingString(model string) string {
	model = strings.ToLower(model)

	convert_to_dolars := float64(10000)

	lang, img := oai.FindModel(model)
	if lang != nil {
		//in, cached, out, image
		return fmt.Sprintf("$%.2f/$%.2f/$%.2f/$%.2f", float64(lang.Prompt_text_token_price)/convert_to_dolars, float64(lang.Prompt_image_token_price)/convert_to_dolars, float64(lang.Cached_prompt_text_token_price)/convert_to_dolars, float64(lang.Completion_text_token_price)/convert_to_dolars)
	}

	if img != nil {
		return fmt.Sprintf("$%.2f", float64(img.Image_price)/convert_to_dolars)
	}

	return fmt.Sprintf("model %s not found", model)
}

func (model *LLMOpenaiLanguageModel) GetTextPrice(in, reason, cached, out int) (float64, float64, float64, float64) {

	convert_to_dolars := float64(10000)

	Input_price := float64(model.Prompt_text_token_price) / convert_to_dolars / 1000000
	Reason_price := float64(model.Prompt_text_token_price) / convert_to_dolars / 1000000
	Cached_price := float64(model.Cached_prompt_text_token_price) / convert_to_dolars / 1000000
	Output_price := float64(model.Completion_text_token_price) / convert_to_dolars / 1000000

	return float64(in) * Input_price, float64(reason) * Reason_price, float64(cached) * Cached_price, float64(out) * Output_price
}

func (oai *LLMOpenai) Complete(st *LLMComplete, router *ToolsRouter, msg *ToolsRouterMsg) error {
	err := oai.Check()
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

			Model: st.Out_model,

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

		fnStreaming := func(chatMsg *ChatMsg) bool {

			chatMsg.Provider = oai.Provider
			chatMsg.Model = st.Out_model
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
		out, status, dt, time_to_first_token, err := OpenAI_completion_Run(jsProps, oai.OpenAI_url, oai.API_key, fnStreaming)
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
				mod, _ := oai.FindModel(st.Out_model)
				if mod != nil {
					usage.Prompt_price, usage.Reasoning_price, usage.Input_cached_price, usage.Completion_price = mod.GetTextPrice(usage.Prompt_tokens, usage.Reasoning_tokens, usage.Input_cached_tokens, usage.Completion_tokens)
				}

				//add
				{
					st.Out_usage.Add(&usage)
				}
			}

			calls := out.Choices[0].Message.Tool_calls
			m2 := msgs.AddAssistentCalls(out.Choices[0].Message.Reasoning_content, out.Choices[0].Message.Content, calls, usage, dt, time_to_first_token, oai.Provider, st.Out_model)
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
			oai.Stats = append(oai.Stats, LLMOpenaiMsgStats{
				Function:       "completion",
				CreatedTimeSec: float64(time.Now().UnixMicro()) / 1000000,
				Model:          st.Out_model,

				Time:             dt,
				TimeToFirstToken: time_to_first_token,

				Usage: usage,
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

func (oai *LLMOpenai) Transcribe(st *LLMTranscribe) error {
	err := oai.Check()
	if err != nil {
		return err
	}

	model := "whisper-1"

	Completion_url := oai.OpenAI_url
	if !strings.HasSuffix(Completion_url, "/") {
		Completion_url += "/"
	}
	Completion_url += "audio/transcriptions"

	//set parameters
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	{
		part, err := writer.CreateFormFile("file", st.BlobFileName)
		if err != nil {
			return fmt.Errorf("CreateFormFile() failed: %w", err)
		}
		part.Write(st.AudioBlob)

		writer.WriteField("temperature", strconv.FormatFloat(st.Temperature, 'f', -1, 64))
		writer.WriteField("response_format", st.Response_format)
		//props.Write(writer)

		writer.WriteField("model", model)

		if st.Response_format == "verbose_json" {
			writer.WriteField("timestamp_granularities[]", "word")
			writer.WriteField("timestamp_granularities[]", "segment")
		}
	}
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, Completion_url, body)
	if err != nil {
		return fmt.Errorf("NewRequest() failed: %w", err)
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Add("Authorization", "Bearer "+oai.API_key)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Do() failed: %w", err)
	}

	resBody, err := io.ReadAll(res.Body) //job.close ...
	st.Out_StatusCode = res.StatusCode

	if err != nil {
		return fmt.Errorf("ReadAll() failed: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("statusCode %d != 200, response: %s", res.StatusCode, string(resBody))
	}

	st.Out_Output = resBody

	return nil
}

func (oai *LLMOpenai) Speak(st *LLMSpeech) error {
	err := oai.Check()
	if err != nil {
		return err
	}

	model := "tts-1" //tts-1-hd
	voice := "alloy"

	Completion_url := oai.OpenAI_url
	if !strings.HasSuffix(Completion_url, "/") {
		Completion_url += "/"
	}
	Completion_url += "audio/speech"

	type TTS struct {
		Model string `json:"model"`
		Input string `json:"input"`
		Voice string `json:"voice"`
	}

	tts := TTS{Model: model, Voice: voice, Input: st.Text}
	js, err := json.Marshal(tts)
	if err != nil {
		return fmt.Errorf("Marshal() failed: %w", err)
	}
	body := bytes.NewReader(js)

	req, err := http.NewRequest(http.MethodPost, Completion_url, body)
	if err != nil {
		return fmt.Errorf("NewRequest() failed: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+oai.API_key)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Do() failed: %w", err)
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	st.Out_StatusCode = res.StatusCode
	if err != nil {
		return fmt.Errorf("ReadAll() failed: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("statusCode %d != 200, response: %s", res.StatusCode, string(js))
	}

	st.Out_Output = resBody

	return nil
}
