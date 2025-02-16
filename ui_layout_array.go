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

type UiLayoutArrayRes struct {
	value float32
}

type UiLayoutArrayItem struct {
	min    float32
	max    float32
	resize *UiLayoutArrayRes
}

type UiLayoutArray struct {
	inputs  []UiLayoutArrayItem
	outputs []int32

	fills []bool
}

func (arr *UiLayoutArray) Clear() {
	arr.inputs = arr.inputs[:0]
}

func (arr *UiLayoutArray) NumInputs() int {
	return len(arr.inputs)
}

func (arr *UiLayoutArray) Resize(num int) {
	//huge amount of RAM if I set few items on position 1000 - all before are alocated and set & reset after frame is rendered ...
	for i := arr.NumInputs(); i < num; i++ {
		arr.inputs = append(arr.inputs, UiLayoutArrayItem{min: 1, max: 0, resize: nil})
	}
}

func (arr *UiLayoutArray) SetFills(childs []*Layout3, cols bool) {
	arr.fills = make([]bool, arr.NumInputs())

	for i := range arr.fills {
		for _, it := range childs {
			if it.IsShown() {
				if (cols && i >= it.props.X && i < it.props.X+it.props.W) ||
					(!cols && i >= it.props.Y && i < it.props.Y+it.props.H) {
					arr.fills[i] = true
					break
				}
			}
		}
	}
}

func LayoutArray_resizerSize(cell int) int {
	v := cell / 4
	if v < 9 {
		return 9
	}
	return v
}

func (arr *UiLayoutArray) IsLastResizeValid() bool {
	n := arr.NumInputs()
	return n >= 2 && arr.inputs[n-2].resize == nil && arr.inputs[n-1].resize != nil
}

func (arr *UiLayoutArray) GetResizeIndex(i int) int {

	if arr.IsLastResizeValid() {
		if i+2 == arr.NumInputs() {
			if arr.inputs[i+1].resize != nil {
				return i + 1 // show resizer before column/row
			}
			return -1
		}
		if i+1 == arr.NumInputs() {
			return -1 // last was return as Previous, so no last
		}
	}

	if i < arr.NumInputs() {
		if arr.inputs[i].resize != nil {
			return i
		}
	}
	return -1
}

func (arr *UiLayoutArray) Convert(cell int, start int, end int) OsV2 {

	var ret OsV2

	for i := 0; i < end; i++ {
		ok := (i < len(arr.outputs))

		if i < start {
			if ok {
				ret.X += int(arr.outputs[i])
			} else {
				ret.X += cell
			}
		} else {
			if ok {
				ret.Y += int(arr.outputs[i])
			} else {
				ret.Y += cell
			}
		}
	}

	if end > 0 && (end-1 < arr.NumInputs()) && arr.GetResizeIndex(int(end)-1) >= 0 {
		ret.Y -= LayoutArray_resizerSize(cell)
	}

	return ret
}

func (arr *UiLayoutArray) ConvertMax(cell int, start int, end int) OsV2 {
	var ret OsV2

	for i := 0; i < end; i++ {
		ok := (i < arr.NumInputs())

		var v int
		if ok {
			v = int(OsMaxFloat32(arr.inputs[i].min, arr.inputs[i].max) * float32(cell))
		} else {
			v = cell
		}

		if i < start {
			ret.X += v
		} else {
			ret.Y += v
		}
	}

	return ret
}

func (arr *UiLayoutArray) GetCellPos(rel_px_pos int, cell int) int {
	if rel_px_pos < 0 {
		return 0
	}
	allPixels := 0
	allPixelsLast := 0
	for i := 0; i < len(arr.outputs); i++ {
		allPixels += int(arr.outputs[i])

		if rel_px_pos >= allPixelsLast && rel_px_pos < allPixels {
			return i //found
		}

		allPixelsLast = allPixels
	}

	return len(arr.outputs) + (rel_px_pos-allPixelsLast)/cell
}

func (arr *UiLayoutArray) GetCloseCellPos(rel_px_pos int) int {
	if rel_px_pos < 0 {
		return 0
	}
	allPixels := 0
	allPixelsLast := 0
	for i := 0; i < len(arr.outputs); i++ {
		allPixels += int(arr.outputs[i])

		if rel_px_pos >= allPixelsLast && rel_px_pos < allPixels {
			return i //found
		}

		allPixelsLast = allPixels
	}

	return len(arr.outputs)
}

