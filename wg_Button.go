package main

import (
	"fmt"
	"image/color"
	"os/exec"
	"runtime"
)

type Button struct {
	Tooltip string
	Value   string //label

	Align int

	Background float64
	Border     bool

	IconPath    string
	IconBlob    []byte
	Icon_margin float64

	Icon2Path    string
	Icon2Blob    []byte
	Icon2_margin float64

	BrowserUrl string

	Cd      color.RGBA
	Cd_fade bool

	ErrorColor   bool
	Shortcut_key rune

	clicked   func()
	clickedEx func(numClicks int, altClick bool)
}

func (layout *Layout) AddButton(x, y, w, h int, label string) *Button {
	props := &Button{Value: label, Align: 1, Background: 1}
	lay := layout._createDiv(x, y, w, h, "Button", props.Build, props.Draw, props.Input)
	lay.fnHasShortcut = props.HasShortcut
	lay.fnAutoResize = props.autoResize
	lay.fnGetAutoResizeMargin = props.getAutoResizeMargin
	lay.fnGetLLMTip = props.getLLMTip
	return props
}
func (layout *Layout) AddButton2(x, y, w, h int, label string) (*Button, *Layout) {
	props := &Button{Value: label, Align: 1, Background: 1}
	lay := layout._createDiv(x, y, w, h, "Button", props.Build, props.Draw, props.Input)
	lay.fnHasShortcut = props.HasShortcut
	lay.fnAutoResize = props.autoResize
	lay.fnGetAutoResizeMargin = props.getAutoResizeMargin
	lay.fnGetLLMTip = props.getLLMTip
	return props, lay
}

func (layout *Layout) AddButtonMenu(x, y, w, h int, label string, icon_path string, icon_margin float64) *Button {
	props := &Button{Value: label, IconPath: icon_path, Icon_margin: icon_margin, Align: 0, Background: 0.25}
	lay := layout._createDiv(x, y, w, h, "Button", props.Build, props.Draw, props.Input)
	lay.fnHasShortcut = props.HasShortcut
	lay.fnAutoResize = props.autoResize
	lay.fnGetAutoResizeMargin = props.getAutoResizeMargin
	lay.fnGetLLMTip = props.getLLMTip
	return props
}

func (layout *Layout) AddButtonIcon(x, y, w, h int, icon_path string, icon_margin float64, tooltip string) *Button {
	props := &Button{IconPath: icon_path, Icon_margin: icon_margin, Background: 1, Tooltip: tooltip}
	lay := layout._createDiv(x, y, w, h, "Button", props.Build, props.Draw, props.Input)
	lay.fnHasShortcut = props.HasShortcut
	lay.fnAutoResize = props.autoResize
	lay.fnGetAutoResizeMargin = props.getAutoResizeMargin
	lay.fnGetLLMTip = props.getLLMTip
	return props
}
func (layout *Layout) AddButtonIcon2(x, y, w, h int, icon_path string, icon_margin float64, tooltip string) (*Button, *Layout) {
	props := &Button{IconPath: icon_path, Icon_margin: icon_margin, Background: 1, Tooltip: tooltip}
	lay := layout._createDiv(x, y, w, h, "Button", props.Build, props.Draw, props.Input)
	lay.fnHasShortcut = props.HasShortcut
	lay.fnAutoResize = props.autoResize
	lay.fnGetAutoResizeMargin = props.getAutoResizeMargin
	lay.fnGetLLMTip = props.getLLMTip
	return props, lay
}

func (layout *Layout) AddButtonDanger(x, y, w, h int, label string) *Button {
	props := &Button{Value: label, Align: 1, Background: 1, Cd: layout.GetPalette().E}
	lay := layout._createDiv(x, y, w, h, "Button", props.Build, props.Draw, props.Input)
	lay.fnHasShortcut = props.HasShortcut
	lay.fnAutoResize = props.autoResize
	lay.fnGetAutoResizeMargin = props.getAutoResizeMargin
	lay.fnGetLLMTip = props.getLLMTip
	return props
}

func (st *Button) Build(layout *Layout) {
	layout.scrollV.Show = false
	layout.scrollH.Show = false
}

func (st *Button) getLLMTip(layout *Layout) string {
	return Layout_buildLLMTip("Button", "labeled ", "\""+st.Value+"\"", st.Tooltip)
}

