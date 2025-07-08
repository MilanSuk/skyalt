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

type AppsClient struct {
	conn *net.TCPConn
}

func NewToolsClient(addr string, port int) (*AppsClient, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", addr, port))
	if LogsError(err) != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if LogsError(err) != nil {
		return nil, err
	}

	return &AppsClient{conn: conn}, nil
}
func (client *AppsClient) Destroy() {
	client.conn.Close()
}

func (client *AppsClient) ReadInt() (uint64, error) {
	var sz [8]byte
	_, err := client.conn.Read(sz[:])
	if LogsError(err) != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint64(sz[:]), nil
}

func (client *AppsClient) WriteInt(value uint64) error {
	var val [8]byte
	binary.LittleEndian.PutUint64(val[:], value)
	_, err := client.conn.Write(val[:])
	if LogsError(err) != nil {
		return err
	}
	return nil
}

func (client *AppsClient) ReadArray() ([]byte, error) {
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
		if LogsError(err) != nil {
			return nil, err
		}
		p += n
	}

	return data, nil
}

func (client *AppsClient) WriteArray(data []byte) error {
	//send size
	err := client.WriteInt(uint64(len(data)))
	if err != nil {
		return err
	}

	//send data
	_, err = client.conn.Write(data)
	if LogsError(err) != nil {
		return err
	}

	return nil
}

type AppsServerInfo struct {
	bytes_written atomic.Int64
	bytes_read    atomic.Int64
}

func (info *AppsServerInfo) AddReadBytes(size int) {
	info.bytes_read.Add(int64(size))
}
func (info *AppsServerInfo) AddWrittenBytes(size int) {
	info.bytes_written.Add(int64(size))
}

func (info *AppsServerInfo) Print() {
	fmt.Println("Server stats: written", Tools_FormatBytes(int(info.bytes_written.Load())))
	fmt.Println("Server stats: read", Tools_FormatBytes(int(info.bytes_read.Load())))
}

type AppsServerClient struct {
	info *AppsServerInfo
	conn net.Conn
}

type AppsServer struct {
	port     int
	listener net.Listener
	exiting  bool

	info *AppsServerInfo
}

func NewAppsServer(port int) *AppsServer {
	server := &AppsServer{}

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
		log.Fatal("can not Listen()")
	}
	server.port = port
	server.info = &AppsServerInfo{}

	fmt.Printf("Server is running on port: %d\n", server.port)
	return server
}

func (server *AppsServer) Destroy() {
	server.exiting = true
	server.listener.Close()

	server.info.Print()
	fmt.Printf("App server port: %d closed\n", server.port)
}

func (server *AppsServer) Accept() (*AppsServerClient, error) {
	conn, err := server.listener.Accept()
	if err != nil {
		if server.exiting {
			return nil, nil
		}
		return nil, err
	}
	return &AppsServerClient{info: server.info, conn: conn}, nil
}

func (client *AppsServerClient) Destroy() {
	client.conn.Close()
}

func (client *AppsServerClient) ReadInt() (uint64, error) {
	var sz [8]byte
	_, err := client.conn.Read(sz[:])
	if LogsError(err) != nil {
		return 0, err
	}
	client.info.AddReadBytes(8)

	return binary.LittleEndian.Uint64(sz[:]), nil
}

func (client *AppsServerClient) WriteInt(value uint64) error {
	var val [8]byte
	binary.LittleEndian.PutUint64(val[:], value)
	_, err := client.conn.Write(val[:])
	if LogsError(err) != nil {
		return err
	}

	client.info.AddWrittenBytes(8)
	return nil
}

func (client *AppsServerClient) ReadArray() ([]byte, error) {
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
		if LogsError(err) != nil {
			return nil, err
		}
		p += n
	}

	client.info.AddReadBytes(int(size))

	return data, nil
}

func (client *AppsServerClient) WriteArray(data []byte) error {
	//send size
	err := client.WriteInt(uint64(len(data)))
	if err != nil {
		return err
	}

	//send data
	_, err = client.conn.Write(data)
	if LogsError(err) != nil {
		return err
	}
	client.info.AddWrittenBytes(len(data))

	return nil
}
