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
)

type WinPaintBuff struct {
	win  *Win
	crop OsV4

	depth int
}

func NewWinPaintBuff(win *Win) *WinPaintBuff {
	var b WinPaintBuff
	b.win = win
	return &b
}

func (b *WinPaintBuff) Destroy() {
}

func (b *WinPaintBuff) StartLevel(crop OsV4, backCd color.RGBA, greyBack OsV4, rounding int) {
	//grey back
	if greyBack.Is() {
		b.depth++
		b.win.render.SetClipRect(b.win.GetScreenCoord(), greyBack)

		//large shadow? ....
		b.win.render.DrawRect(greyBack.Start, greyBack.End(), b.depth, color.RGBA{0, 0, 0, 80}) //grey
	}

	//drawBack
	b.AddCrop(crop)
	b.depth++
	b.AddRectRound(crop, rounding, backCd, 0)

	//prepare for layout
	b.depth++
}

func (b *WinPaintBuff) FinalDraw() {

	win := b.win.GetScreenCoord()
	b.win.render.SetClipRect(b.win.GetScreenCoord(), win)

	b.depth = 0
}

func (b *WinPaintBuff) AddCrop(crop OsV4) OsV4 {
	b.win.render.SetClipRect(b.win.GetScreenCoord(), crop)
	ret := b.crop
	b.crop = crop
	return ret
}

func (b *WinPaintBuff) AddRect(coord OsV4, cd color.RGBA, thick int) {
	start := coord.Start
	end := coord.End()
	if thick == 0 {
		b.win.render.DrawRect(start, end, b.depth, cd)
	} else {
		b.win.DrawRect_border(start, end, b.depth, cd, thick)
	}
}

func (b *WinPaintBuff) AddRectRound(coord OsV4, rad int, cd color.RGBA, thick int) {
	if rad < 3 {
		b.AddRect(coord, cd, thick)
	} else {
		b.win.DrawRectRound(coord, rad, b.depth, cd, thick)
	}
}

func (b *WinPaintBuff) AddLine(start OsV2, end OsV2, cd color.RGBA, thick int) {
	v := end.Sub(start)
	if !v.IsZero() {
		b.win.render.DrawLine(start, end, b.depth, thick, cd)
	}
}
func (b *WinPaintBuff) AddBrush(start OsV2, points []OsV2, cd color.RGBA, thick int, renderEndings bool) {

	circle := b.win.gph.GetCircleGrad(OsV2{thick, thick}, OsV2f{})
	if circle == nil {
		return
	}

	thick = thick / 4
	if thick < 1 {
		thick = 1
	}

	for i := 1; i < len(points); i++ {
		ptA := points[i-1]
		ptB := points[i-0]

		v := ptB.Sub(ptA)
		num_steps := max(int(v.Len()/float32(thick)), 1)
		for ii := range num_steps {
			pt := ptA.Add(v.MulV(float32(ii) / float32(num_steps)))
			circle.item.DrawCut(InitOsV4Mid(pt, circle.size), b.depth, cd, b.win.gph)

		}
	}
}

func (buf *WinPaintBuff) AddBezier(a OsV2, b OsV2, c OsV2, d OsV2, cd color.RGBA, thick int, dash_len float32, move float32) {
	buf.win.render.DrawBezier(a, b, c, d, buf.depth, thick, cd, dash_len, move)
}

func (buf *WinPaintBuff) GetBezier(a OsV2, b OsV2, c OsV2, d OsV2, t float64) (OsV2f, OsV2f) {
	return buf.win.GetBezier(a, b, c, d, t)
}

func (buf *WinPaintBuff) GetPoly(points []OsV2f, width float64) *WinGphItemPoly {
	return buf.win.GetPoly(points, width)
}
func (buf *WinPaintBuff) AddPolyStart(start OsV2, poly *WinGphItemPoly, cd color.RGBA) {
	buf.win.DrawPolyStart(start, poly, buf.depth, cd)
}
func (buf *WinPaintBuff) AddPolyRect(rect OsV4, poly *WinGphItemPoly, cd color.RGBA) {
	buf.win.DrawPolyRect(rect, poly, buf.depth, cd)
}
func (buf *WinPaintBuff) AddPolyQuad(pts [4]OsV2f, uvs [4]OsV2f, poly *WinGphItemPoly, cd color.RGBA) {
	buf.win.DrawPolyQuad(pts, uvs, poly, buf.depth, cd)
}

func (b *WinPaintBuff) AddCircle(coord OsV4, cd color.RGBA, width int) {
	p := coord.Middle()
	b.win.DrawCicle(p, OsV2{coord.Size.X / 2, coord.Size.Y / 2}, b.depth, cd, width)
}

