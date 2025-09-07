package main

import (
	"encoding/base64"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type Chat struct {
	Messages          []*ChatMsg
	Selected_user_msg int

	changed func(regenerate bool)
}

func (layout *Layout) AddChat(x, y, w, h int, messages []*ChatMsg, selected_user_msg int) *Chat {
	props := &Chat{Messages: messages, Selected_user_msg: selected_user_msg}
	layout._createDiv(x, y, w, h, "Chat", props.Build, nil, nil)
	return props
}

func (st *Chat) Build(layout *Layout) {

	layout.SetColumn(0, 1, Layout_MAX_SIZE)

	//Messages
	y := 0 //space

	for msg_i, msg := range st.Messages {
		if msg.Content.Result != nil {
			//space
			//ui.SetRow(y, 0.5, 0.5)
			y++
			continue //skip
		}

		if msg.Usage.Provider == "" {
			y++ //space above user msg
		}

		layout.SetRowFromSub(y, 1, Layout_MAX_SIZE, true)
		st.AddChatMsg(layout.AddLayout(0, y, 1, 1), msg_i, layout)
		y++

		layout.SetRow(y, 0.5, 0.5)
		y++ //space
	}

	/*for msg_i, msg := range source_chat.TempMessages.Messages {
		if msg.Content.Result != nil {
			//space
			ui.SetRow(y, 0.5, 0.5)
			y++
			continue //skip
		}

		ui.SetRowFromSub(y, 1, Layout_MAX_SIZE, true)
		st.AddChatMsg(ui.AddLayout(1, y, 1, 1), &source_chat.TempMessages, msg_i, source_chat, source_root, ui, caller)
		y++

		ui.SetRow(y, 0.5, 0.5)
		y++ //space
	}*/

	//Button Continue
	/*if !isRunning && len(source_chat.TempMessages.Messages) > 0 {

		btContinue := ui.AddButton(1, y, 1, 1, "Continue")
		btContinue.Cd = layout.GetPalette().E
		btContinue.clicked = func() error {
			return source_chat._sendIt(st.AppName, caller, source_root, true)
		}
		y++
	}*/

	//Statistics - total
	if y >= 2 { //1st message is user
		layout.SetRowFromSub(y, 1, 2, true)

		in, inCached, out, sources := st.GetTotalPrice(0, -1)
		info, ilay := layout.AddText2(0, y, 1, 1, fmt.Sprintf("<i>$%s, %d tokens, %s, %d tokens/sec",
			strconv.FormatFloat(in+inCached+out+sources, 'f', 3, 64),
			st.GetTotalOutputTokens(0, -1),
			layout.ConvertDTime(st.GetTotalTime(0, -1)),
			int(st.GetTotalSpeed(0, -1))))
		y++
		info.Align_h = 2 //right
		ilay.Tooltip = fmt.Sprintf("%s seconds\n%d input tokens\n%d output tokens\n%s tokens/sec\nTotal: $%s\n- Input: $%s\n- Input cached: $%s\n- Output: $%s\n- Sources: $%s",
			strconv.FormatFloat(st.GetTotalTime(0, -1), 'f', -1, 64),
			st.GetTotalInputTokens(0, -1),
			st.GetTotalOutputTokens(0, -1),
			strconv.FormatFloat(st.GetTotalSpeed(0, -1), 'f', 3, 64),
			strconv.FormatFloat(in+inCached+out, 'f', -1, 64),
			strconv.FormatFloat(in, 'f', -1, 64),
			strconv.FormatFloat(inCached, 'f', -1, 64),
			strconv.FormatFloat(out, 'f', -1, 64),
			strconv.FormatFloat(sources, 'f', -1, 64))
	}
}

func (st *Chat) AddChatMsg(layout *Layout, msg_i int, MsgsDiv *Layout) {
	msg := st.Messages[msg_i]

	layout.SetColumn(0, 1, 15)
	layout.SetColumn(1, 1, 4)
	layout.SetColumn(2, 1, Layout_MAX_SIZE)

	if msg.Usage.Provider != "" {
		layout.Back_cd = layout.GetPalette().GetGrey(0.09)
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
					layout.SetRowFromSub(y, 1, Layout_MAX_SIZE, true)

					UserDiv := layout.AddLayout(0, y, 3, 1)
					UserDiv.SetColumn(0, 0, Layout_MAX_SIZE)
					UserDiv.SetColumnFromSub(1, 1, Layout_MAX_SIZE, true)
					UserDiv.SetRowFromSub(0, 1, Layout_MAX_SIZE, true)
					y++

					tx, tlay := UserDiv.AddText2(1, 0, 1, 1, txt)
					tx.SetMultilined()
					//tx.Align_v = 0
					tlay.Border_cd = layout.GetPalette().GetGrey(0.2)
					tlay.Back_cd = layout.GetPalette().B
					tlay.Back_rounding = true

				}
			case "image_url":
				hasImage = true
			}
		}

		if hasImage {
			layout.SetRow(y, 2, 2)
			ImgsCards := layout.AddLayoutCards(0, y, 3, 1, true)
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
					ImgDia.Layout.SetColumn(0, 5, 15)
					ImgDia.Layout.SetRow(0, 5, 15)
					ImgDia.Layout.AddImage(0, 0, 1, 1, "", imgBlob)

					imgLay := ImgsCards.AddCardsSubItem()
					imgLay.SetColumn(0, 2, 2)
					imgLay.SetRow(0, 2, 2)
					imgBt, ilay := imgLay.AddButton2(0, 0, 1, 1, "")
					imgBt.IconBlob = imgBlob
					imgBt.Icon_margin = 0

					imgBt.Background = 0
					imgBt.Cd = layout.GetPalette().B
					imgBt.Border = true
					imgBt.clicked = func() {
						ImgDia.OpenRelative(ilay.UID)
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
				layout.SetRowFromSub(y, 1, Layout_MAX_SIZE, true)
				tx := layout.AddText(0, y, 3, 1, txt)
				tx.Multiline = true
				tx.SetMultilined()
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
				layout.SetRowFromSub(y, 1, Layout_MAX_SIZE, true)
				tx := layout.AddText(0, y, 3, 1, rsp_txt)
				tx.SetMultilined()
				//tx.Align_v = 0
				y++
			}
		}

	}

	//list of citations
	if len(msg.Citations) > 0 {
		layout.AddText(0, y, 3, 1, "Citations:")
		for i, c := range msg.Citations {
			layout.AddText(0, y, 3, 1, fmt.Sprintf("\t[%d] %s", i, c))
			y++
		}
	}

	//list of tool calls
	if msg.Content.Calls != nil {
		for _, call := range msg.Content.Calls.Tool_calls {
			toolDiv := layout.AddLayout(0, y, 3, 1)
			layout.SetRowFromSub(y, 1, Layout_MAX_SIZE, true)
			st.toolUse(call, toolDiv, msg)
			y++
		}
	}

	{
		DivIcons := layout.AddLayout(0, y, 3, 1)

		iconsCd := layout.GetPalette().GetGrey(0.5)

		x := 0

		if msg.Usage.Provider != "" {
			// show/hide reasoning
			if rsp_txt != "" {
				ShowRspBt, slay := DivIcons.AddButton2(x, 0, 1, 1, "")
				ShowRspBt.Background = 0.2
				if msg.ShowReasoning {
					ShowRspBt.Background = 1
				}
				slay.Tooltip = "Show reasoning"
				if msg.ShowReasoning {
					slay.Tooltip = "Hide reasoning"
				}
				ShowRspBt.Align = 0
				ShowRspBt.Icon_margin = 0.25
				ShowRspBt.IconPath = "resources/think.png"
				ShowRspBt.clicked = func() {
					msg.ShowReasoning = !msg.ShowReasoning
					if st.changed != nil {
						st.changed(false)
					}
				}
				x++
			}
		}

		//long space
		DivIcons.SetColumn(x, 0, Layout_MAX_SIZE)
		x++

		if msg.Usage.Provider == "" {
			{
				RgnBt, rlay := DivIcons.AddButton2(x, 0, 1, 1, "")
				RgnBt.Cd = iconsCd
				RgnBt.Icon_margin = 0.3
				RgnBt.IconPath = "resources/reload.png"
				RgnBt.Background = 0.2
				rlay.Tooltip = "Re-generate response"
				RgnBt.clicked = func() {

					st.Messages[msg_i].Seed++

					st.Messages = st.Messages[:msg_i+1] //cut
					//chat.TempMessages = ChatMsgs{}      //reset
					//chat._sendIt(st.AppName, caller, root, true)
					if st.changed != nil {
						st.changed(true)
					}
				}
				x++
			}

			{
				DelBt, dlay := DivIcons.AddButtonConfirm2(x, 0, 1, 1, "", "Are you sure?")
				DelBt.Color = iconsCd
				DelBt.Icon_margin = 0.25
				DelBt.IconPath = "resources/delete.png"
				DelBt.Background = 0.2
				dlay.Tooltip = "Delete the message and response"
				DelBt.confirmed = func() {
					//find next user
					next_i := msg_i + 1
					for next_i < len(st.Messages) {
						if st.Messages[next_i].Usage.Provider == "" {
							break
						}
						next_i++
					}

					st.Messages = slices.Delete(st.Messages, msg_i, next_i)
				}
				x++
			}
		} else {

			{
				in := msg.Usage.Prompt_price
				inCached := msg.Usage.Input_cached_price
				out := msg.Usage.Completion_price + msg.Usage.Reasoning_price
				sources := msg.Usage.Sources_price
				inf := fmt.Sprintf("<b>%s</b>\n%s\nTime to first token: %s sec\nTime: %s\n%s tokens/sec\nTotal: $%s\n- Input: $%s(%d toks)\n- Cached: $%s(%d toks)\n- Output: $%s(%d+%d toks)\n- Sources: $%s(%d links)",
					msg.Usage.Provider+":"+msg.Usage.Model,
					layout.ConvertTextDateTime(int64(msg.Usage.CreatedTimeSec)),
					strconv.FormatFloat(msg.Usage.TimeToFirstToken, 'f', 3, 64),
					layout.ConvertDTime(msg.Usage.DTime),
					strconv.FormatFloat(msg.Usage.GetSpeed(), 'f', 3, 64),
					strconv.FormatFloat(in+inCached+out+sources, 'f', -1, 64),
					strconv.FormatFloat(in, 'f', -1, 64),
					msg.Usage.Prompt_tokens,
					strconv.FormatFloat(inCached, 'f', -1, 64),
					msg.Usage.Input_cached_tokens,
					strconv.FormatFloat(out, 'f', -1, 64),
					msg.Usage.Reasoning_tokens,
					msg.Usage.Completion_tokens,
					strconv.FormatFloat(sources, 'f', -1, 64),
					msg.Usage.Num_sources_used)

				InfoDia := DivIcons.AddDialog("info")
				InfoDia.Layout.SetColumn(0, 5, 10)
				InfoDia.Layout.SetRowFromSub(0, 1, 20, true)
				info := InfoDia.Layout.AddText(0, 0, 1, 1, inf)
				info.Align_v = 0
				info.SetMultilined()

				InfoBt, ilay := DivIcons.AddButton2(x, 0, 1, 1, "")
				InfoBt.Cd = iconsCd
				InfoBt.Icon_margin = 0.3
				InfoBt.IconPath = "resources/about.png"
				InfoBt.Background = 0.2
				ilay.Tooltip = inf
				InfoBt.clicked = func() {
					InfoDia.OpenRelative(ilay.UID)
				}
				x++
			}
		}

		{
			CopyBt, clay := DivIcons.AddButton2(x, 0, 1, 1, "")
			CopyBt.Cd = iconsCd
			CopyBt.Icon_margin = 0.3
			CopyBt.IconPath = "resources/copy.png"
			CopyBt.Background = 0.2
			clay.Tooltip = "Copy into clipboard"
			CopyBt.clicked = func() {
				if msg.Content.Msg != nil && len(msg.Content.Msg.Content) > 0 {
					layout.ui.win.SetClipboardText(msg.Content.Msg.Content[0].Text)
				} else if msg.Content.Calls != nil {
					if msg.ShowReasoning {
						layout.ui.win.SetClipboardText(msg.Content.Calls.Content)
					} else {
						layout.ui.win.SetClipboardText(txt)
					}
				}
			}
			x++
		}
		y++
	}

}

