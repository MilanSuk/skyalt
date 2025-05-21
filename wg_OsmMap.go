package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"math"
	"strconv"
	"time"
)

type OsmMapLoc struct {
	Lon   float64
	Lat   float64
	Label string
}
type OsmMapLocators struct {
	Locators []OsmMapLoc
	clicked  func(i int)
}

type OsmMapSegmentTrk struct {
	Lon  float64
	Lat  float64
	Ele  float64
	Time string
	Cd   color.RGBA
}
type OsmMapSegment struct {
	Trkpts []OsmMapSegmentTrk
	Label  string
}

type OsmMapRoute struct {
	Segments []OsmMapSegment
}

type OsmMapCam struct {
	Lon, Lat, Zoom float64
}

type OsmMap struct {
	Cam *OsmMapCam

	Locators []OsmMapLocators
	Routes   []OsmMapRoute

	changed func()

	rect Rect
}

func (layout *Layout) AddOsmMap(x, y, w, h int, cam *OsmMapCam) *OsmMap {
	props := &OsmMap{Cam: cam}
	layout._createDiv(x, y, w, h, "OsmMap", props.Build, props.Draw, props.Input)
	return props
}

func (mp *OsmMap) AddLocators(loc OsmMapLocators) {
	mp.Locators = append(mp.Locators, loc)
}
func (mp *OsmMap) AddRoute(route OsmMapRoute) {
	mp.Routes = append(mp.Routes, route)
}

func (st *OsmMap) Build(layout *Layout) {
	layout.SetColumn(0, 5, 100)
	layout.SetColumn(1, 6, 12)
	layout.SetColumn(2, 0.1, 100)
	layout.SetColumn(3, 2, 6)
	layout.SetRow(0, 1, 100)

	layout.scrollH.Show = false

	//lon,lat,zoom
	{
		Div := layout.AddLayout(1, 1, 1, 1)
		Div.SetColumn(0, 1, 1)
		Div.SetColumn(1, 1, 2)
		Div.SetColumn(2, 1, 1)
		Div.SetColumn(3, 1, 2)
		Div.SetColumn(4, 1, 1.5)
		Div.SetColumn(5, 1, 1)
		Div.scrollH.Show = false
		Div.scrollH.Narrow = true

		tx := Div.AddText(0, 0, 1, 1, "Lon")
		tx.Align_h = 2
		edLon := Div.AddEditbox(1, 0, 1, 1, &st.Cam.Lon)
		edLon.ValueFloatPrec = 4
		edLon.changed = func() {
			if st.changed != nil {
				st.changed()
			}
		}

		tx = Div.AddText(2, 0, 1, 1, "Lat")
		tx.Align_h = 2
		edLat := Div.AddEditbox(3, 0, 1, 1, &st.Cam.Lat)
		edLat.ValueFloatPrec = 4
		edLat.changed = func() {
			if st.changed != nil {
				st.changed()
			}
		}

		tx = Div.AddText(4, 0, 1, 1, "Zoom")
		tx.Align_h = 2
		edZoom := Div.AddEditbox(5, 0, 1, 1, &st.Cam.Zoom)
		edZoom.ValueFloatPrec = 0
		edZoom.changed = func() {
			if st.changed != nil {
				st.changed()
			}
		}

		btA := Div.AddButton(6, 0, 1, 1, "+")
		btS := Div.AddButton(7, 0, 1, 1, "-")
		btT, btTLay := Div.AddButtonIcon2(8, 0, 1, 1, "resources/target.png", 0.1, "Focus")
		btTLay.Enable = len(st.Locators) > 0 || len(st.Routes) > 0
		btA.Background = 0.5
		btS.Background = 0.5
		btT.Background = 0.5
		btA.clicked = func() {
			st.Cam.Zoom = _CompOsmMap_zoomClamp(st.Cam.Zoom + 1)
			if st.changed != nil {
				st.changed()
			}
		}
		btS.clicked = func() {
			st.Cam.Zoom = _CompOsmMap_zoomClamp(st.Cam.Zoom - 1)
			if st.changed != nil {
				st.changed()
			}
		}
		btT.clicked = func() {
			canvas_size := OsV2f{float32(st.rect.W), float32(st.rect.H)}
			tile := 256 / float64(layout.Cell()) * 1
			*st.Cam = st.GetDefaultCam(canvas_size, tile)
			if st.changed != nil {
				st.changed()
			}
		}
	}

	copyright := layout.AddButton(3, 1, 1, 1, layout.ui.sync.map_settings.Copyright)
	copyright.Background = 0
	copyright.BrowserUrl = layout.ui.sync.map_settings.Copyright_url

}

