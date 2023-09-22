package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
)

var conn *net.TCPConn

func main() {
	port, sts_id, asset := Debug()

	if port < 0 {
		port = 8091
	}

	tcpServer, err := net.ResolveTCPAddr("tcp", "localhost:"+strconv.Itoa(port))
	if err != nil {
		fmt.Printf("ResolveTCPAddr() failed: %v\n", err)
		return
	}

	conn, err = net.DialTCP("tcp", nil, tcpServer)
	if err != nil {
		fmt.Printf("DialTCP() failed: %v\n", err)
		return
	}

	WriteUint64(uint64(sts_id))
	WriteBytes([]byte(asset))

	fmt.Printf("Connected on port: %d\n", port)

	for {
		fnName, err := ReadBytes()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Panic(err)
		}
		args, err := ReadBytes()
		if err != nil {
			log.Panic(err)
		}

		switch string(fnName) {
		case "_sa_render":
			Render()

		case "_sa_init":
			var jsStore []byte
			_arrayToArgs(args, &jsStore)
			if !Open(jsStore) {
				json.Unmarshal(jsStore, &store)
			}

		case "_sa_exit":
			js, written := Save()
			if !written {
				js, _ = json.MarshalIndent(&store, "", "")
			}
			_sa_storage_write(_SA_bytesToPtr(js))

		case "_sa_styles":
			var jsStyles []byte
			_arrayToArgs(args, &jsStyles)
			styles = SA_Styles{} //reset ids
			json.Unmarshal(jsStyles, &styles)
			Styles()

		case "_sa_translations":
			e := reflect.ValueOf(&trns).Elem()
			for i := 0; i < e.NumField(); i++ {
				e.Field(i).SetString("{" + e.Type().Field(i).Name + "}")
			}

			var js []byte
			_arrayToArgs(args, &js)
			json.Unmarshal(js, &trns)

		default:
			log.Panic("Unknown function: ", string(fnName))
		}

		WriteUint64(1000) //end of function call
	}

	err = conn.Close()
	if err != nil {
		fmt.Printf("Close() failed: %v\n", err)
		return
	}
}

func _connectionRead(data []byte) error {
	p := 0
	for p < len(data) {
		n, err := conn.Read(data[p:])
		if err != nil {
			return err
		}

		p += n
	}
	return nil
}

func WriteUint64(v uint64) {
	var b [8]byte
	_SA_putUint64(b[:], v)

	_, err := conn.Write(b[:])
	if err != nil {
		log.Panic(err)
	}
}

func ReadUint64() uint64 {
	var b [8]byte
	err := _connectionRead(b[:])
	if err != nil {
		log.Panic(err)
	}
	return _SA_getUint64(b[:])
}

func WriteFloat64(v float64) {
	WriteUint64(math.Float64bits(v))
}

func ReadFloat64() float64 {
	return math.Float64frombits(ReadUint64())
}

func WriteMem(mem SAMem) {
	WriteUint64(uint64(len(mem.v))) //size
	_, err := conn.Write(mem.v)     //data
	if err != nil {
		log.Panic(err)
	}
}

func ReadBytes() ([]byte, error) {

	//size
	//ReadUint64() but with error
	var b [8]byte
	err := _connectionRead(b[:])
	if err != nil {
		return nil, err
	}
	sz := _SA_getUint64(b[:])

	//data
	data := make([]byte, sz)
	_connectionRead(data)
	return data, nil
}
func WriteBytes(data []byte) {
	WriteUint64(uint64(len(data))) //size
	_, err := conn.Write(data)     //data
	if err != nil {
		log.Panic(err)
	}
}

func ReadMem(mem SAMem) {
	sz := int(ReadUint64())
	if sz != len(mem.v) {
		log.Panic("Wrong size")
	}
	_connectionRead(mem.v)
}

//-------

