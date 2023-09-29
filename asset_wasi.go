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
	"encoding/binary"
	"fmt"
	"math"
	"os"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

type AssetWasm struct {
	asset *Asset

	rt     wazero.Runtime
	mod    api.Module
	malloc api.Function
	free   api.Function

	load_tm int64
}

func NewAssetWasm(asset *Asset) (*AssetWasm, error) {
	var aw AssetWasm
	aw.asset = asset

	err := aw.InstantiateEnv()

	return &aw, err
}

func (aw *AssetWasm) SaveData() {
	if aw.mod != nil {
		aw.mod.ExportedFunction("_sa_exit").Call(aw.asset.app.root.ctx)
	}
}

func (aw *AssetWasm) destroyMod() {
	if aw.mod != nil {
		aw.mod.Close(aw.asset.app.root.ctx)
		aw.mod = nil
	}
}

func (aw *AssetWasm) Destroy() {
	aw.destroyMod()
	aw.rt.Close(aw.asset.app.root.ctx)
}

func (aw *AssetWasm) InstantiateEnv() error {

	aw.rt = wazero.NewRuntimeWithConfig(aw.asset.app.root.ctx, aw.asset.app.root.runtimeConfig)
	wasi_snapshot_preview1.MustInstantiate(aw.asset.app.root.ctx, aw.rt)

	env := aw.rt.NewHostModuleBuilder("env")

	//these function are constraint into particular 'asset'!!!
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_info_float).Export("_sa_info_float")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_info_setFloat).Export("_sa_info_setFloat")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_info_string).Export("_sa_info_string")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_info_string_len).Export("_sa_info_string_len")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_info_setString).Export("_sa_info_setString")

	env.NewFunctionBuilder().WithFunc(aw.asset._sa_blob).Export("_sa_blob")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_blob_len).Export("_sa_blob")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_storage_write).Export("_sa_storage_write")

	env.NewFunctionBuilder().WithFunc(aw.asset._sa_sql_write).Export("_sa_sql_write")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_sql_read).Export("_sa_sql_read")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_sql_readRowCount).Export("_sa_sql_readRowCount")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_sql_readRowLen).Export("_sa_sql_readRowLen")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_sql_readRow).Export("_sa_sql_readRow")

	env.NewFunctionBuilder().WithFunc(aw.asset._sa_div_colResize).Export("_sa_div_colResize")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_div_rowResize).Export("_sa_div_rowResize")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_div_colMax).Export("_sa_div_colMax")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_div_rowMax).Export("_sa_div_rowMax")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_div_col).Export("_sa_div_col")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_div_row).Export("_sa_div_row")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_div_start).Export("_sa_div_start")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_div_end).Export("_sa_div_end")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_div_dialogOpen).Export("_sa_div_dialogOpen")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_div_dialogClose).Export("_sa_div_dialogClose")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_div_dialogStart).Export("_sa_div_dialogStart")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_div_dialogEnd).Export("_sa_div_dialogEnd")

	env.NewFunctionBuilder().WithFunc(aw.asset._sa_div_get_info).Export("_sa_div_get_info")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_div_set_info).Export("_sa_div_set_info")

	env.NewFunctionBuilder().WithFunc(aw.asset._sa_div_drag).Export("_sa_div_drag")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_div_drop).Export("_sa_div_drop")

	env.NewFunctionBuilder().WithFunc(aw.asset._sa_register_style).Export("_sa_register_style")

	env.NewFunctionBuilder().WithFunc(aw.asset._sa_render_app).Export("_sa_render_app")

	env.NewFunctionBuilder().WithFunc(aw.asset._sa_paint_rect).Export("_sa_paint_rect")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_paint_circle).Export("_sa_paint_circle")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_paint_line).Export("_sa_paint_line")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_paint_file).Export("_sa_paint_file")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_paint_title).Export("_sa_paint_title")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_paint_text).Export("_sa_paint_text")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_paint_textWidth).Export("_sa_paint_textWidth")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_paint_cursor).Export("_sa_paint_cursor")

	env.NewFunctionBuilder().WithFunc(aw.asset._sa_fn_call).Export("_sa_fn_call")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_fn_setReturn).Export("_sa_fn_setReturn")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_fn_getReturn).Export("_sa_fn_getReturn")

	env.NewFunctionBuilder().WithFunc(aw.asset._sa_comp_drawButton).Export("_sa_comp_drawButton")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_comp_drawSlider).Export("_sa_comp_drawSlider")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_comp_drawProgress).Export("_sa_comp_drawProgress")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_comp_drawText).Export("_sa_comp_drawText")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_comp_drawEdit).Export("_sa_comp_drawEdit")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_comp_getEditValue).Export("_sa_comp_getEditValue")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_comp_drawCombo).Export("_sa_comp_drawCombo")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_comp_drawCheckbox).Export("_sa_comp_drawCheckbox")

	env.NewFunctionBuilder().WithFunc(aw.asset._sa_print).Export("_sa_print")
	env.NewFunctionBuilder().WithFunc(aw.asset._sa_print_float).Export("_sa_print_float")

	_, err := env.Instantiate(aw.asset.app.root.ctx)
	return err
}

