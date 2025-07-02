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
	"bytes"
	"fmt"
	"os"
	"time"
)

func Tools_Time() float64 {
	return float64(time.Now().UnixMicro()) / 1000000 //seconds
}

func Tools_FormatBytes(bytes int) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div := int64(unit)
	exp := 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func Tools_IsFileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func Tools_GetFileTime(path string) int64 {
	inf, err := os.Stat(path)
	if err == nil && inf != nil {
		return inf.ModTime().UnixNano()
	}
	return 0
}

func Tools_WriteJSONFile(path string, st any) (bool, error) {
	// Pack into JSON
	fl, err := LogsJsonMarshal(st)
	if err != nil {
		return false, err
	}

	oldFl, _ := os.ReadFile(path)
	diff := !bytes.Equal(oldFl, fl)

	// Save into file
	if diff {
		err = os.WriteFile(path, fl, 0644)
		if LogsError(err) != nil {
			return false, err
		}
	}

	return diff, nil
}
