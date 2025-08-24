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
	"image/color"
	"strings"
	"unicode/utf8"

	"github.com/go-audio/audio"
)

func (ui *Ui) _Text_getLineNumberWidth(value string, multi_line, line_wrapping, showLineNumbers bool, prop WinFontProps) (int, bool) {
	if !multi_line || line_wrapping {
		showLineNumbers = false
	}

	if showLineNumbers {
		mx, _ := ui.win.GetTextSizeMax(fmt.Sprintf("%d", 1+strings.Count(value, "\n")), 1000000000, prop)
		mx += 2 * ui.CellWidth(0.1) //margin
		return mx, true
	}
	return 0, false
}

func (ui *Ui) _Text_getCoord(coord OsV4, value string, multi_line, line_wrapping, showLineNumbers bool, prop WinFontProps) (OsV4, OsV4, bool) {
	var mx int
	mx, showLineNumbers = ui._Text_getLineNumberWidth(value, multi_line, line_wrapping, showLineNumbers, prop)

	coordLN := OsV4{} //line numbers
	coordText := coord

	if showLineNumbers {
		mx = OsClamp(mx, 0, coordText.Size.X)

		coordLN = coordText
		coordLN.Size.X = mx

		coordText.Start.X += mx
		coordText.Size.X -= mx
	}
	return coordLN, coordText, showLineNumbers
}

func (ui *Ui) _Text_drawHighlighLine(value string, highlight_lowerCase string, cd color.RGBA, align OsV2, coord OsV4, prop WinFontProps, yLine, num_lines int) {
	value_lowerCase := strings.ToLower(value)

	last_i := 0
	for {
		i := strings.Index(value_lowerCase[last_i:], highlight_lowerCase)
		if i < 0 {
			break
		}

		if _UiText_IsOutsideRGBA(last_i+i, value_lowerCase) {
			ui.GetWin().buff.AddTextBack(OsV2{last_i + i, last_i + i + len(highlight_lowerCase)}, value, prop, coord, cd, align, false, yLine, num_lines, ui.Cell())
		}

		last_i += i + 1
	}
}

