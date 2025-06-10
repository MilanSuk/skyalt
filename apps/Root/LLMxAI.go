package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type LLMxAILanguageModel struct {
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

type LLMxAIImageModel struct {
	Id                string
	Created           int64
	Version           string
	Max_prompt_length int

	Image_price int //USD cents per image

	Aliases []string
}

type LLMxAIMsgStats struct {
	Function       string
	CreatedTimeSec float64
	Model          string

	Time             float64
	TimeToFirstToken float64

	Usage ChatMsgUsage
}

// xAI LLM settings.
type LLMxAI struct {
	Provider   string
	OpenAI_url string
	DevUrl     string
	API_key    string

	FastMode  bool
	SmartMode bool

	LanguageModels []*LLMxAILanguageModel
	ImageModels    []*LLMxAIImageModel

	Stats []LLMxAIMsgStats
}

func NewLLMxAI(file string) (*LLMxAI, error) {
	xai := &LLMxAI{}

	xai.Provider = "xAI"
	xai.OpenAI_url = "https://api.x.ai/v1"
	xai.DevUrl = "https://console.x.ai"

	return LoadFile(file, "LLMxAI", "json", xai, true)
}

func (xai *LLMxAI) Check(caller *ToolCaller) error {

	if xai.API_key == "" {
		return fmt.Errorf("%s API key is empty", xai.Provider)
	}

	//reload models
	if len(xai.LanguageModels) == 0 {
		xai.ReloadModels(caller)
	}

	return nil
}

func (xai *LLMxAI) FindProviderModel(name string) (*LLMxAILanguageModel, *LLMxAIImageModel) {
	name = strings.ToLower(name)

	for _, model := range xai.LanguageModels {
		if strings.ToLower(model.Id) == name {
			return model, nil
		}
		for _, alias := range model.Aliases {
			if strings.ToLower(alias) == name {
				return model, nil
			}
		}
	}
	for _, model := range xai.ImageModels {
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

func (xai *LLMxAI) ReloadModels(caller *ToolCaller) error {

	//reset
	xai.LanguageModels = nil
	xai.ImageModels = nil

	//Language models
	{
		js, err := xai.downloadList("language-models")
		if err != nil {
			return err
		}

		type ST struct {
			Models []*LLMxAILanguageModel
		}
		var stt ST
		err = json.Unmarshal(js, &stt)
		if err != nil {
			return err
		}
		xai.LanguageModels = stt.Models
	}

	//Image models
	{
		js, err := xai.downloadList("image-generation-models")
		if err != nil {
			return err
		}

		type ST struct {
			Models []*LLMxAIImageModel
		}
		var stt ST
		err = json.Unmarshal(js, &stt)
		if err != nil {
			return err
		}
		xai.ImageModels = stt.Models
	}

	return nil
}

func (xai *LLMxAI) GetPricingString(model string) string {
	model = strings.ToLower(model)

	convert_to_dolars := float64(10000)

	lang, img := xai.FindProviderModel(model)
	if lang != nil {
		//in, cached, out, image
		return fmt.Sprintf("$%.2f/$%.2f/$%.2f/$%.2f", float64(lang.Prompt_text_token_price)/convert_to_dolars, float64(lang.Prompt_image_token_price)/convert_to_dolars, float64(lang.Cached_prompt_text_token_price)/convert_to_dolars, float64(lang.Completion_text_token_price)/convert_to_dolars)
	}

	if img != nil {
		return fmt.Sprintf("$%.2f", float64(img.Image_price)/convert_to_dolars)
	}

	return fmt.Sprintf("model %s not found", model)
}

func (model *LLMxAILanguageModel) GetTextPrice(in, reason, cached, out int) (float64, float64, float64, float64) {

	convert_to_dolars := float64(10000)

	Input_price := float64(model.Prompt_text_token_price) / convert_to_dolars / 1000000
	Reason_price := float64(model.Prompt_text_token_price) / convert_to_dolars / 1000000
	Cached_price := float64(model.Cached_prompt_text_token_price) / convert_to_dolars / 1000000
	Output_price := float64(model.Completion_text_token_price) / convert_to_dolars / 1000000

	return float64(in) * Input_price, float64(reason) * Reason_price, float64(cached) * Cached_price, float64(out) * Output_price
}

func (xai *LLMxAI) downloadList(url_part string) ([]byte, error) {
	if xai.API_key == "" {
		return nil, fmt.Errorf("%s API key is empty", xai.Provider)
	}

	Completion_url := xai.OpenAI_url
	if !strings.HasSuffix(Completion_url, "/") {
		Completion_url += "/"
	}
	Completion_url += url_part

	body := bytes.NewReader(nil)
	req, err := http.NewRequest(http.MethodGet, Completion_url, body) //http.MethodPost
	if err != nil {
		return nil, fmt.Errorf("NewRequest() failed: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+xai.API_key)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Do() failed: %w", err)
	}
	defer res.Body.Close()

	js, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return js, nil
}
