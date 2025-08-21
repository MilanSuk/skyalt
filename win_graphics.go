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
	"image"
	"image/color"
	"image/draw"
	"unicode/utf8"

	"slices"

	"github.com/fogleman/gg"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type WinGphItem struct {
	realSize     OsV2
	lastDrawTick int64

	alpha   *image.Alpha
	texture *WinTexture
}

func NewWinGphItemAlpha(alpha *image.Alpha, realSize OsV2) *WinGphItem {
	it := &WinGphItem{realSize: realSize}
	it.alpha = alpha
	return it
}

func (it *WinGphItem) getTexture() *WinTexture {
	if it.texture == nil {
		it.texture = InitWinTextureFromImageAlpha(it.alpha)
		it.alpha = nil //not needed anymore
	}
	return it.texture
}

func (it *WinGphItem) Destroy() {
	if it.texture != nil {
		it.texture.Destroy()
	}
}

func (it *WinGphItem) IsUsed(gph *WinGph) bool {
	return gph.tick_now < (it.lastDrawTick + 30000) //30 sec
}
func (it *WinGphItem) UpdateTick(gph *WinGph) {
	it.lastDrawTick = gph.tick_now
}

func (it *WinGphItem) DrawPointsUV(pts [4]OsV2f, uvs [4]OsV2f, depth int, cd color.RGBA, gph *WinGph) error {

	texture := it.getTexture()
	if texture != nil {
		norm := OsV2f{float32(it.realSize.X) / float32(texture.size.X), float32(it.realSize.Y) / float32(texture.size.Y)}
		for i := 0; i < 4; i++ {
			uvs[i] = uvs[i].Mul(norm)
		}
		texture.DrawPointsUV(pts, uvs, depth, cd)
	}

	it.UpdateTick(gph)
	return nil
}

func (it *WinGphItem) DrawCut(coord OsV4, depth int, cd color.RGBA, gph *WinGph) error {
	texture := it.getTexture()
	if texture != nil {
		uv := OsV2f{
			float32(coord.Size.X) / float32(texture.size.X),
			float32(coord.Size.Y) / float32(texture.size.Y)}
		texture.DrawQuadUV(coord, depth, cd, OsV2f{}, uv)
	}

	it.UpdateTick(gph)
	return nil
}

func (it *WinGphItem) DrawCutEx(coord OsV4, textStart OsV2, depth int, cd color.RGBA, gph *WinGph) error {

	texture := it.getTexture()
	if texture != nil {
		uvS := OsV2f{
			float32(textStart.X) / float32(texture.size.X),
			float32(textStart.Y) / float32(texture.size.Y)}
		uvE := OsV2f{
			float32(textStart.X+coord.Size.X) / float32(texture.size.X),
			float32(textStart.Y+coord.Size.Y) / float32(texture.size.Y)}
		texture.DrawQuadUV(coord, depth, cd, uvS, uvE)
	}

	it.UpdateTick(gph)
	return nil
}

func (it *WinGphItem) DrawCutCds(coord OsV4, depth int, defCd color.RGBA, cds []WinGphItemTextCd, gph *WinGph) error {
	texture := it.getTexture()
	if texture != nil {
		if len(cds) > 0 {
			last_pos := 0
			for i := 1; i < len(cds); i++ {
				cd := cds[i-1].cd
				if cd.A == 0 {
					cd = defCd
				}

				var cq OsV4
				cq.Start.X = coord.Start.X + last_pos
				cq.Start.Y = coord.Start.Y
				cq.Size.X = (cds[i].pos - last_pos)
				cq.Size.Y = coord.Size.Y

				it.DrawCutEx(cq, OsV2{last_pos, 0}, depth, cd, gph)

				last_pos = cds[i].pos
			}

			//last
			cd := cds[len(cds)-1].cd
			if cd.A == 0 {
				cd = defCd
			}

			var cq OsV4
			cq.Start.X = coord.Start.X + last_pos
			cq.Start.Y = coord.Start.Y
			cq.Size.X = (texture.size.X - last_pos)
			cq.Size.Y = coord.Size.Y

			it.DrawCutEx(cq, OsV2{last_pos, 0}, depth, cd, gph)
		} else {

			uv := OsV2f{
				float32(coord.Size.X) / float32(texture.size.X),
				float32(coord.Size.Y) / float32(texture.size.Y)}

			texture.DrawQuadUV(coord, depth, defCd, OsV2f{}, uv)
		}
	}

	it.UpdateTick(gph)
	return nil
}

