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
	done func(out string)
}

func (layout *Layout) AddXai_completion(x, y, w, h int, props *Xai_completion) *Xai_completion {
	layout._createDiv(x, y, w, h, "Xai_completion", props.Build, nil, nil)
	return props
}

func OpenMemory_Xai_completion(uid string) *Xai_completion {
	st := &Xai_completion{UID: uid}
	st.Properties.Reset()
	return OpenMemory(uid, st)
}

func (st *Xai_completion) Build(layout *Layout) {

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

func (st *Xai_completion) Start() *Job {

	return StartJob(st.UID, "XAi chat completion", st.Run)
}
func (st *Xai_completion) Stop() {
	job := FindJob(st.UID)
	if job != nil {
		job.Stop()
	}
}
func (st *Xai_completion) FindJob() *Job {
	return FindJob(st.UID)
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
