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
	stack     string
	err       error
	time_unix int64
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

func (lg *ToolsLog) Error(err error) error {
	if err != nil {
		lg.lock.Lock()
		defer lg.lock.Unlock()

		stack := string(debug.Stack())

		lg.errors = append(lg.errors, ToolsLogItem{stack: stack, err: err, time_unix: time.Now().Unix()})

		fmt.Printf("\033[31m%s error: %v\nstack:%s\033[0m\n", lg.Name, err, stack)
	}
	return err
}
func (lg *ToolsLog) GetList(oldest_time_unix int64) []ToolsLogItem {
	lg.lock.Lock()
	defer lg.lock.Unlock()

	start_i := -1
	for i := range lg.errors {
		if lg.errors[i].time_unix > oldest_time_unix {
			start_i = i
			break
		}
	}

	if start_i < 0 {
		return nil
	}
	return lg.errors[start_i:]
}
