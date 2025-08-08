package main

import (
	"image/color"
	"strings"
)

type Search struct {
	Cd    color.RGBA
	Value *string
	Ghost string

	enter func()
}

func (layout *Layout) AddSearch(x, y, w, h int, value *string, ghost string) *Search {
	if ghost == "" {
		ghost = "Search"
	}
	props := &Search{Value: value, Ghost: ghost, Cd: layout.GetPalette().OnB}
	layout._createDiv(x, y, w, h, "Search", props.Build, props.Draw, nil)
	return props
}

func (st *Search) Build(layout *Layout) {
	layout.SetColumn(0, 1, 1)
	layout.SetColumn(1, 1, Layout_MAX_SIZE)
	layout.SetColumn(2, 1, 0)

	Img := layout.AddImage(0, 0, 1, 1, "resources/search.png", nil)
	Img.Cd = color.RGBA{0, 0, 0, 255}
	Img.Margin = 0.2

	ed := layout.AddEditbox(1, 0, 1, 1, st.Value)
	ed.Ghost = st.Ghost
	ed.Refresh = true
	ed.enter = func() {
		if st.enter != nil {
			st.enter()
		}
	}

	bt := layout.AddButtonMenu(2, 0, 1, 1, "⌫", "", 0)
	bt.Align = 1
	bt.clicked = func() {
		if st.Value != nil {
			*st.Value = ""
		}
	}
}

func (st *Search) Draw(rect Rect, layout *Layout) (paint LayoutPaint) {
	if st.Value != nil && *st.Value != "" {
		backCd := Color_Aprox(layout.GetPalette().P, layout.GetPalette().B, 0.5)

		rc := rect.CutLeft(1)
		rc = rc.CutRight(1)
		rc = rc.Cut(0.06)
		paint.RectRad(rc, backCd, backCd, backCd, 0, layout.getRounding())
	}
	return
}

func Search_Prepare(search string) []string {
	search = strings.ToLower(search)

	search = strings.ReplaceAll(search, "\n", " ")
	search = strings.ReplaceAll(search, "\t", " ")
	search = strings.ReplaceAll(search, ";", " ")
	search = strings.ReplaceAll(search, ",", " ")
	search = strings.ReplaceAll(search, ".", " ")
	search = strings.ReplaceAll(search, "?", " ")
	search = strings.ReplaceAll(search, "-", " ")

	//split
	words := strings.Split(search, " ")

	//remove empty items
	n := 0
	for i, s := range words {
		if s != "" {
			words[n] = words[i]
			n++
		}
	}
	words = words[:n]

	return words
}
func Search_Find(str string, words []string) bool {
	if len(words) == 0 {
		return true
	}

	str = strings.ToLower(str)

	for _, w := range words {
		if !strings.Contains(str, w) {
			return false
		}
	}
	return true
}
