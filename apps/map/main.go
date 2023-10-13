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
	"encoding/json"
	"fmt"
	"math"
	"strconv"
)

type Storage struct {
	Cam         Cam
	add_locator bool
}

var store Storage

type Translations struct {
	ADD_LOCATOR string
	ZOOM        string
	REMOVE      string
}

var trns Translations

type V2 struct {
	X, Y float64
}

type Cam struct {
	Lon, Lat, Zoom float64

	lonOld, latOld, zoomOld float64
	start_pos               V2
	start_tile              V2
	start_zoom_time         float64
}

func mmax(x, y float64) float64 {
	if x < y {
		return y
	}
	return x
}
func mmin(x, y float64) float64 {
	if x > y {
		return y
	}
	return x
}
func clamp(v, min, max float64) float64 {
	return mmin(mmax(v, min), max)
}

func MetersPerPixel(lat, zoom float64) float64 {
	return 156543.034 * math.Cos(lat/180*math.Pi) / math.Pow(2, zoom)
}

func LonLatToPos(cam Cam) V2 {
	x := (cam.Lon + 180) / 360 * math.Pow(2, cam.Zoom)
	y := (1 - math.Log(math.Tan(cam.Lat*math.Pi/180)+1/math.Cos(cam.Lat*math.Pi/180))/math.Pi) / 2 * math.Pow(2, cam.Zoom)
	return V2{x, y}
}

func PosToLonLat(pos V2, zoom float64) (float64, float64) {
	lon := pos.X/math.Pow(2, zoom)*360 - 180 //long

	n := math.Pi - 2*math.Pi*pos.Y/math.Pow(2, zoom)
	lat := 180 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(n*-1))) //lat
	return lon, lat
}

func CamBbox(res V2, tile float64, cam Cam) (V2, V2, V2) {
	tilePos := LonLatToPos(cam)
	max_res := math.Pow(2, cam.Zoom)

	var start, end, size V2

	start.X = clamp((tilePos.X*tile-res.X/2)/tile, 0, max_res)
	start.Y = clamp((tilePos.Y*tile-res.Y/2)/tile, 0, max_res)
	end.X = clamp((tilePos.X*tile+res.X/2)/tile, 0, max_res)
	end.Y = clamp((tilePos.Y*tile+res.Y/2)/tile, 0, max_res)

	size.X = end.X - start.X
	size.Y = end.Y - start.Y

	return start, end, size
}

func CamCheck(res V2, tile float64, cam *Cam) {
	if res.X <= 0 || res.Y <= 0 {
		return
	}

	bbStart, bbEnd, bbSize := CamBbox(res, tile, *cam)

	maxTiles := math.Pow(2, cam.Zoom)

	def_bbox_size := V2{res.X / tile, res.Y / tile}

	if bbStart.X <= 0 {
		bbSize.X = def_bbox_size.X
		bbStart.X = 0
	}

	if bbStart.Y <= 0 {
		bbSize.Y = def_bbox_size.Y
		bbStart.Y = 0
	}

	if bbEnd.X >= maxTiles {
		bbSize.X = def_bbox_size.X
		bbStart.X = mmax(0, maxTiles-bbSize.X)
	}

	if bbEnd.Y >= maxTiles {
		bbSize.Y = def_bbox_size.Y
		bbStart.Y = mmax(0, maxTiles-bbSize.Y)
	}

	cam.Lon, cam.Lat = PosToLonLat(V2{bbStart.X + bbSize.X/2, bbStart.Y + bbSize.Y/2}, cam.Zoom)
}

// cam
func Measure(cam Cam) {

	metersPerPixels := MetersPerPixel(cam.Lat, cam.Zoom)

	metersPerWidth := metersPerPixels * SA_DivInfo("screenWidth") * SA_DivInfo("cell")
	metersPerStrip := metersPerWidth / 5
	meters := math.Round(metersPerStrip)

	unitText := "m"
	if meters > 1000 {
		meters = math.Round(meters / 1000)
		unitText = "km"
	}

	SA_Row(0, 0.5)
	SA_Row(1, 0.2)
	SA_Row(2, 0.3)

	for i := 0; i < 5; i++ {
		SA_Col(i, 0.1)
		SA_ColMax(i, 100)
	}

	//texts
	SA_TextCenter("###").ValueFloat(meters*0, 0).Show(0, 0, 2, 1)
	SA_TextCenter("###").ValueFloat(meters*1, 0).Show(1, 0, 2, 1)
	SA_TextCenter("###").ValueFloat(meters*2, 0).Show(2, 0, 2, 1)
	//SA_TextCenter("###").ValueFloat(meters*3, 0).Show(3, 0, 2, 1)
	SA_TextCenter("###"+strconv.FormatFloat(meters*3, 'f', 0, 64)+" "+unitText).Show(3, 0, 2, 1)

	//stripes
	SA_DivStart(1, 1, 1, 1)
	SAPaint_Rect(0, 0, 1, 1, 0, SA_ThemeBlack(), 0)
	SA_DivEnd()

	SA_DivStart(2, 1, 1, 1)
	SAPaint_Rect(0, 0, 1, 1, 0, SA_ThemeWhite(), 0)
	SA_DivEnd()

	SA_DivStart(3, 1, 1, 1)
	SAPaint_Rect(0, 0, 1, 1, 0, SA_ThemeBlack(), 0)
	SA_DivEnd()

	//unit
	//SA_Text("###"+unitText).Show(4, 1, 1, 1)

}

