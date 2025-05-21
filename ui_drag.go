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

type UiDrag struct {
	srcUID uint64
	dstUID uint64

	group string

	pos SA_Drop_POS
}

func (drag *UiDrag) Reset() {
	drag.srcUID = 0
	drag.dstUID = 0
	drag.group = ""
}
func (drag *UiDrag) Set(layout *Layout) {
	drag.srcUID = layout.UID
	drag.group = layout.Drag_group
}
func (drag *UiDrag) IsDraged(layout *Layout) bool {
	return drag.srcUID == layout.UID
}
func (drag *UiDrag) IsDroped(layout *Layout) bool {
	return drag.dstUID == layout.UID
}

func (drag *UiDrag) IsOverDrop(layout *Layout) bool {
	return drag.group != "" && drag.group == layout.Drop_group
}

type SA_Drop_POS int

const (
	SA_Drop_INSIDE  SA_Drop_POS = 0
	SA_Drop_V_LEFT  SA_Drop_POS = 1
	SA_Drop_V_RIGHT SA_Drop_POS = 2
	SA_Drop_H_LEFT  SA_Drop_POS = 3
	SA_Drop_H_RIGHT SA_Drop_POS = 4
)

func UiDrag_MoveElementIndex(src int, dst int, pos SA_Drop_POS, diff_sources bool) int {
	if diff_sources {
		src = dst + 1000 //higher than dst
	}

	//check
	if src < dst && (pos == SA_Drop_V_LEFT || pos == SA_Drop_H_LEFT) {
		dst--
	}
	if src > dst && (pos == SA_Drop_V_RIGHT || pos == SA_Drop_H_RIGHT) {
		dst++
	}
	return dst
}

func Layout_MoveElement[T any](array_src *[]T, array_dst *[]T, src int, dst int) {
	//move(by swap one-by-one)
	if array_src == array_dst {
		for i := src; i < dst; i++ {
			(*array_dst)[i], (*array_dst)[i+1] = (*array_dst)[i+1], (*array_dst)[i]
		}
		for i := src; i > dst; i-- {
			(*array_dst)[i], (*array_dst)[i-1] = (*array_dst)[i-1], (*array_dst)[i]
		}
	} else {
		backup := (*array_src)[src]

		//remove
		*array_src = append((*array_src)[:src], (*array_src)[src+1:]...)

		//insert
		if dst < len(*array_dst) {
			*array_dst = append((*array_dst)[:dst+1], (*array_dst)[dst:]...)
			(*array_dst)[dst] = backup
		} else {
			*array_dst = append(*array_dst, backup)
			//dst = len(*array_dst) - 1
		}
	}
}