func _checkRead(fnTp uint64) {
	WriteUint64(fnTp) //send so other side can check as well

	tp := ReadUint64()
	if tp != fnTp {
		fmt.Printf("Error: Expecting(%d), but it's %d\n", fnTp, tp)
	}
}

func _sa_storage_write(jsonMem SAMem) int64 {
	WriteUint64(0)
	WriteMem(jsonMem)
	ret := int64(ReadUint64())
	_checkRead(0)
	return ret
}

func _sa_info_float(keyMem SAMem) float64 {
	WriteUint64(1)
	WriteMem(keyMem)
	ret := ReadFloat64()
	_checkRead(1)
	return ret
}

func _sa_info_setFloat(keyMem SAMem, value float64) int64 {
	WriteUint64(2)
	WriteMem(keyMem)
	WriteFloat64(value)
	ret := int64(ReadUint64())
	_checkRead(2)
	return ret
}

func _sa_info_string(keyMem SAMem, dstMem SAMem) int64 {
	WriteUint64(3)
	WriteMem(keyMem)

	ReadMem(dstMem)
	ret := int64(ReadUint64())
	_checkRead(3)
	return ret
}

func _sa_info_string_len(keyMem SAMem) int64 {
	WriteUint64(4)
	WriteMem(keyMem)
	ret := int64(ReadUint64())
	_checkRead(4)
	return ret
}

func _sa_info_setString(keyMem SAMem, valueMem SAMem) int64 {
	WriteUint64(5)
	WriteMem(keyMem)
	WriteMem(valueMem)
	ret := int64(ReadUint64())
	_checkRead(5)
	return ret
}

func _sa_resource(pathMem SAMem, dstMem SAMem) int64 {
	WriteUint64(6)
	WriteMem(pathMem)

	ReadMem(dstMem)
	ret := int64(ReadUint64())
	_checkRead(6)
	return ret
}

func _sa_resource_len(pathMem SAMem) int64 {
	WriteUint64(7)
	WriteMem(pathMem)
	ret := int64(ReadUint64())
	_checkRead(7)
	return ret
}

func _sa_print(mem SAMem) {
	WriteUint64(8)
	WriteMem(mem)
	_checkRead(8)
}

func _sa_print_float(val float64) {
	WriteUint64(9)
	WriteFloat64(val)
	_checkRead(9)
}

//-------

func _sa_sql_write(dbMem SAMem, queryMem SAMem) int64 {
	WriteUint64(10)
	WriteMem(dbMem)
	WriteMem(queryMem)
	ret := int64(ReadUint64())
	_checkRead(10)
	return ret
}

func _sa_sql_read(dbMem SAMem, queryMem SAMem) int64 {
	WriteUint64(11)
	WriteMem(dbMem)
	WriteMem(queryMem)
	ret := int64(ReadUint64())
	_checkRead(11)
	return ret
}

func _sa_sql_readRowCount(dbMem SAMem, queryMem SAMem, queryHash int64) int64 {
	WriteUint64(12)
	WriteMem(dbMem)
	WriteMem(queryMem)
	WriteUint64(uint64(queryHash))
	ret := int64(ReadUint64())
	_checkRead(12)
	return ret
}

func _sa_sql_readRowLen(dbMem SAMem, queryMem SAMem, queryHash int64, row_i uint64) int64 {
	WriteUint64(13)
	WriteMem(dbMem)
	WriteMem(queryMem)
	WriteUint64(uint64(queryHash))
	WriteUint64(row_i)
	ret := int64(ReadUint64())
	_checkRead(13)
	return ret
}

func _sa_sql_readRow(dbMem SAMem, queryMem SAMem, queryHash int64, row_i uint64, resultMem SAMem) int64 {
	WriteUint64(14)
	WriteMem(dbMem)
	WriteMem(queryMem)
	WriteUint64(uint64(queryHash))
	WriteUint64(row_i)

	ReadMem(resultMem)
	ret := int64(ReadUint64())
	_checkRead(14)
	return ret
}

