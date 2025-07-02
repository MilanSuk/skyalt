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
	"encoding/json"
	"fmt"
	"runtime/debug"
	"sync"
	"time"
)

type LogsItem struct {
	Stack string
	Msg   string
	Time  float64
}

type Logs struct {
	lock   sync.Mutex
	errors []LogsItem
}

func (lg *Logs) Clear() {
	lg.lock.Lock()
	defer lg.lock.Unlock()
	lg.errors = nil
}

func (lg *Logs) Get(start_i int) []LogsItem {
	lg.lock.Lock()
	defer lg.lock.Unlock()

	if start_i < 0 || start_i >= len(lg.errors) {
		return nil
	}
	return lg.errors[start_i:]
}

func (lg *Logs) Error(err error) error {
	if err != nil {
		lg.lock.Lock()
		defer lg.lock.Unlock()

		stack := string(debug.Stack())

		lg.errors = append(lg.errors, LogsItem{Stack: stack, Msg: err.Error(), Time: float64(time.Now().UnixMicro()) / 1000000})

		fmt.Printf("\033[31m error: %v\nstack:%s\033[0m\n", err, stack)
	}
	return err
}

var g_logs Logs

func LogsError(err error) error {
	return g_logs.Error(err)
}

func LogsErrorf(format string, a ...any) error {
	return LogsError(fmt.Errorf(format, a...))
}

func LogsJsonMarshal(v any) ([]byte, error) {
	data, err := json.Marshal(v)
	if LogsError(err) != nil {
		return nil, err
	}
	return data, err
}
func LogsJsonMarshalIndent(v any) ([]byte, error) {
	data, err := json.MarshalIndent(v, "", " ")
	if LogsError(err) != nil {
		return nil, err
	}
	return data, err
}

func LogsJsonUnmarshal(data []byte, v any) error {
	err := json.Unmarshal(data, v)
	if LogsError(err) != nil {
		return err
	}
	return err
}
