package main

import (
	"encoding/base64"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// [ignore]
type ShowChat struct {
	ChatFileName string
}

func (st *ShowChat) run(caller *ToolCaller, ui *UI) error {

	source_chat, err := NewChat(filepath.Join("Chats", st.ChatFileName))
	if err != nil {
		return err
	}
	source_root, err := NewChats("")
	if err != nil {
		return err
	}

	isRunning := (caller.callFuncFindMsgName(source_chat.GetChatID()) != nil) //(st.isRunning != nil && st.isRunning())

	ui.SetColumn(0, 1, Layout_MAX_SIZE)
	ui.SetColumn(1, 5, 20)
	ui.SetColumn(2, 1, Layout_MAX_SIZE)

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

		ui.SetRowFromSub(y, 1, Layout_MAX_SIZE, true)
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

		ui.SetRowFromSub(y, 1, Layout_MAX_SIZE, true)
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
			return source_chat._sendIt(caller, source_root, true)
		}
		y++

	}

	//Statistics - total
	if y >= 2 { //1st message is user
		ui.SetRowFromSub(y, 1, 2, true)

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
	layout.SetColumn(2, 1, Layout_MAX_SIZE)

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
					layout.SetRowFromSub(y, 1, Layout_MAX_SIZE, true)

					UserDiv := layout.AddLayout(0, y, 3, 1)
					UserDiv.SetColumn(0, 0, Layout_MAX_SIZE)
					UserDiv.SetColumnFromSub(1, 1, Layout_MAX_SIZE, true)
					UserDiv.SetRowFromSub(0, 1, Layout_MAX_SIZE, true)
					y++

					tx := UserDiv.AddText(1, 0, 1, 1, txt)
					tx.setMultilined()
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
					ImgDia.UI.SetColumn(0, 5, 15)
					ImgDia.UI.SetRow(0, 5, 15)
					ImgDia.UI.AddMediaBlob(0, 0, 1, 1, imgBlob)

					imgLay := ImgsCards.AddItem()
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
				layout.SetRowFromSub(y, 1, Layout_MAX_SIZE, true)
				tx := layout.AddText(0, y, 3, 1, txt)
				tx.setMultilined()
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
				tx.setMultilined()
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
		DivIcons.SetColumn(x, 0, Layout_MAX_SIZE)
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

					chat._sendIt(caller, root, true)

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
				info := InfoDia.UI.AddText(0, 0, 1, 1, inf)
				info.Align_v = 0
				info.setMultilined()

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
}
