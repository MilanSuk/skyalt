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

func (app *App) comp_drawButton(style *CompStyle, value string, icon string, icon_margin float64, url string, tooltip string, enable bool) (int, int, int64) {

	root := app.db.root
	st := root.levels.GetStack()

	if style == nil {
		style = &root.styles.Button
	}

	click, rclick, _ := style.IsClicked(enable, app)
	if click > 0 && len(url) > 0 {
		//SA_DialogStart() warning which open dialog ...
		OsUlit_OpenBrowser(url)
	}

	style.Paint(st.stack.canvas, value, "", false, false, icon, icon_margin, enable, app)
	if len(tooltip) > 0 {
		app.paint_tooltip(0, 0, 1, 1, tooltip)
	} else if len(url) > 0 {
		app.paint_tooltip(0, 0, 1, 1, url)
	}

	return click, rclick, 1
}

func (app *App) _sa_comp_drawButton(styleId uint32, valueMem uint64, iconMem uint64, icon_margin float64, urlMem uint64, tooltipMem uint64, enable uint32, outMem uint64) int64 {

	value, err := app.ptrToString(valueMem)
	if app.AddLogErr(err) {
		return -1
	}
	icon, err := app.ptrToString(iconMem)
	if app.AddLogErr(err) {
		return -1
	}
	url, err := app.ptrToString(urlMem)
	if app.AddLogErr(err) {
		return -1
	}
	tooltip, err := app.ptrToString(tooltipMem)
	if app.AddLogErr(err) {
		return -1
	}

	style := app.styles.Get(styleId)

	lclicks, rclicks, ret := app.comp_drawButton(style, value, icon, icon_margin, url, tooltip, enable > 0)

	out, err := app.ptrToBytesDirect(outMem)
	if app.AddLogErr(err) {
		return -1
	}
	binary.LittleEndian.PutUint64(out[0:], uint64(lclicks)) //click
	binary.LittleEndian.PutUint64(out[8:], uint64(rclicks)) //r-click
	return ret
}

func (app *App) comp_drawText(style *CompStyle, value string, icon string, icon_margin float64, tooltip string, enable bool, selection bool) int64 {

	root := app.db.root
	if style == nil {
		style = &root.styles.Text
	}

	app.paint_textGrid(style, value, "", icon, icon_margin, selection, false, enable)
	if len(tooltip) > 0 {
		app.paint_tooltip(0, 0, 1, 1, tooltip)
	}

	return 1
}

func (app *App) _sa_comp_drawText(styleId uint32, valueMem uint64, iconMem uint64, icon_margin float64, tooltipMem uint64, enable uint32, selection uint32) int64 {

	value, err := app.ptrToString(valueMem)
	if app.AddLogErr(err) {
		return -1
	}

	icon, err := app.ptrToString(iconMem)
	if app.AddLogErr(err) {
		return -1
	}

	tooltip, err := app.ptrToString(tooltipMem)
	if app.AddLogErr(err) {
		return -1
	}

	style := app.styles.Get(styleId)

	return app.comp_drawText(style, value, icon, icon_margin, tooltip, enable > 0, selection > 0)
}

func (app *App) comp_getEditValue() string {
	return app.db.root.ui.io.edit.last_edit
}

func (app *App) _sa_comp_getEditValue(outMem uint64) int64 {
	err := app.stringToPtr(app.comp_getEditValue(), outMem)
	if !app.AddLogErr(err) {
		return -1
	}
	return 1
}

func (app *App) comp_drawEdit(style *CompStyle, valueIn string, valueInOrig string, icon string, icon_margin float64, tooltip string, ghost string, enable bool) (string, bool, bool, bool) {

	root := app.db.root
	st := root.levels.GetStack()

	if style == nil {
		style = &root.styles.Editbox
	}

	st.stack.data.scrollH.narrow = true
	st.stack.data.scrollV.show = false

	edit := &root.ui.io.edit

	inDiv := st.stack.FindOrCreate("", InitOsQuad(0, 0, 1, 1), app)
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

	app.paint_textGrid(style, value, valueInOrig, icon, icon_margin, true, true, enable)

	//ghost
	if len(edit.last_edit) == 0 && len(ghost) > 0 {
		stArrow := *style
		stArrow.FontAlignH(1)
		stArrow.ContentCd(OsCd{})
		stArrow.Color(themeGrey(0.7))
		app.paint_text(0, 0, 1, 1, &stArrow, ghost, "", false, false, enable)
	}

	if len(tooltip) > 0 {
		app.paint_tooltip(0, 0, 1, 1, tooltip)
	}

	return edit.last_edit, active, (active && value != edit.last_edit), (active && this_uid != edit.uid)
}

