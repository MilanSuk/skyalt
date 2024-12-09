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

type UiSettings struct {
	Dialogs []UiDialog
	Layouts UiRootSettings
}

type UiDialog struct {
	Path        string
	ParentPath  string
	ParentTouch OsV2
}

type UiRootScroll struct {
	Pos int
	use bool
}
type UiRootResize struct {
	Pos  int
	Size float64
	use  bool
}
type UiRootSettings struct {
	V_scrolls map[uint64]*UiRootScroll `json:",omitempty"`
	H_scrolls map[uint64]*UiRootScroll `json:",omitempty"`

	Cols map[uint64][]UiRootResize `json:",omitempty"`
	Rows map[uint64][]UiRootResize `json:",omitempty"`
}

func (se *UiRootSettings) Init() {
	if se.V_scrolls == nil {
		se.V_scrolls = make(map[uint64]*UiRootScroll)
	}
	if se.H_scrolls == nil {
		se.H_scrolls = make(map[uint64]*UiRootScroll)
	}
	if se.Cols == nil {
		se.Cols = make(map[uint64][]UiRootResize)
	}
	if se.Rows == nil {
		se.Rows = make(map[uint64][]UiRootResize)
	}
}

func (se *UiRootSettings) GetScrollV(id uint64) int {
	wheel, found := se.V_scrolls[id]
	if found {
		se.V_scrolls[id].use = true
		return wheel.Pos
	}
	return 0
}
func (se *UiRootSettings) GetScrollH(id uint64) int {
	wheel, found := se.H_scrolls[id]
	if found {
		se.H_scrolls[id].use = true
		return wheel.Pos
	}
	return 0
}

func (se *UiRootSettings) SetScrollV(id uint64, pos int) {
	val, found := se.V_scrolls[id]
	if found {
		val.Pos = pos
	} else {
		se.V_scrolls[id] = &UiRootScroll{Pos: pos, use: true}
	}
}
func (se *UiRootSettings) SetScrollH(id uint64, pos int) {
	val, found := se.H_scrolls[id]
	if found {
		val.Pos = pos
	} else {
		se.H_scrolls[id] = &UiRootScroll{Pos: pos, use: true}
	}
}

func _UiRootResize_GetSet(items map[uint64][]UiRootResize, id uint64, pos int, size float64, write bool) float64 {
	for i := range items[id] {
		if items[id][i].Pos == pos {
			if write {
				items[id][i].Size = size
			}
			items[id][i].use = true
			return items[id][i].Size
		}
	}
	items[id] = append(items[id], UiRootResize{Pos: pos, Size: size})
	items[id][len(items[id])-1].use = true
	return size
}
func (se *UiRootSettings) GetCol(id uint64, pos int, def_size float64) float64 {
	return _UiRootResize_GetSet(se.Cols, id, pos, def_size, false)
}
func (se *UiRootSettings) SetCol(id uint64, pos int, size float64) {
	_UiRootResize_GetSet(se.Cols, id, pos, size, true)
}

func (se *UiRootSettings) GetRow(id uint64, pos int, def_size float64) float64 {
	return _UiRootResize_GetSet(se.Rows, id, pos, def_size, false)
}
func (se *UiRootSettings) SetRow(id uint64, pos int, size float64) {
	_UiRootResize_GetSet(se.Rows, id, pos, size, true)
}

func _UiRootResize_ColsRowsMaintenance(items map[uint64][]UiRootResize) {
	for id := range items {
		for i := len(items[id]) - 1; i >= 0; i-- {
			if !items[id][i].use {
				items[id] = append(items[id][:i], items[id][i+1:]...) //remove
			} else {
				items[id][i].use = false //reverse back
			}
			if len(items[id]) == 0 {
				delete(items, id)
			}
		}
	}
}
func _UiRootResize_ScrollsMaintenance(items map[uint64]*UiRootScroll) {
	for id, val := range items {
		if val.Pos == 0 || !val.use {
			delete(items, id)
		} else {
			val.use = false //reverse back
		}
	}
}