func (st *OsmMap) Draw(rect Rect, layout *Layout) (paint LayoutPaint) {
	st.rect = rect

	//zoom into default position
	if st.Cam != nil && st.Cam.Zoom == 0 && st.Cam.Lon == 0 && st.Cam.Lat == 0 {
		canvas_size := OsV2f{float32(rect.W), float32(rect.H)}
		tile := 256 / float64(layout.Cell()) * 1.0

		*st.Cam = st.GetDefaultCam(canvas_size, tile)

		if st.changed != nil {
			st.changed()
		}
	}
	lon, lat, zoom, scale, isZooming := g_map.GetAnim(layout, st.Cam.Lon, st.Cam.Lat, st.Cam.Zoom)
	canvas_size := OsV2f{float32(rect.W), float32(rect.H)}
	tile := 256 / float64(layout.Cell()) * scale
	if isZooming {
		layout.RedrawThis()
	}

	lon, lat = _CompOsmMap_camCheck(canvas_size, tile, lon, lat, zoom)
	st.Cam.Lon, st.Cam.Lat = _CompOsmMap_camCheck(canvas_size, tile, st.Cam.Lon, st.Cam.Lat, st.Cam.Zoom)

	bbStart, bbEnd, _ := _CompOsmMap_camBbox(canvas_size, tile, lon, lat, zoom)

	//draw tiles
	for y := float64(int(bbStart.Y)); y < float64(bbEnd.Y); y++ {
		for x := float64(int(bbStart.X)); x < float64(bbEnd.X); x++ {
			if x < 0 || y < 0 {
				continue
			}

			fnGetBlob := func(fnDone func(bytes []byte, err error)) error {

				type GetTile struct {
					X    int //tile's X position
					Y    int //tile's Y position
					Zoom int //map's zoom

					Out_image []byte
				}

				fnAsyncDone := func(bytes []byte, ui *UI, cmds []ToolCmd, err error, start_time float64) {
					if err != nil {
						fnDone(nil, err)
						return
					}

					var out GetTile
					er := json.Unmarshal(bytes, &out)
					if er != nil {
						fnDone(nil, er)
						return
					}

					fnDone(out.Out_image, nil) //ok

				}

				//un-design .....
				_, err := layout.ui.router.CallAsync(0, "GetMapTile", GetTile{X: int(x), Y: int(y), Zoom: int(zoom)}, nil, fnAsyncDone)
				return err

			}

			tileCoord_sx := (x - float64(bbStart.X)) * tile
			tileCoord_sy := (y - float64(bbStart.Y)) * tile
			cdWhite := color.RGBA{255, 255, 255, 255}
			paint.File(Rect{X: tileCoord_sx, Y: tileCoord_sy, W: tile, H: tile},
				InitWinImagePath_load(fmt.Sprintf("map_tile:%d-%d-%d", int(x), int(y), int(zoom)), fnGetBlob),
				cdWhite, cdWhite, cdWhite,
				0, 0)
		}
	}

	st.DrawMeasureStrip(rect, &paint, layout)

	//locators
	for _, locator := range st.Locators {
		cd := color.RGBA{0, 0, 0, 255}
		cd_over := cd //...
		cd_down := cd //...

		for _, it := range locator.Locators {

			p := _CompOsmMap_lonLatToPos(it.Lon, it.Lat, zoom)

			tileCoord_sx := float64(p.X-bbStart.X) * tile
			tileCoord_sy := float64(p.Y-bbStart.Y) * tile
			rad := 1.0

			locRect := Rect{X: tileCoord_sx - rad/2, Y: tileCoord_sy - rad, W: rad, H: rad}

			paint.File(locRect, InitWinImagePath_file("resources/locator.png"), cd, cd_over, cd_down, 1, 1)
		}
	}

	//segments
	tl := OsV2f{float32(tile) / float32(canvas_size.X), float32(tile) / float32(canvas_size.Y)}
	for _, route := range st.Routes {
		for _, segs := range route.Segments {

			var last OsV2f
			for i, pt := range segs.Trkpts {

				p := _CompOsmMap_lonLatToPos(pt.Lon, pt.Lat, zoom)
				pos := p.Sub(bbStart).Mul(tl)

				if pos.Sub(last).Len() < 0.008 {
					continue
				}

				cd := pt.Cd
				if cd.A == 0 {
					cd = color.RGBA{0, 0, 0, 255}
				}

				if i > 0 {
					paint.Line(rect, float64(last.X), float64(last.Y), float64(pos.X), float64(pos.Y), cd, 0.1)
				}
				last = pos
			}
		}
	}

	return
}