func (it *WinGphItem) DrawUV(coord OsV4, depth int, cd color.RGBA, sUV, eUV OsV2f, gph *WinGph) error {
	texture := it.getTexture()
	if texture != nil {
		szUv := OsV2f{
			float32(it.realSize.X) / float32(texture.size.X),
			float32(it.realSize.Y) / float32(texture.size.Y)}

		//normalize by item_size
		sUV = sUV.Mul(szUv)
		eUV = eUV.Mul(szUv)

		texture.DrawQuadUV(coord, depth, cd, sUV, eUV)
	}

	it.UpdateTick(gph)
	return nil
}

type WinGphItemTextCd struct {
	cd  color.RGBA
	pos int
}
type WinGphItemText struct {
	item *WinGphItem
	cds  []WinGphItemTextCd

	size OsV2

	prop WinFontProps
	text string

	letters []int //aggregated!	Represent every byte(multi_byte character = same number for each byte)
	skips   []bool
}

func (txt *WinGphItemText) GetRangePx(st, en int) int {
	if st >= en {
		return 0
	}
	stPx := 0
	if st > 0 {
		stPx = txt.letters[st-1]
	}
	return txt.letters[en-1] - stPx
}

type WinGphLine struct {
	s, e int
}

type WinGphItemTextMax struct {
	const_max_line_px int //const_max_line_px > 0 => line_wrapping is enabled
	prop              WinFontProps
	text              string

	lines      []WinGphLine
	max_size_x int
}

func WinGph_CursorLineY(lines []WinGphLine, cursor int) int {
	for i, p := range lines {
		if cursor <= p.e {
			return i
		}
	}
	return len(lines) - 1
}

func WinGph_PosLineRange(lines []WinGphLine, i int) (int, int) {
	return lines[i].s, lines[i].e
}

func WinGph_CursorLineRange(lines []WinGphLine, cursor int) (int, int) {
	i := WinGph_CursorLineY(lines, cursor)
	return lines[i].s, lines[i].e
}

func WinGph_CursorLine(text string, lines []WinGphLine, cursor int) (string, int) {
	s, e := WinGph_CursorLineRange(lines, cursor)
	return text[s:e], cursor - s
}

type WinGphItemCircle struct {
	item  *WinGphItem
	size  OsV2
	width float64
	arc   OsV2f
	grad  bool
}
type WinGphItemPoly struct {
	item   *WinGphItem
	points []OsV2f
	size   OsV2
	width  float64
}

type WinGphItemRoundedRectangle struct {
	item  *WinGphItem
	size  OsV2
	width float64
	rad   float64
}

func (poly *WinGphItemPoly) CmpPoints(points []OsV2f) bool {
	if len(poly.points) != len(points) {
		return false
	}

	for i, a := range poly.points {
		if !a.Cmp(points[i]) {
			return false
		}
	}
	return true
}

type WinGph struct {
	fonts []*WinFont //array index = textH

	texts             map[string][]*WinGphItemText
	textMaxs          map[string][]*WinGphItemTextMax
	circles           []*WinGphItemCircle
	polys             []*WinGphItemPoly
	roundedRectangles []*WinGphItemRoundedRectangle

	texts_num_created int
	texts_num_remove  int

	tick_now int64
}

