package main

import (
	"fmt"
)

type InstallDataSource struct {
	ChatID int
	Name   string //data source name
}

func (st *InstallDataSource) run(caller *ToolCaller, ui *UI) error {
	source_chat, err := NewChat(fmt.Sprintf("Chats/Chat-%d.json", st.ChatID), caller)
	if err != nil {
		return err
	}

	if st.Name == "" {
		return fmt.Errorf("empty Name parameter")
	}

	source_chat.Sources = append(source_chat.Sources, st.Name)

	return nil
}