func (st *OsmMap) Input(in LayoutInput, layout *Layout) {
	redrawNow := false

	st.Cam.Zoom = _CompOsmMap_zoomClamp(st.Cam.Zoom) //check

	canvas_size := OsV2f{float32(in.Rect.W), float32(in.Rect.H)}
	lon, lat, zoom, scale, isZooming := g_map.GetAnim(layout, st.Cam.Lon, st.Cam.Lat, st.Cam.Zoom)
	if isZooming {
		redrawNow = true
	}
	tile := 256 / float64(layout.Cell()) * scale

	bbStart, _, bbSize := _CompOsmMap_camBbox(canvas_size, tile, lon, lat, zoom)

	start := in.IsStart
	inside := in.IsInside
	end := in.IsEnd
	active := in.IsActive
	touch_x := float32(in.X) / canvas_size.X //<0, 1>
	touch_y := float32(in.Y) / canvas_size.Y //<0, 1>
	wheel := in.Wheel
	clicks := in.NumClicks
	altClick := in.AltClick

	//touch
	if start && inside {
		g_map.start_pos.X = touch_x //rel, not pixels!
		g_map.start_pos.Y = touch_y
		g_map.start_tile = _CompOsmMap_lonLatToPos(lon, lat, zoom)
	}

	if wheel != 0 && inside && !isZooming {

		g_map.zoomOld = st.Cam.Zoom
		zoomNew := _CompOsmMap_zoomClamp(g_map.zoomOld - float64(wheel))
		if g_map.zoomOld != zoomNew {
			g_map.dom_uid = layout.UID

			g_map.lonOld = st.Cam.Lon
			g_map.latOld = st.Cam.Lat

			//get touch lon and lat
			touch_lon, touch_lat := _CompOsmMap_posToLonLat(OsV2f{bbStart.X + bbSize.X*touch_x, bbStart.Y + bbSize.Y*touch_y}, g_map.zoomOld)
			//get new zoom touch pos
			pos := _CompOsmMap_lonLatToPos(touch_lon, touch_lat, zoomNew)
			//get center
			pos.X -= bbSize.X * (touch_x - 0.5)
			pos.Y -= bbSize.Y * (touch_y - 0.5)
			//get new zoom lon and lat
			lonNew, latNew := _CompOsmMap_posToLonLat(pos, zoomNew)
			st.Cam.Lon = lonNew
			st.Cam.Lat = latNew
			st.Cam.Zoom = zoomNew

			g_map.zoom_start_time = _CompOsmMap_getTime()
			g_map.zoom_active = true
			redrawNow = true
		}
	}

	if active {
		g_map.dom_uid = layout.UID
		g_map.zoom_active = false

		var move OsV2f
		move.X = g_map.start_pos.X - touch_x
		move.Y = g_map.start_pos.Y - touch_y

		rx := move.X * bbSize.X
		ry := move.Y * bbSize.Y

		tileX := g_map.start_tile.X + rx
		tileY := g_map.start_tile.Y + ry

		lonNew, latNew := _CompOsmMap_posToLonLat(OsV2f{tileX, tileY}, st.Cam.Zoom)

		if st.Cam.Lon != lonNew || st.Cam.Lat != latNew {
			redrawNow = true
		}

		st.Cam.Lon = lonNew
		st.Cam.Lat = latNew
	}

	//double click
	if clicks > 1 && end && !isZooming {
		z := 1.0
		if altClick {
			z = -1.0
		}

		g_map.zoomOld = st.Cam.Zoom
		zoomNew := _CompOsmMap_zoomClamp(g_map.zoomOld + z)

		if g_map.zoomOld != zoomNew {
			g_map.dom_uid = layout.UID
			g_map.lonOld = st.Cam.Lon
			g_map.latOld = st.Cam.Lat

			//get touch lon and lat
			touch_lon, touch_lat := _CompOsmMap_posToLonLat(OsV2f{bbStart.X + bbSize.X*touch_x, bbStart.Y + bbSize.Y*touch_y}, g_map.zoomOld)
			//get new zoom touch pos
			pos := _CompOsmMap_lonLatToPos(touch_lon, touch_lat, zoomNew)
			//get center
			pos.X -= bbSize.X * (touch_x - 0.5)
			pos.Y -= bbSize.Y * (touch_y - 0.5)
			//get new zoom lon and lat
			lonNew, latNew := _CompOsmMap_posToLonLat(pos, zoomNew)
			st.Cam.Lon = lonNew
			st.Cam.Lat = latNew
			st.Cam.Zoom = zoomNew

			g_map.zoom_start_time = _CompOsmMap_getTime()
			g_map.zoom_active = true

			redrawNow = true
		}
	}

	if redrawNow {
		layout.RedrawThis()
	}

	//locators
	for _, locator := range st.Locators {
		for i, it := range locator.Locators {

			this_i := i

			p := _CompOsmMap_lonLatToPos(it.Lon, it.Lat, zoom)

			tileCoord_sx := float64(p.X-bbStart.X) * tile
			tileCoord_sy := float64(p.Y-bbStart.Y) * tile
			rad := 1.0

			locRect := Rect{X: tileCoord_sx - rad/2, Y: tileCoord_sy - rad, W: rad, H: rad}

			if start && locRect.IsInside(in.X, in.Y) {
				if locator.clicked != nil {
					locator.clicked(this_i)
				}
			}

		}
	}

	if wheel != 0 || clicks > 1 || (end && (g_map.start_pos.X != touch_x || g_map.start_pos.Y != touch_y)) {
		if st.changed != nil {
			st.changed()
		}
	}
}

