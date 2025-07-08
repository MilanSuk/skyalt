package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

type ToolClient struct {
	conn *net.TCPConn
}

func NewToolClient(addr string, port int) *ToolClient {
	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatal(err)
	}

	return &ToolClient{conn: conn}
}
func (client *ToolClient) Destroy() {
	client.conn.Close()
}

func (client *ToolClient) ReadInt() uint64 {
	var sz [8]byte
	_, err := client.conn.Read(sz[:])
	if err != nil {
		log.Fatal(err)
	}

	return binary.LittleEndian.Uint64(sz[:])
}

func (client *ToolClient) WriteInt(value uint64) {
	var val [8]byte
	binary.LittleEndian.PutUint64(val[:], value)
	_, err := client.conn.Write(val[:])
	if err != nil {
		log.Fatal(err)
	}
}
func (client *ToolClient) WriteBool(value bool) {
	if value {
		client.WriteInt(1)
	} else {
		client.WriteInt(0)
	}
}
func (client *ToolClient) ReadArray() []byte {
	//recv size
	size := client.ReadInt()

	//recv data
	data := make([]byte, size)
	p := 0
	for p < int(size) {
		n, err := client.conn.Read(data[p:])
		if err != nil {
			log.Fatal(err)
		}
		p += n
	}

	return data
}

func (client *ToolClient) WriteArray(data []byte) {
	//send size
	client.WriteInt(uint64(len(data)))

	//send data
	_, err := client.conn.Write(data)
	if err != nil {
		log.Fatal(err)
	}
}
