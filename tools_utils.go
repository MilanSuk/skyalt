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
	"encoding/json"
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

func Tools_FileExists(fileName string) bool {
	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func Tools_WriteJSONFile(path string, st interface{}) (bool, error) {
	// Pack into JSON
	fl, err := json.Marshal(st)
	if err != nil {
		return false, fmt.Errorf("Marshal() failed: %w", err)
	}

	oldFl, _ := os.ReadFile(path)
	diff := !bytes.Equal(oldFl, fl)

	// Save into file
	if diff {
		err = os.WriteFile(path, fl, 0644)
		if err != nil {
			return false, fmt.Errorf("WriteFile() failed: %vw", err)
		}
	}

	return diff, nil
}