func NewWinGph() *WinGph {
	gph := &WinGph{}
	gph.texts = make(map[string][]*WinGphItemText)
	gph.textMaxs = make(map[string][]*WinGphItemTextMax)
	return gph
}
func (gph *WinGph) Destroy() {

	for _, it := range gph.fonts {
		it.Destroy()
	}

	for _, it := range gph.circles {
		it.item.Destroy()
	}

	for _, itArr := range gph.texts {
		for _, it := range itArr {
			it.item.Destroy()
		}
	}
	for _, it := range gph.polys {
		it.item.Destroy()
	}
}

func (gph *WinGph) Maintenance() {

	gph.tick_now = OsTicks()

	for i := len(gph.fonts) - 1; i >= 0; i-- {
		gph.fonts[i].Maintenance()
	}

	for i := len(gph.circles) - 1; i >= 0; i-- {
		if !gph.circles[i].item.IsUsed(gph) {
			gph.circles[i].item.Destroy()
			gph.circles = slices.Delete(gph.circles, i, i+1)
		}
	}

	for i := len(gph.polys) - 1; i >= 0; i-- {
		if !gph.polys[i].item.IsUsed(gph) {
			gph.polys[i].item.Destroy()
			gph.polys = slices.Delete(gph.polys, i, i+1)
		}
	}

	for textId, itArr := range gph.texts {
		for i := len(itArr) - 1; i >= 0; i-- {
			if !itArr[i].item.IsUsed(gph) {
				itArr[i].item.Destroy()
				itArr = slices.Delete(itArr, i, i+1)
				gph.texts[textId] = itArr

				gph.texts_num_remove++
			}
		}
	}
}

func (gph *WinGph) GetFont(prop *WinFontProps) *WinFont {

	for i := len(gph.fonts); i < prop.textH+1; i++ {
		gph.fonts = append(gph.fonts, &WinFont{})
	}

	return gph.fonts[prop.textH]
}

func (gph *WinGph) GetTextSize(prop WinFontProps, cur int, text string) OsV2 {
	if text == "" {
		return OsV2{}
	}

	it := gph.GetText(prop, text)
	if it == nil {
		return OsV2{0, 0}
	}
	it.item.UpdateTick(gph)

	if cur < 0 || cur >= len(it.letters) {
		return it.size
	}

	last_i := cur - 1
	if last_i < 0 {
		return OsV2{0, it.size.Y}
	}

	if last_i >= len(it.letters) {
		last_i = len(it.letters) - 1
	}

	return OsV2{it.letters[last_i], it.size.Y}
}

func (gph *WinGph) GetTextPos(prop WinFontProps, curr_px int, text string, roundToClosest bool) int {
	if text == "" {
		return 0
	}

	it := gph.GetText(prop, text)
	if it == nil {
		return 0
	}
	it.item.UpdateTick(gph)

	last_ad := 0
	for i, ad := range it.letters {
		if ad > curr_px {
			if roundToClosest {
				if (ad - curr_px) < (curr_px - last_ad) {
					return OsMin(len(it.letters), i+1) //next
				}
			}
			return i
		}
		last_ad = ad
	}
	return len(it.letters)
}

func (gph *WinGph) GetText(prop WinFontProps, text string) *WinGphItemText {
	if text == "" {
		return nil
	}

	//find
	arr := gph.texts[text]
	for _, it := range arr {
		if it.prop.Cmp(&prop) && it.text == text {
			it.item.UpdateTick(gph)
			return it
		}
	}

	//create
	it := gph.drawString(prop, text)
	if it != nil {
		gph.texts[text] = append(gph.texts[text], it)
		gph.texts_num_created++
	}
	return it
}

