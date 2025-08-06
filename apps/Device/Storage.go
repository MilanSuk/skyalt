package main

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type DeviceSettingsPalette struct {
	P, S, E, B         color.RGBA
	OnP, OnS, OnE, OnB color.RGBA
}

// All device settings, this include date format, volume, dpi, fullscreen mode, theme.
type DeviceSettings struct {
	DateFormat  string
	Rounding    float64
	ScrollThick float64
	Volume      float64

	Dpi         int
	Dpi_default int

	Fullscreen bool
	Stats      bool

	Theme string //light, dark, custom

	LightPalette  DeviceSettingsPalette
	DarkPalette   DeviceSettingsPalette
	CustomPalette DeviceSettingsPalette

	App_provider string
	App_smarter  bool
	App_model    string

	Code_provider string
	Code_smarter  bool
	Code_model    string

	Image_provider string
	Image_model    string

	STT_provider string
}

func NewDeviceSettings(file string) (*DeviceSettings, error) {
	st := &DeviceSettings{}

	//DPI
	st.Dpi = 100
	st.Dpi_default = 100

	//UI rounding
	st.Rounding = 0.2

	//Scroll
	st.ScrollThick = 0.5

	//Speaker
	st.Volume = 0.5

	//Theme
	st.Theme = "light"

	st.LightPalette = DeviceSettingsPalette{
		P:   color.RGBA{37, 100, 120, 255},
		OnP: color.RGBA{255, 255, 255, 255},

		S:   color.RGBA{170, 200, 170, 255},
		OnS: color.RGBA{255, 255, 255, 255},

		E:   color.RGBA{180, 40, 30, 255},
		OnE: color.RGBA{255, 255, 255, 255},

		B:   color.RGBA{250, 250, 250, 255},
		OnB: color.RGBA{25, 27, 30, 255},
	}

	st.DarkPalette = DeviceSettingsPalette{
		P:   color.RGBA{150, 205, 225, 255},
		OnP: color.RGBA{0, 50, 65, 255},

		S:   color.RGBA{190, 200, 205, 255},
		OnS: color.RGBA{40, 50, 55, 255},

		E:   color.RGBA{240, 185, 180, 255},
		OnE: color.RGBA{45, 45, 65, 255},

		B:   color.RGBA{25, 30, 30, 255},
		OnB: color.RGBA{230, 230, 230, 255},
	}

	st.CustomPalette = DeviceSettingsPalette{
		P:   color.RGBA{37, 100, 120, 255},
		OnP: color.RGBA{255, 255, 255, 255},

		S:   color.RGBA{170, 200, 170, 255},
		OnS: color.RGBA{255, 255, 255, 255},

		E:   color.RGBA{180, 40, 30, 255},
		OnE: color.RGBA{255, 255, 255, 255},

		B:   color.RGBA{250, 250, 250, 255},
		OnB: color.RGBA{25, 27, 30, 255},
	}

	//Date format
	{
		_, zn := time.Now().Zone()
		zn = zn / 3600
		if zn <= -3 && zn >= -10 {
			st.DateFormat = "us"
		} else {
			st.DateFormat = "eu"
		}
	}

	return LoadFile(file, "DeviceSettings", "json", st, true)
}

func DeviceSettings_GetPricingStringTooltip() string {
	return "Price of Input_text/Input_image/Input_cached/Output per 1M tokens"
}
func (st *DeviceSettings) GetPricingString(provider, model string) string {

	pricing := "unknown"

	switch strings.ToLower(provider) {
	case "xai":
		st, err := NewLLMxAI("")
		if err == nil {
			pricing = st.GetPricingString(model)
		}
	case "mistral":
		st, err := NewLLMMistral("")
		if err == nil {
			pricing = st.GetPricingString(model)
		}
	case "openai":
		st, err := NewLLMOpenai("")
		if err == nil {
			pricing = st.GetPricingString(model)
		}
	case "groq":
		st, err := NewLLMGroq("")
		if err == nil {
			pricing = st.GetPricingString(model)
		}
	}

	return pricing
}

