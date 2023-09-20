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

type LayoutLevel struct {
	name string
	use  int //init(-1), notUse(0), drawn(1)

	buff *PaintBuff

	infoLayout *RS_LScroll

	rootDiv *LayoutDiv
	stack   *LayoutDiv

	src_coordMoveCut OsV4

	close bool
}

func NewLayoutLevel(name string, src_coordMoveCut OsV4, infoLayout *RS_LScroll, ui *Ui) *LayoutLevel {

	var level LayoutLevel

	level.name = name
	level.src_coordMoveCut = src_coordMoveCut
	level.infoLayout = infoLayout

	level.buff = NewPaintBuff(ui)
	level.rootDiv = NewLayoutPack(nil, "", OsV4{}, infoLayout)

	level.use = -1
	return &level
}

func (level *LayoutLevel) Destroy() {
	level.buff.Destroy()
	level.rootDiv.Destroy(level.infoLayout)
}

func (level *LayoutLevel) GetCoord(q OsV4, winRect OsV4) OsV4 {

	if !level.src_coordMoveCut.IsZero() {
		// relative
		q = OsV4_relativeSurround(level.src_coordMoveCut, q, winRect)
	} else {
		// center
		q = OsV4_center(winRect, q.Size)
	}
	return q
}
