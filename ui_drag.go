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

type UiLayoutDrag struct {
	srcHash uint64
	dstHash uint64

	group string

	pos SA_Drop_POS
}

func (drag *UiLayoutDrag) Reset() {
	drag.srcHash = 0
	drag.dstHash = 0
	drag.group = ""
}
func (drag *UiLayoutDrag) Set(dom *Layout3) {
	drag.srcHash = dom.props.Hash
	drag.group = dom.props.Drag_group
}
func (drag *UiLayoutDrag) IsDraged(dom *Layout3) bool {
	return drag.srcHash == dom.props.Hash
}
func (drag *UiLayoutDrag) IsDroped(dom *Layout3) bool {
	return drag.dstHash == dom.props.Hash
}

func (drag *UiLayoutDrag) IsOverDrop(dom *Layout3) bool {
	return drag.group != "" && drag.group == dom.props.Drop_group
}
