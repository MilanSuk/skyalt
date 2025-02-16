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

		//Messages
		y := 0
		for i := range st.agent.Messages {
			if len(st.agent.Messages[i].Content) > 0 && st.agent.Messages[i].Content[0].Type == "tool_result" {
				continue //skip
			}

			MsgsDiv.SetRowFromSub(y, 1, 100)
			MsgsDiv.AddChatMsg(x, y, 1, 1, &ChatMsg{agent: st.agent, msg_i: i})
			y++

			//space
			MsgsDiv.SetRow(y, 0.5, 0.5)
			y++
		}

		//Statistics - total
		if y >= 2 { //1st message is user
			in, inCached, out := st.agent.GetTotalChatPrice()
			info := MsgsDiv.AddText(x, y, 1, 1, fmt.Sprintf("$%s, %d tokens, %s sec, %d tokens/sec",
				strconv.FormatFloat(in+inCached+out, 'f', 3, 64),
				st.agent.GetTotalOutputTokens(),
				strconv.FormatFloat(st.agent.GetTotalTime(), 'f', 3, 64),
				int(st.agent.GetTotalSpeed())))
			y++
			info.Align_h = 2 //right
			info.Tooltip = fmt.Sprintf("%s sec\n%s tokens/sec\nTotal: $%s\n- Input: $%s\n- Input cached: $%s\n- Output: $%s",
				strconv.FormatFloat(st.agent.GetTotalTime(), 'f', -1, 64),
				strconv.FormatFloat(st.agent.GetTotalSpeed(), 'f', -1, 64),
				strconv.FormatFloat(in+inCached+out, 'f', -1, 64),
				strconv.FormatFloat(in, 'f', -1, 64),
				strconv.FormatFloat(inCached, 'f', -1, 64),
				strconv.FormatFloat(out, 'f', -1, 64))
		}

		//space
		MsgsDiv.SetRow(y, 0.5, 0.5)
		y++

		//Stop
		if st.agent.Call_id == "" {
			stopped := st.agent.Stopped
			if stopped != "" {
				MsgsDiv.AddText(x, y, 1, 1, "("+stopped+")").Align_h = 1
				y++

				{
					ContDiv := MsgsDiv.AddLayout(x, y, 1, 1)
					ContDiv.SetColumn(0, 1, 100)
					ContDiv.SetColumn(1, 1, 7)
					ContDiv.SetColumn(2, 1, 100)
					ContinueBt := ContDiv.AddButton(1, 0, 1, 1, "Continue ...")
					ContinueBt.Background = 0.5
					y++
					ContinueBt.clicked = func() {
						st.agent.RemoveUnfinishedMsg()
						st.Start()
					}
				}

				//space
				MsgsDiv.SetRow(y, 0.5, 0.5)
				y++
			}
		}

		for i := len(st.agent.SubAgents) - 1; i >= 0; i-- {
			dashAgent := st.agent.SubAgents[i]
			if dashAgent.HashUiCalls() {

				MsgsDiv.SetRowFromSub(y, 1, 100)
				DashDiv := MsgsDiv.AddLayout(x, y, 1, 1)
				y++
				DashDiv.Border_cd = Paint_GetPalette().P //Paint_GetPalette().GetGrey(0.2)
				DashDiv.SetColumn(0, 0.5, 0.5)
				DashDiv.SetColumn(1, 1, 100)
				DashDiv.SetColumn(2, 0.5, 0.5)
				DashDiv.SetRow(0, 0.5, 0.5)
				DashDiv.SetRow(1, 1, 100)
				DashDiv.SetRow(2, 0.5, 0.5)

				//Dashboard
				DashDiv.SetRowFromSub(1, 1, 100)
				dash := DashDiv.AddChatDashboard(1, 1, 1, 1, &ChatDashboard{agent: dashAgent})
				y++
				dash.isRunning = func() bool {
					return st.Find() != nil
				}
				dash.uiChanged = func() {
					if st.Find() == nil {
						st.StartDash(dashAgent)
					}
				}

				break
			}
		}
	}

	if st.agent.Call_id == "" {
		//Input
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
		st.agent.Input.file_name = st.file_name
		Input := InputDiv.AddChatInput(x, 0, 1, 1, &st.agent.Input)
		Input.isRunning = func() bool {
			return st.Find() != nil
		}
		Input.stop = func() {
			st.Stop()
		}
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
		//Agent info

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

func (st *Chat) getUID() string {
	return "chat_" + st.file_name
}

func (st *Chat) Find() *Job {
	return FindJob(st.getUID())
}
func (st *Chat) Start() *Job {

	if len(st.agent.Messages) == 1 {
		summarize := func(job *Job) {
			//prepare
			ag := NewAgent("", "", "You are an AI assistant. You summarize long text to clear short description.")
			ag.NoTools = true
			ag.AddUserPromptText("Summarize this text to maximum 10 words:\n"+st.agent.GetFirstMessage(), "")

			//run
			ag.ExeLoop(1, 100, job)

			//save
			st.agent.Description = ag.GetFinalMessage()
		}
		StartJob("summarize_"+st.getUID(), fmt.Sprintf("Summary %s", st.file_name), summarize)
	}

	run := func(job *Job) {
		st.agent.ExeLoop(0, 0, job)
		st.agent.Input.reset()
	}
	return StartJob(st.getUID(), fmt.Sprintf("Agent %s", st.file_name), run)
}

func (st *Chat) StartDash(dashAgent *Agent) *Job {
	run := func(job *Job) {
		dashAgent.ExeLoop(0, 0, job)

		//save previous chat ......

		st.agent.Messages = st.agent.Messages[:1]
		st.agent.SubAgents = nil

		//re-run whole chat
		st.agent.ExeLoop(0, 0, job)
		st.agent.Input.reset()
	}

	return StartJob(st.getUID(), fmt.Sprintf("Dash %s", st.file_name), run)
}

func (st *Chat) Stop() {
	job := FindJob(st.getUID())
	if job != nil {
		job.Stop()
	}
}
