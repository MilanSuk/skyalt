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
	"time"
)

type Process struct {
	uid    string
	fnFree func()

	maxSecUnused float64

	last_use float64
}

func (proc *Process) IsOutOfTime(tm float64) bool {
	return proc.last_use+proc.maxSecUnused < tm
}

type Processes struct {
	items []*Process
	lock  sync.Mutex
}

var g__processes Processes

func AddProcess(uid string, maxSecUnused float64, fnFree func()) *Process {
	g__processes.lock.Lock()
	defer g__processes.lock.Unlock()

	tm := float64(time.Now().UnixMilli()) / 1000

	//find
	for _, it := range g__processes.items {
		if it.uid == uid {
			it.last_use = tm
			return it
		}
	}

	//add
	proc := &Process{uid: uid, maxSecUnused: maxSecUnused, fnFree: fnFree, last_use: tm}
	g__processes.items = append(g__processes.items, proc)

	return proc
}

func (st *Processes) Destroy() {
	st.lock.Lock()
	defer st.lock.Unlock()

	for _, it := range g__processes.items {
		if it.fnFree != nil {
			it.fnFree()
		}
	}
}

func (st *Processes) maintenance() {
	st.lock.Lock()
	defer st.lock.Unlock()

	tm := float64(time.Now().UnixMilli()) / 1000

	//find
	for i := len(st.items) - 1; i >= 0; i-- {
		it := st.items[i]
		if it.IsOutOfTime(tm) {
			if it.fnFree != nil {
				it.fnFree()
			}
			st.items = append(st.items[:i], st.items[i+1:]...) //remove
		}
	}
}

func (layout *Layout) AddProcesses(x, y, w, h int) {
	layout._createDiv(x, y, w, h, "Processes", g__processes.Build, nil, nil)
}

func (st *Processes) Build(layout *Layout) {
	st.lock.Lock()
	defer st.lock.Unlock()

	layout.SetColumn(0, 1, 100)
	layout.SetColumn(1, 1, 4)

	for i, it := range st.items {
		layout.AddText(0, i, 1, 1, it.uid)
		bt := layout.AddButtonDanger(1, i, 1, 1, "Stop")
		bt.clicked = func() {
			st.items = append(st.items[:i], st.items[i+1:]...) //remove
		}
	}
}
