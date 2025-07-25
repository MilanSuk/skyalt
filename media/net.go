package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

type NetClient struct {
	conn *net.TCPConn
}

func NewNetClient(addr string, port int) *NetClient {
	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatal(err)
	}

	return &NetClient{conn: conn}
}
func (client *NetClient) Destroy() {
	client.conn.Close()
}

func (client *NetClient) ReadInt() uint64 {
	var sz [8]byte
	_, err := client.conn.Read(sz[:])
	if err != nil {
		log.Fatal(err)
	}

	return binary.LittleEndian.Uint64(sz[:])
}

func (client *NetClient) WriteInt(value uint64) {
	var val [8]byte
	binary.LittleEndian.PutUint64(val[:], value)
	_, err := client.conn.Write(val[:])
	if err != nil {
		log.Fatal(err)
	}
}
func (client *NetClient) WriteBool(value bool) {
	if value {
		client.WriteInt(1)
	} else {
		client.WriteInt(0)
	}
}
func (client *NetClient) ReadArray() []byte {
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

func (client *NetClient) WriteArray(data []byte) {
	//send size
	client.WriteInt(uint64(len(data)))

	//send data
	_, err := client.conn.Write(data)
	if err != nil {
		log.Fatal(err)
	}
}
