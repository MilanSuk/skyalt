package main

import (
	"fmt"
	"strings"
)

type LLMMistralLanguageModel struct {
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

type LLMMistralImageModel struct {
	Id                string
	Created           int64
	Version           string
	Max_prompt_length int

	Image_price int //USD cents per image

	Aliases []string
}

type LLMMistralMsgStats struct {
	Function       string
	CreatedTimeSec float64
	Model          string

	Time             float64
	TimeToFirstToken float64

	Usage ChatMsgUsage
}

// Mistral LLM settings.
type LLMMistral struct {
	Provider   string
	OpenAI_url string
	DevUrl     string
	API_key    string

	LanguageModels []*LLMMistralLanguageModel
	ImageModels    []*LLMMistralImageModel

	Stats []LLMMistralMsgStats
}

func NewLLMMistral(file string) (*LLMMistral, error) {
	mst := &LLMMistral{}

	mst.Provider = "Mistral"
	mst.OpenAI_url = "https://api.mistral.ai/v1"
	mst.DevUrl = "https://console.mistral.ai"

	return LoadFile(file, "LLMMistral", "json", mst, true)
}

func (mst *LLMMistral) Check(caller *ToolCaller) error {

	if mst.API_key == "" {
		return fmt.Errorf("%s API key is empty", mst.Provider)
	}

	//reload models
	if len(mst.LanguageModels) == 0 {
		mst.ReloadModels(caller)
	}

	return nil
}

func (mst *LLMMistral) FindProviderModel(name string) (*LLMMistralLanguageModel, *LLMMistralImageModel) {
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

func (mst *LLMMistral) ReloadModels(caller *ToolCaller) error {

	//reset
	mst.LanguageModels = nil
	mst.ImageModels = nil

	mst.LanguageModels = append(mst.LanguageModels, &LLMMistralLanguageModel{
		Id:                             "devstral-small-latest",
		Input_modalities:               []string{"text"},
		Prompt_text_token_price:        0,
		Cached_prompt_text_token_price: 0,
		Completion_text_token_price:    0,
	})

	mst.LanguageModels = append(mst.LanguageModels, &LLMMistralLanguageModel{
		Id:                             "mistral-small-latest",
		Input_modalities:               []string{"text"},
		Prompt_text_token_price:        0,
		Cached_prompt_text_token_price: 0,
		Completion_text_token_price:    0,
	})

	mst.LanguageModels = append(mst.LanguageModels, &LLMMistralLanguageModel{
		Id:                             "magistral-small-latest",
		Input_modalities:               []string{"text"},
		Prompt_text_token_price:        0,
		Cached_prompt_text_token_price: 0,
		Completion_text_token_price:    0,
	})

	mst.LanguageModels = append(mst.LanguageModels, &LLMMistralLanguageModel{
		Id:                             "pixtral-12b-latest",
		Input_modalities:               []string{"text", "image"},
		Prompt_text_token_price:        0,
		Cached_prompt_text_token_price: 0,
		Completion_text_token_price:    0,
	})

	return nil
}

func (mst *LLMMistral) GetPricingString(model string) string {
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

func (model *LLMMistralLanguageModel) GetTextPrice(in, reason, cached, out int) (float64, float64, float64, float64) {

	convert_to_dolars := float64(10000)

	Input_price := float64(model.Prompt_text_token_price) / convert_to_dolars / 1000000
	Reason_price := float64(model.Prompt_text_token_price) / convert_to_dolars / 1000000
	Cached_price := float64(model.Cached_prompt_text_token_price) / convert_to_dolars / 1000000
	Output_price := float64(model.Completion_text_token_price) / convert_to_dolars / 1000000

	return float64(in) * Input_price, float64(reason) * Reason_price, float64(cached) * Cached_price, float64(out) * Output_price
}
