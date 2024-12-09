package main

import (
	"fmt"
	"os/exec"
	"runtime"
)

func (st *Button) Draw(rect Rect) {
	st.lock.Lock()
	defer st.lock.Unlock()

	layout := st.layout

	rc := rect
	rc = rc.Cut(0.03)
	rectLabel := rc

	layout.Paint_cursor("hand", rect)
	tip := st.Tooltip
	if st.layout.Shortcut_key != 0 {
		sh := string(st.layout.Shortcut_key)
		if st.layout.Shortcut_key == '\t' {
			sh = "Tab"
		}
		tip += fmt.Sprintf("(Ctrl+%s)", sh)
	}
	layout.Paint_tooltip(tip, rect)

	if st.BrowserUrl != "" {
		layout.Paint_tooltip(st.BrowserUrl, rect)
	}

	//colors
	B := layout.GetPalette().B
	onB := layout.GetPalette().OnB
	cdBack := layout.GetPalette().P
	cdText := layout.GetPalette().OnP

	if st.Background == 0 {
		//no back
		cdText = onB
	} else if st.Background <= 0.3 {
		//menu
		cdText = onB
		cdBack = Color_Aprox(B, onB, 0.2)
	} else if st.Background <= 0.6 {
		//light back
		cdText = cdBack
		cdBack = Color_Aprox(cdBack, B, 0.8)
	}

	if st.BrowserUrl != "" {
		cdText = layout.GetPalette().P
	}

	if st.Cd.A > 0 {

		if st.Background <= 0.3 {
			cdText = st.Cd
		} else {
			cdBack = st.Cd
		}
	}

	cdBack_over := cdBack
	cdText_over := cdText
	cdBack_down := cdBack
	cdText_down := cdText

	cdBack_over = Color_Aprox(cdBack_over, B, 0.2)
	cdText_over = Color_Aprox(cdText_over, B, 0.2)
	cdBack_down = Color_Aprox(cdBack_down, onB, 0.4)
	cdText_down = Color_Aprox(cdText_down, onB, 0.4)

	if st.Cd_fade {
		a := byte(100)
		cdBack.A = a
		cdText.A = a
		cdBack_over.A = a
		cdText_over.A = a
		cdBack_down.A = a
		cdText_down.A = a
	}

	//draw background
	if st.Background > 0 {
		if st.Border {
			rectLabel = rc.Cut(0.06)
		}

		cd := cdBack //new var, because 'Draw_border' also uses 'cdBack'
		if st.Background <= 0.3 {
			cd.A = 0 //no back
		}

		layout.Paint_rect(rectLabel, cd, cdBack_over, cdBack_down, 0)
	}

	//draw border
	if st.Border {
		layout.Paint_rect(rc, cdBack, cdBack_over, cdBack_down, 0.03)
	}

	//draw icon
	if st.Icon != "" {
		rectIcon := rectLabel

		icon_w := rectIcon.W
		if icon_w > 1 {
			icon_w = 1
		}

		switch st.Icon_align {
		case 0:
			rectIcon.W = icon_w

			rectLabel = rectLabel.CutLeft(icon_w)
		case 1:
			//full rect
		case 2:
			rectIcon.X += rectIcon.W - icon_w
			rectIcon.W = icon_w

			rectLabel = rectLabel.CutRight(icon_w)
		}

		rectIcon = rectIcon.Cut(st.Icon_margin)

		layout.Paint_file(rectIcon, false, st.Icon, cdText, cdText_over, cdText_down, 1, 1)
	}

	//draw label
	if st.Value != "" {
		layout.Paint_text(rectLabel, st.Value, "",
			//cdBack, cdBack_over, cdBack_down,
			cdText, cdText_over, cdText_down,
			false, false, uint8(st.Align), 1, true, false, false, 0.1)
	}
}

func (st *Button) Input(in LayoutInput) {
	st.lock.Lock()
	defer st.lock.Unlock()

	clicked := false
	active := in.IsActive
	inside := in.IsInside && (active || !in.IsUse)
	clicked = in.IsUp && active && inside

	if clicked {
		if st.clicked != nil {
			st.clicked()
		}
		if st.clickedEx != nil {
			st.clickedEx(in.NumClicks, in.AltClick)
		}
	}

	//open browser
	if st.BrowserUrl != "" && clicked {
		//dialog Allow/Denied ...
		OsUlit_OpenBrowser(st.BrowserUrl)
	}
}

func OsUlit_OpenBrowser(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return err
}
