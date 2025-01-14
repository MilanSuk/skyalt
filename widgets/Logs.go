package main

import (
	"fmt"
	"sync"
	"time"
)

type Error struct {
	Time_unix   int64
	Layout_hash uint64
	Error       string
}

type Logs struct {
	lock  sync.Mutex
	Items []Error
}

func (layout *Layout) AddLogs(x, y, w, h int, props *Logs) *Logs {
	layout._createDiv(x, y, w, h, "Logs", props.Build, nil, nil)
	return props
}

func (st *Logs) Build(layout *Layout) {
	st.lock.Lock()
	defer st.lock.Unlock()

	layout.SetColumn(0, 1, 4)
	layout.SetColumn(1, 1, 100)

	bt := layout.AddButtonConfirm(0, 0, 1, 1, "Clear", "Are you sure?")
	bt.confirmed = func() {
		st.Items = nil
	}

	y := 1
	for i := len(st.Items) - 1; i >= 0; i-- {
		it := st.Items[i]
		layout.AddText(0, y, 1, 1, time.Unix(it.Time_unix, 0).Format("2006-01-02 15:04:05"))
		layout.AddText(1, y, 1, 1, it.Error)
		y++
	}

}

func (errs *Logs) AddError(err error, layout_hash uint64) {
	if err == nil {
		return
	}

	fmt.Println("AddError():", err)

	errs.Items = append(errs.Items, Error{Time_unix: time.Now().Unix(), Error: err.Error(), Layout_hash: layout_hash})
}

func (errs *Logs) GetError(last_time_unix int64) *Error {
	if len(errs.Items) > 0 && errs.Items[len(errs.Items)-1].Time_unix > last_time_unix {
		return &errs.Items[len(errs.Items)-1]
	}
	return nil
}
