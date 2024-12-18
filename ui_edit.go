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

	temp       string
	start, end int

	RefreshTicks int64

	KeySelectAll bool
	KeyCopy      bool
	KeyCut       bool
	KeyPaste     bool
}

func (edit *UiLayoutEdit) Set(dom *Layout3, editable bool, orig_value, value string, enter_key bool, finish bool) {
	if !finish && edit.hash == dom.props.Hash {

		if editable {
			edit.temp = value //refresh after code
		}

		dom.ui.GetWin().SetTextCursorMove()
		return //no change
	}

	if editable {
		//save old editbox
		diff := (orig_value != value)

		if diff {
			in := LayoutInput{SetEdit: true, EditValue: edit.temp, EditEnter: enter_key}
			dom.ui.parent.CallInput(&dom.props, &in)
		}
	}

	//reset
	edit.hash = 0
	edit.temp = ""
	edit.activate_next_iters = 0
	edit.RefreshTicks = 0
	edit.ResetKeys()

	//set new
	if !finish {
		edit.hash = dom.props.Hash
		edit.temp = value

		edit.start = 0
		edit.end = len(value)

		dom.ui.GetWin().SetTextCursorMove()
	}

	if !finish {
		dom.ui.SetRedraw() //refresh border thickness
	}
}

func (edit *UiLayoutEdit) ResetKeys() {
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
