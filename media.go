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
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

type Media struct {
	lock   sync.Mutex
	server *AppsServer

	cmd_running bool

	client *AppsServerClient
}

func NewMedia(port int) (*Media, error) {
	media := &Media{}

	//create server
	media.server = NewAppsServer(port)

	return media, nil
}

func (media *Media) Destroy() {
	media.lock.Lock()
	defer media.lock.Unlock()

	if media.client != nil {
		media.client.WriteArray([]byte("exit"))
		media.client.Destroy()
		media.client = nil
	}

	media.server.Destroy()

	fmt.Printf("Media server port: %d closed\n", media.server.port)
}

func (media *Media) runProgram() error {
	//reset
	media.cmd_running = false
	media.client = nil

	//start
	cmd := exec.Command("./media/media", strconv.Itoa(media.server.port))
	cmd.Dir = ""
	OutStr := new(strings.Builder)
	ErrStr := new(strings.Builder)
	cmd.Stdout = OutStr
	cmd.Stderr = ErrStr
	err := cmd.Start()
	if err != nil {
		return LogsErrorf("media start failed: %w", err)
	}

	fmt.Printf("Media has started\n")
	media.cmd_running = true

	//run tool
	go func() {
		cmd.Wait()

		if OutStr.Len() > 0 {
			fmt.Printf("Media app output: %s\n", OutStr.String())
		}
		if ErrStr.Len() > 0 {
			fmt.Printf("\033[31mMedia app error:%s\033[0m\n", ErrStr.String())
		}

		media.cmd_running = false
	}()

	return nil
}

func (media *Media) check() error {
	if !media.cmd_running {
		err := media.runProgram()
		if err != nil {
			return err
		}

		//wait for media to connect
		media.client, err = media.server.Accept()
		if err != nil {
			return LogsErrorf("media Accept() failed: %w", err)
		}
	}

	if media.client == nil {
		return LogsErrorf("media.client == nil")
	}

	return nil
}

func (media *Media) GetInfo() ([]byte, error) {
	media.lock.Lock()
	defer media.lock.Unlock()

	err := media.check()
	if err != nil {
		return nil, err
	}

	//write
	{
		err := media.client.WriteArray([]byte("info"))
		if err != nil {
			return nil, err
		}
	}

	//read
	{
		js, err := media.client.ReadArray()
		if err != nil {
			return nil, err
		}

		return js, nil
	}
}

func (media *Media) Type(path string) (int, error) {
	media.lock.Lock()
	defer media.lock.Unlock()

	err := media.check()
	if err != nil {
		return -1, err
	}

	//write
	{
		err := media.client.WriteArray([]byte("type"))
		if err != nil {
			return -1, err
		}
		err = media.client.WriteArray([]byte(path))
		if err != nil {
			return -1, err
		}
	}

	//read
	{
		tp, err := media.client.ReadInt()
		if err != nil {
			return -1, err
		}

		return int(tp), nil
	}
}

func (media *Media) Play(path string, playerID uint64, playIt bool) error {
	media.lock.Lock()
	defer media.lock.Unlock()

	err := media.check()
	if err != nil {
		return err
	}

	//write
	{
		err = media.client.WriteArray([]byte("play"))
		if err != nil {
			return err
		}
		err = media.client.WriteArray([]byte(path))
		if err != nil {
			return err
		}
		err = media.client.WriteInt(playerID)
		if err != nil {
			return err
		}

		if playIt {
			err = media.client.WriteInt(1)
			if err != nil {
				return err
			}
		} else {
			err = media.client.WriteInt(0)
			if err != nil {
				return err
			}
		}

	}

	//read
	{
		errBytes, err := media.client.ReadArray()
		if err != nil {
			return err
		}
		if len(errBytes) > 0 {
			return LogsError(errors.New(string(errBytes)))
		}
	}

	return nil
}

func (media *Media) Seek(path string, playerID uint64, play_pos uint64) error {
	media.lock.Lock()
	defer media.lock.Unlock()

	err := media.check()
	if err != nil {
		return err
	}

	//write
	{
		err = media.client.WriteArray([]byte("seek"))
		if err != nil {
			return err
		}
		err = media.client.WriteArray([]byte(path))
		if err != nil {
			return err
		}
		err = media.client.WriteInt(playerID)
		if err != nil {
			return err
		}
		err = media.client.WriteInt(play_pos)
		if err != nil {
			return err
		}
	}

	//read
	{
		errBytes, err := media.client.ReadArray()
		if err != nil {
			return err
		}
		if len(errBytes) > 0 {
			return LogsError(errors.New(string(errBytes)))
		}
	}

	return nil
}

func (media *Media) Check(path string, playerID uint64) (bool, bool) {
	media.lock.Lock()
	defer media.lock.Unlock()

	err := media.check()
	if err != nil {
		return false, false
	}

	//write
	{
		err = media.client.WriteArray([]byte("check"))
		if err != nil {
			return false, false
		}
		err = media.client.WriteArray([]byte(path))
		if err != nil {
			return false, false
		}
		err = media.client.WriteInt(playerID)
		if err != nil {
			return false, false
		}
	}

	//read
	{
		playing, err := media.client.ReadInt()
		if err != nil {
			return false, false
		}
		changed, err := media.client.ReadInt()
		if err != nil {
			return false, false
		}
		return playing > 0, changed > 0
	}
}

func (media *Media) Frame(path string, blob []byte, playerID uint64, addError bool) (int, int, []byte, uint64, uint64, int, error) {
	media.lock.Lock()
	defer media.lock.Unlock()

	err := media.check()
	if err != nil {
		return 0, 0, nil, 0, 0, -1, err
	}

	//write
	{
		err = media.client.WriteArray([]byte("frame"))
		if err != nil {
			return 0, 0, nil, 0, 0, -1, err
		}
		err = media.client.WriteArray([]byte(path))
		if err != nil {
			return 0, 0, nil, 0, 0, -1, err
		}
		err = media.client.WriteArray(blob)
		if err != nil {
			return 0, 0, nil, 0, 0, -1, err
		}
		err = media.client.WriteInt(playerID)
		if err != nil {
			return 0, 0, nil, 0, 0, -1, err
		}
	}

	//read
	{
		errBytes, err := media.client.ReadArray()
		if err != nil {
			return 0, 0, nil, 0, 0, -1, err
		}
		width, err := media.client.ReadInt()
		if err != nil {
			return 0, 0, nil, 0, 0, -1, err
		}
		height, err := media.client.ReadInt()
		if err != nil {
			return 0, 0, nil, 0, 0, -1, err
		}
		data, err := media.client.ReadArray()
		if err != nil {
			return 0, 0, nil, 0, 0, -1, err
		}
		play_pos, err := media.client.ReadInt()
		if err != nil {
			return 0, 0, nil, 0, 0, -1, err
		}
		play_duration, err := media.client.ReadInt()
		if err != nil {
			return 0, 0, nil, 0, 0, -1, err
		}

		tp, err := media.client.ReadInt()
		if err != nil {
			return 0, 0, nil, 0, 0, -1, err
		}

		if len(errBytes) > 0 {

			err := errors.New(string(errBytes))
			if addError {
				LogsError(err)
			}
			return 0, 0, nil, 0, 0, -1, err
		}

		return int(width), int(height), data, play_pos, play_duration, int(tp), nil
	}
}