func (gph *WinGph) GetTextMax(str string, const_max_line_px int, prop WinFontProps) *WinGphItemTextMax {
	if str == "" {
		return nil
	}

	//find
	{
		arr := gph.textMaxs[str]
		for _, it := range arr {
			if it.prop.Cmp(&prop) /*&& it.text == str*/ && it.const_max_line_px == const_max_line_px {
				return it
			}
		}
	}

	//build lines
	txt := gph.GetText(prop, str)
	var lines []WinGphLine
	start_p := 0
	for p, ch := range str {
		chSz := len(string(ch))

		if ch == '\n' {
			lines = append(lines, WinGphLine{s: start_p, e: p})
			start_p = p + chSz
		} else {
			px := txt.GetRangePx(start_p, p+chSz) //p+1

			if const_max_line_px > 0 && px > const_max_line_px {
				s, _ := _UiText_CursorWordRange(str, p-1) //p-1
				for s > 0 && txt.skips[s-1] {
					s--
				}
				if start_p < s {
					lines = append(lines, WinGphLine{s: start_p, e: s})
					start_p = s
				}
			}
		}
	}

	if len(lines) == 0 || start_p < len(str) || (str != "" && str[len(str)-1] == '\n') {
		lines = append(lines, WinGphLine{s: start_p, e: len(str)})
	}

	//compute max width from all lines
	max_size_x := 0
	for _, ln := range lines {
		max_size_x = OsMax(max_size_x, txt.GetRangePx(ln.s, ln.e))
	}

	it := &WinGphItemTextMax{text: str, const_max_line_px: const_max_line_px, prop: prop, lines: lines, max_size_x: max_size_x}
	gph.textMaxs[str] = append(gph.textMaxs[str], it)

	return it
}

func (gph *WinGph) GetCircle(size OsV2, width float64, arc OsV2f) *WinGphItemCircle {

	//find
	for _, it := range gph.circles {
		if it.size.Cmp(size) && it.width == width && it.arc.Cmp(arc) && !it.grad {
			return it
		}
	}

	//create
	w := OsNextPowOf2(size.X)
	h := OsNextPowOf2(size.Y)

	dc := gg.NewContext(w, h)
	dc.SetRGBA255(255, 255, 255, 255)

	rx := float64(size.X) / 2
	ry := float64(size.Y) / 2
	sx := rx
	sy := ry

	rx -= width //can be zero
	ry -= width

	if arc.X == 0 && arc.Y == 0 {
		dc.DrawEllipse(sx, sy, rx, ry)
	} else {
		dc.NewSubPath()
		dc.MoveTo(sx, sy) //LineTo
		dc.DrawEllipticalArc(sx, sx, rx, ry, float64(arc.X), float64(arc.Y))
		dc.ClosePath()
	}

	if width > 0 {
		dc.SetLineWidth(width)
		dc.Stroke()
	} else {
		dc.Fill()
	}

	//dc.SavePNG("out.png")

	rect := image.Rect(0, 0, w, h)
	dst := image.NewAlpha(rect)
	draw.Draw(dst, rect, dc.Image(), rect.Min, draw.Src)

	//add
	var circle *WinGphItemCircle
	it := NewWinGphItemAlpha(dst, size)
	if it != nil {
		circle = &WinGphItemCircle{item: it, size: size, width: width, arc: arc}
		gph.circles = append(gph.circles, circle)
	}
	return circle
}

func (gph *WinGph) GetRoundedRectangle(width float64, rad float64) *WinGphItemRoundedRectangle {

	size := OsV2{3 * int(rad), 3 * int(rad)}

	//find
	for _, it := range gph.roundedRectangles {
		if it.size.Cmp(size) && it.width == width && it.rad == rad {
			return it
		}
	}

	//create
	w := OsNextPowOf2(size.X)
	h := OsNextPowOf2(size.Y)

	dc := gg.NewContext(w, h)
	dc.SetRGBA255(255, 255, 255, 255)

	dc.DrawRoundedRectangle(width/2, width/2, (3*rad)-width, (3*rad)-width, rad)

	if width > 0 {
		dc.SetLineWidth(width)
		dc.Stroke()
	} else {

		//for shadow will need custom style ....

		dc.Fill()
	}

	//dc.SavePNG("out.png")

	rect := image.Rect(0, 0, w, h)
	dst := image.NewAlpha(rect)
	draw.Draw(dst, rect, dc.Image(), rect.Min, draw.Src)

	//add
	var roundedRect *WinGphItemRoundedRectangle
	it := NewWinGphItemAlpha(dst, size)
	if it != nil {
		roundedRect = &WinGphItemRoundedRectangle{item: it, size: size, width: width, rad: rad}
		gph.roundedRectangles = append(gph.roundedRectangles, roundedRect)
	}
	return roundedRect
}

