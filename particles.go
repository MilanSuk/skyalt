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
	"fmt"
	"image"
	"math"
	"math/rand"
	"os"

	"github.com/veandco/go-sdl2/sdl"
)

type Noise struct {
	noise []byte
	size  OsV2
}

func NewNoise(size OsV2) *Noise {
	var self Noise
	self.size = size
	self.noise = make([]byte, size.X*size.Y)
	return &self
}

func (ptcs *Noise) Draw(offset OsV2, zoom int, octaves int, p float32) {
	for y := 0; y < ptcs.size.Y; y++ {
		for x := 0; x < ptcs.size.X; x++ {
			getnoise := float32(0)
			for i := 0; i < octaves-1; i++ {
				frequency := float32(math.Pow(2.0, float64(i)))
				amplitude := float32(math.Pow(float64(p), float64(i)))
				getnoise += _Noise_get(float32(x+offset.X)*frequency/float32(zoom), float32(y+offset.Y)/float32(zoom)*frequency) * amplitude
			}
			getnoise = (getnoise + 1) * 0.5
			if getnoise < 0 {
				getnoise = 0
			}
			if getnoise > 1 {
				getnoise = 1
			}

			ptcs.noise[y*ptcs.size.X+x] = byte(255.0 * getnoise)
		}
	}
}

func _Noise_find(x float32, y float32) float32 {
	n := int(x + y*57)
	n = (n << 13) ^ n
	nn := (n*(n*n*60493+19990303) + 1376312589) & 0x7fffffff
	return 1.0 - (float32(nn) / 1073741824.0)
}

func _Noise_interpolate(a float32, b float32, x float32) float32 {
	ft := float32(x) * 3.1415927
	f := float32(1.0-math.Cos(float64(ft))) * 0.5
	return a*(1.0-f) + b*f
}

func _Noise_get(x float32, y float32) float32 {

	floorx := float32(int(x))
	floory := float32(int(y))

	s := _Noise_find(floorx, floory)
	t := _Noise_find(floorx+1, floory)
	u := _Noise_find(floorx, floory+1)
	v := _Noise_find(floorx+1, floory+1)

	i1 := _Noise_interpolate(s, t, x-floorx)
	i2 := _Noise_interpolate(u, v, x-floorx)
	return _Noise_interpolate(i1, i2, y-floory)
}

type Particles struct {
	poses  []OsV2f
	vels   []OsV2f
	alphas []float32
	num    int
	time   float64
	noiseX *Noise
	noiseY *Noise

	num_draw int

	logo     *sdl.Texture
	logoSize OsV2

	anim_max_time float32 // zero = deactivated
	anim_act_time float32

	done    float32
	oldDone float32
}

