package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// [ignore]
type WhispercppTranscribe struct {
	AudioBlob    []byte
	BlobFileName string //ext.... (blob.wav, blob.mp3)

	Model           string  //empty = default
	Temperature     float64 //0
	Response_format string

	Out_StatusCode int
	Out_Output     []byte
}

func (st *WhispercppTranscribe) run(caller *ToolCaller, ui *UI) error {
	source_wsp, err := NewLLMWhispercpp_wsp("", caller)
	if err != nil {
		return err
	}

	source_wsp.Check()

	//Set model
	model := st.Model
	if model == "" {
		model = source_wsp.ModelName
	}
	_, err = st.SetModel(model, source_wsp)
	if err != nil {
		return fmt.Errorf("SetModel() failed: %w", err)
	}

	//Transcribe
	st.Out_Output, st.Out_StatusCode, err = st.Transcribe(st.AudioBlob, model, st.Temperature, st.Response_format, source_wsp)
	if err != nil {
		return err
	}

	return nil
}

func (st *WhispercppTranscribe) Transcribe(blob []byte, model string, temperature float64, response_format string, source *LLMWhispercpp) ([]byte, int, error) {
	source.lock.Lock()
	defer source.lock.Unlock()

	//set parameters
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	{
		part, err := writer.CreateFormFile("file", "blob.wav")
		if err != nil {
			return nil, -1, fmt.Errorf("CreateFormFile() failed: %w", err)
		}
		part.Write(blob)

		//write
		writer.WriteField("temperature", fmt.Sprintf("%f", temperature))
		writer.WriteField("model", model)
		writer.WriteField("response_format", response_format)
		if response_format == "verbose_json" {
			writer.WriteField("timestamp_granularities[]", "word")
		}

	}
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, source.GetUrlInference(), body)
	if err != nil {
		return nil, -1, fmt.Errorf("NewRequest() failed: %w", err)
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, -1, fmt.Errorf("Do() failed: %w", err)
	}

	answer, err := io.ReadAll(res.Body) //job.close ....
	if err != nil {
		return nil, res.StatusCode, fmt.Errorf("ReadAll() failed: %w", err)
	}

	if res.StatusCode != 200 {
		return nil, res.StatusCode, fmt.Errorf("statusCode != 200, response: %s", answer)
	}

	return answer, res.StatusCode, nil
}

func (st *WhispercppTranscribe) SetModel(model string, source *LLMWhispercpp) (int, error) {
	source.lock.Lock()
	defer source.lock.Unlock()

	if model == source.ModelName {
		return 200, nil //already set
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("model", source.getModelPath(model))
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, source.GetUrlLoadModel(), body)
	if err != nil {
		return -1, fmt.Errorf("NewRequest() failed: %w", err)
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return -1, fmt.Errorf("Do() failed: %w", err)
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, fmt.Errorf("ReadAll() failed: %w", err)
	}

	if res.StatusCode != 200 {
		return res.StatusCode, fmt.Errorf("statusCode != 200, response: %s", resBody)
	}

	source.ModelName = model

	return res.StatusCode, nil
}
