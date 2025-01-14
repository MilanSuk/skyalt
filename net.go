/*
Copyright 2024 Milan Suk

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
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
)

type NetInfo struct {
	bytes_written atomic.Int64
	bytes_read    atomic.Int64
}

func (info *NetInfo) AddReadBytes(size int) {
	info.bytes_read.Add(int64(size))
}
func (info *NetInfo) AddWrittenBytes(size int) {
	info.bytes_written.Add(int64(size))
}

func (info *NetInfo) Print() {
	fmt.Println("Server stats: written", OsFormatBytes(int(info.bytes_written.Load())))
	fmt.Println("Server stats: read", OsFormatBytes(int(info.bytes_read.Load())))
}

type NetClient struct {
	conn *net.TCPConn
	lock sync.Mutex
}

type NerServerClient struct {
	info *NetInfo
	conn net.Conn
}

type NetServer struct {
	port     int
	listener net.Listener
	exiting  bool

	info *NetInfo
}

func NewNetServer(port int) *NetServer {
	server := &NetServer{}

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
	server.info = &NetInfo{}

	fmt.Printf("Server is running on port: %d\n", server.port)
	return server
}

func (server *NetServer) Destroy() {
	server.exiting = true
	server.listener.Close()

	server.info.Print()
	fmt.Printf("Server port: %d closed\n", server.port)
}

func (server *NetServer) Accept() (*NerServerClient, error) {
	conn, err := server.listener.Accept()
	if err != nil {
		if server.exiting {
			return nil, nil
		}
		return nil, err
	}
	return &NerServerClient{info: server.info, conn: conn}, nil
}

func (client *NerServerClient) Destroy() {
	client.conn.Close()
}

func (client *NerServerClient) ReadInt() (uint64, error) {
	var sz [8]byte
	_, err := client.conn.Read(sz[:])
	if err != nil {
		return 0, err
	}
	client.info.AddReadBytes(8)

	return binary.LittleEndian.Uint64(sz[:]), nil
}

func (client *NerServerClient) WriteInt(value uint64) error {
	var val [8]byte
	binary.LittleEndian.PutUint64(val[:], value)
	_, err := client.conn.Write(val[:])
	if err != nil {
		return err
	}

	client.info.AddWrittenBytes(8)
	return nil
}

func (client *NerServerClient) ReadArray() ([]byte, error) {
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

func (client *NerServerClient) WriteArray(data []byte) error {
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

func NewNetClient(addr string, port int) (*NetClient, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	return &NetClient{conn: conn}, nil
}
func (client *NetClient) Destroy() {
	client.conn.Close()
}
func (client *NetClient) Lock() {
	client.lock.Lock()
}
func (client *NetClient) Unlock() {
	client.lock.Unlock()
}
func (client *NetClient) ReadInt() (uint64, error) {
	var sz [8]byte
	_, err := client.conn.Read(sz[:])
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint64(sz[:]), nil
}

func (client *NetClient) WriteInt(value uint64) error {
	var val [8]byte
	binary.LittleEndian.PutUint64(val[:], value)
	_, err := client.conn.Write(val[:])
	if err != nil {
		return err
	}

	return nil
}

func (client *NetClient) ReadArray() ([]byte, error) {
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

func (client *NetClient) WriteArray(data []byte) error {
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

func OsMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		log.Fatal("OsMarshal failed:", err.Error())
	}
	return data

	//b := new(bytes.Buffer)
	//err := gob.NewEncoder(b).Encode(v)
	//if err != nil {
	//	log.Fatal("NewEncoder failed:", err.Error())
	//}
	//return b.Bytes()
}

func OsUnmarshal(data []byte, v interface{}) {
	err := json.Unmarshal(data, v)
	//b := bytes.NewBuffer(data)
	//err := gob.NewDecoder(b).Decode(v)
	if err != nil {
		fmt.Println(string(data))
		//str := err.Error()
		//fmt.Println(str[err.offset:1])
		log.Fatal("OsUnmarshal failed:", err.Error())
	}
}
