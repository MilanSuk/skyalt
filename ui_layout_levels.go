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
	"encoding/json"
	"os"
)

type UiDomLevel struct {
	ui *Ui

	domPtr    *Layout3
	parentDom *Layout3 //dialog is relative

	parentTouch OsV2
}

func NewUiLayoutLevel(domPtr *Layout3, parentDom *Layout3, parentTouch OsV2, ui *Ui) *UiDomLevel {
	var level UiDomLevel

	level.ui = ui

	level.domPtr = domPtr
	level.parentDom = parentDom

	level.parentTouch = parentTouch

	return &level
}

func (level *UiDomLevel) Destroy() {
}

func (level *UiDomLevel) GetCoord(q OsV4) OsV4 {

	winRect := level.ui.winRect

	if level.parentDom != nil {
		// relative
		q = OsV4_relativeSurround(level.parentDom.crop, q, winRect, false)
	} else {

		if level.parentTouch.IsZero() {
			// center
			q = OsV4_center(winRect, q.Size)
		} else {
			// touch
			q = OsV4_relativeSurround(OsV4{level.parentTouch, OsV2{1, 1}}, q, winRect, false)
		}

	}
	return q
}

type UiLevels struct {
	ui           *Ui
	levels       []*UiDomLevel
	on_top_level bool

	settings UiSettings
}

func InitUiLevels(ui *Ui) UiLevels {
	levels := UiLevels{ui: ui}

	levels.levels = append(levels.levels, NewUiLayoutLevel(ui.dom, nil, OsV2{}, ui))

	levels.on_top_level = true

	levels.settings.Layouts.Init()

	return levels
}

func (levels *UiLevels) Destroy() {
	for _, l := range levels.levels {
		l.Destroy()
	}
}

func (levels *UiLevels) GetUiPath() string {
	return "layouts/Root.json"
}

func (levels *UiLevels) Open() error {

	//open file
	jsRead, err := os.ReadFile(levels.GetUiPath())
	if err != nil {
		return err
	}

	if len(jsRead) > 0 {
		err := json.Unmarshal(jsRead, &levels.settings)
		if err != nil {
			return err
		}
	}

	return nil
}

