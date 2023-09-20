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

import "fmt"

type LayoutLevels struct {
	dialogs []*LayoutLevel
	calls   []*LayoutLevel

	infoLayout RS_LScroll
}

func NewLayoutLevels(scrollPath string, ui *Ui) (*LayoutLevels, error) {

	var levels LayoutLevels

	//scroll
	err := levels.infoLayout.Open(scrollPath)
	if err != nil {
		return nil, fmt.Errorf("Open() failed: %w", err)
	}

	levels.AddDialog("", OsV4{}, ui)

	return &levels, nil
}

func (levels *LayoutLevels) Destroy(scrollPath string) {

	levels.dialogs[0].rootDiv.Save(&levels.infoLayout)
	err := levels.infoLayout.Save(scrollPath)
	if err != nil {
		fmt.Printf("Open() failed: %v\n", err)
	}

	for _, l := range levels.dialogs {
		l.Destroy()
	}
	levels.dialogs = nil
	levels.calls = nil
}

func (levels *LayoutLevels) AddDialog(name string, src_coordMoveCut OsV4, ui *Ui) {

	newDialog := NewLayoutLevel(name, src_coordMoveCut, &levels.infoLayout, ui)
	levels.dialogs = append(levels.dialogs, newDialog)

	//disable bottom dialogs
	for _, l := range levels.calls {
		enabled := (l == newDialog)
		div := l.stack
		for div != nil {
			div.enableInput = enabled
			div = div.parent
		}
	}

}

func (levels *LayoutLevels) StartCall(lev *LayoutLevel) {
	//init level
	lev.stack = lev.rootDiv

	//add
	levels.calls = append(levels.calls, lev)
}
func (levels *LayoutLevels) EndCall() error {

	n := len(levels.calls)
	if n > 1 {
		levels.calls = levels.calls[:n-1]
		return nil
	}

	return fmt.Errorf("trying to EndCall from root level")
}

func (levels *LayoutLevels) isSomeClose() bool {
	for _, l := range levels.dialogs {
		if l.use == 0 || l.close {
			return true
		}
	}
	return false
}

func (levels *LayoutLevels) Maintenance() {

	levels.GetBaseDialog().use = 1 //base level is always use

	//remove unused or closed
	if levels.isSomeClose() {
		var lvls []*LayoutLevel
		for _, l := range levels.dialogs {
			if l.use != 0 && !l.close {
				lvls = append(lvls, l)
			}
		}
		levels.dialogs = lvls

	}

	//layout
	for _, l := range levels.dialogs {
		l.rootDiv.Maintenance(&levels.infoLayout)
		l.use = 0
	}
}

func (levels *LayoutLevels) Draw() {

	for _, l := range levels.dialogs {
		if l.buff != nil {
			l.buff.Draw()
		}
	}
}

func (levels *LayoutLevels) CloseAndAbove(dialog *LayoutLevel) {

	found := false
	for _, l := range levels.dialogs {
		if l == dialog {
			found = true
		}
		if found {
			l.close = true
		}
	}
}
func (levels *LayoutLevels) CloseAll() {

	if len(levels.dialogs) > 1 {
		levels.CloseAndAbove(levels.dialogs[1])
	}
}

func (levels *LayoutLevels) GetBaseDialog() *LayoutLevel {
	return levels.dialogs[0]
}

func (levels *LayoutLevels) GetStack() *LayoutLevel {
	return levels.calls[len(levels.calls)-1] //last call
}

func (levels *LayoutLevels) IsStackTop() bool {
	return levels.dialogs[len(levels.dialogs)-1] == levels.GetStack() //last dialog
}

func (levels *LayoutLevels) ResetStack() {
	levels.calls = nil
	levels.StartCall(levels.GetBaseDialog())
}

func (levels *LayoutLevels) Find(name string) *LayoutLevel {

	for _, l := range levels.dialogs {
		if l.name == name {
			return l
		}
	}
	return nil
}
