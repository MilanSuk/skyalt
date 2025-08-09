package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type LLMGroqLanguageModel struct {
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

type LLMGroqImageModel struct {
	Id                string
	Created           int64
	Version           string
	Max_prompt_length int

	Image_price int //USD cents per image

	Aliases []string
}

type LLMGroqMsgStats struct {
	Function string
	Usage    LLMMsgUsage
}

// OpenAI LLM settings.
type LLMGroq struct {
	Provider   string
	OpenAI_url string
	DevUrl     string
	API_key    string

	LanguageModels []*LLMGroqLanguageModel
	ImageModels    []*LLMGroqImageModel

	Stats []LLMGroqMsgStats
}

func (oai *LLMGroq) Check() error {
	if oai.API_key == "" {
		return LogsErrorf("%s API key is empty", oai.Provider)
	}

	return nil
}

func (oai *LLMGroq) FindModel(name string) (*LLMGroqLanguageModel, *LLMGroqImageModel) {
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

func (oai *LLMGroq) GetPricingString(model string) string {
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

func (model *LLMGroqLanguageModel) GetTextPrice(in, reason, cached, out int) (float64, float64, float64, float64) {

	convert_to_dolars := float64(10000)

	Input_price := float64(model.Prompt_text_token_price) / convert_to_dolars / 1000000
	Reason_price := float64(model.Prompt_text_token_price) / convert_to_dolars / 1000000
	Cached_price := float64(model.Cached_prompt_text_token_price) / convert_to_dolars / 1000000
	Output_price := float64(model.Completion_text_token_price) / convert_to_dolars / 1000000

	return float64(in) * Input_price, float64(reason) * Reason_price, float64(cached) * Cached_price, float64(out) * Output_price
}

func (oai *LLMGroq) Complete(st *LLMComplete, app_port int, tools []*ToolsOpenAI_completion_tool, msg *AppsRouterMsg) error {
	err := oai.Check()
	if err != nil {
		return err
	}

	//Messages
	var msgs ChatMsgs
	if len(st.PreviousMessages) > 0 {
		err := LogsJsonUnmarshal(st.PreviousMessages, &msgs)
		if err != nil {
			return err
		}
	}

	if st.UserMessage != "" || len(st.UserFiles) > 0 {
		m1, err := msgs.AddUserMessage(st.UserMessage, st.UserFiles)
		if err != nil {
			return err
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

		fnStreaming := func(chatMsg *ChatMsg) bool {
			chatMsg.Seed = seed
			chatMsg.Stream = true
			chatMsg.ShowParameters = true
			chatMsg.ShowReasoning = true

			if st.delta != nil {
				st.delta(chatMsg)
			}

			return msg.GetContinue()
		}

		//print
		{
			js, err := LogsJsonMarshalIndent(props)
			if err == nil {
				fmt.Printf("---\n" + string(js) + "\n---\n")
			}
		}

		jsProps, err := LogsJsonMarshal(props)
		if err != nil {
			return err
		}
		out, status, dt, time_to_first_token, err := OpenAI_completion_Run(jsProps, oai.OpenAI_url, oai.API_key, fnStreaming, msg)
		st.Out_StatusCode = status
		if err != nil {
			return err
		}

		if !msg.GetContinue() {
			return nil
		}

		if len(out.Choices) > 0 {

			var usage LLMMsgUsage
			{
				usage.Prompt_tokens = out.Usage.Prompt_tokens
				usage.Input_cached_tokens = out.Usage.Input_cached_tokens
				usage.Completion_tokens = out.Usage.Completion_tokens
				usage.Reasoning_tokens = out.Usage.Completion_tokens_details.Reasoning_tokens

				usage.Provider = oai.Provider
				usage.Model = st.Out_usage.Model
				usage.CreatedTimeSec = float64(time.Now().UnixMicro()) / 1000000
				usage.TimeToFirstToken = time_to_first_token
				usage.DTime = dt

				mod, _ := oai.FindModel(st.Out_usage.Model)
				if mod != nil {
					usage.Prompt_price, usage.Reasoning_price, usage.Input_cached_price, usage.Completion_price = mod.GetTextPrice(usage.Prompt_tokens, usage.Reasoning_tokens, usage.Input_cached_tokens, usage.Completion_tokens)
				}

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

				//call it
				resJs, uiGob, cmdsGob, err := _ToolsCaller_CallBuild(app_port, msg.msg_id, 0, call.Function.Name, []byte(call.Function.Arguments))
				if err != nil {
					return err
				}
				//resJs, tool_ui, err := CallToolApp(st.AppName, call.Function.Name, []byte(call.Function.Arguments), caller)

				//add cmds
				msg.out_flushed_cmdsGob = append(msg.out_flushed_cmdsGob, cmdsGob)

				resMap := make(map[string]interface{})
				err = LogsJsonUnmarshal(resJs, &resMap)
				if err != nil {
					return err
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
				var tool_ui UI
				LogsGobUnmarshal(uiGob, &tool_ui)

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
			oai.Stats = append(oai.Stats, LLMGroqMsgStats{
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
		js, err := LogsJsonMarshalIndent(msgs)
		if err == nil {
			fmt.Printf("+++\n" + string(js) + "\n+++\n")
		}
	}

	st.Out_messages, err = LogsJsonMarshal(msgs)
	if err != nil {
		return err
	}

	return nil
}

func (oai *LLMGroq) Transcribe(st *LLMTranscribe) error {
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
		if LogsError(err) != nil {
			return err
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
	if LogsError(err) != nil {
		return err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Add("Authorization", "Bearer "+oai.API_key)

	client := &http.Client{}
	res, err := client.Do(req)
	if LogsError(err) != nil {
		return err
	}

	resBody, err := io.ReadAll(res.Body) //job.close ...
	st.Out_StatusCode = res.StatusCode

	if LogsError(err) != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return LogsErrorf("statusCode %d != 200, response: %s", res.StatusCode, string(resBody))
	}

	st.Out_Output = resBody

	return nil
}

func (oai *LLMGroq) Speak(st *LLMSpeech) error {
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
	js, err := LogsJsonMarshal(tts)
	if err != nil {
		return err
	}
	body := bytes.NewReader(js)

	req, err := http.NewRequest(http.MethodPost, Completion_url, body)
	if LogsError(err) != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+oai.API_key)

	client := &http.Client{}
	res, err := client.Do(req)
	if LogsError(err) != nil {
		return err
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	st.Out_StatusCode = res.StatusCode
	if LogsError(err) != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return LogsErrorf("statusCode %d != 200, response: %s", res.StatusCode, string(js))
	}

	st.Out_Output = resBody

	return nil
}