func (levels *UiLevels) Save() error {

	js, err := json.MarshalIndent(levels.settings, "", "\t")
	if err != nil {
		return err
	}

	err = os.WriteFile(levels.GetUiPath(), js, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (levels *UiLevels) FindLevel(dom *Layout3) *UiDomLevel {
	for _, l := range levels.levels {
		if l.domPtr == dom {
			return l
		}
	}
	return nil
}

func (levels *UiLevels) GetTopLevel() *UiDomLevel {
	return levels.levels[len(levels.levels)-1]
}

func (levels *UiLevels) IsLevelTop(dom *Layout3) bool {
	if !levels.on_top_level {
		return false
	}

	for dom.parentLevel != nil {
		dom = dom.parentLevel
	}

	return levels.GetTopLevel().domPtr == dom
}
func (levels *UiLevels) IsDialogTop(layout *Layout3) bool {
	if !levels.on_top_level {
		return false
	}

	return levels.settings.Dialogs[len(levels.settings.Dialogs)-1].Path == layout.GetPath()
}

func (levels *UiLevels) IsDialog(layout *Layout3) bool {

	path := layout.GetPath()
	for _, dlg := range levels.settings.Dialogs {
		if dlg.Path == path {
			return true
		}
	}
	return false
}

func (levels *UiLevels) OpenDialog(layout *Layout3, parent *Layout3, parentTouch OsV2) UiDialog {

	layout = layout.GetLevel()

	Path := layout.GetPath()
	ParentPath := ""
	if parent != nil {
		ParentPath = parent.GetPath()
	}

	dia := UiDialog{Path: Path, ParentPath: ParentPath, ParentTouch: parentTouch}

	levels.settings.Dialogs = append(levels.settings.Dialogs, dia)

	levels.ui.SetRelayout()

	return dia
}

func (levels *UiLevels) CloseAllDialogs() {
	levels.settings.Dialogs = nil

}
func (levels *UiLevels) CloseTopDialog() bool {
	if len(levels.settings.Dialogs) > 0 {
		levels.settings.Dialogs = levels.settings.Dialogs[:len(levels.settings.Dialogs)-1]
		return true
	}
	return false
}

func (levels *UiLevels) CloseDialog(layout *Layout3) bool {

	layout = layout.GetLevel()

	path := layout.GetPath()

	found := false
	for i := len(levels.settings.Dialogs) - 1; i >= 0; i-- {
		if levels.settings.Dialogs[i].Path == path {
			levels.settings.Dialogs = append(levels.settings.Dialogs[:i], levels.settings.Dialogs[i+1:]...) //remove
		}
	}
	return found
}

func (levels *UiLevels) _closeDialog(layout *Layout3) bool {

	path := layout.GetPath()

	for i, dlg := range levels.settings.Dialogs {
		if dlg.Path == path {
			levels.settings.Dialogs = levels.settings.Dialogs[:i] //close all above
			return true
		}
	}
	return false
}

func (levels *UiLevels) CloseTouchDialogs() {

	ui := levels.ui
	win := ui.GetWin()

	for i, dia := range levels.levels {
		if i > 0 {
			//close dialog
			if levels.IsLevelTop(dia.domPtr) && !levels.ui.parent.touch.IsAnyActive() {
				outside := ui.winRect.Inside(win.io.Touch.Pos) && !dia.domPtr.CropWithScroll().Inside(win.io.Touch.Pos)

				if (win.io.Touch.End && outside) || win.io.Keys.Esc {
					levels._closeDialog(dia.domPtr)

					ui.parent.ResetIO()
					win.io.Keys.Esc = false
					win.io.Touch.End = false
				}
			}
		}
	}
}

func (levels *UiLevels) TryCloseDialogs() {

	rebuild := false

	//compare
	if 1+len(levels.settings.Dialogs) != len(levels.levels) {
		rebuild = true
	} else {
		for i := 1; i < len(levels.levels); i++ {
			if levels.levels[i].domPtr.GetPath() != levels.settings.Dialogs[i-1].Path {
				rebuild = true
			}
		}
	}

	if !rebuild {
		return
	}

	//add levels
	levels.levels = levels.levels[:1] //keep base
	for i, dlg := range levels.settings.Dialogs {
		dom := levels.ui.dom.FindPath(dlg.Path)
		if dom != nil {
			domParent := levels.ui.dom.FindPath(dlg.ParentPath)
			levels.levels = append(levels.levels, NewUiLayoutLevel(dom, domParent, dlg.ParentTouch, levels.ui))
		} else {
			levels.settings.Dialogs[i].Path = "" // reset
		}
	}

	//remove dialog which don't have valid path
	for i := len(levels.settings.Dialogs) - 1; i >= 0; i-- {
		if levels.settings.Dialogs[i].Path == "" {
			levels.settings.Dialogs = append(levels.settings.Dialogs[:i], levels.settings.Dialogs[i+1:]...) //remove
		}
	}

	if rebuild {
		levels.ui.SetRefresh()
	}
}

func (levels *UiLevels) Relayout() {
	for _, l := range levels.levels {
		l.domPtr.Relayout(true)
	}
}

func (levels *UiLevels) ClearUpperLevels() {
	for i, l := range levels.levels {
		if i > 0 {
			l.Destroy()
		}
	}
	levels.levels = levels.levels[:1] //keep 1st
}

func (levels *UiLevels) drawBuffer() {

	ui := levels.ui
	win := ui.GetWin()

	for i, dia := range levels.levels {
		if i == 0 {
			dia.domPtr.Draw()
		} else {
			win.buff.StartLevel(dia.domPtr.crop, ui.GetPalette().B, ui.dom.canvas)
			dia.domPtr.Draw() //add renderToTexture optimalization ...
		}
	}
}