func (ui *Ui) _Text_draw(layout *Layout, coord OsV4,
	value string, ghost string,
	prop WinFontProps,
	frontCd color.RGBA,
	align OsV2,
	selection, editable bool,
	multi_line, line_wrapping, password, showLineNumbers bool,
	highlight_text string) {

	var coordLN OsV4
	coordLN, coord, showLineNumbers = ui._Text_getCoord(coord, value, multi_line, line_wrapping, showLineNumbers, prop)

	edit := ui.edit

	if !edit.Is(layout) && password && value != "" {
		value = "**********"
	}

	if selection || editable {
		if edit.Is(layout) {
			value = edit.temp
		}

		if edit.Is(layout) && prop.switch_formating_when_edit {
			prop.formating = false //disable formating when edit active
		}
	}

	max_line_px := ui._UiText_getMaxLinePx(coord.Size.X, multi_line, line_wrapping)
	lines := ui.win.GetTextLines(value, max_line_px, prop)
	startY := ui.win.GetTextStart(value, prop, coord, align.X, align.Y, len(lines)).Y

	oldCursor := edit.end
	cursorPos := -1
	if edit.Is(layout) && selection && editable {
		cursorPos = edit.end
	}

	//draw selection
	cdSelection := Color_Aprox(ui.GetPalette().B, frontCd, 0.3)
	var range_sx, range_ex int
	if selection || editable {
		if edit.Is(layout) {
			range_sx = OsMin(edit.start, edit.end)
			range_ex = OsMax(edit.start, edit.end)

			_UiText_CheckSelectionExplode(value, &edit.start, &edit.end, &prop)
		}
	}

	//highlight(search)
	cdHighlight := Color_Aprox(ui.GetPalette().E, ui.GetPalette().GetGrey(0.5), 0.8)

	// draw
	if multi_line {

		lnColor := ui.GetPalette().GetGrey(0.5)

		//select
		var yst, yen int
		var curr_sy, curr_ey int
		if range_sx != range_ex {
			curr_sy = WinGph_CursorLineY(lines, range_sx)
			curr_ey = WinGph_CursorLineY(lines, range_ex)
			if curr_sy > curr_ey {
				curr_sy, curr_ey = curr_ey, curr_sy //swap
			}
			crop_sy, crop_ey := _UiText_GetLineYCrop(startY, len(lines), layout.view, prop) //only rows which are on screen
			yst = OsMax(curr_sy, crop_sy)
			yen = OsMin(curr_ey, crop_ey)
		}

		var highlight_text_lowerCase string
		if highlight_text != "" {
			highlight_text_lowerCase = strings.ToLower(highlight_text)
		}

		sy, ey := _UiText_GetLineYCrop(startY, len(lines), layout.view, prop) //only rows which are on screen
		for y := sy; y < ey; y++ {
			st, en := WinGph_PosLineRange(lines, y)
			ln := value[st:en]

			var rl_sx, rl_ex int
			if range_sx != range_ex && y >= yst && y <= yen { //equal
				rl_ex = len(ln)   //whole line
				if y == curr_sy { //first line
					_, rl_sx = WinGph_CursorLine(value, lines, range_sx)
				}
				if y == curr_ey { //last line
					_, rl_ex = WinGph_CursorLine(value, lines, range_ex)
				}

				ui.GetWin().buff.AddTextBack(OsV2{rl_sx, rl_ex}, ln, prop, coord, cdSelection, align, false, y, len(lines), layout.Cell())
			}

			if showLineNumbers {
				ui.GetWin().buff.AddText(fmt.Sprintf("%d", 1+y), prop, lnColor, coordLN, OsV2{0, 1}, y, len(lines))
			}

			if highlight_text != "" {
				ui._Text_drawHighlighLine(ln, highlight_text_lowerCase, cdHighlight, align, coord, prop, y, len(lines))
			}

			ui.GetWin().buff.AddText(ln, prop, frontCd, coord, align, y, len(lines)) //line
		}
	} else {

		ui.GetWin().buff.AddTextBack(OsV2{range_sx, range_ex}, value, prop, coord, cdSelection, align, false, 0, 1, layout.Cell())

		if highlight_text != "" {
			highlight_text_lowerCase := strings.ToLower(highlight_text)
			ui._Text_drawHighlighLine(value, highlight_text_lowerCase, cdHighlight, align, coord, prop, 0, 1)
		}

		ui.GetWin().buff.AddText(value, prop, frontCd, coord, align, 0, 1)
	}

	// draw cursor
	if cursorPos >= 0 {
		//cursor moved
		if edit.end != oldCursor {
			ui.GetWin().SetTextCursorMove()
		}

		currCd := ui.GetPalette().OnB
		if multi_line {
			y := WinGph_CursorLineY(lines, cursorPos)
			ln, ln_cursorPos := WinGph_CursorLine(value, lines, cursorPos)

			ui.GetWin().buff.AddTextCursor(ln, prop, coord, align, ln_cursorPos, y, len(lines), currCd, ui.Cell())
		} else {
			ui.GetWin().buff.AddTextCursor(value, prop, coord, align, cursorPos, 0, 1, currCd, ui.Cell())
		}
	}

	//ghost
	if ghost != "" {
		if (!edit.Is(layout) && value == "") || (edit.Is(layout) && edit.temp == "") {
			frontCd.A = 100
			layout.ui._Text_draw(layout, coord, ghost, "", prop, frontCd, OsV2{1, 1}, false, false, false, false, false, false, ui.settings.Highlight_text)
		}
	}

}

