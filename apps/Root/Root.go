package main

import (
	"os"
	"path/filepath"
	"slices"
)

type RootChat struct {
	FileName string
	Label    string
}

func (app *RootApp) GetFolderPath() string {
	return filepath.Join("..", app.Name)
}

type RootDev struct {
	Enable bool

	PromptsHistory []string

	ShowSide bool
	SideFile string //Name.go
	MainMode string //"prompts", "secrets"
	SideMode string //"schema", "code", "msg"
}

type RootApp struct {
	Name            string
	Chats           []RootChat
	Selected_chat_i int
	Dev             RootDev
}

// Root
type Root struct {
	ShowSettings bool
	Memory       string

	Autosend float64

	Apps           []*RootApp
	Selected_app_i int

	Last_log_time float64
}

func NewRoot(file string) (*Root, error) {
	st := &Root{}

	return LoadFile(file, "Root", "json", st, true)
}

func (root *Root) IsAppExist(name string) bool {
	_, err := os.Stat(filepath.Join("..", name))
	return err == nil && !os.IsNotExist(err)
}

func (root *Root) refreshApps() (*RootApp, error) {
	fls, err := os.ReadDir("..")
	if err != nil {
		return nil, err
	}
	//add new apps
	for _, fl := range fls {
		if !fl.IsDir() || fl.Name() == "Root" {
			continue
		}

		found := false
		for _, app := range root.Apps {
			if app.Name == fl.Name() {
				found = true
				break
			}
		}
		if !found {
			root.Apps = append(root.Apps, &RootApp{Name: fl.Name()})
		}
	}
	//remove deleted app
	for i := len(root.Apps) - 1; i >= 0; i-- {
		found := false
		for _, fl := range fls {
			if !fl.IsDir() || fl.Name() == "Root" {
				continue
			}

			if fl.Name() == root.Apps[i].Name {
				found = true
				break
			}
		}
		if !found {
			root.Apps = slices.Delete(root.Apps, i, i+1)
		}
	}

	//check select in range
	if root.Selected_app_i >= 0 {
		if root.Selected_app_i >= len(root.Apps) {
			root.Selected_app_i = len(root.Apps) - 1
		}
	}
	//return
	if root.Selected_app_i >= 0 {
		return root.Apps[root.Selected_app_i], nil
	}

	return nil, nil
}

func (app *RootApp) refreshChats(caller *ToolCaller) (*Chat, string, error) {

	chats_folder := filepath.Join("..", app.Name, "Chats")
	if _, err := os.Stat(chats_folder); os.IsNotExist(err) {
		//no chat folder
		app.Chats = nil
		return nil, "", nil //ok
	}

	fls, err := os.ReadDir(chats_folder)
	if err != nil {
		return nil, "", nil //maybe no chat
	}
	//add new chats
	for _, fl := range fls {
		if fl.IsDir() {
			continue
		}

		found := false
		for _, chat := range app.Chats {
			if chat.FileName == fl.Name() {
				found = true
				break
			}
		}
		if !found {
			app.Chats = append(app.Chats, RootChat{FileName: fl.Name()})
		}
	}
	//remove deleted chats
	for i := len(app.Chats) - 1; i >= 0; i-- {
		found := false
		for _, fl := range fls {
			if fl.IsDir() {
				continue
			}

			if fl.Name() == app.Chats[i].FileName {
				found = true
				break
			}
		}
		if !found {
			app.Chats = slices.Delete(app.Chats, i, i+1)
		}
	}

	//check selecte in range
	if app.Selected_chat_i >= 0 {
		if app.Selected_chat_i >= len(app.Chats) {
			app.Selected_chat_i = len(app.Chats) - 1
		}
	}

	//update and return
	if app.Selected_chat_i >= 0 {
		fileName := app.Chats[app.Selected_chat_i].FileName
		sourceChat, err := NewChat(filepath.Join("..", app.Name, "Chats", fileName))
		if err != nil {
			return nil, "", err
		}

		if sourceChat != nil {
			//reload
			app.Chats[app.Selected_chat_i].Label = sourceChat.Label
		}

		return sourceChat, fileName, nil
	}

	return nil, "", nil
}
