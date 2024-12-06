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
	"fmt"
)

func (st *NodeB) Build() {
	st.lock.Lock()
	defer st.lock.Unlock()

	sum := 0
	ms := 500000000
	for i := range ms {
		st.layout.Progress = float64(i) / float64(ms)
		sum += i
	}
	fmt.Println("build sum", sum)

	if st.done != nil {
		st.done()
	}
}

func (st *NodeB) Draw() {
	st.lock.Lock()
	defer st.lock.Unlock()

	st.layout.buffer = append(st.layout.buffer, LayoutDrawPrim{})
	st.layout.buffer = append(st.layout.buffer, LayoutDrawPrim{})

	sum := 0
	ms := 500000000
	for i := range ms {
		st.layout.Progress = float64(i) / float64(ms)
		sum += i
	}
	fmt.Println("draw sum", sum)
}
