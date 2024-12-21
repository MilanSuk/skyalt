package main

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Xai_chat struct {
	UID        string
	Properties Xai_chat_props

	Out  string
	done func()
}

func (layout *Layout) AddXai_chat(x, y, w, h int, props *Xai_chat) *Xai_chat {
	layout._createDiv(x, y, w, h, "Xai_chat", props.Build, nil, nil)
	return props
}

var g_global_Xai_chat = make(map[string]*Xai_chat)

func NewGlobal_Xai_chat(uid string) *Xai_chat {
	uid = fmt.Sprintf("Xai_chat:%s", uid)

	st, found := g_global_Xai_chat[uid]
	if !found {
		st = &Xai_chat{UID: uid}
		st.Properties.Reset()

		g_global_Xai_chat[uid] = st
	}
	return st
}

func (st *Xai_chat) Build(layout *Layout) {

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

func (st *Xai_chat) Start() *Job {
	return StartJob(st.UID, "XAi chat completion", st.Run)
}
func (st *Xai_chat) Stop() {
	job := FindJob(st.UID)
	if job != nil {
		job.Stop()
	}
}
func (st *Xai_chat) IsRunning() bool {
	return FindJob(st.UID) != nil
}

func (st *Xai_chat) Run(job *Job) {
	if !NewFile_Xai().Enable {
		job.AddError(errors.New("Xai is disabled"))
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

	st.Out, err = OpenAI_chat_Complete(jsProps, "https://api.x.ai/v1/chat/completions", NewFile_Xai().Api_key, job)
	if err != nil {
		job.AddError(err)
		return
	}

	if st.done != nil {
		st.done()
	}
}