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
	"os"
	"path/filepath"
	"sync"
)

type ToolsFilesItem struct {
	data []byte

	changed    bool
	time_stamp int64
}

type ToolsFiles struct {
	folder string

	files map[string]*ToolsFilesItem

	lock sync.Mutex

	saveIt bool
}

func NewToolsFiles(Folder string) *ToolsFiles {
	files := &ToolsFiles{folder: Folder}

	files.files = make(map[string]*ToolsFilesItem)

	return files
}

func (files *ToolsFiles) Destroy() {

}

func (files *ToolsFiles) GetFile(path string) (*ToolsFilesItem, []byte) {
	files.lock.Lock()
	defer files.lock.Unlock()

	file, found := files.files[path]
	if !found {
		//open
		data, err := os.ReadFile(filepath.Join(files.folder, string(path)))
		if err == nil {
			file = &ToolsFilesItem{data: data}
			files.files[string(path)] = file
			found = true
		}
	}

	if found {
		return file, file.data
	}
	return nil, nil
}

func (files *ToolsFiles) SetFile(path string, data []byte) {
	files.lock.Lock()
	defer files.lock.Unlock()

	f, found := files.files[path]
	if found {
		f.data = data
		f.changed = true
		f.time_stamp++
	} else {
		files.files[path] = &ToolsFilesItem{data: data, changed: true}
	}

	files.saveIt = true

}

func (files *ToolsFiles) Save() error {
	files.lock.Lock()
	defer files.lock.Unlock()

	//files
	for path, it := range files.files {
		if it.changed {
			os.WriteFile(filepath.Join(files.folder, path), it.data, 0644)
			it.changed = false
		}
	}

	return nil
}

func (files *ToolsFiles) Tick() {

	//delay ....
	if files.saveIt {
		files.saveIt = false
		files.Save()
	}
}
