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
	"errors"
	"fmt"
	"strings"
)

func _ptrg(mem uint64) (uint32, uint32) {
	//endiness ...
	ptr := uint32(mem >> 32)
	size := uint32(mem)
	return ptr, size
}

func (app *App) ptrToString(mem uint64) (string, error) {
	if app.wasm == nil {
		return "", errors.New("wasm is nil")
	}

	ptr, size := _ptrg(mem)

	bytes, ok := app.wasm.mod.Memory().Read(ptr, size)
	if !ok {
		return "", fmt.Errorf("Memory.Read(%d, %d) out of range of memory size %d", ptr, size, app.wasm.mod.Memory().Size())
	}
	return strings.Clone(string(bytes)), nil
}

func (app *App) stringToPtr(str string, dst uint64) error {
	if app.wasm == nil {
		return errors.New("wasm is nil")
	}

	ptr, size := _ptrg(dst)
	n := uint32(len(str))
	if n < size {
		size = n
	}

	if !app.wasm.mod.Memory().Write(ptr, []byte(str)) {
		return fmt.Errorf("Memory.Read(%d, %d) out of range of memory size %d", ptr, size, app.wasm.mod.Memory().Size())
	}
	return nil
}

/*func (app *Asset) ptrToBytes(mem uint64) ([]byte, error) {
	if app.wasm == nil {
		return nil, errors.New("wasm is nil")
	}

	ptr, size := _ptrg(mem)
	bts, ok := app.wasm.mod.Memory().Read(ptr, size)
	if !ok {
		return nil, fmt.Errorf("Memory.Read(%d, %d) out of range of memory size %d", ptr, size, app.wasm.mod.Memory().Size())
	}
	return bytes.Clone(bts)
}*/

func (app *App) ptrToBytesDirect(mem uint64) ([]byte, error) {
	if app.wasm == nil {
		return nil, errors.New("wasm is nil")
	}

	ptr, size := _ptrg(mem)
	bts, ok := app.wasm.mod.Memory().Read(ptr, size)
	if !ok {
		return nil, fmt.Errorf("Memory.Read(%d, %d) out of range of memory size %d", ptr, size, app.wasm.mod.Memory().Size())
	}
	return bts, nil
}

func (app *App) bytesToPtr(src []byte, dst uint64) error {
	if app.wasm == nil {
		return errors.New("wasm is nil")
	}

	ptr, size := _ptrg(dst)
	n := uint32(len(src))
	if n < size {
		size = n
	}

	//copy string into memory
	if !app.wasm.mod.Memory().Write(ptr, src) {
		return fmt.Errorf("Memory.Read(%d, %d) out of range of memory size %d", ptr, size, app.wasm.mod.Memory().Size())
	}
	return nil
}

func (app *App) storage_write(data []byte) (int64, error) {
	err := app.SetStorage(data)
	if err != nil {
		return -1, err
	}

	return 1, nil
}

func (app *App) _sa_storage_write(jsonStorage uint64) int64 {
	data, err := app.ptrToBytesDirect(jsonStorage)
	if app.AddLogErr(err) {
		return -1
	}

	ret, err := app.storage_write(data)
	app.AddLogErr(err)
	return ret
}