func (st *DeviceSettings) CheckProvider(provider string) error {

	switch strings.ToLower(provider) {
	case "xai":
		st, err := NewLLMxAI("")
		if err == nil {
			return st.Check()
		}
	case "mistral":
		st, err := NewLLMMistral("")
		if err == nil {
			return st.Check()
		}
	case "openai":
		st, err := NewLLMOpenai("")
		if err == nil {
			return st.Check()
		}
	case "groq":
		st, err := NewLLMGroq("")
		if err == nil {
			return st.Check()
		}
	case "llama.cpp":
		return nil
	case "whisper.cpp":
		return nil
	}

	return fmt.Errorf("Unknown provider '%s'", provider)
}

func (st *DeviceSettings) UpdateModels() {
	switch strings.ToLower(st.App_provider) {
	case "xai":
		st.App_model = "grok-3-mini"
		if st.App_smarter {
			st.App_model = "grok-4"
		}

	case "mistral":
		st.App_model = "mistral-small-latest"
		if st.App_smarter {
			st.App_model = "mistral-large-latest"
		}

	case "openai":
		st.App_model = "gpt-4.1-mini"
		if st.App_smarter {
			st.App_model = "o4-mini"
		}

	case "groq":
		st.App_model = "qwen/qwen3-32b"
		if st.App_smarter {
			st.App_model = "qwen/qwen3-32b"
		}

	case "llama.cpp":
		st.App_model = "" //....

	}

	switch strings.ToLower(st.Code_provider) {
	case "xai":
		st.Code_model = "grok-3-mini"
		if st.Code_smarter {
			st.Code_model = "grok-4"
		}

	case "mistral":
		st.Code_model = "devstral-small-latest"
		if st.Code_smarter {
			st.Code_model = "codestral-latest"
		}

	case "openai":
		st.Code_model = "gpt-4.1-mini"
		if st.Code_smarter {
			st.Code_model = "o4-mini"
		}

	case "groq":
		st.Code_model = "qwen/qwen3-32b"
		if st.Code_smarter {
			st.Code_model = "qwen/qwen3-32b"
		}

	case "llama.cpp":
		st.Code_model = "" //....

	}
}

func DeviceSettings_getAppProviders() []string {
	return []string{"", "xAI", "Mistral", "OpenAI", "Groq", "Llama.cpp"}
}
func DeviceSettings_getImageProviders() []string {
	return []string{"", "xAI", "OpenAI"}
}
func DeviceSettings_getSTTProviders() []string {
	return []string{"", "Whisper.cpp", "OpenAI"}
}

func (st *DeviceSettings) BuildProvider(ChatDiv *UI, provider string, caller *ToolCaller) {
	ChatDia := ChatDiv.AddDialog(provider + "_settings")
	ChatDia.UI.SetColumn(0, 1, 20)
	ChatDia.UI.SetRowFromSub(0, 1, 100, true)
	found := true
	switch strings.ToLower(provider) {
	case "xai":
		ChatDia.UI.AddTool(0, 0, 1, 1, "xai", (&ShowLLMxAISettings{}).run, caller)
	case "mistral":
		ChatDia.UI.AddTool(0, 0, 1, 1, "mistral", (&ShowLLMMistralSettings{}).run, caller)
	case "groq":
		ChatDia.UI.AddTool(0, 0, 1, 1, "groq", (&ShowLLMGroqSettings{}).run, caller)
	case "openai":
		ChatDia.UI.AddTool(0, 0, 1, 1, "openai", (&ShowLLMOpenAISettings{}).run, caller)
	case "llama.cpp":
		ChatDia.UI.AddTool(0, 0, 1, 1, "llama.cpp", (&ShowLLMLlamacppSettings{}).run, caller)
	case "whisper.cpp":
		ChatDia.UI.AddTool(0, 0, 1, 1, "whisper.cpp", (&ShowLLMWhispercppSettings{}).run, caller)
	default:
		found = false
	}

	if found {
		ChatProvider := ChatDiv.AddButton(1, 1, 1, 1, provider+" Settings")
		ChatProvider.Background = 0.5

		providerErr := st.CheckProvider(provider)
		if providerErr != nil {
			ChatProvider.Cd = UI_GetPalette().E
			ChatProvider.layout.Tooltip = providerErr.Error()
		}
		ChatProvider.clicked = func() error {
			ChatDia.OpenCentered(caller)
			return nil
		}
	} else {
		errTx := ChatDiv.AddText(1, 1, 1, 1, "Disabled")
		errTx.Cd = UI_GetPalette().E
	}
}

