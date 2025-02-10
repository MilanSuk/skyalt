/*
Copyright 2025 Milan Suk

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this db except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

var g_globa_Whispercpp_lock sync.Mutex
var g_globa_Whispercpp_model string

func Whispercpp_getModelPath(model_name string) string {
	return filepath.Join("models/", model_name+".bin")
}

func Whispercpp_SetModel(model string, wsp *Whispercpp) (int, error) {
	g_globa_Whispercpp_lock.Lock()
	defer g_globa_Whispercpp_lock.Unlock()

	if model == g_globa_Whispercpp_model {
		return 200, nil //already set
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("model", Whispercpp_getModelPath(model))
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, wsp.GetUrlLoadModel(), body)
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

	g_globa_Whispercpp_model = model
	return res.StatusCode, nil
}

func Whispercpp_Transcribe(input []byte, props Whispercpp_props, wsp *Whispercpp) ([]byte, int, error) {
	//set model
	_, err := Whispercpp_SetModel(props.Model, wsp)
	if err != nil {
		return nil, -1, fmt.Errorf("SetModel() failed: %w", err)
	}

	g_globa_Whispercpp_lock.Lock()
	defer g_globa_Whispercpp_lock.Unlock()

	//set parameters
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	{
		part, err := writer.CreateFormFile("file", "blob.wav")
		if err != nil {
			return nil, -1, fmt.Errorf("CreateFormFile() failed: %w", err)
		}
		part.Write(input)
		props.Write(writer)
	}
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, wsp.GetUrlInference(), body)
	if err != nil {
		return nil, -1, fmt.Errorf("NewRequest() failed: %w", err)
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, -1, fmt.Errorf("Do() failed: %w", err)
	}

	answer, err := io.ReadAll(res.Body) //job.close ...
	if err != nil {
		return nil, res.StatusCode, fmt.Errorf("ReadAll() failed: %w", err)
	}

	if res.StatusCode != 200 {
		return nil, res.StatusCode, fmt.Errorf("statusCode != 200, response: %s", answer)
	}

	return answer, res.StatusCode, nil
}

func Whispercpp_convertIntoFile(input *audio.IntBuffer, mp3 bool) ([]byte, error) {

	path := "temp/mic.wav"

	//file := &OsWriterSeeker{}	//....
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	enc := wav.NewEncoder(file, input.Format.SampleRate, 16, input.Format.NumChannels, 1)
	err = enc.Write(input)
	if err != nil {
		enc.Close()
		file.Close()
		return nil, err
	}
	enc.Close()
	file.Close()

	if mp3 {
		compress_path := "temp/mic.mp3"
		err := FFMpeg_convert(path, compress_path)
		if err != nil {
			return nil, err
		}
		path = compress_path
	} else {
		resample_path := "temp/mic2.wav"
		err := FFMpeg_convert(path, resample_path)
		if err != nil {
			return nil, err
		}
		path = resample_path
	}

	buff, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return buff, nil
}

func FFMpeg_convert(src, dst string) error {
	OsFileRemove(dst) //ffmpeg complains that 'file already exists'

	cmd := exec.Command("ffmpeg", "-i", src, "-ar", "16000", dst)
	cmd.Dir = ""
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
