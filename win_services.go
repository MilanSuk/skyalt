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
	"database/sql"
	"fmt"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type WinServices struct {
	lock sync.Mutex

	dbs map[string]*WinServiceDb
}

func NewServices(ini *WinIni) *WinServices {
	var srvs WinServices

	srvs.dbs = make(map[string]*WinServiceDb)

	return &srvs
}

func (srvs *WinServices) Destroy() {
	for _, db := range srvs.dbs {
		db.Destroy()
	}
}

func (srvs *WinServices) Maintenance() {
}

func (srvs *WinServices) OpenDb(path string) (*WinServiceDb, bool, error) {
	if path == "" {
		return nil, false, fmt.Errorf("path is empty")
	}

	srvs.lock.Lock()
	defer srvs.lock.Unlock()

	//find
	db, found := srvs.dbs[path]

	if !found {
		folder := filepath.Dir(path)
		err := OsFolderCreate(folder)
		if err != nil {
			return nil, false, fmt.Errorf("OsFolderCreate(%s) failed: %w", folder, err)
		}

		//open
		db, err = NewServiceDb(path, srvs)
		if err != nil {
			return nil, false, fmt.Errorf("NewServiceDb(%s) failed: %w", path, err)
		}

		//add
		srvs.dbs[path] = db
	}

	return db, found, nil
}

type WinServiceDb struct {
	services *WinServices

	path string
	db   *sql.DB

	lastWriteTicks int64
	lastReadTicks  int64

	last_file_tick int64
}

func NewServiceDb(path string, services *WinServices) (*WinServiceDb, error) {
	var db WinServiceDb
	db.path = path
	db.services = services

	var err error
	db.db, err = sql.Open("sqlite3", "file:"+path+"?&_journal_mode=WAL") //sqlite3 -> sqlite3_skyalt
	if err != nil {
		return nil, fmt.Errorf("sql.Open(%s) from file failed: %w", path, err)
	}

	db.updateTime(false)

	return &db, nil
}

func (db *WinServiceDb) Destroy() {
	db.db.Exec("PRAGMA wal_checkpoint(full);")

	//db.Commit()

	err := db.db.Close()
	if err != nil {
		fmt.Printf("db(%s).Destroy() failed: %v\n", db.path, err)
	}
}

func (db *WinServiceDb) updateTime(writeInto bool) {
	db.last_file_tick = OsTicks()
	if writeInto {
		db.lastWriteTicks = int64(OsTicks())
	}
}

func (db *WinServiceDb) Vacuum() error {
	_, err := db.db.Exec("VACUUM;")
	db.updateTime(false)
	return err
}

func (db *WinServiceDb) Write(query string, params ...any) (sql.Result, error) { //call Read_Lock() - Read_Unlock()
	res, err := db.db.Exec(query, params...)
	if err != nil {
		return nil, fmt.Errorf("query(%s) failed: %w", query, err)
	}

	db.updateTime(true)
	return res, nil
}

func (db *WinServiceDb) ReadRow(query string, params ...any) *sql.Row { //call Read_Lock() - Read_Unlock()
	db.lastReadTicks = int64(OsTicks())
	return db.db.QueryRow(query, params...)
}

func (db *WinServiceDb) Read(query string, params ...any) (*sql.Rows, error) { //call Read_Lock() - Read_Unlock()
	db.lastReadTicks = int64(OsTicks())
	return db.db.Query(query, params...)
}
