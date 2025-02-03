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
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var _sdk_client *SDK_NetClient

func main() {
	log.SetFlags(log.Llongfile) //log.LstdFlags | log.Lshortfile

	if len(os.Args) < 2 {
		log.Fatal("missing 'port' argument: ", os.Args)
	}
	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	//connect to server
	_sdk_client = SDK_NewNetClient("localhost", port)
	defer _sdk_client.Destroy()

	//get tool input
	input := _sdk_client.ReadArray()
	var st _replace_with_tool_structure_
	err = json.Unmarshal(input, &st)
	if err != nil {
		log.Fatal(err)
	}

	//exe tool
	output := st.run()
	/*output, err := json.Marshal(st.run())
	if err != nil {
		log.Fatal(err)
	}*/
	/*ret := interface{}(st.run())
	output, ok := ret.([]byte)
	if !ok {
		output, err = json.Marshal(ret)
		if err != nil {
			log.Fatal(err)
		}
	}*/

	//send back result
	_sdk_client.WriteInt(1)
	_sdk_client.WriteArray([]byte(output))

}

// use_case = "main", "programmer", "search"
func SDK_RunAgent(use_case string, max_iters int, max_tokens int, systemPrompt string, userPrompt string) string {
	_sdk_client.WriteInt(2)
	_sdk_client.WriteInt(uint64(max_iters))
	_sdk_client.WriteInt(uint64(max_tokens))
	_sdk_client.WriteArray([]byte(use_case))
	_sdk_client.WriteArray([]byte(systemPrompt))
	_sdk_client.WriteArray([]byte(userPrompt))

	js := _sdk_client.ReadArray()
	return string(js)
}

func SDK_SetToolCode(toolName string, code string) string {
	_sdk_client.WriteInt(3)
	_sdk_client.WriteArray([]byte(toolName))
	_sdk_client.WriteArray([]byte(code))

	js := _sdk_client.ReadArray()
	return string(js)
}
func SDK_Sandbox_violation(err error) bool {
	_sdk_client.WriteInt(4)
	_sdk_client.WriteArray([]byte(err.Error()))

	blockIt := _sdk_client.ReadInt()
	return blockIt != 0
}
func SDK_GetPassword(id string) (string, error) {
	_sdk_client.WriteInt(5)
	_sdk_client.WriteArray([]byte(id))

	ok := _sdk_client.ReadInt()
	if ok == 1 {
		//ok
		password := _sdk_client.ReadArray()
		return string(password), nil
	} else {
		//error
		errStr := _sdk_client.ReadArray()
		return "", errors.New(string(errStr))
	}
}

type SDK_NetClient struct {
	conn *net.TCPConn
}

func SDK_NewNetClient(addr string, port int) *SDK_NetClient {
	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatal(err)
	}

	return &SDK_NetClient{conn: conn}
}
func (client *SDK_NetClient) Destroy() {
	client.conn.Close()
}

func (client *SDK_NetClient) ReadInt() uint64 {
	var sz [8]byte
	_, err := client.conn.Read(sz[:])
	if err != nil {
		log.Fatal(err)
	}

	return binary.LittleEndian.Uint64(sz[:])
}

func (client *SDK_NetClient) WriteInt(value uint64) {
	var val [8]byte
	binary.LittleEndian.PutUint64(val[:], value)
	_, err := client.conn.Write(val[:])
	if err != nil {
		log.Fatal(err)
	}
}

