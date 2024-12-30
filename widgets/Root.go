/*
Copyright 2024 Milan Suk

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this db except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"time"
)

type Root struct {
	ShowPromptList       bool
	Last_error_time_unix int64
}

func (layout *Layout) AddRoot(x, y, w, h int, props *Root) *Root {
	layout._createDiv(x, y, w, h, "Root", props.Build, nil, nil)
	return props
}

var g_Root *Root

func NewFile_Root() *Root {
	if g_Root == nil {
		g_Root = &Root{}
		_read_file("Root-Root", g_Root)
	}
	return g_Root
}

func (st *Root) Build(layout *Layout) {

	layout.SetColumn(0, 1, 100)

	layout.SetRowFromSub(0)
	layout.SetRow(1, 1, 100)

	// Prompt
	{
		HeaderDiv := layout.AddLayout(0, 0, 1, 1)
		HeaderDiv.SetColumn(0, 1, 1.5) //logo
		HeaderDiv.SetColumn(1, 1, 100)
		HeaderDiv.SetColumnFromSub(2)  //mic
		HeaderDiv.SetColumn(3, 1, 15)  //prompt
		HeaderDiv.SetColumn(4, 1, 1)   //clear
		HeaderDiv.SetColumn(5, 1, 3)   //send
		HeaderDiv.SetColumn(6, 1, 100) //empty/errors
		HeaderDiv.SetColumn(7, 1, 1)   //jobs
		HeaderDiv.SetRowFromSub(0)
		HeaderDiv.Back_cd = Paint_GetPalette().GetGrey(0.9)

		ast := NewFile_AssistantChat()

		logoBt := HeaderDiv.AddButton(0, 0, 1, 1, NewButtonIcon("resources/logo_small.png", 0.1, "v0.1")) //v0.1 .......
		if !st.ShowPromptList {
			logoBt.Background = 0
		}
		logoBt.clicked = func() {
			st.ShowPromptList = !st.ShowPromptList
		}

		//microphone
		av := HeaderDiv.AddAssistantVoice(2, 0, 1, 1)
		av.Shortcut = '\t'
		av.Button_background = 0.25

		//prompt
		//...ChatDia := ast.CreateDialog(layout)
		ast.Assistant_recomputePromptColors() //, ast.Picks)
		ed, edLay := HeaderDiv.AddEditboxMultiline(3, 0, 1, 1, &ast.Prompt)
		ed.Ghost = "What can I do for you?"
		ed.Tooltip = "Use Ctrl + Mouse to select widgets."
		edLay.Back_cd = Paint_GetPalette().B
		if ast.Prompt == "" {
			ast.reset()
		}

		//clear
		clearBt := HeaderDiv.AddButton(4, 0, 1, 1, NewButton("X"))
		clearBt.Background = 0.5
		//sendLay.Enable = len(ast.Prompt) > 0
		clearBt.clicked = func() {
			ast.reset()
		}

		//send
		sendBt := HeaderDiv.AddButton(5, 0, 1, 1, NewButton("Send"))
		//sendLay.Enable = len(ast.Prompt) > 0
		sendBt.clicked = func() {
			ast.Send()
		}

		//error info
		{
			logs := NewFile_Logs()

			ErrDia := HeaderDiv.AddDialog("errors")
			ErrDia.Layout.SetColumn(0, 1, 15)
			ErrDia.Layout.SetRow(0, 1, 15)
			ErrDia.Layout.AddLogs(0, 0, 1, 1, logs)

			err := logs.GetError(st.Last_error_time_unix)
			if err != nil {
				HeaderDiv.AddLayout(1, 0, 1, 1) //empty layout, so prompt(editbox) stays centered
				errDiv := HeaderDiv.AddLayout(6, 0, 1, 1)
				errDiv.SetColumn(0, 1, 1)
				errDiv.SetColumn(2, 1, 100) //text
				errDiv.SetColumn(3, 1, 1)

				openBt := errDiv.AddButton(2, 0, 1, 1, NewButtonMenu("Error: "+err.Error, "", 0))
				openBt.Tooltip = "Open Errors App"
				openBt.Cd = Paint_GetPalette().E
				openBt.Background = 0.2
				openBt.clicked = func() {
					ErrDia.OpenCentered()
					st.Last_error_time_unix = err.Time_unix + 1 //hide
				}

				now_unix := time.Now().Unix()
				if now_unix-5 > err.Time_unix {
					st.Last_error_time_unix = now_unix
				}
			}
		}

		//jobs
		{
			JobsDia := HeaderDiv.AddDialog("Jobs")
			JobsDia.Layout.SetColumn(0, 5, 10)
			JobsDia.Layout.SetRow(0, 10, 10)
			JobsDia.Layout.AddJobs(0, 0, 1, 1)

			enable := len(g__jobs.jobs) > 0
			sz := 0.2
			if enable {
				sz = 0.1 //bigger icon
			}
			JobsBt, JobsL := HeaderDiv.AddButton2(7, 0, 1, 1, NewButtonIcon("resources/logo_Counter.png", sz, "List of jobs"))
			if enable {
				JobsBt.Cd = Paint_GetPalette().P
			}
			JobsL.Enable = enable
			JobsBt.Background = 0.25
			JobsBt.clicked = func() {
				JobsDia.OpenRelative(JobsL)
			}
		}

		//panel switch
		/*PanelBt := HeaderDiv.AddButton(7, 0, 1, 1, NewButtonIcon("resources/settings.png", 0.2, "Open/Close AI Assistant"))
		PanelBt.Background = 0.25
		if ast.Show {
			PanelBt.Background = 1
		}
		PanelBt.clicked = func() {
			ast.Show = !ast.Show
		}*/

		//udpate skyalt dom paints
		//layout.CmdSetPicks(ast.Picks)
	}

	{
		AppDiv := layout.AddLayout(0, 1, 1, 1)
		AppDiv.App = true
		AppDiv.SetColumn(0, 1, 100)
		AppDiv.SetRow(0, 1, 100)
		if st.ShowPromptList {
			AppDiv.AddPrompts(0, 0, 1, 1)
		} else {
			//App
			AppDiv.AddShowApp(0, 0, 1, 1)
		}
	}

	//Assistant panel
	/*if NewFile_Assistant().Show {
		layout.SetColumnResizable(1, 5, 20, 6)
		layout.AddAssistant(1, 0, 1, 2, NewFile_Assistant())
	}*/
}
