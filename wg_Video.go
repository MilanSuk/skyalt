package main

import (
	"image/color"
)

type Video struct {
	Path    string
	Tooltip string

	Margin  float64
	Align_h int
	Align_v int

	Cd          color.RGBA
	Draw_border bool
}

func (layout *Layout) AddVideo(x, y, w, h int, path string) *Video {
	props := &Video{Path: path, Align_h: 1, Align_v: 1, Margin: 0.1, Cd: color.RGBA{255, 255, 255, 255}}
	layout._createDiv(x, y, w, h, "Video", props.Build, nil, nil)
	return props
}

func (st *Video) Build(layout *Layout) {

	layout.SetColumn(0, 1, 100)
	layout.SetRow(0, 1, 100)

	im := layout.AddImage(0, 0, 1, 1, st.Path, nil)
	imgDiv := layout.FindGrid(0, 0, 1, 1)

	inited := func() {
		layout.ui.SetRefresh()
	}
	img := layout.ui.win.images.Add(InitWinImagePath_file(st.Path, imgDiv.UID), inited)
	if !img.path.is_playing {
		im.Draw_border = true
	}

	FooterDiv := layout.AddLayout(0, 1, 1, 1)
	FooterDiv.SetColumn(1, 1, 100) //timeline

	statusLabel := "▶"
	if img.path.is_playing {
		statusLabel = "⏸︎"
	}
	PlayBt := FooterDiv.AddButton(0, 0, 1, 1, statusLabel)
	PlayBt.clicked = func() {
		img.SetPlay(!img.path.is_playing, layout.ui.win)
		layout.ui.SetRefresh()
	}

	//show current / total time ....

	//Mute ...

	//Volume ...

	if img != nil && img.path.play_duration > 0 {
		tml := FooterDiv.AddSlider(1, 0, 1, 1, &img.path.play_pos, 0, float64(img.path.play_duration), 1)
		tmlDiv := FooterDiv.FindGrid(1, 0, 1, 1)
		//slider.FuncTooltip() callback converts to data(from miliseconds) ....
		tml.changed = func() {
			if !img.path.is_playing {
				layout.ui.SetRefresh()
			}
			img.SetSeek(img.path.play_pos, layout.ui.win)
		}

		if img.path.is_playing {
			tmlDiv.fnUpdate = func() {
				//CallLayoutUpdates(), will update/redraw after this
			}
		}
	}

}
