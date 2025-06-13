/*
Copyright 2025 Milan Suk

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this db except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type ToolsAppProcess struct {
	Compile *ToolsAppCompile

	port       int
	cmd        *exec.Cmd
	cmd_exited bool
	cmd_error  string
}

func NewToolsAppRun(appName string) *ToolsAppProcess {
	app := &ToolsAppProcess{}

	app.Compile = NewToolsAppCompile(appName)

	return app
}

func (app *ToolsAppProcess) IsRunning() bool {
	return app.cmd != nil && !app.cmd_exited
}

func (app *ToolsAppProcess) Destroy(waitTillEnd bool) error {
	if app.IsRunning() {
		cl, err := NewToolsClient("localhost", app.port)
		if err != nil {
			return err
		}
		err = cl.WriteArray([]byte("exit"))
		if err != nil {
			return err
		}
	}

	if waitTillEnd {
		if !app.cmd_exited {
			app.WaitUntilExited()
		}
	}

	return nil
}

func (app *ToolsAppProcess) WaitUntilExited() string {
	n := 0
	for n < 100 && !app.cmd_exited {
		time.Sleep(10 * time.Millisecond)
		n++
	}
	return app.cmd_error
}

func (app *ToolsAppProcess) CheckRun(router *ToolsRouter) error {
	if !app.IsRunning() {

		if app.cmd_exited {
			app.WaitUntilExited()
		}

		app.cmd_exited = false
		app.cmd_error = ""
		app.port = 0
		app.cmd = nil

		//start
		cmd := exec.Command("./"+app.Compile.GetBinName(), app.Compile.appName, strconv.Itoa(router.server.port))
		cmd.Dir = app.Compile.GetFolderPath()
		OutStr := new(strings.Builder)
		ErrStr := new(strings.Builder)
		cmd.Stdout = OutStr
		cmd.Stderr = ErrStr
		err := cmd.Start()
		if err != nil {
			return fmt.Errorf("'%s' start failed: %w", app.Compile.GetFolderPath(), err)
		}
		app.cmd = cmd //running

		fmt.Printf("App '%s' has started\n", app.Compile.GetFolderPath())

		//run tool
		go func() {
			app.cmd.Wait()

			if OutStr.Len() > 0 {
				fmt.Printf("'%s' app output: %s\n", app.Compile.GetFolderPath(), OutStr.String())
			}
			if ErrStr.Len() > 0 {
				fmt.Printf("\033[31m'%s' app error:%s\033[0m\n", app.Compile.GetFolderPath(), ErrStr.String())
			}

			wd, _ := os.Getwd()
			app.cmd_error = strings.ReplaceAll(ErrStr.String(), wd, "")
			app.cmd_exited = true
			app.cmd = nil
		}()

		//wait one second for recv a port
		{
			n := 0
			for n < 100 && app.port == 0 {
				time.Sleep(10 * time.Millisecond)
				n++
			}
			if app.port == 0 {
				fmt.Printf("'%s' app process hasn't connected in time\n", app.Compile.GetFolderPath())
			}
		}

	}

	return nil //ok
}
