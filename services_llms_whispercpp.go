package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sync"
)

// Whisper.cpp settings.
type LLMWhispercpp struct {
	lock sync.Mutex

	Address string
	Port    int
}

func (wsp *LLMWhispercpp) Check() error {
	if wsp.Address == "" {
		return LogsErrorf("whispercpp address is empty")
	}

	return nil
}
func (wsp *LLMWhispercpp) GetUrlInference() string {
	return fmt.Sprintf("%s:%d/inference", wsp.Address, wsp.Port)
}
func (wsp *LLMWhispercpp) GetUrlLoadModel() string {
	return fmt.Sprintf("%s:%d/load", wsp.Address, wsp.Port)
}

func (wsp *LLMWhispercpp) Transcribe(st *LLMTranscribe) error {
	err := wsp.Check()
	if err != nil {
		return err
	}

	st.Out_Output, st.Out_StatusCode, err = LLMWhispercppTranscribe_Transcribe(st.AudioBlob, "", st.Temperature, st.Response_format, wsp)
	if err != nil {
		return err
	}

	return nil
}

func LLMWhispercppTranscribe_Transcribe(blob []byte, model string, temperature float64, response_format string, source *LLMWhispercpp) ([]byte, int, error) {
	source.lock.Lock()
	defer source.lock.Unlock()

	//set parameters
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	{
		part, err := writer.CreateFormFile("file", "blob.wav")
		if LogsError(err) != nil {
			return nil, -1, err
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
	if LogsError(err) != nil {
		return nil, -1, err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	res, err := client.Do(req)
	if LogsError(err) != nil {
		return nil, -1, err
	}

	answer, err := io.ReadAll(res.Body) //job.close ....
	if LogsError(err) != nil {
		return nil, res.StatusCode, err
	}

	if res.StatusCode != 200 {
		return nil, res.StatusCode, LogsErrorf("statusCode != 200, response: %s", answer)
	}

	return answer, res.StatusCode, nil
}
