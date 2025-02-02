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
	"sync"
)

var g_cmds []LayoutCmd
var g_cmds_lock sync.Mutex

func _addCmd(cmd LayoutCmd) {
	g_cmds_lock.Lock()
	defer g_cmds_lock.Unlock()

	g_cmds = append(g_cmds, cmd)
}
func _getCmds() []LayoutCmd {
	if g__jobs.NeedRefresh() {
		Layout_RefreshDelayed() //calls _addCmd() which has lock() inside
	}

	g_cmds_lock.Lock()
	defer g_cmds_lock.Unlock()

	cmds := g_cmds
	g_cmds = nil
	return cmds
}

func _build(layout *Layout) {
	if layout.fnBuild != nil {
		layout.fnBuild(layout)
	}
	for _, it := range layout.Childs {
		_build(it)
	}
	for _, it := range layout.dialogs {
		_build(it.Layout)
	}
}

func _draw(layout *Layout, rects map[uint64]Rect, out_buffs map[uint64][]LayoutDrawPrim) {
	canvas, found := rects[layout.Hash]
	if found {
		if layout.fnDraw != nil && canvas.Is() {
			//fmt.Println("fnDraw", layout.Name, canvas.H)
			paint := layout.fnDraw(canvas, layout)
			if len(paint.buffer) > 0 {
				out_buffs[layout.Hash] = paint.buffer
			}
		}
	}

	for _, it := range layout.Childs {
		_draw(it, rects, out_buffs)
	}
}
