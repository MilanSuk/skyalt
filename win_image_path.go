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
	"crypto/sha256"
	"encoding/hex"
)

type WinImagePath struct {
	uid      string
	playerID uint64

	path string
	blob []byte

	play_pos      uint64
	play_duration uint64

	is_playing bool

	tp int //type(0=image, 1=video, 2=audio)
}

func InitWinImagePath_file(path string, playerID uint64) WinImagePath {
	return WinImagePath{uid: "file:" + path, path: path, playerID: playerID}
}
func InitWinImagePath_blob(blob []byte, playerID uint64) WinImagePath {
	hash := sha256.Sum256(blob)
	return WinImagePath{uid: "blob: " + hex.EncodeToString(hash[:]), blob: blob, playerID: playerID}
}

func (a *WinImagePath) Cmp(b *WinImagePath) bool {
	return a.uid == b.uid && a.playerID == b.playerID
}
