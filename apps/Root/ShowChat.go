package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

// [ignore]
type ShowChat struct {
	AppName      string
	ChatFileName string
}

const g_ShowChat_prompt_height = 7

func (st *ShowChat) run(caller *ToolCaller, ui *UI) error {

	source_chat, err := NewChat(filepath.Join("..", st.AppName, "Chats", st.ChatFileName))
	if err != nil {
		return err
	}
	source_root, err := NewRoot("")
	if err != nil {
		return err
	}

	var dash *ChatMsg
	if source_chat.Dash_call_id != "" {
		dash = source_chat.FindUI(source_chat.Dash_call_id)
	}

	ui.SetColumn(0, 1, 100)
	ui.SetRow(0, 1, 100)
	ui.SetRowFromSub(1, 1, g_ShowChat_prompt_height, true)

	isRunning := (callFuncFindMsgName(source_chat.GetChatID()) != nil) //(st.isRunning != nil && st.isRunning())

	var MsgsDiv *UI
	if dash != nil {

		DashDiv := ui.AddLayout(0, 0, 1, 1)
		DashDiv.SetColumn(0, 1, 100)
		DashDiv.SetColumnResizable(1, 1, 50, 10)
		DashDiv.SetRow(0, 1, 100)

		//Dash
		{
			appUi, err := DashDiv.AddToolApp(0, 0, 1, 1, st.AppName, dash.UI_func, []byte(dash.UI_paramsJs), caller)
			if err != nil {
				return fmt.Errorf("AddToolApp() failed: %v", err)
			}
			appUi.changed = func(newParamsJs []byte) error {
				dash.UI_paramsJs = string(newParamsJs) //save back changes
				return nil
			}

		}

		ChatDiv := DashDiv.AddLayout(1, 0, 1, 1)
		ChatDiv.SetColumn(0, 1, 100)
		ChatDiv.SetRow(1, 1, 100)

		DashHeaderDiv := ChatDiv.AddLayout(0, 0, 1, 1)
		DashHeaderDiv.ScrollH.Narrow = true
		//DashHeaderDiv.ScrollV.Hide = true
		{
			preUI := source_chat.FindPreviousUI(source_chat.Dash_call_id)
			nxtUI := source_chat.FindNextUI(source_chat.Dash_call_id)

			btClose := DashHeaderDiv.AddButton(0, 0, 1, 1, "<<")
			btClose.layout.Tooltip = "Hide dashboard"
			btClose.Background = 0.5
			btClose.clicked = func() error {
				source_chat.Dash_call_id = "" //reset
				return nil
			}
			btBack := DashHeaderDiv.AddButton(2, 0, 1, 1, "<")
			btBack.layout.Tooltip = "Previous dashboard"
			btBack.Background = 0.5
			btBack.layout.Enable = (preUI != nil)
			btBack.clicked = func() error {
				if preUI != nil {
					source_chat.Dash_call_id = preUI.Content.Result.Tool_call_id
				}
				return nil
			}
			btForward := DashHeaderDiv.AddButton(3, 0, 1, 1, ">")
			btForward.layout.Tooltip = "Next dashboard"
			btForward.Background = 0.5
			btForward.layout.Enable = (nxtUI != nil)
			btForward.clicked = func() error {
				if nxtUI != nil {
					source_chat.Dash_call_id = nxtUI.Content.Result.Tool_call_id
				}
				return nil
			}

			//FuncName
			DashHeaderDiv.SetColumn(5, 3, 100)
			DashHeaderDiv.AddText(5, 0, 1, 1, dash.UI_func+"()").Multiline = false
		}

		//Chat
		MsgsDiv = ChatDiv.AddLayout(0, 1, 1, 1)
		err = st.buildShowMessages(MsgsDiv, caller, source_chat, source_root, isRunning)
		if err != nil {
			return fmt.Errorf("buildShowMessages1() failed: %v", err)
		}
	} else {
		//Chat
		MsgsDiv = ui.AddLayout(0, 0, 1, 1)
		err = st.buildShowMessages(MsgsDiv, caller, source_chat, source_root, isRunning)
		if err != nil {
			return fmt.Errorf("buildShowMessages2() failed: %v", err)
		}
	}

	//User prompt
	{
		DivInput := ui.AddLayout(0, 1, 1, 1)
		d := 0.25
		dd := 0.25
		DivInput.SetColumn(0, d, d) //space
		DivInput.SetColumn(1, 1, 100)
		DivInput.SetColumn(2, d, d) //space
		DivInput.SetRow(0, d, d)
		DivInput.SetRowFromSub(1, 1, g_ShowChat_prompt_height-0.5, true)
		DivInput.SetRow(2, d, d)

		Div := DivInput.AddLayout(1, 1, 1, 1)

		Div.SetColumn(0, dd, dd) //space
		Div.SetColumn(1, 1, 100)
		Div.SetColumn(2, dd, dd) //space
		Div.SetRow(0, dd, dd)
		Div.SetRowFromSub(1, 1, g_ShowChat_prompt_height-1, true)
		Div.SetRow(2, dd, dd)

		Div.Back_cd = UI_GetPalette().B //GetGrey(0.05)
		Div.Back_rounding = true
		Div.Border_cd = UI_GetPalette().GetGrey(0.2)

		err = st.buildInput(Div.AddLayout(1, 1, 1, 1), caller, source_chat, source_root, MsgsDiv, isRunning)
		if err != nil {
			return fmt.Errorf("buildInput() failed: %v", err)
		}
	}

	return nil
}

