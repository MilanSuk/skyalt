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

	_ "github.com/mattn/go-sqlite3"
)

type DbSettings struct {
	root *Root
	db   *Db

	max_sts_uid int
}

func NewDbSettings(root *Root) (*DbSettings, error) {
	var sts DbSettings
	sts.root = root

	var err error
	sts.db, err = root.AddDb(root.folderDatabases + "/settings.sqlite")
	if err != nil {
		return nil, fmt.Errorf("AddDb() failed: %w", err)
	}

	_, err = sts.db.Write("CREATE TABLE IF NOT EXISTS settings(id INT, asset TEXT, content BLOB);")
	if err != nil {
		return nil, fmt.Errorf("Write() failed: %w", err)
	}
	sts.db.Commit()

	{
		rows, err := sts.db.db.Query("SELECT MAX(id) FROM settings")
		if err != nil {
			return nil, fmt.Errorf("query MAX(id) failed: %w", err)
		}

		if rows.Next() {
			rows.Scan(&sts.max_sts_uid)

			//fails when 0 rows ...
			//err := rows.Scan(&sts.max_sts_uid)
			//if err != nil {
			//	return nil, fmt.Errorf("Scan(%s) failed: %w", sts.GetPath(), err)
			//}
		}
	}

	return &sts, nil
}

func (sts *DbSettings) Destroy() error {
	return nil
}

func (sts *DbSettings) DbSettings_GetName() string {
	return sts.db.path + ".sqlite"
}

func (sts *DbSettings) AddSts_uid() int {
	sts.max_sts_uid++
	return sts.max_sts_uid
}

func (sts *DbSettings) Add(id int, asset string) (int, error) {
	sts.max_sts_uid = OsMax(sts.max_sts_uid, id)

	res, err := sts.db.Write("INSERT INTO settings(id, asset) VALUES(?, ?);", id, asset)
	if err != nil {
		return -1, fmt.Errorf("Write() failed: %w", err)
	}

	rowid, err := res.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("LastInsertId() failed: %w", err)
	}

	sts.db.Commit()
	return int(rowid), nil
}

func (sts *DbSettings) Find(id int, asset string) (int, error) {
	rows, err := sts.db.db.Query("SELECT rowid FROM settings WHERE id=? AND asset=?", id, asset)
	if err != nil {
		return -1, fmt.Errorf("query SELECT failed: %w", err)
	}

	rowid := -1 //not found
	if rows.Next() {
		err := rows.Scan(&rowid)
		if err != nil {
			return -1, fmt.Errorf("Scan() failed: %w", err)
		}
	}
	return rowid, nil
}

func (sts *DbSettings) FindOrAdd(id int, asset string) (int, error) {
	rowid, err := sts.Find(id, asset)
	if err != nil {
		return -1, fmt.Errorf("Find() failed: %w", err)
	}

	if rowid < 0 {
		rowid, err = sts.Add(id, asset)
		if err != nil {
			return -1, fmt.Errorf("Add() failed: %w", err)
		}
	}

	return rowid, nil
}

func (sts *DbSettings) GetContent(rowid int) ([]byte, error) {

	rows := sts.db.db.QueryRow("SELECT content FROM settings WHERE rowid=?", rowid)

	var content []byte
	err := rows.Scan(&content)
	if err != nil {
		return nil, fmt.Errorf("Scan() failed: %w", err)
	}
	return content, nil
}

func (sts *DbSettings) SetContent(rowid int, content []byte) error {

	_, err := sts.db.Write("UPDATE settings SET content=? WHERE rowid=?;", content, rowid)
	if err != nil {
		return fmt.Errorf("Write() failed: %w", err)
	}

	sts.db.Commit()
	return nil
}

func (sts *DbSettings) Remove(rowid int) error {
	_, err := sts.db.Write("DELETE FROM settings WHERE rowid=?;", rowid)
	if err != nil {
		return fmt.Errorf("Write() failed: %w", err)
	}
	sts.db.Commit()
	return nil
}

func (sts *DbSettings) Rename(rowid int, file, app, name string) error {

	_, err := sts.db.Write("UPDATE settings SET file=?, app=?, name=? WHERE rowid=?;", file, app, name, rowid)
	if err != nil {
		return fmt.Errorf("query UPDATE(name) failed: %w", err)
	}

	sts.db.Commit()
	return nil
}

func (sts *DbSettings) Duplicate(srcid int) (int, error) {

	dstId := sts.AddSts_uid()

	rows, err := sts.db.db.Query("SELECT asset, content FROM settings WHERE id=?", srcid)
	if err != nil {
		return -1, fmt.Errorf("query SELECT failed: %w", err)
	}

	for rows.Next() {

		//get
		var asset string
		var content []byte
		err := rows.Scan(&asset, &content)
		if err != nil {
			return -1, fmt.Errorf("Scan() failed: %w", err)
		}

		//insert
		_, err = sts.db.Write("INSERT INTO settings(id, asset, content) VALUES(?, ?, ?);", dstId, asset, content)
		if err != nil {
			return -1, fmt.Errorf("Write() failed: %w", err)
		}
	}

	sts.db.Commit()
	return dstId, nil

}
