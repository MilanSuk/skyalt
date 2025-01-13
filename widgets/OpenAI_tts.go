package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type OpenAI_tts struct {
	UID string

	Properties OpenAI_tts_props

	Out  []byte
	done func()
}

func (layout *Layout) AddOpenAI_tts(x, y, w, h int, props *OpenAI_tts) *OpenAI_tts {
	layout._createDiv(x, y, w, h, "OpenAI_tts", props.Build, nil, nil)
	return props
}

func OpenMemory_OpenAI_tts(uid string) *OpenAI_tts {
	st := &OpenAI_tts{UID: uid}
	st.Properties.Reset()
	return OpenMemory(uid, st)
}

func (st *OpenAI_tts) Build(layout *Layout) {

	layout.SetColumn(0, 1, 100)
	layout.SetColumn(1, 1, 3)
	layout.SetRowFromSub(0, 1, 100)

	job := FindJob(st.UID)

	txt := layout.AddTextMultiline(0, 0, 2, 1, "")

	if job != nil {
		txt.Value = job.info
		txt.ScrollToEnd = true
	}

	stopBt := layout.AddButton(1, 1, 1, 1, "Stop")
	stopBt.clicked = func() {
		st.Stop()
	}
}

func (st *OpenAI_tts) Start() *Job {

	return StartJob(st.UID, "OpenAI text-to-speech", st.Run)
}
func (st *OpenAI_tts) Stop() {
	job := FindJob(st.UID)
	if job != nil {
		job.Stop()
	}
}
func (st *OpenAI_tts) FindJob() *Job {
	return FindJob(st.UID)
}
func (st *OpenAI_tts) Run(job *Job) {

	if !OpenFile_OpenAI().Enable {
		job.AddError(errors.New("OpenAI is disabled"))
		return
	}

	st.Out = nil

	if st.Properties.Model == "" {
		job.AddError(errors.New("model is empty"))
		return
	}
	if st.Properties.Voice == "" {
		job.AddError(errors.New("voice is empty"))
		return
	}

	js, err := json.Marshal(st.Properties)
	if err != nil {
		job.AddError(err)
		return
	}
	body := bytes.NewReader(js)

	req, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/audio/speech", body)
	if err != nil {
		job.AddError(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+OpenFile_OpenAI().Api_key)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		job.AddError(err)
		return
	}
	defer res.Body.Close()

	answer, err := io.ReadAll(res.Body)
	if err != nil {
		job.AddError(err)
		return
	}

	if res.StatusCode != 200 {
		job.AddError(fmt.Errorf("statusCode %d != 200, response: %s", res.StatusCode, answer))
		return
	}

	st.Out = answer

	if st.done != nil {
		st.done()
	}
}
