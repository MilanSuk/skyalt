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
	dom   *Layout3
	group string
}

func (drag *UiLayoutDrag) Reset() {
	drag.dom = nil
	drag.group = ""
}
func (drag *UiLayoutDrag) Set(dom *Layout3) {
	drag.dom = dom
	drag.group = dom.props.Drag_group
}
func (drag *UiLayoutDrag) IsDraged(dom *Layout3) bool {
	return drag.dom == dom
}
func (drag *UiLayoutDrag) IsOverDrop(dom *Layout3) bool {
	return drag.group != "" && drag.group == dom.props.Drop_group
}
