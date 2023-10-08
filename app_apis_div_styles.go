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
	"encoding/json"
	"fmt"
	"unicode/utf8"
)

type DivStyle struct {
	Max_width, Max_height             float64
	Max_width_align, Max_height_align int

	Margin_top, Margin_bottom, Margin_left, Margin_right     float64
	Border_top, Border_bottom, Border_left, Border_right     float64
	Padding_top, Padding_bottom, Padding_left, Padding_right float64

	Border_color  OsCd
	Content_color OsCd
	Color         OsCd

	Image_fill                 bool
	Image_alignV, Image_alignH int

	Font_path   string
	Font_height float64 //from cell
	//Font_weight int		//bold		//400 ...
	//Font_angle float64	//italic	//0-10 ...
	//Font_space float64				//0 ...
	Font_alignV, Font_alignH int
	Font_formating           bool

	Cursor string

	Radius float64
	//shadow ...
	//transition_sec(blend between states) ...
}

func (st *DivStyle) Margin(v float64) {
	st.Margin_top = v
	st.Margin_bottom = v
	st.Margin_left = v
	st.Margin_right = v
}
func (st *DivStyle) Border(v float64) {
	st.Border_top = v
	st.Border_bottom = v
	st.Border_left = v
	st.Border_right = v
}
func (st *DivStyle) Padding(v float64) {
	st.Padding_top = v
	st.Padding_bottom = v
	st.Padding_left = v
	st.Padding_right = v
}

func _paintBorder(out OsV4, top, bottom, left, right float64, radius float64, cd OsCd, app *App) OsV4 {

	stt := app.db.root.levels.GetStack()

	in := out.Inner(app.getCellWidth(top), app.getCellWidth(bottom), app.getCellWidth(left), app.getCellWidth(right))

	if cd.A == 0 {
		return in
	}

	if radius > 0 {
		rad := app.getCellWidth(radius)
		if rad*2 > out.Size.X && rad*2 > out.Size.Y {
			//circle
			stt.buff.AddCircle(out, cd, app.getCellWidth(bottom))
		} /* else {
			//rect with radius corners ...
		}*/
	} else {
		//sharp edges
		q := OsV4{Start: out.Start, Size: OsV2{out.Size.X, app.getCellWidth(top)}}
		if q.Is() {
			stt.buff.AddRect(q, cd, 0)
		}

		q = OsV4{Start: OsV2{out.Start.X, in.Start.Y + in.Size.Y}, Size: OsV2{out.Size.X, app.getCellWidth(bottom)}}
		if q.Is() {
			stt.buff.AddRect(q, cd, 0)
		}

		q = OsV4{Start: out.Start, Size: OsV2{app.getCellWidth(left), out.Size.Y}}
		if q.Is() {
			stt.buff.AddRect(q, cd, 0)
		}

		q = OsV4{Start: OsV2{in.Start.X + in.Size.X, out.Start.Y}, Size: OsV2{app.getCellWidth(right), out.Size.Y}}
		if q.Is() {

			stt.buff.AddRect(q, cd, 0)
		}
	}

	return in
}

func DivStyle_getCoord(coord OsV4, x, y, w, h float64) OsV4 {

	return InitOsQuad(
		coord.Start.X+int(float64(coord.Size.X)*x),
		coord.Start.Y+int(float64(coord.Size.Y)*y),
		int(float64(coord.Size.X)*w),
		int(float64(coord.Size.Y)*h))
}

