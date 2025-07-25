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

type NetServerClient struct {
	conn net.Conn
}

type NetServer struct {
	port     int
	listener net.Listener
	exiting  bool
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
		log.Fatal("can not Listen()")
	}
	server.port = port

	fmt.Printf("Media server is running on port: %d\n", server.port)
	return server
}

func (server *NetServer) Destroy() {
	server.exiting = true
	server.listener.Close()

	fmt.Printf("Media server port: %d closed\n", server.port)
}

func (server *NetServer) Accept() (*NetServerClient, error) {
	conn, err := server.listener.Accept()
	if err != nil {
		if server.exiting {
			return nil, nil
		}
		return nil, err
	}
	return &NetServerClient{conn: conn}, nil
}

func (client *NetServerClient) Destroy() {
	client.conn.Close()
}

func (client *NetServerClient) ReadInt() (uint64, error) {
	var sz [8]byte
	_, err := client.conn.Read(sz[:])
	if err != nil {
		log.Fatal(err)
	}
	return binary.LittleEndian.Uint64(sz[:]), nil
}

func (client *NetServerClient) WriteInt(value uint64) error {
	var val [8]byte
	binary.LittleEndian.PutUint64(val[:], value)
	_, err := client.conn.Write(val[:])
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func (client *NetServerClient) ReadArray() ([]byte, error) {
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
			log.Fatal(err)
		}
		p += n
	}

	return data, nil
}

func (client *NetServerClient) WriteArray(data []byte) error {
	//send size
	err := client.WriteInt(uint64(len(data)))
	if err != nil {
		return err
	}

	//send data
	_, err = client.conn.Write(data)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}
