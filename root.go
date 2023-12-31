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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/tetratelabs/wazero"
)

type Root struct {
	ctx             context.Context
	folderApps      string
	folderDatabases string

	cacheDir      string
	cache         wazero.CompilationCache
	runtimeConfig wazero.RuntimeConfig

	dbs map[string]*Db

	dbsList string

	last_ticks int64

	levels *LayoutLevels

	touch LayerTouch
	tile  LayerTile

	ui *Ui

	baseApp *App

	fonts *Fonts

	ui_info Info
	vm_info Info

	editbox_history VmTextHistoryArray

	server *DebugServer

	exit bool
	save bool

	debug_line string

	styles   DivDefaultStyles
	stylesJs []byte
}

func NewRoot(debugPORT int, folderApps string, folderDatabases string, ctx context.Context) (*Root, error) {
	var root Root
	var err error
	root.ctx = ctx

	root.fonts = NewFonts()
	root.dbs = make(map[string]*Db)

	root.folderApps = folderApps
	root.folderDatabases = folderDatabases

	err = OsFolderCreate(folderApps)
	if err != nil {
		return nil, fmt.Errorf("OsFolderCreate() failed: %w", err)
	}

	err = OsFolderCreate(folderDatabases)
	if err != nil {
		return nil, fmt.Errorf("OsFolderCreate() failed: %w", err)
	}

	if !OsFolderExists(folderApps) {
		return nil, fmt.Errorf("Folder(%s) not exist", folderApps)
	}
	if !OsFolderExists(folderDatabases) {
		return nil, fmt.Errorf("Folder(%s) not exist", folderDatabases)
	}

	// init wasm
	root.cacheDir, err = os.MkdirTemp("", "wasm_cache")
	if err != nil {
		return nil, fmt.Errorf("MkdirTemp() failed: %w", err)
	}
	root.cache, err = wazero.NewCompilationCacheWithDir(root.cacheDir)
	if err != nil {
		return nil, fmt.Errorf("NewCompilationCacheWithDir() failed: %w", err)
	}
	root.runtimeConfig = wazero.NewRuntimeConfig().WithCompilationCache(root.cache)

	root.ui, err = NewUi(root.GetIniPath())
	if err != nil {
		return nil, fmt.Errorf("NewUi() failed: %w", err)
	}

	root.server, err = NewDebugServer(debugPORT)
	if err != nil {
		return nil, fmt.Errorf("NewDebugServer() failed: %w", err)
	}

	db, err := root.AddDb(root.folderDatabases + "/base.sqlite")
	if err != nil {
		return nil, fmt.Errorf("AddDb() failed: %w", err)
	}
	err = db.AddFirstRowId("base")
	if err != nil {
		return nil, fmt.Errorf("AddFirstRowId() failed: %w", err)
	}

	root.baseApp, err = db.AddApp(1)
	if err != nil {
		return nil, fmt.Errorf("AddApp() failed: %w", err)
	}

	root.levels, err = NewLayoutLevels(root.baseApp, root.ui)
	if err != nil {
		return nil, fmt.Errorf("NewLayoutLevels() failed: %w", err)
	}

	root.updateDbsList()

	err = root.ReopenApps()
	if err != nil {
		return nil, fmt.Errorf("ReloadApps() failed: %w", err)
	}

	//root.PackageAllApps()
	return &root, nil
}

func (root *Root) Destroy() {

	if root.server != nil {
		root.server.Destroy()
	}

	for _, db := range root.dbs {
		db.Destroy()
	}

	root.fonts.Destroy()

	root.cache.Close(root.ctx)
	os.RemoveAll(root.cacheDir)

	//save storage
	{
		err := root.ui.io.Save(root.GetIniPath())
		if err != nil {
			fmt.Printf("Open() failed: %v\n", err)
		}

		root.levels.Destroy()
	}

	root.ui.Destroy() //also save ini.json
}

func (root *Root) ResetTick() {
	root.last_ticks = 0
}

func (root *Root) ReopenApps() error {

	root.styles = DivStyles_getDefaults(root)
	var err error
	root.stylesJs, err = json.MarshalIndent(&root.styles, "", "")
	if err != nil {
		return fmt.Errorf("MarshalIndent() failed: %w", err)
	}

	for _, it := range root.dbs {
		it.ReopenApps()
	}

	root.ResetTick()
	return nil
}

func (root *Root) GetIniPath() string {
	return "ini.json"
}

func (root *Root) FindDb(path string) *Db {
	db, found := root.dbs[path]
	if found {
		return db
	}
	return nil
}
func (root *Root) AddDb(path string) (*Db, error) {

	//finds
	db := root.FindDb(path)
	if db != nil {
		return db, nil
	}

	//check if file exist
	if !OsFileExists(path) {
		return nil, fmt.Errorf("db(%s) file not exist", path)
	}

	//adds
	var err error
	db, err = NewDb(root, path)
	if err != nil {
		return nil, err
	}

	root.dbs[path] = db
	return db, nil
}

