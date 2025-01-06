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
	"fmt"
	"time"
)

type RootHeader struct {
	ShowPromptList       bool
	Last_error_time_unix int64
}

func (layout *Layout) AddRootHeader(x, y, w, h int, props *RootHeader) *RootHeader {
	layout._createDiv(x, y, w, h, "RootHeader", props.Build, nil, nil)
	return props
}

var g_RootHeader *RootHeader

func OpenFile_RootHeader() *RootHeader {
	if g_RootHeader == nil {
		g_RootHeader = &RootHeader{}
		_read_file("RootHeader-RootHeader", g_RootHeader)
	}
	return g_RootHeader
}

func (st *RootHeader) Build(layout *Layout) {

	layout.SetColumn(0, 1, 1.5) //logo
	layout.SetColumn(1, 1, 100)
	layout.SetColumnFromSub(2)  //mic
	layout.SetColumn(3, 1, 20)  //prompt
	layout.SetColumn(4, 1, 1)   //clear
	layout.SetColumn(5, 1, 3)   //send
	layout.SetColumn(6, 1, 1)   //settings
	layout.SetColumn(7, 1, 100) //empty/errors
	layout.SetColumn(8, 1, 1)   //jobs
	layout.SetRowFromSub(0)
	layout.Back_cd = Paint_GetPalette().GetGrey(0.9)

	ast := OpenFile_AssistantChat()

	logoBt := layout.AddButtonIcon(0, 0, 1, 1, "resources/logo_small.png", 0.1, "v0.1") //v0.1 .......
	if !st.ShowPromptList {
		logoBt.Background = 0
	}
	logoBt.clicked = func() {
		st.ShowPromptList = !st.ShowPromptList
	}

	//microphone
	av := layout.AddAssistantVoice(2, 0, 1, 1)
	av.Shortcut = '\t'
	av.Button_background = 0.25

	//prompt
	ast.Assistant_recomputePromptColors() //, ast.Picks)
	ed, edLay := layout.AddEditboxMultiline(3, 0, 1, 1, &ast.Prompt)
	ed.Ghost = "What can I do for you?"
	ed.Tooltip = "Use Ctrl + Mouse to paint over."
	edLay.Back_cd = Paint_GetPalette().B
	ed.changed = func() {
		if ast.Prompt == "" {
			ast.reset()
		}
	}

	//clear
	clearBt := layout.AddButton(4, 0, 1, 1, "X")
	clearBt.Background = 0.2
	//sendLay.Enable = len(ast.Prompt) > 0
	clearBt.clicked = func() {
		ast.reset()
	}

	//send
	sendBt := layout.AddButton(5, 0, 1, 1, "Send")
	//sendLay.Enable = len(ast.Prompt) > 0
	sendBt.clicked = func() {
		ast.Send()
	}

	//settings
	{
		SDia := layout.AddDialog("settings")
		SDia.Layout.SetColumn(0, 1, 5)
		SDia.Layout.SetRowFromSub(0)
		SDia.Layout.AddModels(0, 0, 1, 1, &ast.Model)

		SettingsBt, SettingsLay := layout.AddButtonIcon2(6, 0, 1, 1, "resources/settings.png", 0.2, "Pick the model")
		SettingsBt.Background = 0.5
		SettingsBt.clicked = func() {
			SDia.OpenRelative(SettingsLay)
		}
	}

	//error info
	{
		logs := OpenFile_Logs()

		ErrDia := layout.AddDialog("errors")
		ErrDia.Layout.SetColumn(0, 1, 15)
		ErrDia.Layout.SetRow(0, 1, 15)
		ErrDia.Layout.AddLogs(0, 0, 1, 1, logs)

		err := logs.GetError(st.Last_error_time_unix)
		if err != nil {
			layout.AddLayout(1, 0, 1, 1) //empty layout, so prompt(editbox) stays centered
			errDiv := layout.AddLayout(7, 0, 1, 1)
			errDiv.SetColumn(0, 1, 1)
			errDiv.SetColumn(2, 1, 100) //text
			errDiv.SetColumn(3, 1, 1)

			openBt := errDiv.AddButtonMenu(2, 0, 1, 1, "Error: "+err.Error, "", 0)
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
		JobsDia := layout.AddDialog("Jobs")
		JobsDia.Layout.SetColumn(0, 5, 10)
		JobsDia.Layout.SetRow(0, 10, 10)
		JobsDia.Layout.AddJobs(0, 0, 1, 1)

		enable := len(g__jobs.jobs) > 0
		doneStr := ""
		slowJob := g__jobs.GetSlowestEstimateJob()
		if slowJob != nil {
			doneStr = fmt.Sprintf("%.1f%%", slowJob.GetEstimateDone()*100) //slowJob.GetEndTime()
		}
		//doneStr ...
		JobsBt, JobsL := layout.AddButton2(8, 0, 1, 1, doneStr)
		JobsL.Enable = enable
		JobsBt.Background = 0.25
		JobsBt.Tooltip = "List of jobs"
		JobsBt.Icon = "resources/logo_Counter.png"
		JobsBt.Icon_margin = 0.2
		if enable {
			JobsBt.Background = 1
			//JobsBt.Cd = Paint_GetPalette().P
			JobsBt.Icon_margin = 0.1 //bigger
		}
		if doneStr == "" {
			JobsBt.Icon_align = 1
		}

		JobsBt.clicked = func() {
			JobsDia.OpenRelative(JobsL)
		}

		if doneStr != "" {
			layout.SetColumn(8, 1, 3)
		}
	}

	//panel switch
	/*PanelBt := layout.AddButtonIcon(7, 0, 1, 1, "resources/settings.png", 0.2, "Open/Close AI Assistant")
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
