package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type Llamacpp_completion struct {
	UID        string
	Properties Llamacpp_completion_props

	Out  string
	done func(Out string)
}

func (layout *Layout) AddLlamacpp_completion(x, y, w, h int, props *Llamacpp_completion) *Llamacpp_completion {
	layout._createDiv(x, y, w, h, "Llamacpp_completion", props.Build, nil, nil)
	return props
}

var g_global_Llamacpp_completion = make(map[string]*Llamacpp_completion)

func NewGlobal_Llamacpp_completion(uid string) *Llamacpp_completion {
	uid = fmt.Sprintf("Llamacpp_completion:%s", uid)

	st, found := g_global_Llamacpp_completion[uid]
	if !found {
		st = &Llamacpp_completion{UID: uid}
		st.Properties.Reset()
		g_global_Llamacpp_completion[uid] = st
	}
	return st
}

func (st *Llamacpp_completion) Build(layout *Layout) {

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

func (st *Llamacpp_completion) Start() *Job {

	return StartJob(st.UID, "Llamacpp chat completion", st.Run)
}
func (st *Llamacpp_completion) Stop() {
	job := FindJob(st.UID)
	if job != nil {
		job.Stop()
	}
}
func (st *Llamacpp_completion) FindJob() *Job {
	return FindJob(st.UID)
}

func (st *Llamacpp_completion) Run(job *Job) {

	llm := OpenFile_Llamacpp()

	//start
	st.Out = ""

	if st.Properties.Model == "" {
		job.AddError(errors.New("model is empty"))
		return
	}

	jsProps, err := json.Marshal(st.Properties)
	if err != nil {
		Layout_WriteError(fmt.Errorf("Marshal() failed: %w", err))
		return
	}

	answer, err := g_LlamacppProcess.Complete(jsProps, st.Properties.Stream, llm, st.Properties.Model, job)
	if err != nil {
		Layout_WriteError(fmt.Errorf("Complete() failed: %w", err))
		return
	}

	st.Out = string(answer)

	if st.done != nil {
		st.done(st.Out)
	}
}

type LlamacppProcess struct {
	selected_model string
	cmd            *exec.Cmd
	lock           sync.Mutex
}

var g_LlamacppProcess LlamacppProcess

func LlamacppProcess_getModelPath(model_name string) string {
	return filepath.Join("models/", model_name+".gguf")
}

func (prc *LlamacppProcess) _check(llm *Llamacpp, init_model string) error {

	if prc.cmd != nil && prc.selected_model != init_model {
		//reset
		prc.Destroy()
	}

	if prc.cmd == nil {
		prc.selected_model = init_model

		//run process
		if llm.RunProcess {
			prc.cmd = exec.Command("./llama-server", "--port", strconv.Itoa(llm.Port), "-m", LlamacppProcess_getModelPath(init_model))
			prc.cmd.Dir = llm.Folder

			prc.cmd.Stdout = os.Stdout
			prc.cmd.Stderr = os.Stderr
			err := prc.cmd.Start()
			if err != nil {
				return fmt.Errorf("Command() failed: %w", err)
			}
		}

		//wait until it's running
		{
			err := errors.New("err")
			st := time.Now().Unix()
			for err != nil && (time.Now().Unix()-st) < 120 { //max 120sec to start
				err = prc._getHealth()
				time.Sleep(250 * time.Millisecond)
			}
			if err != nil {
				return err
			}
		}
	}

	AddProcess("Llamacpp", 60, prc.Destroy)
	return nil
}

func (prc *LlamacppProcess) Destroy() {
	prc.lock.Lock()
	defer prc.lock.Unlock()

	err := prc.cmd.Process.Kill()
	if err != nil {
		//stt.service.AddError(err)
	}
	prc.cmd = nil
}

func (prc *LlamacppProcess) _getHealth() error {
	//prc.lock.Lock()
	//defer prc.lock.Unlock()

	res, err := http.Get(OpenFile_Llamacpp().GetUrlHealth())
	if err != nil {
		return fmt.Errorf("Get() failed: %w", err)
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("ReadAll() failed: %w", err)
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("statusCode: %d, response: %s", res.StatusCode, resBody)
	}

	return nil
}

/*func (prc *LlamacppProcess) GetHealth(llm *Llamacpp) error {
	prc.lock.Lock()
	defer prc.lock.Unlock()
	err := prc._check(llm, )
	if err != nil {
		return err
	}
	return prc._getHealth()
}*/

func (prc *LlamacppProcess) Complete(jsProps []byte, stream bool, llm *Llamacpp, init_model string, job *Job) (string, error) {
	prc.lock.Lock()
	defer prc.lock.Unlock()
	err := prc._check(llm, init_model)
	if err != nil {
		return "", err
	}

	fmt.Println("jsProps", string(jsProps))

	startTime := float64(time.Now().UnixMilli()) / 1000

	body := bytes.NewReader(jsProps)

	req, err := http.NewRequest(http.MethodPost, OpenFile_Llamacpp().GetUrCompletion(), body)
	if err != nil {
		return "", fmt.Errorf("NewRequest() failed: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
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
