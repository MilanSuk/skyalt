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
	"image/color"
	"strings"
	"unicode/utf8"
)

func (ui *Ui) _Text_draw(dom *Layout3, coord OsV4,
	value string, ghost string,
	prop WinFontProps,
	frontCd color.RGBA,
	align OsV2,
	selection, editable bool,
	multi_line, line_wrapping bool) {

	borderCoord := coord
	if editable {
		coord = coord.AddSpace(ui.CellWidth(0.1))
	}

	edit := &ui.parent.edit

	if selection || editable {
		if edit.Is(dom) {
			value = edit.temp
		}

		if edit.Is(dom) && prop.switch_formating_when_edit {
			prop.formating = false //disable formating when edit active
		}
	}

	/*if drawBackWhenNonEmpty && value != "" {
		ui.GetWin().buff.AddRect(coord, dom.ui.GetPalette().S, 0)
	}*/

	max_line_px := ui._Paint_getMaxLinePx(coord, multi_line, line_wrapping)
	lines := ui.GetTextLines(value, max_line_px, prop)
	startY := ui.GetTextStart(value, prop, coord, align.X, align.Y, len(lines)).Y

	oldCursor := edit.end
	cursorPos := -1
	if edit.Is(dom) && selection && editable {
		cursorPos = edit.end
	}

	//draw selection
	cdSelection := ui.GetPalette().GetGrey(0.5)
	var range_sx, range_ex int
	if selection || editable {
		if edit.Is(dom) {
			range_sx = OsMin(edit.start, edit.end)
			range_ex = OsMax(edit.start, edit.end)

			_UiPaint_CheckSelectionExplode(value, &edit.start, &edit.end, &prop)
		}
	}

	// draw
	if multi_line {
		//select
		var yst, yen int
		var curr_sy, curr_ey int
		if range_sx != range_ex {
			curr_sy = WinGph_CursorLineY(lines, range_sx)
			curr_ey = WinGph_CursorLineY(lines, range_ex)
			if curr_sy > curr_ey {
				curr_sy, curr_ey = curr_ey, curr_sy //swap
			}
			crop_sy, crop_ey := _UiPaint_GetLineYCrop(startY, len(lines), dom.view, prop) //only rows which are on screen
			yst = OsMax(curr_sy, crop_sy)
			yen = OsMin(curr_ey, crop_ey)
		}

		sy, ey := _UiPaint_GetLineYCrop(startY, len(lines), dom.view, prop) //only rows which are on screen
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
			}

			ui.GetWin().buff.AddTextBack(OsV2{rl_sx, rl_ex}, ln, prop, coord, cdSelection, align, false, y, len(lines))

			ui.GetWin().buff.AddText(ln, prop, frontCd, coord, align, y, len(lines))
		}
	} else {
		ui.GetWin().buff.AddTextBack(OsV2{range_sx, range_ex}, value, prop, coord, cdSelection, align, false, 0, 1)

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
		if (!edit.Is(dom) && value == "") || (edit.Is(dom) && edit.temp == "") {
			frontCd.A = 100
			dom.ui._Text_draw(dom, coord, ghost, "", prop, frontCd, OsV2{1, 1}, false, false, false, false)
		}
	}

	//draw border
	if editable {
		width := 0.03
		if edit.Is(dom) {
			width *= 2
		}
		ui.GetWin().buff.AddRect(borderCoord, dom.ui.GetPalette().P, dom.ui.CellWidth(width))
	}

}