func (ui *Ui) _Text_update(layout *Layout,
	coord OsV4,
	margin [4]float64,
	value string,
	prop WinFontProps,
	align OsV2,
	selection, editable, tabIsChar bool,
	multi_line, line_wrapping, showLineNumbers bool) {

	_, coord, _ = ui._Text_getCoord(coord, value, multi_line, line_wrapping, showLineNumbers, prop)

	edit := layout.ui.edit
	keys := &ui.GetWin().io.Keys
	touch := &ui.GetWin().io.Touch

	oldCursor := edit.end

	orig_value := value
	if selection || editable {
		if edit.Is(layout) {
			value = edit.temp
		}
	}

	if editable && edit.Is(layout) {
		//if refresh {
		//	edit.Set(layout.UID, editable, orig_value, value, false, false, true, true, ui)
		//}
		edit.UpdateOrigValue(orig_value)
	}

	//wasActive := active

	max_line_px := ui._UiText_getMaxLinePx(coord.Size.X, multi_line, line_wrapping)
	lines := ui.win.GetTextLines(value, max_line_px, prop)
	startY := ui.win.GetTextStart(value, prop, coord, align.X, align.Y, len(lines)).Y

	if selection || editable {
		if edit.Is(layout) && prop.switch_formating_when_edit {
			prop.formating = false //disable formating when edit active
		}

		//touch
		if edit.IsActivateNext() || layout.IsOver() || layout.IsTouchActive() {

			var touchCursor int
			if multi_line {
				y := (ui.GetWin().io.Touch.Pos.Y - startY) / prop.lineH
				y = OsClamp(y, 0, len(lines)-1)

				st, en := WinGph_PosLineRange(lines, y)
				touchCursor = st + ui.win.GetTextPosCoord(ui.GetWin().io.Touch.Pos.X, value[st:en], prop, coord, align)
			} else {
				touchCursor = ui.win.GetTextPosCoord(ui.GetWin().io.Touch.Pos.X, value, prop, coord, align)
			}

			ui._UiText_Touch(layout, editable, orig_value, value, margin, lines, touchCursor, prop)
		}

		if edit.Is(layout) {

			if edit.KeySelectAll {
				keys.SelectAll = true
			}
			if edit.KeyCopy {
				keys.Copy = true
			}
			if edit.KeyCut {
				keys.Cut = true
			}
			if edit.KeyPaste {
				keys.Paste = true
			}
			if edit.KeyRecord {
				keys.RecordMic = true
			}

			ui._UiText_TextSelectKeys(layout, value, margin, lines, prop, multi_line, startY)

			if editable {

				drop_path := ui.GetWin().io.Touch.Drop_path
				if drop_path != "" && layout.IsTouchPosInside() {
					drop_path = "'" + drop_path + "'"

					touchCursor := ui.win.GetTextPosCoord(ui.GetWin().io.Touch.Pos.X, value, prop, coord, align)

					firstCur := OsTrn(edit.start < edit.end, edit.start, edit.end)
					lastCur := OsTrn(edit.start > edit.end, edit.start, edit.end)

					if touchCursor >= firstCur && touchCursor <= lastCur {
						//remove old
						if firstCur != lastCur {
							edit.temp = edit.temp[:firstCur] + edit.temp[lastCur:]
							lastCur = firstCur
						}

						//insert
						edit.temp = edit.temp[:firstCur] + drop_path + edit.temp[lastCur:]
						edit.start = firstCur
						edit.end = firstCur + len(drop_path)
					} else {
						//insert
						edit.temp = edit.temp[:touchCursor] + drop_path + edit.temp[touchCursor:]
						edit.start = touchCursor
						edit.end = touchCursor + len(drop_path)
					}
				}

				//old_value := value
				var tryMoveScroll bool
				value, tryMoveScroll = ui._UiText_Keys(layout, edit.temp, lines, tabIsChar, prop, multi_line, startY) //rewrite 'str' with temp value

				num_old_lines := len(lines)
				lines = ui.win.GetTextLines(value, max_line_px, prop) //refresh

				if num_old_lines != len(lines) {
					layout.ui.SetRelayoutHard()
				}

				if tryMoveScroll {
					ui._UiText_Text_VScrollInto(layout, lines, margin, edit.end, prop)
					ui._UiText_Text_HScrollInto(layout, value, margin, lines, edit.end, prop)
				}

				isTab := !tabIsChar && (keys.Tab && !keys.Ctrl)
				if isTab {
					edit.SetActivateTabNext()
					keys.Tab = false
				}

				if edit.end != oldCursor {
					ui.GetWin().SetTextCursorMove()
				}
			}

			//enter or Tab(key) or outside => save
			isOutside := (touch.Start && layout.CanTouch() && !layout.IsTouchPosInsideOrScroll() && edit.Is(layout) /*&& !keys.Ctrl*/)
			isEnter := keys.Enter && multi_line == keys.Ctrl
			isEsc := keys.Esc
			isTab := (!tabIsChar && keys.Tab && !keys.Ctrl)

			if isEsc && editable {
				edit.temp = value
			}

			if isTab || isEnter || isOutside || isEsc || edit.shortcut_triggered {
				//reset
				edit.Set(layout.UID, editable, orig_value, value, isEnter, true, !isEsc, false, ui)

				keys.Esc = false //don't close dialog
			}
		}
	}

	if edit.Is(layout) {
		edit.ResetShortcutKeys()
	}
}