func (st *DeviceSettings) GetPalette() *DeviceSettingsPalette {
	switch st.Theme {
	case "light":
		return &st.LightPalette
	case "dark":
		return &st.DarkPalette
	}
	return &st.CustomPalette
}

func (st *DeviceSettings) SetDPI(dpi int) {
	//check range
	if dpi < 30 {
		dpi = 30
	}
	if dpi > 600 {
		dpi = 600
	}
	st.Dpi = dpi
}

// Map settings.
type MapSettings struct {
	Enable    bool
	Tiles_url string

	Copyright     string
	Copyright_url string
}

func NewMapSettings(file string) (*MapSettings, error) {
	st := &MapSettings{}

	st.Enable = true
	st.Tiles_url = "https://tile.openstreetmap.org/{z}/{x}/{y}.png"
	st.Copyright = "(c)OpenStreetMap contributors"
	st.Copyright_url = "https://www.openstreetmap.org/copyright"

	return LoadFile(file, "MapSettings", "json", st, true)
}

type MicrophoneSettings struct {
	Enable      bool
	Sample_rate int
	Channels    int
}

func NewMicrophoneSettings(file string) (*MicrophoneSettings, error) {
	st := &MicrophoneSettings{}

	st.Enable = true
	st.Sample_rate = 44100
	st.Channels = 1

	return LoadFile(file, "MicrophoneSettings", "json", st, true)
}

type LLMLlamacppMsgStats struct {
	Function string
	Usage    LLMMsgUsage
}

// Llama.cpp settings.
type LLMLlamacpp struct {
	lock sync.Mutex

	Address string
	Port    int

	Stats []LLMLlamacppMsgStats
}

func NewLLMLlamacpp(file string) (*LLMLlamacpp, error) {
	st := &LLMLlamacpp{}

	st.Address = "http://localhost"
	st.Port = 8070

	return LoadFile(file, "LLMLlamacpp", "json", st, true)
}

func (wsp *LLMLlamacpp) Check() error {
	if wsp.Address == "" {
		return fmt.Errorf("llama.cpp address is empty")
	}

	return nil
}

func (wsp *LLMLlamacpp) GetUrlHealth() string {
	return fmt.Sprintf("%s:%d/health", wsp.Address, wsp.Port)
}

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
	Function string
	Usage    LLMMsgUsage
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

	var err error
	mst, err = LoadFile(file, "LLMMistral", "json", mst, true)
	if err == nil {
		mst.ReloadModels()
	}
	return mst, err
}

