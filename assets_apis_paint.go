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

func (asset *Asset) getCellWidth(width float64) int {
	t := int(width * float64(asset.app.root.ui.Cell())) // cell is ~34
	if width > 0 && t <= 0 {
		t = 1 //at least 1px
	}
	return t
}

func (asset *Asset) addCoordMargin(q OsV4, margin float64, marginX float64, marginY float64) OsV4 {

	q = q.AddSpaceX(asset.getCellWidth(marginX))
	q = q.AddSpaceY(asset.getCellWidth(marginY))
	return q.AddSpace(asset.getCellWidth(margin))
}

func (asset *Asset) getCoord(x, y, w, h float64, margin float64, marginX float64, marginY float64) OsV4 {

	st := asset.app.root.levels.GetStack()
	layoutScreen := st.stack.canvas //st.stackLayout.CoordNoScroll()

	q := InitOsQuad(layoutScreen.Start.X+int(float64(layoutScreen.Size.X)*x),
		layoutScreen.Start.Y+int(float64(layoutScreen.Size.Y)*y),
		int(float64(layoutScreen.Size.X)*w),
		int(float64(layoutScreen.Size.Y)*h))

	return asset.addCoordMargin(q, margin, marginX, marginY)
}

func (asset *Asset) paint_rect(x, y, w, h float64, margin float64, cd OsCd, borderWidth float64) int64 {

	st := asset.app.root.levels.GetStack()
	if st.stack == nil || st.stack.crop.IsZero() {
		return -1
	}
	st.buff.AddRect(asset.getCoord(x, y, w, h, margin, 0, 0), cd, asset.getCellWidth(borderWidth))
	return 1
}

func (asset *Asset) _sa_paint_rect(x, y, w, h float64, margin float64, r, g, b, a uint32, borderWidth float64) int64 {

	return asset.paint_rect(x, y, w, h, margin, OsCd{byte(r), byte(g), byte(b), byte(a)}, borderWidth)
}

func (asset *Asset) _sa_paint_line(x, y, w, h float64, margin float64, sx, sy, ex, ey float64, r, g, b, a uint32, width float64) int64 {

	st := asset.app.root.levels.GetStack()
	if st.stack == nil || st.stack.crop.IsZero() {
		return -1
	}

	coord := asset.getCoord(x, y, w, h, margin, 0, 0)
	var start OsV2
	start.X = coord.Start.X + int(float64(coord.Size.X)*sx)
	start.Y = coord.Start.Y + int(float64(coord.Size.Y)*sy)

	var end OsV2
	end.X = coord.Start.X + int(float64(coord.Size.X)*ex)
	end.Y = coord.Start.Y + int(float64(coord.Size.Y)*ey)

	st.buff.AddLine(start, end, InitOsCd32(r, g, b, a), asset.getCellWidth(width))
	return 1
}

func (asset *Asset) _sa_paint_circle(x, y, w, h float64, margin float64, sx, sy, rad float64, r, g, b, a uint32, borderWidth float64) int64 {

	st := asset.app.root.levels.GetStack()
	if st.stack == nil || st.stack.crop.IsZero() {
		return -1
	}

	coord := asset.getCoord(x, y, w, h, margin, 0, 0)
	var s OsV2
	s.X = coord.Start.X + int(float64(coord.Size.X)*sx)
	s.Y = coord.Start.Y + int(float64(coord.Size.Y)*sy)
	rr := asset.getCellWidth(rad)
	cq := InitOsQuadMid(s, OsV2{rr * 2, rr * 2})

	st.buff.AddCircle(cq, InitOsCd32(r, g, b, a), asset.getCellWidth(borderWidth))
	return 1
}

func (asset *Asset) paint_file(x, y, w, h float64, file string, title string, margin, marginX, marginY float64, r, g, b, a uint32, alignV, alignH uint32, fill uint32) int64 {

	st := asset.app.root.levels.GetStack()
	if st.stack == nil || st.stack.crop.IsZero() {
		return -1
	}

	coord := asset.getCoord(x, y, w, h, margin, marginX, marginY)
	cd := InitOsCd32(r, g, b, a)

	path, err := MediaParseUrl(file, asset)
	if err != nil {
		asset.AddLogErr(err)
		return -1
	}

	st.buff.AddImage(path, coord, cd, int(alignV), int(alignH), fill != 0)

	if len(title) > 0 {
		asset.paint_title(0, 0, 1, 1, title)
	}

	return 1
}
func (asset *Asset) _sa_paint_file(x, y, w, h float64, fileMem uint64, titleMem uint64, margin, marginX, marginY float64, r, g, b, a uint32, alignV, alignH uint32, fill uint32) int64 {

	file, err := asset.ptrToString(fileMem)
	if asset.AddLogErr(err) {
		return -1
	}
	title, err := asset.ptrToString(titleMem)
	if asset.AddLogErr(err) {
		return -1
	}

	return asset.paint_file(x, y, w, h, file, title, margin, marginX, marginY, r, g, b, a, alignV, alignH, fill)
}

func (asset *Asset) paint_title(x, y, w, h float64, text string) int64 {

	st := asset.app.root.levels.GetStack()
	if st.stack == nil || st.stack.crop.IsZero() {
		return -1
	}

	if st.stack.enableInput {
		coord := asset.getCoord(x, y, w, h, 0, 0, 0)

		if coord.HasIntersect(st.stack.crop) {
			asset.app.root.tile.Set(asset.app.root.ui.io.touch.pos, coord, text, OsCd_black())
		}
	}
	return 1
}

func (asset *Asset) _sa_paint_title(x, y, w, h float64, valueMem uint64) int64 {

	value, err := asset.ptrToString(valueMem)
	if asset.AddLogErr(err) {
		return -1
	}

	return asset.paint_title(x, y, w, h, value)
}

func (asset *Asset) paint_cursor(name string) (int64, error) {

	err := asset.app.root.ui.PaintCursor(name)
	if err != nil {
		return -1, err
	}

	return 1, nil
}

func (asset *Asset) _sa_paint_cursor(nameMem uint64) int64 {

	name, err := asset.ptrToString(nameMem)
	if asset.AddLogErr(err) {
		return -1
	}

	ret, err := asset.paint_cursor(name)
	asset.AddLogErr(err)
	return ret
}