func (st *DivStyle) Paint(coord OsV4, text string, textOrig string, textSelect bool, textEdit bool, image_url string, image_margin float64, app *App) OsV4 {

	stt := app.db.root.levels.GetStack()
	if stt.stack == nil || stt.stack.crop.IsZero() {
		return OsV4{}
	}

	if st.Max_width > 0 {
		max := app.getCellWidth(st.Max_width)
		if coord.Size.X > max {
			switch st.Max_width_align {
			case 1:
				coord.Start.X += (coord.Size.X - max) / 2
			case 2:
				coord.Start.X = coord.End().X - max
			}
			coord.Size.X = max
		}
	}

	if st.Max_height > 0 {
		max := app.getCellWidth(st.Max_height)
		if coord.Size.Y > max {
			switch st.Max_height_align {
			case 1:
				coord.Start.Y += (coord.Size.Y - max) / 2
			case 2:
				coord.Start.Y = coord.End().Y - max
			}
			coord.Size.Y = max
		}
	}

	border := _paintBorder(coord, st.Margin_top, st.Margin_bottom, st.Margin_left, st.Margin_right, 0, OsCd{}, app)
	padding := _paintBorder(border, st.Border_top, st.Border_bottom, st.Border_left, st.Border_right, st.Radius, st.Border_color, app)
	content := _paintBorder(padding, st.Padding_top, st.Padding_bottom, st.Padding_left, st.Padding_right, 0, OsCd{}, app)

	//background
	if st.Content_color.A > 0 {
		rad := app.getCellWidth(st.Radius)
		if rad*2 > padding.Size.X && rad*2 > padding.Size.Y {
			stt.buff.AddCircle(padding, st.Content_color, 0)
		} else {
			stt.buff.AddRect(padding, st.Content_color, 0)
		}
	}

	coordImg := content
	coordText := content

	isImg := (len(image_url) > 0)
	isText := (len(text) > 0 || textEdit)

	if isImg && isText {

		w := float64(app.db.root.ui.Cell()) / float64(stt.stack.canvas.Size.X)

		switch st.Image_alignH {
		case 0: //left
			coordImg = DivStyle_getCoord(content, 0, 0, w, 1)
			coordText = DivStyle_getCoord(content, w, 0, 1-w, 1)
		case 1: //center
			coordImg = coord
			coordText = coord
		default: //right
			coordImg = DivStyle_getCoord(content, 1-w, 0, w, 1)
			coordText = DivStyle_getCoord(content, 0, 0, 1-w, 1)
		}
	}

	if isImg {
		path, err := MediaParseUrl(image_url, app)
		if err != nil {
			app.AddLogErr(err)
		} else {
			coordImg = coordImg.Inner(app.getCellWidth(image_margin), app.getCellWidth(image_margin), app.getCellWidth(image_margin), app.getCellWidth(image_margin))

			imgRectBackup := stt.buff.AddCrop(stt.stack.crop.GetIntersect(coordImg))
			stt.buff.AddImage(path, coordImg, st.Color, st.Image_alignV, st.Image_alignH, st.Image_fill)
			stt.buff.AddCrop(imgRectBackup)
		}
	}

	if isText {
		// crop
		imgRectBackup := stt.buff.AddCrop(stt.stack.crop.GetIntersect(coordText))

		//one liner
		active := app._VmDraw_Text_line(coordText, 0, OsV2{utf8.RuneCountInString(text), 0},
			text, textOrig,
			st.Color,
			st.Font_height, 1, 0, 0,
			st.Font_path, st.Font_alignH, st.Font_alignV,
			textSelect, textEdit, false, st.Font_formating)

		if active {
			app._VmDraw_resetKeys(textEdit)
		}

		// crop back
		stt.buff.AddCrop(imgRectBackup)
	}

	return content
}

type CompStyle struct {
	Main        DivStyle
	Hover       DivStyle
	Touch_hover DivStyle
	Touch_out   DivStyle
	Disable     DivStyle
}

func (b *CompStyle) HoverAuto() *CompStyle {
	if b.Main.Color.A > 0 {
		b.Hover.Color = OsCd_Aprox(b.Main.Color, themeWhite(), 0.3)
	}

	if b.Main.Content_color.A > 0 {
		b.Hover.Content_color = OsCd_Aprox(b.Main.Content_color, themeWhite(), 0.3)
	}

	if b.Main.Border_color.A > 0 {
		b.Hover.Border_color = OsCd_Aprox(b.Main.Border_color, themeWhite(), 0.3)
	}

	return b
}

func (b *CompStyle) DisableAuto() *CompStyle {
	if b.Main.Color.A > 0 {
		b.Disable.Color = OsCd_Aprox(b.Main.Color, themeWhite(), 0.35)
	} else {
		b.Disable.Color = OsCd{}
	}

	if b.Main.Content_color.A > 0 {
		b.Disable.Content_color = OsCd_Aprox(b.Main.Content_color, themeWhite(), 0.5)
	} else {
		b.Disable.Content_color = OsCd{}
	}

	return b
}

func (b *CompStyle) Color(v OsCd) *CompStyle {
	b.Main.Color = v
	b.Hover.Color = v
	b.Touch_hover.Color = v
	b.Touch_out.Color = v
	b.Disable.Color = v
	return b
}

func (b *CompStyle) FontPath(v string) *CompStyle {
	b.Main.Font_path = v
	b.Hover.Font_path = v
	b.Touch_hover.Font_path = v
	b.Touch_out.Font_path = v
	b.Disable.Font_path = v
	return b
}

