/*
Copyright 2023 Milan Suk

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
	"encoding/binary"
	"math"
	"strconv"
	"strings"
)

//add: ...
//- Switch
//- Radio button

func themeBack() OsCd {
	return OsCd{220, 220, 220, 255}
}
func themeWhite() OsCd {
	return OsCd{255, 255, 255, 255}
}
func themeBlack() OsCd {
	return OsCd{0, 0, 0, 255}
}
func themeGrey(t float64) OsCd {
	return OsCd{byte(255 * t), byte(255 * t), byte(255 * t), 255}
}
func themeWarning() OsCd {
	return OsCd{230, 110, 50, 255}
}
func themeEdit() OsCd {
	return OsCd{225, 226, 68, 255}
}

func (root *Root) themeCd() OsCd {

	cd := OsCd{90, 180, 180, 255} // ocean
	switch root.ui.io.ini.Theme {
	case 1:
		cd = OsCd{200, 100, 80, 255}
	case 2:
		cd = OsCd{130, 170, 210, 255}
	case 3:
		cd = OsCd{130, 180, 130, 255}
	case 4:
		cd = OsCd{160, 160, 160, 255}
	}
	return cd
}

func (asset *Asset) swp_drawButton(style *SwpStyle, value string, icon string, icon_margin float64, url string, title string, enable bool) (bool, bool, int64) {

	root := asset.app.root
	st := root.levels.GetStack()

	if style == nil {
		style = &root.styles.Button
	}

	click, rclick, _ := style.IsClicked(enable, asset)
	if click && len(url) > 0 {
		//SA_DialogStart() warning which open dialog ...
		OsUlit_OpenBrowser(url)
	}

	style.Paint(st.stack.canvas, value, "", false, false, icon, icon_margin, enable, asset)
	if len(title) > 0 {
		asset.paint_title(0, 0, 1, 1, title)
	} else if len(url) > 0 {
		asset.paint_title(0, 0, 1, 1, url)
	}

	return click, rclick, 1
}

func (asset *Asset) _sa_swp_drawButton(styleId uint32, valueMem uint64, iconMem uint64, icon_margin float64, urlMem uint64, titleMem uint64, enable uint32, outMem uint64) int64 {

	value, err := asset.ptrToString(valueMem)
	if asset.AddLogErr(err) {
		return -1
	}
	icon, err := asset.ptrToString(iconMem)
	if asset.AddLogErr(err) {
		return -1
	}
	url, err := asset.ptrToString(urlMem)
	if asset.AddLogErr(err) {
		return -1
	}
	title, err := asset.ptrToString(titleMem)
	if asset.AddLogErr(err) {
		return -1
	}

	style := asset.styles.Get(styleId)

	click, rclick, ret := asset.swp_drawButton(style, value, icon, icon_margin, url, title, enable > 0)

	out, err := asset.ptrToBytesDirect(outMem)
	if asset.AddLogErr(err) {
		return -1
	}
	binary.LittleEndian.PutUint64(out[0:], uint64(OsTrn(click, 1, 0)))  //click
	binary.LittleEndian.PutUint64(out[8:], uint64(OsTrn(rclick, 1, 0))) //r-click
	return ret
}

func (asset *Asset) swp_drawText(style *SwpStyle, value string, title string, enable bool, selection bool) int64 {

	root := asset.app.root
	if style == nil {
		style = &root.styles.Text
	}

	asset.paint_textGrid(style, value, "", selection, false, enable)
	if len(title) > 0 {
		asset.paint_title(0, 0, 1, 1, title)
	}

	return 1
}

func (asset *Asset) _sa_swp_drawText(styleId uint32, valueMem uint64, titleMem uint64, enable uint32, selection uint32) int64 {

	value, err := asset.ptrToString(valueMem)
	if asset.AddLogErr(err) {
		return -1
	}

	title, err := asset.ptrToString(titleMem)
	if asset.AddLogErr(err) {
		return -1
	}

	style := asset.styles.Get(styleId)

	return asset.swp_drawText(style, value, title, enable > 0, selection > 0)
}

func (asset *Asset) swp_getEditValue() string {
	return asset.app.root.ui.io.edit.last_edit
}

func (asset *Asset) _sa_swp_getEditValue(outMem uint64) int64 {
	err := asset.stringToPtr(asset.swp_getEditValue(), outMem)
	if !asset.AddLogErr(err) {
		return -1
	}
	return 1
}

func (asset *Asset) swp_drawEdit(style *SwpStyle, valueIn string, valueInOrig string, title string, ghost string, enable bool) (string, bool, bool, bool) {

	root := asset.app.root
	st := root.levels.GetStack()

	if style == nil {
		style = &root.styles.Editbox
	}

	st.stack.data.scrollH.narrow = true
	st.stack.data.scrollV.show = false

	edit := &root.ui.io.edit

	inDiv := st.stack.FindOrCreate("", InitOsQuad(0, 0, 1, 1), &root.levels.infoLayout)
	this_uid := inDiv //.Hash()
	edit_uid := edit.uid
	active := (edit_uid != nil && edit_uid == this_uid)

	var value string
	if active {
		value = edit.temp
	} else {
		value = valueIn
	}
	inDiv.data.touch_enabled = enable

	asset.paint_textGrid(style, value, valueInOrig, true, true, enable)

	//ghost
	if len(edit.last_edit) == 0 && len(ghost) > 0 {
		stArrow := *style
		stArrow.FontAlignH(1)
		stArrow.ContentCd(OsCd{})
		stArrow.Color(themeGrey(0.7))
		asset.paint_text(0, 0, 1, 1, &stArrow, ghost, "", false, false, enable)
	}

	if len(title) > 0 {
		asset.paint_title(0, 0, 1, 1, title)
	}

	return edit.last_edit, active, (active && value != edit.last_edit), (active && this_uid != edit.uid)
}

func (asset *Asset) _sa_swp_drawEdit(styleId uint32, valueMem uint64, valueInOrig uint64, titleMem uint64, ghostMem uint64, enable uint32, outMem uint64) int64 {

	value, err := asset.ptrToString(valueMem)
	if asset.AddLogErr(err) {
		return -1
	}
	valueOrig, err := asset.ptrToString(valueInOrig)
	if asset.AddLogErr(err) {
		return -1
	}

	title, err := asset.ptrToString(titleMem)
	if asset.AddLogErr(err) {
		return -1
	}

	ghost, err := asset.ptrToString(ghostMem)
	if asset.AddLogErr(err) {
		return -1
	}

	style := asset.styles.Get(styleId)
	last_edit, active, changed, finished := asset.swp_drawEdit(style, value, valueOrig, title, ghost, enable > 0)

	out, err := asset.ptrToBytesDirect(outMem)
	if asset.AddLogErr(err) {
		return -1
	}
	binary.LittleEndian.PutUint64(out[0:], uint64(OsTrn(active, 1, 0)))    //active
	binary.LittleEndian.PutUint64(out[8:], uint64(OsTrn(changed, 1, 0)))   //changed
	binary.LittleEndian.PutUint64(out[16:], uint64(OsTrn(finished, 1, 0))) //finished
	binary.LittleEndian.PutUint64(out[24:], uint64(len(last_edit)))        //size
	return 1
}

func (asset *Asset) swp_drawProgress(styleFrame *SwpStyle, styleStatus *SwpStyle, value float64, prec int, title string, enable bool) int64 {

	root := asset.app.root
	st := root.levels.GetStack()

	if styleFrame == nil {
		styleFrame = &root.styles.ProgressFrame
	}
	if styleStatus == nil {
		styleStatus = &root.styles.ProgressStatus
	}

	value = OsClampFloat(value, 0, 1)

	styleFrame.Paint(st.stack.canvas, "", "", false, false, "", 0, enable, asset)
	styleStatus.Paint(asset.getCoord(0, 0, value, 1, 0, 0, 0), strconv.FormatFloat(value*100, 'f', prec, 64)+"%", "", false, false, "", 0, enable, asset)

	if len(title) > 0 {
		asset.paint_title(0, 0, 1, 1, title)
	}

	return 1
}

func (asset *Asset) _sa_swp_drawProgress(styleFrameId uint32, styleStatusId uint32, value float64, prec int32, titleMem uint64, enable uint32) int64 {

	title, err := asset.ptrToString(titleMem)
	if asset.AddLogErr(err) {
		return -1
	}

	styleFrame := asset.styles.Get(styleFrameId)
	styleStatus := asset.styles.Get(styleStatusId)

	return asset.swp_drawProgress(styleFrame, styleStatus, value, int(prec), title, enable > 0)
}

func (asset *Asset) swp_drawSlider(styleTrack *SwpStyle, styleThumb *SwpStyle, value float64, minValue float64, maxValue float64, jumpValue float64, title string, enable bool) (float64, bool, bool, bool) {
	root := asset.app.root
	st := root.levels.GetStack()

	if styleTrack == nil {
		styleTrack = &root.styles.SliderTrack
	}
	if styleThumb == nil {
		styleThumb = &root.styles.SliderThumb
	}

	old_value := value

	active := st.stack.data.touch_active
	inside := st.stack.data.touch_inside
	end := st.stack.data.touch_end

	cell := float64(asset.app.root.ui.Cell())
	rad := styleThumb.Main.Max_width / 2
	rad_sp := rad / (float64(st.stack.canvas.Size.X) / cell)

	rpos := root.ui.io.touch.pos.Sub(st.stack.canvas.Start)
	touch_x := float64(rpos.X) / float64(st.stack.canvas.Size.X)

	if enable {
		if active || inside {
			asset.paint_cursor("hand")
		}

		if active {
			//cut space from touch_x: outer(0,1) => inner(0,1)
			touch_x = OsClampFloat(touch_x, rad_sp, 1-rad_sp)
			touch_x = (touch_x - rad_sp) / (1 - 2*rad_sp)

			value = minValue + (maxValue-minValue)*touch_x

			t := math.Round((value - minValue) / jumpValue)
			value = minValue + t*jumpValue
			value = OsClampFloat(value, minValue, maxValue)
		}
	}

	t := (value - minValue) / (maxValue - minValue)
	//inner(0,1) => outer(rad,1-rad)
	t = (t + rad_sp) * (1 - 2*rad_sp)

	//draw
	{
		st2 := *styleTrack
		cd := st2.Main.Content_color
		cd.A = 100
		st2.ContentCd(cd)

		//track
		styleTrack.Paint(asset.getCoord(rad_sp, 0, t, 1, 0, 0, 0), "", "", false, false, "", 0, enable, asset)
		st2.Paint(asset.getCoord(t, 0, 1-t-rad_sp, 1, 0, 0, 0), "", "", false, false, "", 0, enable, asset)
		//thumb
		styleThumb.Paint(asset.getCoord(t-rad_sp, 0, 1, 1, 0, 0, 0), "", "", false, false, "", 0, enable, asset)
	}

	if active {
		p := asset.getCoord(t+rad_sp*2, 0, 1, 1, 0, 0, 0).Start
		asset.app.root.tile.SetForce(p, strconv.FormatFloat(value, 'f', 2, 64), OsCd_black())
	}

	if len(title) > 0 {
		asset.paint_title(0, 0, 1, 1, title)
	}

	return value, active, (active && old_value != value), end
}

func (asset *Asset) _sa_swp_drawSlider(styleTrackId uint32, styleThumbId uint32, value float64, minValue float64, maxValue float64, jumpValue float64, titleMem uint64, enable uint32, outMem uint64) float64 {

	title, err := asset.ptrToString(titleMem)
	if asset.AddLogErr(err) {
		return -1
	}

	styleTrack := asset.styles.Get(styleTrackId)
	styleThumb := asset.styles.Get(styleThumbId)
	value, active, changed, finished := asset.swp_drawSlider(styleTrack, styleThumb, value, minValue, maxValue, jumpValue, title, enable > 0)

	out, err := asset.ptrToBytesDirect(outMem)
	if asset.AddLogErr(err) {
		return -1
	}
	binary.LittleEndian.PutUint64(out[0:], uint64(OsTrn(active, 1, 0)))    //active
	binary.LittleEndian.PutUint64(out[8:], uint64(OsTrn(changed, 1, 0)))   //changed
	binary.LittleEndian.PutUint64(out[16:], uint64(OsTrn(finished, 1, 0))) //finished

	return value
}

func (asset *Asset) swp_drawCombo(style *SwpStyle, styleMenu *SwpStyle, value uint64, optionsIn string, title string, enable bool) int64 {

	root := asset.app.root
	div := root.levels.GetStack().stack

	var options []string
	if len(optionsIn) > 0 {
		options = strings.Split(optionsIn, "|")
	}
	var val string
	if value >= uint64(len(options)) {
		val = ""
	} else {
		val = options[value]
	}

	//w := 0.6 / (float64(div.canvas.Size.X) / float64(asset.app.root.ui.Cell()))

	//back and arrow
	stText := *style
	stText.FontAlignH(2)
	stText.Main.Padding_right = 0.2
	stText.Hover.Padding_right = 0.2
	stText.Touch_hover.Padding_right = 0.2
	stText.Touch_out.Padding_right = 0.2
	stText.Disable.Padding_right = 0.2
	stText.Paint(asset.getCoord(0, 0, 1, 1, 0, 0, 0), "▼", "", false, false, "", 0, enable, asset)

	//text
	div.FindOrCreate("", InitOsQuad(0, 0, 1, 1), &root.levels.infoLayout).data.touch_enabled = false //click through
	stArrow := *style
	stArrow.ContentCd(OsCd{})
	asset.paint_textGrid(&stArrow, val, "", false, false, enable)

	//dialog
	nmd := "combo_" + strconv.Itoa(int(div.Hash()))
	if div.data.touch_end && enable {
		asset.div_dialogOpen(nmd, 1)
	}
	if asset.div_dialogStart(nmd) > 0 {
		//compute minimum dialog width
		mx := 0
		for _, opt := range options {
			mx = OsMax(mx, len(opt))
		}

		asset._sa_div_colMax(0, OsMaxFloat(5, styleMenu.Main.Font_height*float64(mx)))

		menuSt := *styleMenu

		for i, opt := range options {
			asset.div_start(0, uint64(i), 1, 1, "")

			//highlight
			if value == uint64(i) {
				menuSt.Main.Content_color = root.themeCd()
			} else {
				menuSt.Main.Content_color = styleMenu.Main.Content_color //default
			}

			click, _, ret := asset.swp_drawButton(&menuSt, opt, "", 0, "", "", true)
			if ret > 0 && click {
				value = uint64(i)
				asset._sa_div_dialogClose()
				break
			}

			asset._sa_div_end()
		}

		asset._sa_div_dialogEnd()
	}

	if len(title) > 0 {
		asset.paint_title(0, 0, 1, 1, title)
	}

	return int64(value)
}

func (asset *Asset) _sa_swp_drawCombo(styleId uint32, styleMenuId uint32, value uint64, optionsMem uint64, titleMem uint64, enable uint32) int64 {

	options, err := asset.ptrToString(optionsMem)
	if asset.AddLogErr(err) {
		return -1
	}
	title, err := asset.ptrToString(titleMem)
	if asset.AddLogErr(err) {
		return -1
	}

	style := asset.styles.Get(styleId)
	styleMenu := asset.styles.Get(styleMenuId)

	return asset.swp_drawCombo(style, styleMenu, value, options, title, enable > 0)
}

// StyleContainer + StyleIcon(image is empty => default check) ...
func (asset *Asset) swp_drawCheckbox(cd_r, cd_g, cd_b, cd_a uint32,
	value uint64, description string, title string,
	height float64, align uint32, alignV uint32, enable bool) int64 {

	root := asset.app.root
	st := root.levels.GetStack()

	cd := InitOsCd32(cd_r, cd_g, cd_b, cd_a)

	if enable {
		active := st.stack.data.touch_active
		inside := st.stack.data.touch_inside
		end := st.stack.data.touch_end

		if active || inside {
			cd = OsCd_Aprox(cd, OsCd_white(), 0.3)
			asset.paint_cursor("hand")
		}

		if inside && end {
			value = uint64(OsTrn(value > 0, 0, 1))
		}

	} else {
		cd = OsCd_Aprox(OsCd_white(), cd, 0.3)
	}

	ww := float64(st.stack.canvas.Size.X) / float64(root.ui.Cell())
	hh := float64(st.stack.canvas.Size.Y) / float64(root.ui.Cell())

	descSz := asset.paint_textWidth(description, SKYALT_FONT_0, 0.35, -1) //font from style ...

	h := height / hh
	w := h / (ww / hh)

	sx := float64(0)
	switch align {
	case 1:
		sx = OsMaxFloat((1-(w*0.8+descSz))/2, 0)
	case 2:
		sx = OsMaxFloat((1 - (w*0.8 + descSz)), 0)
	}

	sy := float64(0)
	switch alignV {
	case 1:
		sy = OsMaxFloat((1-h)/2, 0)
	case 2:
		sy = OsMaxFloat((1 - h), 0)
	}

	if value > 0 {
		asset.paint_rect(sx, sy, w, h, 0.22, cd, 0)
		asset._sa_paint_line(sx, sy, w, h, 0.33, 1.0/3, 0.9, 0.05, 2.0/3, 255, 255, 255, 255, 0.05)
		asset._sa_paint_line(sx, sy, w, h, 0.33, 1.0/3, 0.9, 0.95, 1.0/4, 255, 255, 255, 255, 0.05)
	} else {
		asset.paint_rect(sx, sy, w, h, 0.22, cd, 0.03)
	}

	div := root.levels.GetStack().stack
	div.FindOrCreate("", InitOsQuad(0, 0, 1, 1), &root.levels.infoLayout).data.touch_enabled = false                //click through
	asset.paint_text(sx+w*1, sy, 1-(sx+w*1), h, &asset.app.root.styles.Text, description, "", false, false, enable) //custom style, cursor is ibeam ...

	if len(title) > 0 {
		asset.paint_title(0, 0, 1, 1, title)
	}

	return int64(value)
}

func (asset *Asset) _sa_swp_drawCheckbox(cd_r, cd_g, cd_b, cd_a uint32,
	value uint64, descriptionMem uint64, titleMem uint64,
	height float64, align uint32, alignV uint32, enable uint32) int64 {

	description, err := asset.ptrToString(descriptionMem)
	if asset.AddLogErr(err) {
		return -1
	}

	title, err := asset.ptrToString(titleMem)
	if asset.AddLogErr(err) {
		return -1
	}

	return asset.swp_drawCheckbox(cd_r, cd_g, cd_b, cd_a, value, description, title, height, align, alignV, enable > 0)
}

func (asset *Asset) paint_textWidth(value string, fontPath string, ratioH float64, cursorPos int64) float64 {

	textH := asset.getCellWidth(ratioH)
	font := asset.app.root.fonts.Get(fontPath)
	cell := float64(asset.app.root.ui.Cell())
	if cursorPos < 0 {

		size, err := font.GetTextSize(value, textH, 0)
		if err == nil {
			return float64(size.X) / cell // pixels for the whole string
		}

	} else {
		px, err := font.GetPxPos(value, textH, int(cursorPos))
		if err == nil {
			return float64(px) / cell // pixels to cursor
		}
	}
	return -1
}

func (asset *Asset) _sa_paint_textWidth(valueMem uint64, fontPathMem uint64, ratioH float64, cursorPos int64) float64 {

	value, err := asset.ptrToString(valueMem)
	if asset.AddLogErr(err) {
		return -1
	}

	fond_path, err := asset.ptrToString(fontPathMem)
	if asset.AddLogErr(err) {
		return -1
	}

	return asset.paint_textWidth(value, fond_path, ratioH, cursorPos)
}
