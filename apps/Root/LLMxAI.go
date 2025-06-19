package main

import (
	"bytes"
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

	xai.ReloadModels()

	return nil
}

func (xai *LLMxAI) FindModel(name string) (*LLMxAILanguageModel, *LLMxAIImageModel) {
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

func (xai *LLMxAI) ReloadModels() error {

	//reset
	xai.LanguageModels = nil
	xai.ImageModels = nil

	xai.LanguageModels = append(xai.LanguageModels, &LLMxAILanguageModel{
		Id:                             "grok-3",
		Input_modalities:               []string{"text"},
		Prompt_text_token_price:        30000,
		Cached_prompt_text_token_price: 7500,
		Completion_text_token_price:    150000,
	})
	xai.LanguageModels = append(xai.LanguageModels, &LLMxAILanguageModel{
		Id:                             "grok-3-fast",
		Input_modalities:               []string{"text"},
		Prompt_text_token_price:        50000,
		Cached_prompt_text_token_price: 12500,
		Completion_text_token_price:    250000,
	})

	xai.LanguageModels = append(xai.LanguageModels, &LLMxAILanguageModel{
		Id:                             "grok-3-mini",
		Input_modalities:               []string{"text"},
		Prompt_text_token_price:        3000,
		Cached_prompt_text_token_price: 75,
		Completion_text_token_price:    5000,
	})

	xai.LanguageModels = append(xai.LanguageModels, &LLMxAILanguageModel{
		Id:                             "grok-3-mini-fast",
		Input_modalities:               []string{"text"},
		Prompt_text_token_price:        6000,
		Cached_prompt_text_token_price: 1500,
		Completion_text_token_price:    40000,
	})

	xai.LanguageModels = append(xai.LanguageModels, &LLMxAILanguageModel{
		Id:                             "grok-2-vision",
		Input_modalities:               []string{"text", "image"},
		Prompt_text_token_price:        20000,
		Cached_prompt_text_token_price: 20000,
		Completion_text_token_price:    100000,
	})

	//Image models

	xai.ImageModels = append(xai.ImageModels, &LLMxAIImageModel{
		Id:          "grok-2-image",
		Image_price: 700,
	})

	return nil
}

func (xai *LLMxAI) GetPricingString(model string) string {
	model = strings.ToLower(model)

	convert_to_dolars := float64(10000)

	lang, img := xai.FindModel(model)
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