func (ui *Ui) _Text_update(dom *Layout3,
	coord OsV4,
	value string,
	prop WinFontProps,
	align OsV2,
	selection, editable, tabIsChar bool,
	multi_line, multi_line_enter_finish, line_wrapping bool) {

	edit := &dom.GetUis().edit
	keys := &ui.GetWin().io.Keys
	touch := &ui.GetWin().io.Touch

	oldCursor := edit.end

	orig_value := value
	if selection || editable {
		if edit.Is(dom) {
			value = edit.temp
		}
	}

	//wasActive := active

	max_line_px := ui._Paint_getMaxLinePx(coord, multi_line, line_wrapping)
	lines := ui.GetTextLines(value, max_line_px, prop)
	startY := ui.GetTextStart(value, prop, coord, align.X, align.Y, len(lines)).Y

	if selection || editable {
		if edit.Is(dom) && prop.switch_formating_when_edit {
			prop.formating = false //disable formating when edit active
		}

		//touch
		if edit.IsActivateNext() || dom.IsOver() || dom.IsTouchActive() {

			var touchCursor int
			if multi_line {
				y := (ui.GetWin().io.Touch.Pos.Y - startY) / prop.lineH
				y = OsClamp(y, 0, len(lines)-1)

				st, en := WinGph_PosLineRange(lines, y)
				touchCursor = st + ui.GetTextPos(ui.GetWin().io.Touch.Pos.X, value[st:en], prop, coord, align)
			} else {
				touchCursor = ui.GetTextPos(ui.GetWin().io.Touch.Pos.X, value, prop, coord, align)
			}

			ui._UiPaint_TextSelectTouch(dom, editable, orig_value, value, lines, touchCursor, prop)
		}

		if edit.Is(dom) {

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

			ui._UiPaint_TextSelectKeys(dom, value, lines, prop, multi_line)

			if editable {

				drop_path := ui.GetWin().io.Touch.Drop_path
				if drop_path != "" && dom.IsTouchPosInside() {
					firstCur := OsTrn(edit.start < edit.end, edit.start, edit.end)
					lastCur := OsTrn(edit.start > edit.end, edit.start, edit.end)
					edit.temp = edit.temp[:firstCur] + drop_path + edit.temp[lastCur:]
					edit.start = firstCur
					edit.end = firstCur + len(drop_path)
				}

				//old_value := value
				var tryMoveScroll bool
				value, tryMoveScroll = ui._UiPaint_TextEditKeys(dom, edit.temp, lines, tabIsChar, prop, multi_line, multi_line_enter_finish) //rewrite 'str' with temp value

				/*if old_value != value && ed.RefreshDelaySec > 0 {
					edit.SetRefreshTicks()
				}*/

				num_old_lines := len(lines)
				lines = ui.GetTextLines(value, max_line_px, prop) //refresh

				if num_old_lines != len(lines) {
					//if dom.parent.needRebuild(dom) {
					dom.ui.SetRelayout()
					//}
				}

				if tryMoveScroll {
					ui._UiPaint_Text_VScrollInto(dom, lines, edit.end, prop)
					ui._UiPaint_Text_HScrollInto(dom, value, lines, edit.end, prop)
				}

				isTab := !tabIsChar && keys.Tab
				if isTab {
					edit.SetActivateNext()
					keys.Tab = false
					//edit.tab = true //edit
				}

				if edit.end != oldCursor {
					ui.GetWin().SetTextCursorMove()
				}
			}

			//enter or Tab(key) or outside => save
			isOutside := (touch.Start && dom.CanTouch() && !dom.IsTouchPosInside() && edit.Is(dom))
			isEnter := keys.Enter && (!multi_line || (multi_line_enter_finish && !keys.Ctrl) || (!multi_line_enter_finish && keys.Ctrl))
			isEsc := keys.Esc
			isTab := !tabIsChar && keys.Tab

			if isEsc && editable {
				edit.temp = value
			}

			if isTab || isEnter || isOutside || isEsc {
				//reset
				edit.Set(dom, editable, orig_value, value, isEnter, true)
			}
		}
	}

	if edit.Is(dom) {
		edit.ResetKeys()
	}
}