func NewParticles(render *sdl.Renderer) (*Particles, error) {
	var ptcs Particles

	// create logo texture
	file, err := os.Open(SKYALT_LOGO)
	if err != nil {
		return nil, fmt.Errorf("Open(%s) failed: %w", SKYALT_LOGO, err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("Decode(%s) failed: %w", SKYALT_LOGO, err)
	}

	ptcs.logo, ptcs.logoSize, err = CreateTextureFromImage(img, render)
	if err != nil {
		return nil, fmt.Errorf("CreateTextureFromImage() failed: %w", err)
	}

	ptcs.noiseX = NewNoise(ptcs.logoSize)
	ptcs.noiseY = NewNoise(ptcs.logoSize)

	return &ptcs, nil
}

func (ptcs *Particles) Destroy() error {
	ptcs.Clear()

	err := ptcs.logo.Destroy()
	if err != nil {
		return fmt.Errorf("texture Destroy() failed: %w", err)
	}

	ptcs.noiseX = nil
	ptcs.noiseY = nil
	return nil
}

func (ptcs *Particles) Clear() {
	ptcs.poses = nil
	ptcs.vels = nil
	ptcs.alphas = nil
	ptcs.num = 0

	ptcs.time = 0
}

func (ptcs *Particles) Emit() error {
	ptcs.Clear()

	pixels, pitch, err := ptcs.logo.Lock(nil)
	if err != nil {
		return fmt.Errorf("Emit() texture Lock() failed: %w", err)
	}

	// get num particles
	n := 0
	for y := 0; y < ptcs.logoSize.Y; y++ {
		for x := 0; x < ptcs.logoSize.X; x++ {
			if pixels[y*pitch+(x*4)+3] != 0 { //get alpha
				n++
			}
		}
	}

	const SUBS = 2 // 3
	n *= (SUBS * SUBS)

	// alloc
	ptcs.poses = make([]OsV2f, n)
	ptcs.vels = make([]OsV2f, n)
	ptcs.alphas = make([]float32, n)
	ptcs.num = n
	ptcs.time = 0

	st := OsV2{OsAbs(int(rand.Int31() % 255)), OsAbs(int(rand.Int31() % 255))}
	ptcs.noiseX.Draw(st, 35, 6, 0.5)
	st = st.Add(ptcs.logoSize)
	ptcs.noiseY.Draw(st, 35, 6, 0.5)

	// set data
	n = 0
	for y := 0; y < ptcs.logoSize.Y; y++ {
		for x := 0; x < ptcs.logoSize.X; x++ {
			cd_a := float32(pixels[y*pitch+(x*4)+3])
			if cd_a != 0 {
				for i := 0; i < SUBS; i++ {
					for j := 0; j < SUBS; j++ {
						ptcs.poses[n] = OsV2f{float32(x) + 1.0/float32(SUBS)*float32(i), float32(y) + 1.0/float32(SUBS)*float32(j)}
						ptcs.vels[n] = OsV2f{0, 0}
						ptcs.alphas[n] = cd_a / float32(-255.0) // negative = not active
						n++
					}
				}
			}
		}
	}
	ptcs.num_draw = n

	ptcs.logo.Unlock()
	return nil
}

func (ptcs *Particles) StartAnim(time_sec float32) {

	ptcs.anim_max_time = time_sec
	ptcs.anim_act_time = 0
	ptcs.Emit()
}

func (ptcs *Particles) UpdateTime() float64 {
	t := OsTime()
	dt := t - ptcs.time
	ptcs.time = t
	if dt > 0.04 { // first tick
		dt = 0.04 // at least 25fps
	}
	return dt
}

func (ptcs *Particles) UpdateDone(DT float32) float32 {

	if ptcs.anim_max_time != 0 {
		ptcs.anim_act_time += DT
		ptcs.done = ptcs.anim_act_time / ptcs.anim_max_time
	}

	if ptcs.done < ptcs.oldDone {
		ptcs.Emit()
	}

	ptcs.oldDone = ptcs.done
	return ptcs.done
}

func (ptcs *Particles) GetLogoCoord(render *sdl.Renderer) (OsV4, error) {

	w, h, err := render.GetOutputSize()
	if err != nil {
		return OsV4{}, fmt.Errorf("GetLogoCoord() RendererOutputSize() failed: %w", err)
	}

	screen := OsV2{int(w), int(h)}
	SX := float32(screen.X) / 4

	size := OsV2{int(SX), int(SX * float32(ptcs.logoSize.Y) / float32(ptcs.logoSize.X))}
	start := screen.Sub(size).MulV(0.5)

	return OsV4{start, size}, nil
}

func (ptcs *Particles) GetPosSmoothRepeat(noise *Noise, p OsV2) byte {

	// repeat
	x := OsAbs(p.X) % ptcs.logoSize.X
	y := OsAbs(p.Y) % ptcs.logoSize.Y

	// smooth - revers odd
	if (p.X/ptcs.logoSize.X)%2 != 0 {
		x = ptcs.logoSize.X - x - 1
	}
	if (p.Y/ptcs.logoSize.Y)%2 != 0 {
		y = ptcs.logoSize.Y - y - 1
	}

	return noise.noise[y*ptcs.logoSize.X+x]
}

func (ptcs *Particles) Draw(cd_theme OsCd, render *sdl.Renderer) (bool, error) {

	front_cd := OsCd{50, 50, 50, 255}

	var logoCd OsCd
	if ptcs.num_draw == 0 {
		logoCd = front_cd
	} else {
		logoCd = cd_theme
	}

	coord, err := ptcs.GetLogoCoord(render)
	if err != nil {
		return false, fmt.Errorf("Draw() GetLogoCoord() failed: %w", err)
	}

	err = ptcs.logo.SetColorMod(logoCd.R, logoCd.G, logoCd.B)
	if err != nil {
		return false, fmt.Errorf("Particles.Draw() SetColorMod() failed: %w", err)
	}
	err = ptcs.logo.SetAlphaMod(255)
	if err != nil {
		return false, fmt.Errorf("Particles.Draw() SetAlphaMod() failed: %w", err)
	}

	err = render.Copy(ptcs.logo, nil, coord.GetSDLRect())
	if err != nil {
		return false, fmt.Errorf("Particles.Draw() RenderCopy() failed: %w", err)
	}

	if ptcs.num_draw == 0 {
		return false, nil
	}

	// draw noise
	/*{
		w, h, _ := render.GetOutputSize()
		for y := int32(-h / 2); y < h/2; y++ {
			for x := int32(-w / 2); x < w/2; x++ {
				a := self.GetPosSmoothRepeat(self.noiseY, OsV2{x, y}.Add(OsV2{10000, 10000}))
				render.SetDrawColor(a, a, 0, 255)
				render.DrawPoint(coord.start.x+x, coord.start.y+y)
			}
		}
	}*/

	ratio := OsV2f{float32(coord.Size.X) / float32(ptcs.logoSize.X), float32(coord.Size.Y) / float32(ptcs.logoSize.Y)}

	err = render.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		return false, fmt.Errorf("Particles.Draw() SetDrawBlendMode() failed: %w", err)
	}

	last_p := OsV2{0, 0}

	for i := 0; i < ptcs.num; i++ {

		a := OsFAbs(ptcs.alphas[i])
		if a > 0.01 && !last_p.Cmp(ptcs.poses[i].toV2()) { // one particles per pixel(in row)
			var p OsV2f
			p.X = float32(coord.Start.X) + ptcs.poses[i].X*ratio.X
			p.Y = float32(coord.Start.Y) + ptcs.poses[i].Y*ratio.Y

			err = render.SetDrawColor(front_cd.R, front_cd.G, front_cd.B, uint8(a*255))
			if err != nil {
				return false, fmt.Errorf("Particles.Draw() SetDrawColor() failed: %w", err)
			}
			err = render.DrawPointF(p.X, p.Y) //DrawPoint<without F>() creates artifacts around y=0
			if err != nil {
				return false, fmt.Errorf("Particles.Draw() DrawPoint() failed: %w", err)
			}

			last_p = ptcs.poses[i].toV2()
		}
	}

	return ptcs.num_draw > 0, nil
}

