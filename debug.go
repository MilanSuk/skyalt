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
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
)

type DebugServer struct {
	mu     sync.Mutex
	listen net.Listener
	apps   []*AppDebug
}

func NewDebugServer(port int) (*DebugServer, error) {
	var server DebugServer

	var err error
	server.listen, err = net.Listen("tcp", "localhost:"+strconv.Itoa(port))
	if err != nil {
		return nil, fmt.Errorf("Listen() failed: %w", err)
	}

	fmt.Printf("Running DebugServer on port: %d\n", port)

	go func() {
		for {
			conn, err := server.listen.Accept()
			if err == nil {
				server.mu.Lock()
				server.apps = append(server.apps, NewAppDebug(conn))
				server.mu.Unlock()
			}
			if errors.Is(err, net.ErrClosed) {
				break
			}
		}
		fmt.Printf("Closing DebugServer on port: %d\n", port)
	}()

	return &server, nil
}

func (server *DebugServer) Destroy() {
	//close connections
	server.mu.Lock()
	defer server.mu.Unlock()
	for _, app := range server.apps {
		app.Destroy()
	}

	//close server
	server.listen.Close()
}

func (server *DebugServer) Get(appName string) *AppDebug {
	server.mu.Lock()
	defer server.mu.Unlock()

	for i, app := range server.apps {
		if app.name == appName {
			server.apps = append(server.apps[:i], server.apps[i+1:]...) //remove
			return app
		}
	}

	return nil
}