func (st *Button) Draw(rect Rect, layout *Layout) (paint LayoutPaint) {

	rc := rect
	rc = rc.Cut(0.03)
	rectLabel := rc

	layout.Back_rounding = true //for disable

	paint.Cursor("hand", rect)
	tip := st.Tooltip
	if st.Shortcut_key != 0 {
		sh := string(st.Shortcut_key)

		switch st.Shortcut_key {
		case '\t':
			sh = "Tab"
		case 37:
			sh = "←"
		case 38:
			sh = "↑"
		case 39:
			sh = "→"
		case 40:
			sh = "↓"
		}

		tip += fmt.Sprintf("(Ctrl+%s)", sh)
	}
	paint.Tooltip(tip, rect)

	if st.BrowserUrl != "" {
		paint.Tooltip(st.BrowserUrl, rect)
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

	if st.ErrorColor {
		cdText_over = layout.GetPalette().E
		cdText_down = layout.GetPalette().E
	}

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

		paint.RectRad(rectLabel, cd, cdBack_over, cdBack_down, 0, layout.getRounding())
	}

	//draw icon
	hasIconL := st.IconPath != "" || len(st.IconBlob) > 0
	hasIconR := st.Icon2Path != "" || len(st.Icon2Blob) > 0

	if hasIconL {
		rectIcon := rectLabel

		icon_w := OsMinFloat(1, rectIcon.W)

		if st.Value != "" || hasIconR {
			//on left side
			rectIcon.W = icon_w
			rectLabel = rectLabel.CutLeft(icon_w)
		}

		var pt WinImagePath
		if st.IconPath != "" {
			pt = InitWinImagePath_file(st.IconPath, layout.UID)
		} else {
			pt = InitWinImagePath_blob(st.IconBlob, layout.UID)
		}
		rectIcon = rectIcon.Cut(st.Icon_margin)
		paint.File(rectIcon, pt, cdText, cdText_over, cdText_down, 1, 1)
	}

	//draw icon2
	if hasIconR {
		rectIcon := rectLabel

		icon_w := OsMinFloat(1, rectIcon.W)

		//if st.Value != "" || hasIconL {
		//on right side
		rectIcon.X += rectIcon.W - icon_w
		rectIcon.W = icon_w
		//rectLabel = rectLabel.CutRight(icon_w)
		//}

		var pt WinImagePath
		if st.Icon2Path != "" {
			pt = InitWinImagePath_file(st.Icon2Path, layout.UID)
		} else {
			pt = InitWinImagePath_blob(st.Icon2Blob, layout.UID)
		}
		rectIcon = rectIcon.Cut(st.Icon2_margin)
		paint.File(rectIcon, pt, cdText, cdText_over, cdText_down, 1, 1)
	}

	//draw label
	if st.Value != "" {
		tx := paint.Text(Rect{}, st.Value, "", cdText, cdText_over, cdText_down, false, false, uint8(st.Align), 1)
		tx.Multiline = true
		tx.Linewrapping = false
		tx.Margin = st.getAutoResizeMargin()
	}

	//draw border
	if st.Border {
		paint.RectRad(rc, cdBack, cdBack_over, cdBack_down, 0.03, layout.getRounding())
	}

	return
}

func (st *Button) Input(in LayoutInput, layout *Layout) {
	active := in.IsActive
	inside := in.IsInside && (active || !in.IsUse)
	clicked := in.IsUp && active && inside

	if st.Shortcut_key != 0 && st.Shortcut_key == in.Shortcut_key {
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
		OsOpenBrowser(st.BrowserUrl)
	}
}

func (st *Button) autoResize(layout *Layout) {
	layout.resizeFromPaintText(st.Value, true, false, false, st.getAutoResizeMargin())
}
func (st *Button) getAutoResizeMargin() [4]float64 {

	m := (1 - WinFontProps_GetDefaultLineH()) / 2
	if st.Value == "" {
		m = 0
	}

	margin := [4]float64{m, m, m, m}

	if st.IconPath != "" || len(st.IconBlob) > 0 {
		margin[2] += 1 //left
	}
	if st.Icon2Path != "" || len(st.Icon2Blob) > 0 {
		margin[3] += 1 //right
	}

	return margin
}

func (st *Button) HasShortcut(key rune) bool {
	return st.Shortcut_key == key
}

func OsOpenBrowser(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = LogsErrorf("unsupported platform")
	}
	return err
}