func (st *ChatInput) SetVoice(js []byte) error {
	//type VerboseJsonWord struct {
	//	Start, End float64
	//	Word       string
	//}
	type VerboseJsonSegment struct {
		Start, End float64
		Text       string
		//Words      []*VerboseJsonWord
	}
	type VerboseJson struct {
		Segments []*VerboseJsonSegment //later are projected to .Words
	}

	var verb VerboseJson
	err := json.Unmarshal(js, &verb)
	if err != nil {
		return fmt.Errorf("verb failed: %v", err)
	}

	//build prompt
	prompt := ""
	for _, seg := range verb.Segments {
		prompt += seg.Text
	}
	prompt = _ShowInput_replaceBackground(prompt)
	prompt = strings.TrimSpace(prompt)

	pick_i := 0
	words := []string{"this", "these", "here"}
	for pick_i < len(st.Picks) {
		min_p := len(prompt)
		word_n := 0
		for _, wd := range words {
			p := strings.Index(prompt, wd)
			if p >= 0 && p < min_p {
				min_p = p
				word_n = len(wd)
			}
		}
		if word_n > 0 {
			prompt = prompt[:min_p] + st.Picks[pick_i].Cd.GetLabel() + prompt[min_p+word_n:] //replace
			pick_i++

		} else {
			break
		}
	}

	st.Text = st.Text_mic + prompt

	return nil
}

func _ShowInput_replaceBackground(str string) string {
	str = strings.ReplaceAll(str, "[BLANK_AUDIO]", "")
	str = strings.ReplaceAll(str, "[NO_SPEECH]", "")
	str = strings.ReplaceAll(str, "[MUSIC]", "")
	str = strings.ReplaceAll(str, "[NOISE]", "")
	str = strings.ReplaceAll(str, "[LAUGHTER]", "")
	str = strings.ReplaceAll(str, "[APPLAUSE]", "")
	str = strings.ReplaceAll(str, "[UNKNOWN]", "")
	str = strings.ReplaceAll(str, "[INAUDIBLE]", "")
	return str
}

func _UiText_RemoveFormatingRGBA(str string) string {
	str = strings.ReplaceAll(str, "</rgba>", "")
	for {
		st := strings.Index(str, "<rgba")
		if st < 0 {
			break
		}
		en := strings.IndexByte(str[st:], '>')
		if en >= 0 {
			str = str[:st] + str[st+en+1:]
		}
	}
	return str
}

func (st *ChatInput) GetFullPrompt() (string, []string) {
	prompt := _UiText_RemoveFormatingRGBA(st.Text)

	legend := ""
	sign := 'A'
	for _, br := range st.Picks {
		label := fmt.Sprintf("{%s}", br.Cd.Label)

		if strings.Contains(prompt, label) {
			new_label := fmt.Sprintf("{%c}", sign)
			prompt = strings.ReplaceAll(prompt, label, new_label)
			sign++

			legend += "\n" + new_label + ": " + br.LLMTip
		}
	}
	if legend != "" {
		prompt += "\n" + legend
	}

	return prompt, st.Files
}

func (st *ShowChat) _sendIt(caller *ToolCaller, chat *Chat, root *Root, continuee bool) error {
	//MsgsDiv.VScrollToTheBottom(false, caller)
	caller.SendFlushCmd()

	if !continuee && chat.Input.Text == "" {
		return nil //empty text
	}

	caller.SetMsgName(chat.GetChatID())

	err := st.complete(caller, chat, root, continuee)
	if err != nil {
		return fmt.Errorf("complete() failed: %v", err)
	}
	if !continuee {
		chat.Input.Reset()
	}

	//MsgsDiv.VScrollToTheBottom(true, caller)
	return nil
}

