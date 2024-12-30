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
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

type Whispercpp_stt struct {
	UID string

	Properties Whispercpp_props

	Input_Data       audio.IntBuffer
	Input_SampleRate int
	Input_Channels   int

	Out  string
	done func()
}

func (layout *Layout) AddWhispercpp_stt(x, y, w, h int, props *Whispercpp_stt) *Whispercpp_stt {
	layout._createDiv(x, y, w, h, "Whispercpp_stt", props.Build, nil, nil)
	return props
}

var g_global_Whispercpp_stt = make(map[string]*Whispercpp_stt)

func NewGlobal_Whispercpp_stt(uid string) *Whispercpp_stt {
	uid = fmt.Sprintf("Whispercpp_stt:%s", uid)

	st, found := g_global_Whispercpp_stt[uid]
	if !found {
		st = &Whispercpp_stt{UID: uid}
		st.Properties = *NewWhispercpp_props()
		st.Input_Channels = OpenFile_Microphone().Channels
		st.Input_SampleRate = OpenFile_Microphone().SampleRate

		g_global_Whispercpp_stt[uid] = st
	}
	return st
}

func (st *Whispercpp_stt) Build(layout *Layout) {

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

func (st *Whispercpp_stt) Start() *Job {
	return StartJob(st.UID, "Whispercpp(speech-to-text)", st.Run)
}
func (st *Whispercpp_stt) Stop() {
	job := FindJob(st.UID)
	if job != nil {
		job.Stop()
	}
}

func (st *Whispercpp_stt) IsRunning() bool {
	return FindJob(st.UID) != nil
}

func (st *Whispercpp_stt) Run(job *Job) {

	wsp := OpenFile_Whispercpp()

	//start
	st.Out = ""

	if st.Properties.Model == "" {
		job.AddError(errors.New("model is empty"))
		return
	}

	//convert stt.Input into buff(mp3)
	var buff []byte
	{
		os.MkdirAll("temp", os.ModePerm)
		path := "temp/mic.wav"

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

			buff, err = os.ReadFile(path)
			if err != nil {
				job.AddError(err)
				return
			}
		}
	}

	//set model
	err := g__WhispercppProcess.SetModel(st.Properties.Model, wsp)
	if err != nil {
		job.AddError(fmt.Errorf("SetModel() failed: %w", err))
		return
	}

	answer, err := g__WhispercppProcess.Transcribe(buff, &st.Properties, wsp)
	if err != nil {
		job.AddError(fmt.Errorf("Transcribe() failed: %w", err))
		return
	}

	st.Out = string(answer)

	if st.done != nil {
		st.done()
	}
}

type WhispercppProcess struct {
	selected_model string
	cmd            *exec.Cmd
	lock           sync.Mutex
}

var g__WhispercppProcess WhispercppProcess

func WhispercppProcess_getModelPath(model_name string) string {
	return filepath.Join("models/", model_name+".bin")
}

func (prc *WhispercppProcess) _check(wsp *Whispercpp) error {

	if prc.cmd == nil {

		//run process
		if wsp.RunProcess {
			prc.cmd = exec.Command("./server", "--port", strconv.Itoa(wsp.Port), "--convert", "-m", WhispercppProcess_getModelPath(wsp.Model))
			prc.cmd.Dir = wsp.Folder

			prc.cmd.Stdout = os.Stdout
			prc.cmd.Stderr = os.Stderr
			err := prc.cmd.Start()
			if err != nil {
				prc.cmd = nil
				return fmt.Errorf("Command() failed: %w", err)
			}
		}

		//wait until it's running
		{
			err := errors.New("err")
			st := time.Now().Unix()
			for err != nil && (time.Now().Unix()-st) < 10 { //max 10sec to start
				err = prc._setModel(wsp.Model, wsp)
				time.Sleep(100 * time.Millisecond)
			}
			if err != nil {
				fmt.Println(err)
				prc.cmd.Process.Kill()
				prc.cmd = nil
				return err
			}
		}
	}

	AddProcess("Whispercpp", 60, prc.Destroy)
	return nil
}

func (prc *WhispercppProcess) Destroy() {
	prc.lock.Lock()
	defer prc.lock.Unlock()

	err := prc.cmd.Process.Kill()
	if err != nil {
		//stt.service.AddError(err)
	}
}

func (prc *WhispercppProcess) _setModel(model string, wsp *Whispercpp) error {
	if model == prc.selected_model {
		return nil //already set
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("model", WhispercppProcess_getModelPath(model))
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, wsp.GetUrlLoadModel(), body)
	if err != nil {
		return fmt.Errorf("NewRequest() failed: %w", err)
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Do() failed: %w", err)
	}
	defer res.Body.Close()

	answer, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("ReadAll() failed: %w", err)
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("statusCode %d != 200, response: %s", res.StatusCode, answer)
	}

	prc.selected_model = model
	return nil
}

func (prc *WhispercppProcess) SetModel(model string, wsp *Whispercpp) error {
	prc.lock.Lock()
	defer prc.lock.Unlock()
	err := prc._check(wsp)
	if err != nil {
		return err
	}
	return prc._setModel(model, wsp)
}

func (prc *WhispercppProcess) Transcribe(intput []byte, props *Whispercpp_props, wsp *Whispercpp) ([]byte, error) {
	prc.lock.Lock()
	defer prc.lock.Unlock()
	err := prc._check(wsp)
	if err != nil {
		return nil, err
	}

	//set parameters
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	{
		part, err := writer.CreateFormFile("file", "blob.wav")
		if err != nil {
			return nil, fmt.Errorf("CreateFormFile() failed: %w", err)
		}
		part.Write(intput)
		props.Write(writer)
	}
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, wsp.GetUrlInference(), body)
	if err != nil {
		return nil, fmt.Errorf("NewRequest() failed: %w", err)
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Do() failed: %w", err)
	}

	answer, err := io.ReadAll(res.Body) //job.close ...
	if err != nil {
		return nil, fmt.Errorf("ReadAll() failed: %w", err)
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("statusCode %d != 200, response: %s", res.StatusCode, answer)
	}

	return answer, nil
}