func _CompOsmMap_getTime() float64 {
	return float64(time.Now().UnixMilli()) / 1000
}

func _CompOsmMap_metersPerPixel(lat, zoom float64) float64 {
	return 156543.034 * math.Cos(lat/180*math.Pi) / math.Pow(2, zoom)
}

func _CompOsmMap_lonLatToPos(lon, lat, zoom float64) OsV2f {
	x := (lon + 180) / 360 * math.Pow(2, zoom)
	y := (1 - math.Log(math.Tan(lat*math.Pi/180)+1/math.Cos(lat*math.Pi/180))/math.Pi) / 2 * math.Pow(2, zoom)
	return OsV2f{float32(x), float32(y)}
}

func _CompOsmMap_posToLonLat(pos OsV2f, zoom float64) (float64, float64) {
	lon := float64(pos.X)/math.Pow(2, zoom)*360 - 180 //long

	n := math.Pi - 2*math.Pi*float64(pos.Y)/math.Pow(2, zoom)
	lat := 180 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(n*-1))) //lat
	return lon, lat
}

func _CompOsmMap_clamp(v, min, max float64) float64 {
	if v < min {
		v = min
	}
	if v > max {
		v = max
	}
	return v
}

func _CompOsmMap_camBbox(res OsV2f, tile float64, lon, lat, zoom float64) (OsV2f, OsV2f, OsV2f) {
	hres := res.MulV(0.5)

	tilePos := _CompOsmMap_lonLatToPos(lon, lat, zoom)
	max_res := math.Pow(2, zoom)

	var start, end OsV2f

	start.X = float32(_CompOsmMap_clamp((float64(tilePos.X)*tile-float64(hres.X))/tile, 0, max_res))
	start.Y = float32(_CompOsmMap_clamp((float64(tilePos.Y)*tile-float64(hres.Y))/tile, 0, max_res))
	end.X = float32(_CompOsmMap_clamp((float64(tilePos.X)*tile+float64(hres.X))/tile, 0, max_res))
	end.Y = float32(_CompOsmMap_clamp((float64(tilePos.Y)*tile+float64(hres.Y))/tile, 0, max_res))

	size := end.Sub(start)

	return start, end, size
}

