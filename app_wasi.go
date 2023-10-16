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
	"fmt"
	"os"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

const TpI64 = byte(0x7e)
const TpF32 = byte(0x7d)
const TpF64 = byte(0x7c)
const TpBytes = byte(0x7b)
const TpString = byte(0x7a)

type AppWasm struct {
	app *App

	rt     wazero.Runtime
	mod    api.Module
	malloc api.Function
	free   api.Function

	load_tm int64
}

func NewAppWasm(app *App) (*AppWasm, error) {
	var aw AppWasm
	aw.app = app

	err := aw.InstantiateEnv()

	return &aw, err
}

func (aw *AppWasm) SaveData() {
	if aw.mod != nil {
		aw.mod.ExportedFunction("_sa_save").Call(aw.app.db.root.ctx)
	}
}

func (aw *AppWasm) destroyMod() {
	if aw.mod != nil {
		aw.mod.Close(aw.app.db.root.ctx)
		aw.mod = nil
	}
}

func (aw *AppWasm) Destroy() {
	aw.destroyMod()
	aw.rt.Close(aw.app.db.root.ctx)
}

func (aw *AppWasm) InstantiateEnv() error {

	aw.rt = wazero.NewRuntimeWithConfig(aw.app.db.root.ctx, aw.app.db.root.runtimeConfig)
	wasi_snapshot_preview1.MustInstantiate(aw.app.db.root.ctx, aw.rt)

	env := aw.rt.NewHostModuleBuilder("env")

	//these function are constraint into particular 'app'!!!
	env.NewFunctionBuilder().WithFunc(aw.app._sa_info_get_prepare).Export("_sa_info_get_prepare")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_info_get).Export("_sa_info_get")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_info_set).Export("_sa_info_set")

	env.NewFunctionBuilder().WithFunc(aw.app._sa_blob).Export("_sa_blob")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_blob_len).Export("_sa_blob_len")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_storage_write).Export("_sa_storage_write")

	env.NewFunctionBuilder().WithFunc(aw.app._sa_sql_commit).Export("_sa_sql_commit")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_sql_rollback).Export("_sa_sql_rollback")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_sql_write).Export("_sa_sql_write")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_sql_read).Export("_sa_sql_read")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_sql_readRowCount).Export("_sa_sql_readRowCount")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_sql_readRowLen).Export("_sa_sql_readRowLen")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_sql_readRow).Export("_sa_sql_readRow")

	env.NewFunctionBuilder().WithFunc(aw.app._sa_div_colResize).Export("_sa_div_colResize")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_div_rowResize).Export("_sa_div_rowResize")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_div_colMax).Export("_sa_div_colMax")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_div_rowMax).Export("_sa_div_rowMax")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_div_col).Export("_sa_div_col")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_div_row).Export("_sa_div_row")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_div_start).Export("_sa_div_start")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_div_startEx).Export("_sa_div_startEx")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_div_end).Export("_sa_div_end")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_div_dialogOpen).Export("_sa_div_dialogOpen")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_div_dialogClose).Export("_sa_div_dialogClose")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_div_dialogStart).Export("_sa_div_dialogStart")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_div_dialogEnd).Export("_sa_div_dialogEnd")

	env.NewFunctionBuilder().WithFunc(aw.app._sa_div_info_get).Export("_sa_div_info_get")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_div_info_set).Export("_sa_div_info_set")

	env.NewFunctionBuilder().WithFunc(aw.app._sa_div_drag).Export("_sa_div_drag")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_div_drop).Export("_sa_div_drop")

	env.NewFunctionBuilder().WithFunc(aw.app._sa_register_style).Export("_sa_register_style")

	env.NewFunctionBuilder().WithFunc(aw.app._sa_render_app).Export("_sa_render_app")

	env.NewFunctionBuilder().WithFunc(aw.app._sa_paint_rect).Export("_sa_paint_rect")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_paint_circle).Export("_sa_paint_circle")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_paint_line).Export("_sa_paint_line")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_paint_file).Export("_sa_paint_file")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_paint_tooltip).Export("_sa_paint_tooltip")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_paint_text).Export("_sa_paint_text")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_paint_textWidth).Export("_sa_paint_textWidth")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_paint_cursor).Export("_sa_paint_cursor")

	env.NewFunctionBuilder().WithFunc(aw.app._sa_comp_drawButton).Export("_sa_comp_drawButton")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_comp_drawSlider).Export("_sa_comp_drawSlider")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_comp_drawProgress).Export("_sa_comp_drawProgress")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_comp_drawText).Export("_sa_comp_drawText")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_comp_drawEdit).Export("_sa_comp_drawEdit")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_comp_getEditValue).Export("_sa_comp_getEditValue")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_comp_drawCombo).Export("_sa_comp_drawCombo")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_comp_drawCheckbox).Export("_sa_comp_drawCheckbox")

	env.NewFunctionBuilder().WithFunc(aw.app._sa_print).Export("_sa_print")
	env.NewFunctionBuilder().WithFunc(aw.app._sa_print_float).Export("_sa_print_float")

	_, err := env.Instantiate(aw.app.db.root.ctx)
	return err
}

func (aw *AppWasm) Call(fnName string) (int64, error) {

	if aw.mod == nil {
		return 0, fmt.Errorf("mod is nil")
	}

	fn := aw.mod.ExportedFunction(fnName)

	//call
	rets, err := fn.Call(aw.app.db.root.ctx)
	if err != nil {
		return 0, fmt.Errorf("wasm module failed: %w", err)
	}

	//return
	ret := int64(0)
	if len(rets) > 0 {
		ret = int64(rets[0])
	}
	return ret, nil
}

func (aw *AppWasm) LoadModule() error {

	wasmFile, err := os.ReadFile(aw.app.getWasmPath())
	if err != nil {
		return fmt.Errorf("ReadFile failed: %w", err)
	}

	aw.SaveData()
	aw.destroyMod()

	aw.mod, err = aw.rt.Instantiate(aw.app.db.root.ctx, wasmFile)
	if err != nil {
		return fmt.Errorf("Instantiate() failed: %w", err)
	}

	if aw.mod != nil {
		aw.malloc = aw.mod.ExportedFunction("malloc")
		aw.free = aw.mod.ExportedFunction("free")
	}

	return nil
}

func (aw *AppWasm) Tick() (bool, error) {

	stat, err := os.Stat(aw.app.getWasmPath())
	if err == nil && !stat.IsDir() {
		if aw.mod == nil || stat.ModTime().UnixMilli() != aw.load_tm {
			err = aw.LoadModule()
			aw.load_tm = stat.ModTime().UnixMilli()
			return true, err
		}
	}

	return false, nil
}
