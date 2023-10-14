//go:build wasi

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

import (
	"unsafe"
)

func main() {}

//export _sa_open
//go:wasmexport _sa_open
func _sa_open() {
	Open()
}

//export _sa_setup_db
//go:wasmexport _sa_setup_db
func _sa_setup_db() {
	SetupDB()
}

//export _sa_save
//go:wasmexport _sa_save
func _sa_save() {
	_sa_storage_write(_SA_bytesToPtr(Save()))
}

//export _sa_render
//go:wasmexport _sa_render
func _sa_render() {
	Render()
}

//export _sa_storage_write
//go:wasmimport env _sa_storage_write
func _sa_storage_write(jsonMem SAMem) int64

//export _sa_info_get_prepare
//go:wasmimport env _sa_info_get_prepare
func _sa_info_get_prepare(keyMem SAMem, prm1Mem SAMem, prm2Mem SAMem) int64

//export _sa_info_get
//go:wasmimport env _sa_info_get
func _sa_info_get(dstMem SAMem) int64

//export _sa_info_set
//go:wasmimport env _sa_info_set
func _sa_info_set(keyMem SAMem, prm1Mem SAMem, prm2Mem SAMem, prm3Mem SAMem) int64

//export _sa_blob
//go:wasmimport env _sa_blob
func _sa_blob(pathMem SAMem, dstMem SAMem) int64

//export _sa_blob_len
//go:wasmimport env _sa_blob_len
func _sa_blob_len(pathMem SAMem) int64

//export _sa_sql_commit
//go:wasmimport env _sa_sql_commit
func _sa_sql_commit(dbUrlMem SAMem) int64

//export _sa_sql_rollback
//go:wasmimport env _sa_sql_rollback
func _sa_sql_rollback(dbUrlMem SAMem) int64

//export _sa_sql_write
//go:wasmimport env _sa_sql_write
func _sa_sql_write(dbUrlMem SAMem, queryMem SAMem) int64

//export _sa_sql_read
//go:wasmimport env _sa_sql_read
func _sa_sql_read(dbUrlMem SAMem, queryMem SAMem) int64

//export _sa_sql_readRowCount
//go:wasmimport env _sa_sql_readRowCount
func _sa_sql_readRowCount(dbUrlMem SAMem, queryMem SAMem, queryHash int64) int64

//export _sa_sql_readRowLen
//go:wasmimport env _sa_sql_readRowLen
func _sa_sql_readRowLen(dbUrlMem SAMem, queryMem SAMem, queryHash int64, row_i uint64) int64

//export _sa_sql_readRow
//go:wasmimport env _sa_sql_readRow
func _sa_sql_readRow(dbUrlMem SAMem, queryMem SAMem, queryHash int64, row_i uint64, resultMem SAMem) int64

//export _sa_div_colResize
//go:wasmimport env _sa_div_colResize
func _sa_div_colResize(pos uint64, nameMem SAMem, val float64) float64

//export _sa_div_rowResize
//go:wasmimport env _sa_div_rowResize
func _sa_div_rowResize(pos uint64, nameMem SAMem, val float64) float64

//export _sa_div_colMax
//go:wasmimport env _sa_div_colMax
func _sa_div_colMax(pos uint64, val float64) float64

//export _sa_div_rowMax
//go:wasmimport env _sa_div_rowMax
func _sa_div_rowMax(pos uint64, val float64) float64

//export _sa_div_col
//go:wasmimport env _sa_div_col
func _sa_div_col(pos uint64, val float64) float64

//export _sa_div_row
//go:wasmimport env _sa_div_row
func _sa_div_row(pos uint64, val float64) float64

//export _sa_div_start
//go:wasmimport env _sa_div_start
func _sa_div_start(x, y, w, h uint64, nameMem SAMem) int64

//export _sa_div_end
//go:wasmimport env _sa_div_end
func _sa_div_end()

//export _sa_div_dialogOpen
//go:wasmimport env _sa_div_dialogOpen
func _sa_div_dialogOpen(nameMem SAMem, tp uint64) int64

//export _sa_div_dialogClose
//go:wasmimport env _sa_div_dialogClose
func _sa_div_dialogClose()

//export _sa_div_dialogStart
//go:wasmimport env _sa_div_dialogStart
func _sa_div_dialogStart(nameMem SAMem) int64

//export _sa_div_dialogEnd
//go:wasmimport env _sa_div_dialogEnd
func _sa_div_dialogEnd()

//export _sa_div_get_info
//go:wasmimport env _sa_div_get_info
func _sa_div_get_info(idMem SAMem, x int64, y int64) float64

//export _sa_div_set_info
//go:wasmimport env _sa_div_set_info
func _sa_div_set_info(idMem SAMem, val float64, x int64, y int64) float64

//export _sa_paint_rect
//go:wasmimport env _sa_paint_rect
func _sa_paint_rect(x, y, w, h float64, margin float64, r, g, b, a uint32, borderWidth float64) int64

//export _sa_paint_line
//go:wasmimport env _sa_paint_line
func _sa_paint_line(x, y, w, h float64, margin float64, sx, sy, ex, ey float64, r, g, b, a uint32, width float64) int64

//export _sa_paint_circle
//go:wasmimport env _sa_paint_circle
func _sa_paint_circle(x, y, w, h float64, margin float64, sx, sy, rad float64, r, g, b, a uint32, borderWidth float64) int64

