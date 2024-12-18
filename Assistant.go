package main

import (
	"encoding/json"
	"fmt"
)

type Assistant struct {
	Show bool

	Prompt string
	Picks  []LayoutPick

	TempPrompt AssistantPrompt

	Chat OpenAI_chat

	AutoSend float64
}

func (layout *Layout) AddAssistant(x, y, w, h int, props *Assistant) *Assistant {
	layout._createDiv(x, y, w, h, "Assistant", props.Build, nil, nil)
	return props
}

var g_Assistant *Assistant

func NewFile_Assistant() *Assistant {
	if g_Assistant == nil {
		g_Assistant = &Assistant{}
		_read_file("Assistant-Assistant", g_Assistant)
	}
	return g_Assistant
}

func (st *Assistant) Build(layout *Layout) {
	//...
}

func (st *Assistant) SetVoice(js []byte, voiceStart_sec float64) {
	type VerboseJsonWord struct {
		Start, End float64
		Word       string
	}
	type VerboseJsonSegment struct {
		Start, End float64
		Text       string
		Words      []VerboseJsonWord
	}
	type VerboseJson struct {
		Segments []VerboseJsonSegment
	}

	var verb VerboseJson
	err := json.Unmarshal(js, &verb)
	if err != nil {
		return
	}

	// jump over older picks
	pick_i := 0
	for _, pk := range st.Picks {
		if pk.Time_sec < voiceStart_sec {
			pick_i++
		}
	}
	// build prompt
	prompt := ""
	for _, seg := range verb.Segments {
		for _, w := range seg.Words {
			for pick_i < len(st.Picks) && st.Picks[pick_i].Time_sec < (voiceStart_sec+w.Start) { //for(!)
				prompt += _Assistant_getMark(pick_i)
				//st.chatVoice_items = st.chatVoice_items[1:]
				pick_i++
			}
			prompt += w.Word
		}
	}

	st.Prompt = prompt
}

func (st *Assistant) Send(dialog *Layout) {
	//...
}

func _Assistant_getMark(i int) string {
	if i < 25 { //25 = 'Z'-'A'
		return fmt.Sprintf("{%c}", 'A'+i)
	} else {
		return fmt.Sprintf("{%c}", '0'+(i-25))
	}
}
