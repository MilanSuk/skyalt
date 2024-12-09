package main

import "time"

func (st *Logs) Build() {
	st.lock.Lock()
	defer st.lock.Unlock()

	st.layout.SetColumn(0, 1, 4)
	st.layout.SetColumn(1, 1, 100)

	bt := st.layout.AddButtonConfirm(0, 0, 1, 1, "Clear", "Are you sure?")
	bt.Confirmed = func() {
		st.Items = nil
	}

	y := 1
	for _, e := range st.Items {
		st.layout.AddText(0, y, 1, 1, time.Unix(e.Time_unix, 0).Format("2006-01-02 15:04:05"))
		st.layout.AddText(1, y, 1, 1, e.Error)
		y++
	}

}

func (errs *Logs) AddError(err error, layout_hash uint64) {
	if err == nil {
		return
	}

	errs.Items = append(errs.Items, Error{Time_unix: time.Now().Unix(), Error: err.Error(), Layout_hash: layout_hash})
}

func (errs *Logs) GetError(last_time_unix int64) *Error {
	if len(errs.Items) > 0 && errs.Items[len(errs.Items)-1].Time_unix > last_time_unix {
		return &errs.Items[len(errs.Items)-1]
	}

	return nil
}
