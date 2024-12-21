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
)

type NetClient struct {
	conn *net.TCPConn
	lock sync.Mutex
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