func (st *ShowChat) buildInput(ui *UI, caller *ToolCaller, chat *Chat, root *Root, MsgsDiv *UI, isRunning bool) error {

	input := &chat.Input

	preview_height := 0.0
	if len(input.Files) > 0 {
		preview_height = 2
	}

	ui.SetRowFromSub(0, 1, g_ShowChat_prompt_height-1-preview_height, true)

	if isRunning {
		MsgsDiv.VScrollToTheBottom(true, caller)
	}

	sendIt := func() error {
		chat.Messages.Messages = append(chat.Messages.Messages, chat.TempMessages.Messages...)
		return st._sendIt(caller, chat, root, false)
	}

	x := 0
	y := 0
	{
		ui.SetColumnFromSub(x, 3, 5, true)
		DivStart := ui.AddLayout(x, y, 1, 1)
		DivStart.SetRow(0, 0, 100)
		DivStart.Enable = !isRunning
		x++

		xx := 0

		//Drop file
		{
			var filePath string
			fls := DivStart.AddFilePickerButton(xx, 1, 1, 1, &filePath, false, false)
			fls.changed = func() error {
				input.Files = append(input.Files, filePath)
				return nil
			}
			xx++
		}

		//Auto-send after recording
		{
			as := DivStart.AddCheckbox(xx, 1, 1, 1, "", &root.Autosend)
			as.layout.Tooltip = "Auto-send after recording"
			xx++
		}

		//Mic
		{
			mic := DivStart.AddMicrophone(xx, 1, 1, 1)
			mic.Transcribe = true
			mic.Transcribe_response_format = "verbose_json"
			mic.Shortcut = '\t'
			mic.Output_onlyTranscript = true

			mic.started = func() error {
				input.Text_mic = input.Text
				return nil
			}
			mic.finished = func(audio []byte, transcript string) error {

				//save
				err := input.SetVoice([]byte(transcript))
				if err != nil {
					return fmt.Errorf("SetVoice() failed: %v", err)
				}

				// auto-send
				if root.Autosend > 0 {
					sendIt()
				}

				return nil
			}
			xx++
		}

		//Reset brushes
		if len(input.Picks) > 0 {
			DivStart.SetColumn(xx, 2, 2)
			ClearBt := DivStart.AddButton(xx, 1, 1, 1, "Clear")
			ClearBt.Background = 0.5
			ClearBt.layout.Tooltip = "Remove Brushes"
			ClearBt.clicked = func() error {
				//remove
				//for i := range st.Picks {
				//	st.Text = strings.ReplaceAll(st.Text, st.Picks[i].Cd.GetLabel(), "")
				//}
				input.Text = ""
				input.Picks = nil
				return nil
			}
			xx++
		}
	}

	//Editbox
	{
		ui.SetColumn(x, 1, 100)
		ed := ui.AddEditboxString(x, y, 1, 1, &input.Text)
		ed.Ghost = "What can I do for you?"
		ed.Multiline = input.Multilined
		ed.enter = sendIt
		ed.Name = "chat_user_prompt"
		ed.layout.Enable = !isRunning
		x++
	}

	//switch multi-lined
	{
		DivML := ui.AddLayout(x, y, 1, 1)
		DivML.SetColumn(0, 1, 100)
		DivML.SetRow(0, 0, 100)
		DivML.Enable = !isRunning

		mt := DivML.AddButton(0, 1, 1, 1, "")
		mt.IconPath = "resources/multiline.png"
		mt.Icon_margin = 0.1
		mt.layout.Tooltip = "Enable/disable multi-line prompt"
		if !input.Multilined {
			mt.Background = 0
		}
		mt.clicked = func() error {
			input.Multilined = !input.Multilined
			return nil
		}
		x++

	}

	//Send button
	{
		ui.SetColumn(x, 2.5, 2.5)
		DivSend := ui.AddLayout(x, y, 1, 1)
		DivSend.SetColumn(0, 1, 100)
		DivSend.SetRow(0, 0, 100)
		if !isRunning {
			SendBt := DivSend.AddButton(0, 1, 1, 1, "Send")
			SendBt.IconPath = "resources/up.png"
			SendBt.Icon_margin = 0.2
			SendBt.Align = 0
			//SendBt.Tooltip = //name of "text" model ....
			SendBt.clicked = sendIt
		} else {
			StopBt := DivSend.AddButton(0, 1, 1, 1, "Stop")
			StopBt.Cd = UI_GetPalette().E
			StopBt.clicked = func() error {
				st.stop(chat)
				return nil
			}
		}
		x++
	}
	y++

	//show file previews
	if len(input.Files) > 0 {
		ui.SetRow(y, preview_height, preview_height)
		ImgsList := ui.AddLayoutList(0, y, x, 1, true)
		y++

		for fi, file := range input.Files {
			ImgDia := ui.AddDialog("image_" + file)
			ImgDia.UI.SetColumn(0, 5, 12)
			ImgDia.UI.SetColumn(1, 3, 3)
			ImgDia.UI.SetRow(1, 5, 15)
			ImgDia.UI.AddMediaPath(0, 1, 2, 1, file)
			ImgDia.UI.AddText(0, 0, 1, 1, file)
			RemoveBt := ImgDia.UI.AddButton(1, 0, 1, 1, "Remove")
			RemoveBt.clicked = func() error {
				input.Files = slices.Delete(input.Files, fi, fi+1)
				ImgDia.Close(caller)
				return nil
			}

			imgLay := ImgsList.AddItem()
			imgLay.SetColumn(0, 2, 2)
			imgLay.SetRow(0, 2, 2)
			imgBt := imgLay.AddButton(0, 0, 1, 1, "")
			imgBt.IconPath = file
			imgBt.Icon_margin = 0
			imgBt.layout.Tooltip = file

			imgBt.Background = 0
			imgBt.Cd = UI_GetPalette().B
			imgBt.Border = true
			imgBt.clicked = func() error {
				ImgDia.OpenRelative(imgBt.layout, caller)
				return nil
			}
		}

		//remove all files
		{
			delLay := ImgsList.AddItem()
			delLay.SetColumn(0, 2, 2)
			delLay.SetRow(0, 2, 2)
			delBt := delLay.AddButton(0, 0, 1, 1, "Delete All")
			delBt.Background = 0.5
			delBt.clicked = func() error {
				input.Files = nil
				return nil
			}
		}
	}

	//LLMTips/Brushes
	if len(input.Picks) > 0 {
		ui.SetRowFromSub(y, 1, 5, true)
		TipsDiv := ui.AddLayout(0, y, x, 1)
		y++
		TipsDiv.SetColumn(0, 2, 2)
		TipsDiv.SetColumn(1, 1, 100)

		yy := 0
		for i, br := range input.Picks {
			found_i := chat.Input.FindPick(br.LLMTip)
			if found_i >= 0 && found_i < i { //unique
				continue //skip
			}

			TipsDiv.SetRowFromSub(yy, 1, 3, true)
			TipsDiv.AddText(0, yy, 1, 1, input.Picks[i].Cd.GetLabel())
			TipsDiv.AddText(1, yy, 1, 1, strings.TrimSpace(br.LLMTip))
			yy++
		}
	}

	return nil
}

