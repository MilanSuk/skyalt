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
	"os"

	"github.com/go-gl/gl/v2.1/gl"
)

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
	if err != nil {
		return nil, nil, err
	}

	tex, err := InitWinTextureFromImage(img)
	return tex, img, err
}

func InitWinTextureFromFile(path string) (*WinTexture, image.Image, error) {
	imgFile, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		return nil, nil, err
	}

	tex, err := InitWinTextureFromImage(img)
	return tex, img, err
}

func (tex *WinTexture) Destroy() {
	if tex.id > 0 {
		gl.DeleteTextures(1, &tex.id)
	}
}

func _WinTexture_drawQuadNoUVs(coord OsV4, depth int) {
	s := coord.Start
	e := coord.End()

	gl.Vertex3f(float32(s.X), float32(s.Y), float32(depth))
	gl.Vertex3f(float32(e.X), float32(s.Y), float32(depth))
	gl.Vertex3f(float32(e.X), float32(e.Y), float32(depth))
	gl.Vertex3f(float32(s.X), float32(e.Y), float32(depth))
}

func _WinTexture_drawQuadNoUVs_cd(coord OsV4, depth int, cd color.RGBA, alphas [4]byte) {
	s := coord.Start
	e := coord.End()

	gl.Color4ub(cd.R, cd.G, cd.B, alphas[0])
	gl.Vertex3f(float32(s.X), float32(s.Y), float32(depth))

	gl.Color4ub(cd.R, cd.G, cd.B, alphas[1])
	gl.Vertex3f(float32(e.X), float32(s.Y), float32(depth))

	gl.Color4ub(cd.R, cd.G, cd.B, alphas[2])
	gl.Vertex3f(float32(e.X), float32(e.Y), float32(depth))

	gl.Color4ub(cd.R, cd.G, cd.B, alphas[3])
	gl.Vertex3f(float32(s.X), float32(e.Y), float32(depth))
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
		for i := 0; i < 4; i++ {
			gl.TexCoord2f(uvs[i].X, uvs[i].Y)
			gl.Vertex3f(float32(pts[i].X), float32(pts[i].Y), float32(depth))
		}
	}
	gl.End()

	gl.BindTexture(gl.TEXTURE_2D, 0) //unbind
}
