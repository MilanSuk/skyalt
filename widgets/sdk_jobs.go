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
)

type Job struct {
	uid    string
	errors []error

	stop atomic.Bool

	title    string
	progress float64 //<0, 1>
	info     string
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
func (job *Job) SetProgress(done float64, info string) {
	job.progress = done
	job.info = info
}

type Jobs struct {
	jobs []*Job
	lock sync.Mutex
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

func StartJob(uid string, title string, fnRun func(job *Job)) *Job {
	g__jobs.lock.Lock()
	defer g__jobs.lock.Unlock()

	//find
	for _, it := range g__jobs.jobs {
		if it.uid == uid {
			return it
		}
	}

	job := &Job{title: title, uid: uid}
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

	//find
	for i := len(st.jobs) - 1; i >= 0; i-- {
		it := st.jobs[i]
		if it.stop.Load() {
			st.jobs = append(st.jobs[:i], st.jobs[i+1:]...) //remove
		}
	}
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
		layout.AddText(0, i*2+0, 1, 1, fmt.Sprintf("%.1f%%: %s", it.progress*100, it.title))
		bt := layout.AddButton(1, i*2+0, 1, 1, NewButtonDanger("Stop"))
		bt.clicked = func() {
			it.Stop()
		}
		layout.AddText(0, i*2+1, 2, 1, it.info)
	}
}