func (aw *AssetWasm) Call(fnName string, args []byte) (int64, error) {

	if aw.mod == nil {
		return 0, fmt.Errorf("mod is nil")
	}

	fn := aw.mod.ExportedFunction(fnName)

	var frees []uint64
	var params []uint64
	{
		fnParamTps := fn.Definition().ParamTypes()

		//copy []bytes -> []uint64
		i := 0
		p := 0
		for p < len(args) && i < len(fnParamTps) {

			tp := args[p]
			fnParamTp := fnParamTps[i]
			p += 1

			arg := binary.LittleEndian.Uint64(args[p:])
			p += 8

			switch tp {

			case TpI64:
				switch fnParamTp {
				case api.ValueTypeF32:
					arg = uint64(math.Float32bits(float32(arg)))
				case api.ValueTypeF64:
					arg = uint64(math.Float64bits(float64(arg)))
				}
				params = append(params, arg)

			case TpF32:
				vf := math.Float32frombits(uint32(arg))
				switch fnParamTp {
				case api.ValueTypeI32, api.ValueTypeI64:
					arg = uint64(vf)
				case api.ValueTypeF64:
					arg = uint64(math.Float64bits(float64(vf)))
				}
				params = append(params, arg)

			case TpF64:
				vf := math.Float64frombits(arg)
				switch fnParamTp {
				case api.ValueTypeI32, api.ValueTypeI64:
					arg = uint64(vf)
				case api.ValueTypeF32:
					arg = uint64(math.Float32bits(float32(vf)))
				}
				params = append(params, arg)

			case TpBytes, TpString:
				src_n := int(arg)
				src := args[p : p+src_n]
				p += int(arg)

				if fnParamTp != TpI64 {
					return -1, fmt.Errorf("parameter is array/string, but pass is not TpI64(pointer)")

				}

				// alloc
				results, err := aw.malloc.Call(aw.asset.app.root.ctx, uint64(src_n))
				if err != nil {
					return -1, fmt.Errorf("wasm malloc() failed: %w", err)
				}
				frees = append(frees, results[0]) //free() later

				// dst ptr
				arg = (uint64(uintptr(results[0])) << uint64(32)) | uint64(src_n)
				params = append(params, arg)

				// copy
				err = aw.asset.bytesToPtr(src, arg)
				if err != nil {
					return -1, err
				}
			}

			i++
		}
	}

	aw.asset.app.fn2Return = nil
	aw.asset.app.fn2Returns = nil

	//call
	res, err := fn.Call(aw.asset.app.root.ctx, params...)
	if err != nil {
		return -1, fmt.Errorf("wasm module failed: %w", err)
	}

	//free
	for _, it := range frees {
		_, err := aw.free.Call(aw.asset.app.root.ctx, it)
		if err != nil {
			return -1, fmt.Errorf("wasm free() failed: %w", err)
		}
	}

	//return
	if len(res) > 0 {
		aw.asset.app.fn2Return = make([]byte, 1+8)
		aw.asset.app.fn2Return[0] = fn.Definition().ResultTypes()[0]
		binary.LittleEndian.PutUint64(aw.asset.app.fn2Return[1:], res[0])
	}

	return int64(len(aw.asset.app.fn2Return) + len(aw.asset.app.fn2Returns)), nil
}

func (aw *AssetWasm) LoadModule() error {

	wasmFile, err := os.ReadFile(aw.asset.getWasmPath())
	if err != nil {
		return fmt.Errorf("ReadFile failed: %w", err)
	}

	aw.SaveData()
	aw.destroyMod()

	aw.mod, err = aw.rt.Instantiate(aw.asset.app.root.ctx, wasmFile)
	if err != nil {
		return fmt.Errorf("Instantiate() failed: %w", err)
	}

	if aw.mod != nil {
		aw.malloc = aw.mod.ExportedFunction("malloc")
		aw.free = aw.mod.ExportedFunction("free")
	}

	return nil
}

func (aw *AssetWasm) Tick() (bool, error) {

	stat, err := os.Stat(aw.asset.getWasmPath())
	if err == nil && !stat.IsDir() {
		if aw.mod == nil || stat.ModTime().UnixMilli() != aw.load_tm {
			err = aw.LoadModule()
			aw.load_tm = stat.ModTime().UnixMilli()
			return true, err
		}
	}

	return false, nil
}
