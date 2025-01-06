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
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Job struct {
	uid    string
	errors []error

	stop atomic.Bool

	title string
	done  float64 //<0, 1>
	info  string

	start_time_sec   float64
	estimate_end_sec float64 //0 = off

	//last_done float64
	//last_time float64 //sec when SetProgress() was called last
	//rest_time float64 //sec to finish job
}

func (job *Job) Stop() {
	job.stop.Store(true)
}

func (job *Job) IsRunning() bool {
	return !job.stop.Load()
}

func (job *Job) AddError(err error) {
	if err != nil {
		job.errors = append(job.errors, err)

		OpenFile_Logs().AddError(err, 0)
	}
}

func (job *Job) SetInfo(info string) {
	job.info = info
}

func (job *Job) SetProgress(done float64) {
	job.done = done

	act := float64(time.Now().UnixMilli()) / 1000
	job.estimate_end_sec = job.start_time_sec + ((act - job.start_time_sec) / done * (1 - done))

	/*time := float64(time.Now().UnixMilli()) / 1000
	dt := time - job.last_time
	if job.last_time > 0 && dt > 0.5 { //need 0.5sec different
		rest_done := (1 - job.last_done)
		diff_done := (done - job.last_done)

		job.rest_time = rest_done / diff_done * dt
		job.last_time = time
		job.last_done = done
	}*/
}

func (job *Job) SetEstimate(end_sec float64) {
	job.estimate_end_sec = end_sec
}

func (job *Job) GetEstimateDone() float64 {
	act := float64(time.Now().UnixMilli()) / 1000

	return (act - job.start_time_sec) / (job.estimate_end_sec - job.start_time_sec)
}

type Jobs struct {
	jobs []*Job
	lock sync.Mutex

	need_refresh bool
}

var g__jobs Jobs

func FindJob(uid string) *Job {
	g__jobs.lock.Lock()
	defer g__jobs.lock.Unlock()

	//find
	for _, it := range g__jobs.jobs {
		if it.uid == uid {
			return it
		}
	}
	return nil
}

func (jobs *Jobs) GetSlowestEstimateJob() *Job {
	g__jobs.lock.Lock()
	defer g__jobs.lock.Unlock()

	//find
	end_time := 0.0
	var end_job *Job
	for _, it := range g__jobs.jobs {
		if it.estimate_end_sec > end_time {
			end_time = it.estimate_end_sec
			end_job = it
		}
	}
	return end_job
}

func StartJob(uid string, title string, fnRun func(job *Job)) *Job {
	g__jobs.lock.Lock()
	defer g__jobs.lock.Unlock()

	//find
	for _, it := range g__jobs.jobs {
		if it.uid == uid {
			return it
		}
	}

	job := &Job{title: title, uid: uid, start_time_sec: float64(time.Now().UnixMilli()) / 1000}
	g__jobs.jobs = append(g__jobs.jobs, job)
	go func() {
		fnRun(job)
		job.stop.Store(true) //done
	}()
	return job
}

func (st *Jobs) maintenance() {
	st.lock.Lock()
	defer st.lock.Unlock()

	had_job := (len(st.jobs) > 0)

	//find
	for i := len(st.jobs) - 1; i >= 0; i-- {
		it := st.jobs[i]
		if it.stop.Load() {
			st.jobs = append(st.jobs[:i], st.jobs[i+1:]...) //remove
		}
	}

	if !st.need_refresh {
		//just finished, one more refresh
		st.need_refresh = had_job && (len(st.jobs) == 0)
	}
}

func (st *Jobs) NeedRefresh() bool {
	st.lock.Lock()
	defer st.lock.Unlock()

	if len(g__jobs.jobs) > 0 {
		return true
	}

	if st.need_refresh {
		st.need_refresh = false
		return true
	}

	return false
}

func (layout *Layout) AddJobs(x, y, w, h int) {
	layout._createDiv(x, y, w, h, "Jobs", g__jobs.Build, nil, nil)
}

func (st *Jobs) Build(layout *Layout) {
	st.lock.Lock()
	defer st.lock.Unlock()

	layout.SetColumn(0, 1, 100)
	layout.SetColumn(1, 1, 4)

	for i, it := range st.jobs {
		layout.AddText(0, i*2+0, 1, 1, fmt.Sprintf("%.1f%%: %s", it.done*100, it.title))
		bt := layout.AddButtonDanger(1, i*2+0, 1, 1, "Stop")
		bt.clicked = func() {
			it.Stop()
		}
		layout.AddText(0, i*2+1, 2, 1, it.info)
	}
}
