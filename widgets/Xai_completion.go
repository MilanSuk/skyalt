package main

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Xai_completion struct {
	UID        string
	Properties Xai_completion_props

	Out  string
	done func(Out string)
}

func (layout *Layout) AddXai_completion(x, y, w, h int, props *Xai_completion) *Xai_completion {
	layout._createDiv(x, y, w, h, "Xai_completion", props.Build, nil, nil)
	return props
}

var g_global_Xai_completion = make(map[string]*Xai_completion)

func NewGlobal_Xai_completion(uid string) *Xai_completion {
	uid = fmt.Sprintf("Xai_completion:%s", uid)

	st, found := g_global_Xai_completion[uid]
	if !found {
		st = &Xai_completion{UID: uid}
		st.Properties.Reset()

		g_global_Xai_completion[uid] = st
	}
	return st
}

func (st *Xai_completion) Build(layout *Layout) {

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

func (st *Xai_completion) Start() *Job {
	return StartJob(st.UID, "XAi chat completion", st.Run)
}
func (st *Xai_completion) Stop() {
	job := FindJob(st.UID)
	if job != nil {
		job.Stop()
	}
}
func (st *Xai_completion) IsRunning() bool {
	return FindJob(st.UID) != nil
}

func (st *Xai_completion) Run(job *Job) {
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
		st.done(st.Out)
	}
}
