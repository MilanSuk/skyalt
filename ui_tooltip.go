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
)

type UiTooltipContent struct {
	coord   OsV4
	priorUp bool
	text    string
	cd      color.RGBA
	force   bool
}

func (ctx *UiTooltipContent) Set(coord OsV4, priorUp bool, text string, cd color.RGBA, force bool) {
	ctx.coord = coord
	ctx.priorUp = priorUp
	ctx.cd = cd
	ctx.text = text
	ctx.force = force
}

func (a *UiTooltipContent) Cmp(b *UiTooltipContent) bool {
	return *a == *b
}

type UiTooltip struct {
	dom *Layout3

	contentSet  UiTooltipContent
	contentShow UiTooltipContent

	change_ticks int64
	show         bool
	coord        OsV4
}

func (p *UiTooltip) SetForce(coord OsV4, priorUp bool, text string, cd color.RGBA) {
	p.contentSet.Set(coord, priorUp, text, cd, true)
}

func (p *UiTooltip) Set(coord OsV4, priorUp bool, text string, cd color.RGBA) {
	if coord.Inside(p.dom.ui.GetWin().io.Touch.Pos) {
		p.contentSet.Set(coord, priorUp, text, cd, false)
	}
}

func (p *UiTooltip) isShow() bool {
	show := p.contentShow.text != "" && (p.contentShow.force || (p.change_ticks+200) <= OsTicks())
	return show

}
func (p *UiTooltip) NeedRedraw() bool {
	return !p.show && p.isShow()
}

func (p *UiTooltip) draw() {
	ui := p.dom.ui

	if !p.show {
		return
	}

	ctx := &p.contentShow

	ui.GetWin().buff.depth = 900
	//background
	{
		rc := p.coord.AddSpace(-ui.CellWidth(0.1))
		ui.GetWin().buff.AddCrop(rc)
		pl := ui.GetPalette()
		ui.GetWin().buff.AddRect(rc, pl.B, 0)
		ui.GetWin().buff.AddRect(rc, pl.OnB, 1)
	}

	prop := InitWinFontPropsDef(ui.Cell())
	ui._Text_draw(p.dom, p.coord, ctx.text, "", prop, ctx.cd, OsV2{0, 0}, false, false, true, true)
}

func (p *UiTooltip) touch() bool {
	redraw := false

	if !p.contentSet.Cmp(&p.contentShow) {
		p.contentShow = p.contentSet
		p.change_ticks = OsTicks()
		redraw = true
		p.show = false
	}

	show := p.isShow()
	if show != p.show {
		redraw = true

		if show {
			ui := p.dom.ui

			prop := InitWinFontPropsDef(ui.Cell())
			max_width_cells := int(p.dom._getWidth())
			if !strings.HasPrefix(p.contentShow.text, "http") {
				max_width_cells = OsMin(max_width_cells, 10)
			}
			mx, my := ui.GetTextSizeMax(p.contentShow.text, ui.Cell()*max_width_cells, prop)

			var final OsV4
			final.Start = ui.GetWin().io.Touch.Pos
			final.Size = OsV2{mx, my * prop.lineH}

			var orig OsV4
			orig.Start = ui.GetWin().io.Touch.Pos
			orig.Size = OsV2{ui.Cell() / 2, ui.Cell() / 2}
			orig = orig.AddSpace(-ui.Cell() / 5)

			p.coord = OsV4_relativeSurround(orig, final, ui.GetWin().GetScreenCoord(), p.contentShow.priorUp)
		}
	}
	p.show = show

	//reset
	p.contentSet = UiTooltipContent{}

	return redraw
}
