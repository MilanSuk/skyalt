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
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/veandco/go-sdl2/sdl"
)

type WinRender struct {
	render sdl.GLContext
}

func NewWinRender(window *sdl.Window) (*WinRender, error) {
	ren := &WinRender{}

	var err error
	ren.render, err = window.GLCreateContext()
	if LogsError(err) != nil {
		return nil, err
	}

	err = gl.Init()
	if LogsError(err) != nil {
		return nil, err
	}
	return ren, nil
}

func (ren *WinRender) Destroy() error {

	sdl.GLDeleteContext(ren.render)

	return nil
}

func (ren *WinRender) ReadGLScreenPixels(screen OsV4, coord OsV4, out *byte) error {
	//winH := win.GetScreenCoord().Size.Y
	winH := screen.Size.Y
	gl.ReadPixels(int32(coord.Start.X), int32(winH-(coord.Start.Y+coord.Size.Y)), int32(coord.Size.X), int32(coord.Size.Y), gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(out))
	return nil
}

func (ren *WinRender) StartRender(screen OsV4, clearCd color.RGBA) error {

	gl.Enable(gl.SCISSOR_TEST)

	gl.Enable(gl.DEPTH_TEST)
	gl.ClearColor(float32(clearCd.R)/255, float32(clearCd.G)/255, float32(clearCd.B)/255, float32(clearCd.A)/255)
	gl.ClearDepth(1)
	gl.DepthFunc(gl.LEQUAL)
	gl.Viewport(0, 0, int32(screen.Size.X), int32(screen.Size.Y))

	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(0, float64(screen.Size.X), float64(screen.Size.Y), 0, -1000, 1000)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()

	gl.Enable(gl.POINT_SMOOTH)
	gl.Hint(gl.POINT_SMOOTH_HINT, gl.NICEST)
	gl.Enable(gl.LINE_SMOOTH)
	gl.Hint(gl.LINE_SMOOTH_HINT, gl.NICEST)
	//gl.Enable(gl.POLYGON_SMOOTH)
	//gl.Hint(gl.POLYGON_SMOOTH_HINT, gl.NICEST)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Enable(gl.ALPHA_TEST)
	gl.ShadeModel(gl.SMOOTH)

	gl.Enable(gl.TEXTURE_2D)

	gl.Scissor(0, 0, int32(screen.Size.X), int32(screen.Size.Y))

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	return nil
}

func (ren *WinRender) SetClipRect(screen OsV4, coord OsV4) {
	winH := screen.Size.Y
	gl.Scissor(int32(coord.Start.X), int32(winH-(coord.Start.Y+coord.Size.Y)), int32(coord.Size.X), int32(coord.Size.Y))
}

func (ren *WinRender) DrawPointStart() {
	gl.Begin(gl.POINTS)
}
func (ren *WinRender) DrawPointCdI(pos OsV2, depth int, cd color.RGBA) {
	gl.Color4ub(cd.R, cd.G, cd.B, cd.A)
	gl.Vertex3i(int32(pos.X), int32(pos.Y), int32(depth))
}
func (ren *WinRender) DrawPointCdF(pos OsV2f, depth int, cd color.RGBA) {
	gl.Color4ub(cd.R, cd.G, cd.B, cd.A)
	gl.Vertex3f(float32(pos.X), float32(pos.Y), float32(depth))
}

func (ren *WinRender) DrawPointEnd() {
	gl.End()
}

func (ren *WinRender) DrawRect(start OsV2, end OsV2, depth int, cd color.RGBA) {
	if start.X != end.X && start.Y != end.Y {
		gl.Color4ub(cd.R, cd.G, cd.B, cd.A)

		gl.Begin(gl.QUADS)
		{
			gl.Vertex3f(float32(start.X), float32(start.Y), float32(depth))
			gl.Vertex3f(float32(end.X), float32(start.Y), float32(depth))
			gl.Vertex3f(float32(end.X), float32(end.Y), float32(depth))
			gl.Vertex3f(float32(start.X), float32(end.Y), float32(depth))
		}
		gl.End()

		//gl.Enable(gl.POLYGON_SMOOTH)
	}
}

