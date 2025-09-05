package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
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

// OpenAI LLM settings.
type LLMGroq struct {
	Provider   string
	OpenAI_url string
	DevUrl     string
	API_key    string

	LanguageModels []*LLMGroqLanguageModel
	ImageModels    []*LLMGroqImageModel

	Stats []LLMMsgStats
}

func (grq *LLMGroq) Check() error {
	if grq.API_key == "" {
		return LogsErrorf("%s API key is empty", grq.Provider)
	}

	return nil
}

func (grq *LLMGroq) FindModel(name string) (*LLMGroqLanguageModel, *LLMGroqImageModel) {
	name = strings.ToLower(name)

	for _, model := range grq.LanguageModels {
		if strings.ToLower(model.Id) == name {
			return model, nil
		}
		for _, alias := range model.Aliases {
			if strings.ToLower(alias) == name {
				return model, nil
			}
		}
	}
	for _, model := range grq.ImageModels {
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

func (grq *LLMGroq) GetPricingString(model string) string {
	model = strings.ToLower(model)

	convert_to_dolars := float64(10000)

	lang, img := grq.FindModel(model)
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

func (grq *LLMGroq) Complete(st *LLMComplete, app_port int, tools []*ToolsOpenAI_completion_tool, msg *AppsRouterMsg) error {
	err := grq.Check()
	if err != nil {
		return err
	}

	mod, _ := grq.FindModel(st.Out_usage.Model)
	if mod == nil {
		return fmt.Errorf("model '%s' not found", st.Out_usage.Model)
	}

	stats, err := OpenAI_Complete(grq.Provider, grq.OpenAI_url, grq.API_key, st, app_port, tools, msg, mod.GetTextPrice)
	grq.Stats = append(grq.Stats, stats...)
	return err
}

func (grq *LLMGroq) Transcribe(st *LLMTranscribe) error {
	err := grq.Check()
	if err != nil {
		return err
	}

	model := "whisper-1"

	Completion_url := grq.OpenAI_url
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
	req.Header.Add("Authorization", "Bearer "+grq.API_key)

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

func (grq *LLMGroq) Speak(st *LLMSpeech) error {
	err := grq.Check()
	if err != nil {
		return err
	}

	model := "tts-1" //tts-1-hd
	voice := "alloy"

	Completion_url := grq.OpenAI_url
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
	req.Header.Add("Authorization", "Bearer "+grq.API_key)

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
