package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type ButtonDownload struct {
	Agent string
	Path  string
	Url   string
	done  func()

	stat_recv int64 //smazat ...
	stat_time float64
}

func (layout *Layout) AddButtonDownload(x, y, w, h int, path string, url string) *ButtonDownload {
	props := &ButtonDownload{Path: path, Url: url}
	layout._createDiv(x, y, w, h, "ButtonDownload", props.Build, nil, nil)
	return props
}

func (st *ButtonDownload) Build(layout *Layout) {

	layout.SetRow(0, 1, 100)

	job := FindJob(st.getUID())

	if job != nil {
		//downloading active
		layout.SetColumn(0, 1, 100)
		layout.SetColumn(1, 2, 2)

		layout.AddProgressBar(0, 0, 1, 1, job.progress)
		layout.AddText(0, 0, 1, 1, job.info)

		pauseBt := layout.AddButton(1, 0, 1, 1, NewButton("Pause"))
		pauseBt.clicked = func() {
			job.Stop()
		}

	} else if ButtonDownload_FileExists(st.GetTempPath()) {
		//downloading paused - no net_service & <path>.temp exist
		layout.SetColumn(0, 1, 100)
		layout.SetColumn(1, 2, 2)

		resumeBt := layout.AddButton(0, 0, 1, 1, NewButton("Resume"))
		resumeBt.Tooltip = st.Url
		resumeBt.clicked = func() {
			st.Start()
		}

		deleteBt := layout.AddButtonConfirm(1, 0, 1, 1, "Delete", "Are you sure?")
		deleteBt.Tooltip = st.GetTempPath()
		deleteBt.confirmed = func() {
			os.Remove(st.GetTempPath())
		}
	} else if ButtonDownload_FileExists(st.Path) {
		//delete - file fully downloaded
		layout.SetColumn(0, 1, 100)
		bt := layout.AddButtonConfirm(0, 0, 1, 1, "Delete", "Are you sure?")
		bt.Draw_back = 0.1
		bt.Draw_border = true
		bt.Tooltip = st.Path
		bt.confirmed = func() {
			os.Remove(st.Path)
		}
	} else {
		//download - file not exist
		layout.SetColumn(0, 1, 100)

		bt := layout.AddButton(0, 0, 1, 1, NewButton("Download"))
		bt.Tooltip = st.Url
		bt.clicked = func() {
			st.Start()
		}
	}
}

func (st *ButtonDownload) getUID() string {
	return fmt.Sprintf("ButtonDownload:%s-%s-%s", st.Path, st.Url, st.Agent)
}

func (st *ButtonDownload) Start() *Job {
	return StartJob(st.getUID(), fmt.Sprintf("Downloading into %s", st.Path), st.Run)
}
func (st *ButtonDownload) Stop() {
	job := FindJob(st.getUID())
	if job != nil {
		job.Stop()
	}
}

func ButtonDownload_FileExists(fileName string) bool {
	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

var g_SAServiceNet_flagTimeout = flag.Duration("timeout", 30*time.Minute, "HTTP timeout")

func (st *ButtonDownload) Run(job *Job) {
	//loop
	temp_path := st.Path + ".download"

	//prepare temp file
	flag := os.O_CREATE | os.O_WRONLY
	if OsFileExists(temp_path) {
		flag = os.O_APPEND | os.O_WRONLY
	}
	var err error
	file, err := os.OpenFile(temp_path, flag, 0666)
	if err != nil {
		job.AddError(err)
		return
	}

	// prepare client
	req, err := http.NewRequest("GET", st.Url, nil)
	if err != nil {
		if file != nil {
			file.Close()
		}
		job.AddError(err)
		return
	}

	if st.Agent != "" {
		req.Header.Set("User-Agent", st.Agent)
	}

	// resume download
	file_bytes := int64(0)
	if file != nil {
		var err error
		file_bytes, err = file.Seek(0, io.SeekEnd)
		if err != nil {
			file.Close()
			job.AddError(err)
			return
		}
	}
	if file_bytes > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", file_bytes)) //https://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html
	}

	// connect
	client := http.Client{
		Timeout: *g_SAServiceNet_flagTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		if file != nil {
			file.Close()
		}
		job.AddError(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		if file != nil {
			file.Close()
		}
		job.AddError(errors.New(resp.Status))
		return
	}
	recv_bytes := file_bytes
	final_bytes := file_bytes + resp.ContentLength

	// Loop
	var out_bytes []byte
	data := make([]byte, 1024*64)
	for job.IsRunning() {
		//download
		n, err := resp.Body.Read(data)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				job.AddError(err)
			}
			break
		}
		//save
		var m int
		if file != nil {
			m, err = file.Write(data[:n])
			if err != nil {
				job.AddError(err)
				break
			}
		} else {
			out_bytes = append(out_bytes, data[:n]...)
			m = n
		}

		recv_bytes += int64(m)
		st.stat_recv += int64(m)

		job.SetProgress(st._getProgress(recv_bytes, final_bytes), st._getProgressStr(recv_bytes, final_bytes))
	}

	if file != nil {
		file.Close()
	}

	if file != nil && recv_bytes == final_bytes {
		OsFileRename(temp_path, st.Path) //<name>.temp -> <name>
	}

	if recv_bytes != final_bytes {
		job.AddError(fmt.Errorf("downloading not finished: Received %dB of %dB", recv_bytes, final_bytes))
		return
	}

	if st.done != nil {
		st.done()
	}
}

func (st *ButtonDownload) GetTempPath() string {
	return st.Path + ".download"
}

func (st *ButtonDownload) _getProgress(recv_bytes, final_bytes int64) float64 {
	if final_bytes > 0 {
		return float64(recv_bytes) / float64(final_bytes)
	}
	return 0
}

func (st *ButtonDownload) _getProgressStr(recv_bytes, final_bytes int64) string {
	speed := st._getAvgRecvBytesPerSec()

	remain_sec := 0
	if speed > 0 {
		remain_sec = int(float64(final_bytes-recv_bytes) / speed)
	}

	now := time.Now()
	predict := now.Add(time.Duration(remain_sec) * time.Second)
	diff := predict.Sub(now)

	return fmt.Sprintf("%.1f%%, %s, %s/s %v", st._getProgress(recv_bytes, final_bytes)*100, OsFormatBytes2(int(recv_bytes), int(final_bytes)), OsFormatBytes(int(speed)), diff)
}

func (st *ButtonDownload) _getAvgRecvBytesPerSec() float64 {
	act_time := OsTime()

	old_time := st.stat_time
	bytes := st.stat_recv

	if (act_time - st.stat_time) > 3 {
		//reset
		st.stat_time = act_time
		st.stat_recv = 0
	}

	return float64(bytes) / (act_time - old_time)
}
