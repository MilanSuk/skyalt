package main

import (
	"fmt"
	"strings"
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
	Function string
	Usage    LLMMsgUsage
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

func NewLLMOpenai(file string) (*LLMOpenai, error) {
	oai := &LLMOpenai{}

	oai.Provider = "OpenAI"
	oai.OpenAI_url = "https://api.openai.com/v1"
	oai.DevUrl = "https://platform.openai.com/"

	var err error
	oai, err = LoadFile(file, "LLMOpenai", "json", oai, true)
	if err == nil {
		oai.ReloadModels()
	}
	return oai, err
}

func (oai *LLMOpenai) Check(caller *ToolCaller) error {

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

func (oai *LLMOpenai) ReloadModels() error {

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
		Id:                             "o4-mini-latest",
		Input_modalities:               []string{"text", "image"},
		Prompt_text_token_price:        11000,
		Cached_prompt_text_token_price: 2750,
		Completion_text_token_price:    44000,
	})

	return nil
}

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
