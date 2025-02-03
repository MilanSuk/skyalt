package main

import (
	"fmt"
	"strconv"
)

type Chat struct {
	file_name string
	center    bool

	parent_agent *Agent
	agent        *Agent
}

func (layout *Layout) AddChat(x, y, w, h int, props *Chat) *Chat {
	layout._createDiv(x, y, w, h, "Chat", props.Build, nil, nil)
	return props
}

func (st *Chat) Build(layout *Layout) {
	job := st.Find()

	layout.SetColumn(0, 1, 100)
	layout.SetRow(0, 0, 100)
	layout.SetRowFromSub(1, 1, 100)

	MsgsDiv := layout.AddLayout(0, 0, 1, 1)
	{
		MsgsDiv.SetColumn(0, 1, 100)

		x := 0
		if st.center {
			MsgsDiv.SetColumn(1, 1, 25)
			MsgsDiv.SetColumn(2, 1, 100)
			x = 1
		}

		//maybe add Button "Show System prompt" ....

		y := 0
		for i := range st.agent.Messages {

			if len(st.agent.Messages[i].Content) > 0 && st.agent.Messages[i].Content[0].Type == "tool_result" {
				continue //skip
			}

			//previous message
			MsgsDiv.SetRowFromSub(y, 1, 100)
			MsgsDiv.AddChatMsg(x, y, 1, 1, &ChatMsg{agent: st.agent, msg_i: i})
			y++

			//space
			MsgsDiv.SetRow(y, 0.5, 0.5)
			y++
		}

		//total info
		if y >= 2 { //1st message is user
			in, inCached, out := st.agent.GetTotalPrice()
			info := MsgsDiv.AddText(x, y, 1, 1, fmt.Sprintf("$%s, %d tokens/sec",
				strconv.FormatFloat(in+inCached+out, 'f', 3, 64),
				int(st.agent.GetTotalSpeed())))
			info.Align_h = 2 //right
			info.Tooltip = fmt.Sprintf("%s tokens/sec\nTotal: $%s\n- Input: $%s\n- Input cached: $%s\n- Output: $%s",
				strconv.FormatFloat(st.agent.GetTotalSpeed(), 'f', -1, 64),
				strconv.FormatFloat(in+inCached+out, 'f', -1, 64),
				strconv.FormatFloat(in, 'f', -1, 64),
				strconv.FormatFloat(inCached, 'f', -1, 64),
				strconv.FormatFloat(out, 'f', -1, 64))
		}
	}

	//stop button ......

	if st.agent.Call_id == "" {
		InputDiv := layout.AddLayout(0, 1, 1, 1)
		InputDiv.SetColumn(0, 1, 100)
		if st.center {
			InputDiv.SetColumn(1, 1, 25)
			InputDiv.SetColumn(2, 1, 100)
		}
		InputDiv.SetRowFromSub(0, 1, 100)

		x := 0
		if st.center {
			x = 1
		}
		Input := InputDiv.AddChatInput(x, 0, 1, 1, &st.agent.Input)
		InputDiv.FindLayout(x, 0, 1, 1).Enable = (job == nil)

		Input.sended = func() {
			st.agent.AddUserPromptTextAndImages(st.agent.Input.Text, st.agent.Input.Files)

			job := st.Find()
			if job != nil {
				return
			}
			job = st.Start()

			InputDiv.VScrollToTheBottom()
		}
	} else {

		headDiv := layout.AddLayout(0, 1, 1, 2)
		headDiv.SetColumn(0, 1, 100)

		msg_i, content_i := st.parent_agent.FindSubCallUseContent(st.agent.Call_id)
		if content_i >= 0 {
			ct := &st.parent_agent.Messages[msg_i].Content[content_i]
			tx := headDiv.AddText(0, 0, 1, 1, fmt.Sprintf("<i>%s(%s)", ct.Name, ct.Id))
			tx.Align_h = 1
		}

		closeBt := headDiv.AddButton(0, 1, 1, 1, "Close")
		closeBt.Background = 0.5
		closeBt.clicked = func() {
			//close subs
			agent := st.agent
			for agent != nil {
				next_agent := agent.FindSubAgent(agent.Selected_sub_call_id)
				agent.Selected_sub_call_id = ""
				agent = next_agent
			}

			//close in parent
			st.parent_agent.Selected_sub_call_id = ""
		}
	}
}

func (st *Chat) Run(job *Job) {
	st.agent.ExeLoop(20, 20000)

	st.agent.Input.reset()
}

func (st *Chat) getUID() string {
	return "chat_" + st.file_name
}

func (st *Chat) Find() *Job {
	return FindJob(st.getUID())
}
func (st *Chat) Start() *Job {
	return StartJob(st.getUID(), fmt.Sprintf("Agent %s", st.file_name), st.Run)
}
func (st *Chat) Stop() {
	job := FindJob(st.getUID())
	if job != nil {
		job.Stop()
	}
}
