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
	"fmt"
	"os"
)

type WinImagePath struct {
	uid       string
	fnGetBlob func(fnDone func(bytes []byte, err error)) error
}

func InitWinImagePath_blob(blob []byte) WinImagePath {
	hash := sha256.Sum256(blob)

	uid := "blob: " + hex.EncodeToString(hash[:])
	fnGetBlob := func(fnDone func(bytes []byte, err error)) error {
		fnDone(blob, nil)
		return nil
	}
	return WinImagePath{uid: uid, fnGetBlob: fnGetBlob}
}
func InitWinImagePath_file(path string) WinImagePath {
	uid := "file: " + path
	fnGetBlob := func(fnDone func(bytes []byte, err error)) error {
		go func() {
			fnDone(os.ReadFile(path))
		}()
		return nil
	}
	return WinImagePath{uid: uid, fnGetBlob: fnGetBlob}
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

func (ip *WinImagePath) GetBlob(fnDone func(bytes []byte, err error)) error {
	if ip.fnGetBlob != nil {
		return ip.fnGetBlob(fnDone)
	}

	return fmt.Errorf("fnGetBlob() is nil")
}
