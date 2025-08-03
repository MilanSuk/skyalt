package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Cache struct {
	lock sync.Mutex
}

func NewCache() *Cache {
	cache := &Cache{}
	os.MkdirAll(filepath.Join("temp", "media"), os.ModePerm)

	return cache
}

func (cache *Cache) Destroy() {
	cache.lock.Lock()
	defer cache.lock.Unlock()
}

func (cache *Cache) Get(url string) (string, error) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	var path string
	{
		id := sha256.Sum256([]byte(url))
		ext, _ := getFileExtensionFromUrl(url)
		path = hex.EncodeToString(id[:])
		if ext != "" {
			path += "." + ext
		}
	}

	path = filepath.Join("temp", "media", path)

	if !isFileExists(path) {
		//download
		err := cache.download(url, path)
		if err != nil {
			return "", err
		}
	}

	return path, nil
}

var g_flagTimeout = flag.Duration("map_tile - timeout", 30*time.Minute, "HTTP timeout")

func (cache *Cache) download(url string, dst_path string) error {

	// prepare client
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Skyalt/0.1") //? ....

	// connect
	client := http.Client{
		Timeout: *g_flagTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	//create file
	file, err := os.Create(dst_path)
	if err != nil {
		return err
	}
	defer file.Close()

	//download
	buff := make([]byte, 4*1024)
	for {
		n, err := resp.Body.Read(buff)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return err
			}
			break
		}
		//save
		nn, err := file.Write(buff[:n])
		if err != nil {
			return err
		}
		if n != nn {
			return errors.New("write failed")
		}
	}

	return nil
}
