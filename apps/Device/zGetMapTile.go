package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Returns map tile from cache or download it [ignore]
type GetMapTile struct {
	X    int //tile's X position
	Y    int //tile's Y position
	Zoom int //map's zoom

	Out_image []byte
}

func (st *GetMapTile) run(caller *ToolCaller, ui *UI) error {
	source_map, err := NewMapSettings("", caller)
	if err != nil {
		return err
	}

	rowid := g_GetMapTile_cache.FindRow(st.X, st.Y, st.Zoom)
	if rowid < 0 {
		if !source_map.Enable {
			return fmt.Errorf("Map is disabled")
		}

		//download
		var err error
		rowid, err = g_GetMapTile_cache.GetOrDownload(st.X, st.Y, st.Zoom, source_map.Tiles_url, caller)
		if err != nil {
			return err
		}
	}

	//get from database
	row := g_GetMapTile_cache.db.QueryRow("SELECT file FROM tiles WHERE rowid=?", rowid)

	err = row.Scan(&st.Out_image)
	if err != nil {
		return err
	}

	return nil
}

type GetMapTileCache struct {
	db   *sql.DB
	lock sync.Mutex
	rows map[string]int64
}

var g_GetMapTile_cache GetMapTileCache

func (cache *GetMapTileCache) FindRow(x, y, zoom int) int64 {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	name := strconv.Itoa(int(zoom)) + "-" + strconv.Itoa(int(x)) + "-" + strconv.Itoa(int(y)) + ".png"
	rowid, found := cache.rows[name]
	if found {
		return rowid
	}
	return -1
}

var g_download_lock sync.Mutex
var g_flagTimeout = flag.Duration("map_tile - timeout", 30*time.Minute, "HTTP timeout")

func (cache *GetMapTileCache) GetOrDownload(x, y, zoom int, Tiles_url string, caller *ToolCaller) (int64, error) {
	//only once at the time
	g_download_lock.Lock()
	defer g_download_lock.Unlock()

	rowid := cache.FindRow(x, y, zoom)
	if rowid >= 0 {
		return -1, nil //already downloaded
	}

	u := Tiles_url
	u = strings.ReplaceAll(u, "{x}", strconv.Itoa(x))
	u = strings.ReplaceAll(u, "{y}", strconv.Itoa(y))
	u = strings.ReplaceAll(u, "{z}", strconv.Itoa(zoom))

	// prepare client
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return -1, err
	}
	req.Header.Set("User-Agent", "Skyalt/0.1")

	// connect
	client := http.Client{
		Timeout: *g_flagTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	// download
	var out []byte
	temp := make([]byte, 1024*64)
	for {
		n, err := resp.Body.Read(temp)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return -1, err
			}
			break
		}
		//save
		out = append(out, temp[:n]...)

		if !caller.Progress(0.5, fmt.Sprintf("Downloading tile %d-%d-%d", x, y, zoom)) { //procentage ....
			return -1, fmt.Errorf("interrupted")
		}
	}

	//save
	name := strconv.Itoa(zoom) + "-" + strconv.Itoa(x) + "-" + strconv.Itoa(y) + ".png"
	res, err := cache.db.Exec("INSERT INTO tiles(name, file) VALUES(?, ?);", name, out)
	if err != nil {
		return -1, err
	}

	rowid, err = res.LastInsertId()
	if err != nil {
		fmt.Printf("LastInsertId() failed: %v\n", err)
	}

	cache.lock.Lock()
	cache.rows[name] = rowid
	cache.lock.Unlock()

	return rowid, nil
}

func GetMapTile_global_init() error {
	var err error
	g_GetMapTile_cache.db, err = sql.Open("sqlite3", "file:cache.sqlite?&_journal_mode=WAL")
	if err != nil {
		return fmt.Errorf("sql.Open() from file failed: %w", err)
	}

	//init table
	_, err = g_GetMapTile_cache.db.Exec("CREATE TABLE IF NOT EXISTS tiles (name TEXT, file BLOB);")
	if err != nil {
		return fmt.Errorf("'CREATE TABLE' failed: %w", err)
	}

	//load all rows
	rows, err := g_GetMapTile_cache.db.Query("SELECT name, rowid FROM tiles")
	if err != nil {
		return fmt.Errorf("'SELECT name, rowid FROM tiles' failed: %w", err)
	}
	g_GetMapTile_cache.rows = make(map[string]int64)
	for rows.Next() {
		var name string
		var rowid int64
		err = rows.Scan(&name, &rowid)
		if err == nil && rowid > 0 {
			g_GetMapTile_cache.rows[name] = rowid
		}
	}

	return nil
}

func GetMapTile_global_destroy() error {
	g_GetMapTile_cache.db.Exec("PRAGMA wal_checkpoint(full);")
	err := g_GetMapTile_cache.db.Close()
	if err != nil {
		return err
	}
	g_GetMapTile_cache.db = nil

	return nil
}
