package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// [ignore]
type DownloadFile struct {
	Agent string
	Url   string
	Path  string //File path
}

var g_DownloadFile_flagTimeout = flag.Duration("download_file - timeout", 30*time.Minute, "HTTP timeout")

func (st *DownloadFile) run(caller *ToolCaller, ui *UI) error {

	if _isFileExists(st.Path) {
		return fmt.Errorf("file '%s' already exists", st.Path)
	}

	file, found := _registerFile(st.Url, st.Path, st.Agent)

	if !found {
		go func() {
			file.Download()
			_unregisterFile(st.Url, st.Path, st.Agent)
		}()
	}

	for !file.Done.Load() {
		if !caller.Progress(file.Progress, "Downloading") {
			file.Stop.Store(true)
		}
	}

	return file.Err
}

type File struct {
	Agent string
	Url   string
	Path  string

	Stop atomic.Bool
	Done atomic.Bool

	Err error
	//Done atomic.Bool

	Progress float64
}

var g_DownloadFile_files map[string]*File //[url]
var g_DownloadFile_files_lock sync.Mutex

func _registerFile(url string, path string, agent string) (file *File, found bool) {
	g_DownloadFile_files_lock.Lock()
	defer g_DownloadFile_files_lock.Unlock()

	if g_DownloadFile_files == nil {
		g_DownloadFile_files = make(map[string]*File)
	}

	uid := url + path
	file, found = g_DownloadFile_files[uid]
	if !found {
		file = &File{Url: url, Path: path, Agent: agent}
		g_DownloadFile_files[uid] = file
	}
	return
}

func _unregisterFile(url string, path string, agent string) {
	g_DownloadFile_files_lock.Lock()
	defer g_DownloadFile_files_lock.Unlock()

	uid := url + path
	delete(g_DownloadFile_files, uid)
}

func (file *File) Download() {

	defer file.Done.Store(true)

	temp_path := file.Path + ".download"

	//prepare temp file
	flag := os.O_CREATE | os.O_WRONLY
	if _isFileExists(temp_path) {
		flag = os.O_APPEND | os.O_WRONLY
	}
	var err error
	fl, err := os.OpenFile(temp_path, flag, 0666)
	if err != nil {
		file.Err = err
		return
	}

	// prepare client
	req, err := http.NewRequest("GET", file.Url, nil)
	if err != nil {
		fl.Close()
		file.Err = err
		return
	}

	if file.Agent != "" {
		req.Header.Set("User-Agent", file.Agent)
	}

	// resume download
	file_bytes := int64(0)
	file_bytes, err = fl.Seek(0, io.SeekEnd)
	if err != nil {
		fl.Close()
		file.Err = err
		return
	}
	if file_bytes > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", file_bytes)) //https://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html
	}

	// connect
	client := http.Client{
		Timeout: *g_DownloadFile_flagTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		fl.Close()
		file.Err = err
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		fl.Close()
		file.Err = errors.New(resp.Status)
		return
	}
	recv_bytes := file_bytes
	final_bytes := file_bytes + resp.ContentLength

	// Loop
	var out_bytes []byte
	data := make([]byte, 1024*64)
	for !file.Stop.Load() {
		//download
		n, err := resp.Body.Read(data)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				fl.Close()
				file.Err = err
				return
			}
			break //finished
		}

		//save
		var m int
		if fl != nil {
			m, err = fl.Write(data[:n])
			if err != nil {
				fl.Close()
				file.Err = err
				return
			}
		} else {
			out_bytes = append(out_bytes, data[:n]...)
			m = n
		}

		recv_bytes += int64(m)

		file.Progress = float64(recv_bytes) / float64(final_bytes)
	}

	err = fl.Close()
	if err != nil {
		file.Err = err
		return
	}

	if recv_bytes != final_bytes {
		file.Err = fmt.Errorf("downloading not finished: Received %dB of %dB", recv_bytes, final_bytes)
		return
	}

	err = os.Rename(temp_path, file.Path) //<name>.temp -> <name>
	if err != nil {
		file.Err = err
		return
	}

}

func _isFileExists(fileName string) bool {
	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
