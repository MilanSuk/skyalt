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

type UiTextHistoryItem struct {
	str string
	cur int
}

type UiTextHistory struct {
	uid uint64

	items []UiTextHistoryItem

	act          int
	lastAddTicks int64
}

func NewUiTextHistoryItem(uid uint64, init UiTextHistoryItem) *UiTextHistory {
	var his UiTextHistory
	his.uid = uid

	his.items = append(his.items, init)
	his.lastAddTicks = OsTicks()

	return &his
}

func (his *UiTextHistory) Add(value UiTextHistoryItem) bool {

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
func (his *UiTextHistory) AddWithTimeOut(value UiTextHistoryItem) bool {
	if !OsIsTicksIn(his.lastAddTicks, 500) {
		return his.Add(value)
	}
	return false
}

func (his *UiTextHistory) Backward(init UiTextHistoryItem) UiTextHistoryItem {

	his.Add(init)

	if his.act-1 >= 0 {
		his.act--
	}
	return his.items[his.act]
}
func (his *UiTextHistory) Forward() UiTextHistoryItem {
	if his.act+1 < len(his.items) {
		his.act++
	}
	return his.items[his.act]
}

type UiTextHistoryArray struct {
	items []*UiTextHistory
}

func (his *UiTextHistoryArray) FindOrAdd(uid uint64, init UiTextHistoryItem) *UiTextHistory {

	//finds
	for _, it := range his.items {
		if it.uid == uid {
			return it
		}
	}

	//adds
	it := NewUiTextHistoryItem(uid, init)
	his.items = append(his.items, it)
	return it
}