//-------

func _sa_div_colResize(pos uint64, nameMem SAMem, val float64) float64 {
	WriteUint64(20)
	WriteUint64(pos)
	WriteMem(nameMem)
	WriteFloat64(val)

	ret := ReadFloat64()
	_checkRead(20)
	return ret
}
func _sa_div_rowResize(pos uint64, nameMem SAMem, val float64) float64 {
	WriteUint64(21)
	WriteUint64(pos)
	WriteMem(nameMem)
	WriteFloat64(val)

	ret := ReadFloat64()
	_checkRead(21)
	return ret
}
func _sa_div_colMax(pos uint64, val float64) float64 {
	WriteUint64(22)
	WriteUint64(pos)
	WriteFloat64(val)

	ret := ReadFloat64()
	_checkRead(22)
	return ret
}

func _sa_div_rowMax(pos uint64, val float64) float64 {
	WriteUint64(23)
	WriteUint64(pos)
	WriteFloat64(val)

	ret := ReadFloat64()
	_checkRead(23)
	return ret
}

func _sa_div_col(pos uint64, val float64) float64 {
	WriteUint64(24)
	WriteUint64(pos)
	WriteFloat64(val)

	ret := ReadFloat64()
	_checkRead(24)
	return ret
}

func _sa_div_row(pos uint64, val float64) float64 {
	WriteUint64(25)
	WriteUint64(pos)
	WriteFloat64(val)

	ret := ReadFloat64()
	_checkRead(25)
	return ret
}

func _sa_div_start(x, y, w, h uint64, nameMem SAMem) int64 {
	WriteUint64(26)
	WriteUint64(x)
	WriteUint64(y)
	WriteUint64(w)
	WriteUint64(h)
	WriteMem(nameMem)

	ret := int64(ReadUint64())
	_checkRead(26)
	return ret
}

func _sa_div_end() {
	WriteUint64(27)
	_checkRead(27)
}

func _sa_div_get_info(idMem SAMem, x int64, y int64) float64 {
	WriteUint64(28)
	WriteMem(idMem)
	WriteUint64(uint64(x))
	WriteUint64(uint64(y))

	ret := ReadFloat64()
	_checkRead(28)
	return ret
}

func _sa_div_set_info(idMem SAMem, val float64, x int64, y int64) float64 {
	WriteUint64(29)
	WriteMem(idMem)
	WriteFloat64(val)
	WriteUint64(uint64(x))
	WriteUint64(uint64(y))

	ret := ReadFloat64()
	_checkRead(29)
	return ret
}

//-------

func _sa_div_dialogOpen(nameMem SAMem, tp uint64) int64 {
	WriteUint64(40)
	WriteMem(nameMem)
	WriteUint64(tp)

	ret := int64(ReadUint64())
	_checkRead(40)
	return ret
}

func _sa_div_dialogClose() {
	WriteUint64(41)
	_checkRead(41)
}

func _sa_div_dialogStart(nameMem SAMem) int64 {
	WriteUint64(42)
	WriteMem(nameMem)

	ret := int64(ReadUint64())
	_checkRead(42)
	return ret
}

func _sa_div_dialogEnd() {
	WriteUint64(43)
	_checkRead(43)
}

//-------

func _sa_paint_rect(x, y, w, h float64, margin float64, r, g, b, a uint32, borderWidth float64) int64 {
	WriteUint64(50)
	WriteFloat64(x)
	WriteFloat64(y)
	WriteFloat64(w)
	WriteFloat64(h)
	WriteFloat64(margin)
	WriteUint64(uint64(r))
	WriteUint64(uint64(g))
	WriteUint64(uint64(b))
	WriteUint64(uint64(a))
	WriteFloat64(borderWidth)

	ret := int64(ReadUint64())
	_checkRead(50)
	return ret
}

