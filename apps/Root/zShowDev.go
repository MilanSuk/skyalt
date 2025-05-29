package main

import "fmt"

// [ignore]
type ShowDev struct {
	AppName string
}

func (st *ShowDev) run(caller *ToolCaller, ui *UI) error {
	source_root, err := NewRoot("", caller)
	if err != nil {
		return err
	}

	//refresh apps
	app, err := source_root.refreshApps()
	if err != nil {
		return err
	}

	//sources ....
	//tools ....
	fmt.Println(app)

	return nil
}