func (gph *WinGph) GetCircleGrad(size OsV2, arc OsV2f) *WinGphItemCircle {

	//find
	for _, it := range gph.circles {
		if it.size.Cmp(size) && it.width == 0 && it.arc.Cmp(arc) && it.grad {
			return it
		}
	}

	//create
	w := OsNextPowOf2(size.X)
	h := OsNextPowOf2(size.Y)

	dc := gg.NewContext(w, h)

	rx := float64(size.X) / 2
	ry := float64(size.Y) / 2

	grad := gg.NewRadialGradient(rx, ry, 0, rx, ry, rx)
	grad.AddColorStop(0, color.RGBA{255, 255, 255, 255})
	grad.AddColorStop(0.5, color.RGBA{255, 255, 255, 255})
	grad.AddColorStop(1, color.RGBA{255, 255, 255, 0})

	dc.SetFillStyle(grad)
	dc.DrawRectangle(0, 0, float64(size.X), float64(size.Y))
	dc.Fill()

	//dc.SavePNG("out.png")

	rect := image.Rect(0, 0, w, h)
	dst := image.NewAlpha(rect)
	draw.Draw(dst, rect, dc.Image(), rect.Min, draw.Src)

	//add
	var circle *WinGphItemCircle
	it := NewWinGphItemAlpha(dst, size)
	if it != nil {
		circle = &WinGphItemCircle{item: it, size: size, width: 0, arc: arc, grad: true}
		gph.circles = append(gph.circles, circle)
	}
	return circle
}

func (gph *WinGph) GetPoly(points []OsV2f, width float64) *WinGphItemPoly {
	if len(points) == 0 {
		return nil
	}

	//find
	for _, it := range gph.polys {
		if it.width == width && it.CmpPoints(points) {
			return it
		}
	}

	//get size
	min := points[0]
	max := points[0]
	for _, p := range points {
		min = min.Min(p)
		max = max.Max(p)
	}

	min = min.toV2().toV2f()
	max = max.toV2().Add(OsV2{1, 1}).toV2f()

	size := max.Sub(min).toV2()
	if size.X == 0 || size.Y == 0 {
		return nil
	}
	w := OsNextPowOf2(size.X)
	h := OsNextPowOf2(size.Y)

	//create
	dc := gg.NewContext(w, h)
	for _, p := range points {
		dc.LineTo(float64(p.X-min.X), float64(p.Y-min.Y))
	}
	dc.ClosePath()

	if width > 0 {
		dc.SetLineWidth(width)
		dc.Stroke()
	} else {
		dc.Fill()
	}

	//dc.SavePNG("out.png")

	rect := image.Rect(0, 0, w, h)
	dst := image.NewAlpha(rect)
	draw.Draw(dst, rect, dc.Image(), rect.Min, draw.Src)

	//add
	var poly *WinGphItemPoly
	it := NewWinGphItemAlpha(dst, size)
	if it != nil {
		poly = &WinGphItemPoly{item: it, points: points, size: size, width: width}
		gph.polys = append(gph.polys, poly)
	}
	return poly

}

func (gph *WinGph) getFixedWidth(prop *WinFontProps) (fixedAdvance fixed.Int26_6) {
	//prop.fixed_width = true
	if prop.fixed_width {
		fixedAdvance, _ = gph.GetFont(prop).GetFace(prop).face.GlyphAdvance('M')
	}
	return
}

