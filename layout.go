/*
Copyright 2024 Milan Suk

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
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"image/color"
	"log"
	"path/filepath"
	"runtime"
)

type Rect struct {
	X, Y, W, H float64
}

func (r Rect) Is() bool {
	return r.W > 0 && r.H > 0
}

type LayoutDrawPrim struct {
	Type uint8

	Rect           Rect
	Sx, Sy, Ex, Ey float64

	Cd, Cd_over, Cd_down color.RGBA

	Border float64

	Text  string
	Text2 string

	Boolean bool

	Align_h uint8
	Align_v uint8

	Text_formating    bool
	Text_multiline    bool
	Text_linewrapping bool
	Text_selection    bool
	Text_editable     bool
}

type Layout struct {
	X, Y, W, H int
	Name       string

	Childs []*Layout
	Hash   uint64

	Caller_file string `json:",omitempty"`
	Caller_line int    `json:",omitempty"`

	fnBuild func()
	fnDraw  func()

	Canvas Rect
	buffer []LayoutDrawPrim

	Progress float64
	updated  bool
	redraw   bool
	done     bool
}

func (layout *Layout) _getName() string {
	return fmt.Sprintf("%s(%d,%d,%d,%d)", layout.Name, layout.X, layout.Y, layout.W, layout.H)
}
func (layout *Layout) _computeHash(parent *Layout) uint64 {
	if parent == nil {
		return 0
	}

	h := sha256.New()

	//parent
	var pt [8]byte
	binary.LittleEndian.PutUint64(pt[:], parent.Hash)
	h.Write(pt[:])

	//this
	h.Write([]byte(layout._getName()))

	return binary.LittleEndian.Uint64(h.Sum(nil))
}

func (layout *Layout) _findChild(x, y, w, h int, name string) *Layout {
	for _, it := range layout.Childs {
		if it.X == x && it.Y == y && it.W == w && it.H == h && it.Name == name {
			return it
		}
	}
	return nil
}

func _newLayout(x, y, w, h int, name string, parent *Layout) *Layout {
	layout := &Layout{X: x, Y: y, W: w, H: h, Name: name} //, canvas_drawn: Rect{0, 0, -1, -1}}

	layout.Hash = layout._computeHash(parent)

	return layout
}

func (layout *Layout) _createDiv(x, y, w, h int, name string, fnBuild func(), fnDraw func()) *Layout {

	lay := layout._findChild(x, y, w, h, name)
	if lay == nil {
		lay = _newLayout(x, y, w, h, name, layout)
		layout.Childs = append(layout.Childs, lay)
	}

	lay.fnBuild = fnBuild
	lay.fnDraw = fnDraw

	var ok bool
	_, lay.Caller_file, lay.Caller_line, ok = runtime.Caller(2)
	if !ok {
		log.Fatal("runtime.Caller failed")
	}
	lay.Caller_file = filepath.Base(lay.Caller_file)

	return lay
}

func (layout *Layout) _findHash(hash uint64) *Layout {
	if layout.Hash == hash {
		return layout
	}

	for _, it := range layout.Childs {
		d := it._findHash(hash)
		if d != nil {
			return d
		}
	}
	return nil
}