func (b *CompStyle) FontH(v float64) *CompStyle {
	b.Main.Font_height = v
	b.Hover.Font_height = v
	b.Touch_hover.Font_height = v
	b.Touch_out.Font_height = v
	b.Disable.Font_height = v
	return b
}

func (b *CompStyle) FontAlignH(v int) *CompStyle {
	b.Main.Font_alignH = v
	b.Hover.Font_alignH = v
	b.Touch_hover.Font_alignH = v
	b.Touch_out.Font_alignH = v
	b.Disable.Font_alignH = v
	return b
}
func (b *CompStyle) FontAlignV(v int) *CompStyle {
	b.Main.Font_alignV = v
	b.Hover.Font_alignV = v
	b.Touch_hover.Font_alignV = v
	b.Touch_out.Font_alignV = v
	b.Disable.Font_alignV = v
	return b
}
func (b *CompStyle) FontFormating(v bool) *CompStyle {
	b.Main.Font_formating = v
	b.Hover.Font_formating = v
	b.Touch_hover.Font_formating = v
	b.Touch_out.Font_formating = v
	b.Disable.Font_formating = v
	return b
}

func (b *CompStyle) Margin(v float64) *CompStyle {
	b.Main.Margin(v)
	b.Hover.Margin(v)
	b.Touch_hover.Margin(v)
	b.Touch_out.Margin(v)
	b.Disable.Margin(v)
	return b
}
func (b *CompStyle) Padding(v float64) *CompStyle {
	b.Main.Padding(v)
	b.Hover.Padding(v)
	b.Touch_hover.Padding(v)
	b.Touch_out.Padding(v)
	b.Disable.Padding(v)
	return b
}

func (b *CompStyle) Border(v float64) *CompStyle {
	b.Main.Border(v)
	b.Hover.Border(v)
	b.Touch_hover.Border(v)
	b.Touch_out.Border(v)
	b.Disable.Border(v)
	return b
}

func (b *CompStyle) BorderCd(v OsCd) *CompStyle {
	b.Main.Border_color = v
	b.Hover.Border_color = v
	b.Touch_hover.Border_color = v
	b.Touch_out.Border_color = v
	b.Disable.Border_color = v
	return b
}

func (b *CompStyle) ContentCd(v OsCd) *CompStyle {
	b.Main.Content_color = v
	b.Hover.Content_color = v
	b.Touch_hover.Content_color = v
	b.Touch_out.Content_color = v
	b.Disable.Content_color = v
	return b
}

func (b *CompStyle) Cursor(v string) {
	b.Main.Cursor = v
	b.Hover.Cursor = v
	b.Touch_hover.Cursor = v
	b.Touch_out.Cursor = v
	b.Disable.Cursor = v
}
func (b *CompStyle) MaxWidth(v float64) *CompStyle {
	b.Main.Max_width = v
	b.Hover.Max_width = v
	b.Touch_hover.Max_width = v
	b.Touch_out.Max_width = v
	b.Disable.Max_width = v
	return b
}
func (b *CompStyle) MaxHeight(v float64) *CompStyle {
	b.Main.Max_height = v
	b.Hover.Max_height = v
	b.Touch_hover.Max_height = v
	b.Touch_out.Max_height = v
	b.Disable.Max_height = v
	return b
}
func (b *CompStyle) MaxWidthAlign(v int) *CompStyle {
	b.Main.Max_width_align = v
	b.Hover.Max_width_align = v
	b.Touch_hover.Max_width_align = v
	b.Touch_out.Max_width_align = v
	b.Disable.Max_width_align = v
	return b
}
func (b *CompStyle) MaxHeightAlign(v int) *CompStyle {
	b.Main.Max_height_align = v
	b.Hover.Max_height_align = v
	b.Touch_hover.Max_height_align = v
	b.Touch_out.Max_height_align = v
	b.Disable.Max_height_align = v
	return b
}

func (b *CompStyle) Radius(v float64) *CompStyle {
	b.Main.Radius = v
	b.Hover.Radius = v
	b.Touch_hover.Radius = v
	b.Touch_out.Radius = v
	b.Disable.Radius = v
	return b
}