func (ui *Ui) _UiPaint_TextSelectTouch(dom *Layout3, editable bool, orig_text string, text string, lines []WinGphLine, cursor int, prop WinFontProps) {

	if !dom.CanTouch() {
		return
	}

	edit := &dom.GetUis().edit
	keys := &ui.GetWin().io.Keys
	touch := &ui.GetWin().io.Touch

	//if touch.Rm && dom.IsTouchPosInside() && edit.Is(dom) && cursor >= OsMin(edit.start, edit.end) && cursor < OsMax(edit.start, edit.end) {
	if touch.Rm && dom.IsTouchPosInside() && edit.Is(dom) {
		return
	}

	if !dom.GetUis().touch.IsScrollOrResizeActive() && (!edit.Is(dom) && editable && edit.IsActivateNext()) /* || (editable && edit.tab) || activate_next_uid)*/ {
		edit.Set(dom, editable, orig_text, text, false, false)

	} else if dom.IsTouchPosInside() && dom.IsMouseButtonDownStart() {

		//click inside
		if !edit.Is(dom) {
			edit.Set(dom, editable, orig_text, text, false, false)
		}
		//set start-end
		edit.end = cursor
		if !edit.Is(dom) || !keys.Shift {
			//set start
			edit.start = cursor
		}

		switch touch.NumClicks {
		case 2:
			st, en := _UiPaint_CursorWordRange(text, cursor)
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
	if edit.Is(dom) && dom.IsTouchActive() && (touch.NumClicks != 2 && touch.NumClicks != 3) {
		edit.end = cursor //set end

		//scroll
		ui._UiPaint_Text_VScrollInto(dom, lines, cursor, prop)
		ui._UiPaint_Text_HScrollInto(dom, text, lines, cursor, prop)

		//root.buff.ResetHost() //SetNoSleep()
	}
}

func (ui *Ui) _UiPaint_TextEditKeys(dom *Layout3, text string, lines []WinGphLine, tabIsChar bool, prop WinFontProps, multi_line bool, multi_line_enter_finish bool) (string, bool) {
	edit := &dom.GetUis().edit
	keys := &ui.GetWin().io.Keys

	shiftKey := keys.Shift

	//uid := edit.uid

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
	}

	//when dialog is active, don't edit
	//lv := ui.GetCall()
	if !dom.CanTouch() {
		return edit.temp, old != *e
	}

	//insert text
	txt := keys.Text
	if tabIsChar && keys.Tab {
		txt += "\t"
	}

	if keys.Enter && multi_line && ((multi_line_enter_finish && keys.Ctrl) || (!multi_line_enter_finish && !keys.Ctrl)) {
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
				p := _UiPaint_CursorMoveLR(text, firstCur, -1, prop)
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
				p := _UiPaint_CursorMoveLR(text, firstCur, +1, prop)
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
					firstCur = _UiPaint_CursorMoveU(text, lines, *e)
					*s = firstCur
					*e = firstCur
				}
				if keys.ArrowD {
					firstCur = _UiPaint_CursorMoveD(text, lines, *e)
					*s = firstCur
					*e = firstCur
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
					p := _UiPaint_CursorMoveLR(text, *s, -1, prop)
					first, _ := _UiPaint_CursorWordRange(text, p)
					if first == p && p > 0 {
						first, _ = _UiPaint_CursorWordRange(text, p-1)
					}
					*s = first
					*e = first
				}
				if keys.ArrowR {
					p := _UiPaint_CursorMoveLR(text, *s, +1, prop)
					_, last := _UiPaint_CursorWordRange(text, p)
					if last == p && p+1 < len(text) {
						_, last = _UiPaint_CursorWordRange(text, p+1)
					}
					*s = last
					*e = last
				}
			} else {
				if multi_line {
					if keys.ArrowU {
						p := _UiPaint_CursorMoveU(text, lines, *e)
						*s = p
						*e = p
					}
					if keys.ArrowD {
						p := _UiPaint_CursorMoveD(text, lines, *e)
						*s = p
						*e = p
					}
				}

				if keys.ArrowL {
					p := _UiPaint_CursorMoveLR(text, *s, -1, prop)
					*s = p
					*e = p
				} else if keys.ArrowR {
					p := _UiPaint_CursorMoveLR(text, *s, +1, prop)
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
		his := UiPaintTextHistoryItem{str: text, cur: *e}

		dom.GetUis().edit_history.FindOrAdd(edit.hash, his).AddWithTimeOut(his)

		if keys.Backward {
			his = dom.GetUis().edit_history.FindOrAdd(edit.hash, his).Backward(his)
			edit.temp = his.str
			*s = his.cur
			*e = his.cur
		}
		if keys.Forward {
			his = dom.GetUis().edit_history.FindOrAdd(edit.hash, his).Forward()
			edit.temp = his.str
			*s = his.cur
			*e = his.cur
		}
	}

	return edit.temp, old != *e
}

func _UiPaint_CursorWordRange(text string, cursor int) (int, int) {
	start := 0
	end := 0

	text = strings.ToLower(text)

	for p, ch := range text {
		if OsIsTextWord(ch) {
			end = p + 1
		} else {
			if p < cursor {
				start = p + 1
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

func (ui *Ui) _UiPaint_TextSelectKeys(dom *Layout3, text string, lines []WinGphLine, prop WinFontProps, multi_line bool) {

	//touch := &ui.GetWin().io.Touch
	keys := &ui.GetWin().io.Keys
	edit := &dom.GetUis().edit

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
			ui.GetWin().SetClipboardText(_UiPaint_RemoveFormating(text[firstCur:lastCur]))
		} else {

			ui.GetWin().SetClipboardText(text[firstCur:lastCur])
		}
	}

	//shift
	if keys.Shift {
		//ctrl
		if keys.Ctrl {
			if keys.ArrowL {
				p := _UiPaint_CursorMoveLR(text, *e, -1, prop)
				first, _ := _UiPaint_CursorWordRange(text, p)
				if first == p && p > 0 {
					first, _ = _UiPaint_CursorWordRange(text, p-1)
				}
				*e = first
			}
			if keys.ArrowR {
				p := _UiPaint_CursorMoveLR(text, *e, +1, prop)
				_, last := _UiPaint_CursorWordRange(text, p)
				if last == p && p+1 < len(text) {
					_, last = _UiPaint_CursorWordRange(text, p+1)
				}
				*e = last
			}
		} else {
			if multi_line {
				if keys.ArrowU {
					*e = _UiPaint_CursorMoveU(text, lines, *e)
				}
				if keys.ArrowD {
					*e = _UiPaint_CursorMoveD(text, lines, *e)
				}
			}

			if keys.ArrowL {
				p := _UiPaint_CursorMoveLR(text, *e, -1, prop)
				*e = p
			}
			if keys.ArrowR {
				p := _UiPaint_CursorMoveLR(text, *e, +1, prop)
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
		ui._UiPaint_Text_VScrollInto(dom, lines, newPos, prop)
	}
	if old != newPos {
		ui._UiPaint_Text_HScrollInto(dom, text, lines, newPos, prop)
	}
}

func (ui *Ui) _UiPaint_Text_VScrollInto(dom *Layout3, lines []WinGphLine, cursor int, prop WinFontProps) {
	/*if dom.parentLevel == nil {
		return
	}*/

	v_pos := WinGph_CursorLineY(lines, cursor) * prop.lineH

	v_st := dom.scrollV.GetWheel()
	v_sz := dom.view.Size.Y - prop.lineH //- ui.CellWidth(2*0.1)
	v_en := v_st + v_sz

	backup_wheel := dom.scrollV.wheel
	if v_pos <= v_st {
		dom.scrollV.SetWheel(OsMax(0, v_pos))
	} else if v_pos >= v_en {
		dom.scrollV.wheel = OsMax(0, v_pos-v_sz) //SetWheel() has boundary check, which is not good here
	}

	if backup_wheel != dom.scrollV.wheel {
		dom.RebuildSoft()
		dom.GetSettings().SetScrollV(dom.props.Hash, dom.scrollV.wheel)
	}

}

func (ui *Ui) _UiPaint_Text_HScrollInto(dom *Layout3, text string, lines []WinGphLine, cursor int, prop WinFontProps) {
	/*if dom.parentLevel == nil {
		return nil
	}*/

	ln, curr := WinGph_CursorLine(text, lines, cursor)
	h_pos := ui.GetTextSize(curr, ln, prop).X

	h_st := dom.scrollH.GetWheel()
	h_sz := dom.view.Size.X - ui.CellWidth(0.1) //text is shifted 0.1 to left ...
	h_en := h_st + h_sz

	backup_wheel := dom.scrollH.wheel
	if h_pos <= h_st {
		dom.scrollH.SetWheel(OsMax(0, h_pos))
	} else if h_pos >= h_en {
		dom.scrollH.wheel = OsMax(0, h_pos-h_sz) //SetWheel() has boundary check, which is not good here
	}

	if backup_wheel != dom.scrollH.wheel {
		dom.RebuildSoft()
		dom.GetSettings().SetScrollH(dom.props.Hash, dom.scrollH.wheel)
	}

}

func _UiPaint_RemoveFormatingRGBA(str string) string {

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

func _UiPaint_RemoveFormating(str string) string {
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

	str = _UiPaint_RemoveFormatingRGBA(str)

	return str
}

func (ui *Ui) _Paint_getMaxLinePx(view OsV4, multi_line, line_wrapping bool) int {
	max_line_px := -1
	if multi_line && line_wrapping {
		max_line_px = view.Size.X //- ui.CellWidth(3*0.1) //3 ...
	}

	return max_line_px
}

func _UiPaint_HashFormatingPreSuf_fix(str string, startWith bool) int {

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

func _UiPaint_CheckSelectionExplode(str string, start *int, end *int, prop *WinFontProps) {
	if !prop.formating {
		return
	}

	if *start < *end {
		*start -= _UiPaint_HashFormatingPreSuf_fix(str[:*start], false)
		*end += _UiPaint_HashFormatingPreSuf_fix(str[*end:], true)
	}
	if *end < *start {
		*end -= _UiPaint_HashFormatingPreSuf_fix(str[:*end], false)
		*start += _UiPaint_HashFormatingPreSuf_fix(str[*start:], true)
	}
}

func _UiPaint_GetLineYCrop(startY int, num_lines int, crop OsV4, prop WinFontProps) (int, int) {

	sy := (crop.Start.Y - startY) / prop.lineH
	ey := OsRoundUp(float64(crop.End().Y-startY) / float64(prop.lineH))

	//check
	sy = OsClamp(sy, 0, num_lines-1)
	ey = OsClamp(ey, 0, num_lines)

	return sy, ey
}

func _UiPaint_CursorMoveLR(text string, cursor int, move int, prop WinFontProps) int {

	//skip formating
	if prop.formating {
		if move < 0 { //left
			cursor -= _UiPaint_HashFormatingPreSuf_fix(text[:cursor], false)
		}

		if move > 0 { //right
			cursor += _UiPaint_HashFormatingPreSuf_fix(text[cursor:], true)
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

func _UiPaint_CursorMoveU(text string, lines []WinGphLine, cursor int) int {
	y := WinGph_CursorLineY(lines, cursor)
	if y > 0 {
		_, pos := WinGph_CursorLine(text, lines, cursor)

		st, en := WinGph_PosLineRange(lines, y-1) //up line
		cursor = st + OsMin(pos, en-st)
	}
	return cursor
}
func _UiPaint_CursorMoveD(text string, lines []WinGphLine, cursor int) int {
	y := WinGph_CursorLineY(lines, cursor)
	if y+1 < len(lines) {
		_, pos := WinGph_CursorLine(text, lines, cursor)

		st, en := WinGph_PosLineRange(lines, y+1) //down line
		cursor = st + OsMin(pos, en-st)
	}
	return cursor
}
