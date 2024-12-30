package main

import (
	"encoding/json"
	"errors"
	"fmt"
)

type OpenAI_completionV struct {
	UID        string
	Properties OpenAI_completionV_props

	Out  string
	done func()
}

func (layout *Layout) AddOpenAI_completionV(x, y, w, h int, props *OpenAI_completionV) *OpenAI_completionV {
	layout._createDiv(x, y, w, h, "OpenAI_completionV", props.Build, nil, nil)
	return props
}

var g_global_OpenAI_completionV = make(map[string]*OpenAI_completionV)

func NewGlobal_OpenAI_completionV(uid string) *OpenAI_completionV {
	uid = fmt.Sprintf("OpenAI_completionV:%s", uid)

	st, found := g_global_OpenAI_completionV[uid]
	if !found {
		st = &OpenAI_completionV{UID: uid}
		st.Properties.Reset()
		g_global_OpenAI_completionV[uid] = st
	}
	return st
}

func (st *OpenAI_completionV) Build(layout *Layout) {

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

func (st *OpenAI_completionV) Start() *Job {
	return StartJob(st.UID, "OpenAI chat completion", st.Run)
}
func (st *OpenAI_completionV) Stop() {
	job := FindJob(st.UID)
	if job != nil {
		job.Stop()
	}
}
func (st *OpenAI_completionV) IsRunning() bool {
	return FindJob(st.UID) != nil
}

func (st *OpenAI_completionV) Run(job *Job) {

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
		job.AddError(err)
		return
	}

	if st.done != nil {
		st.done()
	}
}