func (ui *Ui) _UiText_Touch(layout *Layout, editable bool, orig_text string, text string, tx_margin [4]float64, lines []WinGphLine, cursor int, prop WinFontProps) {
	if !layout.CanTouch() {
		return
	}

	edit := ui.edit
	keys := &ui.GetWin().io.Keys
	touch := &ui.GetWin().io.Touch

	if touch.Rm && layout.IsTouchPosInside() && edit.Is(layout) {
		return
	}

	if !ui.touch.IsScrollOrResizeActive() && (!edit.Is(layout) && editable && edit.IsActivateNext()) {
		if edit.activate_next_uid != 0 {
			if edit.activate_next_uid == layout.UID {
				edit.Set(layout.UID, editable, orig_text, text, false, false, true, false, ui)
			}
		} else {
			//tab
			edit.Set(layout.UID, editable, orig_text, text, false, false, true, false, ui)
		}

	} else if layout.IsTouchPosInside() && layout.IsMouseButtonDownStart() {
		//click inside
		if !edit.Is(layout) {
			edit.Set(layout.UID, editable, orig_text, text, false, false, true, false, ui)
		}
		//set start-end
		edit.end = cursor
		if !edit.Is(layout) || !keys.Shift {
			//set start
			edit.start = cursor
		}

		switch touch.NumClicks {
		case 2:
			st, en := _UiText_CursorWordRange(text, cursor)
			edit.start = st //set start
			edit.end = en   //set end
		case 3:
			//paragraph
			edit.start = strings.LastIndex(text[:cursor], "\n")
			edit.end = strings.Index(text[cursor:], "\n")

			if edit.start < 0 {
				edit.start = 0
			}
			if edit.end < 0 {
				edit.end = len(text)
			} else {
				edit.end += cursor
			}
		}
	}

	//keep selecting
	if edit.Is(layout) && layout.IsTouchActive() && (touch.NumClicks != 2 && touch.NumClicks != 3) {
		edit.end = cursor //set end

		//scroll
		ui._UiText_Text_VScrollInto(layout, lines, tx_margin, cursor, prop)
		ui._UiText_Text_HScrollInto(layout, text, tx_margin, lines, cursor, prop)
	}
}