func (st *ShowChat) stop(chat *Chat) {
	callFuncMsgStop(chat.GetChatID())
}

func (st *ShowChat) complete(caller *ToolCaller, chat *Chat, root *Root, continuee bool) error {
	//sources := callFuncGetSources()

	needSummary := (len(chat.Messages.Messages) == 0 /*|| len(chat.Label) < 8*/)

	prompt, files := chat.Input.GetFullPrompt()

	var comp LLMCompletion
	comp.Temperature = 0.2
	comp.Max_tokens = 4096
	comp.Top_p = 0.95 //1.0
	comp.Frequency_penalty = 0
	comp.Presence_penalty = 0
	comp.Reasoning_effort = "" //low, high

	comp.AppName = st.AppName

	//add default(without source) tools
	/*{
		defSource := sources["_default_"]
		if defSource != nil {
			comp.Tools = append(comp.Tools, defSource.Tools...)
		}
		comp.Tools = append(comp.Tools, "InstallDataSource")

		//add 'installed' tools
		for _, sourceName := range chat.Sources {
			source := sources[sourceName]
			if source != nil {
				comp.Tools = append(comp.Tools, source.Tools...)
			}
		}
	}*/

	old_num_msgs := len(chat.Messages.Messages)
	var err error
	comp.PreviousMessages, err = json.Marshal(chat.Messages)
	if err != nil {
		return fmt.Errorf("comp.PreviousMessages failed: %v", err)
	}

	//system
	{
		system_prompt := `You are an AI assistant, who enjoys precision and carefully follows the user's requirements.
If you need, you can use tool calling. Tools gives you precision and output data which you wouldn't know otherwise. Don't ask to call a tool, just do it! Call tools sequentially. Avoid tool call as parameter value.
If user wants to show/render/visualize some data, search for tools which 'shows'."`

		//sources
		/*system_prompt += "Here is the list of data sources in form of '<name> //<description>':\n"
		for sourceName, source := range sources {
			system_prompt += fmt.Sprintf("%s //%s", sourceName, source.Description)
			if slices.Index(chat.Sources, sourceName) >= 0 {
				system_prompt += " [installed]"
			}
			system_prompt += "\n"
		}
		system_prompt += fmt.Sprintf("Based on what you need you can install new tools by calling InstallDataSource with ChatID=%d. This will give you access to new tools.\n", st.ChatID)
		*/

		//Date time
		system_prompt += fmt.Sprintf("\nCurrent time is %s\n", time.Now().Format("Mon, 02 Jan 2006 15:04"))

		//Memory
		if root.Memory != "" {
			system_prompt += "\n" + root.Memory
		}

		comp.SystemMessage = system_prompt
	}

	if !continuee {
		comp.UserMessage = prompt
		comp.UserFiles = files
	}
	comp.Max_iteration = 10
	comp.delta = func(msgJs []byte) {
		var msg ChatMsg
		err := json.Unmarshal(msgJs, &msg)
		if err != nil {
			return //err ....
		}

		last_i := len(chat.TempMessages.Messages) - 1
		if last_i >= 0 && chat.TempMessages.Messages[last_i].Stream {
			if msg.Stream {
				//copy
				msg.ShowParameters = chat.TempMessages.Messages[last_i].ShowParameters
				msg.ShowReasoning = chat.TempMessages.Messages[last_i].ShowReasoning
			}

			chat.TempMessages.Messages[last_i] = &msg //rewrite last
		} else {
			chat.TempMessages.Messages = append(chat.TempMessages.Messages, &msg)
		}

		//activate dash
		if msg.HasUI() {
			chat.Dash_call_id = msg.Content.Result.Tool_call_id
		}
	}
	chat.TempMessages = ChatMsgs{} //reset

	err = comp.Run(caller)
	if err != nil {
		return fmt.Errorf("LLMxAICompleteChat.run() failed: %v", err)
	}

	chat.TempMessages = ChatMsgs{} //reset
	err = json.Unmarshal(comp.Out_messages, &chat.Messages)
	if err != nil {
		return fmt.Errorf("comp.Out_messages failed: %v", err)
	}

	//activate new dash
	for i := old_num_msgs; i < len(chat.Messages.Messages); i++ {
		msg := chat.Messages.Messages[i]
		if msg.HasUI() {
			chat.Dash_call_id = msg.Content.Result.Tool_call_id
		}
	}

	if needSummary {
		st.summarize(prompt, caller, chat)
	}

	return nil
}

