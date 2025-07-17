package main

import (
	"image/color"
)

type Image struct {
	Tooltip string

	Blob []byte
	Path string

	Cd          color.RGBA
	Draw_border bool

	Margin  float64
	Align_h int
	Align_v int

	Translate_x, Translate_y float64
	Scale_x, Scale_y         float64

	User_input bool //scroll, double-click, move
}

func (layout *Layout) AddImage(x, y, w, h int, path string, blob []byte) *Image {
	props := &Image{Path: path, Blob: blob, Align_h: 1, Align_v: 1, Margin: 0.1, Cd: color.RGBA{255, 255, 255, 255}}
	layout._createDiv(x, y, w, h, "Image", nil, props.Draw, props.Input)
	return props
}

func (st *Image) Draw(rect Rect, layout *Layout) (paint LayoutPaint) {
	paint.Tooltip(st.Tooltip, rect)

	Cd := st.Cd

	rc := rect.Cut(st.Margin)
	if len(st.Blob) > 0 {
		paint.File(rc, InitWinImagePath_blob(st.Blob, layout.UID), Cd, Cd, Cd, uint8(st.Align_h), uint8(st.Align_v))
	} else if st.Path != "" {
		path := InitWinImagePath_file(st.Path, layout.UID)
		paint.File(rc, path, Cd, Cd, Cd, uint8(st.Align_h), uint8(st.Align_v))
	} else {
		paint.Rect(rc, Cd, Cd, Cd, 0)
	}

	if st.Draw_border {
		cd := layout.GetPalette().GetGrey(0.5)
		paint.Rect(rc, cd, cd, cd, 0.03)
	}
	return
}

//....
/*func (st *Image) Paint(layout *Layout, it *LayoutDrawPrim) {


	cell := float64(layout.Cell())
	buff := layout.ui.GetWin().buff

	coord := layout.getCanvasPx(it.Rect)
	cd := layout.GetCd(it.Cd, it.Cd_over, it.Cd_down)

	var tx, ty, sx, sy float64
	if st.User_input {
		tx = st.Translate_x * cell
		ty = st.Translate_y * cell
		sx = st.Scale_x
		sy = st.Scale_y
	}

	path := InitWinMedia_url(it.Text)
	img := buff.AddImage(path, coord, cd, OsV2{int(it.Align_h), int(it.Align_v)}, &tx, &ty, &sx, &sy, layout.GetPalette().E, layout.Cell())
	if img != nil {
		st.orig_size = img.origSize
	}

	if st.User_input {
		st.Translate_x = tx / cell
		st.Translate_y = ty / cell
		st.Scale_x = sx
		st.Scale_y = sy
	}
}*/

func (st *Image) Input(in LayoutInput, layout *Layout) {

	redrawNow := false

	if st.User_input {

		start := in.IsStart
		//end := in.IsEnd
		inside := in.IsInside
		active := in.IsActive
		wheel := in.Wheel
		//clicks := in.NumClicks
		touch_x := in.X
		touch_y := in.Y

		//touch
		if start && inside {
			g_image_active.start_touch_x = touch_x
			g_image_active.start_touch_y = touch_y
			g_image_active.start_tx = st.Translate_x
			g_image_active.start_ty = st.Translate_y
		}

		if wheel != 0 && inside && layout.findParentScroll() == nil {
			zoom := 1.1
			if wheel > 0 {
				zoom = 0.9
			}
			if zoom != 1.0 {
				cell := float64(layout.Cell())

				img := layout.ui.win.images.Add(InitWinImagePath_file(st.Path, layout.UID), nil)
				orig_size := img.texture.size

				iw := (float64(orig_size.X) * st.Scale_x / cell) //image cells
				ih := (float64(orig_size.Y) * st.Scale_y / cell)
				ix := (touch_x - st.Translate_x) / iw //<0-1> in image(zoomed)
				iy := (touch_y - st.Translate_y) / ih

				st.Scale_x = st.Scale_x * zoom //OsClampFloat(st.Scale_x*zoom, 0.1, 5.0)
				st.Scale_y = st.Scale_y * zoom //OsClampFloat(st.Scale_y*zoom, 0.1, 5.0)
				if st.Scale_x < 0.1 {
					st.Scale_x = 0.1
				}
				if st.Scale_y < 0.1 {
					st.Scale_y = 0.1
				}
				if st.Scale_x > 5.0 {
					st.Scale_x = 5.0
				}
				if st.Scale_y > 5.0 {
					st.Scale_y = 5.0
				}

				//compute again, after zoom
				iw = (float64(orig_size.X) * st.Scale_x / cell)
				ih = (float64(orig_size.Y) * st.Scale_y / cell)

				//translation
				st.Translate_x = (touch_x - iw*ix)
				st.Translate_y = (touch_y - ih*iy)

				redrawNow = true
			}
		}
		/*if clicks > 1 && end {
			//open new (almost) fullscreen dialog ....
		}*/

		if active {
			tx := g_image_active.start_tx + (touch_x - g_image_active.start_touch_x)
			ty := g_image_active.start_ty + (touch_y - g_image_active.start_touch_y)
			st.Translate_x = tx
			st.Translate_y = ty

			if st.Translate_x != tx {
				redrawNow = true
			}
			if st.Translate_y != ty {
				redrawNow = true
			}
		}
	}

	if redrawNow {
		layout.RedrawThis()
	}

}

type CompImageActive struct {
	start_touch_x, start_touch_y float64
	start_tx, start_ty           float64
}

var g_image_active CompImageActive
