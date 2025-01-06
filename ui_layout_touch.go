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

type UiLayoutInput struct {
	canvas  uint64
	scrollV uint64
	scrollH uint64
	resize  uint64

	canvasOver uint64 //when scroll over Slider or Map
}

func (in *UiLayoutInput) Set(canvas uint64, scrollV uint64, scrollH uint64, resize uint64) {
	in.canvas = canvas
	in.scrollV = scrollV
	in.scrollH = scrollH
	in.resize = resize
}

func (in *UiLayoutInput) Reset() {
	*in = UiLayoutInput{}
}

func (in *UiLayoutInput) IsActive() bool {
	return in.canvas != 0 || in.scrollV != 0 || in.scrollH != 0 || in.resize != 0
}
func (in *UiLayoutInput) IsCanvasActive() bool {
	return in.canvas != 0
}

func (in *UiLayoutInput) IsResizeActive() bool {
	return in.resize != 0
}
func (in *UiLayoutInput) IsScrollOrResizeActive() bool {
	return in.scrollV != 0 || in.scrollH != 0 || in.resize != 0
}

func (in *UiLayoutInput) IsFnMove(canvas uint64, scrollV uint64, scrollH uint64, resize uint64) bool {
	return ((canvas != 0 && in.canvas == canvas) || (scrollV != 0 && in.scrollV == scrollV) || (scrollH != 0 && in.scrollH == scrollH) || (resize != 0 && in.resize == resize))
}
