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
	"path/filepath"
	"strings"
	"sync"
)

func _read_file(name string, st any) {
	path := filepath.Join("data", name) + ".json"

	js, err := os.ReadFile(path)
	if err != nil {

		if os.IsNotExist(err) {
			//create file
			_write_file(name, st)
			return
		}

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
	path := filepath.Join("data", name) + ".json"

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

type _Struct struct {
	data interface{}
	//job  *Job
}

var g__files = make(map[string]_Struct)
var g__files_lock sync.Mutex

var g__temps = make(map[string]_Struct)
var g__temps_lock sync.Mutex

func OpenFile[T any](path string) *T {
	g__files_lock.Lock()
	defer g__files_lock.Unlock()

	tp := fmt.Sprintf("%T", *new(T))
	tp, _ = strings.CutPrefix(tp, "main.")

	if path == "" {
		path = tp + "-" + tp //default file
	}

	//find
	st, found := g__files[path]
	if found {
		stt, ok := st.data.(*T)
		if !ok {
			fmt.Printf("Runtime error: bad casting(%s) for path(%s)", tp, path)
		}
		return stt
	}

	//add
	//job := &Job{}
	stt := new(T)
	_read_file(path, stt)

	g__files[path] = _Struct{data: stt}
	return stt
}

func OpenMemory[T any](uid string, init *T) *T {
	g__temps_lock.Lock()
	defer g__temps_lock.Unlock()

	tp := fmt.Sprintf("%T", *new(T))
	tp, _ = strings.CutPrefix(tp, "main.")

	uid = tp + ":" + uid

	//find
	st, found := g__temps[uid]
	if found {
		stt, ok := st.data.(*T)
		if !ok {
			fmt.Printf("Runtime error: bad casting(%s) for path(%s)", tp, uid)
		}
		return stt
	}

	//add
	//job := &Job{}
	g__temps[uid] = _Struct{data: init}
	return init
}

type _File struct {
	Folder string
	Type   string
	Name   string
}

func (f *_File) GetPath() string {
	return filepath.Join(f.Folder, fmt.Sprintf("%s-%s", f.Type, f.Name))
}

func GetListOfFiles(folder string) (files []_File, dirs []string) {
	filesDir, err := os.ReadDir(filepath.Join("data", folder))
	if err != nil {
		return
	}
	//sort
	/*sort.Slice(filesDir, func(i, j int) bool {
		info_i, _ := filesDir[i].Info()
		info_j, _ := filesDir[j].Info()
		return info_i.ModTime().Before(info_j.ModTime())
	})*/

	for _, file := range filesDir {

		if file.IsDir() {
			dirs = append(dirs, file.Name())
		} else {
			name := file.Name()
			name = strings.TrimSuffix(name, filepath.Ext(name)) //cut .json

			d := strings.IndexByte(name, '-')
			if d > 0 {
				files = append(files, _File{Folder: folder, Type: name[:d], Name: name[d+1:]})
			}
		}
	}
	return
}

func _skyalt_save() {
	g__files_lock.Lock()
	defer g__files_lock.Unlock()

	for path, it := range g__files {
		_write_file(path, it.data)
	}
	g__files = nil
}
