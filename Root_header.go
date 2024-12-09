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

import "sync"

type Root struct {
	layout *Layout
	lock   sync.Mutex
}

func (layout *Layout) AddRoot(x, y, w, h int, props *Root) *Root {
	props.layout = layout._createDiv(x, y, w, h, "Root", props.Build, nil, nil)
	return props
}

var g_Root *Root

func NewFile_Root() *Root {
	if g_Root == nil {
		g_Root = &Root{}
		_read_file("Root-Root", g_Root)
	}
	return g_Root
}