func (ptcs *Particles) Update() {
	DT := float32(ptcs.UpdateTime())
	FADE := DT * 0.4

	done := ptcs.UpdateDone(DT)
	edge := float32(ptcs.logoSize.X) * (1.0 - done)

	if int(edge) == ptcs.logoSize.X { // edge is on right side = nothing to simulate
		return
	}

	ptcs.num_draw = 0
	for i := 0; i < ptcs.num; i++ {
		p := &ptcs.poses[i]
		alp := &ptcs.alphas[i]

		if p.X > edge {
			*alp = OsFAbs(*alp) // activate
		}

		if *alp > 0 {
			v := &ptcs.vels[i]

			ppp := p.toV2()
			acc := OsV2f{float32(ptcs.GetPosSmoothRepeat(ptcs.noiseX, ppp)) - 128, float32(ptcs.GetPosSmoothRepeat(ptcs.noiseY, ppp)) - 128}
			acc = acc.MulV(2.5)
			//acc.x *= 2.0 //boost ->

			*v = v.Add(acc.MulV(DT)) // v += acc * dt
			*p = p.Add(v.MulV(DT))   // p += v * dt

			*alp -= FADE
			if *alp < 0 {
				*alp = 0
			}
		}

		if *alp != 0 {
			ptcs.num_draw++
		}
	}

	if ptcs.num_draw == 0 { // done
		ptcs.Clear()
	}
}
