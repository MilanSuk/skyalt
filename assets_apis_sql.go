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

import "fmt"

func (app *App) _getDb(dbUrl string, needWrite bool) (*Db, error) {

	var dbPath string
	var err error

	if len(dbUrl) == 0 {
		dbPath = app.db.path
	} else {
		var inAsset bool
		dbPath, _, inAsset, err = FileParseUrl(dbUrl, app)
		if err != nil {
			return nil, fmt.Errorf("DbParseUrl() failed: %w", err)
		}
		if needWrite && inAsset {
			return nil, fmt.Errorf("database is read-only")
		}
	}

	db, err := app.db.root.AddDb(dbPath)
	if err != nil {
		return nil, fmt.Errorf("AddDb() failed: %w", err)
	}

	return db, nil
}

func (app *App) sql_commit(dbUrl string) (int64, error) {

	db, err := app._getDb(dbUrl, true)
	if err != nil {
		return -1, err
	}

	err = db.Commit()
	if err != nil {
		return -1, err
	}

	return 1, nil //just ok
}

func (app *App) _sa_sql_commit(dbUrlMem uint64) int64 {

	db, err := app.ptrToString(dbUrlMem)
	if app.AddLogErr(err) {
		return -1
	}

	ret, err := app.sql_commit(db)
	app.AddLogErr(err)
	return ret
}

func (app *App) sql_rollback(dbUrl string) (int64, error) {

	db, err := app._getDb(dbUrl, true)
	if err != nil {
		return -1, err
	}

	err = db.Rollback()
	if err != nil {
		return -1, err
	}

	return 1, nil //just ok
}

func (app *App) _sa_sql_rollback(dbUrlMem uint64) int64 {

	db, err := app.ptrToString(dbUrlMem)
	if app.AddLogErr(err) {
		return -1
	}

	ret, err := app.sql_rollback(db)
	app.AddLogErr(err)
	return ret
}

func (app *App) sql_write(dbUrl string, query string) (int64, error) {

	db, err := app._getDb(dbUrl, true)
	if err != nil {
		return -1, err
	}

	res, err := db.Write(query)
	if err != nil {
		return -1, fmt.Errorf("Exec(%s) for query(%s) failed: %w", db.GetPath(), query, err)
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return -1, fmt.Errorf("RowsAffected(%s) failed: %w", db.GetPath(), err)
	}

	if aff <= 0 {
		return 0, nil
	}

	insRow, err := res.LastInsertId()
	if err == nil && insRow > 0 {
		return insRow, nil //inserted row(if insert)
	}

	return 1, nil //just ok
}

func (app *App) _sa_sql_write(dbUrlMem uint64, queryMem uint64) int64 {

	db, err := app.ptrToString(dbUrlMem)
	if app.AddLogErr(err) {
		return -1
	}
	query, err := app.ptrToString(queryMem)
	if app.AddLogErr(err) {
		return -1
	}

	ret, err := app.sql_write(db, query)
	app.AddLogErr(err)
	return ret
}

func (app *App) sql_read(dbUrl string, query string) (int64, error) {

	db, err := app._getDb(dbUrl, false)
	if db == nil {
		return -1, err
	}
	cache, err := db.AddCache(query)
	if err != nil {
		return -1, err
	}

	return cache.query_hash, nil
}
func (app *App) _sa_sql_read(dbUrlMem uint64, queryMem uint64) int64 {

	db, err := app.ptrToString(dbUrlMem)
	if app.AddLogErr(err) {
		return -1
	}
	query, err := app.ptrToString(queryMem)
	if app.AddLogErr(err) {
		return -1
	}

	ret, err := app.sql_read(db, query)
	app.AddLogErr(err)
	return ret

}

func (app *App) sql_readRowCount(dbUrl string, query string, queryHash int64) (int64, error) {

	db, err := app._getDb(dbUrl, false)
	if db == nil {
		return -1, err
	}

	cache := db.FindCache(queryHash)

	if cache == nil {
		var err error
		cache, err = db.AddCache(query)
		if err != nil {
			return -1, err
		}
	}

	return int64(len(cache.result_rows)), nil
}

func (app *App) sql_readRowLen(dbUrl string, query string, queryHash int64, row_i uint64) (int64, error) {

	db, err := app._getDb(dbUrl, false)
	if db == nil {
		return -1, err
	}

	cache := db.FindCache(queryHash)

	if cache == nil {
		var err error
		cache, err = db.AddCache(query)
		if err != nil {
			return -1, err
		}
	}

	if row_i < uint64(len(cache.result_rows)) {
		return int64(len(cache.result_rows[row_i])), nil
	}

	return 0, nil //no more rows
}

func (app *App) _sa_sql_readRowCount(dbUrlMem uint64, queryMem uint64, queryHash int64) int64 {

	db, err := app.ptrToString(dbUrlMem)
	if app.AddLogErr(err) {
		return -1
	}
	query, err := app.ptrToString(queryMem)
	if app.AddLogErr(err) {
		return -1
	}

	ret, err := app.sql_readRowCount(db, query, queryHash)
	app.AddLogErr(err)
	return ret
}

func (app *App) _sa_sql_readRowLen(dbUrlMem uint64, queryMem uint64, queryHash int64, row_i uint64) int64 {

	db, err := app.ptrToString(dbUrlMem)
	if app.AddLogErr(err) {
		return -1
	}
	query, err := app.ptrToString(queryMem)
	if app.AddLogErr(err) {
		return -1
	}

	ret, err := app.sql_readRowLen(db, query, queryHash, row_i)
	app.AddLogErr(err)
	return ret
}

func (app *App) sql_readRow(dbUrl string, query string, queryHash int64, row_i uint64) ([]byte, int64, error) {

	db, err := app._getDb(dbUrl, false)
	if db == nil {
		return nil, -1, err
	}

	cache := db.FindCache(queryHash)
	if cache == nil {
		var err error
		cache, err = db.AddCache(query)
		if err != nil {
			return nil, -1, err
		}
	}

	if row_i < uint64(len(cache.result_rows)) {
		return cache.result_rows[row_i], 1, nil
	}

	return nil, 0, nil //no more rows
}

func (app *App) _sa_sql_readRow(dbUrlMem uint64, queryMem uint64, queryHash int64, row_i uint64, resultMem uint64) int64 {

	db, err := app.ptrToString(dbUrlMem)
	if app.AddLogErr(err) {
		return -1
	}
	query, err := app.ptrToString(queryMem)
	if app.AddLogErr(err) {
		return -1
	}

	dst, ret, err := app.sql_readRow(db, query, queryHash, row_i)
	app.AddLogErr(err)

	if ret > 0 {
		err := app.bytesToPtr(dst, resultMem)
		app.AddLogErr(err)
		if err != nil {
			return -1
		}
	}

	return ret
}
