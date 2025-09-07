package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
)

type LLMxAILanguageModel struct {
	Id               string
	Created          int64
	Version          string
	Input_modalities []string //"text", "image"

	Prompt_text_token_price        int //USD cents per million tokens
	Prompt_image_token_price       int
	Cached_prompt_text_token_price int
	Completion_text_token_price    int
	Search_source_price            int //USD cents per thousand tokens

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

// xAI LLM settings.
type LLMxAI struct {
	Provider   string
	OpenAI_url string
	DevUrl     string
	API_key    string

	LanguageModels []*LLMxAILanguageModel
	ImageModels    []*LLMxAIImageModel

	Stats []LLMMsgStats
}

func (xai *LLMxAI) Check() error {

	if xai.API_key == "" {
		return LogsErrorf("%s API key is empty", xai.Provider)
	}

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

func (model *LLMxAILanguageModel) GetTextPrice(in, reason, cached, out int, sources int) (float64, float64, float64, float64, float64) {

	convert_to_dolars := float64(10000)

	Input_price := float64(model.Prompt_text_token_price) / convert_to_dolars / 1000000
	Reason_price := float64(model.Prompt_text_token_price) / convert_to_dolars / 1000000
	Cached_price := float64(model.Cached_prompt_text_token_price) / convert_to_dolars / 1000000
	Output_price := float64(model.Completion_text_token_price) / convert_to_dolars / 1000000
	Source_price := float64(model.Search_source_price) / convert_to_dolars / 1000 //1K

	return float64(in) * Input_price, float64(reason) * Reason_price, float64(cached) * Cached_price, float64(out) * Output_price, float64(sources) * Source_price
}

/*func (xai *LLMxAI) downloadList(url_part string) ([]byte, error) {
	if xai.API_key == "" {
		return nil, LogsErrorf("%s API key is empty", xai.Provider)
	}

	Completion_url := xai.OpenAI_url
	if !strings.HasSuffix(Completion_url, "/") {
		Completion_url += "/"
	}
	Completion_url += url_part

	body := bytes.NewReader(nil)
	req, err := http.NewRequest(http.MethodGet, Completion_url, body) //http.MethodPost
	if LogsError(err) != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+xai.API_key)

	client := &http.Client{}
	res, err := client.Do(req)
	if LogsError(err) != nil {
		return nil, err
	}
	defer res.Body.Close()

	js, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return js, nil
}*/

func (xai *LLMxAI) GenerateImage(st *LLMGenerateImage, msg *AppsRouterMsg) error {
	err := xai.Check()
	if err != nil {
		return err
	}

	model := "grok-2-image"

	props := OpenAI_getImage_props{
		Model:           model,
		N:               st.Num_images,
		Response_format: "b64_json",
	}

	jsProps, err := LogsJsonMarshal(props)
	if err != nil {
		return err
	}
	out, status, dt, err := OpenAI_genImage_Run(jsProps, xai.OpenAI_url, xai.API_key)
	st.Out_dtime_sec = dt
	st.Out_StatusCode = status
	if err != nil {
		return err
	}

	for _, it := range out.Data {
		img, err := base64.StdEncoding.DecodeString(it.B64_json)
		if err != nil {
			return err
		}

		sep := bytes.IndexByte(img, ',')
		if sep < 0 {
			sep = 0
		}

		st.Out_images = append(st.Out_images, img[sep:])
		st.Out_revised_prompts = append(st.Out_revised_prompts, it.Revised_prompt)
	}

	return nil
}

func (xai *LLMxAI) Complete(st *LLMComplete, app_port int, tools []*ToolsOpenAI_completion_tool, msg *AppsRouterMsg) error {
	err := xai.Check()
	if err != nil {
		return err
	}

	mod, _ := xai.FindModel(st.Out_usage.Model)
	if mod == nil {
		return fmt.Errorf("model '%s' not found", st.Out_usage.Model)
	}

	stats, err := OpenAI_Complete(xai.Provider, xai.OpenAI_url, xai.API_key, st, app_port, tools, msg, mod.GetTextPrice)
	xai.Stats = append(xai.Stats, stats...)
	return err
}