func (root *Root) CreateDb(name string) bool {

	newPath := root.folderDatabases + "/" + name
	if OsFileExists(newPath) {
		fmt.Printf("newPath(%s) already exist\n", newPath)
		return false
	}

	f, err := os.Create(newPath)
	if err != nil {
		fmt.Printf("Create(%s) failed: %v\n", newPath, name)
		return false
	}

	err = f.Close()
	if err != nil {
		fmt.Printf("Close(%s) failed: %v\n", newPath, name)
		return false
	}

	root.updateDbsList()
	return true
}

func (root *Root) RenameDb(name string, newName string) bool {

	if strings.ContainsRune(newName, '/') || strings.ContainsRune(newName, '\\') {
		fmt.Printf("newName(%s) has invalid character\n", name)
		return false
	}

	path := root.folderDatabases + "/" + name
	newPath := root.folderDatabases + "/" + newName
	if OsFileExists(newPath) {
		fmt.Printf("newPath(%s) already exist\n", newPath)
		return false
	}

	//finds
	db, found := root.dbs[path]
	if found {
		//close
		db.SaveApps()
		db.Destroy()
		delete(root.dbs, path)
	}

	//rename file
	err := OsFileRename(path, newPath)
	if err != nil {
		fmt.Printf("OsFileRemove(%s) failed: %v\n", path, err)
	}
	if OsFileExists(path + "-shm") {
		err = OsFileRename(path+"-shm", newPath+"-shm")
		if err != nil {
			fmt.Printf("OsFileRemove(%s) failed: %v\n", path, err)
		}
	}
	if OsFileExists(path + "-wal") {
		err = OsFileRename(path+"-wal", newPath+"-wal")
		if err != nil {
			fmt.Printf("OsFileRemove(%s) failed: %v\n", path, err)
		}
	}

	root.updateDbsList()
	return true
}

func (root *Root) DuplicateDb(name string, newName string) bool {

	if strings.ContainsRune(newName, '/') || strings.ContainsRune(newName, '\\') {
		fmt.Printf("newName(%s) has invalid character\n", name)
		return false
	}

	path := root.folderDatabases + "/" + name
	newPath := root.folderDatabases + "/" + newName
	if OsFileExists(newPath) {
		fmt.Printf("newPath(%s) already exist\n", newPath)
		return false
	}

	//finds
	db, found := root.dbs[path]
	if found {
		db.SaveApps()
	}

	//duplicate file
	err := OsFileCopy(path, newPath)
	if err != nil {
		fmt.Printf("OsFileCopy(%s) failed: %v\n", path, err)
	}

	root.updateDbsList()
	return true
}

func (root *Root) RemoveDb(path string) bool {

	//finds
	db, found := root.dbs[path]
	if found {
		//close
		db.Destroy()
		delete(root.dbs, path)
	}

	//delete file
	err := OsFileRemove(path)
	if err != nil {
		fmt.Printf("OsFileRemove(%s) failed: %v\n", path, err)
	}
	if OsFileExists(path + "-shm") {
		err = OsFileRemove(path + "-shm")
		if err != nil {
			fmt.Printf("OsFileRemove(%s-shm) failed: %v\n", path, err)
		}
	}
	if OsFileExists(path + "-wal") {
		err = OsFileRemove(path + "-wal")
		if err != nil {
			fmt.Printf("OsFileRemove(%s-wal) failed: %v\n", path, err)
		}
	}

	root.updateDbsList()
	return true
}

func (root *Root) CommitDbs() {
	for _, db := range root.dbs {
		err := db.Commit()
		if err != nil {
			fmt.Printf("Commit() failed: %v\n", err)
		}
	}
}

func (root *Root) Render() {

	winRect, _ := root.ui.GetScreenCoord()
	root.levels.GetBaseDialog().rootDiv.canvas = winRect
	root.levels.GetBaseDialog().rootDiv.crop = winRect

	// close all levels
	if root.ui.io.keys.shift && root.ui.io.keys.esc {
		root.touch.Reset()
		root.levels.CloseAll()
		root.ui.io.keys.esc = false
	}

	root.levels.ResetStack()

	st := root.levels.GetStack()

	st.buff.Reset(st.stack.canvas) //background

	root.baseApp.Render(true)

	root.levels.Maintenance()
	root.levels.Draw()
}