func (style *CompStyle) GetDiv(enable bool, app *App) *DivStyle {

	st := app.db.root.levels.GetStack()

	var stylee *DivStyle
	var inside bool

	if enable {
		active := st.stack.data.touch_active
		inside = st.stack.data.touch_inside

		if active {
			if inside {
				stylee = &style.Touch_hover

			} else {
				stylee = &style.Touch_out
			}
		} else {
			if inside {
				stylee = &style.Hover
			} else {
				stylee = &style.Main
			}
		}
	} else {
		stylee = &style.Disable
	}

	return stylee
}

func (style *CompStyle) IsClicked(enable bool, app *App) (bool, bool, bool) {
	var click, rclick, inside bool
	if enable {
		st := app.db.root.levels.GetStack()
		inside = st.stack.data.touch_inside
		end := st.stack.data.touch_end
		force := app.db.root.ui.io.touch.rm

		if inside && end {
			click = true
			rclick = force
		}
	}

	return click, rclick, inside
}

func (style *CompStyle) Paint(coord OsV4, text string, textOrig string, textSelect bool, textEdit bool, image_path string, image_margin float64, enable bool, app *App) OsV4 {

	stylee := style.GetDiv(enable, app)

	//draw
	content := stylee.Paint(coord, text, textOrig, textSelect, textEdit, image_path, image_margin, app)

	//cursor
	_, _, inside := style.IsClicked(enable, app)

	if inside && len(stylee.Cursor) > 0 {
		app.paint_cursor(stylee.Cursor)
	}

	return content
}

type DivDefaultStyles struct {
	Button             CompStyle
	ButtonLight        CompStyle
	ButtonAlpha        CompStyle
	ButtonMenu         CompStyle
	ButtonMenuSelected CompStyle

	ButtonBorder CompStyle

	ButtonLogo CompStyle

	ButtonDanger     CompStyle
	ButtonDangerMenu CompStyle

	Text       CompStyle
	TextCenter CompStyle
	TextRight  CompStyle
	TextErr    CompStyle

	Editbox       CompStyle
	EditboxErr    CompStyle
	EditboxYellow CompStyle

	Combo CompStyle

	ProgressFrame  CompStyle
	ProgressStatus CompStyle

	SliderTrack CompStyle
	SliderThumb CompStyle

	CheckboxCheck CompStyle
	CheckboxLabel CompStyle
}