func _CompOsmMap_camCheck(res OsV2f, tile float64, lon, lat, zoom float64) (float64, float64) {
	if res.X <= 0 || res.Y <= 0 {
		return 0, 0
	}

	bbStart, bbEnd, bbSize := _CompOsmMap_camBbox(res, tile, lon, lat, zoom)

	maxTiles := math.Pow(2, zoom)

	def_bbox_size := OsV2f{res.X / float32(tile), res.Y / float32(tile)}

	if bbStart.X <= 0 {
		bbSize.X = def_bbox_size.X
		bbStart.X = 0
	}

	if bbStart.Y <= 0 {
		bbSize.Y = def_bbox_size.Y
		bbStart.Y = 0
	}

	if bbEnd.X >= float32(maxTiles) {
		bbSize.X = def_bbox_size.X
		bbStart.X = float32(maxTiles - float64(bbSize.X))
		if bbStart.X < 0 {
			bbStart.X = 0
		}
	}

	if bbEnd.Y >= float32(maxTiles) {
		bbSize.Y = def_bbox_size.Y
		bbStart.Y = float32(maxTiles - float64(bbSize.Y))
		if bbStart.Y < 0 {
			bbStart.Y = 0
		}
	}

	return _CompOsmMap_posToLonLat(OsV2f{bbStart.X + bbSize.X/2, bbStart.Y + bbSize.Y/2}, zoom)
}

func _CompOsmMap_zoomClamp(z float64) float64 {
	return _CompOsmMap_clamp(z, 0, 19)
}

func (st *OsmMap) DrawMeasureStrip(rect Rect, paint *LayoutPaint, layout *Layout) {

	metersPerPixels := _CompOsmMap_metersPerPixel(st.Cam.Lat, st.Cam.Zoom)

	W := float64(1.5) //strip width
	metersPerStrip := metersPerPixels * W * float64(layout.Cell())

	unitText := "m"
	if metersPerStrip > 1000 {
		metersPerStrip /= 1000
		unitText = "km"
	}

	meters := math.Round(metersPerStrip) //better rounding(10s, 100s, 1000s, etc.) ...

	W *= meters / metersPerStrip //fix strip_width

	cdB := color.RGBA{0, 0, 0, 255}
	cdW := color.RGBA{255, 255, 255, 255}

	SX := 0.3
	SY := rect.H - 0.9
	rbox := Rect{X: SX + 0, Y: SY, W: W, H: 0.2}
	rtext := Rect{X: SX - W/2, Y: SY + 0.3, W: W, H: 0.5}

	//background shadow
	/*{
		rfull := rbox
		rfull.W *= 3
		rfull = rfull.Cut(-0.06)
		cdF := cdW.Aprox(cdB, 0.1)
		layout.Paint_rect(rfull, cdF, cdF, cdF, 0) //background white
	}*/

	paint.Rect(rbox, cdB, cdB, cdB, 0)
	paint.Text(rtext, "<small>"+strconv.Itoa(int(meters*0)), "", cdB, cdB, cdB, false, false, 1, 1)
	rbox.X += W
	rtext.X += W

	paint.Rect(rbox, cdW, cdW, cdW, 0)
	paint.Text(rtext, "<small>"+strconv.Itoa(int(meters*1)), "", cdB, cdB, cdB, false, false, 1, 1)
	rbox.X += W
	rtext.X += W

	paint.Rect(rbox, cdB, cdB, cdB, 0)
	paint.Text(rtext, "<small>"+strconv.Itoa(int(meters*2)), "", cdB, cdB, cdB, false, false, 1, 1)
	rbox.X += W
	rtext.X += W

	paint.Text(rtext, "<small>"+strconv.FormatFloat(meters*3, 'f', 0, 64)+" "+unitText, "", cdB, cdB, cdB, false, false, 1, 1)
}

