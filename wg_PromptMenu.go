package main

type PromptMenuIcon struct {
	Path   string
	Blob   []byte
	Margin float64
}

type PromptMenu struct {
	Tooltip string

	Prompts []string
	Icons   []PromptMenuIcon

	changed func()
}

func (layout *Layout) AddPromptMenu(x, y, w, h int, options []string) *PromptMenu {
	props := &PromptMenu{Prompts: options}
	lay := layout._createDiv(x, y, w, h, "prompt_menu", props.Build, nil, nil)
	lay.fnGetLLMTip = props.getLLMTip
	return props
}

func (st *PromptMenu) getLLMTip(layout *Layout) string {
	return Layout_buildLLMTip("PromptMenu", "", "", st.Tooltip)
}

func (st *PromptMenu) Build(layout *Layout) {

	layout.SetColumnFromSub(0, 1, Layout_MAX_SIZE, false)

	cdialog := layout.AddDialog("dialog")

	//button
	bt := layout.AddButton(0, 0, 1, 1, "")
	bt.Tooltip = st.Tooltip
	bt.IconPath = "resources/ai.png"
	bt.Icon_margin = 0.1
	bt.Background = 0.5

	bt.clicked = func() {
		cdialog.OpenRelative(layout.UID)
	}

	//dialog
	{
		cdialog.Layout.SetColumnFromSub(0, 1, 15, true)
		for i, prompt := range st.Prompts {

			bt := cdialog.Layout.AddButton(0, i, 1, 1, prompt)
			bt.Background = 0.25
			bt.Align = 0

			if i < len(st.Icons) {
				bt.IconBlob = st.Icons[i].Blob
				bt.IconPath = st.Icons[i].Path
				bt.Icon_margin = st.Icons[i].Margin
			}

			bt.clicked = func() {
				layout.ui.runPrompt = prompt + "\nUser prompt above is part of: " + st.Tooltip

				if st.changed != nil {
					st.changed()
				}
				cdialog.Close(layout.ui)
			}
		}
	}
}