func DivStyles_getDefaults(root *Root) DivDefaultStyles {

	stls := DivDefaultStyles{}

	{
		b := &stls.Button.Main
		b.Cursor = "hand"
		b.Content_color = root.themeCd()
		b.Color = themeBlack()
		b.Image_alignV = 1
		b.Image_alignH = 0
		b.Font_path = SKYALT_FONT_PATH
		b.Font_alignV = 1
		b.Font_alignH = 1
		b.Font_height = SKYALT_FONT_HEIGHT
		b.Font_formating = true
		b.Margin(0.06)
		b.Padding(0.1)

		//copy .main to others
		stls.Button.Hover = *b
		stls.Button.Touch_hover = *b
		stls.Button.Touch_out = *b
		stls.Button.Disable = *b

		stls.Button.Hover.Content_color = OsCd_Aprox(stls.Button.Main.Content_color, themeWhite(), 0.5)
		stls.Button.Touch_out.Content_color = stls.Button.Hover.Content_color

		stls.Button.Touch_hover.Content_color = themeBack()
		stls.Button.Touch_hover.Color = root.themeCd()

		stls.Button.DisableAuto()
		//stls.Button.Disable.Color = OsCd_Aprox(stls.Button.Main.Color, themeWhite(), 0.35)
		//stls.Button.Disable.Content_color = OsCd_Aprox(stls.Button.Main.Content_color, themeWhite(), 0.7)
	}

	{
		stls.ButtonLight = stls.Button
		a := byte(127)
		stls.ButtonLight.Main.Content_color.A = a
		stls.ButtonLight.Hover.Content_color.A = a
		stls.ButtonLight.Touch_hover.Content_color.A = a
		stls.ButtonLight.Touch_out.Content_color.A = a
		stls.ButtonLight.Disable.Content_color.A = a
		stls.ButtonLight.Disable.Color.A = a
	}

	{
		stls.ButtonAlpha = stls.Button
		stls.ButtonAlpha.Main.Content_color = OsCd{}
		stls.ButtonAlpha.Hover.Content_color = OsCd_Aprox(root.themeCd(), themeWhite(), 0.7)
		stls.ButtonAlpha.Touch_out.Content_color = OsCd{}
		stls.ButtonAlpha.DisableAuto()
		//stls.ButtonAlpha.Disable.Content_color = OsCd{}
		//stls.ButtonAlpha.Disable.Color = OsCd_Aprox(root.themeCd(), themeWhite(), 0.7)

	}

	{
		stls.ButtonMenu = stls.ButtonAlpha
		stls.ButtonMenu.FontAlignH(0)
	}

	{
		stls.ButtonMenuSelected = stls.Button
		stls.ButtonMenuSelected.FontAlignH(0)
	}

	{
		stls.ButtonBorder = stls.ButtonAlpha
		//stls.ButtonBorder.Margin(0.1)
		stls.ButtonBorder.Border(0.03)
		stls.ButtonBorder.BorderCd(root.themeCd())
	}

	{
		stls.ButtonLogo = stls.ButtonAlpha
		stls.ButtonLogo.Hover.Content_color.A = 0
		stls.ButtonLogo.Touch_hover.Content_color.A = 0
		stls.ButtonLogo.Touch_out.Content_color.A = 0 //refactor ...

		stls.ButtonLogo.Hover.Color = OsCd_Aprox(stls.ButtonLogo.Main.Color, themeWhite(), 0.5)
		stls.ButtonLogo.Touch_hover.Color = root.themeCd()
		stls.ButtonLogo.Touch_out.Color = stls.ButtonLogo.Hover.Color
	}

	{
		stls.ButtonDanger = stls.Button
		stls.ButtonDanger.Main.Content_color = themeWarning()
		stls.ButtonDanger.Hover.Content_color = themeWarning()
		stls.ButtonDanger.Touch_hover.Content_color = themeWarning()
		stls.ButtonDanger.Touch_out.Content_color = themeWarning()
		stls.ButtonDanger.DisableAuto()
		//stls.ButtonDanger.Disable.Content_color = OsCd_Aprox(stls.ButtonDanger.Main.Content_color, themeWhite(), 0.5)
	}

	{
		stls.ButtonDangerMenu = stls.ButtonDanger
		stls.ButtonDangerMenu.FontAlignH(0)
	}

	{
		b := &stls.Text.Main
		b.Cursor = "ibeam"
		b.Color = themeBlack()
		b.Image_alignV = 1
		b.Image_alignH = 0
		b.Font_path = SKYALT_FONT_PATH
		b.Font_alignV = 1
		b.Font_alignH = 0
		b.Font_height = SKYALT_FONT_HEIGHT
		b.Font_formating = true
		b.Padding(0.16)

		//copy .main to others
		stls.Text.Hover = *b
		stls.Text.Touch_hover = *b
		stls.Text.Touch_out = *b
		stls.Text.Disable = *b

		//stls.Text.Disable.Color = OsCd_Aprox(stls.Text.Main.Color, themeWhite(), 0.35)
		stls.Text.DisableAuto()
	}

	{
		stls.TextCenter = stls.Text
		stls.TextCenter.FontAlignH(1)
	}

	{
		stls.TextRight = stls.Text
		stls.TextRight.FontAlignH(2)
	}

	{
		stls.TextErr = stls.Text
		stls.TextErr.Color(themeWarning())
	}

	{
		b := &stls.Editbox.Main
		b.Cursor = "ibeam"
		b.Color = themeBlack()
		b.Image_alignV = 1
		b.Image_alignH = 0
		b.Font_path = SKYALT_FONT_PATH
		b.Font_alignV = 1
		b.Font_alignH = 0
		b.Font_height = SKYALT_FONT_HEIGHT
		b.Font_formating = true
		b.Margin(0.03)
		b.Border(0.03)
		b.Padding(0.1)
		b.Border_color = themeBlack()

		//copy .main to others
		stls.Editbox.Hover = *b
		stls.Editbox.Touch_hover = *b
		stls.Editbox.Touch_out = *b
		stls.Editbox.Disable = *b

		cd := themeGrey(0.9)
		stls.Editbox.Hover.Content_color = cd
		stls.Editbox.Touch_hover.Content_color = cd
		stls.Editbox.Touch_out.Content_color = cd
		stls.Editbox.DisableAuto()
		//stls.Editbox.Disable.Color = OsCd_Aprox(stls.Editbox.Main.Color, themeWhite(), 0.35)
	}

	{
		stls.EditboxErr = stls.Editbox
		stls.EditboxErr.Main.Content_color = themeWarning()
		stls.EditboxErr.Hover.Content_color = themeWarning()
		stls.EditboxErr.Touch_hover.Content_color = themeWarning()
		stls.EditboxErr.Touch_out.Content_color = themeWarning()
		stls.EditboxErr.DisableAuto()
		//stls.EditboxErr.Disable.Content_color = OsCd_Aprox(stls.ButtonDanger.Main.Content_color, themeWhite(), 0.5)
	}

	{
		stls.EditboxYellow = stls.Editbox
		stls.EditboxYellow.Main.Content_color = themeEdit()
		stls.EditboxYellow.Hover.Content_color = themeEdit()
		stls.EditboxYellow.Touch_hover.Content_color = themeEdit()
		stls.EditboxYellow.Touch_out.Content_color = themeEdit()
		stls.EditboxYellow.DisableAuto()
		//stls.EditboxYellow.Disable.Content_color = OsCd_Aprox(stls.EditboxYellow.Main.Content_color, themeEdit(), 0.5)
	}

	{
		stls.Combo = stls.Editbox
		stls.Combo.Cursor("hand")
	}

	{
		//frame
		stls.ProgressFrame.Border(0.03)
		stls.ProgressFrame.BorderCd(root.themeCd())
		stls.ProgressFrame.DisableAuto()
		//stls.ProgressStatus.Disable.Color = OsCd_Aprox(stls.ProgressStatus.Main.Color, themeWhite(), 0.35)
		//stls.ProgressStatus.Disable.Content_color = OsCd_Aprox(stls.ProgressStatus.Main.Content_color, themeWhite(), 0.5)

		//status
		stls.ProgressStatus.Margin(0.1)
		stls.ProgressStatus.Padding(0.1)
		stls.ProgressStatus.ContentCd(root.themeCd())
		stls.ProgressStatus.Color(OsCd_black())
		stls.ProgressStatus.FontPath(SKYALT_FONT_PATH)
		stls.ProgressStatus.FontH(0.35)
		stls.ProgressStatus.FontFormating(true)
		stls.ProgressStatus.FontAlignH(2) //right
		stls.ProgressStatus.FontAlignV(1)
		stls.ProgressStatus.DisableAuto()
		//stls.ProgressStatus.Disable.Content_color = OsCd_Aprox(stls.ProgressStatus.Main.Content_color, themeWhite(), 0.5)

	}

	{
		stls.SliderTrack.MaxHeight(0.1)
		stls.SliderTrack.MaxHeightAlign(1)
		stls.SliderTrack.ContentCd(root.themeCd())
		stls.SliderTrack.DisableAuto()

		stls.SliderThumb.MaxWidth(0.4)
		stls.SliderThumb.MaxHeight(0.4)
		stls.SliderThumb.MaxWidthAlign(0)
		stls.SliderThumb.MaxHeightAlign(1)
		stls.SliderThumb.Radius(1000) //circle
		stls.SliderThumb.ContentCd(root.themeCd())
		stls.SliderThumb.HoverAuto()
		stls.SliderThumb.DisableAuto()
	}

	{
		stls.CheckboxCheck.MaxWidth(0.5)
		stls.CheckboxCheck.MaxHeight(0.5)
		stls.CheckboxCheck.MaxWidthAlign(1)
		stls.CheckboxCheck.MaxHeightAlign(1)
		stls.CheckboxCheck.Border(0.03)
		stls.CheckboxCheck.BorderCd(themeBlack())
		stls.CheckboxCheck.Padding(0.03)
		stls.CheckboxCheck.HoverAuto()
		stls.CheckboxCheck.DisableAuto()

		stls.CheckboxLabel = stls.Text
		stls.CheckboxLabel.Cursor("hand")

	}

	return stls
}

type DivStyles struct {
	styles []*CompStyle
	theme  int
}

func NewDivStyles() *DivStyles {
	var stls DivStyles

	stls.theme = -1

	stls.Add(&CompStyle{}) //empty id=0

	return &stls
}

func (stls *DivStyles) Get(i uint32) *CompStyle {
	if i > 0 && int(i) < len(stls.styles) {
		return stls.styles[i]
	}
	return nil
}

func (stls *DivStyles) Add(st *CompStyle) int {
	stls.styles = append(stls.styles, st)
	return len(stls.styles) - 1
}

func (stls *DivStyles) AddJs(js []byte) (int, error) {

	var div CompStyle
	err := json.Unmarshal(js, &div)
	if err != nil {
		return -1, fmt.Errorf("Unmarshal() failed: %w", err)
	}

	return stls.Add(&div), nil
}
