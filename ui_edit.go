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

import "fmt"

type UiEdit struct {
	uid                    uint64
	activate_next_iters    int    //after pressing tab
	activate_next_uid      uint64 //force to activate
	activate_cursor_at_end bool

	orig_value string

	temp       string
	editable   bool //save
	start, end int

	RefreshTicks int64

	KeySelectAll bool
	KeyCopy      bool
	KeyCut       bool
	KeyPaste     bool
	KeyRecord    bool

	last_refreshInput_unix float64

	shortcut_triggered bool
}

func (edit *UiEdit) send(enter_key bool, ui *Ui) {
	lay := ui.mainLayout.FindUID(edit.uid)
	if lay != nil {
		if lay.fnSetEditbox != nil {
			lay.fnSetEditbox(edit.temp, enter_key)
		}
	} else {
		fmt.Println("Not found")
	}
}

func (edit *UiEdit) Set(uid uint64, editable bool, orig_value, value string, enter_key bool, finish bool, save bool, refreshInput bool, ui *Ui) {
	if !finish && edit.uid == uid {
		if editable {
			edit.temp = value //refresh after code

			tm := OsTime()
			if refreshInput && (edit.orig_value != edit.temp) && (edit.last_refreshInput_unix+1 < OsTime()) {

				edit.orig_value = edit.temp
				edit.send(false, ui)
				edit.last_refreshInput_unix = tm
			}
		}

		ui.GetWin().SetTextCursorMove()
		return //no change
	}

	if edit.editable && save && edit.uid != 0 {
		//save old editbox
		if (edit.orig_value != edit.temp) || enter_key {
			edit.send(enter_key, ui)
		}
	}

	//reset
	edit.uid = 0
	edit.temp = ""
	edit.editable = false
	edit.activate_next_iters = 0
	edit.activate_next_uid = 0
	edit.RefreshTicks = 0
	edit.ResetShortcutKeys()

	//set new
	if !finish {
		edit.uid = uid
		edit.temp = value
		edit.editable = editable
		edit.orig_value = orig_value

		edit.start = 0
		edit.end = len(value)
		if edit.activate_cursor_at_end {
			edit.start = len(value)
		}

		ui.GetWin().SetTextCursorMove()
	}

	if !finish {
		ui.SetRedrawBuffer() //refresh border thickness
	}
}

func (edit *UiEdit) ResetShortcutKeys() {
	edit.KeySelectAll = false
	edit.KeyCopy = false
	edit.KeyCut = false
	edit.KeyPaste = false
	edit.KeyRecord = false
}

func (edit *UiEdit) Is(layout *Layout) bool {
	return edit.IsUID(layout.UID)
}
func (edit *UiEdit) IsUID(uid uint64) bool {
	return edit.uid == uid
}

func (edit *UiEdit) IsActivateNext() bool {
	return edit.activate_next_iters != 0
}
func (edit *UiEdit) SetActivateTabNext() {
	edit.activate_next_iters = 2
}
func (edit *UiEdit) SetActivate(uid uint64) {
	edit.activate_next_uid = uid
	edit.activate_next_iters = 2
	edit.activate_cursor_at_end = true

	edit.shortcut_triggered = false
}

func (edit *UiEdit) SetRefreshTicks() {
	if edit.RefreshTicks == 0 {
		edit.RefreshTicks = OsTicks()
	}
}

func (edit *UiEdit) Tick() {
	if edit.activate_next_iters > 0 {
		edit.activate_next_iters--
	}
}

func (edit *UiEdit) IsActive() bool {
	return edit.uid != 0
}
