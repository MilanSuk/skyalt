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
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sync/atomic"
)

type ToolServerInfo struct {
	bytes_written atomic.Int64
	bytes_read    atomic.Int64
}

func (info *ToolServerInfo) AddReadBytes(size int) {
	info.bytes_read.Add(int64(size))
}
func (info *ToolServerInfo) AddWrittenBytes(size int) {
	info.bytes_written.Add(int64(size))
}

type ToolServerClient struct {
	info *ToolServerInfo
	conn net.Conn
}

type ToolServer struct {
	port     int
	listener net.Listener
	exiting  bool

	info *ToolServerInfo
}

func NewToolServer(port int) *ToolServer {
	server := &ToolServer{}

	port_last := port + 1000
	for port < port_last {
		var err error
		server.listener, err = net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			break
		}
		port++
	}
	if port == port_last {
		log.Fatal(fmt.Errorf("can not Listen()"))
	}
	server.port = port
	server.info = &ToolServerInfo{}

	fmt.Printf("Server is running on port: %d\n", server.port)
	return server
}

func (server *ToolServer) Destroy() {
	server.exiting = true
	server.listener.Close()

	fmt.Printf("Server port: %d closed\n", server.port)
}

func (server *ToolServer) Accept() (*ToolServerClient, error) {
	conn, err := server.listener.Accept()
	if err != nil {
		if server.exiting {
			return nil, nil
		}
		return nil, err
	}
	return &ToolServerClient{info: server.info, conn: conn}, nil
}

func (client *ToolServerClient) Destroy() {
	client.conn.Close()
}

func (client *ToolServerClient) ReadInt() (uint64, error) {
	var sz [8]byte
	_, err := client.conn.Read(sz[:])
	if err != nil {
		return 0, err
	}
	client.info.AddReadBytes(8)

	return binary.LittleEndian.Uint64(sz[:]), nil
}

func (client *ToolServerClient) WriteInt(value uint64) error {
	var val [8]byte
	binary.LittleEndian.PutUint64(val[:], value)
	_, err := client.conn.Write(val[:])
	if err != nil {
		return err
	}

	client.info.AddWrittenBytes(8)
	return nil
}

func (client *ToolServerClient) ReadArray() ([]byte, error) {
	//recv size
	size, err := client.ReadInt()
	if err != nil {
		return nil, err
	}

	//recv data
	data := make([]byte, size)
	p := 0
	for p < int(size) {
		n, err := client.conn.Read(data[p:])
		if err != nil {
			return nil, err
		}
		p += n
	}

	client.info.AddReadBytes(int(size))

	return data, nil
}

func (client *ToolServerClient) WriteArray(data []byte) error {
	//send size
	err := client.WriteInt(uint64(len(data)))
	if err != nil {
		return err
	}

	//send data
	_, err = client.conn.Write(data)
	if err != nil {
		return err
	}
	client.info.AddWrittenBytes(len(data))

	return nil
}

type ToolClient struct {
	conn *net.TCPConn
}

func NewToolClient(addr string, port int) (*ToolClient, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	return &ToolClient{conn: conn}, nil
}
func (client *ToolClient) Destroy() {
	client.conn.Close()
}

func (client *ToolClient) ReadInt() (uint64, error) {
	var sz [8]byte
	_, err := client.conn.Read(sz[:])
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint64(sz[:]), nil
}

func (client *ToolClient) WriteInt(value uint64) error {
	var val [8]byte
	binary.LittleEndian.PutUint64(val[:], value)
	_, err := client.conn.Write(val[:])
	if err != nil {
		return err
	}
	return nil
}

func (client *ToolClient) ReadArray() ([]byte, error) {
	//recv size
	size, err := client.ReadInt()
	if err != nil {
		return nil, err
	}

	//recv data
	data := make([]byte, size)
	p := 0
	for p < int(size) {
		n, err := client.conn.Read(data[p:])
		if err != nil {
			return nil, err
		}
		p += n
	}

	return data, nil
}

func (client *ToolClient) WriteArray(data []byte) error {
	//send size
	err := client.WriteInt(uint64(len(data)))
	if err != nil {
		return err
	}

	//send data
	_, err = client.conn.Write(data)
	if err != nil {
		return err
	}

	return nil
}
