package main

import (
	"fmt"
	"image/color"
	"os/exec"
	"runtime"
)

type Button struct {
	Value string //label

	Tooltip string
	Align   int

	Background float64
	Border     bool

	Icon        string
	Icon_align  int
	Icon_margin float64

	BrowserUrl string

	Cd      color.RGBA
	Cd_fade bool

	//Shortcut_key byte

	clicked   func()
	clickedEx func(numClicks int, altClick bool)
}

func (layout *Layout) AddButton(x, y, w, h int, label string) *Button {
	props := &Button{Value: label, Align: 1, Background: 1}
	layout._createDiv(x, y, w, h, "Button", nil, props.Draw, props.Input)
	return props
}
func (layout *Layout) AddButton2(x, y, w, h int, label string) (*Button, *Layout) {
	props := &Button{Value: label, Align: 1, Background: 1}
	lay := layout._createDiv(x, y, w, h, "Button", nil, props.Draw, props.Input)
	return props, lay
}

func (layout *Layout) AddButtonMenu(x, y, w, h int, label string, icon_path string, icon_margin float64) *Button {
	props := &Button{Value: label, Icon: icon_path, Icon_margin: icon_margin, Align: 0, Background: 0.25}
	layout._createDiv(x, y, w, h, "Button", nil, props.Draw, props.Input)
	return props
}

func (layout *Layout) AddButtonMenu2(x, y, w, h int, label string, icon_path string, icon_margin float64) (*Button, *Layout) {
	props := &Button{Value: label, Icon: icon_path, Icon_margin: icon_margin, Align: 0, Background: 0.25}
	lay := layout._createDiv(x, y, w, h, "Button", nil, props.Draw, props.Input)
	return props, lay
}

func (layout *Layout) AddButtonIcon(x, y, w, h int, icon_path string, icon_margin float64, Tooltip string) *Button {
	props := &Button{Icon: icon_path, Icon_align: 1, Icon_margin: icon_margin, Tooltip: Tooltip, Background: 1}
	layout._createDiv(x, y, w, h, "Button", nil, props.Draw, props.Input)
	return props
}
func (layout *Layout) AddButtonIcon2(x, y, w, h int, icon_path string, icon_margin float64, Tooltip string) (*Button, *Layout) {
	props := &Button{Icon: icon_path, Icon_align: 1, Icon_margin: icon_margin, Tooltip: Tooltip, Background: 1}
	lay := layout._createDiv(x, y, w, h, "Button", nil, props.Draw, props.Input)
	return props, lay
}

func (layout *Layout) AddButtonDanger(x, y, w, h int, label string) *Button {
	props := &Button{Value: label, Align: 1, Background: 1, Cd: Paint_GetPalette().E}
	layout._createDiv(x, y, w, h, "Button", nil, props.Draw, props.Input)
	return props
}

func (st *Button) Draw(rect Rect, layout *Layout) (paint LayoutPaint) {
	rc := rect
	rc = rc.Cut(0.03)
	rectLabel := rc

	paint.Cursor("hand", rect)
	tip := st.Tooltip
	if layout.Shortcut_key != 0 {
		sh := string(layout.Shortcut_key)
		if layout.Shortcut_key == '\t' {
			sh = "Tab"
		}
		tip += fmt.Sprintf("(Ctrl+%s)", sh)
	}
	paint.Tooltip(tip, rect)

	if st.BrowserUrl != "" {
		paint.Tooltip(st.BrowserUrl, rect)
	}

	//colors
	B := Paint_GetPalette().B
	onB := Paint_GetPalette().OnB
	cdBack := Paint_GetPalette().P
	cdText := Paint_GetPalette().OnP

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
		cdText = Paint_GetPalette().P
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

		paint.Rect(rectLabel, cd, cdBack_over, cdBack_down, 0)
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

		paint.File(rectIcon, false, st.Icon, cdText, cdText_over, cdText_down, 1, 1)
	}

	//draw label
	if st.Value != "" {
		tx := paint.Text(rectLabel.Cut(0.1), st.Value, "",
			//cdBack, cdBack_over, cdBack_down,
			cdText, cdText_over, cdText_down,
			false, false, uint8(st.Align), 1)
		tx.Multiline = true
	}

	//draw border
	if st.Border {
		paint.Rect(rc, cdBack, cdBack_over, cdBack_down, 0.03)
	}

	return
}

func (st *Button) Input(in LayoutInput, layout *Layout) {
	clicked := false
	active := in.IsActive
	inside := in.IsInside && (active || !in.IsUse)
	clicked = in.IsUp && active && inside

	if layout.Shortcut_key != 0 && layout.Shortcut_key == in.Shortcut_key {
		clicked = true
	}

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