//export _sa_paint_file
//go:wasmimport env _sa_paint_file
func _sa_paint_file(x, y, w, h float64, fileMem SAMem, tooltipMem SAMem, margin, marginX, marginY float64, r, g, b, a uint32, alignV, alignH uint32, fill uint32) int64

//export _sa_paint_text
//go:wasmimport env _sa_paint_text
func _sa_paint_text(x, y, w, h float64, styleId uint32, valueMem SAMem, selection, edit, enable uint32) int64

//export _sa_paint_textWidth
//go:wasmimport env _sa_paint_textWidth
func _sa_paint_textWidth(styleId uint32, valueMem SAMem, cursorPos int64) float64

//export _sa_paint_tooltip
//go:wasmimport env _sa_paint_tooltip
func _sa_paint_tooltip(x, y, w, h float64, valueMem SAMem) int64

//export _sa_paint_cursor
//go:wasmimport env _sa_paint_cursor
func _sa_paint_cursor(nameMem SAMem) int64

//export _sa_print
//go:wasmimport env _sa_print
func _sa_print(mem SAMem)

//export _sa_print_float
//go:wasmimport env _sa_print_float
func _sa_print_float(val float64)

//export _sa_comp_drawButton
//go:wasmimport env _sa_comp_drawButton
func _sa_comp_drawButton(style uint32, valueMem SAMem, iconMem SAMem, icon_margin float64, urlMem SAMem, tooltipMem SAMem, enable uint32, outMem SAMem) int64

//export _sa_comp_drawSlider
//go:wasmimport env _sa_comp_drawSlider
func _sa_comp_drawSlider(styleTrackId uint32, styleThumbId uint32, value float64, min float64, max float64, jump float64, tooltipMem SAMem, enable uint32, outMem SAMem) float64

//export _sa_comp_drawProgress
//go:wasmimport env _sa_comp_drawProgress
func _sa_comp_drawProgress(styleFrameId uint32, styleStatusId uint32, value float64, prec int32, tooltipMem SAMem, enable uint32) int64

//export _sa_comp_drawText
//go:wasmimport env _sa_comp_drawText
func _sa_comp_drawText(style uint32, valueMem SAMem, tooltipMem SAMem, enable uint32, selection uint32) int64

//export _sa_comp_getEditValue
//go:wasmimport env _sa_comp_getEditValue
func _sa_comp_getEditValue(outMem SAMem) int64

//export _sa_comp_drawEdit
//go:wasmimport env _sa_comp_drawEdit
func _sa_comp_drawEdit(style uint32, valueMem SAMem, valueOrigMem SAMem, tooltipMem SAMem, ghostMem SAMem, enable uint32, outMem SAMem) int64

//export _sa_comp_drawCombo
//go:wasmimport env _sa_comp_drawCombo
func _sa_comp_drawCombo(styleId uint32, styleMenuId uint32, value uint64, optionsMem SAMem, tooltipMem SAMem, enable uint32) int64

//export _sa_comp_drawCheckbox
//go:wasmimport env _sa_comp_drawCheckbox
func _sa_comp_drawCheckbox(styleCheckId uint32, styleLabelId uint32, value uint64, labelMem SAMem, tooltipMem SAMem, enable uint32) int64

//export _sa_register_style
//go:wasmimport env _sa_register_style
func _sa_register_style(jsMem SAMem) int64

//export _sa_div_drag
//go:wasmimport env _sa_div_drag
func _sa_div_drag(groupNameMem SAMem, id uint64) int64

//export _sa_div_drop
//go:wasmimport env _sa_div_drop
func _sa_div_drop(groupNameMem SAMem, vertical uint32, horizontal uint32, inside uint32, outMem SAMem) int64

//export _sa_render_app
//go:wasmimport env _sa_render_app
func _sa_render_app(dbUrlMem SAMem, app_rowid uint64) int64

type SAMem uint64

func _SA_ptrToBytes(mem SAMem) []byte {
	ptr := uint32(mem >> 32)
	size := uint32(mem)
	return unsafe.Slice((*byte)(unsafe.Pointer(uintptr(ptr))), size)
}

func _SA_stringToPtr(s string) SAMem {
	if len(s) > 0 {
		ptr := unsafe.Pointer(unsafe.StringData(s))
		return SAMem((uint64(uintptr(ptr)) << uint64(32)) | uint64(len(s)))
	}
	return 0
}
func _SA_bytesToPtr(s []byte) SAMem {
	if len(s) > 0 {
		ptr := unsafe.Pointer(unsafe.SliceData(s))
		return SAMem((uint64(uintptr(ptr)) << uint64(32)) | uint64(len(s)))
	}
	return 0
}

func _SA_ptrToString(mem SAMem) string {
	ptr := uint32(mem >> 32)
	size := uint32(mem)
	return unsafe.String((*byte)(unsafe.Pointer(uintptr(ptr))), size)
}

/*func _SA_bytes64ToPtr(s []uint64) uint64 {
	ptr := unsafe.Pointer(unsafe.SliceData(s))
	return (uint64(uintptr(ptr)) << uint64(32)) | uint64(len(s)*8)
}*/

func _SA_DebugLine() {
	//empty for .wasm
	//no export neede
}
