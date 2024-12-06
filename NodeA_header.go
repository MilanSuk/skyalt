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

type NodeA struct {
	layout *Layout
	lock   sync.Mutex

	Id int
}

func (layout *Layout) AddNodeA(x, y, w, h int, id int) *NodeA {
	props := &NodeA{Id: id}
	props.layout = layout._createDiv(x, y, w, h, "NodeA", props.Build, props.Draw)
	return props
}