func (mst *LLMMistral) Check() error {

	if mst.API_key == "" {
		return fmt.Errorf("%s API key is empty", mst.Provider)
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

func (mst *LLMMistral) ReloadModels() error {

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
}

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

func (model *LLMMistralLanguageModel) GetTextPrice(in, reason, cached, out int) (float64, float64, float64, float64) {

	convert_to_dolars := float64(10000)

	Input_price := float64(model.Prompt_text_token_price) / convert_to_dolars / 1000000
	Reason_price := float64(model.Prompt_text_token_price) / convert_to_dolars / 1000000
	Cached_price := float64(model.Cached_prompt_text_token_price) / convert_to_dolars / 1000000
	Output_price := float64(model.Completion_text_token_price) / convert_to_dolars / 1000000

	return float64(in) * Input_price, float64(reason) * Reason_price, float64(cached) * Cached_price, float64(out) * Output_price
}

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

// Groq LLM settings.
type LLMGroq struct {
	Provider   string
	OpenAI_url string
	DevUrl     string
	API_key    string

	LanguageModels []*LLMGroqLanguageModel
	ImageModels    []*LLMGroqImageModel

	Stats []LLMGroqMsgStats
}

func NewLLMGroq(file string) (*LLMGroq, error) {
	mst := &LLMGroq{}

	mst.Provider = "Groq"
	mst.OpenAI_url = "https://api.groq.com/openai/v1"
	mst.DevUrl = "https://console.groq.com/keys"

	var err error
	mst, err = LoadFile(file, "LLMGroq", "json", mst, true)
	if err == nil {
		mst.ReloadModels()
	}
	return mst, err
}

func (mst *LLMGroq) Check() error {

	if mst.API_key == "" {
		return fmt.Errorf("%s API key is empty", mst.Provider)
	}

	return nil
}

func (mst *LLMGroq) FindModel(name string) (*LLMGroqLanguageModel, *LLMGroqImageModel) {
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

func (mst *LLMGroq) ReloadModels() error {

	//reset
	mst.LanguageModels = nil
	mst.ImageModels = nil

	mst.LanguageModels = append(mst.LanguageModels, &LLMGroqLanguageModel{
		Id:                             "qwen/qwen3-32b",
		Input_modalities:               []string{"text"},
		Prompt_text_token_price:        0,
		Cached_prompt_text_token_price: 0,
		Completion_text_token_price:    0,
	})

	return nil
}

func (mst *LLMGroq) GetPricingString(model string) string {
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

func (model *LLMGroqLanguageModel) GetTextPrice(in, reason, cached, out int) (float64, float64, float64, float64) {

	convert_to_dolars := float64(10000)

	Input_price := float64(model.Prompt_text_token_price) / convert_to_dolars / 1000000
	Reason_price := float64(model.Prompt_text_token_price) / convert_to_dolars / 1000000
	Cached_price := float64(model.Cached_prompt_text_token_price) / convert_to_dolars / 1000000
	Output_price := float64(model.Completion_text_token_price) / convert_to_dolars / 1000000

	return float64(in) * Input_price, float64(reason) * Reason_price, float64(cached) * Cached_price, float64(out) * Output_price
}

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
		Id:                             "o4-mini",
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

// Whisper.cpp settings.
type LLMWhispercpp struct {
	lock sync.Mutex

	Address string
	Port    int
}

func NewLLMWhispercpp(file string) (*LLMWhispercpp, error) {
	st := &LLMWhispercpp{}

	st.Address = "http://localhost"
	st.Port = 8090

	return LoadFile(file, "LLMWhispercpp", "json", st, true)
}

func (wsp *LLMWhispercpp) Check() error {
	if wsp.Address == "" {
		return fmt.Errorf("whisper.cpp address is empty")
	}

	return nil
}
func (wsp *LLMWhispercpp) getModelPath(model_name string) string {
	return filepath.Join("models", model_name+".bin")
}

func (wsp *LLMWhispercpp) GetUrlInference() string {
	return fmt.Sprintf("%s:%d/inference", wsp.Address, wsp.Port)
}
func (wsp *LLMWhispercpp) GetUrlLoadModel() string {
	return fmt.Sprintf("%s:%d/load", wsp.Address, wsp.Port)
}

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
	Function string
	Usage    LLMMsgUsage
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

	var err error
	xai, err = LoadFile(file, "LLMxAI", "json", xai, true)
	if err == nil {
		xai.ReloadModels()
	}
	return xai, err
}

func (xai *LLMxAI) Check() error {

	if xai.API_key == "" {
		return fmt.Errorf("%s API key is empty", xai.Provider)
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

func (xai *LLMxAI) ReloadModels() error {

	//reset
	xai.LanguageModels = nil
	xai.ImageModels = nil

	xai.LanguageModels = append(xai.LanguageModels, &LLMxAILanguageModel{
		Id:                             "grok-4",
		Input_modalities:               []string{"text"},
		Prompt_text_token_price:        30000,
		Cached_prompt_text_token_price: 7500,
		Completion_text_token_price:    150000,
	})

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