func _sa_paint_line(x, y, w, h float64, margin float64, sx, sy, ex, ey float64, r, g, b, a uint32, width float64) int64 {
	WriteUint64(51)
	WriteFloat64(x)
	WriteFloat64(y)
	WriteFloat64(w)
	WriteFloat64(h)
	WriteFloat64(margin)
	WriteFloat64(sx)
	WriteFloat64(sy)
	WriteFloat64(ex)
	WriteFloat64(ey)
	WriteUint64(uint64(r))
	WriteUint64(uint64(g))
	WriteUint64(uint64(b))
	WriteUint64(uint64(a))
	WriteFloat64(width)

	ret := int64(ReadUint64())
	_checkRead(51)
	return ret
}

func _sa_paint_circle(x, y, w, h float64, margin float64, sx, sy, rad float64, r, g, b, a uint32, borderWidth float64) int64 {
	WriteUint64(52)
	WriteFloat64(x)
	WriteFloat64(y)
	WriteFloat64(w)
	WriteFloat64(h)
	WriteFloat64(margin)
	WriteFloat64(sx)
	WriteFloat64(sy)
	WriteFloat64(rad)
	WriteUint64(uint64(r))
	WriteUint64(uint64(g))
	WriteUint64(uint64(b))
	WriteUint64(uint64(a))
	WriteFloat64(borderWidth)

	ret := int64(ReadUint64())
	_checkRead(52)
	return ret
}

func _sa_paint_file(x, y, w, h float64, fileMem SAMem, titleMem SAMem, margin, marginX, marginY float64, r, g, b, a uint32, alignV, alignH uint32, fill uint32) int64 {
	WriteUint64(53)
	WriteFloat64(x)
	WriteFloat64(y)
	WriteFloat64(w)
	WriteFloat64(h)
	WriteMem(fileMem)
	WriteMem(titleMem)
	WriteFloat64(margin)
	WriteFloat64(marginX)
	WriteFloat64(marginY)
	WriteUint64(uint64(r))
	WriteUint64(uint64(g))
	WriteUint64(uint64(b))
	WriteUint64(uint64(a))
	WriteUint64(uint64(alignV))
	WriteUint64(uint64(alignH))
	WriteUint64(uint64(fill))

	ret := int64(ReadUint64())
	_checkRead(53)
	return ret
}

func _sa_paint_text(x, y, w, h float64,
	valueMem SAMem,
	margin float64, marginX float64, marginY float64,
	r, g, b, a uint32,
	ratioH, lineHeight float64,
	fontId, align, alignV uint32,
	selection, edit, tabIsChar, enable uint32) int64 {

	WriteUint64(54)
	WriteFloat64(x)
	WriteFloat64(y)
	WriteFloat64(w)
	WriteFloat64(h)

	WriteMem(valueMem)

	WriteFloat64(margin)
	WriteFloat64(marginX)
	WriteFloat64(marginY)

	WriteUint64(uint64(r))
	WriteUint64(uint64(g))
	WriteUint64(uint64(b))
	WriteUint64(uint64(a))

	WriteFloat64(ratioH)
	WriteFloat64(lineHeight)

	WriteUint64(uint64(fontId))
	WriteUint64(uint64(align))
	WriteUint64(uint64(alignV))
	WriteUint64(uint64(selection))
	WriteUint64(uint64(edit))
	WriteUint64(uint64(tabIsChar))
	WriteUint64(uint64(enable))

	ret := int64(ReadUint64())
	_checkRead(54)
	return ret
}

func _sa_paint_textWidth(valueMem SAMem, fontPathMem SAMem, ratioH float64, cursorPos int64) float64 {
	WriteUint64(55)
	WriteMem(valueMem)
	WriteMem(fontPathMem)
	WriteFloat64(ratioH)
	WriteUint64(uint64(cursorPos))

	ret := ReadFloat64()
	_checkRead(55)
	return ret
}

