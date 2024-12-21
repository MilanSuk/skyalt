package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type OpenAI_chat struct {
	UID        string
	Properties OpenAI_chat_props

	Out  string
	done func()
}

func (layout *Layout) AddOpenAI_chat(x, y, w, h int, props *OpenAI_chat) *OpenAI_chat {
	layout._createDiv(x, y, w, h, "OpenAI_chat", props.Build, nil, nil)
	return props
}

var g_global_OpenAI_chat = make(map[string]*OpenAI_chat)

func NewGlobal_OpenAI_chat(uid string) *OpenAI_chat {
	uid = fmt.Sprintf("OpenAI_chat:%s", uid)

	st, found := g_global_OpenAI_chat[uid]
	if !found {
		st = &OpenAI_chat{UID: uid}
		st.Properties.Reset()
		g_global_OpenAI_chat[uid] = st
	}
	return st
}

func (st *OpenAI_chat) Build(layout *Layout) {

	layout.SetColumn(0, 1, 100)
	layout.SetColumn(1, 1, 3)
	layout.SetRow(0, 1, 10)

	job := FindJob(st.UID)

	txt, txtLay := layout.AddTextMultiline(0, 0, 2, 1, "")
	txt.Align_h = 0
	if job != nil {
		txt.Value = job.info
		txtLay.VScrollToTheBottom()
	}

	stopBt := layout.AddButton(1, 1, 1, 1, NewButton("Stop"))
	stopBt.clicked = func() {
		st.Stop()
	}
}

func (st *OpenAI_chat) Start() *Job {
	return StartJob(st.UID, "OpenAI chat completion", st.Run)
}
func (st *OpenAI_chat) Stop() {
	job := FindJob(st.UID)
	if job != nil {
		job.Stop()
	}
}
func (st *OpenAI_chat) IsRunning() bool {
	return FindJob(st.UID) != nil
}

func (st *OpenAI_chat) Run(job *Job) {

	if !NewFile_OpenAI().Enable {
		job.AddError(errors.New("OpenAI is disabled"))
		return
	}

	st.Out = ""

	if st.Properties.Model == "" {
		job.AddError(errors.New("model is empty"))
		return
	}

	st.Properties.Stream = true
	jsProps, err := json.Marshal(&st.Properties)
	if err != nil {
		job.AddError(fmt.Errorf("Marshal() failed: %w", err))
		return
	}

	st.Out, err = OpenAI_chat_Complete(jsProps, "https://api.openai.com/v1/chat/completions", NewFile_OpenAI().Api_key, job)
	if err != nil {
		job.AddError(err)
		return
	}

	if st.done != nil {
		st.done()
	}
}

func OpenAI_chat_Complete(jsProps []byte, Completion_url string, Api_key string, job *Job) (string, error) {

	body := bytes.NewReader(jsProps)

	req, err := http.NewRequest(http.MethodPost, Completion_url, body)
	if err != nil {
		return "", fmt.Errorf("NewRequest() failed: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+Api_key)
	//req.Header.Set("Accept", "text/event-stream")
	//req.Header.Set("Cache-Control", "no-cache")
	//req.Header.Set("Connection", "keep-alive")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Do() failed: %w", err)
	}
	defer res.Body.Close()

	answer, err := OpenAI_chat_parseStream(res, job)
	if err != nil {
		return "", err
	}

	if res.StatusCode != 200 {
		return "", fmt.Errorf("statusCode %d != 200, response: %s", res.StatusCode, answer)
	}

	return answer, nil
}

func OpenAI_chat_parseStream(res *http.Response, job *Job) (string, error) {
	// https://platform.openai.com/docs/api-reference/assistants-streaming/events
	type STMsg struct {
		Content string
	}
	type STChoice struct {
		Message STMsg
		Delta   STMsg
	}
	type ST struct {
		Choices []STChoice
	}

	job.info = ""
	buff := make([]byte, 0, 1024)
	buff_last := 0
	for job.IsRunning() {
		var tb [256]byte
		n, readErr := res.Body.Read(tb[:])
		if readErr != nil && readErr != io.EOF {
			return "", fmt.Errorf("Read() failed: %w", readErr)
		}
		if n > 0 {
			buff = append(buff, tb[:n]...)
		}

		for {
			strBlock := string(buff[buff_last:])

			separ := "\n\n"
			d := strings.Index(strBlock, separ)
			if d < 0 {
				break
			}
			buff_last += d + len(separ)

			strData := strBlock[:d] //cut end

			js, found := strings.CutPrefix(strData, "data:") //cut start
			if !found {
				return "", fmt.Errorf("missing 'data:'")
			}
			js = strings.TrimSpace(js)

			if js != "[DONE]" {
				var st ST
				err := json.Unmarshal([]byte(js), &st)
				if err != nil {
					return "", fmt.Errorf("Unmarshal() failed: %w", err)
				}

				if len(st.Choices) > 0 {
					job.info = job.info + st.Choices[0].Delta.Content
					fmt.Print(st.Choices[0].Delta.Content)
				}
			}
		}

		if readErr != nil && readErr == io.EOF {
			break //done
		}
	}

	if !job.IsRunning() {
		return "", fmt.Errorf("user Canceled the job")
	}

	return job.info, nil
}
