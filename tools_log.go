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
	"fmt"
	"runtime/debug"
	"sync"
	"time"
)

type ToolsLogItem struct {
	Stack string
	Msg   string
	Time  float64
}

type ToolsLog struct {
	Name   string
	lock   sync.Mutex
	errors []ToolsLogItem
}

func (lg *ToolsLog) Clear() {
	lg.lock.Lock()
	defer lg.lock.Unlock()
	lg.errors = nil
}

func (lg *ToolsLog) Get(start_i int) []ToolsLogItem {
	lg.lock.Lock()
	defer lg.lock.Unlock()

	if start_i < 0 || start_i >= len(lg.errors) {
		return nil
	}
	return lg.errors[start_i:]
}

func (lg *ToolsLog) Error(err error) error {
	if err != nil {
		lg.lock.Lock()
		defer lg.lock.Unlock()

		stack := string(debug.Stack())

		lg.errors = append(lg.errors, ToolsLogItem{Stack: stack, Msg: err.Error(), Time: float64(time.Now().UnixMicro()) / 1000000})

		fmt.Printf("\033[31m%s error: %v\nstack:%s\033[0m\n", lg.Name, err, stack)
	}
	return err
}
