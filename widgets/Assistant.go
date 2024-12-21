package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"image/color"
	"os"
	"path/filepath"
	"strings"
)

type Assistant struct {
	Show bool

	Prompt string
	Picks  []LayoutPick

	TempPrompt AssistantPrompt

	Chat OpenAI_chat

	AutoSend float64
}

func (layout *Layout) AddAssistant(x, y, w, h int, props *Assistant) *Assistant {
	layout._createDiv(x, y, w, h, "Assistant", props.Build, nil, nil)
	return props
}

var g_Assistant *Assistant

func NewFile_Assistant() *Assistant {
	if g_Assistant == nil {
		g_Assistant = &Assistant{}
		_read_file("Assistant-Assistant", g_Assistant)
	}
	return g_Assistant
}

func (st *Assistant) Build(layout *Layout) {

	//...
}

func (st *Assistant) SetVoice(js []byte, voiceStart_sec float64) {
	type VerboseJsonWord struct {
		Start, End float64
		Word       string
	}
	type VerboseJsonSegment struct {
		Start, End float64
		Text       string
		Words      []VerboseJsonWord
	}
	type VerboseJson struct {
		Segments []VerboseJsonSegment
	}

	var verb VerboseJson
	err := json.Unmarshal(js, &verb)
	if err != nil {
		return
	}

	// jump over older picks
	pick_i := 0
	for _, pk := range st.Picks {
		if pk.Time_sec < voiceStart_sec {
			pick_i++
		}
	}
	// build prompt
	prompt := ""
	for _, seg := range verb.Segments {
		for _, w := range seg.Words {
			for pick_i < len(st.Picks) && st.Picks[pick_i].Time_sec < (voiceStart_sec+w.Start) { //for(!)
				prompt += LayoutPick_getMark(pick_i)
				pick_i++
			}
			prompt += w.Word
		}
	}

	st.Prompt = prompt
}

func (st *Assistant) Send(dialog *Layout) {
	//...
}

func _Assistant_RemoveFormatingRGBA(str string) string {
	str = strings.ReplaceAll(str, "</rgba>", "")

	for {
		st := strings.Index(str, "<rgba")
		if st < 0 {
			break
		}
		en := strings.IndexByte(str[st:], '>')
		if en >= 0 {
			str = str[:st] + str[st+en+1:]
		}
	}

	return str
}

func (ast *Assistant) findPickOrAdd(item LayoutPick) bool {
	for _, it := range ast.Picks {
		if it.Cmp(&item) {
			return true //found
		}
	}

	//add
	ast.Picks = append(ast.Picks, item)
	ast.Prompt += LayoutPick_getMark(len(ast.Picks) - 1)

	return false //new
}

func Assistant_recomputeColors(prompt string, picks []LayoutPick) string {
	var hsl HSL
	hsl.S = 0.8
	hsl.L = 0.5

	n := len(picks)
	hAdd := 360 / (n + 1)
	for pki := range picks {

		if strings.Contains(prompt, LayoutPick_getMark(pki)) {
			picks[pki].Cd = hsl.ToRGB()
			hsl.H += hAdd
		} else {
			picks[pki].Cd.A = 0 //invisible
		}

	}

	prompt = _Assistant_RemoveFormatingRGBA(prompt)
	for pki := range picks {
		mark := LayoutPick_getMark(pki)

		start := len(prompt)
		for start >= 0 {
			start = strings.LastIndex(prompt[:start], mark)
			if start >= 0 {
				end := start + len(mark)

				cd := picks[pki].Cd
				midStr := fmt.Sprintf("<rgba%d,%d,%d,%d>", cd.R, cd.G, cd.B, cd.A) + prompt[start:end] + "</rgba>"
				prompt = prompt[:start] + midStr + prompt[end:]
				//pos = start + len(midStr)
			}
		}
	}

	return prompt
}

type HSL struct {
	H int     //0-360
	S float64 //0-1
	L float64 //0-1
}

func _hueToRGB(v1, v2, vH float64) float64 {
	if vH < 0 {
		vH++
	}
	if vH > 1 {
		vH--
	}

	if (6 * vH) < 1 {
		return v1 + (v2-v1)*6*vH
	} else if (2 * vH) < 1 {
		return v2
	} else if (3 * vH) < 2 {
		return v1 + (v2-v1)*((2.0/3)-vH)*6
	}

	return v1
}

func (hsl HSL) ToRGB() color.RGBA {
	cd := color.RGBA{A: 255}

	if hsl.S == 0 {
		ll := hsl.L * 255
		if ll < 0 {
			ll = 0
		}
		if ll > 255 {
			ll = 255
		}
		cd.R = uint8(ll)
		cd.G = uint8(ll)
		cd.B = uint8(ll)
	} else {
		var v2 float64
		if hsl.L < 0.5 {
			v2 = (hsl.L * (1 + hsl.S))
		} else {
			v2 = ((hsl.L + hsl.S) - (hsl.L * hsl.S))
		}
		v1 := 2*hsl.L - v2

		hue := float64(hsl.H) / 360
		cd.R = uint8(255 * _hueToRGB(v1, v2, hue+(1.0/3)))
		cd.G = uint8(255 * _hueToRGB(v1, v2, hue))
		cd.B = uint8(255 * _hueToRGB(v1, v2, hue-(1.0/3)))
	}

	return cd
}

func _Assistant_GetGridBuildFuncPosition(layout *Layout) (string, int) {

	Caller_file := layout.Caller_file
	Caller_line := layout.Caller_line

	//'best_layout' is Caller position of Add<name>(). But we need position of Build()!
	if layout.Name != "_layout" {
		path := filepath.Join("widgets", layout.Name+".go")
		fileCode, err := os.ReadFile(path)
		if err == nil {
			build_pos, _, _, _, err := _Assistant_findBuildFunction(path, string(fileCode), layout.Name)
			if err == nil && build_pos >= 0 {
				//rewrite Caller
				Caller_file = layout.Name + ".go"
				Caller_line = strings.Count(string(fileCode[:build_pos]), "\n") + 1
			}
		} else {
			fmt.Println("File not found", err)
		}
	}

	return Caller_file, Caller_line
}
func _Assistant_findBuildFunction(ghostPath string, code string, stName string) (int, int, int, int, error) {
	node, err := parser.ParseFile(token.NewFileSet(), ghostPath, code, parser.ParseComments)
	if err != nil {
		return -1, -1, -1, -1, err
	}

	build_pos := -1
	touch_pos := -1
	draw_pos := -1
	shortcut_pos := -1
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:

			tp := ""
			if x.Recv != nil && len(x.Recv.List) > 0 {
				tp = string(code[x.Recv.List[0].Type.Pos()-1 : x.Recv.List[0].Type.End()-1])
			}

			//function
			if tp == "*"+stName {
				if x.Name.Name == "Build" {
					build_pos = int(x.Pos())
				}
				if x.Name.Name == "Touch" {
					touch_pos = int(x.Pos())
				}
				if x.Name.Name == "Draw" {
					draw_pos = int(x.Pos())
				}
				if x.Name.Name == "Shortcut" {
					shortcut_pos = int(x.Pos())
				}
			}
		}
		return true
	})

	return build_pos, touch_pos, draw_pos, shortcut_pos, nil
}
