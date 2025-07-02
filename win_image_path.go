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
	"os"
	"path/filepath"
	"strings"
)

type WinImagePath struct {
	uid            string
	fnGetBlob      func(fnDone func(bytes []byte, err error)) error
	fnGetTimestamp func() (int64, error)

	service_path string
}

func InitWinImagePath_blob(blob []byte) WinImagePath {
	hash := sha256.Sum256(blob)

	uid := "hash: " + hex.EncodeToString(hash[:])
	fnGetBlob := func(fnDone func(bytes []byte, err error)) error {
		fnDone(blob, nil)
		return nil
	}
	return WinImagePath{uid: uid, fnGetBlob: fnGetBlob}
}
func InitWinImagePath_file(path string) WinImagePath {
	uid := path

	var fnGetBlob func(fnDone func(bytes []byte, err error)) error
	var service_path string
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".mp4" || ext == ".mkv" || ext == ".webm" || ext == ".mov" || ext == ".avi" || ext == ".flv" ||
		ext == ".wav" || ext == ".mp3" || ext == ".opus" || ext == ".aac" || ext == ".ogg" || ext == ".flac" || ext == ".pcm" {
		//audio/video
		service_path = path
	} else {
		//images
		fnGetBlob = func(fnDone func(bytes []byte, err error)) error {
			go func() {
				fnDone(os.ReadFile(path))
			}()
			return nil
		}
	}

	fnGetTimestamp := func() (int64, error) {
		inf, err := os.Stat(path)
		if err == nil && inf != nil {
			return inf.ModTime().UnixNano(), nil //ok
		}
		return -1, err
	}
	return WinImagePath{uid: uid, fnGetBlob: fnGetBlob, fnGetTimestamp: fnGetTimestamp, service_path: service_path}
}
func InitWinImagePath_load(load_uid string, fnGetBlob func(fnDone func(bytes []byte, err error)) error) WinImagePath {
	uid := "func: " + load_uid
	return WinImagePath{uid: uid, fnGetBlob: fnGetBlob}
}

func (ip *WinImagePath) GetString() string {
	return ip.uid
}

func (a *WinImagePath) Cmp(b *WinImagePath) bool {
	return a.uid == b.uid
}