func (ren *WinRender) DrawLine(start OsV2, end OsV2, depth int, thick int, cd color.RGBA) {

	gl.Color4ub(cd.R, cd.G, cd.B, cd.A)

	v := end.Sub(start)
	if !v.IsZero() {

		if start.Y == end.Y {
			ren.DrawRect(start, OsV2{end.X, start.Y + thick}, depth, cd) // horizontal
		} else if start.X == end.X {
			ren.DrawRect(start, OsV2{start.X + thick, end.Y}, depth, cd) // vertical
		} else {
			gl.LineWidth(float32(thick))
			gl.Begin(gl.LINES)
			gl.Vertex3f(float32(start.X), float32(start.Y), float32(depth))
			gl.Vertex3f(float32(end.X), float32(end.Y), float32(depth))
			gl.End()
		}
	}
}

func _Win_getBezierPoint(t float64, a, b, c, d OsV2f) OsV2f {
	af := a.MulV(float32(math.Pow(t, 3)))
	bf := b.MulV(float32(3 * math.Pow(t, 2) * (1 - t)))
	cf := c.MulV(float32(3 * t * math.Pow((1-t), 2)))
	df := d.MulV(float32(math.Pow((1 - t), 3)))

	return af.Add(bf).Add(cf).Add(df)
}

func (ren *WinRender) DrawBezier(a OsV2, b OsV2, c OsV2, d OsV2, depth int, thick int, cd color.RGBA, dash_px float32, move float32) {

	gl.Color4ub(cd.R, cd.G, cd.B, cd.A)

	aa := a.toV2f()
	bb := b.toV2f()
	cc := c.toV2f()
	dd := d.toV2f()

	gl.LineWidth(float32(thick))
	if dash_px > 0 {
		gl.Begin(gl.LINES)
	} else {
		gl.Begin(gl.LINE_STRIP)
	}

	{
		//compute length
		len := float32(0)
		last_a := a.toV2f()
		N := 10
		div := 1 / float64(N)
		for t := float64(0); t <= 1.001; t += div {
			len += last_a.Sub(_Win_getBezierPoint(t, aa, bb, cc, dd)).Len()
		}

		N = OsTrn(dash_px > 0, int(len/dash_px), int(len/5)) // 5 = 5px jump
		div = 1 / float64(N)

		pre_p := _Win_getBezierPoint(0-div, aa, bb, cc, dd)
		for t := float64(0); t <= 1.001; t += div {
			p := _Win_getBezierPoint(t, aa, bb, cc, dd)

			if move != 0 {
				old_p := p
				v := p.Sub(pre_p)
				v.X, v.Y = -v.Y, v.X
				v = v.MulV(move / v.Len())
				p = p.Add(v)
				pre_p = old_p
			}

			gl.Vertex3f(p.X, p.Y, float32(depth))

		}
	}
	gl.End()
}

type WinTexture struct {
	id   uint32
	size OsV2
}

func InitWinTextureSize(size OsV2) (*WinTexture, error) {
	var tex WinTexture

	gl.GenTextures(1, &tex.id)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, tex.id)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	tex.size = size
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.SRGB_ALPHA, int32(tex.size.X), int32(tex.size.Y), 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)

	return &tex, nil
}

func InitWinVideo(size OsV2) (*WinTexture, error) {
	var tex WinTexture

	gl.GenTextures(1, &tex.id)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, tex.id)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP)

	tex.size = size

	return &tex, nil
}

func InitWinTextureFromImageRGBAPix(rgba []byte, size OsV2) (*WinTexture, error) {
	var tex WinTexture

	gl.GenTextures(1, &tex.id)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, tex.id)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	tex.size = size
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, int32(tex.size.X), int32(tex.size.Y), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba))

	//gl.GenerateMipmap(texture.id)
	gl.BindTexture(gl.TEXTURE_2D, 0) //unbind

	return &tex, nil
}
func InitWinTextureFromImageAlphaPix(alpha []byte, size OsV2) (*WinTexture, error) {
	var tex WinTexture

	gl.GenTextures(1, &tex.id)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, tex.id)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	tex.size = size
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.ALPHA, int32(tex.size.X), int32(tex.size.Y), 0, gl.ALPHA, gl.UNSIGNED_BYTE, gl.Ptr(alpha))

	gl.BindTexture(gl.TEXTURE_2D, 0) //unbind

	return &tex, nil
}

