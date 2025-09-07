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

// Mistral LLM settings.
type LLMMistral struct {
	Provider   string
	OpenAI_url string
	DevUrl     string
	API_key    string

	LanguageModels []*LLMMistralLanguageModel
	ImageModels    []*LLMMistralImageModel

	Stats []LLMMsgStats
}

func (mst *LLMMistral) Check() error {
	if mst.API_key == "" {
		return LogsErrorf("%s API key is empty", mst.Provider)
	}

	return nil
}

func (mst *LLMMistral) FindModel(name string) (*LLMMistralLanguageModel, *LLMMistralImageModel) {
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

/*func (mst *LLMMistral) ReloadModels() error {

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
	mst.LanguageModels = append(mst.LanguageModels, &LLMMistralLanguageModel{
		Id:                             "pixtral-large-latest",
		Input_modalities:               []string{"text", "image"},
		Prompt_text_token_price:        20000,
		Cached_prompt_text_token_price: 20000,
		Completion_text_token_price:    60000,
	})

	mst.LanguageModels = append(mst.LanguageModels, &LLMMistralLanguageModel{
		Id:                             "codestral-latest",
		Input_modalities:               []string{"text"},
		Prompt_text_token_price:        3000,
		Cached_prompt_text_token_price: 3000,
		Completion_text_token_price:    9000,
	})

	mst.LanguageModels = append(mst.LanguageModels, &LLMMistralLanguageModel{
		Id:                             "mistral-large-latest",
		Input_modalities:               []string{"text"},
		Prompt_text_token_price:        20000,
		Cached_prompt_text_token_price: 20000,
		Completion_text_token_price:    60000,
	})

	return nil
}*/

func (mst *LLMMistral) GetPricingString(model string) string {
	model = strings.ToLower(model)

	convert_to_dolars := float64(10000)

	lang, img := mst.FindModel(model)
	if lang != nil {
		//in, cached, out, image
		return fmt.Sprintf("$%.2f/$%.2f/$%.2f/$%.2f", float64(lang.Prompt_text_token_price)/convert_to_dolars, float64(lang.Prompt_image_token_price)/convert_to_dolars, float64(lang.Cached_prompt_text_token_price)/convert_to_dolars, float64(lang.Completion_text_token_price)/convert_to_dolars)
	}

	if img != nil {
		return fmt.Sprintf("$%.2f", float64(img.Image_price)/convert_to_dolars)
	}

	return fmt.Sprintf("model %s not found", model)
}

func (model *LLMMistralLanguageModel) GetTextPrice(in, reason, cached, out int, sources int) (float64, float64, float64, float64, float64) {
	convert_to_dolars := float64(10000)

	Input_price := float64(model.Prompt_text_token_price) / convert_to_dolars / 1000000
	Reason_price := float64(model.Prompt_text_token_price) / convert_to_dolars / 1000000
	Cached_price := float64(model.Cached_prompt_text_token_price) / convert_to_dolars / 1000000
	Output_price := float64(model.Completion_text_token_price) / convert_to_dolars / 1000000

	return float64(in) * Input_price, float64(reason) * Reason_price, float64(cached) * Cached_price, float64(out) * Output_price, 0
}

func (mst *LLMMistral) Complete(st *LLMComplete, app_port int, tools []*ToolsOpenAI_completion_tool, msg *AppsRouterMsg) error {
	err := mst.Check()
	if err != nil {
		return err
	}

	mod, _ := mst.FindModel(st.Out_usage.Model)
	if mod == nil {
		return fmt.Errorf("model '%s' not found", st.Out_usage.Model)
	}

	stats, err := OpenAI_Complete(mst.Provider, mst.OpenAI_url, mst.API_key, st, app_port, tools, msg, mod.GetTextPrice)
	mst.Stats = append(mst.Stats, stats...)
	return err
}