func (ui *Ui) _UiText_Keys(layout *Layout, text string, lines []WinGphLine, tabIsChar bool, prop WinFontProps, multi_line bool, startY int) (string, bool) {
	edit := layout.ui.edit
	keys := &ui.GetWin().io.Keys

	shiftKey := keys.Shift

	s := &edit.start
	e := &edit.end
	old := *e

	firstCur := OsTrn(*s < *e, *s, *e)
	lastCur := OsTrn(*s > *e, *s, *e)

	//cut/paste(copy() is in selectKeys)
	if keys.Cut {

		if firstCur == lastCur {
			firstCur, lastCur = WinGph_CursorLineRange(lines, firstCur) //select whole line
		}

		//remove
		text = text[:firstCur] + text[lastCur:]
		edit.temp = text

		//select
		*s = firstCur
		*e = firstCur
	} else if keys.Paste {
		//remove old selection
		if *s != *e {
			text = text[:firstCur] + text[lastCur:]
		}

		//insert
		cb := ui.GetWin().GetClipboardText()
		text = text[:firstCur] + cb + text[firstCur:]
		edit.temp = text

		firstCur += len(cb)
		*s = firstCur
		*e = firstCur
	} else if keys.RecordMic {
		if layout.ui.router.services.mic.Find(layout.UID) == nil {
			layout.ui.router.services.mic.FinishAll(false) //finish all previous

			mic, _ := layout.ui.router.services.mic.Start(layout.UID) //err ....

			mic.fnFinished = func(buff *audio.IntBuffer) {
				transcript, err := layout.ui.router.services.llms.TranscribeBuff(buff, "wav", "text")
				if err == nil {

					transcript = strings.TrimSpace(transcript)

					//remove old selection
					if *s != *e {
						text = text[:firstCur] + text[lastCur:]
					}
					//insert
					text = text[:firstCur] + transcript + text[firstCur:]
					edit.temp = text

					firstCur += len(transcript)
					*s = firstCur
					*e = firstCur
				}
			}
		} else {
			layout.ui.router.services.mic.Finished(layout.UID, false)
		}

		layout.ui.SetRefresh()
	}

	//when dialog is active, don't edit
	if !layout.CanTouch() {
		return edit.temp, old != *e
	}

	//insert text
	txt := keys.Text
	if tabIsChar && (keys.Tab && !keys.Ctrl) {
		txt += "\t"
	}

	if keys.Enter && multi_line && !keys.Ctrl {
		txt = "\n"
	}

	if len(txt) > 0 {
		//remove old selection
		if *s != *e {
			text = text[:firstCur] + text[lastCur:]
			*e = *s
		}

		//insert
		text = text[:firstCur] + txt + text[firstCur:]
		edit.temp = text

		//cursor
		firstCur += len(txt)
		*s = firstCur
		*e = firstCur

		//reset
		keys.Text = ""
	}

	//delete/backspace
	if *s != *e {
		if keys.Delete || keys.Backspace {

			//remove
			text = text[:firstCur] + text[lastCur:]
			edit.temp = text

			//cursor
			*s = firstCur
			*e = firstCur
		}
	} else {
		if keys.Backspace {
			//remove
			if *s > 0 {
				//removes one letter
				p := _UiText_CursorMoveLR(text, firstCur, -1, prop)
				text = text[:p] + text[firstCur:]
				edit.temp = text

				//cursor
				firstCur = p
				*s = firstCur
				*e = firstCur
			}
		} else if keys.Delete {
			//remove
			if *s < len(text) {
				//removes one letter
				p := _UiText_CursorMoveLR(text, firstCur, +1, prop)
				text = text[:firstCur] + text[p:]
				edit.temp = text
			}
		}
	}

	if !shiftKey {
		//arrows
		if *s != *e {
			if multi_line {
				if keys.ArrowU {
					p := _UiText_CursorMoveU(text, lines, *e, ui.win, prop)
					*s = p
					*e = p
				}
				if keys.ArrowD {
					p := _UiText_CursorMoveD(text, lines, *e, ui.win, prop)
					*s = p
					*e = p
				}
				if keys.PageU || keys.PageD {
					sy, ey := _UiText_GetLineYCrop(startY, len(lines), layout.view, prop)
					page_n := OsMax(1, ey-sy-2)

					var p int
					if keys.PageU {
						p = lines[OsMax(0, WinGph_CursorLineY(lines, *e)-page_n)].s
					} else {
						p = lines[OsMin(len(lines)-1, WinGph_CursorLineY(lines, *e)+page_n)].s
					}
					*s = p
					*e = p
				}
			}

			if keys.ArrowL {
				//from select -> single start
				*s = firstCur
				*e = firstCur
			} else if keys.ArrowR {
				//from select -> single end
				*s = lastCur
				*e = lastCur
			}
		} else {
			if keys.Ctrl {
				if keys.ArrowL {
					p := _UiText_CursorMoveLR(text, *s, -1, prop)
					first, _ := _UiText_CursorWordRange(text, p)
					if first == p && p > 0 {
						first, _ = _UiText_CursorWordRange(text, p-1)
					}
					*s = first
					*e = first
				}
				if keys.ArrowR {
					p := _UiText_CursorMoveLR(text, *s, +1, prop)
					_, last := _UiText_CursorWordRange(text, p)
					if last == p && p+1 < len(text) {
						_, last = _UiText_CursorWordRange(text, p+1)
					}
					*s = last
					*e = last
				}
			} else {
				if multi_line {
					if keys.ArrowU {
						p := _UiText_CursorMoveU(text, lines, *e, ui.win, prop)
						*s = p
						*e = p
					}
					if keys.ArrowD {
						p := _UiText_CursorMoveD(text, lines, *e, ui.win, prop)
						*s = p
						*e = p
					}

					if keys.PageU || keys.PageD {
						sy, ey := _UiText_GetLineYCrop(startY, len(lines), layout.view, prop)
						page_n := OsMax(1, ey-sy-2)

						var p int
						if keys.PageU {
							p = lines[OsMax(0, WinGph_CursorLineY(lines, *e)-page_n)].s
						} else {
							p = lines[OsMin(len(lines)-1, WinGph_CursorLineY(lines, *e)+page_n)].s
						}
						*s = p
						*e = p
					}

				}

				if keys.ArrowL {
					p := _UiText_CursorMoveLR(text, *s, -1, prop)
					*s = p
					*e = p
				} else if keys.ArrowR {
					p := _UiText_CursorMoveLR(text, *s, +1, prop)
					*s = p
					*e = p
				}
			}
		}

		//home/end
		if keys.Home {
			if multi_line {
				firstCur, _ = WinGph_CursorLineRange(lines, *e) //line start
			} else {
				firstCur = 0
			}
			*s = firstCur
			*e = firstCur
		} else if keys.End {
			if multi_line {
				_, firstCur = WinGph_CursorLineRange(lines, *e) //line start
			} else {
				firstCur = len(text)
			}

			*s = firstCur
			*e = firstCur
		}
	}

	//history
	{
		his := UiTextHistoryItem{str: text, cur: *e}

		layout.ui.edit_history.FindOrAdd(edit.uid, his).AddWithTimeOut(his)

		if keys.TextBackward {
			his = layout.ui.edit_history.FindOrAdd(edit.uid, his).Backward(his)
			edit.temp = his.str
			*s = his.cur
			*e = his.cur
		}
		if keys.TextForward {
			his = layout.ui.edit_history.FindOrAdd(edit.uid, his).Forward()
			edit.temp = his.str
			*s = his.cur
			*e = his.cur
		}
	}

	return edit.temp, old != *e
}

