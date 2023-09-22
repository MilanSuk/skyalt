/*
Copyright 2023 Milan Suk

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this db except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by assetlicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
)

type AssetDebug struct {
	conn net.Conn

	sts_id int
	asset  string
}

func NewAssetDebug(conn net.Conn) *AssetDebug {
	var as AssetDebug
	as.conn = conn

	as.sts_id = int(as.ReadUint64())
	as.asset = string(as.ReadBytes())

	return &as
}

func (ad *AssetDebug) Destroy() {
	if ad.conn != nil {
		ad.conn.Close()
	}
}

func (ad *AssetDebug) Is(sts_id int, assetName string) bool {
	return ad.sts_id == sts_id && ad.asset == assetName
}

func (ad *AssetDebug) _connectionRead(data []byte) error {
	p := 0
	for p < len(data) {
		n, err := ad.conn.Read(data[p:])
		if err != nil {
			return fmt.Errorf("ad.conn.Read() failed: %w", err)
		}

		p += n
	}
	return nil
}

func (ad *AssetDebug) WriteUint64(v uint64) {
	if ad.conn == nil {
		return
	}

	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], v)

	_, err := ad.conn.Write(b[:])
	if err != nil {
		ad.conn = nil
	}
}

func (ad *AssetDebug) ReadUint64() uint64 {
	if ad.conn == nil {
		return 0
	}

	var b [8]byte
	err := ad._connectionRead(b[:])
	if err != nil {
		ad.conn = nil
	}
	return binary.LittleEndian.Uint64(b[:])
}

func (ad *AssetDebug) WriteFloat64(v float64) {
	ad.WriteUint64(math.Float64bits(v))
}

func (ad *AssetDebug) ReadFloat64() float64 {
	return math.Float64frombits(ad.ReadUint64())
}

func (ad *AssetDebug) ReadBytes() []byte {
	if ad.conn == nil {
		return nil
	}

	sz := int(ad.ReadUint64())
	data := make([]byte, sz)

	if ad.conn == nil {
		return nil
	}
	err := ad._connectionRead(data)
	if err != nil {
		ad.conn = nil
	}
	return data
}
func (ad *AssetDebug) WriteBytes(data []byte) {
	if ad.conn == nil {
		return
	}
	ad.WriteUint64(uint64(len(data))) //size

	if ad.conn == nil {
		return
	}

	_, err := ad.conn.Write(data) //data
	if err != nil {
		ad.conn = nil
	}
}

func (ad *AssetDebug) SaveData(asset *Asset) {
	ad.Call("_sa_exit", nil, asset)
}

func (ad *AssetDebug) _checkRead(fnTp uint64) {

	ad.WriteUint64(fnTp) //send so other side can check as well

	tp := ad.ReadUint64()
	if tp != fnTp {
		fmt.Printf("Error: Expecting(%d), but it's %d\n", fnTp, tp)
	}
}

func (ad *AssetDebug) Call(fnName string, args []byte, asset *Asset) (int64, error) {

	if ad.conn == nil {
		return -1, fmt.Errorf("no connection")
	}

	//function name
	ad.WriteBytes([]byte(fnName))

	//arguments
	ad.WriteBytes(args)

	for ad.conn != nil {
		//recv
		fnTp := ad.ReadUint64()

		switch fnTp {
		case 0:
			json := ad.ReadBytes()
			ret, err := asset.storage_write(json)
			asset.AddLogErr(err)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 1:
			key := ad.ReadBytes()
			ret := asset.info_float(string(key))
			ad.WriteFloat64(ret)
			ad._checkRead(fnTp)

		case 2:
			key := ad.ReadBytes()
			value := ad.ReadFloat64()
			ret := asset.info_setFloat(string(key), value)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 3:
			key := ad.ReadBytes()
			dst, ret := asset.info_string(string(key))
			ad.WriteBytes([]byte(dst))
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 4:
			key := ad.ReadBytes()
			ret := asset.info_string_len(string(key))
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 5:
			key := ad.ReadBytes()
			value := ad.ReadBytes()
			ret := asset.info_setString(string(key), string(value))
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 6:
			path := ad.ReadBytes()
			dst, ret, err := asset.resource(string(path))
			asset.AddLogErr(err)
			ad.WriteBytes([]byte(dst))
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 7:
			path := ad.ReadBytes()
			ret, err := asset.resource_len(string(path))
			asset.AddLogErr(err)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 8:
			name := string(ad.ReadBytes())
			asset.print(name)
			ad._checkRead(fnTp)

		case 9:
			val := ad.ReadFloat64()
			asset._sa_print_float(val)
			ad._checkRead(fnTp)

		case 10:
			db := ad.ReadBytes()
			query := ad.ReadBytes()
			ret, err := asset.sql_write(string(db), string(query))
			asset.AddLogErr(err)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 11:
			db := ad.ReadBytes()
			query := ad.ReadBytes()
			ret, err := asset.sql_read(string(db), string(query))
			asset.AddLogErr(err)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 12:
			db := ad.ReadBytes()
			query := ad.ReadBytes()
			queryHash := int64(ad.ReadUint64())
			ret, err := asset.sql_readRowCount(string(db), string(query), queryHash)
			asset.AddLogErr(err)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 13:
			db := ad.ReadBytes()
			query := ad.ReadBytes()
			queryHash := int64(ad.ReadUint64())
			row_i := ad.ReadUint64()
			ret, err := asset.sql_readRowLen(string(db), string(query), queryHash, row_i)
			asset.AddLogErr(err)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 14:
			db := ad.ReadBytes()
			query := ad.ReadBytes()
			queryHash := int64(ad.ReadUint64())
			row_i := ad.ReadUint64()
			dst, ret, err := asset.sql_readRow(string(db), string(query), queryHash, row_i)
			asset.AddLogErr(err)
			ad.WriteBytes([]byte(dst))
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 20:
			pos := ad.ReadUint64()
			name := string(ad.ReadBytes())
			val := ad.ReadFloat64()
			ret := asset.div_colResize(pos, name, val)
			ad.WriteFloat64(ret)
			ad._checkRead(fnTp)

		case 21:
			pos := ad.ReadUint64()
			name := string(ad.ReadBytes())
			val := ad.ReadFloat64()
			ret := asset.div_rowResize(pos, name, val)
			ad.WriteFloat64(ret)
			ad._checkRead(fnTp)

		case 22:
			pos := ad.ReadUint64()
			val := ad.ReadFloat64()
			ret := asset._sa_div_colMax(pos, val)
			ad.WriteFloat64(ret)
			ad._checkRead(fnTp)

		case 23:
			pos := ad.ReadUint64()
			val := ad.ReadFloat64()
			ret := asset._sa_div_rowMax(pos, val)
			ad.WriteFloat64(ret)
			ad._checkRead(fnTp)

		case 24:
			pos := ad.ReadUint64()
			val := ad.ReadFloat64()
			ret := asset._sa_div_col(pos, val)
			ad.WriteFloat64(ret)
			ad._checkRead(fnTp)

		case 25:
			pos := ad.ReadUint64()
			val := ad.ReadFloat64()
			ret := asset._sa_div_row(pos, val)
			ad.WriteFloat64(ret)
			ad._checkRead(fnTp)

		case 26:
			x := ad.ReadUint64()
			y := ad.ReadUint64()
			w := ad.ReadUint64()
			h := ad.ReadUint64()
			name := string(ad.ReadBytes())
			ret := asset.div_start(x, y, w, h, name)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 27:
			asset._sa_div_end()
			ad._checkRead(fnTp)

		case 28:
			name := string(ad.ReadBytes())
			x := int64(ad.ReadUint64())
			y := int64(ad.ReadUint64())
			ret := asset.div_get_info(name, x, y)
			ad.WriteFloat64(ret)
			ad._checkRead(fnTp)

		case 29:
			name := string(ad.ReadBytes())
			val := ad.ReadFloat64()
			x := int64(ad.ReadUint64())
			y := int64(ad.ReadUint64())
			ret := asset.div_set_info(name, val, x, y)
			ad.WriteFloat64(ret)
			ad._checkRead(fnTp)

		case 40:
			name := string(ad.ReadBytes())
			tp := ad.ReadUint64()
			ret := asset.div_dialogOpen(name, tp)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 41:
			asset._sa_div_dialogClose()
			ad._checkRead(fnTp)

		case 42:
			name := string(ad.ReadBytes())

			ret := asset.div_dialogStart(name)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 43:
			asset._sa_div_dialogEnd()
			ad._checkRead(fnTp)

		case 50:
			x := ad.ReadFloat64()
			y := ad.ReadFloat64()
			w := ad.ReadFloat64()
			h := ad.ReadFloat64()
			margin := ad.ReadFloat64()
			r := uint32(ad.ReadUint64())
			g := uint32(ad.ReadUint64())
			b := uint32(ad.ReadUint64())
			a := uint32(ad.ReadUint64())
			borderWidth := ad.ReadFloat64()
			ret := asset._sa_paint_rect(x, y, w, h, margin, r, g, b, a, borderWidth)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 51:
			x := ad.ReadFloat64()
			y := ad.ReadFloat64()
			w := ad.ReadFloat64()
			h := ad.ReadFloat64()
			margin := ad.ReadFloat64()

			sx := ad.ReadFloat64()
			sy := ad.ReadFloat64()
			ex := ad.ReadFloat64()
			ey := ad.ReadFloat64()
			r := uint32(ad.ReadUint64())
			g := uint32(ad.ReadUint64())
			b := uint32(ad.ReadUint64())
			a := uint32(ad.ReadUint64())
			width := ad.ReadFloat64()
			ret := asset._sa_paint_line(x, y, w, h, margin, sx, sy, ex, ey, r, g, b, a, width)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 52:
			x := ad.ReadFloat64()
			y := ad.ReadFloat64()
			w := ad.ReadFloat64()
			h := ad.ReadFloat64()
			margin := ad.ReadFloat64()
			sx := ad.ReadFloat64()
			sy := ad.ReadFloat64()
			rad := ad.ReadFloat64()
			r := uint32(ad.ReadUint64())
			g := uint32(ad.ReadUint64())
			b := uint32(ad.ReadUint64())
			a := uint32(ad.ReadUint64())
			borderWidth := ad.ReadFloat64()
			ret := asset._sa_paint_circle(x, y, w, h, margin, sx, sy, rad, r, g, b, a, borderWidth)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 53:
			x := ad.ReadFloat64()
			y := ad.ReadFloat64()
			w := ad.ReadFloat64()
			h := ad.ReadFloat64()
			file := string(ad.ReadBytes())
			title := string(ad.ReadBytes())
			margin := ad.ReadFloat64()
			marginX := ad.ReadFloat64()
			marginY := ad.ReadFloat64()
			r := uint32(ad.ReadUint64())
			g := uint32(ad.ReadUint64())
			b := uint32(ad.ReadUint64())
			a := uint32(ad.ReadUint64())
			alignV := uint32(ad.ReadUint64())
			alignH := uint32(ad.ReadUint64())
			fill := uint32(ad.ReadUint64())
			ret := asset.paint_file(x, y, w, h, file, title, margin, marginX, marginY, r, g, b, a, alignV, alignH, fill)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 54:
			x := ad.ReadFloat64()
			y := ad.ReadFloat64()
			w := ad.ReadFloat64()
			h := ad.ReadFloat64()
			styleId := uint32(ad.ReadUint64())
			value := string(ad.ReadBytes())
			selection := uint32(ad.ReadUint64())
			edit := uint32(ad.ReadUint64())
			enable := uint32(ad.ReadUint64())

			style := asset.styles.Get(styleId)
			ret := asset.paint_text(x, y, w, h, style, value, value, selection > 0, edit > 0, enable > 0)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 55:
			value := string(ad.ReadBytes())
			font := string(ad.ReadBytes())
			ratioH := ad.ReadFloat64()
			cursorPos := int64(ad.ReadUint64())
			ret := asset.paint_textWidth(value, font, ratioH, cursorPos)
			ad.WriteFloat64(ret)
			ad._checkRead(fnTp)

		case 56:
			x := ad.ReadFloat64()
			y := ad.ReadFloat64()
			w := ad.ReadFloat64()
			h := ad.ReadFloat64()
			value := string(ad.ReadBytes())
			ret := asset.paint_title(x, y, w, h, value)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 57:
			name := string(ad.ReadBytes())
			ret, err := asset.paint_cursor(name)
			asset.AddLogErr(err)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 70:
			assetName := string(ad.ReadBytes())
			fnName := string(ad.ReadBytes())
			args := ad.ReadBytes()

			v, err := asset.fn_call(assetName, fnName, args)
			asset.AddLogErr(err)
			ad.WriteUint64(uint64(v))
			ad._checkRead(fnTp)

		case 71:
			args := ad.ReadBytes()
			v := asset.fn_setReturn(args)
			ad.WriteUint64(uint64(v))
			ad._checkRead(fnTp)

		case 72:
			dst := asset.fn_getReturn()
			ad.WriteBytes(dst)
			ad.WriteUint64(1)
			ad._checkRead(fnTp)

		case 80:
			styleId := uint32(ad.ReadUint64())

			value := string(ad.ReadBytes())
			icon := string(ad.ReadBytes())
			icon_margin := ad.ReadFloat64()
			url := string(ad.ReadBytes())
			title := string(ad.ReadBytes())
			enable := uint32(ad.ReadUint64())

			style := asset.styles.Get(styleId)
			click, rclick, ret := asset.swp_drawButton(style, value, icon, icon_margin, url, title, enable > 0)

			var dst [2 * 8]byte
			binary.LittleEndian.PutUint64(dst[0:], uint64(OsTrn(click, 1, 0)))
			binary.LittleEndian.PutUint64(dst[8:], uint64(OsTrn(rclick, 1, 0)))
			ad.WriteBytes(dst[:])
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 81:
			styleTrackId := uint32(ad.ReadUint64())
			styleThumbId := uint32(ad.ReadUint64())
			value := ad.ReadFloat64()
			min := ad.ReadFloat64()
			max := ad.ReadFloat64()
			jump := ad.ReadFloat64()
			title := string(ad.ReadBytes())
			enable := uint32(ad.ReadUint64())

			styleTrack := asset.styles.Get(styleTrackId)
			styleThumb := asset.styles.Get(styleThumbId)
			value, active, changed, finished := asset.swp_drawSlider(styleTrack, styleThumb, value, min, max, jump, title, enable > 0)

			var dst [3 * 8]byte
			binary.LittleEndian.PutUint64(dst[0:], uint64(OsTrn(active, 1, 0)))    //active
			binary.LittleEndian.PutUint64(dst[8:], uint64(OsTrn(changed, 1, 0)))   //changed
			binary.LittleEndian.PutUint64(dst[16:], uint64(OsTrn(finished, 1, 0))) //finished
			ad.WriteBytes(dst[:])
			ad.WriteFloat64(value)
			ad._checkRead(fnTp)

		case 82:
			styleFrameId := uint32(ad.ReadUint64())
			styleStatusId := uint32(ad.ReadUint64())

			value := ad.ReadFloat64()
			prec := int32(ad.ReadUint64())
			title := string(ad.ReadBytes())
			enable := uint32(ad.ReadUint64())

			styleFrame := asset.styles.Get(styleFrameId)
			styleStatus := asset.styles.Get(styleStatusId)
			ret := asset.swp_drawProgress(styleFrame, styleStatus, value, int(prec), title, enable > 0)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 83:
			styleId := uint32(ad.ReadUint64())
			value := string(ad.ReadBytes())
			title := string(ad.ReadBytes())
			enable := uint32(ad.ReadUint64()) > 0
			selection := uint32(ad.ReadUint64()) > 0

			style := asset.styles.Get(styleId)
			ret := asset.swp_drawText(style, value, title, enable, selection)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 84:
			edit := asset.swp_getEditValue()
			ad.WriteBytes([]byte(edit))
			ad.WriteUint64(1)
			ad._checkRead(fnTp)

		case 85:
			styleId := uint32(ad.ReadUint64())
			value := string(ad.ReadBytes())
			valueOrig := string(ad.ReadBytes())
			title := string(ad.ReadBytes())
			ghost := string(ad.ReadBytes())
			enable := uint32(ad.ReadUint64()) > 0

			style := asset.styles.Get(styleId)
			last_edit, active, changed, finished := asset.swp_drawEdit(style, value, valueOrig, title, ghost, enable)

			var dst [4 * 8]byte
			binary.LittleEndian.PutUint64(dst[0:], uint64(OsTrn(active, 1, 0)))    //active
			binary.LittleEndian.PutUint64(dst[8:], uint64(OsTrn(changed, 1, 0)))   //changed
			binary.LittleEndian.PutUint64(dst[16:], uint64(OsTrn(finished, 1, 0))) //finished
			binary.LittleEndian.PutUint64(dst[24:], uint64(len(last_edit)))        //size
			ad.WriteBytes(dst[:])
			ad.WriteUint64(1)
			ad._checkRead(fnTp)

		case 86:
			styleId := uint32(ad.ReadUint64())
			styleMenuId := uint32(ad.ReadUint64())
			value := ad.ReadUint64()
			options := string(ad.ReadBytes())
			title := string(ad.ReadBytes())
			enable := uint32(ad.ReadUint64())

			style := asset.styles.Get(styleId)
			styleMenu := asset.styles.Get(styleMenuId)

			valueOut := asset.swp_drawCombo(style, styleMenu, value, options, title, enable > 0)
			ad.WriteUint64(uint64(valueOut))
			ad._checkRead(fnTp)

		case 87:
			cd_r := uint32(ad.ReadUint64())
			cd_g := uint32(ad.ReadUint64())
			cd_b := uint32(ad.ReadUint64())
			cd_a := uint32(ad.ReadUint64())

			value := ad.ReadUint64()
			description := string(ad.ReadBytes())
			title := string(ad.ReadBytes())

			height := ad.ReadFloat64()
			align := uint32(ad.ReadUint64())
			alignV := uint32(ad.ReadUint64())
			enable := ad.ReadUint64() != 0

			valueOut := asset.swp_drawCheckbox(cd_r, cd_g, cd_b, cd_a,
				value, description, title,
				height, align, alignV, enable)
			ad.WriteUint64(uint64(valueOut))
			ad._checkRead(fnTp)

		case 100:
			js := ad.ReadBytes()
			ret := asset.register_style(js)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 110:
			groupName := string(ad.ReadBytes())
			id := ad.ReadUint64()
			ret := asset.div_drag(groupName, id)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 111:
			groupName := string(ad.ReadBytes())
			vertical := uint32(ad.ReadUint64())
			horizontal := uint32(ad.ReadUint64())
			inside := uint32(ad.ReadUint64())

			id, pos, done := asset.div_drop(groupName, vertical, horizontal, inside)

			var dst [2 * 8]byte
			binary.LittleEndian.PutUint64(dst[0:], uint64(id))
			binary.LittleEndian.PutUint64(dst[8:], uint64(pos))
			ad.WriteBytes(dst[:])
			ad.WriteUint64(uint64(done))
			ad._checkRead(fnTp)

		case 120:
			app := string(ad.ReadBytes())
			db := string(ad.ReadBytes())
			sts_id := ad.ReadUint64()

			ret, err := asset.render_app(app, db, sts_id)
			asset.AddLogErr(err)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 130:
			line := string(ad.ReadBytes())
			asset.SetDebugLine(line)
			ad._checkRead(fnTp)

		case 1000:
			//must return len(returnBytes)
			return 0, nil //render() is done

		default:
			return -1, fmt.Errorf("unknown type: %d", fnTp)
		}
	}

	//connection closed
	return -1, nil
}
