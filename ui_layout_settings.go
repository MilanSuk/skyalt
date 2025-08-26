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
	"slices"
)

type UiSettings struct {
	Dialogs []*UiDialog
	Layouts UiRootSettings

	Highlight_text string

	changed_dialogs bool //open, close
}

func (s *UiSettings) IsChanged() bool {
	changed := s.changed_dialogs
	s.changed_dialogs = false
	return changed
}

func (s *UiSettings) CloseAllDialogs(ui *Ui) {
	for _, dia := range s.Dialogs {
		dia.Close(ui)
	}
	s.Dialogs = nil
}

func (s *UiSettings) OpenDialog(uid uint64, relative_uid uint64, parentTouch OsV2) *UiDialog {
	dia := &UiDialog{UID: uid, Relative_uid: relative_uid, TouchPos: parentTouch}
	s.Dialogs = append(s.Dialogs, dia)
	s.changed_dialogs = true
	return dia
}

func (s *UiSettings) FindDialog(uid uint64) *UiDialog {
	for _, dia := range s.Dialogs {
		if dia.UID == uid {
			return dia
		}
	}
	return nil
}

func (s *UiSettings) CloseDialog(dia *UiDialog, ui *Ui) {
	for i := len(s.Dialogs) - 1; i >= 0; i-- {
		if s.Dialogs[i] == dia {
			dia.Close(ui)
			s.Dialogs = slices.Delete(s.Dialogs, i, i+1)
			s.changed_dialogs = true
		}
	}
}

func (s *UiSettings) CloseTouchDialogs(ui *Ui) {
	win := ui.GetWin()

	for i := len(s.Dialogs) - 1; i >= 0; i-- {
		dia := s.Dialogs[i]

		layDia := ui.mainLayout.FindUID(dia.UID)
		if layDia != nil && layDia.touchDia { //touchDia == true
			layApp := layDia.GetApp()
			if layApp != nil {
				app_crop := layApp.crop
				outside := app_crop.Inside(win.io.Touch.Pos) && !layDia.CropWithScroll().Inside(win.io.Touch.Pos)
				if (win.io.Touch.Start && outside) || win.io.Keys.Esc {
					s.CloseDialog(dia, ui)

					ui.ResetIO()
					win.io.Keys.Esc = false
					win.io.Touch.Start = false
				}
			}
		}
	}
}

func (s *UiSettings) GetHigherDialogApp(findApp *Layout, ui *Ui) int {
	for i := len(s.Dialogs) - 1; i >= 0; i-- {
		layDia := ui.mainLayout.FindUID(s.Dialogs[i].UID)
		if layDia != nil {
			layApp := layDia.parent.GetApp()
			if layApp == findApp {
				return i
			}
		}
	}
	return -1
}

type UiDialog struct {
	UID          uint64
	Relative_uid uint64
	TouchPos     OsV2
}

func (dia *UiDialog) Close(ui *Ui) {
	d := ui.mainLayout.FindDialog(dia.UID)
	if d != nil {
		if d.fnCloseDialog != nil {
			d.fnCloseDialog()
		}
	}
}

func (dia *UiDialog) GetDialogCoord(q OsV4, ui *Ui) OsV4 {

	winRect := ui.winRect

	var relLayout *Layout
	if dia.Relative_uid != 0 {
		relLayout = ui.mainLayout.FindUID(dia.Relative_uid)
	}

	if relLayout != nil {
		// relative
		q = OsV4_relativeSurround(relLayout.crop, q, winRect, false)
	} else {

		if dia.TouchPos.IsZero() {
			// center
			q = OsV4_center(winRect, q.Size)
		} else {
			// touch
			q = OsV4_relativeSurround(OsV4{dia.TouchPos, OsV2{1, 1}}, q, winRect, false)
		}

	}
	return q
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

func (se *UiRootSettings) GetScrollV(uid uint64) int {
	wheel, found := se.V_scrolls[uid]
	if found {
		se.V_scrolls[uid].use = true
		return wheel.Pos
	}
	return 0
}
func (se *UiRootSettings) GetScrollH(uid uint64) int {
	wheel, found := se.H_scrolls[uid]
	if found {
		se.H_scrolls[uid].use = true
		return wheel.Pos
	}
	return 0
}

func UiRootSettings_GetMaxScroll() int {
	return 1000000000 //1B
}

func (se *UiRootSettings) SetScrollV(uid uint64, pos int) {
	val, found := se.V_scrolls[uid]
	if found {
		val.Pos = pos
	} else {
		se.V_scrolls[uid] = &UiRootScroll{Pos: pos, use: true}
	}
}
func (se *UiRootSettings) SetScrollH(uid uint64, pos int) {
	val, found := se.H_scrolls[uid]
	if found {
		val.Pos = pos
	} else {
		se.H_scrolls[uid] = &UiRootScroll{Pos: pos, use: true}
	}
}

func _UiRootResize_GetSet(items map[uint64][]UiRootResize, uid uint64, pos int, size float64, write bool) float64 {
	for i := range items[uid] {
		if items[uid][i].Pos == pos {
			if write {
				items[uid][i].Size = size
			}
			items[uid][i].use = true
			return items[uid][i].Size
		}
	}
	items[uid] = append(items[uid], UiRootResize{Pos: pos, Size: size})
	items[uid][len(items[uid])-1].use = true
	return size
}
func (se *UiRootSettings) GetCol(uid uint64, pos int, def_size float64) float64 {
	return _UiRootResize_GetSet(se.Cols, uid, pos, def_size, false)
}
func (se *UiRootSettings) SetCol(uid uint64, pos int, size float64) {
	_UiRootResize_GetSet(se.Cols, uid, pos, size, true)
}

func (se *UiRootSettings) GetRow(uid uint64, pos int, def_size float64) float64 {
	return _UiRootResize_GetSet(se.Rows, uid, pos, def_size, false)
}
func (se *UiRootSettings) SetRow(uid uint64, pos int, size float64) {
	_UiRootResize_GetSet(se.Rows, uid, pos, size, true)
}
