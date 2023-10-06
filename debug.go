/*
Copyright 2023 Milan Suk

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this db except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by assetlicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"net"
	"strconv"
	"sync"
)

type DebugServer struct {
	mu     sync.Mutex
	listen net.Listener
	assets []*AssetDebug
}

func NewDebugServer(PORT int) (*DebugServer, error) {
	var server DebugServer

	var err error
	server.listen, err = net.Listen("tcp", "localhost:"+strconv.Itoa(PORT))
	if err != nil {
		return nil, fmt.Errorf("Listen() failed: %w", err)
	}

	go func() {
		for {
			conn, err := server.listen.Accept()
			if err == nil {
				server.mu.Lock()
				server.assets = append(server.assets, NewAssetDebug(conn))
				server.mu.Unlock()
			}
		}
	}()

	return &server, nil
}

func (server *DebugServer) Destroy() {
	//close connections
	server.mu.Lock()
	defer server.mu.Unlock()
	for _, asset := range server.assets {
		asset.Destroy()
	}

	//close server
	server.listen.Close()
}

func (server *DebugServer) Find(assetName string) *AssetDebug {
	server.mu.Lock()
	defer server.mu.Unlock()

	for _, asset := range server.assets {
		if asset.name == assetName && asset.conn != nil {
			return asset
		}
	}

	return nil
}