func (st *Chat) GetTotalPrice(st_i, en_i int) (input, inCached, output, sources float64) {
	if en_i < 0 {
		en_i = len(st.Messages)
	}
	for i := st_i; i < en_i; i++ {
		input += st.Messages[i].Usage.Prompt_price
		inCached += st.Messages[i].Usage.Input_cached_price
		output += st.Messages[i].Usage.Completion_price + st.Messages[i].Usage.Reasoning_price
		sources += st.Messages[i].Usage.Sources_price
	}

	return
}

func (st *Chat) GetTotalSpeed(st_i, en_i int) float64 {
	toks := st.GetTotalOutputTokens(st_i, en_i)
	dt := st.GetTotalTime(st_i, en_i)
	if dt == 0 {
		return 0
	}
	return float64(toks) / dt

}

func (st *Chat) GetTotalTime(st_i, en_i int) float64 {
	dt := 0.0

	if en_i < 0 {
		en_i = len(st.Messages)
	}
	for i := st_i; i < en_i; i++ {
		dt += st.Messages[i].Usage.DTime
	}

	return dt
}

func (st *Chat) GetTotalInputTokens(st_i, en_i int) int {
	tokens := 0

	if en_i < 0 {
		en_i = len(st.Messages)
	}
	for i := st_i; i < en_i; i++ {
		tokens += st.Messages[i].Usage.Prompt_tokens
	}

	return tokens
}
func (st *Chat) GetTotalOutputTokens(st_i, en_i int) int {
	tokens := 0

	if en_i < 0 {
		en_i = len(st.Messages)
	}
	for i := st_i; i < en_i; i++ {
		tokens += st.Messages[i].Usage.Completion_tokens + st.Messages[i].Usage.Reasoning_tokens
	}

	return tokens
}

