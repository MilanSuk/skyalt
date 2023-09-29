package main

import (
	"encoding/json"
	"reflect"
	"unsafe"
)

func main() {}

//export _sa_render
func _sa_render() uint32 {
	return Render()
}

//export _sa_init
func _sa_init(storeMem SAMem) {
	jsStore := _SA_ptrToBytes(storeMem)
	if !Open(jsStore) {
		json.Unmarshal(jsStore, &store)
	}
}

//export _sa_exit
func _sa_exit() {
	js, written := Save()
	if !written {
		js, _ = json.MarshalIndent(&store, "", "")
	}
	_sa_storage_write(_SA_bytesToPtr(js))
}

//export _sa_storage_write
func _sa_storage_write(jsonMem SAMem) int64

//export _sa_styles
func _sa_styles(jsonMem SAMem) {
	jsStyles := _SA_ptrToBytes(jsonMem)
	styles = SA_Styles{} //reset ids
	json.Unmarshal(jsStyles, &styles)
	Styles()
}

//export _sa_translations
func _sa_translations(jsonMem SAMem) {
	e := reflect.ValueOf(&trns).Elem()
	for i := 0; i < e.NumField(); i++ {
		e.Field(i).SetString("{" + e.Type().Field(i).Name + "}")
	}

	json.Unmarshal(_SA_ptrToBytes(jsonMem), &trns)
}

//export _sa_info_float
func _sa_info_float(keyMem SAMem) float64

//export _sa_info_setFloat
func _sa_info_setFloat(keyMem SAMem, value float64) int64

//export _sa_info_string
func _sa_info_string(keyMem SAMem, dstMem SAMem) int64

//export _sa_info_string_len
func _sa_info_string_len(keyMem SAMem) int64

//export _sa_info_setString
func _sa_info_setString(keyMem SAMem, valueMem SAMem) int64

//export _sa_blob
func _sa_blob(pathMem SAMem, dstMem SAMem) int64

//export _sa_blob_len
func _sa_blob_len(pathMem SAMem) int64

//export _sa_sql_write
func _sa_sql_write(dbUrlMem SAMem, queryMem SAMem) int64

//export _sa_sql_read
func _sa_sql_read(dbUrlMem SAMem, queryMem SAMem) int64

//export _sa_sql_readRowCount
func _sa_sql_readRowCount(dbUrlMem SAMem, queryMem SAMem, queryHash int64) int64

//export _sa_sql_readRowLen
func _sa_sql_readRowLen(dbUrlMem SAMem, queryMem SAMem, queryHash int64, row_i uint64) int64

//export _sa_sql_readRow
func _sa_sql_readRow(dbUrlMem SAMem, queryMem SAMem, queryHash int64, row_i uint64, resultMem SAMem) int64

//export _sa_div_colResize
func _sa_div_colResize(pos uint64, nameMem SAMem, val float64) float64

//export _sa_div_rowResize
func _sa_div_rowResize(pos uint64, nameMem SAMem, val float64) float64

//export _sa_div_colMax
func _sa_div_colMax(pos uint64, val float64) float64

//export _sa_div_rowMax
func _sa_div_rowMax(pos uint64, val float64) float64

//export _sa_div_col
func _sa_div_col(pos uint64, val float64) float64

//export _sa_div_row
func _sa_div_row(pos uint64, val float64) float64

//export _sa_div_start
func _sa_div_start(x, y, w, h uint64, nameMem SAMem) int64

//export _sa_div_end
func _sa_div_end()

//export _sa_div_dialogOpen
func _sa_div_dialogOpen(nameMem SAMem, tp uint64) int64

//export _sa_div_dialogClose
func _sa_div_dialogClose()

//export _sa_div_dialogStart
func _sa_div_dialogStart(nameMem SAMem) int64

//export _sa_div_dialogEnd
func _sa_div_dialogEnd()

//export _sa_div_get_info
func _sa_div_get_info(idMem SAMem, x int64, y int64) float64

//export _sa_div_set_info
func _sa_div_set_info(idMem SAMem, val float64, x int64, y int64) float64

//export _sa_paint_rect
func _sa_paint_rect(x, y, w, h float64, margin float64, r, g, b, a uint32, borderWidth float64) int64

//export _sa_paint_line
func _sa_paint_line(x, y, w, h float64, margin float64, sx, sy, ex, ey float64, r, g, b, a uint32, width float64) int64

//export _sa_paint_circle
func _sa_paint_circle(x, y, w, h float64, margin float64, sx, sy, rad float64, r, g, b, a uint32, borderWidth float64) int64