func zoomClamp(z float64) float64 {
	return clamp(z, 0, 6) //19
}

func isZooming(cam *Cam) (bool, float64, float64) {
	ANIM_TIME := 0.4
	dt := SA_Time() - cam.start_zoom_time
	return (dt < ANIM_TIME), dt, ANIM_TIME
}

func Map(cam *Cam) {
	zooming := 0

	cam.Zoom = zoomClamp(cam.Zoom) //check

	lon := cam.Lon
	lat := cam.Lat
	zoom := cam.Zoom

	scale := float64(1)
	isZooming, dt, ANIM_TIME := isZooming(cam)
	if isZooming {
		t := dt / ANIM_TIME
		if cam.Zoom > cam.zoomOld {
			scale = 1 + t
		} else {
			scale = 1 - t/2
		}
		zoom = cam.zoomOld
		lon = cam.lonOld + (cam.Lon-cam.lonOld)*t
		lat = cam.latOld + (cam.Lat-cam.latOld)*t
		zooming = 1
		SA_InfoSet("nosleep", "", "", "")
	}

	cell := SA_DivInfo("cell")
	width := SA_DivInfo("screenWidth")
	height := SA_DivInfo("screenHeight")

	touch_x := SA_DivInfo("touchX")
	touch_y := SA_DivInfo("touchY")
	inside := SA_DivInfo("touchInside") > 0
	active := SA_DivInfo("touchActive") > 0
	end := SA_DivInfo("touchEnd") > 0
	start := SA_DivInfo("touchStart") > 0
	wheel := SA_DivInfo("touchWheel")
	clicks := SA_DivInfo("touchClicks")

	coord := V2{width, height}

	tile := 256 / cell * scale
	tileW := tile / width
	tileH := tile / height

	CamCheck(coord, tile, cam)
	bbStart, bbEnd, bbSize := CamBbox(coord, tile, Cam{Lon: lon, Lat: lat, Zoom: zoom})

	//draw tiles
	for y := float64(int(bbStart.Y)); y < bbEnd.Y; y++ {
		for x := float64(int(bbStart.X)); x < bbEnd.X; x++ {
			if x < 0 || y < 0 {
				continue
			}

			tileCoord_sx := (x - bbStart.X) * tileW
			tileCoord_sy := (y - bbStart.Y) * tileH

			q := SA_SqlRead("", "SELECT rowid FROM tiles WHERE name=='"+strconv.Itoa(int(zoom))+"-"+strconv.Itoa(int(x))+"-"+strconv.Itoa(int(y))+".png'")
			var rowid int
			if q.Next(&rowid) {
				file := fmt.Sprintf("dbs::tiles/file/%d", rowid)

				//extra margin will fix white spaces during zooming
				SAPaint_File(tileCoord_sx, tileCoord_sy, tileW, tileH, file, "", float64(zooming)*-0.03, 0, 0, SA_ThemeWhite(), 0, 0, false)
			}

		}
	}

	//Locators
	{
		query := SA_SqlRead("", "SELECT rowid, title, pos FROM locators")
		var rowid int
		var title string
		var pos string
		for query.Next(&rowid, &title, &pos) {
			var ln, lt float64
			_, err := fmt.Sscanf(pos, "%f,%f", &ln, &lt)
			if err != nil {
				continue
			}

			p := LonLatToPos(Cam{Lon: ln, Lat: lt, Zoom: zoom})

			x := (p.X - bbStart.X) * tileW
			y := (p.Y - bbStart.Y) * tileH

			rad := 1.0
			rad_x := rad / width
			rad_y := rad / height

			//SAPaint_Text()	//...
			SAPaint_File(x-rad_x/2, y-rad_y, rad_x, rad_y, "app:resources/locator.png", "", 0, 0, 0, SA_ThemeError(), 1, 0, false)
			//SAPaint_Circle(x, y, 0.1, SA_ThemeError(), 0)

			dnm := fmt.Sprintf("locator_%d", rowid)
			if touch_x > x-rad_x/2 && touch_x < x+rad_x/2 &&
				touch_y > y-rad_y && touch_y < y {

				if end {
					SA_DialogOpen(dnm, 2)
					end = false
				}
				SAPaint_Cursor("hand")
			}

			//bug: bottom editbox is active ...
			if SA_DialogStart(dnm) {
				SA_ColMax(0, 5)

				if SA_Editbox(&title).Show(0, 0, 1, 1).finished {
					SA_SqlWrite("", fmt.Sprintf("UPDATE locators SET title='%s' WHERE rowid=%d;", title, rowid))
				}
				SA_Text(fmt.Sprintf("Lon: %.3f", ln)).Show(0, 1, 1, 1)
				SA_Text(fmt.Sprintf("Lat: %.3f", lt)).Show(0, 2, 1, 1)

				if SA_Button(trns.REMOVE).Show(0, 3, 1, 1).click {
					SA_SqlWrite("", fmt.Sprintf("DELETE FROM locators WHERE rowid=%d;", rowid))
				}

				SA_DialogEnd()
			}
		}
	}

	//touch
	if start && inside {
		cam.start_pos.X = touch_x //rel, not pixels!
		cam.start_pos.Y = touch_y
		cam.start_tile = LonLatToPos(Cam{Lon: lon, Lat: lat, Zoom: zoom})
	}

	if wheel != 0 && inside && !isZooming {
		cam.zoomOld = cam.Zoom
		cam.Zoom = zoomClamp(cam.Zoom - wheel)
		if cam.zoomOld != cam.Zoom {
			cam.lonOld = cam.Lon
			cam.latOld = cam.Lat

			//where the mouse is
			if wheel < 0 {
				var pos V2
				pos.X = bbStart.X + bbSize.X*touch_x
				pos.Y = bbStart.Y + bbSize.Y*touch_y
				cam.Lon, cam.Lat = PosToLonLat(pos, zoom)
			}

			cam.start_zoom_time = SA_Time()
		}
	}

	if active {
		var move V2
		move.X = cam.start_pos.X - touch_x
		move.Y = cam.start_pos.Y - touch_y

		rx := move.X * bbSize.X
		ry := move.Y * bbSize.Y

		tileX := cam.start_tile.X + rx
		tileY := cam.start_tile.Y + ry

		cam.Lon, cam.Lat = PosToLonLat(V2{tileX, tileY}, cam.Zoom)
	}

	//double click
	if clicks > 1 && end && !isZooming {
		cam.zoomOld = cam.Zoom
		cam.Zoom = zoomClamp(cam.Zoom + 1)

		if cam.zoomOld != cam.Zoom {
			cam.lonOld = cam.Lon
			cam.latOld = cam.Lat

			var pos V2
			pos.X = bbStart.X + bbSize.X*touch_x
			pos.Y = bbStart.Y + bbSize.Y*touch_y
			cam.Lon, cam.Lat = PosToLonLat(pos, zoom)

			cam.start_zoom_time = SA_Time()
		}
	}

	if store.add_locator {
		if end {
			var pos V2
			pos.X = bbStart.X + bbSize.X*touch_x
			pos.Y = bbStart.Y + bbSize.Y*touch_y
			ln, lt := PosToLonLat(pos, zoom)

			SA_SqlWrite("", fmt.Sprintf("INSERT INTO locators(title, pos) VALUES('un-named', '%f, %f');", ln, lt))
			store.add_locator = false
		}

		SAPaint_Cursor("cross")
	}

	SA_ColMax(0, 100)
	SA_RowMax(1, 100)

	//top
	SA_DivStart(0, 0, 1, 1)
	{
		if SA_ButtonBorder("+").Tooltip(trns.ADD_LOCATOR).Pressed(store.add_locator).Show(0, 0, 1, 1).click {
			store.add_locator = !store.add_locator
		}
	}
	SA_DivEnd()

	//bottom
	SA_DivStart(0, 2, 1, 1)
	{
		SA_ColMax(1, 14)
		SA_ColMax(2, 100) //space
		SA_ColMax(3, 3)
		SA_ColMax(4, 3)
		SA_ColMax(5, 2.5)
		SA_ColMax(6, 100) //space
		SA_ColMax(7, 7)

		SA_DivStart(0, 0, 2, 1)
		Measure(*cam)
		SA_DivEnd()

		SA_Editbox(&cam.Lon).Precision(3).ShowDescription(3, 0, 1, 1, "Lon", 1.5, &styles.TextRight)
		SA_Editbox(&cam.Lat).Precision(3).ShowDescription(4, 0, 1, 1, "Lat", 1.5, &styles.TextRight)
		SA_Editbox(&cam.Zoom).Precision(0).ShowDescription(5, 0, 1, 1, trns.ZOOM, 1.5, &styles.TextRight)
		SA_ButtonAlpha("(c)OpenStreetMap contributors").Url("https://www.openstreetmap.org/copyright").Show(7, 0, 1, 1)
	}
	SA_DivEnd()
}

func Render() {
	SA_ColMax(0, 100)
	SA_RowMax(0, 100)

	SA_DivStart(0, 0, 1, 1)
	Map(&store.Cam)
	SA_DivEnd()
}

var styles SA_Styles

func Open() {
	//default
	json.Unmarshal(SA_File("storage_json"), &store)
	json.Unmarshal(SA_File("translations_json:app:resources/translations.json"), &trns)
	json.Unmarshal(SA_File("styles_json"), &styles)
}

func SetupDB() {
	SA_SqlWrite("", "CREATE TABLE IF NOT EXISTS tiles(name TEXT, file BLOB);")
	SA_SqlWrite("", "CREATE TABLE IF NOT EXISTS locators(title TEXT, pos TEXT);")
}

func Save() []byte {
	js, _ := json.MarshalIndent(&store, "", "")
	return js
}
