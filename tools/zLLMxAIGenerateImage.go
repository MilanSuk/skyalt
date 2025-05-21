package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// [ignore]
type LLMxAIGenerateImage struct {
	Model      string //Model name to be used.
	Prompt     string //Prompt for image generation.
	Num_images int    //Number of images to be generated

	Out_StatusCode      int
	Out_images          [][]byte
	Out_revised_prompts []string
	Out_dtime_sec       float64
}

func (st *LLMxAIGenerateImage) run(caller *ToolCaller, ui *UI) error {
	source_llm, err := NewLLMxAI("", caller)
	if err != nil {
		return err
	}

	props := OpenAI_getImage_props{
		Model:           st.Model,
		N:               st.Num_images,
		Response_format: "b64_json",
	}

	jsProps, err := json.Marshal(props)
	if err != nil {
		return err
	}
	out, status, dt, err := OpenAI_genImage_Run(jsProps, source_llm.OpenAI_url, source_llm.API_key)
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

type OpenAI_getImage_props struct {
	Prompt string `json:"prompt"`
	Model  string `json:"model"`
	N      int    `json:"n"`

	Response_format string `json:"response_format"` //Response format to return the image in. Can be url or b64_json.

	//Quality string `json:"quality"` //Quality of the image.
	//Size    string `json:"size"`    //Size of the image.
	//Style   string `json:"style"`   //Style of the image.
	//User    string `json:"user"`    //A unique identifier representing your end-user, which can help xAI to monitor and detect abuse.
}

var g_global_OpenAI_genImage_lock sync.Mutex

type OpenAIGenImageOutData struct {
	B64_json       string //"data:image/png;base64,..."
	Revised_prompt string
}

type OpenAIGenImage_UsageDetails struct {
	Reasoning_tokens int
}
type OpenAIGenImage_Usage struct {
	Prompt_tokens       int
	Input_cached_tokens int
	Completion_tokens   int
	Total_tokens        int

	Completion_tokens_details OpenAIGenImage_UsageDetails
}

type OpenAIGenImage_Error struct {
	Message string
}

type OpenAIGenImageOut struct {
	Data  []OpenAIGenImageOutData
	Usage OpenAIGenImage_Usage
	Error *OpenAIGenImage_Error
}

func OpenAI_genImage_Run(jsProps []byte, Completion_url string, api_key string) (OpenAIGenImageOut, int, float64, error) {
	g_global_OpenAI_genImage_lock.Lock()
	defer g_global_OpenAI_genImage_lock.Unlock()

	st := time.Now().UnixMicro()

	if !strings.HasSuffix(Completion_url, "/") {
		Completion_url += "/"
	}
	Completion_url += "images/generations"

	body := bytes.NewReader(jsProps)

	req, err := http.NewRequest(http.MethodPost, Completion_url, body)
	if err != nil {
		return OpenAIGenImageOut{}, -1, 0, fmt.Errorf("NewRequest() failed: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+api_key)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return OpenAIGenImageOut{}, -1, 0, fmt.Errorf("Do() failed: %w", err)
	}
	defer res.Body.Close()

	js, err := io.ReadAll(res.Body)
	if err != nil {
		return OpenAIGenImageOut{}, res.StatusCode, 0, err
	}

	if len(js) == 0 {
		return OpenAIGenImageOut{}, res.StatusCode, 0, fmt.Errorf("output is empty")
	}

	var out OpenAIGenImageOut
	err = json.Unmarshal(js, &out)
	if err != nil {
		return OpenAIGenImageOut{}, res.StatusCode, 0, fmt.Errorf("%w. %s", err, string(js))
	}
	if out.Error != nil && out.Error.Message != "" {
		return OpenAIGenImageOut{}, res.StatusCode, 0, errors.New(out.Error.Message)
	}

	if res.StatusCode != 200 {
		return OpenAIGenImageOut{}, res.StatusCode, 0, fmt.Errorf("statusCode %d != 200, response: %s", res.StatusCode, string(js))
	}

	tm := float64(time.Now().UnixMicro()-st) / 1000000

	return out, res.StatusCode, tm, nil
}