func (gph *WinGph) GetStringSize(prop WinFontProps, str string) (OsV2, fixed.Int26_6) {
	fixedAdvance := gph.getFixedWidth(&prop)

	var w fixed.Int26_6 //round to int after!
	prevCh := rune(-1)

	var maxH int
	var maxAscent fixed.Int26_6
	skip := 0
	act_prop := prop
	i := 0
	var cd color.RGBA
	for p, ch := range str {
		if prop.formating && !prop.processLetter(str[p:], &act_prop, &cd, &skip) {
			i++
			continue
		}

		isTab := (ch == '\t')
		if isTab {
			ch = ' '
		}

		face := gph.GetFont(&act_prop).GetFace(&act_prop).face

		var advance fixed.Int26_6
		if prop.fixed_width {
			advance = fixedAdvance
		} else {
			if prevCh >= 0 {
				w += face.Kern(prevCh, ch)
			}
			advance, _ = face.GlyphAdvance(ch)
		}

		if isTab {
			advance *= 8
		}

		w += advance
		prevCh = ch

		m := face.Metrics()
		maxH = OsMax(maxH, int(m.Ascent+m.Descent)>>6)
		if m.Ascent > maxAscent {
			maxAscent = m.Ascent
		}
		i++
	}

	return OsV2{int(w >> 6), maxH + 2}, maxAscent
}

func (gph *WinGph) drawString(prop WinFontProps, str string) *WinGphItemText {
	fixedAdvance := gph.getFixedWidth(&prop)

	size, maxAscent := gph.GetStringSize(prop, str)

	w := OsNextPowOf2(size.X)
	h := OsNextPowOf2(size.Y)

	a := image.NewAlpha(image.Rect(0, 0, w, h)) //[alpha]

	var letters []int
	var skips []bool
	d := &font.Drawer{
		Dst: a,                                                 //[alpha]
		Src: image.NewUniform(color.NRGBA{255, 255, 255, 255}), //[alpha]
		Dot: fixed.Point26_6{X: fixed.Int26_6(0), Y: fixed.Int26_6(maxAscent)},
	}

	prevCh := rune(-1)

	var cd color.RGBA
	var cds []WinGphItemTextCd

	word_count := 1
	skip := 0
	act_prop := prop
	i := 0
	for p, ch := range str {
		if prop.formating && !prop.processLetter(str[p:], &act_prop, &cd, &skip) {
			i++
			letters = append(letters, int(d.Dot.X>>6)) //same
			skips = append(skips, true)
			continue
		}

		if !OsIsTextWord(ch) && ch != ' ' && ch != '\t' {
			word_count++
		}

		if len(cds) == 0 || cds[len(cds)-1].cd != cd {
			cds = append(cds, WinGphItemTextCd{cd: cd, pos: int(d.Dot.X >> 6)})
		}

		isTab := (ch == '\t')
		if isTab {
			ch = ' '
		}

		if !prop.fixed_width {
			if prevCh >= 0 {
				d.Dot.X += d.Face.Kern(prevCh, ch)
			}
		}

		d.Face = gph.GetFont(&act_prop).GetFace(&act_prop).face
		dr, mask, maskp, advance, _ := d.Face.Glyph(d.Dot, ch)

		if prop.fixed_width {

			centeredDot := fixed.Point26_6{
				X: d.Dot.X + (fixedAdvance-advance)/2,
				Y: d.Dot.Y,
			}
			advance = fixedAdvance

			dr, mask, maskp, _, _ = d.Face.Glyph(centeredDot, ch)
		}
		if !dr.Empty() {
			draw.DrawMask(d.Dst, dr, d.Src, image.Point{}, mask, maskp, draw.Over)
		}

		if isTab {
			advance *= 8
		}

		d.Dot.X += advance

		//add letter and skip
		{
			letters = append(letters, int(d.Dot.X>>6))
			skips = append(skips, false)
			//other bytes
			_, n := utf8.DecodeRuneInString(string(ch))
			for ii := 1; ii < n; ii++ {
				letters = append(letters, int(d.Dot.X>>6)) //same
				skips = append(skips, false)
			}
		}

		prevCh = ch
	}

	return &WinGphItemText{item: NewWinGphItemAlpha(a, size), size: size, prop: prop, text: str, letters: letters, skips: skips, cds: cds} //[alpha]
}