func (st *Chat) FindResultContent(call_id string) (*ChatMsg, int) {
	for i, m := range st.Messages {
		if m.Content.Result != nil && m.Content.Result.Tool_call_id == call_id {
			return m, i
		}
	}
	return nil, -1
}

func (st *Chat) FindToolCallUserMessage(tool_call_id string) int {
	n := 0
	for _, msg := range st.Messages {
		if msg.Content.Msg != nil {
			n++
		}
		if msg.Content.Result != nil && msg.Content.Result.Tool_call_id == tool_call_id {
			return n - 1
		}
	}
	/*for _, msg := range st.TempMessages.Messages {
		if msg.Content.Msg != nil {
			n++
		}
		if msg.Content.Result != nil && msg.Content.Result.Tool_call_id == tool_call_id {
			return n - 1
		}
	}*/
	return -1
}

func (st *Chat) toolUse(it OpenAI_completion_msg_Content_ToolCall, layout *Layout, msg *ChatMsg) {

	msg_result, _ := st.FindResultContent(it.Id)

	x := 0

	if msg_result != nil && msg_result.HasUI() {
		user_msg_i := st.FindToolCallUserMessage(msg_result.Content.Result.Tool_call_id)
		isOpen := (st.Selected_user_msg == user_msg_i)

		stateStr := "Show"

		layout.SetColumn(x, 3, 3)
		bt, blay := layout.AddButton2(x, 0, 1, 1, stateStr)
		x++
		blay.Enable = !isOpen
		if isOpen {
			bt.Background = 1
		} else {
			bt.Border = true
			bt.Background = 0.25
			blay.Tooltip = "Show dashboard"
		}

		bt.clicked = func() {
			st.Selected_user_msg = user_msg_i //msg_result.Content.Result.Tool_call_id
			if st.changed != nil {
				st.changed(false)
			}
		}
	}

	// open sub-agent
	layout.SetColumn(x, 1, Layout_MAX_SIZE)
	ShowParamsBt := layout.AddButton(x, 0, 1, 1, fmt.Sprintf("<i>%s(%s)", it.Function.Name, it.Id))
	x++
	ShowParamsBt.Background = 0.5
	ShowParamsBt.Align = 0
	ShowParamsBt.Icon_margin = 0.1

	if msg.ShowParameters {
		ShowParamsBt.IconPath = "resources/arrow_down.png"
	} else {
		ShowParamsBt.IconPath = "resources/arrow_right.png"
	}
	ShowParamsBt.clicked = func() {
		msg.ShowParameters = !msg.ShowParameters
		if st.changed != nil {
			st.changed(false)
		}
	}

	// parameters
	if msg.ShowParameters {
		layout.SetRowFromSub(1, 1, Layout_MAX_SIZE, true)
		CallDiv := layout.AddLayout(0, 1, 3, 1)
		CallDiv.SetColumnFromSub(1, 0, Layout_MAX_SIZE, true)
		CallDiv.SetColumn(2, 1, Layout_MAX_SIZE)
		CallDiv.Back_cd = Color_Aprox(layout.GetPalette().P, layout.GetPalette().B, 0.8)

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

				CallDiv.SetRowFromSub(yy, 1, Layout_MAX_SIZE, true)

				CallDiv.AddText(1, yy, 1, 1, "<i>"+_ShowChat_cutString(key)) //name
				//tx.Tooltip = //description from agent tool ....

				if strings.Count(valStr, "\n") > 0 {
					tx := CallDiv.AddText(2, yy, 1, 1, _ShowChat_cutString(valStr))
					tx.SetMultilined()
					//tx.Align_v = 0
				} else {
					CallDiv.AddText(2, yy, 1, 1, _ShowChat_cutString(valStr))
				}
				yy++
			}
		}

		if msg_result != nil {
			CallDiv.AddText(0, yy, 3, 1, "Output:")
			yy++
			CallDiv.SetRowFromSub(yy, 1, Layout_MAX_SIZE, true)

			tx := CallDiv.AddText(1, yy, 2, 1, _ShowChat_cutString(msg_result.Content.Result.Content))
			tx.SetMultilined()
			//tx.Align_v = 0
			yy++
		}
	}

}

func _ShowChat_cutString(str string) string {
	if len(str) > 250 {
		str = str[:200] + fmt.Sprintf(" (+ %d more)", len(str)-200)
	}
	return str
}
