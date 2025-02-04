package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
)

type ChatMsg struct {
	agent *Agent
	msg_i int

	isRunning func() bool
	uiChanged func()
}

func (layout *Layout) AddChatMsg(x, y, w, h int, props *ChatMsg) *ChatMsg {
	layout._createDiv(x, y, w, h, "ChatMsg", props.Build, nil, nil)
	return props
}

func (st *ChatMsg) Build(layout *Layout) {
	layout.SetColumn(0, 1, 100)
	layout.SetColumn(1, 3, 3) //date

	msg := st.agent.Messages[st.msg_i]

	sender := "User"
	if msg.CreatedBy != "" {
		sender = msg.CreatedBy
		layout.Back_cd = Paint_GetPalette().GetGrey(0.1)
	} else {
		//tool result
		//if len(st.Content) > 0 && st.Content[0].Type == "tool_result" {
		//	sender = fmt.Sprintf("User's %s result", st.Content[0].Tool_use_id)
		//}
	}

	//model name
	{
		model := layout.AddText(0, 0, 1, 1, "<b>"+sender)
		in, inCached, out := msg.GetPrice(st.agent)
		model.Tooltip = fmt.Sprintf("%s tokens/sec\n$%s(%d toks) total</b>\n- $%s(%d toks) input\n- $%s(%d toks) cached\n- $%s(%d toks) output",
			strconv.FormatFloat(msg.GetSpeed(), 'f', 3, 64),
			strconv.FormatFloat(in+inCached+out, 'f', -1, 64),
			msg.InputTokens+msg.InputCachedTokens+msg.OutputTokens,
			strconv.FormatFloat(in, 'f', -1, 64),
			msg.InputTokens,
			strconv.FormatFloat(inCached, 'f', -1, 64),
			msg.InputCachedTokens,
			strconv.FormatFloat(out, 'f', -1, 64),
			msg.OutputTokens)
	}

	date := layout.AddText(1, 0, 1, 1, "<small>"+Layout_ConvertTextDateTime(int64(msg.CreatedTimeSec)))
	date.Align_h = 2

	y := 1
	for content_i, it := range msg.Content {

		switch it.Type {
		case "text":
			txt := strings.TrimSpace(it.Text)
			if txt != "" {
				layout.SetRowFromSub(y, 1, 100)
				layout.AddTextMultiline(0, y, 2, 1, it.Text)
				y++
			}
		case "tool_use":

			switch it.Name {
			case "ui_slider":
				type ui_slider struct {
					Min   float64 //Minimum range
					Max   float64 //Maximum range
					Step  float64 //Step size
					Value float64 //Current value
				}
				var ui ui_slider
				js, _ := it.Input.MarshalJSON()
				_ = json.Unmarshal(js, &ui)

				gui := layout.AddSlider(0, y, 2, 1, &ui.Value, ui.Min, ui.Max, ui.Step)
				layout.FindLayout(0, y, 2, 1).Enable = (st.isRunning == nil || !st.isRunning())
				gui.changed = func() {
					args, _ := json.Marshal(ui)
					st.sendChange(content_i, string(args), it)
				}

			case "ui_map":
				//....

			default:
				y = st.toolUse(it, layout, y)
			}

		//case "tool_result":
		//	layout.SetRowFromSub(y, 1, 100)
		//	layout.AddTextMultiline(0, y, 2, 1, it.Content)
		//	y++

		case "image":
			//inside /disk - but don't save them into chat(anthropic messages) ....
			//every chat 45645123165.json has folder with same name ...
			//+pemanent delete folder as well ...

			//....
			/*if len(st.Files) > 0 {
				layout.SetRowFromSub(2, 1, 2)
				ImgsList := layout.AddLayoutList(0, 2, 2, 1, true)

				for _, file := range st.Files {
					ImgDia := layout.AddDialog("image_" + file)
					ImgDia.Layout.SetColumn(0, 5, 15)
					ImgDia.Layout.SetRow(0, 5, 15)
					ImgDia.Layout.AddImage(0, 0, 1, 1, file)

					imgLay := ImgsList.AddListSubItem()
					imgLay.SetColumn(0, 2, 2)
					imgLay.SetRow(0, 2, 2)
					imgBt := imgLay.AddButtonIcon(0, 0, 1, 1, file, 0, file)
					imgBt.Background = 0
					imgBt.Cd = Paint_GetPalette().B
					imgBt.Border = true
					imgBt.clicked = func() {
						ImgDia.OpenRelative(imgLay)
					}
				}
			}*/
			y++
		}
	}
}