func (st *ShowChat) summarize(userMessage string, caller *ToolCaller, chat *Chat) {
	if len(userMessage) < 40 {
		chat.Label = userMessage //use user prompt as chat label

		//clean
		chat.Label = strings.ReplaceAll(chat.Label, "\n", " ")
		chat.Label = strings.TrimSpace(chat.Label)
	} else {
		var comp LLMCompletion
		comp.Temperature = 0.2
		comp.Max_tokens = 1024
		comp.Top_p = 0.95 //1.0
		comp.Frequency_penalty = 0
		comp.Presence_penalty = 0
		comp.Reasoning_effort = ""

		comp.SystemMessage = "You are an AI assistant. You summarize long text to clear short description."
		comp.UserMessage = "Summarize following text to maximum 10 words:\n" + userMessage
		comp.Max_iteration = 1

		err := comp.Run(caller)
		if err == nil {

			label := comp.Out_answer
			label = strings.ReplaceAll(label, "\n", " ")
			label = strings.TrimSpace(label)
			chat.Label = label
		}
	}
}

func (st *ShowChat) buildShowMessages(ui *UI, caller *ToolCaller, source_chat *Chat, source_root *Root, isRunning bool) error {
	ui.SetColumn(0, 1, 100)
	ui.SetColumn(1, 5, 20)
	ui.SetColumn(2, 1, 100)

	//Messages
	y := 0 //space

	for msg_i, msg := range source_chat.Messages.Messages {
		if msg.Content.Result != nil {
			//space
			//ui.SetRow(y, 0.5, 0.5)
			y++
			continue //skip
		}

		if msg.Usage.Provider == "" {
			y++ //space above user msg
		}

		ui.SetRowFromSub(y, 1, 100, true)
		st.AddChatMsg(ui.AddLayout(1, y, 1, 1), &source_chat.Messages, msg_i, source_chat, source_root, ui, caller)
		y++

		ui.SetRow(y, 0.5, 0.5)
		y++ //space
	}

	for msg_i, msg := range source_chat.TempMessages.Messages {
		if msg.Content.Result != nil {
			//space
			ui.SetRow(y, 0.5, 0.5)
			y++
			continue //skip
		}

		ui.SetRowFromSub(y, 1, 100, true)
		st.AddChatMsg(ui.AddLayout(1, y, 1, 1), &source_chat.TempMessages, msg_i, source_chat, source_root, ui, caller)
		y++

		ui.SetRow(y, 0.5, 0.5)
		y++ //space
	}

	//Button Continue
	if !isRunning && len(source_chat.TempMessages.Messages) > 0 {

		btContinue := ui.AddButton(1, y, 1, 1, "Continue")
		btContinue.Cd = UI_GetPalette().E
		btContinue.clicked = func() error {
			return st._sendIt(caller, source_chat, source_root, true)
		}
		y++

	}

	//Statistics - total
	if y >= 2 { //1st message is user
		in, inCached, out := source_chat.Messages.GetTotalPrice(0, -1)
		info := ui.AddText(1, y, 1, 1, fmt.Sprintf("<i>$%s, %d tokens, %s sec, %d tokens/sec",
			strconv.FormatFloat(in+inCached+out, 'f', 3, 64),
			source_chat.Messages.GetTotalOutputTokens(0, -1),
			strconv.FormatFloat(source_chat.Messages.GetTotalTime(0, -1), 'f', 3, 64),
			int(source_chat.Messages.GetTotalSpeed(0, -1))))
		y++
		info.Align_h = 2 //right
		info.layout.Tooltip = fmt.Sprintf("%s seconds\n%d input tokens\n%d output tokens\n%s tokens/sec\nTotal: $%s\n- Input: $%s\n- Input cached: $%s\n- Output: $%s",
			strconv.FormatFloat(source_chat.Messages.GetTotalTime(0, -1), 'f', -1, 64),
			source_chat.Messages.GetTotalInputTokens(0, -1),
			source_chat.Messages.GetTotalOutputTokens(0, -1),
			strconv.FormatFloat(source_chat.Messages.GetTotalSpeed(0, -1), 'f', 3, 64),
			strconv.FormatFloat(in+inCached+out, 'f', -1, 64),
			strconv.FormatFloat(in, 'f', -1, 64),
			strconv.FormatFloat(inCached, 'f', -1, 64),
			strconv.FormatFloat(out, 'f', -1, 64))
	}

	return nil
}