func (b *WinPaintBuff) AddImage(path WinImagePath, screen OsV4, cd color.RGBA, align OsV2, Translate_x, Translate_y, Scale_x, Scale_y *float64, errCd color.RGBA, cell int) {
	img := b.win.images.Add(path, nil)

	//origSize := img.origSize

	//position
	q := screen

	sz := img.loaded_size
	if sz.Is() {

		fill := OsV2_OutRatio(screen.Size, sz)
		fit := OsV4_center(screen, OsV2_InRatio(screen.Size, sz))

		if *Scale_x < 0 {
			//fill
			q.Size.X = fill.X //from layout
		} else if *Scale_x == 0 {
			//fit
			q.Start.X = fit.Start.X
			q.Size.X = fit.Size.X //from layout
		} else {
			q.Size.X = int(float64(sz.X) * *Scale_x) //from orig
		}

		if *Scale_y < 0 {
			//fill
			q.Size.Y = fill.Y //from layout
		} else if *Scale_y == 0 {
			//fit
			q.Start.Y = fit.Start.Y
			q.Size.Y = fit.Size.Y //from layout
		} else {
			q.Size.Y = int(float64(sz.Y) * *Scale_y) //from orig
		}

		//align
		if *Scale_x <= 0 {
			switch align.X {
			case 0:
				q.Start.X = screen.Start.X
			case 1:
				q.Start.X = OsV4_centerFull(screen, q.Size).Start.X
			case 2:
				q.Start.X = screen.End().X - q.Size.X
			}
		}
		if *Scale_y <= 0 {
			switch align.Y {
			case 0:
				q.Start.Y = screen.Start.Y
			case 1:
				q.Start.Y = OsV4_centerFull(screen, q.Size).Start.Y
			case 2:
				q.Start.Y = screen.End().Y - q.Size.Y
			}
		}

		if *Scale_x > 0 {
			q.Start.X = screen.Start.X + int(*Translate_x)
		}
		if *Scale_y > 0 {
			q.Start.Y = screen.Start.Y + int(*Translate_y)
		}

		//check translate boundaries
		{
			min_x := screen.Start.X
			min_y := screen.Start.Y
			max_x := min_x + (screen.Size.X - q.Size.X)
			max_y := min_y + (screen.Size.Y - q.Size.Y)

			if q.Size.X > screen.Size.X {
				min_x -= (q.Size.X - screen.Size.X)
				max_x += (q.Size.X - screen.Size.X)
			} else {
				if align.X == 1 {
					q.Start.X = OsV4_centerFull(screen, q.Size).Start.X //smaller than screen => auto-center
				}
			}
			if q.Size.Y > screen.Size.Y {
				min_y -= (q.Size.Y - screen.Size.Y)
				max_y += (q.Size.Y - screen.Size.Y)
			} else {
				if align.Y == 1 {
					q.Start.Y = OsV4_centerFull(screen, q.Size).Start.Y //smaller than screen => auto-center
				}
			}

			q.Start.X = OsClamp(q.Start.X, min_x, max_x)
			q.Start.Y = OsClamp(q.Start.Y, min_y, max_y)
		}

		*Scale_x = float64(q.Size.X) / float64(sz.X)
		*Scale_y = float64(q.Size.Y) / float64(sz.Y)
		*Translate_x = float64(q.Start.X - screen.Start.X)
		*Translate_y = float64(q.Start.Y - screen.Start.Y)
	}

	//draw image
	imgRectBackup := b.AddCrop(b.crop.GetIntersect(screen))
	alt := img.Draw(q, b.depth, cd, b.win)
	if alt != "" {
		b.win.DrawTextLine(alt, InitWinFontPropsDef(cell), errCd, q, b.depth, OsV2{1, 1}, 0, 1)
	}

	b.AddCrop(imgRectBackup)
}

func (b *WinPaintBuff) AddText(ln string, prop WinFontProps, frontCd color.RGBA, coord OsV4, align OsV2, yLine, num_lines int) {

	imgRectBackup := b.AddCrop(b.crop.GetIntersect(coord))

	b.win.DrawTextLine(ln, prop, frontCd, coord, b.depth, align, yLine, num_lines)

	b.AddCrop(imgRectBackup)
}

func (b *WinPaintBuff) AddTextBack(rangee OsV2, ln string, prop WinFontProps, coord OsV4, cd color.RGBA, align OsV2, underline bool, yLine, num_lines int, cell int) {
	if rangee.X == rangee.Y && num_lines <= 1 {
		return
	}

	start := b.win.GetTextStartLine(ln, prop, coord, align, num_lines)
	start.Y += yLine * prop.lineH

	var rng OsV2
	rng.X = b.win.GetTextSize(rangee.X, ln, prop).X
	rng.Y = b.win.GetTextSize(rangee.Y, ln, prop).X

	rng.Sort()

	if num_lines > 1 {
		coord.Size.Y = prop.lineH
	}

	if rng.X != rng.Y {
		if underline {
			Y := start.Y + coord.Size.Y
			b.AddRect(OsV4{Start: OsV2{start.X + rng.X, Y - 2}, Size: OsV2{rng.Y, 2}}, cd, 0)
		} else {
			c := InitOsV4(start.X+rng.X, start.Y, rng.Y-rng.X, prop.lineH)
			b.AddRect(c, cd, 0)
		}
	} else {
		//empty line
		c := InitOsV4(start.X+rng.X, start.Y, OsMax(2, cell/10), prop.lineH)
		b.AddRect(c, cd, 0)
	}
}

func WinPaintBuff_GetCursorWidth(cell int) int {
	return OsMax(1, cell/15)
}

func (b *WinPaintBuff) AddTextCursor(text string, prop WinFontProps, coord OsV4, align OsV2, cursorPos int, yLine, numLines int, cd color.RGBA, cell int) OsV4 {
	b.win.cursorEdit = true

	start := b.win.GetTextStartLine(text, prop, coord, align, numLines)
	start.Y += yLine * prop.lineH

	rngX := b.win.GetTextSize(cursorPos, text, prop).X

	c := InitOsV4(start.X+rngX, start.Y, WinPaintBuff_GetCursorWidth(cell), prop.lineH)

	cd.A = b.win.cursorCdA
	b.AddRect(c, cd, 0)

	return c
}
