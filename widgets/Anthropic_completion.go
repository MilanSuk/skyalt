package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Anthropic_completion struct {
	UID        string
	Properties Anthropic_completion_props

	Out  string
	done func(Out string)
}

func (layout *Layout) AddAnthropic_completion(x, y, w, h int, props *Anthropic_completion) *Anthropic_completion {
	layout._createDiv(x, y, w, h, "Anthropic_completion", props.Build, nil, nil)
	return props
}

var g_global_Anthropic_completion = make(map[string]*Anthropic_completion)

func NewGlobal_Anthropic_completion(uid string) *Anthropic_completion {
	uid = fmt.Sprintf("Anthropic_completion:%s", uid)

	st, found := g_global_Anthropic_completion[uid]
	if !found {
		st = &Anthropic_completion{UID: uid}
		st.Properties.Reset()

		g_global_Anthropic_completion[uid] = st
	}
	return st
}

func (st *Anthropic_completion) Build(layout *Layout) {

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

func (st *Anthropic_completion) Start() *Job {
	return StartJob(st.UID, "Anthropic chat completion", st.Run)
}
func (st *Anthropic_completion) Stop() {
	job := FindJob(st.UID)
	if job != nil {
		job.Stop()
	}
}
func (st *Anthropic_completion) IsRunning() bool {
	return FindJob(st.UID) != nil
}

func (st *Anthropic_completion) Run(job *Job) {
	if !OpenFile_Anthropic().Enable {
		job.AddError(errors.New("Anthropic is disabled"))
		return
	}

	st.Out = ""

	if st.Properties.Model == "" {
		job.AddError(errors.New("model is empty"))
		return
	}

	//st.Properties.Stream = true
	jsProps, err := json.Marshal(&st.Properties)
	if err != nil {
		job.AddError(fmt.Errorf("Marshal() failed: %w", err))
		return
	}

	st.Out, err = Anthropic_completion_Run(jsProps, st.Properties.Stream, "https://api.anthropic.com/v1/messages", OpenFile_Anthropic().Api_key, job)
	if err != nil {
		job.AddError(err)
		return
	}

	if st.done != nil {
		st.done(st.Out)
	}
}

func Anthropic_completion_Run(jsProps []byte, stream bool, Completion_url string, Api_key string, job *Job) (string, error) {
	fmt.Println("jsProps", string(jsProps))

	startTime := float64(time.Now().UnixMilli()) / 1000

	body := bytes.NewReader(jsProps)

	req, err := http.NewRequest(http.MethodPost, Completion_url, body)
	if err != nil {
		return "", fmt.Errorf("NewRequest() failed: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-api-key", Api_key)
	req.Header.Add("anthropic-version", "2023-06-01")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Do() failed: %w", err)
	}
	defer res.Body.Close()

	var answer string
	if stream {
		answer, err = Anthropic_completion_parseStream(res, job)
		if err != nil {
			return "", err
		}
	} else {
		js, err := io.ReadAll(res.Body) //job.close ...
		if err != nil {
			job.AddError(fmt.Errorf("ReadAll() failed: %E", err))
			return "", err
		}

		type STContent struct {
			Text string
			Type string
		}
		type STUsage struct {
			Input_tokens  int
			Output_tokens int
		}
		type ST struct {
			Content []STContent
			Usage   STUsage
		}
		var st ST
		err = json.Unmarshal(js, &st)
		if err != nil {
			job.AddError(fmt.Errorf("Unmarshal() of %s failed: %w", js, err))
			return "", err
		}
		if len(st.Content) > 0 {
			answer = st.Content[0].Text
		} else {
			job.AddError(fmt.Errorf("Missing answer in js: %s", string(js)))
			return "", err
		}

		dt := (float64(time.Now().UnixMilli()) / 1000) - startTime
		fmt.Printf("info: generated %dtoks which took %.1fsec = %.1f toks/sec\n", st.Usage.Output_tokens, dt, float64(st.Usage.Output_tokens)/dt)
	}

	if res.StatusCode != 200 {
		return "", fmt.Errorf("statusCode %d != 200, response: %s", res.StatusCode, answer)
	}
	return answer, nil

	/*body := bytes.NewReader(jsProps)

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

	var answer string
	if stream {
		answer, err = Anthropic_completion_parseStream(res, job)
		if err != nil {
			return "", err
		}
	} else {

		js, err := io.ReadAll(res.Body) //job.close ...
		if err != nil {
			job.AddError(fmt.Errorf("ReadAll() failed: %E", err))
			return "", err
		}

		type STMsg struct {
			Content string
		}
		type STChoice struct {
			Message STMsg
			Delta   STMsg
		}
		type Usage struct {
			Prompt_tokens     int
			Completion_tokens int
			Total_tokens      int
			//prompt_tokens_details ...
		}
		type ST struct {
			Choices []STChoice
			Usage   Usage
		}
		var st ST
		err = json.Unmarshal(js, &st)
		if err != nil {
			job.AddError(fmt.Errorf("Unmarshal() of %s failed: %w", js, err))
			return "", err
		}
		if len(st.Choices) > 0 {
			answer = st.Choices[0].Message.Content
		} else {
			job.AddError(fmt.Errorf("Missing answer in js: %s", string(js)))
			return "", err
		}

		dt := (float64(time.Now().UnixMilli()) / 1000) - startTime
		fmt.Printf("info: generated %dtoks which took %.1fsec = %.1f toks/sec\n", st.Usage.Completion_tokens, dt, float64(st.Usage.Completion_tokens)/dt)
	}

	if res.StatusCode != 200 {
		return "", fmt.Errorf("statusCode %d != 200, response: %s", res.StatusCode, answer)
	}

	return string(answer), nil*/
}

func Anthropic_completion_parseStream(res *http.Response, job *Job) (string, error) {

	type STDelta struct {
		Type string
		Text string
	}
	type ST struct {
		Type  string
		Delta STDelta
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

			{
				strData := strBlock[:d] //cut end
				{
					d := strings.Index(strData, "\n")
					if d < 0 {
						return "", fmt.Errorf("wrong syntax")
					}
					strData = strData[d+1:] //cut event: message_start
				}

				js, found := strings.CutPrefix(strData, "data:") //cut start from 2nd part
				if !found {
					return "", fmt.Errorf("missing 'data:'")
				}
				js = strings.TrimSpace(js)

				var st ST
				err := json.Unmarshal([]byte(js), &st)
				if err != nil {
					return "", fmt.Errorf("Unmarshal() failed: %w", err)
				}

				if st.Type == "content_block_delta" {
					job.info = job.info + st.Delta.Text
					fmt.Print(st.Delta.Text)
				}
			}
		}

		if readErr != nil && readErr == io.EOF {
			break //done
		}
	}

	if !job.IsRunning() {
		return "", fmt.Errorf("user Cancel the job")
	}

	return job.info, nil
}
