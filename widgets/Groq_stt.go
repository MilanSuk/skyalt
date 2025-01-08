package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

type Groq_stt struct {
	UID string

	Properties Groq_stt_props

	Input_Data       audio.IntBuffer
	Input_SampleRate int
	Input_Channels   int

	Out  string
	done func(out string)
}

func (layout *Layout) AddGroq_stt(x, y, w, h int, props *Groq_stt) *Groq_stt {
	layout._createDiv(x, y, w, h, "Groq_stt", props.Build, nil, nil)
	return props
}

var g_global_Groq_stt = make(map[string]*Groq_stt)

func NewGlobal_Groq_stt(uid string) *Groq_stt {
	uid = fmt.Sprintf("Groq_stt:%s", uid)

	st, found := g_global_Groq_stt[uid]
	if !found {
		st = &Groq_stt{UID: uid}
		st.Properties.Reset()
		st.Input_Channels = OpenFile_Microphone().Channels
		st.Input_SampleRate = OpenFile_Microphone().SampleRate

		g_global_Groq_stt[uid] = st
	}
	return st
}

func (st *Groq_stt) Build(layout *Layout) {
	layout.SetColumn(0, 1, 100)
	layout.SetColumn(1, 1, 3)
	layout.SetRow(0, 1, 10)

	job := FindJob(st.UID)

	txt := layout.AddTextMultiline(0, 0, 2, 1, "")
	txt.Align_h = 0
	if job != nil {
		txt.Value = job.info
		txt.ScrollToEnd = true
	}

	stopBt := layout.AddButton(1, 1, 1, 1, "Stop")
	stopBt.clicked = func() {
		st.Stop()
	}
}

func (st *Groq_stt) Start() *Job {
	return StartJob(st.UID, "Groq speech-to-text", st.Run)
}
func (st *Groq_stt) Stop() {
	job := FindJob(st.UID)
	if job != nil {
		job.Stop()
	}
}
func (st *Groq_stt) IsRunning() bool {
	return FindJob(st.UID) != nil
}

func (st *Groq_stt) Run(job *Job) {

	oai := OpenFile_Groq()

	if !oai.Enable {
		job.AddError(errors.New("Groq is disabled"))
		return
	}

	st.Out = ""

	if st.Properties.Model == "" {
		job.AddError(errors.New("model is empty"))
		return
	}

	if len(st.Input_Data.Data) == 0 {
		return
	}

	//convert stt.Input into buff(mp3)
	var buff []byte
	{
		os.MkdirAll("temp", os.ModePerm)
		path := "temp/mic.wav"
		compress_path := "temp/mic.mp3"

		//encode & save
		{
			//file := &OsWriterSeeker{}
			file, err := os.Create(path)
			if err != nil {
				job.AddError(err)
				return
			}

			enc := wav.NewEncoder(file, st.Input_SampleRate, 16, st.Input_Channels, 1)
			err = enc.Write(&st.Input_Data)
			if err != nil {
				enc.Close()
				file.Close()
				job.AddError(err)
				return
			}
			enc.Close()
			file.Close()
		}

		err := Groq_stt_ffmpeg_convert(path, compress_path)
		if err != nil {
			job.AddError(err)
			return
		}
		path = compress_path

		buff, err = os.ReadFile(path)
		if err != nil {
			job.AddError(err)
			return
		}
	}

	//set parameters
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	{
		part, err := writer.CreateFormFile("file", "blob.mp3")
		if err != nil {
			job.AddError(fmt.Errorf("CreateFormFile() failed: %w", err))
			return
		}
		part.Write(buff)
		st.Properties.Write(writer)
	}
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, "https://api.groq.com/openai/v1/audio/transcriptions", body)
	if err != nil {
		job.AddError(fmt.Errorf("NewRequest() failed: %w", err))
		return
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Add("Authorization", "Bearer "+oai.Api_key)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		job.AddError(fmt.Errorf("Do() failed: %w", err))
		return
	}

	answer, err := io.ReadAll(res.Body) //job.close ...
	if err != nil {
		job.AddError(fmt.Errorf("ReadAll() failed: %E", err))
		return
	}

	if res.StatusCode != 200 {
		job.AddError(fmt.Errorf("statusCode %d != 200, response: %s", res.StatusCode, answer))
		return
	}

	st.Out = string(answer)

	if st.done != nil {
		st.done(st.Out)
	}
}

func Groq_stt_ffmpeg_convert(src, dst string) error {
	os.Remove(dst) //ffmpeg complains that 'file already exists'

	cmd := exec.Command("ffmpeg", "-i", src, dst)
	cmd.Dir = ""
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
