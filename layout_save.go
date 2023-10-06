/*
Copyright 2023 Milan Suk

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

type LayoutSaveItemResize struct {
	Name  string
	Value float32
}

type LayoutSaveItem struct {
	Hash                   uint64
	ScrollVpos, ScrollHpos int
	Cols_resize            []LayoutSaveItemResize
	Rows_resize            []LayoutSaveItemResize
}

type LayoutSaveBase struct {
	Hostname string
	Items    []*LayoutSaveItem
}

type LayoutSave struct {
	base LayoutSaveBase //this device
}

func NewRS_LScroll(js []byte) (*LayoutSave, error) {

	var save LayoutSave

	hostname, bases, found_i, err := LayoutSave_parseBases(js)
	if err != nil {
		return nil, fmt.Errorf("LayoutSave_parseBases() failed: %w", err)
	}

	if found_i < 0 {
		//new
		save.base.Hostname = hostname
	} else {
		//copy
		save.base = bases[found_i]
	}

	return &save, nil
}

func LayoutSave_parseBases(js []byte) (string, []LayoutSaveBase, int, error) {

	var bases []LayoutSaveBase
	found_i := -1

	hostname, err := os.Hostname()
	if err != nil {
		return "", nil, -1, fmt.Errorf("Hostname() failed: %w", err)
	}

	if len(js) > 0 {
		//extract
		err := json.Unmarshal(js, &bases)
		if err != nil {
			return "", nil, -1, fmt.Errorf("Unmarshal() failed: %w", err)
		}

		//find
		for i, b := range bases {
			if b.Hostname == hostname {
				found_i = i
				break
			}
		}
	}

	return hostname, bases, found_i, nil
}

func (save *LayoutSave) Save(js []byte) ([]byte, error) {

	_, bases, found_i, err := LayoutSave_parseBases(js)
	if err != nil {
		return nil, fmt.Errorf("LayoutSave_parseBases() failed: %w", err)
	}

	if found_i < 0 {
		//new
		bases = append(bases, save.base)
	} else {
		//copy
		bases[found_i] = save.base
	}

	file, err := json.MarshalIndent(&bases, "", "")
	if err != nil {
		return nil, fmt.Errorf("MarshalIndent() failed: %w", err)
	}
	return file, nil
}

func (save *LayoutSave) FindGlobalScrollHash(hash uint64) *LayoutSaveItem {
	if save == nil {
		return nil
	}

	for _, it := range save.base.Items {
		if it.Hash == hash {
			return it
		}
	}

	return nil
}

func (save *LayoutSave) AddGlobalScrollHash(hash uint64) *LayoutSaveItem {
	if save == nil {
		return nil
	}

	sc := save.FindGlobalScrollHash(hash)
	if sc != nil {
		return sc
	}

	nw := &LayoutSaveItem{Hash: hash}
	save.base.Items = append(save.base.Items, nw)
	return nw
}
