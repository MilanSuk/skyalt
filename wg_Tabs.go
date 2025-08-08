package main

type Tabs struct {
	Labels []string
	Values []string
	Value  *string

	changed func()
}

func (layout *Layout) AddTabs(x, y, w, h int, value *string, labels []string, values []string) *Tabs {
	props := &Tabs{Value: value, Labels: labels, Values: values}
	layout._createDiv(x, y, w, h, "Tabs", props.Build, nil, nil)
	return props
}

func (st *Tabs) Build(layout *Layout) {

	layout.Back_cd = layout.GetPalette().GetGrey(0.1)

	layout.scrollV.Show = false
	layout.scrollH.Narrow = true

	for i, val := range st.Values {
		this_val := val

		//check
		if i >= len(st.Labels) {
			break
		}

		//button
		layout.SetColumn(i*2+0, 1, Layout_MAX_SIZE)
		bt := layout.AddButton(i*2+0, 0, 1, 1, st.Labels[i])
		bt.Tooltip = val
		if *st.Value == val {
			bt.Background = 1
		} else {
			bt.Background = 0.1
		}
		bt.clicked = func() {
			*st.Value = this_val
			if st.changed != nil {
				st.changed()
			}
		}

		if i+1 < len(st.Values) {
			//divider
			layout.SetColumn(i*2+1, 0.2, 0.2)
			layout.AddDivider(i*2+1, 0, 1, 1, false)
		}
	}
}
