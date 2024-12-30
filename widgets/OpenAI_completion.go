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

type OpenAI_completion struct {
	UID        string
	Properties OpenAI_completion_props

	Out  string
	done func(Out string)
}

func (layout *Layout) AddOpenAI_completion(x, y, w, h int, props *OpenAI_completion) *OpenAI_completion {
	layout._createDiv(x, y, w, h, "OpenAI_completion", props.Build, nil, nil)
	return props
}

var g_global_OpenAI_completion = make(map[string]*OpenAI_completion)

func NewGlobal_OpenAI_completion(uid string) *OpenAI_completion {
	uid = fmt.Sprintf("OpenAI_completion:%s", uid)

	st, found := g_global_OpenAI_completion[uid]
	if !found {
		st = &OpenAI_completion{UID: uid}
		st.Properties.Reset()
		g_global_OpenAI_completion[uid] = st
	}
	return st
}

func (st *OpenAI_completion) Build(layout *Layout) {

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

func (st *OpenAI_completion) Start() *Job {
	return StartJob(st.UID, "OpenAI chat completion", st.Run)
}
func (st *OpenAI_completion) Stop() {
	job := FindJob(st.UID)
	if job != nil {
		job.Stop()
	}
}
func (st *OpenAI_completion) IsRunning() bool {
	return FindJob(st.UID) != nil
}

func (st *OpenAI_completion) Run(job *Job) {

	if !OpenFile_OpenAI().Enable {
		job.AddError(errors.New("OpenAI is disabled"))
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

	st.Out, err = OpenAI_completion_Run(jsProps, st.Properties.Stream, "https://api.openai.com/v1/chat/completions", OpenFile_OpenAI().Api_key, job)
	if err != nil {
		fmt.Println("--OpenAI_completion_Run error", err)
		job.AddError(err)
		return
	}

	fmt.Println("--OpenAI_completion_Run done")

	if st.done != nil {
		st.done(st.Out)
	}
}

func OpenAI_completion_Run(jsProps []byte, stream bool, Completion_url string, Api_key string, job *Job) (string, error) {
	fmt.Println("jsProps", string(jsProps))

	startTime := float64(time.Now().UnixMilli()) / 1000

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

	var answer string
	if stream {
		answer, err = OpenAI_completion_parseStream(res, job)
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
		type STUsage struct {
			Prompt_tokens     int
			Completion_tokens int
			Total_tokens      int
			//prompt_tokens_details ...
		}
		type ST struct {
			Choices []STChoice
			Usage   STUsage
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

	return answer, nil
}

func OpenAI_completion_parseStream(res *http.Response, job *Job) (string, error) {
	// https://platform.openai.com/docs/api-reference/assistants-streaming/events
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
			fmt.Println(string(tb[:n])) //bug somewhere around: sometime never return [done] .........
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
			if js == "[DONE]" {
				fmt.Println("---done")
				readErr = io.EOF
				break
			} else {
				var st ST
				err := json.Unmarshal([]byte(js), &st)
				if err != nil {
					return "", fmt.Errorf("Unmarshal() failed: %w", err)
				}

				if len(st.Choices) > 0 {
					job.info = job.info + st.Choices[0].Delta.Content
					//fmt.Print(st.Choices[0].Delta.Content)
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