func (st *ShowChat) AddChatMsg(layout *UI, msgs *ChatMsgs, msg_i int, chat *Chat, root *Root, MsgsDiv *UI, caller *ToolCaller) {
	msg := msgs.Messages[msg_i]

	layout.SetColumn(0, 1, 15)
	layout.SetColumn(1, 1, 4)
	layout.SetColumn(2, 1, 100)

	if msg.Usage.Provider != "" {
		layout.Back_cd = UI_GetPalette().GetGrey(0.09)
		layout.Back_rounding = true
	}

	y := 0

	if msg.Content.Msg != nil {
		hasImage := false
		for _, it := range msg.Content.Msg.Content {
			switch it.Type {
			case "text":
				txt := strings.TrimSpace(it.Text)
				if txt != "" {
					layout.SetRowFromSub(y, 1, 100, true)

					UserDiv := layout.AddLayout(0, y, 3, 1)
					UserDiv.SetColumn(0, 0, 100)
					UserDiv.SetColumnFromSub(1, 1, 100, true)
					UserDiv.SetRowFromSub(0, 1, 100, true)
					y++

					tx := UserDiv.AddText(1, 0, 1, 1, txt)
					tx.Multiline = true
					//tx.Align_v = 0
					tx.layout.Border_cd = UI_GetPalette().GetGrey(0.2)
					tx.layout.Back_cd = UI_GetPalette().B
					tx.layout.Back_rounding = true

				}
			case "image_url":
				hasImage = true
			}
		}

		if hasImage {
			layout.SetRow(y, 2, 2)
			ImgsList := layout.AddLayoutList(0, y, 3, 1, true)
			y++

			for i, it := range msg.Content.Msg.Content {
				if it.Image_url == nil {
					continue
				}

				switch it.Type {
				case "image_url":

					var imgBlob []byte
					sep := strings.IndexByte(it.Image_url.Url, ',')

					if sep >= 0 {
						imgBlob, _ = base64.StdEncoding.DecodeString(it.Image_url.Url[sep+1:])
					}

					ImgDia := layout.AddDialog(fmt.Sprintf("msg_%d_image_%d", msg_i, i))
					ImgDia.UI.SetColumn(0, 5, 15)
					ImgDia.UI.SetRow(0, 5, 15)
					ImgDia.UI.AddMediaBlob(0, 0, 1, 1, imgBlob)

					imgLay := ImgsList.AddItem()
					imgLay.SetColumn(0, 2, 2)
					imgLay.SetRow(0, 2, 2)
					imgBt := imgLay.AddButton(0, 0, 1, 1, "")
					imgBt.IconBlob = imgBlob
					imgBt.Icon_margin = 0

					imgBt.Background = 0
					imgBt.Cd = UI_GetPalette().B
					imgBt.Border = true
					imgBt.clicked = func() error {
						ImgDia.OpenRelative(imgBt.layout, caller)
						return nil
					}

				}
			}
		}
	}

	txt := ""
	rsp_txt := ""
	if msg.Content.Calls != nil {
		txt = msg.Content.Calls.Content
		{
			if msg.ReasoningSize > 0 && len(txt) >= msg.ReasoningSize {
				rsp_txt = txt[:msg.ReasoningSize]
				txt = txt[msg.ReasoningSize:]
			}
			rsp_txt = strings.TrimSpace(rsp_txt)
			txt = strings.TrimSpace(txt)

			if txt != "" {
				layout.SetRowFromSub(y, 1, 100, true)
				tx := layout.AddText(0, y, 3, 1, txt)
				tx.Multiline = true
				//tx.Align_v = 0
				y++
			}

			if rsp_txt != "" && msg.ShowReasoning {
				//divider
				if txt != "" {
					layout.SetRow(y, 0.2, 0.2)
					layout.AddDivider(0, y, 3, 1, true)
					y++
				}

				//text
				layout.SetRowFromSub(y, 1, 100, true)
				tx := layout.AddText(0, y, 3, 1, rsp_txt)
				tx.Multiline = true
				//tx.Align_v = 0
				y++
			}
		}

	}

	{
		DivIcons := layout.AddLayout(0, y, 3, 1)

		iconsCd := UI_GetPalette().GetGrey(0.5)

		x := 0

		if msg.Usage.Provider != "" {
			// show/hide reasoning
			if rsp_txt != "" {
				ShowRspBt := DivIcons.AddButton(x, 0, 1, 1, "")
				ShowRspBt.Background = 0.2
				if msg.ShowReasoning {
					ShowRspBt.Background = 1
				}
				ShowRspBt.layout.Tooltip = "Show reasoning"
				if msg.ShowReasoning {
					ShowRspBt.layout.Tooltip = "Hide reasoning"
				}
				ShowRspBt.Align = 0
				ShowRspBt.Icon_align = 0
				ShowRspBt.Icon_margin = 0.25
				ShowRspBt.IconPath = "resources/think.png"
				ShowRspBt.clicked = func() error {
					msg.ShowReasoning = !msg.ShowReasoning
					return nil
				}
				x++
			}
		}

		//long space
		DivIcons.SetColumn(x, 0, 100)
		x++

		if msg.Usage.Provider == "" {
			{
				RgnBt := DivIcons.AddButton(x, 0, 1, 1, "")
				RgnBt.Cd = iconsCd
				RgnBt.Icon_margin = 0.3
				RgnBt.IconPath = "resources/reload.png"
				RgnBt.Background = 0.2
				RgnBt.layout.Tooltip = "Re-generate response"
				RgnBt.clicked = func() error {

					msgs.Messages[msg_i].Seed++

					msgs.Messages = msgs.Messages[:msg_i+1] //cut
					chat.TempMessages = ChatMsgs{}          //reset

					st._sendIt(caller, chat, root, true)

					return nil
				}
				x++
			}

			{
				DelBt := DivIcons.AddButton(x, 0, 1, 1, "")
				DelBt.ConfirmQuestion = "Are you sure?"
				DelBt.Cd = iconsCd
				DelBt.Icon_margin = 0.25
				DelBt.IconPath = "resources/delete.png"
				DelBt.Background = 0.2
				DelBt.layout.Tooltip = "Delete the message and response"
				DelBt.clicked = func() error {
					//find next user
					next_i := msg_i + 1
					for next_i < len(msgs.Messages) {
						if msgs.Messages[next_i].Usage.Provider == "" {
							break
						}
						next_i++
					}

					msgs.Messages = slices.Delete(msgs.Messages, msg_i, next_i)
					return nil
				}
				x++
			}
		} else {

			{
				in := msg.Usage.Prompt_price
				inCached := msg.Usage.Input_cached_price
				out := msg.Usage.Completion_price + msg.Usage.Reasoning_price
				inf := fmt.Sprintf("<b>%s</b>\n%s\nTime to first token: %s sec\nTime: %s sec\n%s tokens/sec\nTotal: $%s\n- Input: $%s(%d toks)\n- Cached: $%s(%d toks)\n- Output: $%s(%d+%d toks)",
					msg.Usage.Provider+":"+msg.Usage.Model,
					SdkGetDateTime(int64(msg.Usage.CreatedTimeSec)),
					strconv.FormatFloat(msg.Usage.TimeToFirstToken, 'f', 3, 64),
					strconv.FormatFloat(msg.Usage.DTime, 'f', 3, 64),
					strconv.FormatFloat(msg.Usage.GetSpeed(), 'f', 3, 64),
					strconv.FormatFloat(in+inCached+out, 'f', -1, 64),
					strconv.FormatFloat(in, 'f', -1, 64),
					msg.Usage.Prompt_tokens,
					strconv.FormatFloat(inCached, 'f', -1, 64),
					msg.Usage.Input_cached_tokens,
					strconv.FormatFloat(out, 'f', -1, 64),
					msg.Usage.Reasoning_tokens,
					msg.Usage.Completion_tokens)

				InfoDia := DivIcons.AddDialog("info")
				InfoDia.UI.SetColumn(0, 5, 10)
				InfoDia.UI.SetRowFromSub(0, 1, 20, true)
				InfoDia.UI.AddText(0, 0, 1, 1, inf)

				InfoBt := DivIcons.AddButton(x, 0, 1, 1, "")
				InfoBt.Cd = iconsCd
				InfoBt.Icon_margin = 0.3
				InfoBt.IconPath = "resources/about.png"
				InfoBt.Background = 0.2
				InfoBt.layout.Tooltip = inf
				InfoBt.clicked = func() error {
					InfoDia.OpenRelative(InfoBt.layout, caller)
					return nil
				}
				x++
			}

		}

		{
			CopyBt := DivIcons.AddButton(x, 0, 1, 1, "")
			CopyBt.Cd = iconsCd
			CopyBt.Icon_margin = 0.3
			CopyBt.IconPath = "resources/copy.png"
			CopyBt.Background = 0.2
			CopyBt.layout.Tooltip = "Copy into clipboard"
			CopyBt.clicked = func() error {
				if msg.Content.Msg != nil && len(msg.Content.Msg.Content) > 0 {
					caller.SetClipboardText(msg.Content.Msg.Content[0].Text)
				} else if msg.Content.Calls != nil {
					if msg.ShowReasoning {
						caller.SetClipboardText(msg.Content.Calls.Content)
					} else {
						caller.SetClipboardText(txt)
					}
				}

				return nil
			}
			x++
		}
		y++
	}

	//list of tool calls
	if msg.Content.Calls != nil {
		for _, call := range msg.Content.Calls.Tool_calls {
			y = st.toolUse(call, layout, y, msg, chat)
		}
	}

}