func (client *SDK_NetClient) ReadArray() []byte {
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

func (client *SDK_NetClient) WriteArray(data []byte) {
	//send size
	client.WriteInt(uint64(len(data)))

	//send data
	_, err := client.conn.Write(data)
	if err != nil {
		log.Fatal(err)
	}
}

func _sandbox_isPathValid(name string) bool {
	curr, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	curr1 := filepath.Join(curr, "disk")
	curr2 := filepath.Join(curr, "tools")

	path, err := filepath.Abs(name)
	if err != nil {
		log.Fatal(err)
	}

	valid1 := strings.HasPrefix(path, curr1)
	valid2 := strings.HasPrefix(path, curr2)
	ok := (valid1 || valid2)
	if !ok {
		SDK_Sandbox_violation(fmt.Errorf("path '%s' is outside of program '%s' folder", path, curr))
	}
	return ok
}

func _exec_Command(name string, arg ...string) *exec.Cmd {
	SDK_Sandbox_violation(fmt.Errorf("command '%s' blocked", name))
	//exec.Command()
	return nil
}

func _exec_CommandContext(ctx context.Context, name string, arg ...string) *exec.Cmd {
	//exec.CommandContext()
	return nil
}

func _exec_StartProcess(name string, argv []string, attr *os.ProcAttr) (*os.Process, error) {
	//os.StartProcess()
	return nil, fmt.Errorf("StartProcess(%s, %v, %v) was blocked", name, argv, attr)
}

func _os_WriteFile(name string, data []byte, perm os.FileMode) error {
	if !_sandbox_isPathValid(name) {
		return fmt.Errorf("WriteFile(%s) outside program folder", name)
	}
	return os.WriteFile(name, data, perm)
}

func _os_Mkdir(path string, perm os.FileMode) error {
	if !_sandbox_isPathValid(path) {
		return fmt.Errorf("Mkdir(%s) outside program folder", path)
	}
	return os.Mkdir(path, perm)
}
func _os_MkdirAll(path string, perm os.FileMode) error {
	if !_sandbox_isPathValid(path) {
		return fmt.Errorf("MkdirAll(%s) outside program folder", path)
	}
	return os.MkdirAll(path, perm)
}

func _os_Remove(path string) error {
	if !_sandbox_isPathValid(path) {
		return fmt.Errorf("Remove(%s) outside program folder", path)
	}
	return os.Remove(path)
}

func _os_RemoveAll(path string) error {
	if !_sandbox_isPathValid(path) {
		return fmt.Errorf("RemoveAll(%s) outside program folder", path)
	}
	return os.RemoveAll(path)
}

func _os_Rename(oldpath, newpath string) error {
	if !_sandbox_isPathValid(oldpath) || !_sandbox_isPathValid(newpath) {
		return fmt.Errorf("Rename(%s, %s) outside program folder", oldpath, newpath)
	}
	return os.Rename(oldpath, newpath)
}

func _os_Chmod(path string, mode fs.FileMode) error {
	if !_sandbox_isPathValid(path) {
		return fmt.Errorf(" Chmod(%s) outside program folder", path)
	}
	return os.Chmod(path, mode)
}
func _os_Chdir(path string) error {
	if !_sandbox_isPathValid(path) {
		return fmt.Errorf("Chdir(%s) outside program folder", path)
	}
	return os.Chdir(path)
}

func _os_Create(path string) (*os.File, error) {
	if !_sandbox_isPathValid(path) {
		return nil, fmt.Errorf("Create(%s) outside program folder", path)
	}
	return os.Create(path)
}

func _os_OpenFile(path string, flag int, perm fs.FileMode) (*os.File, error) {
	if !_sandbox_isPathValid(path) && flag != os.O_RDONLY {
		return nil, fmt.Errorf("OpenFile(%s) outside program folder", path)
	}
	return os.OpenFile(path, flag, perm)
}

func _os_Lchown(path string, uid, gid int) error {
	if !_sandbox_isPathValid(path) {
		return fmt.Errorf("Lchown(%s) outside program folder", path)
	}
	return os.Lchown(path, uid, gid)
}

func _os_Truncate(path string, size int64) error {
	if !_sandbox_isPathValid(path) {
		return fmt.Errorf("Truncate(%s) outside program folder", path)
	}
	return os.Truncate(path, size)
}

func _os_Link(oldpath, newpath string) error {
	if !_sandbox_isPathValid(oldpath) || !_sandbox_isPathValid(newpath) {
		return fmt.Errorf("Link(%s, %s) outside program folder", oldpath, newpath)
	}
	return os.Link(oldpath, newpath)
}

func _os_Symlink(oldpath, newpath string) error {
	if !_sandbox_isPathValid(oldpath) || !_sandbox_isPathValid(newpath) {
		return fmt.Errorf("Symlink(%s, %s) outside program folder", oldpath, newpath)
	}
	return os.Symlink(oldpath, newpath)
}

func _os_NewFile(fd uintptr, path string) *os.File {
	if !_sandbox_isPathValid(path) {
		//fmt.Errorf("NewFile(%s) outside program folder", path)
		return nil
	}
	return os.NewFile(fd, path)
}
