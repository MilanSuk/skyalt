package main

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type ChatDashboard struct {
	agent *Agent

	isRunning func() bool
	uiChanged func()
}

func (layout *Layout) AddChatDashboard(x, y, w, h int, props *ChatDashboard) *ChatDashboard {
	layout._createDiv(x, y, w, h, "ChatDashboard", props.Build, nil, nil)
	return props
}

func (st *ChatDashboard) Build(layout *Layout) {
	//layout.SetColumn(0, 1, 100)

	y := 1

	for _, msg := range st.agent.Messages {

		for _, it := range msg.Content {

			if it.Type == "tool_use" {

				switch it.Name {

				case "ui_set_column":
					type ui_set_column struct {
						Index int     //Column's index. Starts with 0.
						Min   float64 //Minimum size of column.
						Max   float64 //Maximum size of column.
					}
					var ui ui_set_column
					js, _ := it.Input.MarshalJSON()
					_ = json.Unmarshal(js, &ui)
					layout.SetColumn(ui.Index, ui.Min, ui.Max)

				case "ui_set_row":
					type ui_set_row struct {
						Index int     //Row's index. Starts with 0.
						Min   float64 //Minimum size of row.
						Max   float64 //Maximum size of row.
					}
					var ui ui_set_row
					js, _ := it.Input.MarshalJSON()
					_ = json.Unmarshal(js, &ui)
					layout.SetRow(ui.Index, ui.Min, ui.Max)

				case "ui_switch":
					type ui_switch struct {
						X, Y, W, H  int
						Description string

						Label string //Description
						Value bool   //Current value
					}
					var ui ui_switch
					js, _ := it.Input.MarshalJSON()
					_ = json.Unmarshal(js, &ui)
					gui := layout.AddSwitch(ui.X, ui.Y, ui.W, ui.H, ui.Label, &ui.Value)
					layout.FindLayout(ui.X, ui.Y, ui.W, ui.H).Enable = (st.isRunning == nil || !st.isRunning())
					gui.changed = func() {
						st.sendChange(ui.Description, OsTrnString(ui.Value, "true", "false"))
					}

				case "ui_editbox":
					type ui_editbox struct {
						X, Y, W, H  int
						Description string

						Value string //Current value
					}
					var ui ui_editbox
					js, _ := it.Input.MarshalJSON()
					_ = json.Unmarshal(js, &ui)
					gui := layout.AddEditbox(ui.X, ui.Y, ui.W, ui.H, &ui.Value)
					layout.FindLayout(ui.X, ui.Y, ui.W, ui.H).Enable = (st.isRunning == nil || !st.isRunning())
					gui.changed = func() {
						st.sendChange(ui.Description, ui.Value)
					}

				case "ui_combo":
					type ui_combo struct {
						X, Y, W, H  int
						Description string

						Value string //Current value

						Values []string //Options to pick value from
						Labels []string //Labels for Values
					}
					var ui ui_combo
					js, _ := it.Input.MarshalJSON()
					_ = json.Unmarshal(js, &ui)

					gui := layout.AddCombo(ui.X, ui.Y, ui.W, ui.H, &ui.Value, ui.Labels, ui.Values)
					layout.FindLayout(ui.X, ui.Y, ui.W, ui.H).Enable = (st.isRunning == nil || !st.isRunning())
					gui.changed = func() {
						st.sendChange(ui.Description, ui.Value)
					}

				case "ui_slider":
					type ui_slider struct {
						X, Y, W, H  int
						Description string

						Min   float64 //Minimum range
						Max   float64 //Maximum range
						Step  float64 //Step size
						Value float64 //Current value
					}
					var ui ui_slider
					js, _ := it.Input.MarshalJSON()
					_ = json.Unmarshal(js, &ui)

					gui := layout.AddSlider(ui.X, ui.Y, ui.W, ui.H, &ui.Value, ui.Min, ui.Max, ui.Step)
					layout.FindLayout(ui.X, ui.Y, ui.W, ui.H).Enable = (st.isRunning == nil || !st.isRunning())
					gui.changed = func() {
						st.sendChange(ui.Description, strconv.FormatFloat(ui.Value, 'f', -1, 64))
					}

				case "ui_map":
					//....

				}
				y++
			}
		}
	}
}

func (st *ChatDashboard) sendChange(param string, new_value string) {
	if st.isRunning != nil && st.isRunning() {
		return
	}

	st.agent.AddUserPromptText(fmt.Sprintf("User just changed '%s' to new value: %s", param, new_value), "")

	// run
	if st.uiChanged != nil {
		st.uiChanged()
	}
}