func (st *OsmMap) GetDefaultCam(canvas_size OsV2f, tile float64) OsmMapCam {

	var cam OsmMapCam
	if len(st.Locators) > 0 {
		lon := 0.0
		lat := 0.0
		n := 0
		for _, loc := range st.Locators {
			for _, lc := range loc.Locators {
				lon += lc.Lon
				lat += lc.Lat
				n++
			}
		}
		cam.Lon = lon / float64(n)
		cam.Lat = lat / float64(n)

		//zoom
		var starts [20]OsV2f
		var ends [20]OsV2f
		for z := range starts {
			starts[z], ends[z], _ = _CompOsmMap_camBbox(canvas_size, tile, cam.Lon, cam.Lat, float64(z))
		}

		zooms_fail := 20
		for _, loc := range st.Locators {
			for _, lc := range loc.Locators {
				for z := 0; z < zooms_fail; z++ {
					p := _CompOsmMap_lonLatToPos(lc.Lon, lc.Lat, float64(z))
					if p.X < starts[z].X || p.Y < starts[z].Y || p.X >= ends[z].X || p.Y >= ends[z].Y {
						//fail
						if z < zooms_fail {
							zooms_fail = z
						}
						break
					}
				}
			}
		}
		if zooms_fail > 0 {
			zooms_fail--
		}
		cam.Zoom = float64(zooms_fail)

	} else if len(st.Routes) > 0 {
		lon := 0.0
		lat := 0.0
		n := 0
		for _, route := range st.Routes {
			for _, seg := range route.Segments {
				for _, pts := range seg.Trkpts {
					lon += pts.Lon
					lat += pts.Lat
					n++
				}
			}
		}
		cam.Lon = lon / float64(n)
		cam.Lat = lat / float64(n)

		//zoom
		var starts [20]OsV2f
		var ends [20]OsV2f
		for z := range starts {
			starts[z], ends[z], _ = _CompOsmMap_camBbox(canvas_size, tile, cam.Lon, cam.Lat, float64(z))
		}

		zooms_fail := 20
		for _, route := range st.Routes {
			for _, seg := range route.Segments {
				for _, pts := range seg.Trkpts {
					for z := 0; z < zooms_fail; z++ {
						p := _CompOsmMap_lonLatToPos(pts.Lon, pts.Lat, float64(z))
						if p.X < starts[z].X || p.Y < starts[z].Y || p.X >= ends[z].X || p.Y >= ends[z].Y {
							//fail
							if z < zooms_fail {
								zooms_fail = z
							}
							break
						}
					}
				}
			}
		}
		if zooms_fail > 0 {
			zooms_fail--
		}
		cam.Zoom = float64(zooms_fail)
	}
	return cam
}

type OsmMapActive struct {
	dom_uid uint64

	lonOld, latOld, zoomOld float64
	start_pos               OsV2f
	start_tile              OsV2f
	zoom_start_time         float64
	zoom_active             bool
}

var g_map OsmMapActive

func (mp *OsmMapActive) isZooming() (bool, float64, float64) {

	ANIM_TIME := 0.2
	dt := _CompOsmMap_getTime() - mp.zoom_start_time

	if mp.zoom_active && dt > ANIM_TIME {
		mp.zoom_active = false
	}
	isZooming := mp.zoom_active

	return isZooming, dt, ANIM_TIME
}

func (mp *OsmMapActive) GetAnim(layout *Layout, lon float64, lat float64, zoom float64) (float64, float64, float64, float64, bool) {

	scale := float64(1)
	isZooming, dt, ANIM_TIME := mp.isZooming()
	if isZooming && mp.dom_uid == layout.UID {
		t := dt / ANIM_TIME
		if zoom > mp.zoomOld {
			scale = 1 + t
		} else {
			scale = 1 - t/2
		}
		zoom = mp.zoomOld
		lon = mp.lonOld + (lon-mp.lonOld)*t
		lat = mp.latOld + (lat-mp.latOld)*t
	} else {
		isZooming = false
	}

	return lon, lat, zoom, scale, isZooming
}