func _UiText_CursorWordRange(text string, cursor int) (int, int) {
	start := 0
	end := 0

	text = strings.ToLower(text)

	for p, ch := range text {
		chSz := len(string(ch))

		if OsIsTextWord(ch) {
			end = p + chSz
		} else {
			if p < cursor {
				start = p + chSz
			} else {
				break
			}
		}
	}
	if end < start {
		end = start
	}

	return start, end
}

func (ui *Ui) _UiText_TextSelectKeys(layout *Layout, text string, tx_margin [4]float64, lines []WinGphLine, prop WinFontProps, multi_line bool, startY int) {
	keys := &ui.GetWin().io.Keys
	edit := layout.ui.edit

	s := &edit.start
	e := &edit.end

	//check
	*s = OsClamp(*s, 0, len(text))
	*e = OsClamp(*e, 0, len(text))

	old := *e

	//select all
	if keys.SelectAll {
		*s = 0
		*e = len(text)
	}

	//copy, cut
	if keys.Copy || keys.Cut {
		firstCur := OsTrn(*s < *e, *s, *e)
		lastCur := OsTrn(*s > *e, *s, *e)

		if firstCur == lastCur {
			firstCur, lastCur = WinGph_CursorLineRange(lines, firstCur) //select whole line
		}

		if keys.Shift {
			ui.GetWin().SetClipboardText(_UiText_RemoveFormating(text[firstCur:lastCur]))
		} else {

			ui.GetWin().SetClipboardText(text[firstCur:lastCur])
		}
	}

	//shift
	if keys.Shift {
		//ctrl
		if keys.Ctrl {
			if keys.ArrowL {
				p := _UiText_CursorMoveLR(text, *e, -1, prop)
				first, _ := _UiText_CursorWordRange(text, p)
				if first == p && p > 0 {
					first, _ = _UiText_CursorWordRange(text, p-1)
				}
				*e = first
			}
			if keys.ArrowR {
				p := _UiText_CursorMoveLR(text, *e, +1, prop)
				_, last := _UiText_CursorWordRange(text, p)
				if last == p && p+1 < len(text) {
					_, last = _UiText_CursorWordRange(text, p+1)
				}
				*e = last
			}
		} else {
			if multi_line {
				if keys.ArrowU {
					*e = _UiText_CursorMoveU(text, lines, *e, ui.win, prop)
				}
				if keys.ArrowD {
					*e = _UiText_CursorMoveD(text, lines, *e, ui.win, prop)
				}

				if keys.PageU || keys.PageD {
					sy, ey := _UiText_GetLineYCrop(startY, len(lines), layout.view, prop)
					page_n := OsMax(1, ey-sy-2)

					if keys.PageU {
						*e = lines[OsMax(0, WinGph_CursorLineY(lines, *e)-page_n)].s
					} else {
						*e = lines[OsMin(len(lines)-1, WinGph_CursorLineY(lines, *e)+page_n)].e
					}
				}
			}

			if keys.ArrowL {
				p := _UiText_CursorMoveLR(text, *e, -1, prop)
				*e = p
			}
			if keys.ArrowR {
				p := _UiText_CursorMoveLR(text, *e, +1, prop)
				*e = p
			}
		}

		//home & end
		if keys.Home {
			*e, _ = WinGph_CursorLineRange(lines, *e) //line start
		}
		if keys.End {
			_, *e = WinGph_CursorLineRange(lines, *e) //line end
		}
	}

	//scroll
	newPos := *e
	if old != newPos {
		ui._UiText_Text_VScrollInto(layout, lines, tx_margin, newPos, prop)
	}
	if old != newPos {
		ui._UiText_Text_HScrollInto(layout, text, tx_margin, lines, newPos, prop)
	}
}

