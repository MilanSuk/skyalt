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
	"encoding/binary"
	"fmt"
	"math"
	"net"
)

type AppDebug struct {
	conn net.Conn
	name string
}

func NewAppDebug(conn net.Conn) *AppDebug {
	var as AppDebug
	as.conn = conn
	as.name = string(as.ReadBytes())
	return &as
}

func (ad *AppDebug) Destroy() {
	if ad.conn != nil {
		ad.conn.Close()
	}
}

func (ad *AppDebug) _connectionRead(data []byte) error {
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

func (ad *AppDebug) WriteUint64(v uint64) {
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

func (ad *AppDebug) ReadUint64() uint64 {
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

func (ad *AppDebug) WriteFloat64(v float64) {
	ad.WriteUint64(math.Float64bits(v))
}

func (ad *AppDebug) ReadFloat64() float64 {
	return math.Float64frombits(ad.ReadUint64())
}

func (ad *AppDebug) ReadBytes() []byte {
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
func (ad *AppDebug) WriteBytes(data []byte) {
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

func (ad *AppDebug) SaveData(app *App) {
	ad.Call("_sa_save", app)
}

func (ad *AppDebug) _checkRead(fnTp uint64) {

	ad.WriteUint64(fnTp) //send so other side can check as well

	tp := ad.ReadUint64()
	if tp != fnTp && ad.conn != nil {
		fmt.Printf("Error: Expecting(%d), but it's %d\n", fnTp, tp)
	}
}

func (ad *AppDebug) Call(fnName string, app *App) (int64, error) {

	if ad.conn == nil {
		return -1, fmt.Errorf("no connection")
	}

	//function name
	ad.WriteBytes([]byte(fnName))

	//arguments
	//ad.WriteBytes(nil)

	for ad.conn != nil {
		//recv
		fnTp := ad.ReadUint64()

		switch fnTp {
		case 0:
			json := ad.ReadBytes()
			ret, err := app.storage_write(json)
			app.AddLogErr(err)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 1:
			cmd := ad.ReadBytes()
			prm1 := ad.ReadBytes()
			prm2 := ad.ReadBytes()
			ret := app.info_get_prepare(string(cmd), string(prm1), string(prm2))
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 2:
			ad.WriteBytes([]byte(app.info_string))
			ad.WriteUint64(uint64(1))
			ad._checkRead(fnTp)

		case 3:
			cmd := ad.ReadBytes()
			prm1 := ad.ReadBytes()
			prm2 := ad.ReadBytes()
			prm3 := ad.ReadBytes()
			ret := app.info_set(string(cmd), string(prm1), string(prm2), string(prm3))
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 6:
			path := ad.ReadBytes()
			dst, ret, err := app.resource(string(path))
			app.AddLogErr(err)
			ad.WriteBytes([]byte(dst))
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 7:
			path := ad.ReadBytes()
			ret, err := app.resource_len(string(path))
			app.AddLogErr(err)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 8:
			name := string(ad.ReadBytes())
			app.print(name)
			ad._checkRead(fnTp)

		case 9:
			val := ad.ReadFloat64()
			app._sa_print_float(val)
			ad._checkRead(fnTp)

		case 10:
			db := ad.ReadBytes()
			ret, err := app.sql_commit(string(db))
			app.AddLogErr(err)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 11:
			db := ad.ReadBytes()
			ret, err := app.sql_rollback(string(db))
			app.AddLogErr(err)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 12:
			db := ad.ReadBytes()
			query := ad.ReadBytes()
			ret, err := app.sql_write(string(db), string(query))
			app.AddLogErr(err)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 13:
			db := ad.ReadBytes()
			query := ad.ReadBytes()
			ret, err := app.sql_read(string(db), string(query))
			app.AddLogErr(err)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 14:
			db := ad.ReadBytes()
			query := ad.ReadBytes()
			queryHash := int64(ad.ReadUint64())
			ret, err := app.sql_readRowCount(string(db), string(query), queryHash)
			app.AddLogErr(err)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 15:
			db := ad.ReadBytes()
			query := ad.ReadBytes()
			queryHash := int64(ad.ReadUint64())
			row_i := ad.ReadUint64()
			ret, err := app.sql_readRowLen(string(db), string(query), queryHash, row_i)
			app.AddLogErr(err)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 16:
			db := ad.ReadBytes()
			query := ad.ReadBytes()
			queryHash := int64(ad.ReadUint64())
			row_i := ad.ReadUint64()
			dst, ret, err := app.sql_readRow(string(db), string(query), queryHash, row_i)
			app.AddLogErr(err)
			ad.WriteBytes([]byte(dst))
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 20:
			pos := ad.ReadUint64()
			name := string(ad.ReadBytes())
			val := ad.ReadFloat64()
			ret := app.div_colResize(pos, name, val)
			ad.WriteFloat64(ret)
			ad._checkRead(fnTp)

		case 21:
			pos := ad.ReadUint64()
			name := string(ad.ReadBytes())
			val := ad.ReadFloat64()
			ret := app.div_rowResize(pos, name, val)
			ad.WriteFloat64(ret)
			ad._checkRead(fnTp)

		case 22:
			pos := ad.ReadUint64()
			val := ad.ReadFloat64()
			ret := app._sa_div_colMax(pos, val)
			ad.WriteFloat64(ret)
			ad._checkRead(fnTp)

		case 23:
			pos := ad.ReadUint64()
			val := ad.ReadFloat64()
			ret := app._sa_div_rowMax(pos, val)
			ad.WriteFloat64(ret)
			ad._checkRead(fnTp)

		case 24:
			pos := ad.ReadUint64()
			val := ad.ReadFloat64()
			ret := app._sa_div_col(pos, val)
			ad.WriteFloat64(ret)
			ad._checkRead(fnTp)

		case 25:
			pos := ad.ReadUint64()
			val := ad.ReadFloat64()
			ret := app._sa_div_row(pos, val)
			ad.WriteFloat64(ret)
			ad._checkRead(fnTp)

		case 26:
			x := ad.ReadUint64()
			y := ad.ReadUint64()
			w := ad.ReadUint64()
			h := ad.ReadUint64()
			name := string(ad.ReadBytes())
			ret := app.div_start(x, y, w, h, name)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 27:
			x := ad.ReadUint64()
			y := ad.ReadUint64()
			w := ad.ReadUint64()
			h := ad.ReadUint64()
			rx := ad.ReadFloat64()
			ry := ad.ReadFloat64()
			rw := ad.ReadFloat64()
			rh := ad.ReadFloat64()
			name := string(ad.ReadBytes())
			ret := app.div_startEx(x, y, w, h, rx, ry, rw, rh, name)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 28:
			app._sa_div_end()
			ad._checkRead(fnTp)

		case 29:
			name := string(ad.ReadBytes())
			x := int64(ad.ReadUint64())
			y := int64(ad.ReadUint64())
			ret := app.div_get_info(name, x, y)
			ad.WriteFloat64(ret)
			ad._checkRead(fnTp)

		case 30:
			name := string(ad.ReadBytes())
			val := ad.ReadFloat64()
			x := int64(ad.ReadUint64())
			y := int64(ad.ReadUint64())
			ret := app.div_set_info(name, val, x, y)
			ad.WriteFloat64(ret)
			ad._checkRead(fnTp)

		case 40:
			name := string(ad.ReadBytes())
			tp := ad.ReadUint64()
			ret := app.div_dialogOpen(name, tp)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 41:
			app._sa_div_dialogClose()
			ad._checkRead(fnTp)

		case 42:
			name := string(ad.ReadBytes())

			ret := app.div_dialogStart(name)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 43:
			app._sa_div_dialogEnd()
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
			ret := app._sa_paint_rect(x, y, w, h, margin, r, g, b, a, borderWidth)
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
			ret := app._sa_paint_line(x, y, w, h, margin, sx, sy, ex, ey, r, g, b, a, width)
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
			ret := app._sa_paint_circle(x, y, w, h, margin, sx, sy, rad, r, g, b, a, borderWidth)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 53:
			x := ad.ReadFloat64()
			y := ad.ReadFloat64()
			w := ad.ReadFloat64()
			h := ad.ReadFloat64()
			file := string(ad.ReadBytes())
			tooltip := string(ad.ReadBytes())
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
			ret := app.paint_file(x, y, w, h, file, tooltip, margin, marginX, marginY, r, g, b, a, alignV, alignH, fill)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 54:
			x := ad.ReadFloat64()
			y := ad.ReadFloat64()
			w := ad.ReadFloat64()
			h := ad.ReadFloat64()
			styleId := uint32(ad.ReadUint64())
			value := string(ad.ReadBytes())
			selection := ad.ReadUint64() > 0
			edit := ad.ReadUint64() > 0
			enable := ad.ReadUint64() > 0

			style := app.styles.Get(styleId)
			ret := app.paint_text(x, y, w, h, style, value, value, selection, edit, enable)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 55:
			styleId := uint32(ad.ReadUint64())
			value := string(ad.ReadBytes())
			cursorPos := int64(ad.ReadUint64())

			style := app.styles.Get(styleId)
			ret := app.paint_textWidth(style, value, cursorPos)
			ad.WriteFloat64(ret)
			ad._checkRead(fnTp)

		case 56:
			x := ad.ReadFloat64()
			y := ad.ReadFloat64()
			w := ad.ReadFloat64()
			h := ad.ReadFloat64()
			value := string(ad.ReadBytes())
			ret := app.paint_tooltip(x, y, w, h, value)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 57:
			name := string(ad.ReadBytes())
			ret, err := app.paint_cursor(name)
			app.AddLogErr(err)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 80:
			styleId := uint32(ad.ReadUint64())

			value := string(ad.ReadBytes())
			icon := string(ad.ReadBytes())
			icon_margin := ad.ReadFloat64()
			url := string(ad.ReadBytes())
			tooltip := string(ad.ReadBytes())
			enable := ad.ReadUint64() > 0

			style := app.styles.Get(styleId)
			click, rclick, ret := app.comp_drawButton(style, value, icon, icon_margin, url, tooltip, enable)

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
			tooltip := string(ad.ReadBytes())
			enable := ad.ReadUint64() > 0

			styleTrack := app.styles.Get(styleTrackId)
			styleThumb := app.styles.Get(styleThumbId)
			value, active, changed, finished := app.comp_drawSlider(styleTrack, styleThumb, value, min, max, jump, tooltip, enable)

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
			tooltip := string(ad.ReadBytes())
			enable := ad.ReadUint64() > 0

			styleFrame := app.styles.Get(styleFrameId)
			styleStatus := app.styles.Get(styleStatusId)
			ret := app.comp_drawProgress(styleFrame, styleStatus, value, int(prec), tooltip, enable)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 83:
			styleId := uint32(ad.ReadUint64())
			value := string(ad.ReadBytes())
			tooltip := string(ad.ReadBytes())
			enable := uint32(ad.ReadUint64()) > 0
			selection := uint32(ad.ReadUint64()) > 0

			style := app.styles.Get(styleId)
			ret := app.comp_drawText(style, value, tooltip, enable, selection)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 84:
			edit := app.comp_getEditValue()
			ad.WriteBytes([]byte(edit))
			ad.WriteUint64(1)
			ad._checkRead(fnTp)

		case 85:
			styleId := uint32(ad.ReadUint64())
			value := string(ad.ReadBytes())
			valueOrig := string(ad.ReadBytes())
			tooltip := string(ad.ReadBytes())
			ghost := string(ad.ReadBytes())
			enable := uint32(ad.ReadUint64()) > 0

			style := app.styles.Get(styleId)
			last_edit, active, changed, finished := app.comp_drawEdit(style, value, valueOrig, tooltip, ghost, enable)

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
			tooltip := string(ad.ReadBytes())
			enable := ad.ReadUint64() > 0

			style := app.styles.Get(styleId)
			styleMenu := app.styles.Get(styleMenuId)

			valueOut := app.comp_drawCombo(style, styleMenu, value, options, tooltip, enable)
			ad.WriteUint64(uint64(valueOut))
			ad._checkRead(fnTp)

		case 87:
			styleCheckId := uint32(ad.ReadUint64())
			styleLabelId := uint32(ad.ReadUint64())

			value := ad.ReadUint64()
			label := string(ad.ReadBytes())
			tooltip := string(ad.ReadBytes())
			enable := ad.ReadUint64() != 0

			styleCheck := app.styles.Get(styleCheckId)
			styleLabel := app.styles.Get(styleLabelId)

			valueOut := app.comp_drawCheckbox(styleCheck, styleLabel, value, label, tooltip, enable)
			ad.WriteUint64(uint64(valueOut))
			ad._checkRead(fnTp)

		case 100:
			js := ad.ReadBytes()
			ret := app.register_style(js)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 110:
			groupName := string(ad.ReadBytes())
			id := ad.ReadUint64()
			ret := app.div_drag(groupName, id)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 111:
			groupName := string(ad.ReadBytes())
			vertical := uint32(ad.ReadUint64())
			horizontal := uint32(ad.ReadUint64())
			inside := uint32(ad.ReadUint64())

			id, pos, done := app.div_drop(groupName, vertical, horizontal, inside)

			var dst [2 * 8]byte
			binary.LittleEndian.PutUint64(dst[0:], uint64(id))
			binary.LittleEndian.PutUint64(dst[8:], uint64(pos))
			ad.WriteBytes(dst[:])
			ad.WriteUint64(uint64(done))
			ad._checkRead(fnTp)

		case 120:
			db := string(ad.ReadBytes())
			app_rowid := ad.ReadUint64()

			ret, err := app.render_app(db, app_rowid)
			app.AddLogErr(err)
			ad.WriteUint64(uint64(ret))
			ad._checkRead(fnTp)

		case 130:
			line := string(ad.ReadBytes())
			app.SetDebugLine(line)
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