func (st *ChatMsg) sendChange(content_i int, new_args string, it Anthropic_completion_msg_Content) {
	if st.isRunning != nil && st.isRunning() {
		return
	}

	// remove follow ups
	st.agent.Messages = st.agent.Messages[:st.msg_i+1]
	st.agent.Messages[st.msg_i].Content = st.agent.Messages[st.msg_i].Content[:content_i+1]

	// call
	answerJs := st.agent.callTool(it.Id, it.Name, new_args, nil)

	// save answer
	st.agent.AddCallResult(it.Name, it.Id, answerJs)

	// run
	if st.uiChanged != nil {
		st.uiChanged()
	}
}

func (st *ChatMsg) toolUse(it Anthropic_completion_msg_Content, layout *Layout, y int) int {

	// if open, highlight it
	if st.agent.Selected_sub_call_id == it.Id {
		layout.Back_cd = Paint_GetPalette().S
		//layout.Border_cd = Paint_GetPalette().OnB
	}

	args, err := it.Input.MarshalJSON()
	if err != nil {
		log.Fatal(err)
	}

	showParams := st.agent.IsShowToolParameters(it.Id)

	// open sub-agent
	ShowParamsBt := layout.AddButton(0, y, 1, 1, fmt.Sprintf("<i>%s(%s)", it.Name, it.Id))
	ShowParamsBt.Background = 0.5
	ShowParamsBt.Align = 0
	ShowParamsBt.Icon_align = 0
	ShowParamsBt.Icon_margin = 0.1
	// ShowParamsBt.Border = true
	if showParams {
		ShowParamsBt.Icon = "resources/arrow_down.png"
		ShowParamsBt.Tooltip = "Hide tool call's parameters"
	} else {
		ShowParamsBt.Icon = "resources/arrow_right.png"
		ShowParamsBt.Tooltip = "Show tool call's parameters"
	}
	ShowParamsBt.clicked = func() {
		st.agent.SetShowToolParameters(it.Id, !showParams)
	}

	if st.agent.FindSubAgent(it.Id) != nil {
		if st.agent.Selected_sub_call_id != it.Id {
			bt := layout.AddButton(1, y, 1, 1, "Open")
			bt.Background = 0.5
			bt.clicked = func() {
				st.agent.Selected_sub_call_id = it.Id
			}
		} else {
			bt := layout.AddButton(1, y, 1, 1, "Close")
			bt.Background = 0.5
			bt.clicked = func() {
				st.agent.Selected_sub_call_id = ""
			}
		}
	}
	y++

	// parameters
	if showParams {
		layout.SetRowFromSub(y, 1, 100)
		CallDiv := layout.AddLayout(0, y, 2, 1)
		CallDiv.SetColumn(0, 1, 1)
		CallDiv.SetColumnFromSub(1, 1, 10)
		CallDiv.SetColumn(2, 1, 100)
		//CallDiv.Border_cd = Paint_GetPalette().OnB
		CallDiv.Back_cd = Color_Aprox(Paint_GetPalette().P, Paint_GetPalette().B, 0.8)
		y++

		yy := 0
		CallDiv.AddText(0, yy, 3, 1, "Inputs:")
		yy++

		//render tool
		attrs := make(map[string]interface{})
		err = json.Unmarshal(args, &attrs)
		if err == nil {

			//sort map by key
			// Extract keys from map
			var keys []string
			for k := range attrs {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, key := range keys {
				val := attrs[key]
				CallDiv.SetRowFromSub(yy, 1, 100)
				/*tx :=*/ CallDiv.AddText(1, yy, 1, 1, "<i>"+key) //name
				//tx.Tooltip = //description from agent tool ....

				varStr := fmt.Sprintf("%v", val)
				if strings.Count(varStr, "\n") > 0 {
					CallDiv.AddTextMultiline(2, yy, 1, 1, varStr)
				} else {
					CallDiv.AddText(2, yy, 1, 1, varStr)
				}
				yy++
			}
		}

		result := st.agent.FindSubCallResultContent(it.Id)
		if result != nil {
			CallDiv.AddText(0, yy, 3, 1, "Output:")
			yy++
			CallDiv.SetRowFromSub(yy, 1, 100)
			CallDiv.AddTextMultiline(1, yy, 2, 1, result.Content)
			yy++
		}
	}

	return y
}
