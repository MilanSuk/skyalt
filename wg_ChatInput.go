package main

import (
	"encoding/json"

	"github.com/go-audio/audio"
)

type ChatInput struct {
	Text string

	Files          []string
	FilePickerPath string

	file_name string

	isRunning func() bool
	stop      func()

	sended func()
}

func (layout *Layout) AddChatInput(x, y, w, h int, props *ChatInput) *ChatInput {
	layout._createDiv(x, y, w, h, "ChatInput", props.Build, nil, nil)
	return props
}

func (st *ChatInput) Build(layout *Layout) {
	layout.SetColumn(0, 1, 100)
	layout.SetColumn(1, 3, 3)
	layout.SetRowFromSub(0, 1, 7)

	isRunning := (st.isRunning != nil && st.isRunning())

	sendIt := func() {
		if st.Text != "" && st.sended != nil {
			st.sended()
		}
	}

	{
		layoutIn := layout.AddLayout(0, 0, 1, 1)
		layoutIn.SetColumn(1, 1, 100)
		layoutIn.SetRowFromSub(0, 1, 5)
		layoutIn.SetRowFromSub(1, 1, 2)
		layoutIn.Enable = !isRunning

		{
			mic := layoutIn.AddMicrophone_recorder(0, 0, 1, 1, &Microphone_recorder{UID: "chat_mic_" + st.file_name})
			mic.done = func(out audio.IntBuffer, startUnix float64) {

				props := Agent_findSTTAgentProperties("main")
				if props != nil {
					login, whispercpp, _ := FindLoginSTTModel(props.Model)

					blob, err := Whispercpp_convertIntoFile(&out, login != nil)
					if err != nil {
						//.....
						return
					}

					var js []byte
					if login != nil {
						api_key, err := g_agent_passwords.Find(login.Api_key_id)
						if err != nil {
							//...
							return
						}

						js, _, err = OpenAI_stt_Run(props.Model, "blob.mp3", blob, props.Temperature, props.Response_format, login.OpenAI_completion_url, api_key)
						if err != nil {
							//...
							return
						}
					}

					if whispercpp != nil {
						js, _, err = Whispercpp_Transcribe(blob, Whispercpp_props{Model: props.Model, Temperature: props.Temperature, Response_format: props.Response_format}, OpenFile_Whispercpp())
						if err != nil {
							//...
							return
						}
					}

					st.SetVoice(js, startUnix)
					Layout_RefreshDelayed()
				}
			}

			text := layoutIn.AddEditboxMultiline(1, 0, 1, 1, &st.Text)
			text.enter = sendIt
		}

		//image(s)
		{
			FileDia := layoutIn.AddDialog("file")
			FileDia.Layout.SetColumn(0, 5, 20)
			FileDia.Layout.SetRow(0, 5, 10)
			pk := FileDia.Layout.AddFilePicker(0, 0, 1, 1, &st.FilePickerPath, true)
			pk.changed = func(close bool) {
				if close {
					st.Files = append(st.Files, st.FilePickerPath)
					st.FilePickerPath = "" //reset
					FileDia.Close()
				}
			}

			ImgsList := layoutIn.AddLayoutList(0, 1, 2, 1, true)
			ImgsList.dropFile = func(path string) {
				st.Files = append(st.Files, path)
			}

			for fi, file := range st.Files {
				ImgDia := layout.AddDialog("image_" + file)
				ImgDia.Layout.SetColumn(0, 5, 12)
				ImgDia.Layout.SetColumn(1, 3, 3)
				ImgDia.Layout.SetRow(1, 5, 15)
				ImgDia.Layout.AddImage(0, 1, 2, 1, file)
				RemoveBt := ImgDia.Layout.AddButton(1, 0, 1, 1, "Remove")
				RemoveBt.clicked = func() {
					st.Files = append(st.Files[:fi], st.Files[fi+1:]...) //remove
					ImgDia.Close()
				}

				imgLay := ImgsList.AddListSubItem()
				imgLay.SetColumn(0, 2, 2)
				imgLay.SetRow(0, 2, 2)
				imgBt := imgLay.AddButtonIcon(0, 0, 1, 1, file, 0, file)
				imgBt.Background = 0
				imgBt.Cd = Paint_GetPalette().B
				imgBt.Border = true
				imgBt.clicked = func() {
					ImgDia.OpenRelative(imgLay)
				}
			}

			addImgLay := ImgsList.AddListSubItem()
			AddImgBt := addImgLay.AddButton(0, 0, 1, 1, "+")
			AddImgBt.clicked = func() {
				FileDia.OpenRelative(addImgLay)
			}
		}
	}

	if !isRunning {
		SendBt := layout.AddButton(1, 0, 1, 1, "Send")
		SendBt.clicked = sendIt
	} else {
		StopBt := layout.AddButton(1, 0, 1, 1, "Stop")
		StopBt.Cd = Paint_GetPalette().E
		StopBt.clicked = func() {
			if st.stop != nil {
				st.stop()
			}
		}
	}
}

func (st *ChatInput) reset() {
	st.Text = ""
	st.Files = nil
}

func (st *ChatInput) SetVoice(js []byte, voiceStart_sec float64) {
	type VerboseJsonWord struct {
		Start, End float64
		Word       string
	}
	type VerboseJsonSegment struct {
		Start, End float64
		Text       string
		Words      []*VerboseJsonWord
	}
	type VerboseJson struct {
		Segments []*VerboseJsonSegment //later are projected to .Words
	}

	var verb VerboseJson
	err := json.Unmarshal(js, &verb)
	if err != nil {
		return
	}
	for _, seg := range verb.Segments {
		if len(seg.Words) == 0 {
			//copy whole segment .Text as single word, because Groq doesn't support timestamp_granularities[]
			seg.Words = append(seg.Words, &VerboseJsonWord{Start: seg.Start, End: seg.End, Word: seg.Text})
		}
	}

	//jump over older picks
	/*pick_i := 0
	for _, pk := range ast.Picks {
		if pk.Time_sec < voiceStart_sec {
			pick_i++
		}
	}*/

	//build prompt
	prompt := ""
	for _, seg := range verb.Segments {
		for _, w := range seg.Words {
			/*for pick_i < len(ast.Picks) && ast.Picks[pick_i].Time_sec < (voiceStart_sec+w.Start) { //for(!)
				prompt += _getMark(pick_i)
				//st.chatVoice_items = st.chatVoice_items[1:]
				pick_i++
			}*/
			prompt += w.Word
		}
	}

	//ast.Prompt = prompt

	st.Text += prompt
}
