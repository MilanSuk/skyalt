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
	"encoding/json"
	"fmt"
	"os"
)

func _read_file(name string, st any) {
	path := "data/" + name + ".json"

	js, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("warning: ReadFile(): ", err)
		return
	}

	err = json.Unmarshal(js, st)
	if err != nil {
		fmt.Println("warning: Unmarshal(): ", err)
		return
	}

	fmt.Println("File open:", path)
}
func _write_file(name string, st any) {
	path := "data/" + name + ".json"

	js, err := json.MarshalIndent(st, "", "")
	if err != nil {
		fmt.Println("warning: MarshalIndent(): ", err)
	}

	err = os.WriteFile(path, js, 0644)
	if err != nil {
		fmt.Println("warning: WriteFile(): ", err)
	}

	fmt.Println("File saved:", path)
}
