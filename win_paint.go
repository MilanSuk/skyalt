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
)

type WinPaintBuff struct {
	win  *Win
	crop OsV4

	level_depth int
	depth       int
}

func NewWinPaintBuff(win *Win) *WinPaintBuff {
	var b WinPaintBuff
	b.win = win
	return &b
}

func (b *WinPaintBuff) Destroy() {
}

func (b *WinPaintBuff) StartLevel(crop OsV4, backCd color.RGBA, greyBack OsV4) {
	b.level_depth = b.depth

	//grey back - maybe do it as large shadow ...
	if greyBack.Is() {
		b.depth++
		b.win.SetClipRect(greyBack)
		b.win.DrawRect(greyBack.Start, greyBack.End(), b.depth, color.RGBA{0, 0, 0, 80}) //grey
	}

	//drawBack
	b.AddCrop(crop)
	b.depth++
	b.AddRect(crop, backCd, 0) //depth=100

	//prepare for layout
	b.depth++
}

func (b *WinPaintBuff) FinalDraw() {

	win := b.win.GetScreenCoord()
	b.win.SetClipRect(win)

	b.depth = 0
}

func (b *WinPaintBuff) AddCrop(crop OsV4) OsV4 {
	b.win.SetClipRect(crop)
	ret := b.crop
	b.crop = crop
	return ret
}

func (b *WinPaintBuff) AddRect(coord OsV4, cd color.RGBA, thick int) {
	start := coord.Start
	end := coord.End()
	if thick == 0 {
		b.win.DrawRect(start, end, b.depth, cd)
	} else {
		b.win.DrawRect_border(start, end, b.depth, cd, thick)
	}

}
func (b *WinPaintBuff) AddRectRound(coord OsV4, rad int, cd color.RGBA, thick int) {
	b.win.DrawRectRound(coord, rad, b.depth, cd, thick, false)
}
func (b *WinPaintBuff) AddRectRoundGrad(coord OsV4, rad int, cd color.RGBA, thick int) {
	b.win.DrawRectRound(coord, rad, b.depth, cd, thick, true)
}

func (b *WinPaintBuff) AddLine(start OsV2, end OsV2, cd color.RGBA, thick int) { //cd_over, cd_down ...
	v := end.Sub(start)
	if !v.IsZero() {
		b.win.DrawLine(start, end, b.depth, thick, cd)
	}
}

func (buf *WinPaintBuff) AddBezier(a OsV2, b OsV2, c OsV2, d OsV2, cd color.RGBA, thick int, dash_len float32, move float32) {
	buf.win.DrawBezier(a, b, c, d, buf.depth, thick, cd, dash_len, move)
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

func (b *WinPaintBuff) AddImage(path WinMedia, screen OsV4, cd color.RGBA, align OsV2, Translate_x, Translate_y, Scale_x, Scale_y *float64, errCd color.RGBA, cell int) *WinImage {
	img, err := b.win.AddImage(path)
	if err != nil {
		prop := InitWinFontPropsDef(cell)
		//props.frontCd = b.win.io.GetPalette().E
		b.AddText(path.GetString()+" has error", prop, errCd, screen, OsV2{1, 1}, 0, 1)
		return nil
	}

	origSize := img.origSize

	//position
	q := screen
	{
		fill := OsV2_OutRatio(screen.Size, origSize)
		fit := OsV4_center(screen, OsV2_InRatio(screen.Size, origSize))

		if *Scale_x < 0 {
			//fill
			q.Size.X = fill.X //from layout
		} else if *Scale_x == 0 {
			//fit
			q.Start.X = fit.Start.X
			q.Size.X = fit.Size.X //from layout
		} else {
			q.Size.X = int(float64(origSize.X) * *Scale_x) //from orig
		}

		if *Scale_y < 0 {
			//fill
			q.Size.Y = fill.Y //from layout
		} else if *Scale_y == 0 {
			//fit
			q.Start.Y = fit.Start.Y
			q.Size.Y = fit.Size.Y //from layout
		} else {
			q.Size.Y = int(float64(origSize.Y) * *Scale_y) //from orig
		}

		//align
		if *Scale_x <= 0 {
			if align.X == 0 {
				q.Start.X = screen.Start.X
			} else if align.X == 1 {
				q.Start.X = OsV4_centerFull(screen, q.Size).Start.X
			} else if align.X == 2 {
				q.Start.X = screen.End().X - q.Size.X
			}
		}
		if *Scale_y <= 0 {
			if align.Y == 0 {
				q.Start.Y = screen.Start.Y
			} else if align.Y == 1 {
				q.Start.Y = OsV4_centerFull(screen, q.Size).Start.Y
			} else if align.Y == 2 {
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
				q.Start.X = OsV4_centerFull(screen, q.Size).Start.X //smaller than screen => auto-center
			}
			if q.Size.Y > screen.Size.Y {
				min_y -= (q.Size.Y - screen.Size.Y)
				max_y += (q.Size.Y - screen.Size.Y)
			} else {
				q.Start.Y = OsV4_centerFull(screen, q.Size).Start.Y //smaller than screen => auto-center
			}

			q.Start.X = OsClamp(q.Start.X, min_x, max_x)
			q.Start.Y = OsClamp(q.Start.Y, min_y, max_y)
		}

		*Scale_x = float64(q.Size.X) / float64(origSize.X)
		*Scale_y = float64(q.Size.Y) / float64(origSize.Y)
		*Translate_x = float64(q.Start.X - screen.Start.X)
		*Translate_y = float64(q.Start.Y - screen.Start.Y)
	}

	//draw image
	imgRectBackup := b.AddCrop(b.crop.GetIntersect(screen))
	err = img.Draw(q, b.depth, cd)
	if err != nil {
		fmt.Printf("Draw() failed: %v\n", err)
	}
	b.AddCrop(imgRectBackup)

	return img
}

func (b *WinPaintBuff) AddText(ln string, prop WinFontProps, frontCd color.RGBA, coord OsV4, align OsV2, yLine, num_lines int) {

	imgRectBackup := b.AddCrop(b.crop.GetIntersect(coord))

	b.win.DrawText(ln, prop, frontCd, coord, b.depth, align, yLine, num_lines)

	b.AddCrop(imgRectBackup)
}

func (b *WinPaintBuff) AddTextBack(rangee OsV2, ln string, prop WinFontProps, coord OsV4, cd color.RGBA, align OsV2, underline bool, yLine, num_lines int) {
	if rangee.X == rangee.Y {
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
	}
}

func (b *WinPaintBuff) AddTextCursor(text string, prop WinFontProps, coord OsV4, align OsV2, cursorPos int, yLine, numLines int, cd color.RGBA, cell int) OsV4 {
	b.win.cursorEdit = true

	start := b.win.GetTextStartLine(text, prop, coord, align, numLines)
	start.Y += yLine * prop.lineH

	rngX := b.win.GetTextSize(cursorPos, text, prop).X

	c := InitOsV4(start.X+rngX, start.Y, OsMax(1, cell/15), prop.lineH)

	cd.A = b.win.cursorCdA
	b.AddRect(c, cd, 0)

	return c
}