func InitWinTextureFromImageRGBA(rgba *image.RGBA) (*WinTexture, error) {
	return InitWinTextureFromImageRGBAPix(rgba.Pix, OsV2{rgba.Rect.Size().X, rgba.Rect.Size().Y})
}

func InitWinTextureFromImageAlpha(alpha *image.Alpha) (*WinTexture, error) {
	return InitWinTextureFromImageAlphaPix(alpha.Pix, OsV2{alpha.Rect.Size().X, alpha.Rect.Size().Y})
}

func InitWinTextureFromImage(img image.Image) (*WinTexture, error) {

	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Pt(0, 0), draw.Src)

	return InitWinTextureFromImageRGBA(rgba)
}

func InitWinTextureFromBlob(blob []byte) (*WinTexture, image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(blob))
	if LogsError(err) != nil {
		return nil, nil, err
	}

	tex, err := InitWinTextureFromImage(img)
	return tex, img, err
}

func InitWinTextureFromFile(path string) (*WinTexture, image.Image, error) {
	imgFile, err := os.Open(path)
	if LogsError(err) != nil {
		return nil, nil, err
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if LogsError(err) != nil {
		return nil, nil, err
	}

	tex, err := InitWinTextureFromImage(img)
	if LogsError(err) != nil {
		return nil, nil, err
	}

	return tex, img, err
}

func (tex *WinTexture) Destroy() {
	if tex.id > 0 {
		gl.DeleteTextures(1, &tex.id)
		tex.id = 0
	}
}

func (tex *WinTexture) UpdateContent(size OsV2, pixels unsafe.Pointer) {
	gl.BindTexture(gl.TEXTURE_2D, tex.id)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(tex.size.X), int32(tex.size.Y), 0, gl.RGBA, gl.UNSIGNED_BYTE, pixels)
}

func (tex *WinTexture) DrawQuadUV(coord OsV4, depth int, cd color.RGBA, sUV, eUV OsV2f) {
	gl.Color4ub(cd.R, cd.G, cd.B, cd.A)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, tex.id)

	gl.Begin(gl.QUADS)
	{
		s := coord.Start
		e := coord.End()

		gl.TexCoord2f(sUV.X, sUV.Y)
		gl.Vertex3f(float32(s.X), float32(s.Y), float32(depth))

		gl.TexCoord2f(eUV.X, sUV.Y)
		gl.Vertex3f(float32(e.X), float32(s.Y), float32(depth))

		gl.TexCoord2f(eUV.X, eUV.Y)
		gl.Vertex3f(float32(e.X), float32(e.Y), float32(depth))

		gl.TexCoord2f(sUV.X, eUV.Y)
		gl.Vertex3f(float32(s.X), float32(e.Y), float32(depth))
	}
	gl.End()

	gl.BindTexture(gl.TEXTURE_2D, 0) //unbind
}

func (tex *WinTexture) DrawQuad(coord OsV4, depth int, cd color.RGBA) {
	tex.DrawQuadUV(coord, depth, cd, OsV2f{}, OsV2f{1, 1})
}

func (tex *WinTexture) DrawPointsUV(pts [4]OsV2f, uvs [4]OsV2f, depth int, cd color.RGBA) {
	gl.Color4ub(cd.R, cd.G, cd.B, cd.A)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, tex.id)

	gl.Begin(gl.QUADS)
	{
		for i := range 4 {
			gl.TexCoord2f(uvs[i].X, uvs[i].Y)
			gl.Vertex3f(float32(pts[i].X), float32(pts[i].Y), float32(depth))
		}
	}
	gl.End()

	gl.BindTexture(gl.TEXTURE_2D, 0) //unbind
}
