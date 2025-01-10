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

type UiLayoutEdit struct {
	hash                uint64
	activate_next_iters int //after pressing tab
	reload_hash         uint64

	orig_value string

	temp       string
	start, end int

	RefreshTicks int64

	KeySelectAll bool
	KeyCopy      bool
	KeyCut       bool
	KeyPaste     bool

	last_refreshInput_unix float64
}

func (edit *UiLayoutEdit) send(enter_key bool, ui *Ui) {

	in := LayoutInput{SetEdit: true, EditValue: edit.temp, EditEnter: enter_key}
	lay := ui.dom.FindHash(edit.hash)
	if lay != nil {
		ui.parent.CallInput(&lay.props, &in)
	}

}

func (edit *UiLayoutEdit) Set(dom *Layout3, editable bool, orig_value, value string, enter_key bool, finish bool, save bool, refreshInput bool) {
	if !finish && edit.hash == dom.props.Hash {
		if editable {
			edit.temp = value //refresh after code

			tm := OsTime()
			if refreshInput && (edit.orig_value != edit.temp) && (edit.last_refreshInput_unix+1 < OsTime()) {

				edit.orig_value = edit.temp
				edit.send(false, dom.ui)
				edit.last_refreshInput_unix = tm
			}
		}

		dom.ui.GetWin().SetTextCursorMove()
		return //no change
	}

	if editable && save && edit.hash != 0 {
		//save old editbox
		if (edit.orig_value != edit.temp) || enter_key {
			edit.send(enter_key, dom.ui)
		}
	}

	//reset
	edit.hash = 0
	edit.temp = ""
	edit.activate_next_iters = 0
	edit.RefreshTicks = 0
	edit.ResetShortcutKeys()

	//set new
	if !finish {
		edit.hash = dom.props.Hash
		edit.temp = value
		edit.orig_value = orig_value

		edit.start = 0
		edit.end = len(value)

		dom.ui.GetWin().SetTextCursorMove()
	}

	if !finish {
		dom.ui.SetRedrawBuffer() //refresh border thickness
	}
}

func (edit *UiLayoutEdit) ResetShortcutKeys() {
	edit.KeySelectAll = false
	edit.KeyCopy = false
	edit.KeyCut = false
	edit.KeyPaste = false
}

func (edit *UiLayoutEdit) Is(dom *Layout3) bool {
	return edit.hash == dom.props.Hash
}

func (edit *UiLayoutEdit) IsActivateNext() bool {
	return edit.activate_next_iters != 0
}
func (edit *UiLayoutEdit) SetActivateNext() {
	edit.activate_next_iters = 2
}

func (edit *UiLayoutEdit) SetRefreshTicks() {
	if edit.RefreshTicks == 0 {
		edit.RefreshTicks = OsTicks()
	}
}

func (edit *UiLayoutEdit) Maintenance(rs *UiClients) {
	if edit.activate_next_iters > 0 {
		edit.activate_next_iters--
	}
}

func (edit *UiLayoutEdit) IsActive() bool {
	return edit.hash != 0
}