func _sa_paint_title(x, y, w, h float64, valueMem SAMem) int64 {
	WriteUint64(56)
	WriteFloat64(x)
	WriteFloat64(y)
	WriteFloat64(w)
	WriteFloat64(h)
	WriteMem(valueMem)

	ret := int64(ReadUint64())
	_checkRead(56)
	return ret
}

func _sa_paint_cursor(nameMem SAMem) int64 {
	WriteUint64(57)
	WriteMem(nameMem)

	ret := int64(ReadUint64())
	_checkRead(57)
	return ret
}

func _sa_fn_call(assetMem SAMem, fnMem SAMem, argsMem SAMem) int64 {
	WriteUint64(70)
	WriteMem(assetMem)
	WriteMem(fnMem)
	WriteMem(argsMem)

	ret := int64(ReadUint64())
	_checkRead(70)
	return ret
}

func _sa_fn_setReturn(argsMem SAMem) int64 {
	WriteUint64(71)
	WriteMem(argsMem)

	ret := int64(ReadUint64())
	_checkRead(71)
	return ret
}

func _sa_fn_getReturn(argsMem SAMem) int64 {
	WriteUint64(72)

	ReadMem(argsMem)
	ret := int64(ReadUint64())
	_checkRead(72)
	return ret
}

func _sa_swp_drawButton(styleId uint32, valueMem SAMem, iconMem SAMem, icon_margin float64, urlMem SAMem, titleMem SAMem, enable uint32, outMem SAMem) int64 {
	WriteUint64(80)

	WriteUint64(uint64(styleId))

	WriteMem(valueMem)
	WriteMem(iconMem)
	WriteFloat64(icon_margin)
	WriteMem(urlMem)
	WriteMem(titleMem)
	WriteUint64(uint64(enable))

	ReadMem(outMem)
	ret := int64(ReadUint64())

	_checkRead(80)
	return ret
}

func _sa_swp_drawSlider(styleTrackId uint32, styleThumbId uint32, value float64, min float64, max float64, jump float64, titleMem SAMem, enable uint32, outMem SAMem) float64 {
	WriteUint64(81)
	WriteUint64(uint64(styleTrackId))
	WriteUint64(uint64(styleThumbId))
	WriteFloat64(value)
	WriteFloat64(min)
	WriteFloat64(max)
	WriteFloat64(jump)
	WriteMem(titleMem)
	WriteUint64(uint64(enable))

	ReadMem(outMem)
	ret := ReadFloat64()
	_checkRead(81)
	return ret
}

func _sa_swp_drawProgress(styleFrameId uint32, styleStatusId uint32, value float64, prec int32, titleMem SAMem, enable uint32) int64 {
	WriteUint64(82)
	WriteUint64(uint64(styleFrameId))
	WriteUint64(uint64(styleStatusId))
	WriteFloat64(value)
	WriteUint64(uint64(prec))
	WriteMem(titleMem)
	WriteUint64(uint64(enable))
	ret := int64(ReadUint64())
	_checkRead(82)
	return ret
}

func _sa_swp_drawText(styleId uint32, valueMem SAMem, titleMem SAMem, enable uint32, selection uint32) int64 {
	WriteUint64(83)
	WriteUint64(uint64(styleId))
	WriteMem(valueMem)
	WriteMem(titleMem)
	WriteUint64(uint64(enable))
	WriteUint64(uint64(selection))

	ret := int64(ReadUint64())
	_checkRead(83)
	return ret
}

func _sa_swp_getEditValue(outMem SAMem) int64 {
	WriteUint64(84)

	ReadMem(outMem)
	ret := int64(ReadUint64())
	_checkRead(84)
	return ret
}

