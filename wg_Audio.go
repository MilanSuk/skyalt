package main

type Audio struct {
	Path    string
	Tooltip string
}

func (layout *Layout) AddAudio(x, y, w, h int, path string) *Audio {
	props := &Audio{Path: path}
	layout._createDiv(x, y, w, h, "Audio", props.Build, nil, nil)
	return props
}

func (st *Audio) Build(layout *Layout) {
	inited := func() {
		layout.ui.SetRefresh()
	}
	img := layout.ui.win.images.Add(InitWinImagePath_file(st.Path, layout.UID), inited)

	layout.SetRow(0, 0, 100)
	layout.SetRow(2, 0, 100)

	layout.SetColumn(1, 1, 100) //timeline

	statusLabel := "▶"
	if img.path.is_playing {
		statusLabel = "⏸︎"
	}
	PlayBt := layout.AddButton(0, 1, 1, 1, statusLabel)
	PlayBt.clicked = func() {
		img.SetPlay(!img.path.is_playing, layout.ui.win)
		layout.ui.SetRefresh()
	}

	//Mute ...

	//Volume ...

	if img != nil && img.path.play_duration > 0 {
		tml := layout.AddSlider(1, 1, 1, 1, &img.path.play_pos, 0, float64(img.path.play_duration), 1)
		tmlDiv := layout.FindGrid(1, 1, 1, 1)
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
