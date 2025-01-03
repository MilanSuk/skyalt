package main

type UserInfo struct {
	Gender string  //"male" or "female"
	Born   int64   //unix time
	Height float64 //meters
	Weight float64 //kilograms
}

func (layout *Layout) AddUserInfo(x, y, w, h int, props *UserInfo) *UserInfo {
	layout._createDiv(x, y, w, h, "UserInfo", props.Build, nil, nil)
	return props
}

var g_UserInfo *UserInfo

func OpenFile_UserInfo() *UserInfo {
	if g_UserInfo == nil {
		g_UserInfo = &UserInfo{Born: 946681200, Gender: "male", Height: 1.7, Weight: 60}
		_read_file("UserInfo-UserInfo", g_UserInfo)
	}
	return g_UserInfo
}

func (st *UserInfo) Build(layout *Layout) {
	layout.SetColumn(0, 3, 3)
	layout.SetColumn(1, 1, 100)

	y := 0

	layout.AddText(0, y, 1, 1, "Gender")
	cb := layout.AddCombo(1, y, 1, 1, &st.Gender, []string{"Male", "Female"}, []string{"male", "female"})
	cb.DialogWidth = 4
	y++

	layout.AddText(0, y, 1, 1, "Height(meters)")
	layout.AddEditboxFloat(1, y, 1, 1, &st.Height, 2)
	y++

	layout.AddText(0, y, 1, 1, "Weight(kg)")
	layout.AddEditboxFloat(1, y, 1, 1, &st.Weight, 2)
	y++

	layout.AddText(0, y, 1, 1, "Born")
	layout.AddDatePickerButton(1, y, 1, 1, &st.Born, nil)
	y++
}