func _sa_swp_drawEdit(styleId uint32, valueMem SAMem, valueOrigMem SAMem, titleMem SAMem, ghostMem SAMem, enable uint32, outMem SAMem) int64 {
	WriteUint64(85)
	WriteUint64(uint64(styleId))
	WriteMem(valueMem)
	WriteMem(valueOrigMem)
	WriteMem(titleMem)
	WriteMem(ghostMem)
	WriteUint64(uint64(enable))

	ReadMem(outMem)
	ret := int64(ReadUint64())
	_checkRead(85)
	return ret
}

func _sa_swp_drawCombo(styleId uint32, styleMenuId uint32, value uint64, optionsMem SAMem, titleMem SAMem, enable uint32) int64 {
	WriteUint64(86)
	WriteUint64(uint64(styleId))
	WriteUint64(uint64(styleMenuId))
	WriteUint64(value)
	WriteMem(optionsMem)
	WriteMem(titleMem)
	WriteUint64(uint64(enable))

	ret := int64(ReadUint64())
	_checkRead(86)
	return ret
}

func _sa_swp_drawCheckbox(cd_r, cd_g, cd_b, cd_a uint32,
	value uint64, descriptionMem SAMem, titleMem SAMem,
	height float64, align uint32, alignV uint32, enable uint32) int64 {
	WriteUint64(87)
	WriteUint64(uint64(cd_r))
	WriteUint64(uint64(cd_g))
	WriteUint64(uint64(cd_b))
	WriteUint64(uint64(cd_a))

	WriteUint64(value)
	WriteMem(descriptionMem)
	WriteMem(titleMem)

	WriteFloat64(height)
	WriteUint64(uint64(align))
	WriteUint64(uint64(alignV))
	WriteUint64(uint64(enable))

	ret := int64(ReadUint64())
	_checkRead(87)
	return ret
}

func _sa_register_style(jsMem SAMem) int64 {
	WriteUint64(100)
	WriteMem(jsMem)
	ret := int64(ReadUint64())
	_checkRead(100)
	return ret
}

func _sa_div_drag(groupName SAMem, id uint64) int64 {
	WriteUint64(110)
	WriteMem(groupName)
	WriteUint64(id)
	ret := int64(ReadUint64())
	_checkRead(110)
	return ret
}

func _sa_div_drop(groupName SAMem, vertical uint32, horizontal uint32, inside uint32, outMem SAMem) int64 {
	WriteUint64(111)
	WriteMem(groupName)
	WriteUint64(uint64(vertical))
	WriteUint64(uint64(horizontal))
	WriteUint64(uint64(inside))

	ReadMem(outMem)
	ret := int64(ReadUint64())
	_checkRead(111)
	return ret
}

func _sa_render_app(appMem SAMem, dbMem SAMem, sts_id uint64) int64 {
	WriteUint64(120)
	WriteMem(appMem)
	WriteMem(dbMem)
	WriteUint64(sts_id)

	ret := int64(ReadUint64())
	_checkRead(120)
	return ret
}

func _sa_debug_line(lineMem SAMem) {
	WriteUint64(130)
	WriteMem(lineMem)
	_checkRead(130)
}

func _SA_DebugLine() {

	ok := true
	i := 2
	for ok {
		var file string
		var line int
		_, file, line, ok = runtime.Caller(i)

		if filepath.Base(file) != "sdk.go" {
			str := file + "/" + strconv.Itoa(line)
			_sa_debug_line(_SA_stringToPtr(str))
			return
		}
		i++
	}

	fmt.Println("Debug caller not found")
}

/*func _SA_ptrToString(mem SAMem) string {
	ptr := uint32(mem >> 32)
	size := uint32(mem)
	return unsafe.String((*byte)(unsafe.Pointer(uintptr(ptr))), size)
}*/

type SAMem struct {
	v []byte
}

func _SA_stringToPtr(s string) SAMem {
	return SAMem{v: []byte(s)}
}

func _SA_bytesToPtr(s []byte) SAMem {
	return SAMem{v: s}
}
func _SA_ptrToBytes(mem SAMem) []byte {
	return mem.v
}
