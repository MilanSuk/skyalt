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
	Function       string
	CreatedTimeSec float64
	Model          string

	Time             float64
	TimeToFirstToken float64

	Usage ChatMsgUsage
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
	mst := &LLMOpenai{}

	mst.Provider = "OpenAI"
	mst.OpenAI_url = "https://api.openai.com/v1"
	mst.DevUrl = "https://platform.openai.com/"

	return LoadFile(file, "LLMOpenai", "json", mst, true)
}

func (mst *LLMOpenai) Check(caller *ToolCaller) error {

	if mst.API_key == "" {
		return fmt.Errorf("%s API key is empty", mst.Provider)
	}

	//reload models
	if len(mst.LanguageModels) == 0 {
		mst.ReloadModels(caller)
	}

	return nil
}

func (mst *LLMOpenai) FindProviderModel(name string) (*LLMOpenaiLanguageModel, *LLMOpenaiImageModel) {
	name = strings.ToLower(name)

	for _, model := range mst.LanguageModels {
		if strings.ToLower(model.Id) == name {
			return model, nil
		}
		for _, alias := range model.Aliases {
			if strings.ToLower(alias) == name {
				return model, nil
			}
		}
	}
	for _, model := range mst.ImageModels {
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

func (mst *LLMOpenai) ReloadModels(caller *ToolCaller) error {

	//reset
	mst.LanguageModels = nil
	mst.ImageModels = nil

	mst.LanguageModels = append(mst.LanguageModels, &LLMOpenaiLanguageModel{
		Id:                             "gpt-4.1-nano-latest",
		Input_modalities:               []string{"text", "image"},
		Prompt_text_token_price:        1000,
		Cached_prompt_text_token_price: 250,
		Completion_text_token_price:    4000,
	})
	mst.LanguageModels = append(mst.LanguageModels, &LLMOpenaiLanguageModel{
		Id:                             "gpt-4.1-mini-latest",
		Input_modalities:               []string{"text", "image"},
		Prompt_text_token_price:        4000,
		Cached_prompt_text_token_price: 1000,
		Completion_text_token_price:    16000,
	})

	mst.LanguageModels = append(mst.LanguageModels, &LLMOpenaiLanguageModel{
		Id:                             "gpt-4o-mini-latest",
		Input_modalities:               []string{"text", "image"},
		Prompt_text_token_price:        1500,
		Cached_prompt_text_token_price: 750,
		Completion_text_token_price:    6000,
	})

	mst.LanguageModels = append(mst.LanguageModels, &LLMOpenaiLanguageModel{
		Id:                             "o4-mini-latest",
		Input_modalities:               []string{"text", "image"},
		Prompt_text_token_price:        11000,
		Cached_prompt_text_token_price: 2750,
		Completion_text_token_price:    44000,
	})

	return nil
}

func (mst *LLMOpenai) GetPricingString(model string) string {
	model = strings.ToLower(model)

	convert_to_dolars := float64(10000)

	lang, img := mst.FindProviderModel(model)
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
