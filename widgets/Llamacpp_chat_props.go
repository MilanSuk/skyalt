package main

import "strings"

type Llamacpp_completion_props struct {
	Model    string                  `json:"model"`
	Messages []OpenAI_completion_msg `json:"messages"`
	Stream   bool                    `json:"stream"`

	Max_tokens int `json:"max_tokens"`

	Seed              int      `json:"seed"`
	N_predict         int      `json:"n_predict"`
	Temperature       float64  `json:"temperature"`
	Dynatemp_range    float64  `json:"dynatemp_range"`
	Dynatemp_exponent float64  `json:"dynatemp_exponent"`
	Stop              []string `json:"stop"`
	Repeat_last_n     int      `json:"repeat_last_n"`
	Repeat_penalty    float64  `json:"repeat_penalty"`
	Top_k             int      `json:"top_k"`
	Top_p             float64  `json:"top_p"`
	Min_p             float64  `json:"min_p"`
	Tfs_z             float64  `json:"tfs_z"`
	Typical_p         float64  `json:"typical_p"`
	Presence_penalty  float64  `json:"presence_penalty"`
	Frequency_penalty float64  `json:"frequency_penalty"`
	Mirostat          bool     `json:"mirostat"` //not int?
	Mirostat_tau      float64  `json:"mirostat_tau"`
	Mirostat_eta      float64  `json:"mirostat_eta"`
	//Grammar           string   `json:"grammar"` //[]string?
	N_probs int `json:"n_probs"`
	//Image_data //{"data": "<BASE64_STRING>", "id": 12}
	Cache_prompt bool `json:"cache_prompt"`
	Slot_id      int  `json:"slot_id"`

	Response_format *OpenAI_completion_format `json:"response_format"`
}

func (layout *Layout) AddLlamacpp_completion_props(x, y, w, h int, props *Llamacpp_completion_props) *Llamacpp_completion_props {
	layout._createDiv(x, y, w, h, "Llamacpp_completion_props", props.Build, nil, nil)
	return props
}

func (st *Llamacpp_completion_props) Build(layout *Layout) {

	layout.SetColumn(0, 2, 3.5)
	layout.SetColumn(1, 1, 100)

	y := 0

	layout.AddText(0, y, 1, 1, "Model")
	models := OpenFile_Llamacpp().GetModelList()
	layout.AddCombo(1, y, 1, 1, &st.Model, models, models)
	y++

	tx := layout.AddText(0, y, 1, 1, "Streaming")
	tx.Tooltip = "See result as is generated."
	layout.AddSwitch(1, y, 1, 1, "", &st.Stream)
	y++

	sl := layout.AddSliderEdit(0, y, 2, 1, &st.Temperature, 0, 2, 0.1)
	sl.ValuePointerPrec = 1
	sl.Description = "Temperature"
	sl.Tooltip = "The sampling temperature, between 0 and 1. Higher values like 0.8 will make the output more random, while lower values like 0.2 will make it more focused and deterministic. If set to 0, the model will use log probability to automatically increase the temperature until certain thresholds are hit."
	sl.Description_width = 3.5
	sl.Edit_width = 2
	sl.Legend = true
	y++

	sl = layout.AddSliderEdit(0, y, 2, 1, &st.Max_tokens, 128, 4096, 1)
	sl.ValuePointerPrec = 0
	sl.Description = "Max Tokens"
	sl.Tooltip = "The maximum number of tokens that can be generated in the chat completion. The total length of input tokens and generated tokens is limited by the model's context length."
	sl.Description_width = 3.5
	sl.Edit_width = 2
	sl.Legend = true
	y++

	sl = layout.AddSliderEdit(0, y, 2, 1, &st.Top_p, 0, 1, 0.1)
	sl.ValuePointerPrec = 1
	sl.Description = "Top P"
	sl.Tooltip = "An alternative to sampling with temperature, called nucleus sampling, where the model considers the results of the tokens with top_p probability mass. So 0.1 means only the tokens comprising the top 10% probability mass are considered."
	sl.Description_width = 3.5
	sl.Edit_width = 2
	sl.Legend = true
	y++

	sl = layout.AddSliderEdit(0, y, 2, 1, &st.Frequency_penalty, -2, 2, 0.1)
	sl.ValuePointerPrec = 1
	sl.Description = "Frequency Penalty"
	sl.Tooltip = "Number between -2.0 and 2.0. Positive values penalize new tokens based on their existing frequency in the text so far, decreasing the model's likelihood to repeat the same line verbatim."
	sl.Description_width = 3.5
	sl.Edit_width = 2
	sl.Legend = true
	y++

	sl = layout.AddSliderEdit(0, y, 2, 1, &st.Presence_penalty, -2, 2, 0.1)
	sl.ValuePointerPrec = 1
	sl.Description = "Presence Penalty"
	sl.Tooltip = "Number between -2.0 and 2.0. Positive values penalize new tokens based on whether they appear in the text so far, increasing the model's likelihood to talk about new topics."
	sl.Description_width = 3.5
	sl.Edit_width = 2
	sl.Legend = true
	y++

	tx = layout.AddText(0, y, 1, 1, "Seed")
	tx.Tooltip = "Set the random number generator (RNG) seed.  Default: `-1`, which is a random seed."
	layout.AddEditbox(1, y, 1, 1, &st.Seed)
	y++

	tx = layout.AddText(0, y, 1, 1, "Cache prompt")
	tx.Tooltip = "Re-use KV cache from a previous request if possible. This way the common prefix does not have to be re-processed, only the suffix that differs between the requests. Because (depending on the backend) the logits are **not** guaranteed to be bit-for-bit identical for different batch sizes (prompt processing vs. token generation) enabling this option can cause nondeterministic results. Default: `false`"
	layout.AddSwitch(1, y, 1, 1, "", &st.Cache_prompt)
	y++

	tx = layout.AddText(0, y, 1, 1, "Stop")
	tx.Tooltip = "Specify a array(separated by ;) of stopping strings.\nThese words will not be included in the completion, so make sure to add them to the prompt for the next iteration."
	stop := strings.Join(st.Stop, ";")
	stp := layout.AddEditbox(1, y, 1, 1, &stop)
	stp.changed = func() {
		st.Stop = strings.Split(stop, ";")
	}
	y++

	ResetBt := layout.AddButton(0, y, 1, 1, "Reset")
	ResetBt.Background = 0.5
	ResetBt.clicked = func() {
		st.Reset()
	}
	y++
}

func (props *Llamacpp_completion_props) Reset() {
	if props.Model == "" {
		props.Model = OpenFile_Llamacpp().Model
	}

	props.Stream = true
	props.Seed = -1
	props.Max_tokens = 4046
	props.N_predict = 400
	props.Temperature = 0.8
	props.Dynatemp_range = 0.0
	props.Dynatemp_exponent = 1.0
	props.Stop = []string{"</s>", "<|im_start|>", "<|im_end|>", "Llama:", "User:"}
	props.Repeat_last_n = 256
	props.Repeat_penalty = 1.18

	props.Top_k = 40
	props.Top_p = 0.5
	props.Tfs_z = 1.0
	props.Typical_p = 1.0
	props.Presence_penalty = 0.0
	props.Frequency_penalty = 0.0
	props.Mirostat = false
	props.Mirostat_tau = 5.0
	props.Mirostat_eta = 0.1

	//Grammar
	props.N_probs = 0

	//Image_data
	props.Cache_prompt = false
	props.Slot_id = -1

}
