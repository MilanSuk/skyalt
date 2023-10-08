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

type VmTextHistoryItem struct {
	str string
	cur int
}

type VmTextHistory struct {
	uid *LayoutDiv

	items []VmTextHistoryItem

	act          int
	lastAddTicks int
}

func NewVmTextHistoryItem(uid *LayoutDiv, init VmTextHistoryItem) *VmTextHistory {
	var his VmTextHistory
	his.uid = uid

	his.items = append(his.items, init)
	his.lastAddTicks = OsTicks()

	return &his
}

func (his *VmTextHistory) Add(value VmTextHistoryItem) bool {

	//same as previous
	if his.items[his.act].str == value.str {
		return false
	}

	//cut all after
	his.items = his.items[:his.act+1]

	//adds new snapshot
	his.items = append(his.items, value)
	his.act++
	his.lastAddTicks = OsTicks()

	return true
}
func (his *VmTextHistory) AddWithTimeOut(value VmTextHistoryItem) bool {
	if !OsIsTicksIn(his.lastAddTicks, 1000) {
		return his.Add(value)
	}
	return false
}

func (his *VmTextHistory) Backward(init VmTextHistoryItem) VmTextHistoryItem {

	his.Add(init)

	if his.act-1 >= 0 {
		his.act--
	}
	return his.items[his.act]
}
func (his *VmTextHistory) Forward() VmTextHistoryItem {
	if his.act+1 < len(his.items) {
		his.act++
	}
	return his.items[his.act]
}

type VmTextHistoryArray struct {
	items []*VmTextHistory
}

func (his *VmTextHistoryArray) FindOrAdd(uid *LayoutDiv, init VmTextHistoryItem) *VmTextHistory {

	//finds
	for _, it := range his.items {
		if it.uid == uid {
			return it
		}
	}

	//adds
	it := NewVmTextHistoryItem(uid, init)
	his.items = append(his.items, it)
	return it
}