func (arr *UiLayoutArray) GetResizerPos(i int, cell int) int {
	if i >= len(arr.outputs) {
		return 0
	}

	allPixels := 0
	for ii := 0; ii <= i; ii++ {
		allPixels += int(arr.outputs[ii])
	}

	return allPixels - LayoutArray_resizerSize(cell)
}

func (arr *UiLayoutArray) IsResizerTouch(touchPos int, cell int) int {
	space := LayoutArray_resizerSize(cell)

	for i := 0; i < arr.NumInputs(); i++ {
		if arr.GetResizeIndex(i) >= 0 {
			p := arr.GetResizerPos(i, cell)
			if touchPos > p && touchPos < p+space {
				return i
			}
		}
	}
	return -1
}

func (arr *UiLayoutArray) OutputAll() int {
	sum := 0
	for i := 0; i < len(arr.outputs); i++ {
		sum += int(arr.outputs[i])
	}
	return sum
}

func (arr *UiLayoutArray) makeLarger(cell int, window int, allPixels int, fill bool) int {

	hasSpace := (len(arr.outputs) > 0)
	for allPixels < window && hasSpace {
		rest := window - allPixels
		tryAdd := OsMax(1, rest/int(len(arr.outputs)))

		hasSpace = false
		for i := 0; i < len(arr.outputs) && allPixels < window; i++ {
			if arr.fills[i] == fill {
				maxAdd := int(arr.inputs[i].max*float32(cell)) - int(arr.outputs[i])
				add := OsClamp(tryAdd, 0, maxAdd)

				arr.outputs[i] += int32(add)
				allPixels += add

				if maxAdd > tryAdd {
					hasSpace = true
				}
			}
		}
	}
	return allPixels
}

func (arr *UiLayoutArray) Update(cell int, window int) {

	arr.outputs = make([]int32, arr.NumInputs())

	//project in -> out
	for i := 0; i < len(arr.inputs); i++ {
		//min
		minV := float64(arr.inputs[i].min)
		minV = OsClampFloat(minV, 0.001, 100000000)

		if arr.inputs[i].resize == nil {
			// max
			maxV := minV
			if arr.inputs[i].max > 0 {
				maxV = float64(arr.inputs[i].max)
				maxV = OsMaxFloat(minV, maxV)
			}

			arr.inputs[i].min = float32(minV)
			arr.inputs[i].max = float32(maxV)
		} else {
			resV := float64(arr.inputs[i].resize.value)
			resV = OsMaxFloat(resV, minV)

			if arr.inputs[i].max > 0 {
				maxV := float64(arr.inputs[i].max)
				maxV = OsMaxFloat(minV, maxV)

				resV = OsClampFloat(resV, minV, maxV)
			}

			arr.inputs[i].min = float32(resV)
			arr.inputs[i].max = float32(resV)

			arr.inputs[i].resize.value = float32(resV)
		}
	}

	//sum
	allPixels := 0
	for i := 0; i < len(arr.outputs); i++ {
		arr.outputs[i] = int32(arr.inputs[i].min * float32(cell))
		allPixels += int(arr.outputs[i])
	}

	//make larger(when maxs allow)
	allPixels = arr.makeLarger(cell, window, allPixels, true) //fills
	arr.makeLarger(cell, window, allPixels, false)            //non-fills
}

func (arr *UiLayoutArray) HasResize() bool {
	for _, c := range arr.inputs {
		if c.resize != nil {
			return true
		}
	}
	return false
}

func (arr *UiLayoutArray) GetOutput(i int) int {

	if i < len(arr.outputs) {
		return int(arr.outputs[i])
	}
	return -1
}

func (arr *UiLayoutArray) GetSumOutput(i int) int32 {

	if i < 0 {
		i = len(arr.outputs)
	}

	p := int32(0)
	for ii, it := range arr.outputs {
		if ii >= i {
			break
		}
		p += it
	}
	return p
}

func (arr *UiLayoutArray) findInput(pos int) *UiLayoutArrayItem {

	if pos < len(arr.inputs) {
		return &arr.inputs[pos]
	}

	return nil
}

func (arr *UiLayoutArray) findOrAdd(pos int) *UiLayoutArrayItem {

	if pos >= len(arr.inputs) {
		arr.Resize(pos + 1)
	}

	return &arr.inputs[pos]
}
