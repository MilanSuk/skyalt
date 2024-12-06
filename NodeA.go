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

import "fmt"

func (st *NodeA) Build() {
	st.lock.Lock()
	defer st.lock.Unlock()

	for i := range 2 {
		b := st.layout.AddNodeB(0, st.layout.Y*10+i, 10, 1, i)
		b.done = func() {
			st.lock.Lock()
			defer st.lock.Unlock()
		}
	}
}

func (st *NodeA) Draw() {
	st.lock.Lock()
	defer st.lock.Unlock()

	st.layout.buffer = append(st.layout.buffer, LayoutDrawPrim{})

	fmt.Println("NodeA Draw()", st.layout.Y)
}
