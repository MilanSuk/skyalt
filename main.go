/*
Copyright 2023 Milan Suk

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
	"context"
	"fmt"
)

const SKYALT_LOGO = "resources/logo.png"

func main() {
	InitImageGlobal()
	err := InitSDLGlobal()
	if err != nil {
		fmt.Printf("InitSDLGlobal() failed: %v\n", err)
		return
	}

	defer DestroySDLGlobal()

	InitDbGlobal()

	// Server
	/*var server *Server
	server, err = NewServer("server", 7000, "ssl")
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(250 * time.Millisecond) //give 'server thread' time to start
	defer server.Destroy()*/

	ctx := context.Background()
	root, err := NewRoot(8091, "apps", "databases", ctx)
	if err != nil {
		fmt.Printf("NewRoot() failed: %v\n", err)
		return
	}
	defer root.Destroy()

	/*{
		g_file_profile, err := os.Create("skyalt.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(g_file_profile)
		defer pprof.StopCPUProfile()
		//defer g_file_profile.Close()
	}*/

	run := true
	for run {
		run, err = root.Tick()
		if err != nil {
			fmt.Printf("Tick() failed: %v\n", err)
			return
		}
	}
}