func (ui *Ui) _UiText_Text_VScrollInto(layout *Layout, lines []WinGphLine, tx_margin [4]float64, cursor int, prop WinFontProps) {

	margin_t := layout.ui.CellWidth(tx_margin[0])
	margin_b := layout.ui.CellWidth(tx_margin[1])

	v_pos := WinGph_CursorLineY(lines, cursor)*prop.lineH + margin_t //- ui.CellWidth(2*0.1)

	extra_space := margin_b

	v_st := layout.scrollV.GetWheel()
	v_sz := layout.view.Size.Y - prop.lineH
	v_en := v_st + v_sz

	backup_wheel := layout.scrollV.wheel
	if v_pos-extra_space <= v_st {
		layout.scrollV.SetWheel(OsMax(0, v_pos) - extra_space)
	} else if v_pos+extra_space >= v_en {
		layout.scrollV.wheel = OsMax(0, v_pos-v_sz) + extra_space //SetWheel() has boundary check, which is not good here
	}

	if backup_wheel != layout.scrollV.wheel {
		//dom.RebuildSoft()
		layout.GetSettings().SetScrollV(layout.UID, layout.scrollV.wheel)
		layout.ui.SetRelayoutSoft()
	}

}

func (ui *Ui) _UiText_Text_HScrollInto(layout *Layout, text string, tx_margin [4]float64, lines []WinGphLine, cursor int, prop WinFontProps) {
	margin_l := layout.ui.CellWidth(tx_margin[2])
	margin_r := layout.ui.CellWidth(tx_margin[3])

	ln, curr := WinGph_CursorLine(text, lines, cursor)
	h_pos := ui.win.GetTextSize(curr, ln, prop).X + margin_l

	cursor_space := WinPaintBuff_GetCursorWidth(layout.ui.Cell())
	extra_space := margin_r //set 0 for cursor being exactly at the border(side)

	h_st := layout.scrollH.GetWheel()
	h_sz := layout.view.Size.X
	h_en := h_st + h_sz

	backup_wheel := layout.scrollH.wheel
	if h_pos-extra_space <= h_st {
		layout.scrollH.SetWheel(OsMax(0, h_pos) - extra_space)
	} else if h_pos+cursor_space+extra_space >= h_en {
		//SetWheel() has boundary check, which is not good here
		layout.scrollH.wheel = OsMax(0, h_pos-h_sz) + cursor_space + extra_space
	}

	if backup_wheel != layout.scrollH.wheel {
		layout.GetSettings().SetScrollH(layout.UID, layout.scrollH.wheel)
		layout.ui.SetRelayoutSoft()
	}
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

func _UiText_RemoveFormating(str string) string {
	str = strings.ReplaceAll(str, "<b>", "")
	str = strings.ReplaceAll(str, "</b>", "")

	str = strings.ReplaceAll(str, "<i>", "")
	str = strings.ReplaceAll(str, "</i>", "")

	str = strings.ReplaceAll(str, "<h1>", "")
	str = strings.ReplaceAll(str, "</h1>", "")

	str = strings.ReplaceAll(str, "<h2>", "")
	str = strings.ReplaceAll(str, "</h2>", "")

	str = strings.ReplaceAll(str, "<small>", "")
	str = strings.ReplaceAll(str, "</small>", "")

	str = _UiText_RemoveFormatingRGBA(str)

	return str
}

func _UiText_IsOutsideRGBA(cur int, str string) bool {
	if strings.LastIndex(str[:cur], "</rgba") >= strings.LastIndex(str[:cur], "<rgba") {
		return true
	}

	return false
}

func (ui *Ui) _UiText_getMaxLinePx(max_px int, multi_line, line_wrapping bool) int {
	if max_px < 0 {
		max_px = -1
	}

	max_line_px := -1
	if multi_line && line_wrapping {
		max_line_px = max_px
	}

	return max_line_px
}

func _UiText_HashFormatingPreSuf_fix(str string, startWith bool) int {

	var fn func(a, b string) bool
	if startWith {
		fn = strings.HasPrefix
	} else {
		fn = strings.HasSuffix
	}

	if fn(str, "<b>") || fn(str, "<i>") {
		return 3
	}
	if fn(str, "</b>") || fn(str, "</i>") {
		return 3 + 1
	}

	if fn(str, "<h1>") || fn(str, "<h2>") {
		return 4
	}
	if fn(str, "</h1>") || fn(str, "</h2>") {
		return 4 + 1
	}

	if fn(str, "<small>") {
		return 7
	}
	if fn(str, "</small>") {
		return 8
	}

	if fn(str, "</rgba>") {
		return 7
	}

	if startWith {
		if strings.HasPrefix(str, "<rgba") {
			return strings.IndexByte(str, '>') + 1
		}
	} else {
		if strings.HasSuffix(str, ">") {
			d := strings.LastIndex(str, "<rgba")
			if d >= 0 {
				return len(str) - d
			}
		}
	}

	return 0
}

func _UiText_CheckSelectionExplode(str string, start *int, end *int, prop *WinFontProps) {
	if !prop.formating {
		return
	}

	if *start < *end {
		*start -= _UiText_HashFormatingPreSuf_fix(str[:*start], false)
		*end += _UiText_HashFormatingPreSuf_fix(str[*end:], true)
	}
	if *end < *start {
		*end -= _UiText_HashFormatingPreSuf_fix(str[:*end], false)
		*start += _UiText_HashFormatingPreSuf_fix(str[*start:], true)
	}
}

func _UiText_GetLineYCrop(startY int, num_lines int, crop OsV4, prop WinFontProps) (int, int) {

	sy := (crop.Start.Y - startY) / prop.lineH
	ey := OsRoundUp(float64(crop.End().Y-startY) / float64(prop.lineH))

	//check
	sy = OsClamp(sy, 0, num_lines-1)
	ey = OsClamp(ey, 0, num_lines)

	return sy, ey
}

func _UiText_CursorMoveLR(text string, cursor int, move int, prop WinFontProps) int {

	//skip formating
	if prop.formating {
		if move < 0 { //left
			cursor -= _UiText_HashFormatingPreSuf_fix(text[:cursor], false)
		}

		if move > 0 { //right
			cursor += _UiText_HashFormatingPreSuf_fix(text[cursor:], true)
		}
	}

	//shift rune
	if move < 0 { //left
		_, l := utf8.DecodeLastRuneInString(text[:cursor])
		cursor -= l
	}
	if move > 0 { //right
		_, l := utf8.DecodeRuneInString(text[cursor:])
		cursor += l
	}

	//check
	cursor = OsClamp(cursor, 0, len(text))

	return cursor
}

func _UiText_CursorMoveU(text string, lines []WinGphLine, cursor int, win *Win, prop WinFontProps) int {
	y := WinGph_CursorLineY(lines, cursor)
	if y > 0 {
		curr_ln_text, old_pos := WinGph_CursorLine(text, lines, cursor)
		curr_px := win.GetTextSize(old_pos, curr_ln_text, prop).X

		st, en := WinGph_PosLineRange(lines, y-1) //up line
		new_pos := win.GetTextPos(curr_px, text[st:en], prop, true)

		cursor = st + OsMin(new_pos, en-st)
	}
	return cursor
}
func _UiText_CursorMoveD(text string, lines []WinGphLine, cursor int, win *Win, prop WinFontProps) int {
	y := WinGph_CursorLineY(lines, cursor)
	if y+1 < len(lines) {
		curr_ln_text, old_pos := WinGph_CursorLine(text, lines, cursor)
		curr_px := win.GetTextSize(old_pos, curr_ln_text, prop).X

		st, en := WinGph_PosLineRange(lines, y+1) //down line
		new_pos := win.GetTextPos(curr_px, text[st:en], prop, true)

		cursor = st + OsMin(new_pos, en-st)
	}
	return cursor
}
