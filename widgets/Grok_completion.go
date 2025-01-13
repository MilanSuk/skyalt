package main

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Groq_completion struct {
	UID        string
	Properties Groq_completion_props

	Out  string
	done func(Out string)
}

func (layout *Layout) AddGroq_completion(x, y, w, h int, props *Groq_completion) *Groq_completion {
	layout._createDiv(x, y, w, h, "Groq_completion", props.Build, nil, nil)
	return props
}

func OpenMemory_Groq_completion(uid string) *Groq_completion {
	st := &Groq_completion{UID: uid}
	st.Properties.Reset()

	return OpenMemory(uid, st)
}

func (st *Groq_completion) Build(layout *Layout) {

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

func (st *Groq_completion) Start() *Job {

	return StartJob(st.UID, "Groq chat completion", st.Run)
}
func (st *Groq_completion) Stop() {
	job := FindJob(st.UID)
	if job != nil {
		job.Stop()
	}
}
func (st *Groq_completion) FindJob() *Job {
	return FindJob(st.UID)
}

func (st *Groq_completion) Run(job *Job) {
	if !OpenFile_Groq().Enable {
		job.AddError(errors.New("Groq is disabled"))
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

	st.Out, err = OpenAI_completion_Run(jsProps, st.Properties.Stream, "https://api.groq.com/openai/v1/chat/completions", OpenFile_Groq().Api_key, job)
	if err != nil {
		job.AddError(err)
		return
	}

	if st.done != nil {
		st.done(st.Out)
	}
}
