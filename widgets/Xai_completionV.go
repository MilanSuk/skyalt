package main

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Xai_completionV struct {
	UID        string
	Properties Xai_completionV_props

	Out  string
	done func()
}

func (layout *Layout) AddXai_completionV(x, y, w, h int, props *Xai_completionV) *Xai_completionV {
	layout._createDiv(x, y, w, h, "Xai_completionV", props.Build, nil, nil)
	return props
}

var g_global_Xai_completionV = make(map[string]*Xai_completionV)

func NewGlobal_Xai_completionV(uid string) *Xai_completionV {
	uid = fmt.Sprintf("Xai_completionV:%s", uid)

	st, found := g_global_Xai_completionV[uid]
	if !found {
		st = &Xai_completionV{UID: uid}
		st.Properties.Reset()

		g_global_Xai_completionV[uid] = st
	}
	return st
}

func (st *Xai_completionV) Build(layout *Layout) {

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

func (st *Xai_completionV) Start() *Job {

	return StartJob(st.UID, "XAi chat completion", st.Run)
}
func (st *Xai_completionV) Stop() {
	job := FindJob(st.UID)
	if job != nil {
		job.Stop()
	}
}
func (st *Xai_completionV) FindJob() *Job {
	return FindJob(st.UID)
}

func (st *Xai_completionV) Run(job *Job) {
	if !OpenFile_Xai().Enable {
		job.AddError(errors.New("Xai is disabled"))
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

	st.Out, err = OpenAI_completion_Run(jsProps, st.Properties.Stream, "https://api.x.ai/v1/chat/completions", OpenFile_Xai().Api_key, job)
	if err != nil {
		job.AddError(err)
		return
	}

	if st.done != nil {
		st.done()
	}
}