func (app *App) _sa_comp_drawEdit(styleId uint32, valueMem uint64, valueInOrig uint64, iconMem uint64, icon_margin float64, tooltipMem uint64, ghostMem uint64, enable uint32, outMem uint64) int64 {

	value, err := app.ptrToString(valueMem)
	if app.AddLogErr(err) {
		return -1
	}
	valueOrig, err := app.ptrToString(valueInOrig)
	if app.AddLogErr(err) {
		return -1
	}

	icon, err := app.ptrToString(iconMem)
	if app.AddLogErr(err) {
		return -1
	}

	tooltip, err := app.ptrToString(tooltipMem)
	if app.AddLogErr(err) {
		return -1
	}

	ghost, err := app.ptrToString(ghostMem)
	if app.AddLogErr(err) {
		return -1
	}

	style := app.styles.Get(styleId)
	last_edit, active, changed, finished := app.comp_drawEdit(style, value, valueOrig, icon, icon_margin, tooltip, ghost, enable > 0)

	out, err := app.ptrToBytesDirect(outMem)
	if app.AddLogErr(err) {
		return -1
	}
	binary.LittleEndian.PutUint64(out[0:], uint64(OsTrn(active, 1, 0)))    //active
	binary.LittleEndian.PutUint64(out[8:], uint64(OsTrn(changed, 1, 0)))   //changed
	binary.LittleEndian.PutUint64(out[16:], uint64(OsTrn(finished, 1, 0))) //finished
	binary.LittleEndian.PutUint64(out[24:], uint64(len(last_edit)))        //size
	return 1
}

func (app *App) comp_drawProgress(styleFrame *CompStyle, styleStatus *CompStyle, value float64, prec int, tooltip string, enable bool) int64 {

	root := app.db.root
	st := root.levels.GetStack()

	if styleFrame == nil {
		styleFrame = &root.styles.ProgressFrame
	}
	if styleStatus == nil {
		styleStatus = &root.styles.ProgressStatus
	}

	value = OsClampFloat(value, 0, 1)

	styleFrame.Paint(st.stack.canvas, "", "", false, false, "", 0, enable, app)
	styleStatus.Paint(app.getCoord(0, 0, value, 1, 0, 0, 0), strconv.FormatFloat(value*100, 'f', prec, 64)+"%", "", false, false, "", 0, enable, app)

	if len(tooltip) > 0 {
		app.paint_tooltip(0, 0, 1, 1, tooltip)
	}

	return 1
}

func (app *App) _sa_comp_drawProgress(styleFrameId uint32, styleStatusId uint32, value float64, prec int32, tooltipMem uint64, enable uint32) int64 {

	tooltip, err := app.ptrToString(tooltipMem)
	if app.AddLogErr(err) {
		return -1
	}

	styleFrame := app.styles.Get(styleFrameId)
	styleStatus := app.styles.Get(styleStatusId)

	return app.comp_drawProgress(styleFrame, styleStatus, value, int(prec), tooltip, enable > 0)
}