//export _sa_paint_file
func _sa_paint_file(x, y, w, h float64, fileMem SAMem, tooltipMem SAMem, margin, marginX, marginY float64, r, g, b, a uint32, alignV, alignH uint32, fill uint32) int64

//export _sa_paint_text
func _sa_paint_text(x, y, w, h float64,
	valueMem SAMem,
	margin float64, marginX float64, marginY float64,
	r, g, b, a uint32,
	ratioH, lineHeight float64,
	fontId, align, alignV uint32,
	selection, edit, tabIsChar, enable uint32) int64

//export _sa_paint_textWidth
func _sa_paint_textWidth(valueMem SAMem, fontPathMem SAMem, ratioH float64, cursorPos int64) float64

//export _sa_paint_tooltip
func _sa_paint_tooltip(x, y, w, h float64, valueMem SAMem) int64

//export _sa_paint_cursor
func _sa_paint_cursor(nameMem SAMem) int64

//export _sa_print
func _sa_print(mem SAMem)

//export _sa_print_float
func _sa_print_float(val float64)

//export _sa_fn_call
func _sa_fn_call(assetMem SAMem, fnMem SAMem, argsMem SAMem) int64

//export _sa_fn_setReturn
func _sa_fn_setReturn(argsMem SAMem) int64

//export _sa_fn_getReturn
func _sa_fn_getReturn(argsMem SAMem) int64

//export _sa_comp_drawButton
func _sa_comp_drawButton(style uint32, valueMem SAMem, iconMem SAMem, icon_margin float64, urlMem SAMem, tooltipMem SAMem, enable uint32, outMem SAMem) int64

//export _sa_comp_drawSlider
func _sa_comp_drawSlider(styleTrackId uint32, styleThumbId uint32, value float64, min float64, max float64, jump float64, tooltipMem SAMem, enable uint32, outMem SAMem) float64

//export _sa_comp_drawProgress
func _sa_comp_drawProgress(styleFrameId uint32, styleStatusId uint32, value float64, prec int32, tooltipMem SAMem, enable uint32) int64

//export _sa_comp_drawText
func _sa_comp_drawText(style uint32, valueMem SAMem, tooltipMem SAMem, enable uint32, selection uint32) int64

//export _sa_comp_getEditValue
func _sa_comp_getEditValue(outMem SAMem) int64

//export _sa_comp_drawEdit
func _sa_comp_drawEdit(style uint32, valueMem SAMem, valueOrigMem SAMem, tooltipMem SAMem, ghostMem SAMem, enable uint32, outMem SAMem) int64

//export _sa_comp_drawCombo
func _sa_comp_drawCombo(styleId uint32, styleMenuId uint32, value uint64, optionsMem SAMem, tooltipMem SAMem, enable uint32) int64

//export _sa_comp_drawCheckbox
func _sa_comp_drawCheckbox(styleCheckId uint32, styleLabelId uint32, value uint64, labelMem SAMem, tooltipMem SAMem, enable uint32) int64

//export _sa_register_style
func _sa_register_style(jsMem SAMem) int64

//export _sa_div_drag
func _sa_div_drag(groupNameMem SAMem, id uint64) int64

//export _sa_div_drop
func _sa_div_drop(groupNameMem SAMem, vertical uint32, horizontal uint32, inside uint32, outMem SAMem) int64

//export _sa_render_app
func _sa_render_app(appNameMem SAMem, dbUrlMem SAMem, sts_id uint64) int64

type SAMem struct {
	v uint64
}

func _SA_ptrToBytes(mem SAMem) []byte {
	ptr := uint32(mem.v >> 32)
	size := uint32(mem.v)
	return unsafe.Slice((*byte)(unsafe.Pointer(uintptr(ptr))), size)
}

func _SA_stringToPtr(s string) SAMem {
	if len(s) > 0 {
		ptr := unsafe.Pointer(unsafe.StringData(s))
		return SAMem{v: (uint64(uintptr(ptr)) << uint64(32)) | uint64(len(s))}
	}
	return SAMem{v: 0}
}
func _SA_bytesToPtr(s []byte) SAMem {
	if len(s) > 0 {
		ptr := unsafe.Pointer(unsafe.SliceData(s))
		return SAMem{v: (uint64(uintptr(ptr)) << uint64(32)) | uint64(len(s))}
	}
	return SAMem{v: 0}
}

func _SA_ptrToString(mem SAMem) string {
	ptr := uint32(mem.v >> 32)
	size := uint32(mem.v)
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
