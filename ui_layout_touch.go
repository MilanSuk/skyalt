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
}

func (layTouch *UiLayoutInput) Set(canvas uint64, scrollV uint64, scrollH uint64, resize uint64) {
	layTouch.canvas = canvas
	layTouch.scrollV = scrollV
	layTouch.scrollH = scrollH
	layTouch.resize = resize
}

func (layTouch *UiLayoutInput) Reset() {
	*layTouch = UiLayoutInput{}
}

func (layTouch *UiLayoutInput) IsActive() bool {
	return layTouch.canvas != 0 || layTouch.scrollV != 0 || layTouch.scrollH != 0 || layTouch.resize != 0
}
func (layTouch *UiLayoutInput) IsCanvasActive() bool {
	return layTouch.canvas != 0
}

func (layTouch *UiLayoutInput) IsResizeActive() bool {
	return layTouch.resize != 0
}
func (layTouch *UiLayoutInput) IsScrollOrResizeActive() bool {
	return layTouch.scrollV != 0 || layTouch.scrollH != 0 || layTouch.resize != 0
}

func (layTouch *UiLayoutInput) IsFnMove(canvas uint64, scrollV uint64, scrollH uint64, resize uint64) bool {
	return ((canvas != 0 && layTouch.canvas == canvas) || (scrollV != 0 && layTouch.scrollV == scrollV) || (scrollH != 0 && layTouch.scrollH == scrollH) || (resize != 0 && layTouch.resize == resize))
}
