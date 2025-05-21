/*
Copyright 2025 Milan Suk

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
	"math"
)

type OsV2f struct {
	X float32
	Y float32
}

func (a OsV2f) Add(b OsV2f) OsV2f {
	return OsV2f{a.X + b.X, a.Y + b.Y}
}
func (a OsV2f) Sub(b OsV2f) OsV2f {
	return OsV2f{a.X - b.X, a.Y - b.Y}
}
func (a OsV2f) Mul(b OsV2f) OsV2f {
	return OsV2f{a.X * b.X, a.Y * b.Y}
}
func (a OsV2f) Div(b OsV2f) OsV2f {
	return OsV2f{a.X / b.X, a.Y / b.Y}
}
func (a OsV2f) MulV(t float32) OsV2f {
	return OsV2f{a.X * t, a.Y * t}
}
func (a OsV2f) DivV(t float32) OsV2f {
	return a.MulV(1 / t)
}
func (a OsV2f) toV2() OsV2 {
	return OsV2{int(a.X), int(a.Y)}
}
func (a OsV2f) Cmp(b OsV2f) bool {
	return a.X == b.X && a.Y == b.Y
}
func (a OsV2f) Min(b OsV2f) OsV2f {
	return OsV2f{OsMinFloat32(a.X, b.X), OsMinFloat32(a.Y, b.Y)}
}
func (a OsV2f) Max(b OsV2f) OsV2f {
	return OsV2f{OsMaxFloat32(a.X, b.X), OsMaxFloat32(a.Y, b.Y)}
}
func (v OsV2f) Len() float32 {
	return float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y)))
}

type OsV2 struct {
	X int
	Y int
}

func OsV2_32(x, y int32) OsV2 {
	return OsV2{int(x), int(y)}
}

func (v *OsV2) Get32() (int32, int32) {
	return int32(v.X), int32(v.Y)
}

func (v *OsV2) EqAdd(val OsV2) {
	v.X += val.X
	v.Y += val.Y
}
func (v *OsV2) EqSub(vel OsV2) {
	v.X -= vel.X
	v.Y -= vel.Y
}

func (a OsV2) Add(b OsV2) OsV2 {
	return OsV2{a.X + b.X, a.Y + b.Y}
}
func (a OsV2) Sub(b OsV2) OsV2 {
	return OsV2{a.X - b.X, a.Y - b.Y}
}
func (a OsV2) MulV(t float32) OsV2 {
	return OsV2{int(float32(a.X) * t), int(float32(a.Y) * t)}
}

func (a OsV2) Aprox(b OsV2, t float32) OsV2 {
	return a.Add(b.Sub(a).MulV(t))
}

func (a OsV2) toV2f() OsV2f {
	return OsV2f{float32(a.X), float32(a.Y)}
}

func (a OsV2) Cmp(b OsV2) bool {
	return a.X == b.X && a.Y == b.Y
}

func (start OsV2) Inside(end OsV2, test OsV2) bool {
	return test.X >= start.X && test.Y >= start.Y && test.X < end.X && test.Y < end.Y
}

func (a OsV2) Min(b OsV2) OsV2 {
	return OsV2{OsMin(a.X, b.X), OsMin(a.Y, b.Y)}
}

func (a OsV2) Max(b OsV2) OsV2 {
	return OsV2{OsMax(a.X, b.X), OsMax(a.Y, b.Y)}
}

func (v OsV2) Is() bool {
	return v.X != 0 && v.Y != 0
}

func (v OsV2) IsZero() bool {
	return v.X == 0 && v.Y == 0
}

func (v *OsV2) Switch() {
	*v = OsV2{v.Y, v.X}
}

func (v *OsV2) Sort() {
	if v.X > v.Y {
		v.Switch()
	}
}

func (v OsV2) Len() float32 {
	return float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y)))
}

func (v OsV2) Angle() float32 {
	return float32(math.Atan2(float64(v.Y), float64(v.X))) //<-PI, PI>
}

func (a OsV2) Distance(b OsV2) float32 {
	return a.Sub(b).Len()
}

func OsV2_Intersect(a OsV2, b OsV2) OsV2 {
	v := OsV2{OsMax(a.X, b.X), OsMax(a.Y, b.Y)}

	if v.X > v.Y {
		return OsV2{}
	}
	return v
}

func OsV2_InRatio(rect OsV2, orig OsV2) OsV2 {
	rectRatio := float32(rect.X) / float32(rect.Y)
	origRatio := float32(orig.X) / float32(orig.Y)

	var ratio float32
	if origRatio > rectRatio {
		ratio = float32(rect.X) / float32(orig.X)
	} else {
		ratio = float32(rect.Y) / float32(orig.Y)
	}
	return orig.MulV(ratio)
}

func OsV2_OutRatio(rect OsV2, orig OsV2) OsV2 {
	rectRatio := float32(rect.X) / float32(rect.Y)
	origRatio := float32(orig.X) / float32(orig.Y)

	var ratio float32
	if origRatio < rectRatio {
		ratio = float32(rect.X) / float32(orig.X)
	} else {
		ratio = float32(rect.Y) / float32(orig.Y)
	}
	return orig.MulV(ratio)
}

func (coord OsV4) Align(size OsV2, align OsV2) OsV2 {
	start := coord.Start

	if align.X == 0 {
		// left
	} else if align.X == 1 {
		// center
		if size.X > coord.Size.X {
			start.X = coord.Start.X // + H / 2
		} else {
			start.X = coord.Middle().X - size.X/2
		}
	} else {
		// right
		start.X = coord.End().X - size.X
	}

	// y
	if size.Y >= coord.Size.Y {
		start.Y += (coord.Size.Y - size.Y) / 2
	} else {
		if align.Y == 0 {
			start.Y = coord.Start.Y // + H / 2
		} else if align.Y == 1 {
			start.Y += (coord.Size.Y - size.Y) / 2
		} else if align.Y == 2 {
			start.Y += (coord.Size.Y) - size.Y
		}
	}

	return start
}

type OsV4 struct {
	Start OsV2
	Size  OsV2
}

func InitOsV4(x, y, w, h int) OsV4 {
	return OsV4{OsV2{x, y}, OsV2{w, h}}
}

func InitOsV4Mid(mid OsV2, size OsV2) OsV4 {
	return InitOsV4(mid.X-size.X/2, mid.Y-size.Y/2, size.X, size.Y)
}

func InitOsV4ab(a OsV2, b OsV2) OsV4 {
	st := OsV2{OsMin(a.X, b.X), OsMin(a.Y, b.Y)}
	sz := OsV2{OsAbs(a.X - b.X), OsAbs(a.Y - b.Y)}
	return InitOsV4(st.X, st.Y, sz.X, sz.Y)
}

func (v OsV4) End() OsV2 {
	return OsV2{v.Start.X + v.Size.X, v.Start.Y + v.Size.Y}
}

func (v OsV4) Is() bool {
	return v.Size.Is()
}

func (v OsV4) IsZero() bool {
	return v.Size.IsZero()
}

func (v OsV4) Check() {
	if v.Size.X < 0 {
		v.Size.X = 0
	}
	if v.Size.Y < 0 {
		v.Size.Y = 0
	}
}

func (v OsV4) GetPos(x, y float64) OsV2 {
	return OsV2{v.Start.X + int(float64(v.Size.X)*x), v.Start.Y + int(float64(v.Size.Y)*y)}
}

func (q OsV4) CropX(space int) OsV4 {
	space *= 2
	if space > q.Size.X {
		space = q.Size.X
	}
	return InitOsV4(q.Start.X+space/2, q.Start.Y, q.Size.X-space, q.Size.Y)
}

func (q OsV4) CropY(space int) OsV4 {
	space *= 2
	if space > q.Size.Y {
		space = q.Size.Y
	}
	return InitOsV4(q.Start.X, q.Start.Y+space/2, q.Size.X, q.Size.Y-space)
}

func (q OsV4) Crop(space int) OsV4 {
	r := q.CropX(space)
	return r.CropY(space)
}

func (q OsV4) Inner(top, bottom, left, right int) OsV4 {
	for q.Size.X < (left + right) { //for!
		left--
		right--
	}
	for q.Size.Y < (top + bottom) { //for!
		top--
		bottom--
	}
	return InitOsV4(q.Start.X+left, q.Start.Y+top, q.Size.X-(left+right), q.Size.Y-(top+bottom))
}

func (v OsV4) Middle() OsV2 {
	return v.Start.Add(v.Size.MulV(0.5))
}

func (v OsV4) Inside(test OsV2) bool {
	return v.Start.Inside(v.End(), test)
}
func (a OsV4) Cmp(b OsV4) bool {
	return a.Start.Cmp(b.Start) && a.Size.Cmp(b.Size)
}

func OsV4_center(out OsV4, in OsV2) OsV4 {
	r := OsV4{out.Start, in}

	if out.Size.X > in.X {
		r.Start.X += (out.Size.X - in.X) / 2
	}
	if out.Size.Y > in.Y {
		r.Start.Y += (out.Size.Y - in.Y) / 2
	}
	return r
}

func OsV4_centerFull(out OsV4, in OsV2) OsV4 {
	r := OsV4{out.Start, in}

	if out.Size.X != in.X {
		r.Start.X += (out.Size.X - in.X) / 2
	}
	if out.Size.Y != in.Y {
		r.Start.Y += (out.Size.Y - in.Y) / 2
	}
	return r
}

func (a OsV4) Area() int {
	return a.Size.X * a.Size.Y
}

func (a OsV4) Extend(b OsV4) OsV4 {

	start := OsV2{OsMin(a.Start.X, b.Start.X), OsMin(a.Start.Y, b.Start.Y)}

	ae := a.End()
	be := b.End()

	end := OsV2{OsMax(ae.X, be.X), OsMax(ae.Y, be.Y)}

	return OsV4{start, end.Sub(start)}
}

func (a OsV4) Extend2(q OsV4, v OsV2) OsV4 {

	start := OsV2{OsMin(q.Start.X, v.X), OsMin(q.Start.Y, v.Y)}

	end := q.End()
	end.X = OsMax(end.X, v.X)
	end.Y = OsMax(end.Y, v.Y)

	return OsV4{start, end.Sub(start)}
}

func (a OsV4) HasCover(b OsV4) bool {
	q := a.Extend(b)
	return q.Size.X < (a.Size.X+b.Size.X) && q.Size.Y < (a.Size.Y+b.Size.Y)
}

func (qA OsV4) GetIntersect(qB OsV4) OsV4 {

	if qA.HasCover(qB) {
		v_start := qA.Start.Max(qB.Start)
		v_end := qA.End().Min(qB.End())

		return OsV4{v_start, v_end.Sub(v_start)}
	}
	return OsV4{Start: qA.Start}
}

func (qA OsV4) HasIntersect(qB OsV4) bool {

	q := qA.GetIntersect(qB)
	return q.Is()
}

func OsV4_relativeSurround(src OsV4, dst OsV4, screen OsV4, priorUp bool) OsV4 {

	q := dst
	q.Start = q.Start.Sub(screen.Start)

	srcStart := src.Start.Sub(screen.Start)
	srcSize := src.Size

	up := srcStart.Y > (screen.Size.Y - srcStart.Y - srcSize.Y)
	if !up && priorUp {
		up = srcStart.Y > q.Size.Y //check enough space
	}

	q.Start.X = srcStart.X
	if q.Start.X+q.Size.X > screen.Size.X {
		q.Start.X = screen.Size.X - q.Size.X //move to left
	}

	if up {
		q.Start.Y = srcStart.Y - q.Size.Y
		if q.Start.Y < 0 {
			q.Size.Y = srcStart.Y
			q.Start.Y = 0
		}
	} else {
		q.Start.Y = srcStart.Y + srcSize.Y
		if q.Start.Y+q.Size.Y > screen.Size.Y {
			q.Size.Y = screen.Size.Y - q.Start.Y
		}
	}

	q.Start = q.Start.Add(screen.Start)
	return q
}

func (v *OsV4) RelativePos(p OsV2) OsV2f {
	s := p.Sub(v.Start)
	return OsV2f{float32(s.X) / float32(v.Size.X), float32(s.Y) / float32(v.Size.Y)}
}

func (v *OsV4) Relative(q OsV4) (x, y, w, h float32) {
	s := v.RelativePos(q.Start)
	e := v.RelativePos(q.End())
	return s.X, s.Y, (e.X - s.X), (e.Y - s.Y)
}

func (v OsV4) Cut(x, y, w, h float64) OsV4 {

	return InitOsV4(
		v.Start.X+int(float64(v.Size.X)*x),
		v.Start.Y+int(float64(v.Size.Y)*y),
		int(float64(v.Size.X)*w),
		int(float64(v.Size.Y)*h))
}

func (v OsV4) CutEx(x, y, w, h float64, space, spaceX, spaceY int) OsV4 {
	v = v.CropX(spaceX)
	v = v.CropY(spaceY)
	v = v.Crop(space)
	v = v.Cut(x, y, w, h)
	return v
}
