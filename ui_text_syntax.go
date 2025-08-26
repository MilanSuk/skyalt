/*
Copyright 2025 Milan Suk

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
	"cmp"
	"fmt"
	"image/color"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

var (
	g_syntax_basic  = color.RGBA{50, 50, 150, 255} // blue
	g_syntax_struct = color.RGBA{150, 150, 50, 255}
	g_syntax_func   = color.RGBA{50, 150, 50, 255} //green

	g_syntax_comment     = color.RGBA{50, 150, 150, 255}
	g_syntax_stringConst = color.RGBA{150, 50, 50, 255} //red
)

func isStdType(name string) bool {
	var g_syntax_stds = []string{"string",
		"float32", "float64",
		"byte", "rune", "bool",
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"error",
	}
	return slices.Contains(g_syntax_stds, name)
}
func isStdKeyword(name string) bool {
	var g_syntax_stds = []string{"type", "struct",
		"for", "if", "else", "return", "break", "continue", "range",
		"switch", "case", "default",
		"var", "const",
		"map", "chan",
		"func", "nil",
		"import", "package",
	}
	return slices.Contains(g_syntax_stds, name)
}

type UiTextSyntax struct {
	Start int
	End   int
	Color color.RGBA

	Text string

	Ignore bool
}

func InitSyntaxItem(str string, start int, end int, cd color.RGBA, ignore bool) UiTextSyntax {
	return UiTextSyntax{Start: start, End: end, Text: str[start:end], Color: cd, Ignore: ignore}
}

func (e *UiTextSyntax) Replace(code string) string {
	st := e.Start
	en := e.End
	return code[:st] + fmt.Sprintf("<rgba%d,%d,%d,255>%s</rgba>", e.Color.R, e.Color.G, e.Color.B, code[st:en]) + code[en:]
}

func _UiText_FormatAsCode(code string) string {
	outliers := _UiText_findOutliers(code)

	var finalElem []UiTextSyntax
	for _, it := range outliers {
		if it.Ignore {
			continue
		}
		finalElem = append(finalElem, it)
	}

	elms := _UiText_getWords(code)

	//Ignore elements inside comment or string
	for i, it := range elms {
		for _, it2 := range outliers {
			if (it.Start >= it2.Start && it.Start < it2.End) || (it.End >= it2.Start && it.End < it2.End) {
				elms[i].Ignore = true
				break
			}
		}
	}

	//Create structs list
	var structs []string
	for _, it := range elms {
		if it.Ignore {
			continue
		}
		if it.End+1 < len(code) && strings.HasPrefix(code[it.End:], " struct") {
			structs = append(structs, it.Text)
			break
		}
	}

	//Std
	for _, it := range elms {
		if it.Ignore {
			continue
		}

		//Standard
		if isStdType(it.Text) || isStdKeyword(it.Text) {
			it.Color = g_syntax_basic
			finalElem = append(finalElem, it)
		}

		if _UiText_isNumber(it.Text) {
			it.Color = g_syntax_stringConst
			finalElem = append(finalElem, it)
		}

		//Function
		if it.End+1 < len(code) && code[it.End] == '(' {
			it.Color = g_syntax_func
			finalElem = append(finalElem, it)
		}

		//Struct
		{
			if slices.Contains(structs, it.Text) {
				it.Color = g_syntax_struct
				finalElem = append(finalElem, it)
			}
		}
	}

	//Add marks into code
	slices.SortFunc(finalElem, func(a, b UiTextSyntax) int { return cmp.Compare(a.Start, b.Start) })
	finalElem = slices.CompactFunc(finalElem, func(a, b UiTextSyntax) bool { return a.Start == b.Start || a.End == b.End })
	slices.Reverse(finalElem)
	for _, e := range finalElem {
		code = e.Replace(code)
	}

	return code
}

func _UiText_isNumber(str string) bool {
	_, err := strconv.ParseFloat(str, 64)
	return err == nil
}

func _UiText_getWords(input string) []UiTextSyntax {
	var words []UiTextSyntax

	// Handle empty string
	if len(input) == 0 {
		return words
	}

	start := -1
	var wordBuilder strings.Builder

	for i, ch := range input {
		// Check if character is alphabetic or numeric
		if unicode.IsLetter(ch) || unicode.IsNumber(ch) || ch == '_' || (ch == '.' && start >= 0 && _UiText_isNumber(input[start:i])) {
			// Start a new word if not already started
			if start == -1 {
				start = i
			}
			wordBuilder.WriteRune(ch)
		} else {
			// If we were building a word, save it
			if start != -1 {
				wordText := wordBuilder.String()
				if wordText != "" {
					words = append(words, InitSyntaxItem(input, start, i, color.RGBA{}, false))
				}
				// Reset for next word
				start = -1
				wordBuilder.Reset()
			}
		}
	}

	// Handle case where string ends with a word
	if start != -1 {
		wordText := wordBuilder.String()
		if wordText != "" {
			words = append(words, InitSyntaxItem(input, start, len(input), color.RGBA{}, false))
		}
	}

	return words
}

func _UiText_processOutlierPair(s string, i int, prefix, postfix string, cd color.RGBA, ignore bool) (bool, int, []UiTextSyntax) {

	if i <= len(s)-len(prefix) && s[i:i+len(prefix)] == prefix {
		var items []UiTextSyntax

		start := i
		i += len(prefix)
		foundEnd := false
		for i < len(s) {
			if i <= len(s)-len(postfix) && s[i:i+len(postfix)] == postfix {
				items = append(items, InitSyntaxItem(s, start, i+1, cd, ignore))
				i += len(postfix)
				foundEnd = true
				break
			} else if s[i] == '\n' {
				items = append(items, InitSyntaxItem(s, start, i-1, cd, ignore))
				start = i + 1
			}
			i++
		}
		if !foundEnd {
			items = append(items, InitSyntaxItem(s, start, len(s)-1, cd, ignore))
		}

		return true, i, items
	}

	return false, i, nil

}

func _UiText_findOutliers(s string) []UiTextSyntax {
	var items []UiTextSyntax
	i := 0
	for i < len(s) {
		// Check for single-line comment
		if i < len(s)-1 && s[i:i+2] == "//" {
			j := i + 2
			for j < len(s) && s[j] != '\n' {
				j++
			}
			end := j - 1
			if j == len(s) {
				end = len(s) - 1
			}
			items = append(items, InitSyntaxItem(s, i, end+1, g_syntax_comment, false))
			i = j
			continue
		}

		var found bool
		var addItems []UiTextSyntax

		found, i, addItems = _UiText_processOutlierPair(s, i, "/*", "*/", g_syntax_comment, false)
		if found {
			items = append(items, addItems...)
			continue
		}

		found, i, addItems = _UiText_processOutlierPair(s, i, "\"", "\"", g_syntax_stringConst, false)
		if found {
			items = append(items, addItems...)
			continue
		}

		found, i, addItems = _UiText_processOutlierPair(s, i, "<rgba", "</rgba>", g_syntax_stringConst, true) //ignore=true -> don't highlight syntax inside(probably red error)
		if found {
			items = append(items, addItems...)
			continue
		}

		i++
	}
	return items
}