func (st *ShowChat) toolUse(it OpenAI_completion_msg_Content_ToolCall, layout *UI, y int, msg *ChatMsg, chat *Chat) int {
	// open sub-agent
	ShowParamsBt := layout.AddButton(0, y, 1, 1, fmt.Sprintf("<i>%s(%s)", it.Function.Name, it.Id))
	ShowParamsBt.Background = 0.5
	ShowParamsBt.Align = 0
	ShowParamsBt.Icon_align = 0
	ShowParamsBt.Icon_margin = 0.1

	if msg.ShowParameters {
		ShowParamsBt.IconPath = "resources/arrow_down.png"
	} else {
		ShowParamsBt.IconPath = "resources/arrow_right.png"
	}
	ShowParamsBt.clicked = func() error {
		msg.ShowParameters = !msg.ShowParameters
		return nil
	}

	msg_result, _ := chat.Messages.FindResultContent(it.Id)

	if msg_result != nil && msg_result.HasUI() {
		isOpen := (chat.Dash_call_id == msg_result.Content.Result.Tool_call_id)

		stateStr := "Show"
		if isOpen {
			stateStr = "Hide"
		}

		bt := layout.AddButton(1, y, 1, 1, stateStr)
		if isOpen {
			bt.Background = 1
			bt.layout.Tooltip = "Close dashboard"
		} else {
			bt.Border = true
			bt.Background = 0.25
			bt.layout.Tooltip = "Show dashboard"
		}

		bt.clicked = func() error {
			if isOpen {
				chat.Dash_call_id = "" //close
			} else {
				chat.Dash_call_id = msg_result.Content.Result.Tool_call_id
			}
			return nil
		}
	}

	y++

	// parameters
	if msg.ShowParameters {
		layout.SetRowFromSub(y, 1, 100, true)
		CallDiv := layout.AddLayout(0, y, 3, 1)
		CallDiv.SetColumnFromSub(1, 0, 100, true)
		CallDiv.SetColumn(2, 1, 100)
		CallDiv.Back_cd = Color_Aprox(UI_GetPalette().P, UI_GetPalette().B, 0.8)
		y++

		yy := 0
		CallDiv.AddText(0, yy, 3, 1, "Inputs:")
		yy++

		//render tool
		attrs, err := it.Function.GetArgsAsStrings()
		if err == nil {
			//sort map by key
			var keys []string
			for k := range attrs {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, key := range keys {
				val := attrs[key]

				valStr := fmt.Sprintf("%s", val)

				CallDiv.SetRowFromSub(yy, 1, 100, true)

				CallDiv.AddText(1, yy, 1, 1, "<i>"+key) //name
				//tx.Tooltip = //description from agent tool ....

				if strings.Count(valStr, "\n") > 0 {
					tx := CallDiv.AddText(2, yy, 1, 1, valStr)
					tx.Multiline = true
					//tx.Align_v = 0
				} else {
					CallDiv.AddText(2, yy, 1, 1, valStr)
				}
				yy++
			}
		}

		if msg_result != nil {
			CallDiv.AddText(0, yy, 3, 1, "Output:")
			yy++
			CallDiv.SetRowFromSub(yy, 1, 100, true)
			tx := CallDiv.AddText(1, yy, 2, 1, msg_result.Content.Result.Content)
			tx.Multiline = true
			//tx.Align_v = 0
			yy++
		}
	}

	return y
}
