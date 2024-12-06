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
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
)

func OsFormatBytes(bytes int) string {
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

func OsMarshal(v interface{}) []byte {
	//data, err := json.Marshal(v)
	//return data, err

	b := new(bytes.Buffer)
	err := gob.NewEncoder(b).Encode(v)
	if err != nil {
		log.Fatal("NewEncoder failed:", err.Error())
	}
	return b.Bytes()
}

func OsUnmarshal(data []byte, v interface{}) {
	//err := json.Unmarshal(data, v)
	//return err

	b := bytes.NewBuffer(data)
	err := gob.NewDecoder(b).Decode(v)
	if err != nil {
		log.Fatal("NewDecoder failed:", err.Error())
	}
}