func (app *App) comp_drawSlider(styleTrack *CompStyle, styleThumb *CompStyle, value float64, minValue float64, maxValue float64, jumpValue float64, tooltip string, enable bool) (float64, bool, bool, bool) {
	root := app.db.root
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

	cell := float64(app.db.root.ui.Cell())
	rad := styleThumb.Main.Max_width / 2
	rad_sp := rad / (float64(st.stack.canvas.Size.X) / cell)

	rpos := root.ui.io.touch.pos.Sub(st.stack.canvas.Start)
	touch_x := float64(rpos.X) / float64(st.stack.canvas.Size.X)

	if enable {
		if active || inside {
			app.paint_cursor("hand")
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
		styleTrack.Paint(app.getCoord(rad_sp, 0, t, 1, 0, 0, 0), "", "", false, false, "", 0, enable, app)
		st2.Paint(app.getCoord(t, 0, 1-t-rad_sp, 1, 0, 0, 0), "", "", false, false, "", 0, enable, app)
		//thumb
		styleThumb.Paint(app.getCoord(t-rad_sp, 0, 1, 1, 0, 0, 0), "", "", false, false, "", 0, enable, app)
	}

	if active {
		p := app.getCoord(t+rad_sp*2, 0, 1, 1, 0, 0, 0).Start
		app.db.root.tile.SetForce(p, strconv.FormatFloat(value, 'f', 2, 64), OsCd_black())
	}

	if len(tooltip) > 0 {
		app.paint_tooltip(0, 0, 1, 1, tooltip)
	}

	return value, active, (active && old_value != value), end
}

func (app *App) _sa_comp_drawSlider(styleTrackId uint32, styleThumbId uint32, value float64, minValue float64, maxValue float64, jumpValue float64, tooltipMem uint64, enable uint32, outMem uint64) float64 {

	tooltip, err := app.ptrToString(tooltipMem)
	if app.AddLogErr(err) {
		return -1
	}

	styleTrack := app.styles.Get(styleTrackId)
	styleThumb := app.styles.Get(styleThumbId)
	value, active, changed, finished := app.comp_drawSlider(styleTrack, styleThumb, value, minValue, maxValue, jumpValue, tooltip, enable > 0)

	out, err := app.ptrToBytesDirect(outMem)
	if app.AddLogErr(err) {
		return -1
	}
	binary.LittleEndian.PutUint64(out[0:], uint64(OsTrn(active, 1, 0)))    //active
	binary.LittleEndian.PutUint64(out[8:], uint64(OsTrn(changed, 1, 0)))   //changed
	binary.LittleEndian.PutUint64(out[16:], uint64(OsTrn(finished, 1, 0))) //finished

	return value
}

func (app *App) comp_drawCombo(style *CompStyle, styleMenu *CompStyle, value uint64, optionsIn string, tooltip string, enable bool) int64 {

	root := app.db.root
	div := root.levels.GetStack().stack

	if style == nil {
		style = &root.styles.Combo
	}
	if styleMenu == nil {
		styleMenu = &root.styles.ButtonMenu
	}

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

	//back and arrow
	stText := *style
	stText.FontAlignH(2)
	stText.FontH(0.3)
	stText.Main.Padding_right = 0.2
	stText.Hover.Padding_right = 0.2
	stText.Touch_hover.Padding_right = 0.2
	stText.Touch_out.Padding_right = 0.2
	stText.Disable.Padding_right = 0.2
	stText.Paint(app.getCoord(0, 0, 1, 1, 0, 0, 0), "▼", "", false, false, "", 0, enable, app)

	//text
	div.FindOrCreate("", InitOsQuad(0, 0, 1, 1), app).data.touch_enabled = false //click through
	stArrow := *style
	stArrow.ContentCd(OsCd{})
	app.paint_textGrid(&stArrow, val, "", "", 0, false, false, enable)

	//dialog
	nmd := "combo_" + strconv.Itoa(int(div.data.hash))
	if div.data.touch_end && enable {
		app.div_dialogOpen(nmd, 1)
	}
	if app.div_dialogStart(nmd) > 0 {
		//compute minimum dialog width
		mx := 0
		for _, opt := range options {
			mx = OsMax(mx, len(opt))
		}

		app._sa_div_colMax(0, OsMaxFloat(5, styleMenu.Main.Font_height*float64(mx)))

		menuSt := *styleMenu

		for i, opt := range options {
			app.div_start(0, uint64(i), 1, 1, "")

			//highlight
			if value == uint64(i) {
				menuSt.Main.Content_color = root.themeCd()
			} else {
				menuSt.Main.Content_color = styleMenu.Main.Content_color //default
			}

			lclicks, _, ret := app.comp_drawButton(&menuSt, opt, "", 0, "", "", true)
			if ret > 0 && lclicks > 0 {
				value = uint64(i)
				app._sa_div_dialogClose()
				break
			}

			app._sa_div_end()
		}

		app._sa_div_dialogEnd()
	}

	if len(tooltip) > 0 {
		app.paint_tooltip(0, 0, 1, 1, tooltip)
	}

	return int64(value)
}

func (app *App) _sa_comp_drawCombo(styleId uint32, styleMenuId uint32, value uint64, optionsMem uint64, tooltipMem uint64, enable uint32) int64 {

	options, err := app.ptrToString(optionsMem)
	if app.AddLogErr(err) {
		return -1
	}
	tooltip, err := app.ptrToString(tooltipMem)
	if app.AddLogErr(err) {
		return -1
	}

	style := app.styles.Get(styleId)
	styleMenu := app.styles.Get(styleMenuId)

	return app.comp_drawCombo(style, styleMenu, value, options, tooltip, enable > 0)
}

func (app *App) comp_drawCheckbox(styleCheck *CompStyle, styleLabel *CompStyle, value uint64, label string, tooltip string, enable bool) int64 {

	root := app.db.root
	st := root.levels.GetStack()

	if styleCheck == nil {
		styleCheck = &root.styles.CheckboxCheck
	}
	if styleLabel == nil {
		styleLabel = &root.styles.CheckboxLabel
	}

	if enable {
		active := st.stack.data.touch_active
		inside := st.stack.data.touch_inside
		end := st.stack.data.touch_end

		if active || inside {
			app.paint_cursor("hand")
		}

		if inside && end {
			value = uint64(OsTrn(value > 0, 0, 1))
		}

	}

	var content OsV4
	if len(label) > 0 {
		w := 1.0 / (float64(st.stack.canvas.Size.X) / float64(root.ui.Cell()))
		w *= styleCheck.Main.Max_width * 1.2

		content = styleCheck.Paint(app.getCoord(0, 0, w, 1, 0, 0, 0), "", "", false, false, "", 0, enable, app)

		app.paint_text(w, 0, 1-w, 1, styleLabel, label, "", false, false, enable)

	} else {
		//center
		content = styleCheck.Paint(st.stack.canvas, "", "", false, false, "", 0, enable, app)

	}

	if value > 0 {
		st.buff.AddRect(content, styleCheck.Main.Border_color, 0)

		//draw check
		//content = content.AddSpace(app.getCellWidth(0.1))
		//st.buff.AddLine(content.GetPos(1.0/3, 0.9), content.GetPos(0.05, 2.0/3), themeWhite(), app.getCellWidth(0.05))
		//st.buff.AddLine(content.GetPos(1.0/3, 0.9), content.GetPos(0.95, 1.0/4), themeWhite(), app.getCellWidth(0.05))
	}

	if len(tooltip) > 0 {
		app.paint_tooltip(0, 0, 1, 1, tooltip)
	}

	return int64(value)
}

func (app *App) _sa_comp_drawCheckbox(styleCheckId uint32, styleLabelId uint32, value uint64, labelMem uint64, tooltipMem uint64, enable uint32) int64 {

	label, err := app.ptrToString(labelMem)
	if app.AddLogErr(err) {
		return -1
	}

	tooltip, err := app.ptrToString(tooltipMem)
	if app.AddLogErr(err) {
		return -1
	}

	styleCheck := app.styles.Get(styleCheckId)
	styleLabel := app.styles.Get(styleLabelId)

	return app.comp_drawCheckbox(styleCheck, styleLabel, value, label, tooltip, enable > 0)
}

func (app *App) paint_textWidth(style *CompStyle, value string, cursorPos int64) float64 {

	sdiv := style.GetDiv(true, app)

	ratioH := sdiv.Font_height
	if ratioH <= 0 {
		ratioH = 0.35
	}
	font := app.db.root.fonts.Get(sdiv.Font_path)
	textH := app.getCellWidth(ratioH)
	cell := float64(app.db.root.ui.Cell())
	if cursorPos < 0 {
		size, err := font.GetTextSize(value, g_Font_DEFAULT_Weight, textH, 0, sdiv.Font_formating)
		if err == nil {
			return float64(size.X) / cell // pixels for the whole string
		}
	} else {
		px, err := font.GetPxPos(value, g_Font_DEFAULT_Weight, textH, int(cursorPos), sdiv.Font_formating)
		if err == nil {
			return float64(px) / cell // pixels to cursor
		}
	}
	return -1
}

func (app *App) _sa_paint_textWidth(styleId uint32, valueMem uint64, cursorPos int64) float64 {

	value, err := app.ptrToString(valueMem)
	if app.AddLogErr(err) {
		return -1
	}

	style := app.styles.Get(styleId)
	return app.paint_textWidth(style, value, cursorPos)
}