func (root *Root) Tick() (bool, error) {

	if time.Now().UnixMilli() > root.last_ticks+2000 {
		root.last_ticks = time.Now().UnixMilli() //before, because root.ResetTick() can be called inside

		for _, db := range root.dbs {
			db.Tick()
		}

		root.updateDbsList()

		root.ExtractUnpackedApps()

		root.fonts.Maintenance()
	}

	run, err := root.ui.UpdateIO()
	if err != nil {
		return false, fmt.Errorf("UpdateIO() failed: %w", err)
	}

	//tile
	{
		if root.tile.NeedsRedrawFromSleep(root.ui.io.touch.pos) {
			root.ui.ResendInput()
		}
		root.tile.NextTick()
	}

	//debug
	root.debug_line = ""

	if root.ui.NeedRedraw() {

		stUiTicks := OsTicks()
		root.ui.StartRender()
		stVmTicks := OsTicks()

		if root.ui.io.touch.start {
			root.touch.Reset()
		}

		root.Render()

		if root.ui.io.touch.end {
			root.touch.Reset()
			root.ui.io.drag.group = ""
		}

		// tile - redraw If mouse is over tile
		if root.tile.IsActive(root.ui.io.touch.pos) {
			err := root.ui.RenderTile(root.tile.text, root.tile.coord, root.tile.cd, root.fonts.Get(SKYALT_FONT_PATH))
			if err != nil {
				fmt.Printf("RenderTile() failed: %v\n", err)
			}
		}

		if len(root.debug_line) > 0 {
			err := root.ui.RenderTile(root.debug_line, OsV4{root.ui.io.touch.pos, OsV2{1, 1}}, themeBlack(), root.fonts.Get(SKYALT_FONT_PATH))
			if err != nil {
				fmt.Printf("RenderTile() failed: %v\n", err)
			}
		}

		// show fps
		if root.ui.io.ini.Stats {
			root.RenderInfoStats(&root.ui_info, &root.vm_info, root.fonts.Get(SKYALT_FONT_PATH))
		}

		root.vm_info.Update(int(OsTicks() - stVmTicks))
		root.ui.EndRender()
		root.ui_info.Update(int(OsTicks() - stUiTicks))

		root.CommitDbs()

		if root.save {
			for _, db := range root.dbs {
				db.SaveApps()
			}
			root.save = false
		}
	} else {
		time.Sleep(10 * time.Millisecond)
	}

	return (run && !root.exit), err
}

func (root *Root) RenderInfoStats(ui_info *Info, vm_info *Info, font *Font) error {
	if root.ui == nil {
		return nil
	}

	textH := root.ui.io.GetDPI() / 6

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	text := fmt.Sprintf("worst FPS(ui: %.1f, vm: %.1f), avg FPS(ui: %.1f, vm: %.1f), Memory(imgs(%dx): %.2fMB, process: %.2fMB), Threads(%d)",
		ui_info.out_worst_fps, vm_info.out_worst_fps,
		ui_info.out_avg_fps, vm_info.out_avg_fps,
		root.NumTextures(), float64(root.GetImagesBytes())/1024/1024, float64(mem.Sys)/1024/1024,
		runtime.NumGoroutine())
	//netStats.num_connections_opened-netStats.num_connections_closed, netStats.num_sends, netStats.num_recvs)	//, Net(connections: %d, send: %dx, recv: %dx)

	sz, _ := font.GetTextSize(text, g_Font_DEFAULT_Weight, textH, int(float32(textH)*1.2), true)

	cq := OsV4{root.ui.io.GetCoord().Middle().Sub(sz.MulV(0.5)), sz}

	err := root.ui.render.SetClipRect(cq.GetSDLRect())
	if err != nil {
		fmt.Printf("SetClipRect() failed: %v\n", err)
	}
	_Ui_boxSE(root.ui.render, cq.Start, cq.End(), OsCd_white())
	err = font.Print(text, g_Font_DEFAULT_Weight, textH, cq, OsV2{0, 1}, OsCd{255, 50, 50, 255}, nil, true, true, root.ui.render)
	if err != nil {
		fmt.Printf("Print() failed: %v\n", err)
	}

	return nil
}

func (root *Root) updateDbsList() {

	dir, err := os.ReadDir(root.folderDatabases)
	if err != nil {
		fmt.Printf("ReadDir() failed: %v\n", err)
		return
	}

	root.dbsList = ""
	for _, file := range dir {
		if !file.IsDir() {
			if strings.HasSuffix(file.Name(), "-shm") || strings.HasSuffix(file.Name(), "-wal") {
				continue //skip
			}
			root.dbsList += file.Name()
			root.dbsList += "/"
		}
	}
	root.dbsList = strings.TrimSuffix(root.dbsList, "/") //remove '/' at the end
}

func (root *Root) GetAppsList() []OsFileList {
	list := OsFileListBuild(root.folderApps, root.folderApps, true)
	return list.Subs
}

func (root *Root) SetDebugLine(line string) {
	root.debug_line = line
}

func (root *Root) NumTextures() int {
	n := 0
	for _, db := range root.dbs {
		for _, app := range db.apps {
			for _, img := range app.images {
				if img.texture != nil {
					n++
				}
			}
		}
	}
	return n
}

func (root *Root) GetImagesBytes() int64 {
	bytes := int64(0)
	for _, db := range root.dbs {
		for _, app := range db.apps {
			for _, img := range app.images {
				bytes += img.GetBytes()
			}
		}
	}
	return bytes
}

func (root *Root) Vacuum() {
	for _, db := range root.dbs {
		db.Vacuum()
	}
}
