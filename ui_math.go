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

type Rect struct {
	X, Y, W, H float64
}

func (r Rect) check() Rect {
	if r.W < 0 {
		r.W = 0
	}
	if r.H < 0 {
		r.H = 0
	}
	return r
}

func (r Rect) Cut(v float64) Rect {
	r.X += v
	r.Y += v
	r.W -= 2 * v
	r.H -= 2 * v
	return r.check()
}
func (r Rect) Is() bool {
	return r.W > 0 && r.H > 0
}

func (r Rect) CutLeft(v float64) Rect {
	r.X += v
	r.W -= v
	return r.check()
}
func (r Rect) CutTop(v float64) Rect {
	r.Y += v
	r.H -= v
	return r.check()
}
func (r Rect) CutRight(v float64) Rect {
	r.W -= v
	return r.check()
}
func (r Rect) CutBottom(v float64) Rect {
	r.H -= v
	return r.check()
}
func (r Rect) GetPos(x, y float64) (float64, float64) {
	return r.X + r.W*x, r.Y + r.H*y
}
func (r Rect) Move(x, y float64) Rect {
	r.X -= x
	r.Y -= y
	return r
}

func Rect_centerFull(out Rect, in_w, in_h float64) Rect {
	var r Rect
	r.X = out.X
	r.Y = out.Y
	r.W = in_w
	r.H = in_h

	if out.W != in_w {
		r.X += (out.W - in_w) / 2
	}
	if out.Y != in_h {
		r.Y += (out.H - in_h) / 2
	}
	return r
}

func (r Rect) IsInside(x, y float64) bool {
	return (x > r.X && x < r.X+r.W && y > r.Y && y < r.Y+r.H)
}
