package main

// Search the web.
type web_search struct {
	Prompt string //Search input.
}

func (st *web_search) run() string {

	SystemPrompt := "Be precise and concise."

	answer := SDK_RunAgent("search", 20, 20000, SystemPrompt, st.Prompt)

	return answer
}
